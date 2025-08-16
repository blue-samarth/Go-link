package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
	"errors"
	"path/filepath"
	"os"
	"sort"
	"strconv"
	
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/driver/postgres"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/blue-samarth/go-link/config"
	"github.com/blue-samarth/go-link/utils"
)

var (
	ErrDatabaseNotConnected = errors.New("database not connected")
	ErrInvalidConfig        = errors.New("invalid database configuration")
	ErrDatabaseNil          = errors.New("database instance is nil")
	ErrConnectionPoolExhausted = errors.New("connection pool exhausted")
	ErrConnectionTimeout    = errors.New("connection timeout")
	ErrUnsupportedDatabase  = errors.New("unsupported database type")
	ErrCircuitBreakerOpen   = errors.New("circuit breaker is open")
	ErrMigrationNotFound    = errors.New("migration not found")
	ErrMigrationFailed      = errors.New("migration failed")
)

// Circuit Breaker Implementation
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

type CircuitBreakerConfig struct {
	MaxFailures     int
	ResetTimeout    time.Duration
	FailureThreshold float64
	MinRequests     int
}

func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		MaxFailures:     5,
		ResetTimeout:    60 * time.Second,
		FailureThreshold: 0.5,
		MinRequests:     10,
	}
}

type CircuitBreaker struct {
	config        *CircuitBreakerConfig
	state         CircuitBreakerState
	failures      int
	requests      int
	lastFailTime  time.Time
	mutex         sync.RWMutex
	tracer        trace.Tracer
}

func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}
	
	return &CircuitBreaker{
		config: config,
		state:  CircuitBreakerClosed,
		tracer: otel.Tracer("database-circuit-breaker"),
	}
}

func (cb *CircuitBreaker) Execute(ctx context.Context, operation func() error) error {
	ctx, span := cb.tracer.Start(ctx, "circuit_breaker_execute")
	defer span.End()
	
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// Check if circuit breaker should transition from Open to Half-Open
	if cb.state == CircuitBreakerOpen {
		if time.Since(cb.lastFailTime) > cb.config.ResetTimeout {
			cb.state = CircuitBreakerHalfOpen
			cb.failures = 0
			cb.requests = 0
			span.AddEvent("Circuit breaker transitioned to half-open")
		} else {
			span.SetStatus(codes.Error, "Circuit breaker is open")
			span.SetAttributes(attribute.String("state", "open"))
			return ErrCircuitBreakerOpen
		}
	}

	// Execute operation
	cb.requests++
	err := operation()
	
	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()
		
		// Check if we should open the circuit
		if cb.shouldOpen() {
			cb.state = CircuitBreakerOpen
			span.AddEvent("Circuit breaker opened due to failures")
			span.SetAttributes(
				attribute.Int("failures", cb.failures),
				attribute.Int("requests", cb.requests),
				attribute.String("state", "open"),
			)
		}
		
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	// Success - reset if in half-open state
	if cb.state == CircuitBreakerHalfOpen {
		cb.state = CircuitBreakerClosed
		cb.failures = 0
		cb.requests = 0
		span.AddEvent("Circuit breaker closed after successful operation")
	}

	span.SetAttributes(attribute.String("state", string(rune(cb.state))))
	return nil
}

func (cb *CircuitBreaker) shouldOpen() bool {
	if cb.requests < cb.config.MinRequests {
		return false
	}
	
	failureRate := float64(cb.failures) / float64(cb.requests)
	return failureRate >= cb.config.FailureThreshold || cb.failures >= cb.config.MaxFailures
}

func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// Migration Support
type Migration struct {
	Version     int
	Description string
	UpSQL       string
	DownSQL     string
	AppliedAt   *time.Time
}

type MigrationManager struct {
	db             *gorm.DB
	migrationsPath string
	tableName      string
	tracer         trace.Tracer
}

func NewMigrationManager(db *gorm.DB, migrationsPath string) *MigrationManager {
	return &MigrationManager{
		db:             db,
		migrationsPath: migrationsPath,
		tableName:      "schema_migrations",
		tracer:         otel.Tracer("database-migrations"),
	}
}

