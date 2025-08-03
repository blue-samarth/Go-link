# Pagination Package Documentation

A comprehensive Go package for handling pagination with both SQL (GORM) and MongoDB databases. This package provides type-safe, generic pagination functions with automatic parameter validation and consistent response formatting.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Types](#types)
- [Functions](#functions)
- [Usage Examples](#usage-examples)
- [API Reference](#api-reference)
- [Best Practices](#best-practices)
- [Error Handling](#error-handling)

## Features

- ✅ **Generic Type Support**: Works with any struct type using Go generics
- ✅ **Dual Database Support**: Compatible with both GORM (SQL) and MongoDB
- ✅ **Automatic Validation**: Built-in parameter normalization with sensible defaults
- ✅ **Context Support**: Proper context handling for timeouts and cancellations
- ✅ **Flexible Options**: Support for custom MongoDB find options
- ✅ **Consistent Response**: Unified response format across both databases
- ✅ **Production Ready**: Includes limits and error handling for production use

## Installation

```bash
go get -u gorm.io/gorm
go get -u go.mongodb.org/mongo-driver/mongo
```

## Types

### PaginationParams

```go
type PaginationParams struct {
    Page  int `form:"page" json:"page"`   // Page number (starts from 1)
    Limit int `form:"limit" json:"limit"` // Items per page (max 100)
}
```

**Validation Rules:**
- `Page`: Defaults to 1 if <= 0
- `Limit`: Defaults to 10 if <= 0 or > 100

### PaginationMeta

```go
type PaginationMeta struct {
    CurrentPage int   `json:"currentPage"` // Current page number
    TotalPages  int   `json:"totalPages"`  // Total number of pages
    PageSize    int   `json:"pageSize"`    // Items per page
    TotalItems  int64 `json:"totalItems"`  // Total number of items
}
```

### PaginatedResponse

```go
type PaginatedResponse[T any] struct {
    Message    string         `json:"message"`    // Response message
    Data       []T            `json:"data"`       // Paginated data
    Pagination PaginationMeta `json:"pagination"` // Pagination metadata
}
```

## Functions

### PaginateSQL

Paginates GORM queries with automatic parameter validation.

```go
func PaginateSQL[T any](ctx context.Context, db *gorm.DB, params PaginationParams) (PaginatedResponse[T], error)
```

**Parameters:**
- `ctx`: Context for request lifecycle management
- `db`: GORM database instance with your query conditions
- `params`: Pagination parameters

**Returns:**
- `PaginatedResponse[T]`: Paginated response with data and metadata
- `error`: Error if pagination fails

### PaginateMongo

Paginates MongoDB queries with support for custom find options.

```go
func PaginateMongo[T any](ctx context.Context, collection *mongo.Collection, filter interface{}, params PaginationParams, opts ...*options.FindOptions) (PaginatedResponse[T], error)
```

**Parameters:**
- `ctx`: Context for request lifecycle management
- `collection`: MongoDB collection
- `filter`: MongoDB filter (bson.M, bson.D, etc.)
- `params`: Pagination parameters
- `opts`: Optional MongoDB find options (sorting, projection, etc.)

**Returns:**
- `PaginatedResponse[T]`: Paginated response with data and metadata
- `error`: Error if pagination fails

## Usage Examples

### Basic SQL Pagination

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "your-project/pagination"
    "gorm.io/gorm"
)

type User struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"createdAt"`
}

func GetUsers(db *gorm.DB, page, limit int) {
    ctx := context.Background()
    
    // Prepare pagination parameters
    params := pagination.PaginationParams{
        Page:  page,
        Limit: limit,
    }
    
    // Create base query
    query := db.Model(&User{})
    
    // Add your filters/conditions
    query = query.Where("created_at > ?", time.Now().AddDate(0, -1, 0))
    
    // Apply pagination
    response, err := pagination.PaginateSQL[User](ctx, query, params)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    // Use the response
    fmt.Printf("Message: %s\n", response.Message)
    fmt.Printf("Total Users: %d\n", response.Pagination.TotalItems)
    fmt.Printf("Current Page: %d/%d\n", 
        response.Pagination.CurrentPage, 
        response.Pagination.TotalPages)
    
    for _, user := range response.Data {
        fmt.Printf("User: %s (%s)\n", user.Name, user.Email)
    }
}
```

### Advanced SQL Pagination with Complex Queries

```go
func GetActiveUsers(db *gorm.DB, params pagination.PaginationParams) (pagination.PaginatedResponse[User], error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Complex query with joins and conditions
    query := db.Model(&User{}).
        Joins("LEFT JOIN user_profiles ON users.id = user_profiles.user_id").
        Where("users.status = ?", "active").
        Where("users.email_verified = ?", true).
        Where("user_profiles.subscription_active = ?", true).
        Order("users.created_at DESC")
    
    return pagination.PaginateSQL[User](ctx, query, params)
}
```

### Basic MongoDB Pagination

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "your-project/pagination"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type Product struct {
    ID          string    `bson:"_id,omitempty" json:"id,omitempty"`
    Name        string    `bson:"name" json:"name"`
    Price       int       `bson:"price" json:"price"`
    Category    string    `bson:"category" json:"category"`
    InStock     bool      `bson:"inStock" json:"inStock"`
    CreatedAt   time.Time `bson:"createdAt" json:"createdAt"`
}

func GetProducts(collection *mongo.Collection, page, limit int) {
    ctx := context.Background()
    
    params := pagination.PaginationParams{
        Page:  page,
        Limit: limit,
    }
    
    // Simple filter
    filter := bson.M{"inStock": true}
    
    response, err := pagination.PaginateMongo[Product](ctx, collection, filter, params)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Found %d products\n", response.Pagination.TotalItems)
    for _, product := range response.Data {
        fmt.Printf("Product: %s - $%d\n", product.Name, product.Price)
    }
}
```

### Advanced MongoDB Pagination with Sorting and Projection

```go
func GetProductsByCategory(collection *mongo.Collection, category string, params pagination.PaginationParams) (pagination.PaginatedResponse[Product], error) {
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    
    // Complex filter
    filter := bson.M{
        "category": category,
        "inStock":  true,
        "price":    bson.M{"$gte": 10}, // Price >= $10
    }
    
    // Find options with sorting and projection
    opts := options.Find().
        SetSort(bson.D{{"price", 1}, {"name", 1}}). // Sort by price asc, then name asc
        SetProjection(bson.M{
            "name":      1,
            "price":     1,
            "category":  1,
            "createdAt": 1,
            // Exclude some fields for performance
            "description": 0,
            "images":      0,
        })
    
    return pagination.PaginateMongo[Product](ctx, collection, filter, params, opts)
}
```

### Web Framework Integration (Gin)

```go
package handlers

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "your-project/pagination"
)

func GetUsersHandler(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Parse pagination parameters from query string
        var params pagination.PaginationParams
        if err := c.ShouldBindQuery(&params); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination parameters"})
            return
        }
        
        // Alternative manual parsing
        if params.Page == 0 {
            if pageStr := c.Query("page"); pageStr != "" {
                if page, err := strconv.Atoi(pageStr); err == nil {
                    params.Page = page
                }
            }
        }
        
        if params.Limit == 0 {
            if limitStr := c.Query("limit"); limitStr != "" {
                if limit, err := strconv.Atoi(limitStr); err == nil {
                    params.Limit = limit
                }
            }
        }
        
        // Apply filters based on query parameters
        query := db.Model(&User{})
        
        if name := c.Query("name"); name != "" {
            query = query.Where("name ILIKE ?", "%"+name+"%")
        }
        
        if status := c.Query("status"); status != "" {
            query = query.Where("status = ?", status)
        }
        
        // Get paginated results
        response, err := pagination.PaginateSQL[User](c.Request.Context(), query, params)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
            return
        }
        
        c.JSON(http.StatusOK, response)
    }
}

// Usage: GET /users?page=2&limit=20&name=john&status=active
```

### Web Framework Integration (Fiber)

```go
package handlers

import (
    "github.com/gofiber/fiber/v2"
    "your-project/pagination"
)

func GetProductsHandler(collection *mongo.Collection) fiber.Handler {
    return func(c *fiber.Ctx) error {
        var params pagination.PaginationParams
        if err := c.QueryParser(&params); err != nil {
            return c.Status(400).JSON(fiber.Map{"error": "Invalid parameters"})
        }
        
        // Build MongoDB filter from query parameters
        filter := bson.M{}
        
        if category := c.Query("category"); category != "" {
            filter["category"] = category
        }
        
        if inStock := c.Query("inStock"); inStock == "true" {
            filter["inStock"] = true
        }
        
        if minPrice := c.QueryInt("minPrice", 0); minPrice > 0 {
            filter["price"] = bson.M{"$gte": minPrice}
        }
        
        // Sorting options
        var opts *options.FindOptions
        if sortBy := c.Query("sortBy"); sortBy != "" {
            sortOrder := 1
            if c.Query("sortOrder") == "desc" {
                sortOrder = -1
            }
            opts = options.Find().SetSort(bson.M{sortBy: sortOrder})
        }
        
        response, err := pagination.PaginateMongo[Product](c.Context(), collection, filter, params, opts)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch products"})
        }
        
        return c.JSON(response)
    }
}

