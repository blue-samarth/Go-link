# Success Response Handler

A comprehensive, production-ready success response handler for Go HTTP applications that provides consistent JSON responses with enhanced logging, context support, and graceful error handling.

## Features

- ✅ **Consistent JSON Structure** - Standardized success response format
- ✅ **Enhanced Logging** - Comprehensive request logging with client IP, user agent, and request tracking
- ✅ **Context Support** - Full context propagation for request IDs and cancellation handling  
- ✅ **Graceful Error Handling** - Fallback mechanisms for JSON encoding failures
- ✅ **Input Validation** - Message validation with sensible defaults
- ✅ **Flexible API** - Multiple convenience functions for common HTTP patterns
- ✅ **Production Ready** - Nil-safe operations, no panics, comprehensive error handling

## Installation

```go
import "your-module/response"
```

## Quick Start

### 1. Initialize Logger (Optional)

```go
logger, _ := zap.NewProduction()
response.InitSuccessLogger(logger)
```

### 2. Basic Usage

```go
func getUserHandler(w http.ResponseWriter, r *http.Request) {
    user := User{ID: 123, Name: "John Doe"}
    response.WriteOK(w, r, "User retrieved successfully", user)
}
```

### 3. With Context Support

```go
func createUserHandler(w http.ResponseWriter, r *http.Request) {
    ctx := context.WithValue(r.Context(), "request_id", "req-123")
    user := User{ID: 123, Name: "John Doe"}
    response.WriteCreatedWithContext(ctx, w, r, "User created successfully", user)
}
```

## Response Structure

All success responses follow this consistent JSON structure:

```json
{
  "status": "success",
  "message": "Operation completed successfully",
  "data": {...},
  "meta": {...},
  "timestamp": "2025-08-03T10:30:00Z"
}
```

### Fields

- **`status`** - Always "success" for successful responses
- **`message`** - Human-readable success message
- **`data`** *(optional)* - Response payload/data
- **`meta`** *(optional)* - Metadata (pagination, counts, etc.)
- **`timestamp`** - ISO 8601 UTC timestamp

## API Reference

### Core Functions

#### `WriteSuccessWithContext`
The primary function with full context support and enhanced logging.

```go
func WriteSuccessWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request, 
    statusCode int, message string, data interface{}, meta interface{})
```

#### `WriteSuccess`
Backward compatibility function without context.

```go
func WriteSuccess(w http.ResponseWriter, r *http.Request, 
    statusCode int, message string, data interface{}, meta interface{})
```

### HTTP Status Code Convenience Functions

#### With Context Support

```go
// 200 OK
WriteOKWithContext(ctx, w, r, "User retrieved", userData)

// 201 Created  
WriteCreatedWithContext(ctx, w, r, "User created", userData)

// 202 Accepted
WriteAcceptedWithContext(ctx, w, r, "Request accepted", nil)

// 204 No Content
WriteNoContentWithContext(ctx, w, r, "User deleted")

// 206 Partial Content
WritePartialContentWithContext(ctx, w, r, "Partial data", data, meta)
```

#### Without Context (Backward Compatible)

```go
WriteOK(w, r, "User retrieved", userData)
WriteCreated(w, r, "User created", userData)  
WriteAccepted(w, r, "Request accepted", nil)
WriteNoContent(w, r, "User deleted")
WritePartialContent(w, r, "Partial data", data, meta)
```

### Pagination Support

```go
pagination := map[string]interface{}{
    "page":  1,
    "limit": 10, 
    "total": 100,
    "pages": 10,
}

WriteOKWithPagination(w, r, "Users retrieved", users, pagination)
WriteOKWithPaginationContext(ctx, w, r, "Users retrieved", users, pagination)
```

### Resource-Specific Helpers

Convenient functions for common REST API patterns:

```go
// Create resource
WriteResourceCreated(w, r, "User", userData)
// → "User created successfully"

// Update resource  
WriteResourceUpdated(w, r, "User", userData)
// → "User updated successfully"

// Delete resource
WriteResourceDeleted(w, r, "User") 
// → "User deleted successfully"

// List resources
WriteResourceList(w, r, "Users", userList, paginationInfo)
// → "Users retrieved successfully"

// Get single resource
WriteResourceDetail(w, r, "User", userData)
// → "User retrieved successfully"
```

## Examples

### Basic CRUD Operations

```go
// GET /users/123
func getUser(w http.ResponseWriter, r *http.Request) {
    user := getUserFromDB(123)
    response.WriteResourceDetail(w, r, "User", user)
}

// POST /users
func createUser(w http.ResponseWriter, r *http.Request) {
    user := createUserInDB(userData)
    response.WriteResourceCreated(w, r, "User", user)
}

// PUT /users/123
func updateUser(w http.ResponseWriter, r *http.Request) {
    user := updateUserInDB(123, userData)
    response.WriteResourceUpdated(w, r, "User", user)
}

// DELETE /users/123
func deleteUser(w http.ResponseWriter, r *http.Request) {
    deleteUserFromDB(123)
    response.WriteResourceDeleted(w, r, "User")
}
```

