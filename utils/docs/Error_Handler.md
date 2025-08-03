# Go HTTP Error Handler - Complete Documentation

A comprehensive Go package for handling HTTP errors with structured JSON responses, enhanced logging, and context support.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [API Reference](#api-reference)
- [Usage Examples](#usage-examples)
- [Best Practices](#best-practices)
- [Error Response Format](#error-response-format)
- [Logging Integration](#logging-integration)
- [Context Support](#context-support)
- [Migration Guide](#migration-guide)

## Overview

This package provides a robust and consistent way to handle HTTP errors in Go web applications. It automatically generates structured JSON error responses, provides comprehensive logging with request context, and supports both traditional and context-aware error handling patterns.

## Features

- **Structured JSON Responses**: Consistent error response format across your API
- **Enhanced Logging**: Automatic logging with request context, client IP, user agent, and more
- **Context Support**: Full support for Go's context package with request tracing
- **Flexible Details**: Support for additional error details and metadata
- **Client IP Detection**: Intelligent client IP extraction from various headers
- **Graceful Fallbacks**: Handles JSON encoding failures and logger initialization issues
- **Comprehensive Status Codes**: Support for all standard HTTP status codes
- **Backward Compatibility**: Non-breaking API with legacy function support

## Installation

```bash
go get your-module/utils
```

## Quick Start

### 1. Initialize the Logger

```go
package main

import (
    "your-module/utils"
    "go.uber.org/zap"
)

func main() {
    // Initialize with zap logger
    logger, _ := zap.NewProduction()
    utils.InitLogger(logger)
    
    // Or initialize with nil for no-op logging
    utils.InitLogger(nil)
}
```

### 2. Basic Error Handling

```go
func handler(w http.ResponseWriter, r *http.Request) {
    if someCondition {
        utils.HandleBadRequest(w, r, "Invalid input provided", nil)
        return
    }
    
    // Your handler logic here
}
```

### 3. Context-Aware Error Handling

```go
func handlerWithContext(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    if err := someOperation(ctx); err != nil {
        utils.HandleInternalErrorWithContext(ctx, w, r, err)
        return
    }
    
    // Your handler logic here
}
```

## API Reference

### Core Functions

#### `InitLogger(l *zap.Logger)`
Initializes the global logger instance. Pass `nil` to use a no-op logger.

#### `HandleErrorWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request, status int, message string, err error, details ...interface{})`
The main error handling function with full context support.

#### `HandleError(w http.ResponseWriter, r *http.Request, status int, message string, err error, details ...interface{})`
Legacy error handling function for backward compatibility.

### Specialized Error Handlers

#### Client Errors (4xx)
- `HandleBadRequest[WithContext]` - 400 Bad Request
- `HandleUnauthorized[WithContext]` - 401 Unauthorized
- `HandlePaymentRequired` - 402 Payment Required
- `HandleForbidden[WithContext]` - 403 Forbidden
- `HandleNotFound[WithContext]` - 404 Not Found
- `HandleMethodNotAllowed` - 405 Method Not Allowed
- `HandleRequestTimeout` - 408 Request Timeout
- `HandleConflict[WithContext]` - 409 Conflict
- `HandleUnprocessableEntity[WithContext]` - 422 Unprocessable Entity
- `HandleTooManyRequests[WithContext]` - 429 Too Many Requests

#### Server Errors (5xx)
- `HandleInternalError[WithContext]` - 500 Internal Server Error
- `HandleNotImplemented` - 501 Not Implemented
- `HandleBadGateway` - 502 Bad Gateway
- `HandleServiceUnavailable[WithContext]` - 503 Service Unavailable
- `HandleGatewayTimeout` - 504 Gateway Timeout

## Usage Examples

### Basic REST API Error Handling

```go
package main

import (
    "encoding/json"
    "net/http"
    "your-module/utils"
    "go.uber.org/zap"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
    // Initialize logger
    logger, _ := zap.NewProduction()
    utils.InitLogger(logger)
    
    http.HandleFunc("/users", handleUsers)
    http.HandleFunc("/users/", handleUserByID)
    http.ListenAndServe(":8080", nil)
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        getUsersList(w, r)
    case http.MethodPost:
        createUser(w, r)
    default:
        utils.HandleMethodNotAllowed(w, r, "Method not allowed", nil)
    }
}

func createUser(w http.ResponseWriter, r *http.Request) {
    var user User
    
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        utils.HandleBadRequest(w, r, "Invalid JSON payload", err, 
            map[string]string{"expected": "User object with id and name"})
        return
    }
    
    if user.Name == "" {
        utils.HandleUnprocessableEntity(w, r, "Name is required", nil,
            map[string]interface{}{
                "field": "name",
                "constraint": "non-empty string",
            })
        return
    }
    
    // Simulate database error
    if user.ID == 999 {
        utils.HandleInternalError(w, r, 
            fmt.Errorf("database connection failed"), 
            map[string]string{"operation": "user_create"})
        return
    }
    
    // Success response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

func getUsersList(w http.ResponseWriter, r *http.Request) {
    // Simulate rate limiting
    if r.Header.Get("X-Rate-Limit-Exceeded") == "true" {
        utils.HandleTooManyRequests(w, r, "Rate limit exceeded", nil,
            map[string]interface{}{
                "retry_after": "60s",
                "limit": 100,
                "window": "1h",
            })
        return
    }
    
    users := []User{{1, "Alice"}, {2, "Bob"}}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}
```

### Context-Aware Error Handling with Request Tracing

```go
package main

import (
    "context"
    "database/sql"
    "net/http"
    "time"
    "your-module/utils"
    "go.uber.org/zap"
    "github.com/google/uuid"
)

// Middleware to add request ID to context
func requestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := uuid.New().String()
        ctx := context.WithValue(r.Context(), "request_id", requestID)
        w.Header().Set("X-Request-ID", requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Timeout middleware
func timeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx, cancel := context.WithTimeout(r.Context(), timeout)
            defer cancel()
            
            done := make(chan struct{})
            go func() {
                defer close(done)
                next.ServeHTTP(w, r.WithContext(ctx))
            }()
            
            select {
            case <-done:
                // Request completed
            case <-ctx.Done():
                if ctx.Err() == context.DeadlineExceeded {
                    utils.HandleRequestTimeout(w, r, "Request timeout", ctx.Err(),
                        map[string]interface{}{
                            "timeout": timeout.String(),
                            "suggestion": "Try reducing request complexity",
                        })
                }
            }
        })
    }
}

func databaseHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Simulate database operation with context
    result, err := queryDatabaseWithContext(ctx)
    if err != nil {
        if err == sql.ErrNoRows {
            utils.HandleNotFoundWithContext(ctx, w, r, "Resource not found", err)
            return
        }
        
        utils.HandleInternalErrorWithContext(ctx, w, r, err,
            map[string]string{
                "operation": "database_query",
                "table": "users",
            })
        return
    }
    
    // Success response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

func queryDatabaseWithContext(ctx context.Context) (interface{}, error) {
    // Simulate database operation that respects context
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    case <-time.After(100 * time.Millisecond):
        return map[string]string{"result": "data"}, nil
    }
}

func main() {
    logger, _ := zap.NewProduction()
    utils.InitLogger(logger)
    
    mux := http.NewServeMux()
    mux.HandleFunc("/data", databaseHandler)
    
    // Apply middleware
    handler := requestIDMiddleware(timeoutMiddleware(5 * time.Second)(mux))
    
    http.ListenAndServe(":8080", handler)
}
```

### Authentication and Authorization

```go
func authHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    token := r.Header.Get("Authorization")
    if token == "" {
        utils.HandleUnauthorizedWithContext(ctx, w, r, 
            "Authentication required", nil,
            map[string]interface{}{
                "required_header": "Authorization",
                "format": "Bearer <token>",
                "docs": "https://api.example.com/docs/auth",
            })
        return
    }
    
    user, err := validateToken(ctx, token)
    if err != nil {
        utils.HandleUnauthorizedWithContext(ctx, w, r,
            "Invalid or expired token", err,
            map[string]string{
                "action": "obtain_new_token",
                "endpoint": "/auth/refresh",
            })
        return
    }
    
    if !user.HasPermission("read:data") {
        utils.HandleForbiddenWithContext(ctx, w, r,
            "Insufficient permissions", nil,
            map[string]interface{}{
                "required_permission": "read:data",
                "user_permissions": user.Permissions,
                "contact": "admin@example.com",
            })
        return
    }
    
    // Authorized - continue with handler logic
}
```

### Input Validation with Detailed Error Messages

```go
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

func validateCreateUserRequest(req CreateUserRequest) []string {
    var errors []string
    
    if req.Name == "" {
        errors = append(errors, "name is required")
    }
    if len(req.Name) > 100 {
        errors = append(errors, "name must be less than 100 characters")
    }
    if req.Email == "" {
        errors = append(errors, "email is required")
    }
    if !isValidEmail(req.Email) {
        errors = append(errors, "email format is invalid")
    }
    if req.Age < 0 || req.Age > 150 {
        errors = append(errors, "age must be between 0 and 150")
    }
    
    return errors
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.HandleBadRequestWithContext(ctx, w, r, 
            "Invalid JSON payload", err,
            map[string]interface{}{
                "example": CreateUserRequest{
                    Name:  "John Doe",
                    Email: "john@example.com",
                    Age:   30,
                },
                "content_type": "application/json",
            })
        return
    }
    
    if validationErrors := validateCreateUserRequest(req); len(validationErrors) > 0 {
        utils.HandleUnprocessableEntityWithContext(ctx, w, r,
            "Validation failed", nil,
            map[string]interface{}{
                "errors": validationErrors,
                "received": req,
            })
        return
    }
    
    // Process valid request...
}
```

### Service Integration with Circuit Breaker Pattern

```go
func externalServiceHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Check if external service is available
    if !isServiceHealthy("payment-service") {
        utils.HandleServiceUnavailableWithContext(ctx, w, r,
            "Payment service temporarily unavailable", nil,
            map[string]interface{}{
                "service": "payment-service",
                "retry_after": "30s",
                "status_page": "https://status.example.com",
            })
        return
    }
    
    result, err := callExternalService(ctx)
    if err != nil {
        // Determine error type and respond appropriately
        switch {
        case isTimeoutError(err):
            utils.HandleGatewayTimeout(w, r, 
                "External service timeout", err,
                map[string]interface{}{
                    "service": "payment-service",
                    "timeout": "10s",
                })
        case isServiceError(err):
            utils.HandleBadGateway(w, r,
                "External service error", err,
                map[string]string{
                    "service": "payment-service",
                    "incident_id": generateIncidentID(),
                })
        default:
            utils.HandleInternalErrorWithContext(ctx, w, r, err)
        }
        return
    }
    
    // Success response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

## Best Practices

### 1. Logger Initialization
Always initialize the logger early in your application lifecycle:

```go
func main() {
    // Production logger
    logger, err := zap.NewProduction()
    if err != nil {
        log.Fatal("Failed to initialize logger:", err)
    }
    defer logger.Sync()
    
    utils.InitLogger(logger)
    
    // Rest of your application...
}
```

### 2. Use Context-Aware Functions
Prefer context-aware functions when available:

```go
// Good
utils.HandleBadRequestWithContext(ctx, w, r, "Invalid input", err)

// Less preferred (but still supported)
utils.HandleBadRequest(w, r, "Invalid input", err)
```

### 3. Provide Meaningful Error Messages
Use clear, actionable error messages:

```go
// Good
utils.HandleBadRequest(w, r, "Email format is invalid", nil,
    map[string]string{
        "expected_format": "user@domain.com",
        "received": userInput.Email,
    })

// Less helpful
utils.HandleBadRequest(w, r, "Bad input", nil)
```

### 4. Include Relevant Details
Add context-specific details to help with debugging and user experience:

```go
utils.HandleTooManyRequests(w, r, "Rate limit exceeded", nil,
    map[string]interface{}{
        "limit": 100,
        "window": "1h",
        "retry_after": "3600s",
        "current_usage": userCurrentUsage,
    })
```

### 5. Handle Different Error Types Appropriately
Use specific error handlers for different scenarios:

```go
switch err {
case sql.ErrNoRows:
    utils.HandleNotFound(w, r, "User not found", err)
case context.DeadlineExceeded:
    utils.HandleRequestTimeout(w, r, "Operation timeout", err)
default:
    utils.HandleInternalError(w, r, err)
}
```

## Error Response Format

All error responses follow this consistent JSON structure:

```json
{
  "status": "error",
  "code": "BAD_REQUEST",
  "message": "Invalid input provided",
  "details": {
    "field": "email",
    "constraint": "valid email format required"
  },
  "timestamp": "2025-08-03T10:30:00Z"
}
```

### Fields Description

- **status**: Always "error" for error responses
- **code**: Machine-readable error code (e.g., "BAD_REQUEST", "NOT_FOUND")
- **message**: Human-readable error message
- **details**: Optional additional information (can be any JSON-serializable type)
- **timestamp**: ISO 8601 formatted timestamp in UTC

## Logging Integration

The package integrates seamlessly with Uber's Zap logger, providing structured logging with rich context:

### Log Fields Included

- **status**: HTTP status code
- **message**: Error message
- **method**: HTTP method
- **url**: Request URL path
- **client_ip**: Client IP address (extracted from headers)
- **user_agent**: User agent string
- **request_id**: Request ID from context (if available)
- **details**: Any additional details provided
- **error**: The actual error (if provided)

### Log Levels

- **Error**: Used when an actual error object is provided
- **Warn**: Used for HTTP errors without an underlying error object

### Example Log Output

```json
{
  "level": "error",
  "ts": 1691058600.123,
  "caller": "utils/errors.go:89",
  "msg": "HTTP error",
  "status": 400,
  "message": "Invalid email format",
  "method": "POST",
  "url": "/users",
  "client_ip": "192.168.1.100",
  "user_agent": "curl/7.68.0",
  "request_id": "req-123-456-789",
  "details": {
    "field": "email",
    "value": "invalid-email"
  },
  "error": "mail: missing '@' sign"
}
```

## Context Support

The package provides full support for Go's context package, enabling:

### Request Tracing
Use request IDs to trace requests across your application:

```go
ctx := context.WithValue(r.Context(), "request_id", "req-123")
utils.HandleErrorWithContext(ctx, w, r, 400, "Error message", err)
```

### Cancellation Handling
The package automatically detects context cancellation:

```go
ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
defer cancel()

// If context times out, it will be logged automatically
utils.HandleErrorWithContext(ctx, w, r, 500, "Operation failed", err)
```

### Custom Context Values
Add any custom values to the context for enhanced logging:

```go
ctx := context.WithValue(r.Context(), "user_id", userID)
ctx = context.WithValue(ctx, "operation", "user_update")
utils.HandleErrorWithContext(ctx, w, r, 403, "Access denied", nil)
```

## Migration Guide

### From Basic Error Handling

If you're currently using basic error handling:

```go
// Before
http.Error(w, "Bad Request", 400)

// After
utils.HandleBadRequest(w, r, "Bad Request", nil)
```

### Adding Context Support

To add context support to existing handlers:

```go
// Before
func handler(w http.ResponseWriter, r *http.Request) {
    utils.HandleBadRequest(w, r, "Error", err)
}

// After  
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    utils.HandleBadRequestWithContext(ctx, w, r, "Error", err)
}
```

### Adding Logger Support

```go
// Add to your main function
logger, _ := zap.NewProduction()
utils.InitLogger(logger)
```

The package is designed to be backward compatible, so you can migrate gradually without breaking existing functionality.