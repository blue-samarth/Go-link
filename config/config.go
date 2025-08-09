package config

import (
	"fmt"
	"log"
	"net"
	"os"
	"reflect"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type DbType string

const (
	Postgres DbType = "postgres"
	MongoDb  DbType = "mongodb"
	MySQL    DbType = "mysql"
	MsSQL    DbType = "mssql"
	SQLite   DbType = "sqlite"
)

func (d DbType) IsValid() bool {
	switch d {
	case Postgres, MongoDb, MySQL, MsSQL, SQLite:
		return true
	}
	return false
}

type DbConfig struct {
	DbHost  string `yaml:"db_host" env:"DB_HOST" env-default:"localhost"`
	DbPort  int    `yaml:"db_port" env:"DB_PORT" env-default:"5432"`
	DbUser  string `yaml:"db_user" env:"DB_USER" env-default:"Admin"`
	DbPass  string `yaml:"db_pass" env:"DB_PASS" env-default:"Admin123" secret:"true"`
	DbName  string `yaml:"db_name" env:"DB_NAME" env-default:"dbname"`
	SSLMode string `yaml:"ssl_mode" env:"SSL_MODE" env-default:"disable"`
}

// MongoConfig holds configuration for MongoDB database connections.
type MongoConfig struct {
	Host       string `yaml:"host" env:"MONGO_HOST" env-default:"localhost"`
	Port       int    `yaml:"port" env:"MONGO_PORT" env-default:"27017"`
	User       string `yaml:"user" env:"MONGO_USER" env-default:"admin"`
	Pass       string `yaml:"pass" env:"MONGO_PASS" env-default:"changeme" secret:"true"`
	DbName     string `yaml:"db_name" env:"MONGO_DB_NAME" env-default:"app_db"`
	ReplicaSet string `yaml:"replica_set" env:"REPLICA_SET" env-default:""`

	// Optional: Support for connection string as alternative
	ConnectionString string `yaml:"connection_string" env:"MONGO_CONNECTION_STRING" env-default:""`
}


type LogConfig struct {
	LogLevel         string `yaml:"log_level" env:"LOG_LEVEL" env-default:"info"`       // Log level (debug, info, warn, error)
	LogMode          string `yaml:"log_mode" env:"LOG_MODE" env-default:"dev"`          // Logging mode: dev, prod, minimal
	LogFile          string `yaml:"log_file" env:"LOG_FILE" env-default:""`             // Log file path (empty = stdout/stderr)
	EnableRotation   bool   `yaml:"enable_rotation" env:"ENABLE_ROTATION" env-default:"false"`
	RotationMaxSize  int    `yaml:"rotation_max_size" env:"ROTATION_MAX_SIZE" env-default:"100"`       // Max size in MB before rotation
	RotationMaxBackups int  `yaml:"rotation_max_backups" env:"ROTATION_MAX_BACKUPS" env-default:"10"`  // Max number of backup files to keep
	RotationMaxAge   int    `yaml:"rotation_max_age" env:"ROTATION_MAX_AGE" env-default:"30"`          // Max age of log files in days
	RotationCompress bool   `yaml:"rotation_compress" env:"ROTATION_COMPRESS" env-default:"true"`      // Compress rotated files
	LogBufferSize    int    `yaml:"log_buffer_size" env:"LOG_BUFFER_SIZE" env-default:"1000"`          // Buffer size for async logging
	LogFlushInterval int    `yaml:"log_flush_interval" env:"LOG_FLUSH_INTERVAL" env-default:"5"`      // Flush interval in seconds
}


type AppConfig struct {
	// Server configuration
	AppPort int    `yaml:"app_port" env:"APP_PORT" env-default:"8080"`
	Env     string `yaml:"env" env:"ENV" env-default:"development"`

	// Security configuration
	JWTSecretKey       string `yaml:"jwt_secret_key" env:"JWT_SECRET_KEY" env-required:"true" secret:"true"`
	RefreshTokenSecret string `yaml:"refresh_token_secret" env:"REFRESH_TOKEN_SECRET" env-required:"true" secret:"true"`
	JWTExpiryMinutes   int    `yaml:"jwt_expiry_minutes" env:"JWT_EXPIRY_MINUTES" env-default:"60"`

	// Rate limiting
	RateLimit int `yaml:"rate_limit" env:"RATE_LIMIT" env-default:"100"`

	// Database configuration
	DbType      DbType       `yaml:"db_type" env:"DB_TYPE" env-default:"postgres" omitempty:"true"`
	DbConfig    *DbConfig    `yaml:"database" omitempty:"true"`
	MongoConfig *MongoConfig `yaml:"mongodb" omitempty:"true"`

	// Logging
	LogConfig LogConfig `yaml:"logging" env-required:"true"`
}

func validatePort(port int, name string, checkAvailability bool) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid %s: %d — must be in range 1-65535", name, port)
	}

	// Check if port is already in use
	if !checkAvailability {
		return nil
	}

	address := fmt.Sprintf("localhost:%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("%s %d is already in use or unavailable: %v", name, port, err)
	}
	listener.Close()

	return nil
}