// Usage: GET /products?page=1&limit=25&category=electronics&inStock=true&sortBy=price&sortOrder=asc
```

## API Reference

### Parameter Validation

The package automatically validates and normalizes pagination parameters:

| Parameter | Invalid Value | Normalized To | Notes |
|-----------|---------------|---------------|-------|
| `Page`    | ≤ 0          | 1             | Page numbers start from 1 |
| `Limit`   | ≤ 0          | 10            | Default page size |
| `Limit`   | > 100        | 10            | Maximum limit for performance |

### Response Format

All pagination functions return a consistent response format:

```json
{
  "message": "Success",
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com"
    }
  ],
  "pagination": {
    "currentPage": 1,
    "totalPages": 5,
    "pageSize": 10,
    "totalItems": 50
  }
}
```

### MongoDB Find Options

The `PaginateMongo` function supports all standard MongoDB find options:

```go
opts := options.Find().
    SetSort(bson.M{"createdAt": -1}).              // Sorting
    SetProjection(bson.M{"password": 0}).          // Field projection
    SetHint(bson.M{"email": 1}).                   // Index hints
    SetCollation(&options.Collation{               // Collation
        Locale: "en",
        Strength: 2,
    })

response, err := pagination.PaginateMongo[User](ctx, collection, filter, params, opts)
```

## Best Practices

### 1. Always Use Context

```go
// ✅ Good - with timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