func (mm *MigrationManager) Initialize(ctx context.Context) error {
	ctx, span := mm.tracer.Start(ctx, "initialize_migrations")
	defer span.End()

	// Create migrations table if it doesn't exist
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`

	if err := mm.db.WithContext(ctx).Exec(createTableSQL).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create migrations table")
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	span.SetStatus(codes.Ok, "Migrations table initialized")
	return nil
}

func (mm *MigrationManager) LoadMigrations(ctx context.Context) ([]*Migration, error) {
	ctx, span := mm.tracer.Start(ctx, "load_migrations")
	defer span.End()

	files, err := filepath.Glob(filepath.Join(mm.migrationsPath, "*.sql"))
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to glob migration files: %w", err)
	}

	migrations := make([]*Migration, 0)
	for _, file := range files {
		migration, err := mm.parseMigrationFile(file)
		if err != nil {
			span.RecordError(err)
			utils.Log(zapcore.ErrorLevel, "Failed to parse migration file", []string{"prod"},
				zap.String("file", file), zap.Error(err))
			continue
		}
		migrations = append(migrations, migration)
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	span.SetAttributes(attribute.Int("migrations_loaded", len(migrations)))
	return migrations, nil
}

func (mm *MigrationManager) parseMigrationFile(filePath string) (*Migration, error) {
	fileName := filepath.Base(filePath)
	
	// Expected format: 001_create_users.sql
	parts := strings.SplitN(fileName, "_", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid migration file name format: %s", fileName)
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid version number in file %s: %w", fileName, err)
	}

	description := strings.TrimSuffix(parts[1], ".sql")
	description = strings.ReplaceAll(description, "_", " ")

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration file %s: %w", filePath, err)
	}

	// Split UP and DOWN migrations
	sections := strings.Split(string(content), "-- DOWN")
	upSQL := strings.TrimSpace(sections[0])
	downSQL := ""
	
	if len(sections) > 1 {
		downSQL = strings.TrimSpace(sections[1])
	}

	return &Migration{
		Version:     version,
		Description: description,
		UpSQL:       upSQL,
		DownSQL:     downSQL,
	}, nil
}

func (mm *MigrationManager) GetAppliedMigrations(ctx context.Context) (map[int]*Migration, error) {
	ctx, span := mm.tracer.Start(ctx, "get_applied_migrations")
	defer span.End()

	var records []struct {
		Version     int       `gorm:"column:version"`
		Description string    `gorm:"column:description"`
		AppliedAt   time.Time `gorm:"column:applied_at"`
	}

	if err := mm.db.WithContext(ctx).Table(mm.tableName).Find(&records).Error; err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}

	applied := make(map[int]*Migration)
	for _, record := range records {
		applied[record.Version] = &Migration{
			Version:     record.Version,
			Description: record.Description,
			AppliedAt:   &record.AppliedAt,
		}
	}

	span.SetAttributes(attribute.Int("applied_migrations", len(applied)))
	return applied, nil
}

func (mm *MigrationManager) Migrate(ctx context.Context) error {
	ctx, span := mm.tracer.Start(ctx, "migrate")
	defer span.End()

	migrations, err := mm.LoadMigrations(ctx)
	if err != nil {
		span.RecordError(err)
		return err
	}

	applied, err := mm.GetAppliedMigrations(ctx)
	if err != nil {
		span.RecordError(err)
		return err
	}

	migrationsRun := 0
	for _, migration := range migrations {
		if _, exists := applied[migration.Version]; exists {
			continue // Already applied
		}

		if err := mm.applyMigration(ctx, migration); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}
		migrationsRun++
	}

	span.SetAttributes(attribute.Int("migrations_run", migrationsRun))
	utils.Log(zapcore.InfoLevel, "Migrations completed", []string{"prod"},
		zap.Int("migrations_run", migrationsRun))
	
	return nil
}

func (mm *MigrationManager) applyMigration(ctx context.Context, migration *Migration) error {
	ctx, span := mm.tracer.Start(ctx, "apply_migration")
	defer span.End()
	
	span.SetAttributes(
		attribute.Int("migration_version", migration.Version),
		attribute.String("migration_description", migration.Description),
	)

	tx := mm.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		span.RecordError(tx.Error)
		return tx.Error
	}

	// Execute migration SQL
	if err := tx.Exec(migration.UpSQL).Error; err != nil {
		tx.Rollback()
		span.RecordError(err)
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration as applied
	if err := tx.Exec("INSERT INTO schema_migrations (version, description) VALUES (?, ?)",
		migration.Version, migration.Description).Error; err != nil {
		tx.Rollback()
		span.RecordError(err)
		return fmt.Errorf("failed to record migration: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	utils.Log(zapcore.InfoLevel, "Applied migration", []string{"prod"},
		zap.Int("version", migration.Version),
		zap.String("description", migration.Description))

	span.SetStatus(codes.Ok, "Migration applied successfully")
	return nil
}

// Enhanced ConnectionMetrics with tracing
type ConnectionMetrics struct {
	MaxOpenConnections int
	OpenConnections    int
	InUse             int
	Idle              int
	WaitCount         int64
	WaitDuration      time.Duration
	MaxIdleClosed     int64
	MaxLifetimeClosed int64
	Connected         bool
	LastPingTime      time.Time
	LastPingError     error
	CircuitBreakerState CircuitBreakerState
}

type Database interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	IsConnected() bool
}

type DatabaseSQL interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	IsConnected() bool
	HealthCheck(ctx context.Context) error

	// GORM methods
	DB() (*gorm.DB, error)
	GetSQLDB() (*sql.DB, error)
	GetMetrics() ConnectionMetrics
	
	// Connection management
	ReconnectWithRetry(ctx context.Context) error
	ValidateConnection(ctx context.Context) error
	
	// Migration support
	GetMigrationManager() *MigrationManager
	RunMigrations(ctx context.Context) error
}

type ConnectionPoolConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func DefaultConnectionPoolConfig() *ConnectionPoolConfig {
	return &ConnectionPoolConfig{
		MaxIdleConns:    25,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
	}
}

type SQLDatabase struct {
	db         *gorm.DB
	sqlDB      *sql.DB
	connected  bool
	mutex      sync.RWMutex
	dbType     config.DbType
	cfg        *config.DbConfig
	poolConfig *ConnectionPoolConfig
	
	// Health monitoring
	lastPingTime  time.Time
	lastPingError error
	reconnectAttempts int
	
	// Circuit breaker
	circuitBreaker *CircuitBreaker
	
	// Migration manager
	migrationManager *MigrationManager
	
	// Tracing
	tracer trace.Tracer
}

// NewSQLDatabase creates a new SQL database instance with circuit breaker and tracing
func NewSQLDatabase(dbType config.DbType, cfg *config.DbConfig) (DatabaseSQL, error) {
	if cfg == nil {
		utils.Log(zapcore.ErrorLevel, "Database configuration is nil", []string{"prod"}, 
			zap.String("operation", "NewSQLDatabase"))
		return nil, fmt.Errorf("%w: configuration is nil", ErrInvalidConfig)
	}

	// Validate database type
	if !dbType.IsValid() {
		utils.Log(zapcore.ErrorLevel, "Invalid database type", []string{"prod"}, 
			zap.String("dbType", string(dbType)),
			zap.String("operation", "NewSQLDatabase"))
		return nil, fmt.Errorf("%w: unsupported database type %s", ErrUnsupportedDatabase, dbType)
	}

	// Use the config package's validation
	appCfg := &config.AppConfig{
		DbConfig: cfg,
	}
	if err := appCfg.Validate(); err != nil {
		utils.Log(zapcore.ErrorLevel, "Database configuration validation failed", []string{"prod"}, 
			zap.Error(err),
			zap.String("operation", "NewSQLDatabase"))
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	poolConfig := DefaultConnectionPoolConfig()
	circuitBreaker := NewCircuitBreaker(DefaultCircuitBreakerConfig())

	utils.Log(zapcore.InfoLevel, "Creating new SQL database instance", []string{"prod"}, 
		zap.String("dbType", string(dbType)),
		zap.String("host", cfg.DbHost),
		zap.Int("port", cfg.DbPort),
		zap.String("database", cfg.DbName))

	return &SQLDatabase{
		dbType:         dbType,
		cfg:            cfg,
		mutex:          sync.RWMutex{},
		poolConfig:     poolConfig,
		circuitBreaker: circuitBreaker,
		tracer:         otel.Tracer("database-sql"),
	}, nil
}

// configureConnectionPool applies connection pool configuration with tracing
func (s *SQLDatabase) configureConnectionPool(sqlDB *sql.DB) {
	_, span := s.tracer.Start(context.Background(), "configure_connection_pool")
	defer span.End()

	sqlDB.SetMaxIdleConns(s.poolConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(s.poolConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(s.poolConfig.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(s.poolConfig.ConnMaxIdleTime)
	
	span.SetAttributes(
		attribute.Int("max_idle_conns", s.poolConfig.MaxIdleConns),
		attribute.Int("max_open_conns", s.poolConfig.MaxOpenConns),
		attribute.String("conn_max_lifetime", s.poolConfig.ConnMaxLifetime.String()),
	)
	
	utils.Log(zapcore.InfoLevel, "Connection pool configured", []string{"prod"}, 
		zap.Int("maxIdleConns", s.poolConfig.MaxIdleConns),
		zap.Int("maxOpenConns", s.poolConfig.MaxOpenConns),
		zap.Duration("connMaxLifetime", s.poolConfig.ConnMaxLifetime))
}

// Connect establishes database connection with circuit breaker and tracing
func (s *SQLDatabase) Connect(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "database_connect")
	defer span.End()

	// Use circuit breaker for connection attempts
	return s.circuitBreaker.Execute(ctx, func() error {
		return s.doConnect(ctx)
	})
}

func (s *SQLDatabase) doConnect(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.connected {
		utils.Log(zapcore.DebugLevel, "Database already connected", []string{"dev"})
		return nil
	}

	ctx, span := s.tracer.Start(ctx, "do_connect")
	defer span.End()

	span.SetAttributes(
		attribute.String("db_type", string(s.dbType)),
		attribute.String("db_host", s.cfg.DbHost),
		attribute.Int("db_port", s.cfg.DbPort),
	)

	utils.Log(zapcore.InfoLevel, "Connecting to database", []string{"prod"}, 
		zap.String("dbType", string(s.dbType)),
		zap.String("host", s.cfg.DbHost),
		zap.Int("port", s.cfg.DbPort))

	dsn, err := s.buildDSN(s.dbType, s.cfg)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to build DSN")
		utils.Log(zapcore.ErrorLevel, "Failed to build DSN", []string{"prod"}, 
			zap.Error(err),
			zap.String("dbType", string(s.dbType)))
		return fmt.Errorf("failed to build DSN: %w", err)
	}

	var dialector gorm.Dialector

	switch s.dbType {
	case config.Postgres:
		dialector = postgres.Open(dsn)
	case config.MySQL:
		dialector = mysql.Open(dsn)
	case config.SQLite:
		dialector = sqlite.Open(dsn)
	case config.MsSQL:
		dialector = sqlserver.Open(dsn)
	default:
		err := fmt.Errorf("%w: %s", ErrUnsupportedDatabase, s.dbType)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Unsupported database type")
		utils.Log(zapcore.ErrorLevel, "Unsupported database type", []string{"prod"}, 
			zap.Error(err),
			zap.String("dbType", string(s.dbType)))
		return err
	}

	// Configure GORM for production with tracing
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time {
			return time.Now()
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		PrepareStmt: true,
	}

	// Add timeout to context if not present
	connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Attempt connection
	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to connect to database")
		utils.Log(zapcore.ErrorLevel, "Failed to connect to database", []string{"prod"}, 
			zap.Error(err),
			zap.String("dbType", string(s.dbType)),
			zap.String("host", s.cfg.DbHost))
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	
	sqlDB, err := db.DB()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get underlying sql.DB")
		utils.Log(zapcore.ErrorLevel, "Failed to get underlying sql.DB", []string{"prod"}, 
			zap.Error(err))
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	s.configureConnectionPool(sqlDB)

	// Test the connection
	if err := sqlDB.PingContext(connectCtx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Database ping failed")
		utils.Log(zapcore.ErrorLevel, "Database ping failed after connection", []string{"prod"}, 
			zap.Error(err))
		return fmt.Errorf("database ping failed: %w", err)
	}

	s.db = db
	s.sqlDB = sqlDB
	s.connected = true
	s.lastPingTime = time.Now()
	s.lastPingError = nil
	s.reconnectAttempts = 0

	// Initialize migration manager
	s.migrationManager = NewMigrationManager(s.db, "migrations")

	span.SetStatus(codes.Ok, "Database connected successfully")
	utils.Log(zapcore.InfoLevel, "Database connected successfully", []string{"prod"}, 
		zap.String("dbType", string(s.dbType)),
		zap.String("host", s.cfg.DbHost),
		zap.Duration("connectionTime", time.Since(s.lastPingTime)))

	return nil
}

// GetMigrationManager returns the migration manager instance
func (s *SQLDatabase) GetMigrationManager() *MigrationManager {
	return s.migrationManager
}

// RunMigrations executes pending database migrations
func (s *SQLDatabase) RunMigrations(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "run_migrations")
	defer span.End()

	if s.migrationManager == nil {
		err := errors.New("migration manager not initialized")
		span.RecordError(err)
		span.SetStatus(codes.Error, "Migration manager not initialized")
		return err
	}

	if err := s.migrationManager.Initialize(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	return s.migrationManager.Migrate(ctx)
}

// Ping tests database connectivity with circuit breaker and tracing
func (s *SQLDatabase) Ping(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "database_ping")
	defer span.End()

	return s.circuitBreaker.Execute(ctx, func() error {
		return s.doPing(ctx)
	})
}

func (s *SQLDatabase) doPing(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.connected || s.sqlDB == nil {
		err := ErrDatabaseNotConnected
		s.lastPingError = err
		return err
	}

	ctx, span := s.tracer.Start(ctx, "do_ping")
	defer span.End()

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.sqlDB.PingContext(pingCtx)
	s.lastPingTime = time.Now()
	s.lastPingError = err

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Database ping failed")
		utils.Log(zapcore.ErrorLevel, "Database ping failed", []string{"prod"}, 
			zap.Error(err))
		s.connected = false
		return fmt.Errorf("database ping failed: %w", err)
	}

	span.SetStatus(codes.Ok, "Database ping successful")
	utils.Log(zapcore.DebugLevel, "Database ping successful", []string{"dev"})
	return nil
}

// GetMetrics returns connection pool metrics including circuit breaker state
func (s *SQLDatabase) GetMetrics() ConnectionMetrics {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	metrics := ConnectionMetrics{
		Connected:           s.connected,
		LastPingTime:        s.lastPingTime,
		LastPingError:       s.lastPingError,
		CircuitBreakerState: s.circuitBreaker.GetState(),
	}

	if s.sqlDB != nil {
		stats := s.sqlDB.Stats()
		metrics.MaxOpenConnections = stats.MaxOpenConnections
		metrics.OpenConnections = stats.OpenConnections
		metrics.InUse = stats.InUse
		metrics.Idle = stats.Idle
		metrics.WaitCount = stats.WaitCount
		metrics.WaitDuration = stats.WaitDuration
		metrics.MaxIdleClosed = stats.MaxIdleClosed
		metrics.MaxLifetimeClosed = stats.MaxLifetimeClosed
	}

	return metrics
}

// ReconnectWithRetry attempts to reconnect with exponential backoff and tracing
func (s *SQLDatabase) ReconnectWithRetry(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "reconnect_with_retry")
	defer span.End()

	maxRetries := 5
	baseDelay := time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(1<<uint(attempt-1)) * baseDelay
			span.AddEvent("Retrying connection", trace.WithAttributes(
				attribute.Int("attempt", attempt+1),
				attribute.String("delay", delay.String()),
			))
			
			utils.Log(zapcore.WarnLevel, "Retrying database connection", []string{"prod"}, 
				zap.Int("attempt", attempt+1),
				zap.Int("maxRetries", maxRetries),
				zap.Duration("delay", delay))
			
			select {
			case <-ctx.Done():
				span.RecordError(ctx.Err())
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		if err := s.Connect(ctx); err != nil {
			span.AddEvent("Connection attempt failed", trace.WithAttributes(
				attribute.Int("attempt", attempt+1),
				attribute.String("error", err.Error()),
			))
			utils.Log(zapcore.WarnLevel, "Connection attempt failed", []string{"prod"}, 
				zap.Error(err),
				zap.Int("attempt", attempt+1))
			continue
		}

		span.SetStatus(codes.Ok, "Database reconnected successfully")
		utils.Log(zapcore.InfoLevel, "Database reconnected successfully", []string{"prod"}, 
			zap.Int("attempts", attempt+1))
		return nil
	}

	err := fmt.Errorf("failed to reconnect after %d attempts", maxRetries)
	span.RecordError(err)
	span.SetStatus(codes.Error, "Failed to reconnect after all retries")
	utils.Log(zapcore.ErrorLevel, "Failed to reconnect after all retries", []string{"prod"}, 
		zap.Int("maxRetries", maxRetries))
	return err
}

// Additional methods remain the same but with tracing added...
func (s *SQLDatabase) Disconnect(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "database_disconnect")
	defer span.End()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.connected {
		utils.Log(zapcore.DebugLevel, "Database already disconnected", []string{"dev"})
		return nil
	}

	utils.Log(zapcore.InfoLevel, "Disconnecting from database", []string{"prod"})

	if s.sqlDB != nil {
		if err := s.sqlDB.Close(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to close sql.DB")
			utils.Log(zapcore.ErrorLevel, "Failed to close sql.DB", []string{"prod"}, 
				zap.Error(err))
			return fmt.Errorf("failed to close sql.DB: %w", err)
		}
	}

	s.connected = false
	s.db = nil
	s.sqlDB = nil
	
	span.SetStatus(codes.Ok, "Database disconnected successfully")
	utils.Log(zapcore.InfoLevel, "Database disconnected successfully", []string{"prod"})
	return nil
}

// HealthCheck performs comprehensive database health validation with tracing
func (s *SQLDatabase) HealthCheck(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "database_health_check")
	defer span.End()

	if err := s.Ping(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	if err := s.ValidateConnection(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	span.SetStatus(codes.Ok, "Database health check passed")
	return nil
}

// ValidateConnection checks connection pool health with tracing
func (s *SQLDatabase) ValidateConnection(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "validate_connection")
	defer span.End()

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.connected || s.sqlDB == nil {
		err := ErrDatabaseNotConnected
		span.RecordError(err)
		span.SetStatus(codes.Error, "Database not connected")
		return err
	}

	stats := s.sqlDB.Stats()
	span.SetAttributes(
		attribute.Int("open_connections", stats.OpenConnections),
		attribute.Int("max_open_connections", stats.MaxOpenConnections),
		attribute.Int("in_use", stats.InUse),
		attribute.Int("idle", stats.Idle),
	)

	if stats.OpenConnections >= stats.MaxOpenConnections {
		err := ErrConnectionPoolExhausted
		span.RecordError(err)
		span.SetStatus(codes.Error, "Connection pool exhausted")
		utils.Log(zapcore.WarnLevel, "Connection pool nearly exhausted", []string{"prod"}, 
			zap.Int("openConnections", stats.OpenConnections),
			zap.Int("maxOpenConnections", stats.MaxOpenConnections))
		return err
	}

	span.SetStatus(codes.Ok, "Connection validation passed")
	return nil
}

// IsConnected returns current connection status
func (s *SQLDatabase) IsConnected() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.connected
}

// DB returns GORM database instance with error handling and tracing
func (s *SQLDatabase) DB() (*gorm.DB, error) {
	_, span := s.tracer.Start(context.Background(), "get_db")
	defer span.End()

	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	if s.db == nil {
		err := ErrDatabaseNotConnected
		span.RecordError(err)
		span.SetStatus(codes.Error, "Database not connected")
		return nil, err
	}
	
	span.SetStatus(codes.Ok, "Database instance retrieved")
	return s.db, nil
}

// GetSQLDB returns underlying sql.DB with error handling and tracing
func (s *SQLDatabase) GetSQLDB() (*sql.DB, error) {
	_, span := s.tracer.Start(context.Background(), "get_sql_db")
	defer span.End()

	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	if s.sqlDB == nil {
		err := ErrDatabaseNotConnected
		span.RecordError(err)
		span.SetStatus(codes.Error, "SQL database not connected")
		return nil, err
	}
	
	span.SetStatus(codes.Ok, "SQL database instance retrieved")
	return s.sqlDB, nil
}

// buildDSN constructs database connection string with proper field mapping and security
func (s *SQLDatabase) buildDSN(dbType config.DbType, cfg *config.DbConfig) (string, error) {
	switch dbType {
	case config.Postgres:
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			cfg.DbHost, cfg.DbUser, cfg.DbPass, cfg.DbName, cfg.DbPort, cfg.SSLMode)
		
		sanitized := fmt.Sprintf("host=%s user=%s dbname=%s port=%d sslmode=%s",
			cfg.DbHost, cfg.DbUser, cfg.DbName, cfg.DbPort, cfg.SSLMode)
		utils.Log(zapcore.DebugLevel, "Built PostgreSQL DSN", []string{"dev"}, 
			zap.String("dsn", sanitized))
		
		return dsn, nil
		
	case config.MySQL:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DbUser, cfg.DbPass, cfg.DbHost, cfg.DbPort, cfg.DbName)
		
		sanitized := fmt.Sprintf("%s:***@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DbUser, cfg.DbHost, cfg.DbPort, cfg.DbName)
		utils.Log(zapcore.DebugLevel, "Built MySQL DSN", []string{"dev"}, 
			zap.String("dsn", sanitized))
		
		return dsn, nil
		
	case config.SQLite:
		utils.Log(zapcore.DebugLevel, "Built SQLite DSN", []string{"dev"}, 
			zap.String("database", cfg.DbName))
		return cfg.DbName, nil
		
	case config.MsSQL:
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			cfg.DbUser, cfg.DbPass, cfg.DbHost, cfg.DbPort, cfg.DbName)
		
		sanitized := fmt.Sprintf("sqlserver://%s:***@%s:%d?database=%s",
			cfg.DbUser, cfg.DbHost, cfg.DbPort, cfg.DbName)
		utils.Log(zapcore.DebugLevel, "Built SQL Server DSN", []string{"dev"}, 
			zap.String("dsn", sanitized))
		
		return dsn, nil
		
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedDatabase, dbType)
	}
}

// SanitizeDSN removes sensitive information from DSN for logging
func SanitizeDSN(dsn string) string {
	sanitized := dsn
	if strings.Contains(sanitized, "password=") {
		parts := strings.Split(sanitized, " ")
		for i, part := range parts {
			if strings.HasPrefix(part, "password=") {
				parts[i] = "password=***"
			}
		}
		sanitized = strings.Join(parts, " ")
	}
	
	// Handle MySQL format
	if strings.Contains(sanitized, ":") && strings.Contains(sanitized, "@") {
		parts := strings.Split(sanitized, "@")
		if len(parts) > 1 {
			userPass := strings.Split(parts[0], ":")
			if len(userPass) > 1 {
				sanitized = userPass[0] + ":***@" + strings.Join(parts[1:], "@")
			}
		}
	}
	
	return sanitized
}