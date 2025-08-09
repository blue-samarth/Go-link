# Go Logger Package Documentation

## Overview

The `utils` logger package provides a high-performance, thread-safe logging solution for Go applications. It features asynchronous logging with buffering, file rotation, colorized console output, and graceful shutdown handling.

## Key Features

- **Asynchronous Logging**: Non-blocking log writes with background processing
- **File Rotation**: Automatic log file rotation based on size, age, and backup count
- **Colorized Output**: Color-coded log levels for console output
- **Thread-Safe**: Safe for concurrent use across goroutines
- **Graceful Shutdown**: Ensures all logs are flushed before application exit
- **Configurable**: Highly customizable through configuration struct

## Quick Start

### 1. Basic Setup

```go
package main

import (
    "time"
    "your-project/utils"
    "go.uber.org/zap/zapcore"
)

func main() {
    // Initialize logger with basic configuration
    config := utils.LoggerConfig{
        Level:         "info",
        Mode:          "production",
        BufferSize:    1000,
        FlushInterval: 5 * time.Second,
    }
    
    if err := utils.InitLogger(config); err != nil {
        panic(err)
    }
    
    // Ensure proper cleanup
    defer utils.ShutdownLogger()
    
    // Your application code here
    utils.Log(zapcore.InfoLevel, "Application started")
}
```

### 2. Console Logging (Development)

```go
config := utils.LoggerConfig{
    Level:         "debug",
    Mode:          "dev",          // Enables development mode
    BufferSize:    500,
    FlushInterval: 2 * time.Second,
}
```

### 3. File Logging with Rotation

```go
config := utils.LoggerConfig{
    Level:         "info",
    Mode:          "production",
    FilePath:      "/var/log/myapp.log",
    EnableRotate:  true,
    MaxSizeMB:     100,           // Rotate when file reaches 100MB
    MaxBackups:    5,             // Keep 5 backup files
    MaxAgeDays:    30,            // Delete files older than 30 days
    Compress:      true,          // Compress rotated files
    BufferSize:    2000,
    FlushInterval: 10 * time.Second,
}
```

## Configuration Options

### LoggerConfig Struct

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `Level` | `string` | Log level: "debug", "info", "warn", "error", "fatal" | "info" |
| `Mode` | `string` | Logger mode: "dev" or "production" | "production" |
| `FilePath` | `string` | File path for logging. If empty, logs to console | "" (console) |
| `EnableRotate` | `bool` | Enable file rotation (only for file logging) | `false` |
| `MaxSizeMB` | `int` | Maximum file size in MB before rotation | 100 |
| `MaxBackups` | `int` | Maximum number of backup files to keep | 3 |
| `MaxAgeDays` | `int` | Maximum age of log files in days | 28 |
| `Compress` | `bool` | Compress rotated log files | `false` |
| `BufferSize` | `int` | Size of the log buffer for async processing | 1000 |
| `FlushInterval` | `time.Duration` | Interval for flushing buffered logs | 5s |

## API Reference

### Initialization Functions

#### `InitLogger(cfg LoggerConfig) error`

Initializes the global logger with the provided configuration. This function is thread-safe and uses a singleton pattern - multiple calls will only initialize once.

**Parameters:**
- `cfg`: Logger configuration struct

**Returns:**
- `error`: Returns error if initialization fails

**Example:**
```go
config := utils.LoggerConfig{
    Level: "info",
    Mode:  "production",
}
if err := utils.InitLogger(config); err != nil {
    log.Fatal("Failed to initialize logger:", err)
}
```

### Logging Functions

#### `Log(level zapcore.Level, msg string, fields ...zap.Field)`

Asynchronously logs a message at the specified level.

**Parameters:**
- `level`: Log level (zapcore.DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel)
- `msg`: Log message string
- `fields`: Optional structured fields (zap.Field)

**Example:**
```go
import "go.uber.org/zap"

// Simple message
utils.Log(zapcore.InfoLevel, "User logged in")

// With structured fields
utils.Log(zapcore.InfoLevel, "User action", 
    zap.String("user_id", "12345"),
    zap.String("action", "login"),
    zap.Duration("response_time", 150*time.Millisecond),
)
```

#### `Logger() *zap.Logger`

Returns the underlying zap.Logger instance for direct use. This provides access to all zap logging methods.

**Returns:**
- `*zap.Logger`: The global logger instance

**Example:**
```go
logger := utils.Logger()
logger.Info("Direct zap usage",
    zap.String("component", "auth"),
    zap.Int("user_count", 42),
)
```

### Cleanup Functions

#### `ShutdownLogger()`

Gracefully shuts down the logger, ensuring all buffered logs are flushed. This function is idempotent - multiple calls are safe.

**Example:**
```go
defer utils.ShutdownLogger()
```

## Usage Patterns

### 1. Application Startup Pattern