response, err := pagination.PaginateSQL[User](ctx, query, params)

// ❌ Bad - no timeout
ctx := context.Background()
response, err := pagination.PaginateSQL[User](ctx, query, params)
```

### 2. Handle Errors Appropriately

```go
// ✅ Good - proper error handling
response, err := pagination.PaginateSQL[User](ctx, query, params)
if err != nil {
    log.Printf("Pagination error: %v", err)
    return handleDatabaseError(err)
}

// ❌ Bad - ignoring errors
response, _ := pagination.PaginateSQL[User](ctx, query, params)
```

### 3. Use Appropriate Limits

```go
// ✅ Good - reasonable limits
params := pagination.PaginationParams{
    Page:  1,
    Limit: 25, // Good for most use cases
}

// ❌ Bad - too large limits
params := pagination.PaginationParams{
    Page:  1,
    Limit: 1000, // Will be normalized to 10
}
```

### 4. Optimize Database Queries

```go
// ✅ Good - use indexes and efficient queries
query := db.Model(&User{}).
    Where("status = ?", "active").           // Use indexed fields
    Where("created_at > ?", lastMonth).      // Use indexed timestamp
    Select("id, name, email, created_at")    // Select only needed fields

// ❌ Bad - full table scan
query := db.Model(&User{}).
    Where("LOWER(description) LIKE ?", "%search%") // Avoid functions in WHERE
```

### 5. Use Proper Types

```go
// ✅ Good - use specific struct types
type UserResponse struct {
    ID    uint   `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

response, err := pagination.PaginateSQL[UserResponse](ctx, query, params)

// ❌ Bad - using interface{}
response, err := pagination.PaginateSQL[interface{}](ctx, query, params)
```

## Error Handling

The package returns specific errors for different scenarios:

### Common Error Types

1. **Database Connection Errors**
```go
response, err := pagination.PaginateSQL[User](ctx, query, params)
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        // Handle no records found
        return handleEmptyResult()
    }
    if errors.Is(err, context.DeadlineExceeded) {
        // Handle timeout
        return handleTimeout()
    }
    // Handle other database errors
    return handleDatabaseError(err)
}
```

2. **MongoDB Specific Errors**
```go
response, err := pagination.PaginateMongo[Product](ctx, collection, filter, params)
if err != nil {
    if errors.Is(err, mongo.ErrNoDocuments) {
        // Handle no documents found
        return handleEmptyResult()
    }
    if mongo.IsDuplicateKeyError(err) {
        // Handle duplicate key error
        return handleDuplicateError(err)
    }
    // Handle other MongoDB errors
    return handleMongoError(err)
}
```

### Custom Error Handler

```go
func handlePaginationError(err error) error {
    switch {
    case errors.Is(err, context.DeadlineExceeded):
        return fmt.Errorf("request timeout: %w", err)
    case errors.Is(err, gorm.ErrRecordNotFound):
        return fmt.Errorf("no records found: %w", err)
    case errors.Is(err, mongo.ErrNoDocuments):
        return fmt.Errorf("no documents found: %w", err)
    default:
        return fmt.Errorf("database error: %w", err)
    }
}
```

## Performance Considerations

### SQL Optimization

1. **Use Indexes**: Ensure your WHERE clauses use indexed columns
2. **Limit Columns**: Use `Select()` to fetch only needed columns
3. **Count Optimization**: For large tables, consider implementing count caching

```go
// Optimized query
query := db.Model(&User{}).
    Select("id, name, email").              // Only needed columns
    Where("status = ?", "active").          // Indexed column
    Where("created_at > ?", lastWeek)       // Indexed timestamp
```

### MongoDB Optimization

1. **Use Indexes**: Create compound indexes for your filter fields
2. **Projection**: Use projection to limit returned fields
3. **Aggregation**: For complex queries, consider using aggregation pipeline

```go
// Create compound index
collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
    Keys: bson.D{{"category", 1}, {"inStock", 1}, {"price", 1}},
})

// Optimized query with projection
opts := options.Find().
    SetProjection(bson.M{"name": 1, "price": 1}). // Only needed fields
    SetSort(bson.M{"createdAt": -1})

response, err := pagination.PaginateMongo[Product](ctx, collection, filter, params, opts)
```

---
