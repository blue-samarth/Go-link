# Go Configuration Management

A flexible and robust configuration management package for Go applications with support for multiple database types, environment variables, YAML files, and comprehensive validation.

## Features

- 🔧 **Multi-Database Support**: PostgreSQL, MySQL, MSSQL, SQLite, and MongoDB
- 🌍 **Environment Variables**: Automatic loading from `.env` files or system environment
- 📝 **YAML Configuration**: Support for YAML configuration files
- 🔒 **Security**: Built-in secret masking and production validation
- ✅ **Validation**: Comprehensive configuration validation with port availability checking
- 🚀 **Flexible Loading**: Dynamic environment loading or file-based configuration
- 📊 **Logging**: Safe configuration logging with secret redaction

## Installation

```bash
go get github.com/joho/godotenv
go get github.com/ilyakaznacheev/cleanenv
```

## Quick Start

### Basic Usage

```go
package main

import (
    "log"
    // Import your config package
)

func main() {
    // Load configuration from .env file
    config := LoadConfig(false)
    
    // Or load directly from environment variables
    // config := LoadConfig(true)
    
    // Log configuration (secrets will be masked)
    LogConfig(config, "")
    
    // Use your configuration
    log.Printf("Starting server on port %d", config.AppPort)
}
```

## Configuration Structure

### Application Configuration

The main `AppConfig` struct contains all application settings:

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
    LogLevel string `yaml:"log_level" env:"LOG_LEVEL" env-default:"info"`
}
```

### Supported Database Types

The configuration supports the following database types:

- `postgres` - PostgreSQL
- `mysql` - MySQL
- `mssql` - Microsoft SQL Server
- `sqlite` - SQLite
- `mongodb` - MongoDB

## Environment Variables

### Required Variables

These variables are required and must be set:

```bash
JWT_SECRET_KEY=your-jwt-secret-key-at-least-32-chars
REFRESH_TOKEN_SECRET=your-refresh-token-secret-at-least-32-chars
```

### Application Variables

```bash
# Server Configuration
APP_PORT=8080
ENV=development  # or production

# Security
JWT_EXPIRY_MINUTES=60

# Rate Limiting
RATE_LIMIT=100

# Logging
LOG_LEVEL=info  # debug, info, warn, error

# Database Type
DB_TYPE=postgres  # postgres, mysql, mssql, sqlite, mongodb
```

### SQL Database Configuration (PostgreSQL, MySQL, MSSQL, SQLite)

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=Admin
DB_PASS=Admin123
DB_NAME=dbname
SSL_MODE=disable
```

### MongoDB Configuration

```bash
MONGO_HOST=localhost
MONGO_PORT=27017
MONGO_USER=admin
MONGO_PASS=changeme
MONGO_DB_NAME=app_db
REPLICA_SET=
MONGO_CONNECTION_STRING=  # Optional: Use instead of individual fields
```

## Configuration Files

### .env File Example

Create a `.env` file in your project root:

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

# Optional
RATE_LIMIT=100
LOG_LEVEL=info
JWT_EXPIRY_MINUTES=60
```

### YAML Configuration Example

Create a `config.yaml` file:

```yaml
app_port: 8080
env: "development"
jwt_secret_key: "your-super-secret-jwt-key-must-be-32-chars-minimum"
refresh_token_secret: "your-super-secret-refresh-token-32-chars-min"
jwt_expiry_minutes: 60
rate_limit: 100
log_level: "info"
db_type: "postgres"

database:
  db_host: "localhost"
  db_port: 5432
  db_user: "myuser"
  db_pass: "mypassword"
  db_name: "myapp"
  ssl_mode: "disable"

# For MongoDB
# mongodb:
#   host: "localhost"
#   port: 27017
#   user: "admin"
#   pass: "changeme"
#   db_name: "app_db"
#   replica_set: ""
```

## Usage Examples

### Loading Configuration

```go
// Method 1: Load from .env file
config := LoadConfig(false)

// Method 2: Load directly from environment (useful for containers)
config := LoadConfig(true)
```

### Using Different Database Types

```go
// PostgreSQL
os.Setenv("DB_TYPE", "postgres")
config := LoadConfig(true)
// Access: config.DbConfig.DbHost, config.DbConfig.DbPort, etc.

// MongoDB
os.Setenv("DB_TYPE", "mongodb")
config := LoadConfig(true)
// Access: config.MongoConfig.Host, config.MongoConfig.Port, etc.
```

### Production vs Development

```go
// Development
config := LoadConfig(false)
if config.Env == "development" {
    // Development-specific logic
    log.Println("Running in development mode")
}

// Production
os.Setenv("ENV", "production")
config := LoadConfig(true)
// Production validation will enforce stronger security requirements
```

## Validation

The package includes comprehensive validation:

### Port Validation
- Validates port ranges (1-65535)
- Checks port availability (can be skipped in containers)
- Set `SKIP_PORT_CHECK=true` to disable port availability checking

### Production Security
In production environment (`ENV=production`):
- JWT secrets must be at least 32 characters
- Stronger validation for security-related configurations

### Database Validation
- Ensures correct database configuration based on selected type
- Validates required fields for each database type

## Logging

### Safe Configuration Logging

```go
// This will log all configuration values with secrets masked
LogConfig(config, "")

// Output example:
// AppPort: 8080
// Env: development
// JWTSecretKey: *****
// RefreshTokenSecret: *****
// DbConfig.DbHost: localhost
// DbConfig.DbPass: *****
```

### Custom Logging

```go
// Log with custom prefix
LogConfig(config, "MyApp.")

// Log specific structs
LogConfig(config.DbConfig, "Database.")
```

## Docker/Container Support

For containerized environments:

```dockerfile
# Skip port availability checks in containers
ENV SKIP_PORT_CHECK=true

# Use dynamic environment loading
ENV ENV=production
```

```go
// In your application
config := LoadConfig(true) // Load directly from environment
```

## Error Handling

The package provides detailed error messages for common issues:

```go
config := LoadConfig(false)
// If validation fails, the application will log the error and exit

// Common errors:
// - Invalid port ranges
// - Missing required secrets in production
// - Invalid database type
// - Port already in use
// - Missing database configuration
```

## Best Practices

### Development
1. Use `.env` files for local development
2. Keep secrets out of version control
3. Use default values for non-critical settings

### Production
1. Set `ENV=production`
2. Use strong, unique secrets (32+ characters)
3. Load configuration from environment variables
4. Set `SKIP_PORT_CHECK=true` in containers
5. Use proper logging levels

### Security
1. Always use the `secret:"true"` tag for sensitive fields
2. Use strong JWT secrets in production
3. Rotate secrets regularly
4. Use environment variables for secrets, not files

## Troubleshooting

### Common Issues

**Port already in use:**
```bash
# Set this environment variable to skip port checks
export SKIP_PORT_CHECK=true
```

**Invalid database type:**
```bash
# Ensure DB_TYPE is one of: postgres, mysql, mssql, sqlite, mongodb
export DB_TYPE=postgres
```

**Missing required secrets:**
```bash
# Set required JWT secrets
export JWT_SECRET_KEY="your-32-character-secret-key-here"
export REFRESH_TOKEN_SECRET="your-32-character-refresh-secret"
```

**Configuration not loading:**
```go
// Check if .env file exists and is readable
// Use LoadConfig(true) to load directly from environment
// Check environment variable names match exactly
```

## Contributing

1. Ensure all new configuration fields include appropriate struct tags
2. Add validation for new fields in the `Validate()` method
3. Update documentation and examples
4. Test with different database types and environments
