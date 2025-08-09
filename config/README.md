# Go Configuration Management Package

A flexible and robust configuration management package for Go applications with support for multiple database types, environment variables, YAML files, and comprehensive validation.

## Overview

The `config` package provides a complete configuration solution that handles application settings, database configurations (SQL and NoSQL), security settings, and logging configuration. It automatically loads from environment variables or `.env` files with comprehensive validation and security features.

## Key Features

- **Multi-Database Support**: PostgreSQL, MySQL, MSSQL, SQLite, and MongoDB
- **Environment Variables**: Automatic loading from `.env` files or system environment  
- **YAML Configuration**: Support for YAML configuration files
- **Security**: Built-in secret masking and production validation
- **Validation**: Comprehensive configuration validation with port availability checking
- **Flexible Loading**: Dynamic environment loading or file-based configuration
- **Logging Config**: Built-in logging configuration with rotation and buffering support
- **Production Ready**: Enhanced validation and security for production environments

## Installation

```bash
go get github.com/joho/godotenv
go get github.com/ilyakaznacheev/cleanenv
```

## Quick Start

### 1. Basic Configuration Loading

```go
package main

import (
    "log"
    "your-project/config"
)

func main() {
    // Load configuration from .env file
    cfg := config.LoadConfig(false)
    
    // Or load directly from environment variables (for containers)
    // cfg := config.LoadConfig(true)
    
    // Log configuration safely (secrets will be masked)
    config.LogConfig(cfg, "")
    
    // Use your configuration
    log.Printf("Starting server on port %d", cfg.AppPort)
    log.Printf("Database type: %s", cfg.DbType)
}
```

### 2. Environment Variables Setup

Create a `.env` file:
```env
# Application
APP_PORT=8080
ENV=development

# Security (Required)
JWT_SECRET_KEY=your-super-secret-jwt-key-must-be-32-chars-minimum
REFRESH_TOKEN_SECRET=your-super-secret-refresh-token-32-chars-min

# Database
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=myuser
DB_PASS=mypassword
DB_NAME=myapp
SSL_MODE=disable

# Logging
LOG_LEVEL=info
LOG_MODE=dev
LOG_FILE=app.log
ENABLE_ROTATION=true
LOG_BUFFER_SIZE=1000
```

### 3. Using Different Database Types

```go
// PostgreSQL
os.Setenv("DB_TYPE", "postgres")
cfg := config.LoadConfig(true)
// Access: cfg.DbConfig.DbHost, cfg.DbConfig.DbPort, etc.

// MongoDB  
os.Setenv("DB_TYPE", "mongodb")
cfg := config.LoadConfig(true)
// Access: cfg.MongoConfig.Host, cfg.MongoConfig.Port, etc.
```

## Configuration Structure

### Main Configuration Types

#### `AppConfig` - Main application configuration
```go
type AppConfig struct {
    // Server configuration
    AppPort int    `yaml:"app_port" env:"APP_PORT" env-default:"8080"`
    Env     string `yaml:"env" env:"ENV" env-default:"development"`

    // Security configuration  
    JWTSecretKey       string `yaml:"jwt_secret_key" env:"JWT_SECRET_KEY" env-required:"true"`
    RefreshTokenSecret string `yaml:"refresh_token_secret" env:"REFRESH_TOKEN_SECRET" env-required:"true"`
    JWTExpiryMinutes   int    `yaml:"jwt_expiry_minutes" env:"JWT_EXPIRY_MINUTES" env-default:"60"`

    // Rate limiting
    RateLimit int `yaml:"rate_limit" env:"RATE_LIMIT" env-default:"100"`

    // Database configuration
    DbType      DbType       `yaml:"db_type" env:"DB_TYPE" env-default:"postgres"`
    DbConfig    *DbConfig    `yaml:"database"`
    MongoConfig *MongoConfig `yaml:"mongodb"`

    // Logging
    LogConfig LogConfig `yaml:"logging"`
}
```

#### `DbType` - Supported database types
- `postgres` - PostgreSQL
- `mysql` - MySQL  
- `mssql` - Microsoft SQL Server
- `sqlite` - SQLite
- `mongodb` - MongoDB

#### `DbConfig` - SQL database configuration
```go
type DbConfig struct {
    DbHost  string `yaml:"db_host" env:"DB_HOST" env-default:"localhost"`
    DbPort  int    `yaml:"db_port" env:"DB_PORT" env-default:"5432"`
    DbUser  string `yaml:"db_user" env:"DB_USER" env-default:"Admin"`
    DbPass  string `yaml:"db_pass" env:"DB_PASS" env-default:"Admin123"`
    DbName  string `yaml:"db_name" env:"DB_NAME" env-default:"dbname"`
    SSLMode string `yaml:"ssl_mode" env:"SSL_MODE" env-default:"disable"`
}
```

