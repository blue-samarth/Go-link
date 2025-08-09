# Organization Models Package - Complete Documentation

## Table of Contents
1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Core Data Types](#core-data-types)
4. [FlexibleID System](#flexibleid-system)
5. [Organization Entity](#organization-entity)
6. [Request/Response Types](#requestresponse-types)
7. [Validation Functions](#validation-functions)
8. [Constructor Functions](#constructor-functions)
9. [Database Operations](#database-operations)
10. [Organization Methods](#organization-methods)
11. [GORM Hooks](#gorm-hooks)
12. [Utility Functions](#utility-functions)
13. [Error Handling](#error-handling)
14. [Concurrency & Thread Safety](#concurrency--thread-safety)
15. [Logging System](#logging-system)
16. [Usage Examples](#usage-examples)
17. [Performance Considerations](#performance-considerations)

## Overview

The `models` package provides a comprehensive organization management system designed for enterprise applications that need to support both MongoDB and SQL databases simultaneously. It implements a sophisticated hybrid ID system, comprehensive validation, secure password management, and extensive logging capabilities.

### Key Features
- **Dual Database Architecture**: Seamless support for MongoDB and SQL databases with unified APIs
- **Hybrid ID Management**: FlexibleID system handles both MongoDB ObjectIDs and SQL auto-increment IDs
- **Enterprise Security**: BCrypt password hashing, input validation, and secure data handling
- **Production Logging**: Multi-level structured logging with environment-specific verbosity
- **Concurrency Safety**: Thread-safe operations with read-write mutex protection
- **Lifecycle Management**: Complete organization status management with audit trails
- **Validation Framework**: Comprehensive input validation with detailed error reporting
- **Context Support**: Full context cancellation and timeout support for all operations

## Architecture

The package follows a layered architecture with clear separation of concerns:

```
┌─────────────────────────┐
│   API Layer             │ ← Response Types, Validation
├─────────────────────────┤
│   Business Logic        │ ← Organization Methods, Status Management
├─────────────────────────┤
│   Data Access Layer     │ ← Database Operations (MongoDB/SQL)
├─────────────────────────┤
│   Core Types            │ ← FlexibleID, Organization Entity
├─────────────────────────┤
│   Infrastructure        │ ← Logging, Validation, Utilities
└─────────────────────────┘
```

## Core Data Types

### IdType Enumeration

Defines the type of ID system being used for an organization.

```go
type IdType string

const (
    MongoIdType  IdType = "mongo"  // For MongoDB ObjectIDs
    SerialIdType IdType = "serial" // For SQL auto-increment IDs
)
```

**Usage Context:**
- `MongoIdType`: Used when working with MongoDB collections
- `SerialIdType`: Used when working with SQL databases (PostgreSQL, MySQL, etc.)

### OrgStatus Enumeration

Represents the current status of an organization in its lifecycle.

```go
type OrgStatus string

const (
    OrgStatusActive    OrgStatus = "active"    // Organization is operational
    OrgStatusInactive  OrgStatus = "inactive"  // Organization is temporarily disabled
    OrgStatusSuspended OrgStatus = "suspended" // Organization is suspended due to policy violation
    OrgStatusDeleted   OrgStatus = "deleted"   // Organization is soft-deleted
)
```

#### OrgStatus Methods

##### IsValid
Validates if the status value is one of the allowed enum values.

**Signature:**
```go
func (s OrgStatus) IsValid() bool
```

**Returns:**
- `bool`: True if status is valid, false otherwise

**Behavior:**
- Logs warning for invalid status values
- Returns false for empty or unrecognized status values

**Usage:**
```go
status := OrgStatus("active")
if status.IsValid() {
    // Proceed with valid status
}
```

##### Value (SQL Driver Interface)
Implements the `driver.Valuer` interface for SQL database storage.

**Signature:**
```go
func (s OrgStatus) Value() (driver.Value, error)
```

**Returns:**
- `driver.Value`: String representation of the status
- `error`: Always nil for this implementation

##### Scan (SQL Driver Interface)
Implements the `sql.Scanner` interface for SQL database retrieval.

**Signature:**
```go
func (s *OrgStatus) Scan(value interface{}) error
```

**Parameters:**
- `value`: Database value (string, []byte, or nil)

**Returns:**
- `error`: Scanning or validation error

**Behavior:**
- Handles nil values by setting to `OrgStatusInactive`
- Converts string and []byte values to OrgStatus
- Validates scanned value using `IsValid()`
- Logs errors for unsupported types or invalid values

## FlexibleID System

The FlexibleID system is a sophisticated hybrid identification mechanism that allows the same codebase to work seamlessly with both MongoDB ObjectIDs and SQL auto-increment IDs.

### FlexibleID Structure

```go
type FlexibleID struct {
    Type     IdType              `json:"type" bson:"type"`
    MongoID  *primitive.ObjectID `json:"mongo_id,omitempty" bson:"mongo_id,omitempty"`
    SerialID *int64              `json:"serial_id,omitempty" bson:"serial_id,omitempty"`
}
```

**Design Principles:**
- Only one ID field is populated based on the Type
- JSON serialization includes both type and value information
- BSON serialization stores the actual ID value directly
- SQL scanning automatically detects ID type based on database value

### FlexibleID Constructor Functions

#### NewMongoID
Creates a new FlexibleID with a fresh MongoDB ObjectID.

**Signature:**
```go
func NewMongoID() *FlexibleID
```

**Returns:**
- `*FlexibleID`: New FlexibleID with generated MongoDB ObjectID

**Usage:**
```go
id := NewMongoID()
fmt.Println(id.String()) // "507f1f77bcf86cd799439011"
```

#### FromMongoID
Creates a FlexibleID from an existing MongoDB ObjectID.

**Signature:**
```go
func FromMongoID(id primitive.ObjectID) *FlexibleID
```

**Parameters:**
- `id`: Existing MongoDB ObjectID

**Returns:**
- `*FlexibleID`: FlexibleID wrapping the provided ObjectID

#### NewSerialID
Creates a FlexibleID with a specific serial ID value.

**Signature:**
```go
func NewSerialID(id int64) *FlexibleID
```

**Parameters:**
- `id`: Serial ID value

**Returns:**
- `*FlexibleID`: FlexibleID with the specified serial ID

#### FromSerialID
Alias for NewSerialID for consistency.

**Signature:**
```go
func FromSerialID(id int64) *FlexibleID
```

### FlexibleID Methods

#### String
Returns string representation of the ID.

**Signature:**
```go
func (f *FlexibleID) String() string
```

**Returns:**
- `string`: Hex string for MongoDB ObjectID, decimal string for serial ID, empty for nil/invalid

**Behavior:**
- MongoDB ObjectIDs: Returns 24-character hex string
- Serial IDs: Returns decimal number as string
- Nil or invalid IDs: Returns empty string

#### IsEmpty
Checks if the FlexibleID contains no valid ID.

**Signature:**
```go
func (f *FlexibleID) IsEmpty() bool
```

**Returns:**
- `bool`: True if ID is nil, zero, or invalid

**Behavior:**
- Returns true for nil FlexibleID
- Returns true for MongoDB ObjectIDs that are zero values
- Returns true for serial IDs that are 0 or nil
- Returns true for unrecognized ID types

#### GetValue
Returns the raw ID value for database operations.

**Signature:**
```go
func (f *FlexibleID) GetValue() interface{}
```

**Returns:**
- `interface{}`: `primitive.ObjectID` for MongoDB, `int64` for serial, `nil` for empty

#### Value (SQL Driver Interface)
Implements `driver.Valuer` for SQL database storage.

**Signature:**
```go
func (f *FlexibleID) Value() (driver.Value, error)
```

**Returns:**
- `driver.Value`: `int64` for serial IDs, hex string for MongoDB ObjectIDs, `nil` for empty
- `error`: Always nil for this implementation

#### Scan (SQL Driver Interface)
Implements `sql.Scanner` for SQL database retrieval.

**Signature:**
```go
func (f *FlexibleID) Scan(value interface{}) error
```

**Parameters:**
- `value`: Database value (`int64`, `string`, `[]byte`, or `nil`)

**Returns:**
- `error`: Parsing or type error

**Behavior:**
- `int64` values: Stored as serial IDs
- `string` values: First attempts MongoDB ObjectID parsing, falls back to serial ID parsing
- `[]byte` values: Converted to string and processed
- `nil` values: Leaves FlexibleID unchanged
- Logs errors for unsupported types or parsing failures

#### MarshalJSON
Custom JSON serialization for API responses.

**Signature:**
```go
func (f *FlexibleID) MarshalJSON() ([]byte, error)
```

**Returns:**
- `[]byte`: JSON representation with type and value fields
- `error`: JSON marshaling error

**JSON Format:**
```json
{
  "type": "mongo",
  "value": "507f1f77bcf86cd799439011"
}
```

#### UnmarshalJSON
Custom JSON deserialization for API requests.

**Signature:**
```go
func (f *FlexibleID) UnmarshalJSON(data []byte) error
```

**Parameters:**
- `data`: JSON bytes to unmarshal

**Returns:**
- `error`: JSON parsing or validation error

**Behavior:**
- Expects JSON object with "type" and "value" fields
- Validates type field against known IdType values
- Parses value field according to the specified type
- Logs errors for invalid formats or parsing failures

#### MarshalBSONValue
Custom BSON serialization for MongoDB storage.

**Signature:**
```go
func (f *FlexibleID) MarshalBSONValue() (bson.ValueType, []byte, error)
```

**Returns:**
- `bson.ValueType`: `TypeObjectID` for MongoDB IDs, `TypeInt64` for serial IDs, `TypeNull` for empty
- `[]byte`: BSON-encoded value
- `error`: BSON marshaling error

**Behavior:**
- Stores actual ID values directly in BSON (not wrapped objects)
- MongoDB ObjectIDs are stored as native BSON ObjectIDs
- Serial IDs are stored as BSON Int64 values

#### UnmarshalBSONValue
Custom BSON deserialization for MongoDB retrieval.

**Signature:**
```go
func (f *FlexibleID) UnmarshalBSONValue(t bson.ValueType, data []byte) error
```

**Parameters:**
- `t`: BSON value type
- `data`: BSON-encoded data

**Returns:**
- `error`: BSON parsing error

**Behavior:**
- `TypeObjectID`: Unmarshals as MongoDB ObjectID
- `TypeInt64`/`TypeInt32`: Unmarshals as serial ID (converts Int32 to Int64)
- Other types: Returns error with logging

## Organization Entity

The Organization struct represents the core entity with full database compatibility and thread safety.

### Organization Structure

```go
type Organization struct {
    ID          *FlexibleID `json:"id" bson:"_id" gorm:"-"`
    SerialID    *int64      `json:"-" gorm:"column:id;primaryKey;autoIncrement"`
    Name        string      `json:"name" bson:"name" gorm:"column:name;size:100;not null"`
    Email       string      `json:"email" bson:"email" gorm:"column:email;unique;not null"`
    Password    string      `json:"-" bson:"password" gorm:"column:password;not null"`
    Status      OrgStatus   `json:"status" bson:"status" gorm:"column:status;type:varchar(20);not null"`
    CreatedAt   time.Time   `json:"created_at" bson:"created_at" gorm:"column:created_at;autoCreateTime"`
    LastUpdated time.Time   `json:"last_updated" bson:"last_updated" gorm:"column:last_updated;autoUpdateTime"`
    DeletedAt   *time.Time  `json:"deleted_at,omitempty" bson:"deleted_at,omitempty" gorm:"column:deleted_at"`
    mu          sync.RWMutex `json:"-" bson:"-" gorm:"-"`
}
```

**Field Details:**

- **ID**: Hybrid identifier using FlexibleID system
- **SerialID**: SQL-specific auto-increment ID (hidden from JSON)
- **Name**: Organization name (3-100 characters, required)
- **Email**: Unique email address (validated, required)
- **Password**: BCrypt-hashed password (hidden from JSON, required)
- **Status**: Organization status enum (required)
- **CreatedAt**: Creation timestamp (auto-managed)
- **LastUpdated**: Last modification timestamp (auto-managed)
- **DeletedAt**: Soft deletion timestamp (optional)
- **mu**: Read-write mutex for thread safety (hidden)

**Database Mapping:**
- **GORM Tags**: SQL database column definitions and constraints
- **BSON Tags**: MongoDB field mappings
- **JSON Tags**: API serialization control (passwords and internal fields hidden)

### Organization Database Interface Methods

#### TableName
Returns the SQL table name for GORM.

**Signature:**
```go
func (Organization) TableName() string
```

**Returns:**
- `string`: "organizations"

**Usage:**
GORM automatically uses this method to determine the table name for SQL operations.

#### CollectionName
Returns the MongoDB collection name.

**Signature:**
```go
func (Organization) CollectionName() string
```

**Returns:**
- `string`: "organizations"

**Usage:**
Used by MongoDB operations to determine the collection name.

## Request/Response Types

### OrganizationCreate

Request type for creating new organizations with comprehensive validation tags.

```go
type OrganizationCreate struct {
    Name        string     `json:"name" validate:"required,min=3,max=100"`
    Email       string     `json:"email" validate:"required,email"`
    Status      OrgStatus  `json:"status" validate:"required,oneof=active inactive suspended deleted"`
    Password    string     `json:"password" validate:"required,min=8"`
    CreatedAt   time.Time  `json:"created_at"`
    LastUpdated time.Time  `json:"last_updated"`
    DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}
```

**Validation Rules:**
- **Name**: Required, 3-100 characters
- **Email**: Required, valid email format
- **Status**: Required, must be valid OrgStatus
- **Password**: Required, minimum 8 characters
- **Timestamps**: Usually set programmatically

### OrganizationUpdate

Request type for updating existing organizations with optional fields.

```go
type OrganizationUpdate struct {
    Name        *string    `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
    Email       *string    `json:"email,omitempty" validate:"omitempty,email"`
    Status      *OrgStatus `json:"status,omitempty" validate:"omitempty,oneof=active inactive suspended deleted"`
    Password    *string    `json:"password,omitempty" validate:"omitempty,min=8"`
    CreatedAt   *time.Time `json:"created_at,omitempty"`
    LastUpdated *time.Time `json:"last_updated,omitempty"`
    DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}
```

**Design Notes:**
- All fields are optional pointers
- Only non-nil fields are updated
- Same validation rules as create when fields are present

### OrganizationResponse

Safe response type for API consumers, excluding sensitive information.

```go
type OrganizationResponse struct {
    ID          interface{} `json:"id"`
    Name        string      `json:"name"`
    Email       string      `json:"email"`
    Status      OrgStatus   `json:"status"`
    CreatedAt   time.Time   `json:"created_at"`
    LastUpdated time.Time   `json:"last_updated"`
    DeletedAt   *time.Time  `json:"deleted_at,omitempty"`
}
```

**Security Features:**
- Password field completely excluded
- ID field dynamically typed (string for MongoDB, int64 for SQL)
- All timestamps preserved for audit purposes

## Validation Functions

### IsValidEmail

Comprehensive email validation using regex pattern matching.

**Signature:**
```go
func IsValidEmail(email string) bool
```

**Parameters:**
- `email`: Email address to validate

**Returns:**
- `bool`: True if email format is valid

**Validation Pattern:**
```regexp
^[a-zA-Z0-9](\.?[a-zA-Z0-9_\-+%]){0,63}@[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z]{2,})+$
```

**Behavior:**
- Returns false for empty strings
- Logs warnings for invalid emails
- Supports common email formats including plus addressing and subdomain emails

**Examples:**
```go
IsValidEmail("user@example.com")     // true
IsValidEmail("user+tag@example.com") // true
IsValidEmail("user@sub.example.co.uk") // true
IsValidEmail("invalid.email")        // false (logs warning)
```

### validateOrganizationCreateRequest

Internal comprehensive validation for organization creation requests.

**Signature:**
```go
func validateOrganizationCreateRequest(req OrganizationCreate) error
```

**Parameters:**
- `req`: Organization creation request to validate

**Returns:**
- `error`: Detailed validation error or nil if valid

**Validation Rules:**
1. **Name**: Must not be empty after trimming whitespace
2. **Email**: Must pass `IsValidEmail()` check
3. **Status**: Must be non-empty and pass `IsValid()` check
4. **Password**: Must be at least 8 characters
5. **CreatedAt**: Must not be zero time
6. **LastUpdated**: Must not be zero time
7. **DeletedAt**: Can only be set if status is "deleted"

**Error Examples:**
- `"name is required"`
- `"invalid email format: invalid@"`
- `"status is required and must be one of: active, inactive, suspended, deleted"`
- `"password must be at least 8 characters long"`
- `"deleted_at can only be set if status is 'deleted'"`

**Logging:**
- Logs warnings for all validation failures
- Logs info message for successful validation

### validatePaginationParams

Internal validation for pagination parameters.

**Signature:**
```go
func validatePaginationParams(limit, offset int) error
```

**Parameters:**
- `limit`: Maximum number of records to return
- `offset`: Number of records to skip

**Returns:**
- `error`: Validation error or nil if valid

**Validation Rules:**
- `limit`: Must be >= 0 and <= 1000
- `offset`: Must be >= 0

**Error Examples:**
- `"limit cannot be negative"`
- `"offset cannot be negative"`
- `"limit cannot exceed 1000"`

## Constructor Functions

### NewOrganization

Master constructor function that creates organization instances with specified ID types.

**Signature:**
```go
func NewOrganization(ctx context.Context, req OrganizationCreate, idType IdType, serialID ...int64) (*Organization, error)
```

**Parameters:**
- `ctx`: Context for cancellation and timeout handling
- `req`: Organization creation request with all required fields
- `idType`: `MongoIdType` or `SerialIdType`
- `serialID`: Optional serial ID (required if `idType` is `SerialIdType`)

**Returns:**
- `*Organization`: Created organization instance (not persisted)
- `error`: Validation, context, or creation error

**Behavior:**
1. **Context Check**: Returns immediately if context is cancelled
2. **Validation**: Runs comprehensive validation on the request
3. **Password Hashing**: Uses BCrypt with default cost (currently 10)
4. **ID Assignment**: Creates appropriate FlexibleID based on idType
5. **Field Population**: Copies all fields from request to organization
6. **Logging**: Logs creation process with organization details

**Error Scenarios:**
- Context cancelled: Returns `ctx.Err()`
- Validation failure: Returns detailed validation error
- BCrypt failure: Returns wrapped hashing error
- Invalid ID type: Returns "invalid ID type" error
- Missing serial ID: Returns "serial ID required for SerialIdType" error

**Usage Examples:**
```go
// MongoDB organization
org, err := NewOrganization(ctx, req, MongoIdType)

// SQL organization with auto-generated ID
org, err := NewOrganization(ctx, req, SerialIdType)

// SQL organization with specific ID
org, err := NewOrganization(ctx, req, SerialIdType, 12345)
```

### NewOrganizationForMongo

Convenience constructor specifically for MongoDB organizations.

**Signature:**
```go
func NewOrganizationForMongo(ctx context.Context, req OrganizationCreate) (*Organization, error)
```

**Parameters:**
- `ctx`: Context for cancellation and timeout handling
- `req`: Organization creation request

**Returns:**
- `*Organization`: Organization with MongoDB ObjectID
- `error`: Creation error

**Behavior:**
- Internally calls `NewOrganization(ctx, req, MongoIdType)`
- Automatically generates a new MongoDB ObjectID
- Logs MongoDB-specific creation

### NewOrganizationForSQL

Convenience constructor specifically for SQL database organizations.

**Signature:**
```go
func NewOrganizationForSQL(ctx context.Context, req OrganizationCreate, id ...int64) (*Organization, error)
```

**Parameters:**
- `ctx`: Context for cancellation and timeout handling
- `req`: Organization creation request
- `id`: Optional specific serial ID

**Returns:**
- `*Organization`: Organization with serial ID
- `error`: Creation error

**Behavior:**
- Internally calls `NewOrganization(ctx, req, SerialIdType, id...)`
- If no ID provided, allows database to auto-generate
- If ID provided, uses that specific value
- Syncs FlexibleID with SerialID field
- Logs SQL-specific creation

## Database Operations

### Creation Operations

#### CreateOrganizationMongo

Creates and persists a new organization in MongoDB.

**Signature:**
```go
func CreateOrganizationMongo(ctx context.Context, req OrganizationCreate) (*Organization, error)
```

**Parameters:**
- `ctx`: Database operation context
- `req`: Organization creation request

**Returns:**
- `*Organization`: Created and persisted organization
- `error`: Creation, validation, or database error

**Behavior:**
1. **Creation**: Calls `NewOrganizationForMongo()` to create organization instance
2. **Collection Access**: Gets "organizations" collection from `config.MongoDB`
3. **Insertion**: Uses MongoDB `InsertOne()` with provided context
4. **Logging**: Logs creation attempts and results
5. **Error Handling**: Returns detailed errors for validation or database issues

**Error Scenarios:**
- Validation failures from `NewOrganizationForMongo()`
- MongoDB connection issues
- Duplicate key violations (email uniqueness)
- Context cancellation or timeout

#### CreateOrganizationSQL

Creates and persists a new organization in SQL database using transactions.

**Signature:**
```go
func CreateOrganizationSQL(ctx context.Context, req OrganizationCreate) (*Organization, error)
```

**Parameters:**
- `ctx`: Database operation context
- `req`: Organization creation request

**Returns:**
- `*Organization`: Created and persisted organization with auto-generated ID
- `error`: Creation, validation, or database error

**Behavior:**
1. **Creation**: Calls `NewOrganizationForSQL()` to create organization instance
2. **Transaction**: Wraps creation in GORM transaction for atomicity
3. **GORM Hooks**: Triggers `AfterCreate` hook to sync FlexibleID
4. **Auto-Increment**: Database generates SerialID, synced to FlexibleID
5. **Rollback**: Automatically rolls back on any error

**Error Scenarios:**
- Validation failures from `NewOrganizationForSQL()`
- SQL constraint violations (unique email, not null constraints)
- Transaction failures
- Context cancellation or timeout

### Read Operations

#### GetOrganizationByIDMongo

Retrieves a single organization from MongoDB by FlexibleID.

**Signature:**
```go
func GetOrganizationByIDMongo(ctx context.Context, id *FlexibleID) (*Organization, error)
```

**Parameters:**
- `ctx`: Database operation context
- `id`: FlexibleID containing MongoDB ObjectID

**Returns:**
- `*Organization`: Found organization or nil
- `error`: "not found" error or database error

**Behavior:**
1. **ID Validation**: Checks for nil ID
2. **Query**: Uses MongoDB `FindOne()` with `_id` filter
3. **Decoding**: Decodes BSON document to Organization struct
4. **Not Found Handling**: Returns specific "not found" error
5. **Error Logging**: Logs database errors with ID context

**Error Scenarios:**
- `id` parameter is nil: Returns "id cannot be nil"
- Document not found: Returns formatted "organization not found with id: ..." error
- Database connection issues: Returns database error with logging
- BSON decoding errors: Returns decoding error

#### GetOrganizationByIDSQL

Retrieves a single organization from SQL database by FlexibleID.

**Signature:**
```go
func GetOrganizationByIDSQL(ctx context.Context, id *FlexibleID) (*Organization, error)
```

**Parameters:**
- `ctx`: Database operation context
- `id`: FlexibleID containing serial ID

**Returns:**
- `*Organization`: Found organization or nil
- `error`: "not found" error or database error

**Behavior:**
1. **ID Validation**: Checks for nil ID and correct type
2. **Type Check**: Ensures ID is SerialIdType with non-nil SerialID
3. **Query**: Uses GORM `First()` with WHERE clause
4. **Not Found Handling**: Detects `gorm.ErrRecordNotFound`
5. **Error Logging**: Logs database errors with ID context

**Error Scenarios:**
- `id` parameter is nil: Returns "id cannot be nil"
- Invalid ID type: Returns "invalid ID type for SQL query"
- Record not found: Returns formatted "organization not found with id: ..." error
- Database connection issues: Returns database error with logging

#### GetOrganizationByEmailMongo

Retrieves organization from MongoDB by unique email address.

**Signature:**
```go
func GetOrganizationByEmailMongo(ctx context.Context, email string) (*Organization, error)
```

**Parameters:**
- `ctx`: Database operation context
- `email`: Email address to search for

**Returns:**
- `*Organization`: Found organization or nil
- `error`: "not found" error or database error

**Behavior:**
1. **Email Validation**: Checks for empty email
2. **Query**: Uses MongoDB `FindOne()` with email filter
3. **Not Found Handling**: Returns specific "not found" error
4. **Warning Logging**: Logs warnings for not found cases
5. **Error Logging**: Logs database errors

#### GetOrganizationByEmailSQL

Retrieves organization from SQL database by unique email address.

**Signature:**
```go
func GetOrganizationByEmailSQL(ctx context.Context, email string) (*Organization, error)
```

**Parameters:**
- `ctx`: Database operation context
- `email`: Email address to search for

**Returns:**
- `*Organization`: Found organization or nil
- `error`: "not found" error or database error

**Behavior:**
- Similar to MongoDB version but uses GORM `First()` with WHERE clause
- Detects `gorm.ErrRecordNotFound` for not found cases

#### GetOrganizationByNameMongo

Retrieves organization from MongoDB by name.

**Signature:**
```go
func GetOrganizationByNameMongo(ctx context.Context, name string) (*Organization, error)
```

**Parameters:**
- `ctx`: Database operation context
- `name`: Organization name to search for

**Returns:**
- `*Organization`: Found organization or nil
- `error`: "not found" error or database error

**Note:** Name is not necessarily unique, so this returns the first match found.

#### GetOrganizationByNameSQL

Retrieves organization from SQL database by name.

**Signature:**
```go
func GetOrganizationByNameSQL(ctx context.Context, name string) (*Organization, error)
```

**Parameters:**
- `ctx`: Database operation context
- `name`: Organization name to search for

**Returns:**
- `*Organization`: Found organization or nil
- `error`: "not found" error or database error

#### GetOrganizationsMongo

Retrieves multiple organizations from MongoDB with pagination.

**Signature:**
```go
func GetOrganizationsMongo(ctx context.Context, limit, offset int) ([]*Organization, error)
```

**Parameters:**
- `ctx`: Database operation context
- `limit`: Maximum number of organizations to return (0-1000)
- `offset`: Number of organizations to skip

**Returns:**
- `[]*Organization`: Array of organizations (may be empty)
- `error`: Database or validation error

**Behavior:**
1. **Validation**: Validates pagination parameters
2. **Options**: Creates MongoDB find options with limit and skip
3. **Cursor**: Uses MongoDB cursor for efficient iteration
4. **Decoding**: Decodes each document individually
5. **Resource Management**: Properly closes cursor
6. **Error Handling**: Handles cursor errors and decoding errors

#### GetOrganizationsSQL

Retrieves multiple organizations from SQL database with pagination.

**Signature:**
```go
func GetOrganizationsSQL(ctx context.Context, limit, offset int) ([]*Organization, error)
```

**Parameters:**
- `ctx`: Database operation context
- `limit`: Maximum number of organizations to return (0-1000)
- `offset`: Number of organizations to skip

**Returns:**
- `[]*Organization`: Array of organizations (may be empty)
- `error`: Database or validation error

**Behavior:**
- Uses GORM `Limit()` and `Offset()` for pagination
- Returns slice of organization pointers
- Handles empty results gracefully

### Update Operations

#### applyOrganizationUpdates

Internal function that applies update request to organization instance.

**Signature:**
```go
func applyOrganizationUpdates(ctx context.Context, req OrganizationUpdate, org *Organization) error
```

**Parameters:**
- `ctx`: Context (currently for future use)
- `req`: Update request with optional fields
- `org`: Organization instance to modify

**Returns:**
- `error`: Validation or application error

**Behavior:**
1. **Password Update**: If password provided, uses `SetPassword()` method
2. **Field Updates**: Updates only non-nil fields from request
3. **Email Validation**: Validates email if provided
4. **Status Validation**: Validates status if provided
5. **Timestamp Management**: Always updates `LastUpdated` to current time
6. **Soft Delete Logic**: Sets `DeletedAt` if status is "deleted"
7. **Thread Safety**: Uses organization's mutex for safe updates

**Update Logic:**
- Only non-nil pointer fields are updated
- Password updates trigger BCrypt hashing
- Status changes to "deleted" automatically set `DeletedAt`
- Status changes away from "deleted" clear `DeletedAt` (unless explicitly set)
- `LastUpdated` is always set to current time

#### UpdateOrganizationMongo

Performs full update of organization in MongoDB.

**Signature:**
```go
func UpdateOrganizationMongo(ctx context.Context, req OrganizationUpdate, org *Organization) error
```

**Parameters:**
- `ctx`: Database operation context
- `req`: Update request with optional fields
- `org`: Organization instance to update (modified in-place)

**Returns:**
- `error`: Database or validation error

**Behavior:**
1. **Nil Check**: Validates organization is not nil
2. **Apply Updates**: Uses `applyOrganizationUpdates()` to modify organization
3. **Database Update**: Uses MongoDB `UpdateOne()` with `$set` operation
4. **Filter Creation**: Uses organization's MongoDB ObjectID as filter
5. **Error Logging**: Logs update failures with organization context

**Error Scenarios:**
- Organization is nil: Returns "organization cannot be nil"
- Validation failures from `applyOrganizationUpdates()`
- MongoDB connection issues
- Document not found (if organization was deleted externally)

#### UpdateOrganizationSQL

Performs full update of organization in SQL database using transactions.

**Signature:**
```go
func UpdateOrganizationSQL(ctx context.Context, req OrganizationUpdate, org *Organization) error
```

**Parameters:**
- `ctx`: Database operation context
- `req`: Update request with optional fields
- `org`: Organization instance to update (modified in-place)

**Returns:**
- `error`: Database, validation, or transaction error

**Behavior:**
1. **Nil Check**: Validates organization is not nil
2. **Transaction**: Wraps entire update in GORM transaction
3. **Copy Creation**: Creates copy of organization for atomic updates
4. **Apply Updates**: Uses `applyOrganizationUpdates()` on copy
5. **Database Save**: Uses GORM `Save()` to persist changes
6. **GORM Hooks**: Triggers `BeforeUpdate` hook automatically
7. **Atomic Update**: Copies updated values back to original on success
8. **Rollback**: Automatically rolls back transaction on any error

**Error Scenarios:**
- Organization is nil: Returns "organization cannot be nil"
- Validation failures from `applyOrganizationUpdates()`
- SQL constraint violations
- Transaction failures
- GORM hook failures

#### PatchOrganizationMongo

Alias for `UpdateOrganizationMongo` - performs partial update.

**Signature:**
```go
func PatchOrganizationMongo(ctx context.Context, req OrganizationUpdate, org *Organization) error
```

**Note:** Functionally identical to `UpdateOrganizationMongo`. The separate function exists for API semantic clarity (PATCH vs PUT operations).

#### PatchOrganizationSQL

Alias for `UpdateOrganizationSQL` - performs partial update.

**Signature:**
```go
func PatchOrganizationSQL(ctx context.Context, req OrganizationUpdate, org *Organization) error
```

**Note:** Functionally identical to `UpdateOrganizationSQL`. The separate function exists for API semantic clarity.

### Delete Operations

#### DeleteOrganizationMongo

Permanently removes organization from MongoDB.

**Signature:**
```go
func DeleteOrganizationMongo(ctx context.Context, org *Organization) error
```

**Parameters:**
- `ctx`: Database operation context
- `org`: Organization to delete

**Returns:**
- `error`: Database error

**Behavior:**
1. **Nil Check**: Validates organization is not nil
2. **Logging**: Logs deletion attempt with organization ID
3. **Database Delete**: Uses MongoDB `DeleteOne()` with ObjectID filter
4. **Success Logging**: Logs successful deletion
5. **Error Logging**: Logs deletion failures with context

**Important Notes:**
- This is a hard delete - the document is permanently removed
- For soft deletion, use `org.SoftDeleteOrganization()` followed by an update operation
- No cascading delete logic is implemented

#### DeleteOrganizationSQL

Permanently removes organization from SQL database using transactions.

**Signature:**
```go
func DeleteOrganizationSQL(ctx context.Context, org *Organization) error
```

**Parameters:**
- `ctx`: Database operation context
- `org`: Organization to delete

**Returns:**
- `error`: Database or transaction error

**Behavior:**
1. **Nil Check**: Validates organization is not nil
2. **Transaction**: Wraps deletion in GORM transaction
3. **Logging**: Logs deletion attempt and results
4. **Database Delete**: Uses GORM `Delete()` method
5. **Rollback**: Automatically rolls back on any error

**Important Notes:**
- This is a hard delete - the record is permanently removed
- GORM may perform soft delete if `DeletedAt` field is configured
- For explicit soft deletion, use `org.SoftDeleteOrganization()` followed by update

## Organization Methods

### ID Management Methods

#### SetMongoID

Sets the organization's FlexibleID to a specific MongoDB ObjectID.

**Signature:**
```go
func (o *Organization) SetMongoID(id primitive.ObjectID)
```

**Parameters:**
- `id`: MongoDB ObjectID to set

**Behavior:**
- **Nil Check**: Logs warning and returns if organization is nil
- **Thread Safety**: Uses write lock for safe concurrent access
- **ID Assignment**: Creates new FlexibleID with MongoIdType
- **No Validation**: Does not validate if the ID exists in database

**Usage:**
```go
objectID := primitive.NewObjectID()
org.SetMongoID(objectID)
```

#### SetSerialID

Sets the organization's FlexibleID to a specific serial ID.

**Signature:**
```go
func (o *Organization) SetSerialID(id int64)
```

**Parameters:**
- `id`: Serial ID to set

**Behavior:**
- **Nil Check**: Logs warning and returns if organization is nil
- **Thread Safety**: Uses write lock for safe concurrent access
- **ID Assignment**: Creates new FlexibleID with SerialIdType
- **No Validation**: Does not validate if the ID exists in database

#### GetIDString

Thread-safe method to get string representation of organization ID.

**Signature:**
```go
func (o *Organization) GetIDString() string
```

**Returns:**
- `string`: String representation of ID, or empty string if nil/invalid

**Behavior:**
- **Nil Check**: Returns empty string if organization is nil
- **Thread Safety**: Uses read lock for safe concurrent access
- **Delegation**: Delegates to FlexibleID's `String()` method
- **Empty Handling**: Returns empty string for nil or invalid IDs

#### IsMongoID

Thread-safe check if organization uses MongoDB ObjectID.

**Signature:**
```go
func (o *Organization) IsMongoID() bool
```

**Returns:**
- `bool`: True if organization has valid MongoDB ObjectID

**Behavior:**
- **Nil Check**: Returns false for nil organization
- **Thread Safety**: Uses read lock
- **Type Check**: Verifies FlexibleID type is MongoIdType
- **Value Check**: Ensures MongoID field is not nil

#### IsSerialID

Thread-safe check if organization uses serial ID.

**Signature:**
```go
func (o *Organization) IsSerialID() bool
```

**Returns:**
- `bool`: True if organization has valid serial ID

**Behavior:**
- **Nil Check**: Returns false for nil organization
- **Thread Safety**: Uses read lock
- **Type Check**: Verifies FlexibleID type is SerialIdType
- **Value Check**: Ensures SerialID field is not nil

### Password Management Methods

#### SetPassword

Securely sets and hashes organization password.

**Signature:**
```go
func (o *Organization) SetPassword(password string) error
```

**Parameters:**
- `password`: Plain text password (minimum 8 characters)

**Returns:**
- `error`: Validation or hashing error

**Behavior:**
1. **Nil Check**: Returns error if organization is nil
2. **Length Validation**: Ensures password is at least 8 characters
3. **BCrypt Hashing**: Uses `bcrypt.GenerateFromPassword()` with default cost
4. **Thread Safety**: Uses write lock during password update
5. **Timestamp Update**: Sets `LastUpdated` to current time
6. **Logging**: Logs validation failures and hashing errors

**Security Features:**
- Uses BCrypt with default cost (currently 10)
- Plain text password is never stored
- Logs warnings for short passwords with organization context
- Thread-safe operation prevents race conditions

**Error Scenarios:**
- Organization is nil: Returns "organization cannot be nil"
- Password too short: Returns "password must be at least 8 characters long"
- BCrypt failure: Returns wrapped error with context

#### CheckPassword

Verifies plain text password against stored BCrypt hash.

**Signature:**
```go
func (o *Organization) CheckPassword(password string) bool
```

**Parameters:**
- `password`: Plain text password to verify

**Returns:**
- `bool`: True if password matches stored hash

**Behavior:**
1. **Nil Check**: Returns false with warning if organization is nil
2. **Thread Safety**: Uses read lock to access stored password
3. **Empty Check**: Returns false if no password is stored
4. **BCrypt Verification**: Uses `bcrypt.CompareHashAndPassword()`
5. **Logging**: Logs failures for security monitoring
6. **Timing Safety**: BCrypt comparison is resistant to timing attacks

**Security Features:**
- Constant-time comparison via BCrypt
- Logs failed attempts for security monitoring
- Never logs actual passwords
- Thread-safe read access

### Status Management Methods

#### ActivateOrganization

Sets organization status to active with timestamp updates.

**Signature:**
```go
func (o *Organization) ActivateOrganization()
```

**Behavior:**
1. **Nil Check**: Logs warning and returns if organization is nil
2. **Logging**: Logs activation attempt with organization ID
3. **Thread Safety**: Uses write lock during status change
4. **Status Update**: Sets status to `OrgStatusActive`
5. **Timestamp Update**: Sets `LastUpdated` to current time
6. **Audit Logging**: Logs status change with old and new values

**Usage:**
```go
org.ActivateOrganization()
// Don't forget to persist changes with UpdateOrganization...()
```

#### DeactivateOrganization

Sets organization status to inactive with timestamp updates.

**Signature:**
```go
func (o *Organization) DeactivateOrganization()
```

**Behavior:**
- Similar to `ActivateOrganization()` but sets status to `OrgStatusInactive`
- Includes same logging and thread safety features
- Updates `LastUpdated` timestamp

#### SuspendOrganization

Sets organization status to suspended with enhanced logging.

**Signature:**
```go
func (o *Organization) SuspendOrganization()
```

**Behavior:**
- Similar to activation/deactivation but sets status to `OrgStatusSuspended`
- Uses warning-level logging due to the severity of suspension
- Includes audit trail with old and new status values
- Updates `LastUpdated` timestamp

#### SoftDeleteOrganization

Marks organization as deleted with deletion timestamp.

**Signature:**
```go
func (o *Organization) SoftDeleteOrganization()
```

**Behavior:**
1. **Nil Check**: Logs warning and returns if organization is nil
2. **Logging**: Logs soft deletion attempt
3. **Thread Safety**: Uses write lock during updates
4. **Status Update**: Sets status to `OrgStatusDeleted`
5. **Timestamp Updates**: Sets both `DeletedAt` and `LastUpdated` to current time
6. **Audit Logging**: Uses warning-level logging for deletion events

**Important Notes:**
- This does not remove the organization from the database
- Sets `DeletedAt` timestamp for audit purposes
- Status change and timestamp update are atomic
- Requires subsequent database update to persist changes

### Conversion Methods

#### ToResponse

Converts organization to safe response format for API consumption.

**Signature:**
```go
func (o *Organization) ToResponse() *OrganizationResponse
```

**Returns:**
- `*OrganizationResponse`: Safe response object without sensitive data
- `nil`: If organization is nil

**Behavior:**
1. **Nil Check**: Returns nil with warning if organization is nil
2. **Thread Safety**: Uses read lock during field access
3. **ID Conversion**: Converts FlexibleID to appropriate interface{} type
4. **Field Mapping**: Copies all safe fields to response struct
5. **Security**: Completely excludes password and internal fields
6. **Logging**: Debug-level logging for conversion process

**ID Conversion Logic:**
- MongoDB ObjectIDs: Converted to hex string
- Serial IDs: Kept as int64
- Empty/invalid IDs: Set to nil
- Type information is lost in response (client must infer from value type)

**Excluded Fields:**
- Password (security)
- Internal mutex (implementation detail)
- SerialID (internal SQL field)

## GORM Hooks

GORM hooks provide lifecycle callbacks for SQL database operations.

### BeforeUpdate

Automatic hook called before any GORM update operation.

**Signature:**
```go
func (o *Organization) BeforeUpdate(tx *gorm.DB) error
```

**Parameters:**
- `tx`: GORM database transaction instance

**Returns:**
- `error`: Hook error that cancels the update

**Behavior:**
1. **Context Handling**: Checks for context cancellation
2. **Nil Check**: Returns error if organization is nil
3. **Timestamp Update**: Sets `LastUpdated` to current time
4. **Soft Delete Logic**: Sets `DeletedAt` if status is "deleted"
5. **Delete Clear Logic**: Clears `DeletedAt` if status is not "deleted"
6. **Thread Safety**: Uses write lock during updates
7. **Logging**: Logs hook execution and soft delete events

**Automatic Behaviors:**
- Always updates `LastUpdated` timestamp
- Manages `DeletedAt` timestamp based on status
- Provides audit logging for deletion events

### AfterCreate

Automatic hook called after successful GORM create operation.

**Signature:**
```go
func (o *Organization) AfterCreate(tx *gorm.DB) error
```

**Parameters:**
- `tx`: GORM database transaction instance

**Returns:**
- `error`: Hook error (rarely used for AfterCreate)

**Behavior:**
1. **Context Handling**: Checks for context cancellation
2. **Nil Check**: Returns error if organization is nil
3. **ID Synchronization**: Syncs auto-generated SerialID to FlexibleID
4. **Logging**: Logs ID synchronization process
5. **Thread Safety**: Direct field access (called within transaction)

**ID Synchronization:**
- If SerialID is populated but FlexibleID is empty
- Creates new FlexibleID with SerialIdType
- Ensures FlexibleID reflects database-generated ID

**Usage Notes:**
- Only relevant for SQL database operations
- Called automatically by GORM after successful INSERT
- Critical for maintaining FlexibleID consistency

## Utility Functions

### Email Validation

The package uses a sophisticated email regex pattern for validation:

```regexp
^[a-zA-Z0-9](\.?[a-zA-Z0-9_\-+%]){0,63}@[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z]{2,})+$
```

**Supported Features:**
- Standard email formats
- Plus addressing (user+tag@domain.com)
- Subdomain emails
- International domain extensions
- Special characters in local part

**Validation Limits:**
- Local part: Maximum 64 characters
- Domain labels: Maximum 63 characters
- Minimum TLD length: 2 characters

## Error Handling

The package implements comprehensive error handling with different categories:

### Validation Errors

**Sources:**
- `validateOrganizationCreateRequest()`
- `validatePaginationParams()`
- `IsValidEmail()`
- `OrgStatus.IsValid()`

**Characteristics:**
- Detailed error messages
- Field-specific validation failures
- Business rule violations

**Examples:**
```go
"name is required"
"invalid email format: invalid@email"
"password must be at least 8 characters long"
"limit cannot exceed 1000"
```

### Database Errors

**Sources:**
- MongoDB connection/operation failures
- SQL constraint violations
- Transaction rollbacks
- Record not found scenarios

**Characteristics:**
- Wrapped with context information
- Logged with appropriate severity levels
- Include operation details for debugging

**Examples:**
```go
"organization not found with id: 12345"
"failed to hash password: <bcrypt error>"
"cannot scan <type> into FlexibleID"
```

### Context Errors

**Sources:**
- Context cancellation
- Timeout exceeded
- Deadline exceeded

**Characteristics:**
- Standard Go context errors
- Respected by all long-running operations
- Allow graceful cancellation

### Type Errors

**Sources:**
- FlexibleID type mismatches
- Unsupported data type scanning
- Invalid BSON/JSON formats

**Characteristics:**
- Include type information in error messages
- Logged for debugging purposes
- Fail fast to prevent data corruption

## Concurrency & Thread Safety

The package is designed for high-concurrency environments with comprehensive thread safety.

### Read-Write Mutex Protection

Each Organization instance contains a `sync.RWMutex` that protects:
- **Read Operations**: `GetIDString()`, `IsMongoID()`, `IsSerialID()`, `CheckPassword()`, `ToResponse()`
- **Write Operations**: `SetMongoID()`, `SetSerialID()`, `SetPassword()`, status change methods

### Locking Strategy

**Read Locks (RLock):**
- Allow multiple concurrent readers
- Used for data access without modification
- Automatically released when function returns

**Write Locks (Lock):**
- Exclusive access during modifications
- Used for field updates and status changes
- Prevent race conditions during updates

### Database-Level Concurrency

**MongoDB:**
- Uses MongoDB's native concurrency controls
- Document-level locking for updates
- Optimistic concurrency for reads

**SQL (GORM):**
- Transaction-based operations for consistency
- Database-level locking for updates
- Connection pooling for concurrent access

### Safe Usage Patterns

```go
// Safe concurrent reads
go func() {
    id := org.GetIDString() // Thread-safe
    isValid := org.CheckPassword("test") // Thread-safe
}()

// Safe concurrent writes  
go func() {
    org.SetPassword("newpass") // Thread-safe
    org.ActivateOrganization() // Thread-safe
}()

// Database operations are inherently safe
go func() {
    err := UpdateOrganizationSQL(ctx, req, org) // Safe
}()
```

## Logging System

The package uses structured logging with multiple verbosity levels and environment-specific filtering.

### Log Levels

**Error Level:**
- Database operation failures
- Validation errors that prevent operations
- Critical system failures
- Password hashing failures

**Warning Level:**
- Validation failures (invalid emails, short passwords)
- Operations on nil objects
- Organization suspensions and deletions
- Business rule violations

**Info Level:**
- Organization creation and major lifecycle events
- Successful operations completion
- Status changes and activations

**Debug Level:**
- Method entry/exit for development
- Detailed operation progress
- ID synchronization events

### Environment Tags

**prod**: Production environment logs
- Focus on errors and warnings
- Minimal debug information
- Security-sensitive operations

**dev**: Development environment logs
- Detailed debug information
- Method tracing
- Development-specific events

**minimal**: High-level production logs
- Critical errors only
- Major lifecycle events
- Security events

### Logging Context

All log entries include relevant context:
- Organization IDs for traceability
- Operation names and parameters
- Error details and stack traces
- Performance metrics where applicable

**Example Log Entries:**
```go
// Info level
utils.Log(zapcore.InfoLevel, "Organization created successfully", []string{"dev", "prod", "minimal"}, 
    zap.String("id", org.GetIDString()), zap.String("name", org.Name))

// Error level
utils.Log(zapcore.ErrorLevel, "Failed to hash password during organization creation", []string{"prod", "minimal"}, 
    zap.Error(err))

// Warning level
utils.Log(zapcore.WarnLevel, "Password check failed", []string{"prod"}, 
    zap.String("org_id", orgIDStr))
```

## Usage Examples

### Complete CRUD Workflow

```go
package main

import (
    "context"
    "log"
    "time"
    "your-module/models"
)

func main() {
    ctx := context.Background()
    
    // Create organization
    createReq := models.OrganizationCreate{
        Name:        "TechCorp Industries",
        Email:       "admin@techcorp.com",
        Status:      models.OrgStatusActive,
        Password:    "securepassword123",
        CreatedAt:   time.Now(),
        LastUpdated: time.Now(),
    }
    
    // For MongoDB
    org, err := models.CreateOrganizationMongo(ctx, createReq)
    if err != nil {
        log.Fatalf("Failed to create organization: %v", err)
    }
    
    log.Printf("Created organization: %s", org.GetIDString())
    
    // Read organization
    retrieved, err := models.GetOrganizationByIDMongo(ctx, org.ID)
    if err != nil {
        log.Fatalf("Failed to retrieve organization: %v", err)
    }
    
    // Verify password
    if retrieved.CheckPassword("securepassword123") {
        log.Println("Password verification successful")
    }
    
    // Update organization
    updateReq := models.OrganizationUpdate{
        Name:   stringPtr("TechCorp Industries Ltd"),
        Status: &models.OrgStatusSuspended,
    }
    
    err = models.UpdateOrganizationMongo(ctx, updateReq, org)
    if err != nil {
        log.Fatalf("Failed to update organization: %v", err)
    }
    
    // Get safe response for API
    response := org.ToResponse()
    log.Printf("API Response: %+v", response)
    
    // List organizations with pagination
    orgs, err := models.GetOrganizationsMongo(ctx, 10, 0)
    if err != nil {
        log.Fatalf("Failed to list organizations: %v", err)
    }
    
    log.Printf("Found %d organizations", len(orgs))
    
    // Soft delete
    org.SoftDeleteOrganization()
    err = models.UpdateOrganizationMongo(ctx, models.OrganizationUpdate{}, org)
    if err != nil {
        log.Fatalf("Failed to soft delete organization: %v", err)
    }
    
    // Hard delete
    err = models.DeleteOrganizationMongo(ctx, org)
    if err != nil {
        log.Fatalf("Failed to delete organization: %v", err)
    }
}

func stringPtr(s string) *string {
    return &s
}
```

### Advanced Password Management

```go
func demonstratePasswordManagement(org *models.Organization) {
    // Set initial password
    err := org.SetPassword("initialpassword123")
    if err != nil {
        log.Printf("Failed to set password: %v", err)
        return
    }
    
    // Verify correct password
    if org.CheckPassword("initialpassword123") {
        log.Println("Password verification successful")
    } else {
        log.Println("Password verification failed")
    }
    
    // Test incorrect password
    if !org.CheckPassword("wrongpassword") {
        log.Println("Incorrect password properly rejected")
    }
    
    // Change password
    err = org.SetPassword("newpassword456")
    if err != nil {
        log.Printf("Failed to change password: %v", err)
        return
    }
    
    // Verify old password no longer works
    if !org.CheckPassword("initialpassword123") {
        log.Println("Old password properly invalidated")
    }
    
    // Verify new password works
    if org.CheckPassword("newpassword456") {
        log.Println("New password works correctly")
    }
}
```

### Status Lifecycle Management

```go
func demonstrateStatusLifecycle(ctx context.Context, org *models.Organization) {
    // Start with active organization
    org.ActivateOrganization()
    err := models.UpdateOrganizationMongo(ctx, models.OrganizationUpdate{}, org)
    if err != nil {
        log.Printf("Failed to activate: %v", err)
        return
    }
    
    log.Printf("Organization status: %s", org.Status)
    
    // Suspend for policy violation
    org.SuspendOrganization()
    err = models.UpdateOrganizationMongo(ctx, models.OrganizationUpdate{}, org)
    if err != nil {
        log.Printf("Failed to suspend: %v", err)
        return
    }
    
    log.Printf("Organization suspended: %s", org.Status)
    
    // Reactivate after issue resolution
    org.ActivateOrganization()
    err = models.UpdateOrganizationMongo(ctx, models.OrganizationUpdate{}, org)
    if err != nil {
        log.Printf("Failed to reactivate: %v", err)
        return
    }
    
    log.Printf("Organization reactivated: %s", org.Status)
    
    // Soft delete when organization closes
    org.SoftDeleteOrganization()
    err = models.UpdateOrganizationMongo(ctx, models.OrganizationUpdate{}, org)
    if err != nil {
        log.Printf("Failed to soft delete: %v", err)
        return
    }
    
    log.Printf("Organization soft deleted: %s, DeletedAt: %v", org.Status, org.DeletedAt)
}
```

### Dual Database Usage

```go
func demonstrateDualDatabase(ctx context.Context) {
    createReq := models.OrganizationCreate{
        Name:        "Multi-DB Corp",
        Email:       "admin@multidb.com",
        Status:      models.OrgStatusActive,
        Password:    "securepass123",
        CreatedAt:   time.Now(),
        LastUpdated: time.Now(),
    }
    
    // Create in MongoDB
    mongoOrg, err := models.CreateOrganizationMongo(ctx, createReq)
    if err != nil {
        log.Printf("MongoDB creation failed: %v", err)
        return
    }
    
    // Create in SQL (different email to avoid unique constraint)
    createReq.Email = "admin-sql@multidb.com"
    sqlOrg, err := models.CreateOrganizationSQL(ctx, createReq)
    if err != nil {
        log.Printf("SQL creation failed: %v", err)
        return
    }
    
    log.Printf("MongoDB ID: %s (type: %v)", mongoOrg.GetIDString(), mongoOrg.IsMongoID())
    log.Printf("SQL ID: %s (type: %v)", sqlOrg.GetIDString(), sqlOrg.IsSerialID())
    
    // Both can be used with same API
    updateReq := models.OrganizationUpdate{
        Name: stringPtr("Updated Multi-DB Corp"),
    }
    
    // Update MongoDB version
    err = models.UpdateOrganizationMongo(ctx, updateReq, mongoOrg)
    if err != nil {
        log.Printf("MongoDB update failed: %v", err)
    }
    
    // Update SQL version  
    err = models.UpdateOrganizationSQL(ctx, updateReq, sqlOrg)
    if err != nil {
        log.Printf("SQL update failed: %v", err)
    }
    
    // Search operations work similarly
    foundMongo, err := models.GetOrganizationByEmailMongo(ctx, mongoOrg.Email)
    if err != nil {
        log.Printf("MongoDB search failed: %v", err)
    } else {
        log.Printf("Found in MongoDB: %s", foundMongo.Name)
    }
    
    foundSQL, err := models.GetOrganizationByEmailSQL(ctx, sqlOrg.Email)
    if err != nil {
        log.Printf("SQL search failed: %v", err)
    } else {
        log.Printf("Found in SQL: %s", foundSQL.Name)
    }
}
```

### Error Handling Patterns

```go
func handleOrganizationOperations(ctx context.Context) {
    createReq := models.OrganizationCreate{
        Name:        "Test Corp",
        Email:       "invalid-email", // Intentionally invalid
        Status:      models.OrgStatusActive,
        Password:    "short", // Intentionally too short
        CreatedAt:   time.Now(),
        LastUpdated: time.Now(),
    }
    
    // This will fail validation
    org, err := models.CreateOrganizationMongo(ctx, createReq)
    if err != nil {
        log.Printf("Expected validation error: %v", err)
        // Fix the validation errors
        createReq.Email = "admin@testcorp.com"
        createReq.Password = "securepassword123"
    }
    
    // Now creation should succeed
    org, err = models.CreateOrganizationMongo(ctx, createReq)
    if err != nil {
        log.Printf("Unexpected error: %v", err)
        return
    }
    
    // Try to find non-existent organization
    invalidID := models.NewMongoID()
    _, err = models.GetOrganizationByIDMongo(ctx, invalidID)
    if err != nil {
        log.Printf("Expected not found error: %v", err)
    }
    
    // Try invalid pagination
    _, err = models.GetOrganizationsMongo(ctx, -1, 0)
    if err != nil {
        log.Printf("Expected pagination error: %v", err)
    }
    
    // Try to set invalid password
    err = org.SetPassword("short")
    if err != nil {
        log.Printf("Expected password error: %v", err)
    }
}
```

## Performance Considerations

### Database Performance

**MongoDB:**
- Uses single document operations for atomicity
- Leverages MongoDB's native ObjectID indexing
- Cursor-based pagination for large result sets
- Proper connection pooling via driver

**SQL:**
- Transaction-based operations ensure consistency
- Primary key indexing on auto-increment IDs
- Unique constraints on email field for performance
- GORM connection pooling and prepared statements

### Memory Management

**FlexibleID:**
- Only one ID field populated at a time
- Minimal memory overhead for hybrid system
- Efficient string conversion caching

**Organization:**
- Single mutex per instance minimizes lock contention
- Read locks allow concurrent access
- Password hashes stored efficiently

### Logging Performance

**Structured Logging:**
- Environment-based filtering reduces overhead
- Lazy evaluation of log messages
- Efficient JSON serialization via Zap

**Log Level Management:**
- Debug logs filtered out in production
- Context-aware logging reduces volume
- Error aggregation prevents log spam

### Concurrency Performance

**Read-Heavy Workloads:**
- Read locks allow multiple concurrent readers
- Thread-safe methods prevent contention
- Efficient password verification

**Write Operations:**
- Exclusive locks only during modifications
- Database-level optimizations for updates
- Transaction batching where appropriate

## Dependencies and Requirements

### Required Go Packages

```go
import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "regexp"
    "strconv"
    "strings"
    "sync"
    "time"
    "database/sql/driver"
    
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)
```

### Configuration Dependencies

The package expects these configuration objects to be properly initialized:

**config.MongoDB:**
- Type: `*mongo.Database`
- Must be connected and authenticated
- Should have proper indexes on organizations collection

**config.DB:**
- Type: `*gorm.DB`
- Must be connected with proper driver (PostgreSQL, MySQL, SQLite)
- Should have organizations table with proper schema

### Database Schema Requirements

**MongoDB Collections:**
```javascript
// organizations collection
{
  "_id": ObjectId,
  "name": String, // required, 3-100 chars
  "email": String, // required, unique, valid format
  "password": String, // required, BCrypt hash
  "status": String, // required, enum: active|inactive|suspended|deleted
  "created_at": Date, // required
  "last_updated": Date, // required
  "deleted_at": Date // optional
}

// Recommended indexes
db.organizations.createIndex({ "email": 1 }, { unique: true })
db.organizations.createIndex({ "status": 1 })
db.organizations.createIndex({ "created_at": 1 })
db.organizations.createIndex({ "name": 1 })
```

**SQL Table Schema:**
```sql
CREATE TABLE organizations (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password TEXT NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'inactive', 'suspended', 'deleted')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- Recommended indexes
CREATE INDEX idx_organizations_email ON organizations(email);
CREATE INDEX idx_organizations_status ON organizations(status);
CREATE INDEX idx_organizations_created_at ON organizations(created_at);
CREATE INDEX idx_organizations_name ON organizations(name);
```

### Minimum Go Version

- **Go 1.19+** required for generics and latest context features
- **Go 1.20+** recommended for optimal performance

### Environment Setup

**Development:**
```bash
# Install dependencies
go mod tidy

# Set up logging configuration
export LOG_LEVEL=debug
export LOG_ENV=dev

# Configure database connections
export MONGODB_URI="mongodb://localhost:27017/myapp"
export SQL_DB_DSN="postgres://user:password@localhost/myapp?sslmode=disable"
```

**Production:**
```bash
# Production logging
export LOG_LEVEL=info
export LOG_ENV=prod

# Production database connections with proper security
export MONGODB_URI="mongodb+srv://user:password@cluster.mongodb.net/myapp"
export SQL_DB_DSN="postgres://user:password@db.example.com/myapp?sslmode=require"
```

## Testing Considerations

### Unit Testing

**Mocking Requirements:**
- Mock `config.MongoDB` for MongoDB operations
- Mock `config.DB` for SQL operations  
- Mock logging functions for test isolation

**Test Categories:**
1. **Constructor Tests**: Validate organization creation with different ID types
2. **Validation Tests**: Test all validation functions with edge cases
3. **Method Tests**: Test all organization methods with concurrent access
4. **Conversion Tests**: Test JSON/BSON marshaling and unmarshaling
5. **Error Handling Tests**: Test all error scenarios and edge cases

**Example Test Structure:**
```go
func TestOrganizationCreation(t *testing.T) {
    tests := []struct {
        name        string
        req         OrganizationCreate
        idType      IdType
        expectedErr bool
    }{
        {
            name: "valid mongo organization",
            req: OrganizationCreate{
                Name:        "Test Corp",
                Email:       "test@example.com",
                Status:      OrgStatusActive,
                Password:    "password123",
                CreatedAt:   time.Now(),
                LastUpdated: time.Now(),
            },
            idType:      MongoIdType,
            expectedErr: false,
        },
        // Add more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            org, err := NewOrganization(ctx, tt.req, tt.idType)
            
            if tt.expectedErr {
                assert.Error(t, err)
                assert.Nil(t, org)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, org)
                assert.Equal(t, tt.req.Name, org.Name)
            }
        })
    }
}
```

### Integration Testing

**Database Setup:**
- Use test databases (separate from development/production)
- Implement database seeding and cleanup
- Test both MongoDB and SQL backends

**Concurrent Testing:**
```go
func TestConcurrentPasswordUpdates(t *testing.T) {
    org := createTestOrganization(t)
    
    var wg sync.WaitGroup
    errors := make(chan error, 10)
    
    // Simulate concurrent password updates
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(password string) {
            defer wg.Done()
            err := org.SetPassword(password)
            if err != nil {
                errors <- err
            }
        }(fmt.Sprintf("password%d", i))
    }
    
    wg.Wait()
    close(errors)
    
    // Verify no race conditions occurred
    for err := range errors {
        assert.NoError(t, err)
    }
}
```

### Performance Testing

**Benchmark Tests:**
```go
func BenchmarkPasswordHashing(b *testing.B) {
    org := &Organization{}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        err := org.SetPassword("benchmarkpassword123")
        if err != nil {
            b.Fatalf("Password hashing failed: %v", err)
        }
    }
}

func BenchmarkPasswordVerification(b *testing.B) {
    org := &Organization{}
    org.SetPassword("benchmarkpassword123")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        valid := org.CheckPassword("benchmarkpassword123")
        if !valid {
            b.Fatalf("Password verification failed")
        }
    }
}
```

## Migration and Upgrade Considerations

### Database Migrations

**From Single Database to Dual Database:**
```go
func MigrateToFlexibleID(ctx context.Context) error {
    // Step 1: Add FlexibleID support to existing records
    var orgs []*Organization
    err := config.DB.Find(&orgs).Error
    if err != nil {
        return err
    }
    
    for _, org := range orgs {
        if org.ID == nil && org.SerialID != nil {
            org.ID = NewSerialID(*org.SerialID)
            err = config.DB.Save(org).Error
            if err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

**Adding New Status Values:**
```go
// Add new status to enum
const (
    OrgStatusPending   OrgStatus = "pending"   // New status
    OrgStatusArchived  OrgStatus = "archived"  // New status
)

// Update IsValid method
func (s OrgStatus) IsValid() bool {
    switch s {
    case OrgStatusActive, OrgStatusInactive, OrgStatusSuspended, 
         OrgStatusDeleted, OrgStatusPending, OrgStatusArchived:
        return true
    }
    return false
}
```

### Version Compatibility

**Backward Compatibility:**
- Existing FlexibleID JSON format remains supported
- Database schema changes are additive only
- Old status values continue to work

**Forward Compatibility:**
- New fields added as optional with sensible defaults
- Version detection in JSON unmarshaling
- Graceful degradation for unknown enum values

## Best Practices

### Security Best Practices

**Password Management:**
```go
// DO: Use strong passwords and proper validation
func (o *Organization) SetSecurePassword(password string) error {
    if len(password) < 12 {
        return errors.New("password must be at least 12 characters")
    }
    if !hasUppercase(password) || !hasLowercase(password) || !hasDigit(password) {
        return errors.New("password must contain uppercase, lowercase, and digits")
    }
    return o.SetPassword(password)
}

// DON'T: Store plain text passwords or log them
func (o *Organization) UnsafePasswordHandling(password string) {
    // NEVER DO THIS
    o.Password = password  // Plain text storage
    log.Printf("Password: %s", password)  // Logging passwords
}
```

**Data Access Patterns:**
```go
// DO: Always use ToResponse() for API responses
func GetOrganizationAPI(ctx context.Context, id string) (*OrganizationResponse, error) {
    org, err := GetOrganizationByIDMongo(ctx, parseID(id))
    if err != nil {
        return nil, err
    }
    return org.ToResponse(), nil  // Safe response
}

// DON'T: Return raw organization with password
func UnsafeGetOrganization(ctx context.Context, id string) (*Organization, error) {
    return GetOrganizationByIDMongo(ctx, parseID(id))  // Exposes password hash
}
```

### Performance Best Practices

**Database Operations:**
```go
// DO: Use context with timeouts for database operations
func SafeDatabaseOperation(id string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    org, err := GetOrganizationByIDMongo(ctx, parseID(id))
    if err != nil {
        return err
    }
    
    return UpdateOrganizationMongo(ctx, updateReq, org)
}

// DO: Use pagination for list operations
func GetOrganizationsPaginated(page, size int) ([]*Organization, error) {
    if size > 100 {
        size = 100  // Limit page size
    }
    offset := (page - 1) * size
    return GetOrganizationsMongo(ctx, size, offset)
}
```

**Concurrency Patterns:**
```go
// DO: Use organization methods for thread-safe operations
func SafeConcurrentAccess(org *Organization) {
    var wg sync.WaitGroup
    
    // Multiple readers
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            id := org.GetIDString()  // Thread-safe read
            log.Printf("Organization ID: %s", id)
        }()
    }
    
    wg.Wait()
}

// DON'T: Access fields directly without locks
func UnsafeConcurrentAccess(org *Organization) {
    go func() {
        // UNSAFE: Direct field access without locks
        if org.ID != nil {
            log.Printf("ID: %s", org.ID.String())
        }
    }()
}
```

### Error Handling Best Practices

**Comprehensive Error Handling:**
```go
func RobustOrganizationOperation(ctx context.Context, req OrganizationCreate) (*Organization, error) {
    // Validate input
    if err := validateInput(req); err != nil {
        return nil, fmt.Errorf("input validation failed: %w", err)
    }
    
    // Create with timeout
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    
    org, err := CreateOrganizationMongo(ctx, req)
    if err != nil {
        // Log error with context
        utils.Log(zapcore.ErrorLevel, "Organization creation failed", 
            []string{"prod", "minimal"}, 
            zap.String("email", req.Email),
            zap.Error(err))
        
        // Return user-friendly error
        if isValidationError(err) {
            return nil, fmt.Errorf("invalid organization data: %w", err)
        }
        if isDuplicateError(err) {
            return nil, fmt.Errorf("organization with email %s already exists", req.Email)
        }
        return nil, fmt.Errorf("failed to create organization: %w", err)
    }
    
    return org, nil
}
```

### Logging Best Practices

**Structured Logging:**
```go
// DO: Use structured logging with appropriate levels
func LogOrganizationEvent(org *Organization, event string, level zapcore.Level) {
    utils.Log(level, event, []string{"prod", "dev"}, 
        zap.String("org_id", org.GetIDString()),
        zap.String("org_name", org.Name),
        zap.String("org_status", string(org.Status)),
        zap.Time("timestamp", time.Now()))
}

// DON'T: Use unstructured logging or log sensitive data
func BadLogging(org *Organization) {
    // AVOID: Unstructured logging
    log.Printf("Organization: %+v", org)  // May expose password hash
    
    // AVOID: Logging sensitive information
    log.Printf("Password hash: %s", org.Password)  // Security risk
}
```

## Advanced Usage Patterns

### Custom Validation

**Extended Validation:**
```go
type ExtendedOrganizationCreate struct {
    OrganizationCreate
    Domain        string   `json:"domain"`
    AllowedEmails []string `json:"allowed_emails"`
}

func ValidateExtendedOrganization(req ExtendedOrganizationCreate) error {
    // Base validation
    if err := validateOrganizationCreateRequest(req.OrganizationCreate); err != nil {
        return err
    }
    
    // Extended validation
    if req.Domain != "" {
        if !strings.HasSuffix(req.Email, "@"+req.Domain) {
            return errors.New("email must belong to organization domain")
        }
    }
    
    if len(req.AllowedEmails) > 0 {
        allowed := false
        for _, email := range req.AllowedEmails {
            if req.Email == email {
                allowed = true
                break
            }
        }
        if !allowed {
            return errors.New("email not in allowed list")
        }
    }
    
    return nil
}
```

### Custom Status Workflows

**Workflow Management:**
```go
type StatusTransition struct {
    From OrgStatus
    To   OrgStatus
    ValidTransitions map[OrgStatus][]OrgStatus
}

func NewStatusTransition() *StatusTransition {
    return &StatusTransition{
        ValidTransitions: map[OrgStatus][]OrgStatus{
            OrgStatusActive: {OrgStatusInactive, OrgStatusSuspended, OrgStatusDeleted},
            OrgStatusInactive: {OrgStatusActive, OrgStatusDeleted},
            OrgStatusSuspended: {OrgStatusActive, OrgStatusDeleted},
            OrgStatusDeleted: {}, // Terminal state
        },
    }
}

func (st *StatusTransition) CanTransition(from, to OrgStatus) bool {
    validStates, exists := st.ValidTransitions[from]
    if !exists {
        return false
    }
    
    for _, validState := range validStates {
        if validState == to {
            return true
        }
    }
    return false
}

func (o *Organization) TransitionStatus(newStatus OrgStatus, st *StatusTransition) error {
    if !st.CanTransition(o.Status, newStatus) {
        return fmt.Errorf("invalid status transition from %s to %s", o.Status, newStatus)
    }
    
    switch newStatus {
    case OrgStatusActive:
        o.ActivateOrganization()
    case OrgStatusInactive:
        o.DeactivateOrganization()
    case OrgStatusSuspended:
        o.SuspendOrganization()
    case OrgStatusDeleted:
        o.SoftDeleteOrganization()
    }
    
    return nil
}
```

### Audit Trail Implementation

**Audit Logging:**
```go
type OrganizationAudit struct {
    OrganizationID string    `json:"organization_id"`
    Action         string    `json:"action"`
    OldValues      string    `json:"old_values,omitempty"`
    NewValues      string    `json:"new_values,omitempty"`
    UserID         string    `json:"user_id,omitempty"`
    Timestamp      time.Time `json:"timestamp"`
    IPAddress      string    `json:"ip_address,omitempty"`
}

func AuditOrganizationChange(org *Organization, action string, oldOrg *Organization, userID, ipAddress string) {
    audit := OrganizationAudit{
        OrganizationID: org.GetIDString(),
        Action:         action,
        UserID:         userID,
        Timestamp:      time.Now(),
        IPAddress:      ipAddress,
    }
    
    if oldOrg != nil {
        oldJSON, _ := json.Marshal(oldOrg.ToResponse())
        audit.OldValues = string(oldJSON)
    }
    
    newJSON, _ := json.Marshal(org.ToResponse())
    audit.NewValues = string(newJSON)
    
    // Log audit event
    utils.Log(zapcore.InfoLevel, "Organization audit event", 
        []string{"prod", "audit"},
        zap.String("org_id", audit.OrganizationID),
        zap.String("action", audit.Action),
        zap.String("user_id", audit.UserID),
        zap.String("ip_address", audit.IPAddress))
    
    // Store in audit collection/table
    storeAuditEvent(audit)
}

// Usage with updates
func AuditedUpdateOrganization(ctx context.Context, req OrganizationUpdate, org *Organization, userID, ipAddress string) error {
    // Create backup for audit
    oldOrg := *org
    
    // Perform update
    err := UpdateOrganizationMongo(ctx, req, org)
    if err != nil {
        return err
    }
    
    // Log audit trail
    AuditOrganizationChange(org, "update", &oldOrg, userID, ipAddress)
    
    return nil
}
```