### List with Pagination

```go
func listUsers(w http.ResponseWriter, r *http.Request) {
    users, total := getUsersFromDB(page, limit)
    
    pagination := map[string]interface{}{
        "page":       page,
        "limit":      limit,
        "total":      total,
        "total_pages": (total + limit - 1) / limit,
    }
    
    response.WriteResourceList(w, r, "Users", users, pagination)
}
```

### With Request Context

```go
func createUserWithTracking(w http.ResponseWriter, r *http.Request) {
    // Add request ID to context for tracking
    requestID := generateRequestID()
    ctx := context.WithValue(r.Context(), "request_id", requestID)
    
    user := createUserInDB(userData)
    response.WriteCreatedWithContext(ctx, w, r, "User created successfully", user)
}
```

### Custom Metadata

```go
func getUserWithStats(w http.ResponseWriter, r *http.Request) {
    user := getUserFromDB(123)
    
    meta := map[string]interface{}{
        "last_login":    user.LastLogin,
        "login_count":   user.LoginCount,
        "account_type":  user.AccountType,
    }
    
    response.WriteOKWithContext(ctx, w, r, "User retrieved with statistics", user, meta)
}
```

## Advanced Usage

### Custom Message Validation

The handler automatically validates messages and provides defaults:

```go
// Empty message gets default
response.WriteOK(w, r, "", userData) 
// → "Operation completed successfully"

// Whitespace is trimmed
response.WriteOK(w, r, "  User found  ", userData)
// → "User found"
```

### Error Handling

The handler gracefully handles JSON encoding failures:

```go
// If JSON encoding fails, falls back to plain text
response.WriteOK(w, r, "Success", cyclicalData) // would cause JSON error
// → Falls back to plain text: "Operation completed successfully"
```

### Logging Features

When a logger is configured, the handler logs:

- HTTP method and URL path
- Response status code and message  
- Client IP address (with proxy support)
- User agent
- Request ID (from context)
- Data type information
- Context cancellation status

```json
{
  "level": "info",
  "msg": "Success response sent",
  "status": 200,
  "message": "User created successfully", 
  "method": "POST",
  "url": "/api/users",
  "client_ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "request_id": "req-123",
  "data_type": "object"
}
```

## Configuration

### Logger Initialization

```go
// Production logger
logger, _ := zap.NewProduction()
response.InitSuccessLogger(logger)

// Development logger  
logger, _ := zap.NewDevelopment()
response.InitSuccessLogger(logger)

// Disable logging
response.InitSuccessLogger(nil) // Uses no-op logger
```

### Client IP Detection

The handler automatically detects client IPs through:

1. `X-Forwarded-For` header (proxy support)
2. `X-Real-IP` header  
3. `RemoteAddr` fallback

## JSON Response Examples

### Simple Success

```json
{
  "status": "success",
  "message": "User retrieved successfully",
  "data": {
    "id": 123,
    "name": "John Doe",
    "email": "john@example.com"
  },
  "timestamp": "2025-08-03T10:30:00Z"
}
```

### With Pagination

```json
{
  "status": "success", 
  "message": "Users retrieved successfully",
  "data": [
    {"id": 1, "name": "John"},
    {"id": 2, "name": "Jane"}
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 25,
    "total_pages": 3
  },
  "timestamp": "2025-08-03T10:30:00Z"
}
```

### No Content Response

```json
{
  "status": "success",
  "message": "User deleted successfully", 
  "timestamp": "2025-08-03T10:30:00Z"
}
```

## Best Practices

1. **Always use context-aware functions** in new code for better request tracking
2. **Initialize logger early** in your application startup
3. **Use resource-specific helpers** for consistent messaging  
4. **Include pagination metadata** for list endpoints
5. **Validate inputs** before calling response functions
6. **Use meaningful messages** that help API consumers

## Thread Safety

The success response handler is thread-safe. The global logger is set once during initialization and safely accessed by concurrent goroutines.

## Error Handling Philosophy

This handler follows a "never panic" philosophy:

- Nil loggers are handled gracefully with no-op loggers
- JSON encoding failures fall back to plain text
- Empty messages get sensible defaults
- All operations are nil-safe

## Integration with Error Handler

This success handler is designed to work alongside the companion error handler, sharing:

- Consistent logging patterns
- Same context support  
- Identical client IP detection
- Similar graceful failure handling
- Matching timestamp formats

## Contributing

When contributing, ensure:

- All functions are nil-safe
- Context support is maintained
- Logging is comprehensive but not verbose
- Backward compatibility is preserved
- Error handling is graceful