#### `MongoConfig` - MongoDB configuration
```go
type MongoConfig struct {
    Host       string `yaml:"host" env:"MONGO_HOST" env-default:"localhost"`
    Port       int    `yaml:"port" env:"MONGO_PORT" env-default:"27017"`
    User       string `yaml:"user" env:"MONGO_USER" env-default:"admin"`
    Pass       string `yaml:"pass" env:"MONGO_PASS" env-default:"changeme"`
    DbName     string `yaml:"db_name" env:"MONGO_DB_NAME" env-default:"app_db"`
    ReplicaSet string `yaml:"replica_set" env:"REPLICA_SET" env-default:""`
    ConnectionString string `yaml:"connection_string" env:"MONGO_CONNECTION_STRING" env-default:""`
}
```

#### `LogConfig` - Logging configuration
```go
type LogConfig struct {
    LogLevel         string `yaml:"log_level" env:"LOG_LEVEL" env-default:"info"`
    LogMode          string `yaml:"log_mode" env:"LOG_MODE" env-default:"dev"`
    LogFile          string `yaml:"log_file" env:"LOG_FILE" env-default:""`
    EnableRotation   bool   `yaml:"enable_rotation" env:"ENABLE_ROTATION" env-default:"false"`
    RotationMaxSize  int    `yaml:"rotation_max_size" env:"ROTATION_MAX_SIZE" env-default:"100"`
    RotationMaxBackups int  `yaml:"rotation_max_backups" env:"ROTATION_MAX_BACKUPS" env-default:"10"`
    RotationMaxAge   int    `yaml:"rotation_max_age" env:"ROTATION_MAX_AGE" env-default:"30"`
    RotationCompress bool   `yaml:"rotation_compress" env:"ROTATION_COMPRESS" env-default:"true"`
    LogBufferSize    int    `yaml:"log_buffer_size" env:"LOG_BUFFER_SIZE" env-default:"1000"`
    LogFlushInterval int    `yaml:"log_flush_interval" env:"LOG_FLUSH_INTERVAL" env-default:"5"`
}
```

## Environment Variables Reference

### Required Variables
```bash
# Security (Required in all environments)
JWT_SECRET_KEY=your-jwt-secret-key-at-least-32-chars
REFRESH_TOKEN_SECRET=your-refresh-token-secret-at-least-32-chars
```

### Application Variables
```bash
# Server Configuration
APP_PORT=8080                    # Server port (1-65535)
ENV=development                  # Environment: development, production

# Security
JWT_EXPIRY_MINUTES=60           # JWT token expiry in minutes
RATE_LIMIT=100                  # Rate limiting requests per minute

# Database Type Selection
DB_TYPE=postgres                 # postgres, mysql, mssql, sqlite, mongodb
```

### SQL Database Variables (PostgreSQL, MySQL, MSSQL, SQLite)
```bash
DB_HOST=localhost               # Database host
DB_PORT=5432                    # Database port
DB_USER=Admin                   # Database username  
DB_PASS=Admin123                # Database password
DB_NAME=dbname                  # Database name
SSL_MODE=disable                # SSL mode for connection
```

### MongoDB Variables
```bash  
MONGO_HOST=localhost            # MongoDB host
MONGO_PORT=27017                # MongoDB port
MONGO_USER=admin                # MongoDB username
MONGO_PASS=changeme             # MongoDB password
MONGO_DB_NAME=app_db            # MongoDB database name
REPLICA_SET=                    # MongoDB replica set (optional)
MONGO_CONNECTION_STRING=        # Full connection string (optional)
```

### Logging Variables
```bash
LOG_LEVEL=info                  # Log level: debug, info, warn, error
LOG_MODE=dev                    # Logging mode: dev, prod, minimal
LOG_FILE=                       # Log file path (empty = console)
ENABLE_ROTATION=false           # Enable log file rotation
ROTATION_MAX_SIZE=100           # Max log file size in MB
ROTATION_MAX_BACKUPS=10         # Max backup files to keep
ROTATION_MAX_AGE=30             # Max age of log files in days
ROTATION_COMPRESS=true          # Compress rotated files
LOG_BUFFER_SIZE=1000            # Buffer size for async logging
LOG_FLUSH_INTERVAL=5            # Flush interval in seconds
```

## API Reference

### Core Functions

#### `LoadConfig(useDynamicEnv bool) *AppConfig`

Loads and validates the complete application configuration.

**Parameters:**
- `useDynamicEnv`: If `true`, loads only from environment variables. If `false`, loads from `.env` file first, then environment variables.

**Returns:**
- `*AppConfig`: Fully validated configuration struct