```go
func main() {
    // Initialize logger first
    config := utils.LoggerConfig{
        Level:         "info",
        FilePath:      "app.log",
        EnableRotate:  true,
        MaxSizeMB:     50,
        BufferSize:    1000,
        FlushInterval: 5 * time.Second,
    }
    
    if err := utils.InitLogger(config); err != nil {
        fmt.Fprintf(os.Stderr, "Logger init failed: %v\n", err)
        os.Exit(1)
    }
    
    // Ensure cleanup on exit
    defer utils.ShutdownLogger()
    
    // Log application start
    utils.Log(zapcore.InfoLevel, "Application starting")
    
    // Your application logic
    runApplication()
    
    utils.Log(zapcore.InfoLevel, "Application shutting down")
}
```

### 2. Different Log Levels

```go
import "go.uber.org/zap/zapcore"

// Debug information (only shown in debug level)
utils.Log(zapcore.DebugLevel, "Processing request details")

// General information
utils.Log(zapcore.InfoLevel, "Server started on port 8080")

// Warning conditions
utils.Log(zapcore.WarnLevel, "High memory usage detected")

// Error conditions
utils.Log(zapcore.ErrorLevel, "Database connection failed")

// Critical errors
utils.Log(zapcore.FatalLevel, "Unable to start server")
```

### 3. Structured Logging

```go
import "go.uber.org/zap"

// HTTP request logging
utils.Log(zapcore.InfoLevel, "HTTP request processed",
    zap.String("method", "GET"),
    zap.String("path", "/api/users"),
    zap.Int("status_code", 200),
    zap.Duration("duration", 45*time.Millisecond),
    zap.String("user_agent", "Mozilla/5.0..."),
)

// Database operation
utils.Log(zapcore.InfoLevel, "Database query executed",
    zap.String("table", "users"),
    zap.String("operation", "SELECT"),
    zap.Int("rows_affected", 15),
    zap.Duration("query_time", 12*time.Millisecond),
)

// Error with context
utils.Log(zapcore.ErrorLevel, "Payment processing failed",
    zap.String("payment_id", "pay_123456"),
    zap.Float64("amount", 99.99),
    zap.String("currency", "USD"),
    zap.String("error", "insufficient_funds"),
)
```

### 4. Using Direct Zap Logger

```go
logger := utils.Logger()

// All zap methods are available
logger.Debug("Debug message")
logger.Info("Info message")
logger.Warn("Warning message")
logger.Error("Error message")

// With fields
logger.Info("User created",
    zap.String("username", "johndoe"),
    zap.String("email", "john@example.com"),
    zap.Time("created_at", time.Now()),
)

// Sugar logger for printf-style logging
sugar := logger.Sugar()
sugar.Infof("Processing %d items", itemCount)
sugar.Errorw("Failed to process item",
    "item_id", itemID,
    "error", err,
)
```

## Configuration Examples

### Development Environment

```go
config := utils.LoggerConfig{
    Level:         "debug",        // Show all log levels
    Mode:          "dev",          // Development formatting
    BufferSize:    500,            // Smaller buffer for quick feedback
    FlushInterval: 1 * time.Second, // Frequent flushes
}
```

### Production Environment

```go
config := utils.LoggerConfig{
    Level:         "info",         // Production level
    Mode:          "production",   // Structured JSON output
    FilePath:      "/var/log/app.log",
    EnableRotate:  true,
    MaxSizeMB:     200,            // Larger files for production
    MaxBackups:    10,             // More backups
    MaxAgeDays:    90,             // Longer retention
    Compress:      true,           // Save disk space
    BufferSize:    5000,           // Larger buffer for performance
    FlushInterval: 30 * time.Second, // Less frequent flushes
}
```

### High-Performance Setup

```go
config := utils.LoggerConfig{
    Level:         "warn",         // Only warnings and errors
    Mode:          "production",
    FilePath:      "/var/log/app.log",
    EnableRotate:  true,
    MaxSizeMB:     500,            // Very large files
    BufferSize:    10000,          // Large buffer
    FlushInterval: 60 * time.Second, // Minimal I/O interruption
}
```

## Best Practices

1. **Initialize Once**: Call `InitLogger()` only once at application startup
2. **Always Defer Cleanup**: Use `defer utils.ShutdownLogger()` to ensure logs are flushed
3. **Use Structured Logging**: Prefer structured fields over string formatting
4. **Choose Appropriate Levels**: Use debug for development, info for normal operations
5. **Configure Buffer Size**: Larger buffers for high-throughput applications
6. **Monitor Disk Space**: Enable rotation in production to prevent disk filling
7. **Handle Initialization Errors**: Always check for errors from `InitLogger()`

## Signal Handling

The logger automatically handles SIGTERM and SIGINT signals to ensure graceful shutdown and log flushing. No additional setup required.

## Thread Safety

All functions in this package are thread-safe and can be called concurrently from multiple goroutines without additional synchronization.