func (cfg *AppConfig) Validate() error {
	// Check if we're in a containerized environment
	checkPortAvailability := os.Getenv("SKIP_PORT_CHECK") != "true"

	if err := validatePort(cfg.AppPort, "APP_PORT", checkPortAvailability); err != nil {
		return err
	}

	if !cfg.DbType.IsValid() {
		return fmt.Errorf("invalid DB_TYPE: %s", cfg.DbType)
	}

	// Validate required secrets in production
	if cfg.Env == "production" {
		if cfg.JWTSecretKey == "secret" || len(cfg.JWTSecretKey) < 32 {
			return fmt.Errorf("JWT_SECRET_KEY must be at least 32 characters in production")
		}
		if cfg.RefreshTokenSecret == "refresh_secret" || len(cfg.RefreshTokenSecret) < 32 {
			return fmt.Errorf("REFRESH_TOKEN_SECRET must be at least 32 characters in production")
		}
	}

	// Validate database configuration
	switch cfg.DbType {
	case Postgres, MySQL, MsSQL, SQLite:
		if cfg.DbConfig == nil {
			return fmt.Errorf("DbConfig must not be nil for database type %s", cfg.DbType)
		}
		if err := validatePort(cfg.DbConfig.DbPort, "DB_PORT", checkPortAvailability); err != nil {
			return err
		}
	case MongoDb:
		if cfg.MongoConfig == nil {
			return fmt.Errorf("MongoConfig must not be nil for database type %s", cfg.DbType)
		}
		if err := validatePort(cfg.MongoConfig.Port, "MONGO_PORT", checkPortAvailability); err != nil {
			return err
		}
	}

	if cfg.LogConfig.EnableRotation && cfg.LogConfig.LogFile == "" {
		return fmt.Errorf("LOG_FILE must be set when ENABLE_ROTATION is true")
	}
	if cfg.LogConfig.RotationMaxSize <= 0 {
		return fmt.Errorf("ROTATION_MAX_SIZE must be greater than 0")
	}
	if cfg.LogConfig.LogFlushInterval <= 0 {
		return fmt.Errorf("LOG_FLUSH_INTERVAL must be greater than 0")
	}

	return nil
}

func LoadConfig(useDynamicEnv bool) *AppConfig {
	var config AppConfig

	if useDynamicEnv {
		if err := cleanenv.ReadEnv(&config); err != nil {
			log.Println("Error loading environment variables, using default values")
		}
	} else {
		if err := godotenv.Load(); err != nil {
			log.Println("Error loading .env file, using default values")
		}
		if err := cleanenv.ReadEnv(&config); err != nil {
			log.Println("Error reading environment variables into config, using default values")
		}
	}

	// Validate DbType
	if !config.DbType.IsValid() {
		log.Fatalf("Invalid DB_TYPE provided: %s. Supported: postgres, mysql, mssql, sqlite, mongodb", config.DbType)
	}

	// Load sub-configs based on DbType
	switch config.DbType {
	case Postgres, MySQL, MsSQL, SQLite:
		if config.DbConfig == nil {
			config.DbConfig = &DbConfig{}
		}
		if err := cleanenv.ReadEnv(config.DbConfig); err != nil {
			log.Println("Failed to override database config from env, using defaults")
		}
	case MongoDb:
		if config.MongoConfig == nil {
			config.MongoConfig = &MongoConfig{}
		}
		if err := cleanenv.ReadEnv(config.MongoConfig); err != nil {
			log.Println("Failed to override MongoDB config from env, using defaults")
		}
	}

	if err := config.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	return &config
}

func LogConfig(cfg any, prefix string) {
	val := reflect.ValueOf(cfg)
	typ := reflect.TypeOf(cfg)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		fieldVal := val.Field(i)
		fieldType := typ.Field(i)
		fieldName := fieldType.Name

		// Skip unexported fields
		if !fieldVal.CanInterface() {
			continue
		}

		isSecret := fieldType.Tag.Get("secret") == "true"

		// Recurse if nested struct or pointer to struct
		if fieldVal.Kind() == reflect.Struct {
			LogConfig(fieldVal.Interface(), prefix+fieldName+".")
		} else if fieldVal.Kind() == reflect.Ptr && fieldVal.Type().Elem().Kind() == reflect.Struct && !fieldVal.IsNil() {
			LogConfig(fieldVal.Interface(), prefix+fieldName+".")
		} else {
			if isSecret {
				log.Printf("%s%s: *****", prefix, fieldName)
			} else {
				log.Printf("%s%s: %v", prefix, fieldName, fieldVal.Interface())
			}
		}
	}
}