**Example:**
```go
// Load from .env file + environment
config := config.LoadConfig(false)

// Load only from environment (containers)
config := config.LoadConfig(true)
```

#### `(cfg *AppConfig) Validate() error`

Validates the configuration struct for consistency and security requirements.

**Returns:**
- `error`: Returns validation error or nil if valid

**Example:**
```go
config := &AppConfig{...}
if err := config.Validate(); err != nil {
    log.Fatalf("Invalid configuration: %v", err)
}
```

#### `LogConfig(cfg any, prefix string)`

Safely logs configuration values with automatic secret masking.

**Parameters:**
- `cfg`: Configuration struct or interface to log
- `prefix`: String prefix for log entries

**Example:**
```go
// Log entire configuration
config.LogConfig(cfg, "")

// Log with prefix
config.LogConfig(cfg, "MyApp.")

// Log specific sub-config
config.LogConfig(cfg.DbConfig, "Database.")
```

#### `(d DbType) IsValid() bool`

Validates if a database type is supported.

**Returns:**
- `bool`: True if database type is valid

**Example:**
```go
dbType := config.Postgres
if dbType.IsValid() {
    // Use database type
}
```

## Usage Examples

### 1. Complete Application Setup

```go
package main

import (
    "log"
    "os"
    "your-project/config"
)

func main() {
    // Load configuration
    cfg := config.LoadConfig(false)
    
    // Log configuration safely
    config.LogConfig(cfg, "App.")
    
    // Use configuration throughout app
    startServer(cfg)
}

func startServer(cfg *config.AppConfig) {
    log.Printf("Starting server on port %d", cfg.AppPort)
    log.Printf("Environment: %s", cfg.Env)
    log.Printf("Database type: %s", cfg.DbType)
    
    // Setup database based on type
    switch cfg.DbType {
    case config.Postgres, config.MySQL, config.MsSQL, config.SQLite:
        setupSQLDatabase(cfg.DbConfig)
    case config.MongoDb:
        setupMongoDB(cfg.MongoConfig)
    }
}
```

### 2. Environment-Specific Configuration

```go
// Development environment
func setupDevelopment() *config.AppConfig {
    os.Setenv("ENV", "development")
    os.Setenv("LOG_LEVEL", "debug")
    os.Setenv("LOG_MODE", "dev")
    return config.LoadConfig(true)
}

// Production environment  
func setupProduction() *config.AppConfig {
    os.Setenv("ENV", "production")
    os.Setenv("LOG_LEVEL", "info")
    os.Setenv("LOG_MODE", "prod")
    os.Setenv("LOG_FILE", "/var/log/app.log")
    os.Setenv("ENABLE_ROTATION", "true")
    return config.LoadConfig(true)
}

// Container environment
func setupContainer() *config.AppConfig {
    os.Setenv("SKIP_PORT_CHECK", "true")
    return config.LoadConfig(true) // Load from environment only
}
```

### 3. Database Connection Examples

```go
// PostgreSQL connection
func connectPostgres(cfg *config.AppConfig) {
    if cfg.DbType == config.Postgres && cfg.DbConfig != nil {
        dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
            cfg.DbConfig.DbHost,
            cfg.DbConfig.DbPort,  
            cfg.DbConfig.DbUser,
            cfg.DbConfig.DbPass,
            cfg.DbConfig.DbName,
            cfg.DbConfig.SSLMode,
        )
        // Use dsn to connect...
    }
}

// MongoDB connection
func connectMongo(cfg *config.AppConfig) {
    if cfg.DbType == config.MongoDb && cfg.MongoConfig != nil {
        // Use connection string if provided
        if cfg.MongoConfig.ConnectionString != "" {
            // Connect using connection string
        } else {
            // Build connection from individual fields
            uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
                cfg.MongoConfig.User,
                cfg.MongoConfig.Pass,
                cfg.MongoConfig.Host,
                cfg.MongoConfig.Port,
                cfg.MongoConfig.DbName,
            )
            // Use uri to connect...
        }
    }
}
```

### 4. Configuration Integration with Logger

```go
func setupApplicationLogging(cfg *config.AppConfig) {
    // Convert config.LogConfig to logger format
    loggerConfig := utils.LoggerConfig{
        Level:         cfg.LogConfig.LogLevel,
        Mode:          cfg.LogConfig.LogMode,
        FilePath:      cfg.LogConfig.LogFile,
        EnableRotate:  cfg.LogConfig.EnableRotation,
        MaxSizeMB:     cfg.LogConfig.RotationMaxSize,
        MaxBackups:    cfg.LogConfig.RotationMaxBackups,
        MaxAgeDays:    cfg.LogConfig.RotationMaxAge,
        Compress:      cfg.LogConfig.RotationCompress,
        BufferSize:    cfg.LogConfig.LogBufferSize,
        FlushInterval: time.Duration(cfg.LogConfig.LogFlushInterval) * time.Second,
    }
    
    if err := utils.InitLogger(loggerConfig); err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
    }
    
    defer utils.ShutdownLogger()
}
```

### 5. Multiple Database Type Support

```go
func setupDatabase(cfg *config.AppConfig) interface{} {
    switch cfg.DbType {
    case config.Postgres:
        return setupPostgreSQL(cfg.DbConfig)
    case config.MySQL:
        return setupMySQL(cfg.DbConfig)
    case config.MsSQL:
        return setupMSSQL(cfg.DbConfig)
    case config.SQLite:
        return setupSQLite(cfg.DbConfig)
    case config.MongoDb:
        return setupMongoDB(cfg.MongoConfig)
    default:
        log.Fatalf("Unsupported database type: %s", cfg.DbType)
        return nil
    }
}
```

## Security Features

### Secret Protection
- Automatic secret masking in logs using `secret:"true"` struct tags
- Passwords, tokens, and keys are never displayed in plain text
- Safe configuration logging with `LogConfig()` function

### Production Security Enforcement
- Enhanced validation in production environment
- Minimum length requirements for JWT secrets
- Prevention of default/weak secret values

### Configuration Security Best Practices
```go
// Fields marked as secret are automatically masked
type SecureConfig struct {
    PublicField  string `yaml:"public" env:"PUBLIC"`
    SecretField  string `yaml:"secret" env:"SECRET" secret:"true"`  // Will show as *****
}
```

## Troubleshooting

### Common Issues and Solutions

#### Port Already in Use
```bash
Error: APP_PORT 8080 is already in use or unavailable
```
**Solutions:**
- Change the port: `export APP_PORT=8081`
- Skip port check in containers: `export SKIP_PORT_CHECK=true`
- Find and stop the process using the port

#### Invalid Database Type
```bash
Error: Invalid DB_TYPE provided: mysql2. Supported: postgres, mysql, mssql, sqlite, mongodb
```
**Solution:**
```bash
# Use exact supported values
export DB_TYPE=mysql  # not mysql2
```

#### Missing Required Secrets
```bash
Error: JWT_SECRET_KEY must be at least 32 characters in production
```
**Solution:**
```bash
# Set proper length secrets
export JWT_SECRET_KEY="your-32-character-secret-key-here"
export REFRESH_TOKEN_SECRET="your-32-character-refresh-secret"
```

#### Configuration Not Loading
```bash
Error: Error loading .env file, using default values
```
**Solutions:**
- Ensure `.env` file exists in working directory
- Check file permissions are readable
- Use `LoadConfig(true)` to load from environment only
- Verify environment variable names match exactly

#### Database Configuration Missing
```bash
Error: DbConfig must not be nil for database type postgres
```
**Solution:**
```bash
# Set database environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=myuser
export DB_PASS=mypass
export DB_NAME=mydb
```

#### Log Rotation Error
```bash
Error: LOG_FILE must be set when ENABLE_ROTATION is true
```
**Solution:**
```bash
export LOG_FILE=/var/log/app.log
export ENABLE_ROTATION=true
```

### Debugging Configuration

#### Enable Debug Output
```go
// Add debug logging to see what's loaded
config.LogConfig(cfg, "Debug.")
```

#### Validate Configuration Manually
```go
cfg := &config.AppConfig{...}
if err := cfg.Validate(); err != nil {
    fmt.Printf("Validation error: %v\n", err)
}
```

#### Check Environment Variables
```bash
# List all environment variables
env | grep -E "(APP_|DB_|MONGO_|JWT_|LOG_)"

# Check specific variable
echo $DB_TYPE
echo $JWT_SECRET_KEY
```

### Environment-Specific Troubleshooting

#### Development Environment
```bash
# Ensure .env file is properly formatted
cat .env

# Check file permissions
ls -la .env

# Test with different loading method
export USE_DYNAMIC_ENV=true
```

#### Production Environment  
```bash
# Verify all required variables are set
env | grep JWT_SECRET_KEY
env | grep REFRESH_TOKEN_SECRET

# Check production validation
export ENV=production
```

#### Container Environment
```bash
# Skip port checks in containers
export SKIP_PORT_CHECK=true

# Use environment-only loading
# Don't rely on .env files in containers
```

### Getting Help

If you encounter issues not covered here:

1. **Check the validation error messages** - they provide specific guidance
2. **Verify environment variable names** match the struct tags exactly  
3. **Test with minimal configuration** to isolate issues
4. **Enable debug logging** to see configuration loading process
5. **Check file permissions** for .env files and log files Errors**: 