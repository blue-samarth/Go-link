# Go Models Package Documentation

## Overview
This package provides flexible data models for organizations that can work with both MongoDB and SQL databases. It includes a flexible ID system and comprehensive organization management functionality.

## Types

### IdType
```go
type IdType string
```
**Description**: Enum type for identifying ID storage types.

**Constants**:
- `MongoIdType`: "mongo" - Uses MongoDB ObjectID
- `SerialIdType`: "serial" - Uses SQL serial/integer ID

---

### FlexibleID
```go
type FlexibleID struct {
    Type     IdType               `json:"type" bson:"type"`
    MongoID  *primitive.ObjectID  `json:"mongo_id,omitempty" bson:"mongo_id,omitempty"`
    SerialID *int64               `json:"serial_id,omitempty" bson:"serial_id,omitempty"`
}
```
**Description**: A flexible ID type that can store either MongoDB ObjectID or SQL serial ID.

---

### OrgStatus
```go
type OrgStatus string
```
**Description**: Enum type for organization status.

**Constants**:
- `OrgStatusActive`: "active"
- `OrgStatusInactive`: "inactive"
- `OrgStatusSuspended`: "suspended"
- `OrgStatusDeleted`: "deleted"

---

### Organization
```go
type Organization struct {
    ID          *FlexibleID `json:"id" bson:"_id"`
    Name        string      `json:"name" bson:"name"`
    Email       string      `json:"email" bson:"email"`
    Password    string      `json:"-" bson:"password"`
    Status      OrgStatus   `json:"status" bson:"status"`
    CreatedAt   time.Time   `json:"created_at" bson:"created_at"`
    LastUpdated time.Time   `json:"last_updated" bson:"last_updated"`
    DeletedAt   *time.Time  `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}
```
**Description**: Main organization model with flexible ID support.

---

### OrganizationCreate
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
**Description**: Request struct for creating new organizations.

---

### OrganizationUpdate
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
**Description**: Request struct for updating organizations with optional fields.

---

### OrganizationResponse
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
**Description**: Response struct for API returns, excludes password field.

## FlexibleID Functions

### NewMongoID
```go
func NewMongoID() *FlexibleID
```
**Description**: Creates a new FlexibleID with a fresh MongoDB ObjectID.
**Arguments**: None
**Returns**: `*FlexibleID` - New FlexibleID with MongoDB type

### FromMongoID
```go
func FromMongoID(id primitive.ObjectID) *FlexibleID
```
**Description**: Creates a FlexibleID from an existing MongoDB ObjectID.
**Arguments**: 
- `id primitive.ObjectID` - Existing MongoDB ObjectID
**Returns**: `*FlexibleID` - FlexibleID containing the provided ObjectID

### NewSerialID
```go
func NewSerialID(id int64) *FlexibleID
```
**Description**: Creates a new FlexibleID with a serial ID.
**Arguments**: 
- `id int64` - Serial ID value
**Returns**: `*FlexibleID` - New FlexibleID with serial type

### FromSerialID
```go
func FromSerialID(id int64) *FlexibleID
```
**Description**: Creates a FlexibleID from an existing serial ID.
**Arguments**: 
- `id int64` - Existing serial ID
**Returns**: `*FlexibleID` - FlexibleID containing the provided serial ID

## FlexibleID Methods

### String
```go
func (f *FlexibleID) String() string
```
**Description**: Returns string representation of the ID.
**Returns**: `string` - Hex string for MongoDB ObjectID or decimal string for serial ID

### IsEmpty
```go
func (f *FlexibleID) IsEmpty() bool
```
**Description**: Checks if the FlexibleID is empty or nil.
**Returns**: `bool` - true if empty, false otherwise

### GetValue
```go
func (f *FlexibleID) GetValue() interface{}
```
**Description**: Returns the underlying ID value.
**Returns**: `interface{}` - primitive.ObjectID for mongo, int64 for serial, nil if empty

### Value
```go
func (f *FlexibleID) Value() (driver.Value, error)
```
**Description**: SQL driver Value method for database operations.
**Returns**: 
- `driver.Value` - Database-compatible value
- `error` - Error if conversion fails

### Scan
```go
func (f *FlexibleID) Scan(value interface{}) error
```
**Description**: SQL driver Scan method for reading from database.
**Arguments**: 
- `value interface{}` - Value from database
**Returns**: `error` - Error if scan fails

### MarshalJSON
```go
func (f *FlexibleID) MarshalJSON() ([]byte, error)
```
**Description**: Custom JSON marshaling.
**Returns**: 
- `[]byte` - JSON bytes
- `error` - Marshaling error

### UnmarshalJSON
```go
func (f *FlexibleID) UnmarshalJSON(data []byte) error
```
**Description**: Custom JSON unmarshaling.
**Arguments**: 
- `data []byte` - JSON data to unmarshal
**Returns**: `error` - Unmarshaling error

### MarshalBSONValue
```go
func (f *FlexibleID) MarshalBSONValue() (bson.ValueType, []byte, error)
```
**Description**: Custom BSON marshaling for MongoDB.
**Returns**: 
- `bson.ValueType` - BSON type
- `[]byte` - BSON bytes  
- `error` - Marshaling error

### UnmarshalBSONValue
```go
func (f *FlexibleID) UnmarshalBSONValue(t bson.ValueType, data []byte) error
```
**Description**: Custom BSON unmarshaling for MongoDB.
**Arguments**: 
- `t bson.ValueType` - BSON value type
- `data []byte` - BSON data
**Returns**: `error` - Unmarshaling error

## OrgStatus Methods

### IsValid
```go
func (s OrgStatus) IsValid() bool
```
**Description**: Validates if the OrgStatus value is one of the defined constants.
**Returns**: `bool` - true if valid, false otherwise

### Value
```go
func (s OrgStatus) Value() (driver.Value, error)
```
**Description**: SQL driver Value method for database operations.
**Returns**: 
- `driver.Value` - String value for database
- `error` - Always nil for this implementation

### Scan
```go
func (s *OrgStatus) Scan(value interface{}) error
```
**Description**: SQL driver Scan method for reading from database.
**Arguments**: 
- `value interface{}` - Value from database
**Returns**: `error` - Error if value is invalid

## Email Validation Function

### IsValidEmail
```go
func (email *string) IsValidEmail() bool
```
**Description**: Validates email format using regex pattern.
**Returns**: `bool` - true if email format is valid, false otherwise

## Organization Constructor Functions

### NewOrganization
```go
func NewOrganization(req OrganizationCreate, idType IdType, serialID ...int64) (*Organization, error)
```
**Description**: Creates a new Organization with specified ID type.
**Arguments**: 
- `req OrganizationCreate` - Organization creation request
- `idType IdType` - Type of ID to use (mongo or serial)
- `serialID ...int64` - Optional serial ID (required for SerialIdType)
**Returns**: 
- `*Organization` - Created organization
- `error` - Validation or creation error

### NewOrganizationForMongo
```go
func NewOrganizationForMongo(req OrganizationCreate) (*Organization, error)
```
**Description**: Creates a new Organization with MongoDB ObjectID.
**Arguments**: 
- `req OrganizationCreate` - Organization creation request
**Returns**: 
- `*Organization` - Created organization with MongoDB ID
- `error` - Validation or creation error

### NewOrganizationForSQL
```go
func NewOrganizationForSQL(req OrganizationCreate, id ...int64) (*Organization, error)
```
**Description**: Creates a new Organization with SQL serial ID.
**Arguments**: 
- `req OrganizationCreate` - Organization creation request
- `id ...int64` - Optional serial ID values
**Returns**: 
- `*Organization` - Created organization with serial ID
- `error` - Validation or creation error

## Organization Methods

### SetMongoID
```go
func (o *Organization) SetMongoID(id primitive.ObjectID)
```
**Description**: Sets the organization ID to a MongoDB ObjectID.
**Arguments**: 
- `id primitive.ObjectID` - MongoDB ObjectID to set

### SetSerialID
```go
func (o *Organization) SetSerialID(id int64)
```
**Description**: Sets the organization ID to a serial ID.
**Arguments**: 
- `id int64` - Serial ID to set

### GetIDString
```go
func (o *Organization) GetIDString() string
```
**Description**: Returns the ID as a string representation.
**Returns**: `string` - String representation of the ID

### IsMongoID
```go
func (o *Organization) IsMongoID() bool
```
**Description**: Checks if the organization uses MongoDB ObjectID.
**Returns**: `bool` - true if using MongoDB ID, false otherwise

### IsSerialID
```go
func (o *Organization) IsSerialID() bool
```
**Description**: Checks if the organization uses serial ID.
**Returns**: `bool` - true if using serial ID, false otherwise

### SetPassword
```go
func (o *Organization) SetPassword(password string) error
```
**Description**: Sets a new password with bcrypt hashing and updates LastUpdated timestamp.
**Arguments**: 
- `password string` - Plain text password (minimum 8 characters)
**Returns**: `error` - Error if validation fails or hashing fails

### CheckPassword
```go
func (o *Organization) CheckPassword(password string) bool
```
**Description**: Verifies a password against the stored hash.
**Arguments**: 
- `password string` - Plain text password to check
**Returns**: `bool` - true if password matches, false otherwise

### ActivateOrganization
```go
func (o *Organization) ActivateOrganization()
```
**Description**: Sets organization status to active and updates LastUpdated timestamp.

### DeactivateOrganization
```go
func (o *Organization) DeactivateOrganization()
```
**Description**: Sets organization status to inactive and updates LastUpdated timestamp.

### SuspendOrganization
```go
func (o *Organization) SuspendOrganization()
```
**Description**: Sets organization status to suspended and updates LastUpdated timestamp.

### SoftDeleteOrganization
```go
func (o *Organization) SoftDeleteOrganization()
```
**Description**: Sets organization status to deleted, sets DeletedAt timestamp, and updates LastUpdated.

### ToResponse
```go
func (o *Organization) ToResponse() *OrganizationResponse
```
**Description**: Converts Organization to OrganizationResponse, excluding sensitive data like password.
**Returns**: `*OrganizationResponse` - Response struct suitable for API returns

### TableName
```go
func (Organization) TableName() string
```
**Description**: Returns the SQL table name for the Organization model.
**Returns**: `string` - "organizations"

### CollectionName
```go
func (Organization) CollectionName() string
```
**Description**: Returns the MongoDB collection name for the Organization model.
**Returns**: `string` - "organizations"

### BeforeUpdate
```go
func (o *Organization) BeforeUpdate()
```
**Description**: Hook method that updates LastUpdated timestamp and manages DeletedAt field based on status. Called before database updates.

## Validation Functions

### validateOrganizationCreateRequest
```go
func validateOrganizationCreateRequest(req OrganizationCreate) error
```
**Description**: Validates OrganizationCreate request data.
**Arguments**: 
- `req OrganizationCreate` - Request to validate
**Returns**: `error` - Validation error with specific message, nil if valid

**Validation Rules**:
- Name: Required, non-empty after trimming whitespace
- Email: Required, valid email format
- Status: Required, must be valid OrgStatus value
- CreatedAt: Required, non-zero time
- LastUpdated: Required, non-zero time
- DeletedAt: Can only be set if status is 'deleted'
- Password: Required, minimum 8 characters

## Dependencies

**Required Imports**:
- `time` - Time handling
- `encoding/json` - JSON marshaling/unmarshaling
- `errors` - Error handling
- `regexp` - Email validation
- `strings` - String operations
- `fmt` - String formatting
- `strconv` - String conversions
- `go.mongodb.org/mongo-driver/bson` - MongoDB BSON support
- `go.mongodb.org/mongo-driver/bson/primitive` - MongoDB ObjectID
- `database/sql/driver` - SQL driver interface
- `golang.org/x/crypto/bcrypt` - Password hashing

## Usage Examples

### Creating Organizations
```go
// For MongoDB
req := OrganizationCreate{
    Name: "Example Corp",
    Email: "contact@example.com",
    Status: OrgStatusActive,
    Password: "securepassword123",
    CreatedAt: time.Now(),
    LastUpdated: time.Now(),
}
org, err := NewOrganizationForMongo(req)

// For SQL
org, err := NewOrganizationForSQL(req, 1) // with specific ID
```

### Working with IDs
```go
// Create FlexibleID
mongoID := NewMongoID()
serialID := NewSerialID(123)

// Check ID type
if org.IsMongoID() {
    fmt.Println("Using MongoDB ID:", org.GetIDString())
}
```

### Password Management
```go
// Set password
err := org.SetPassword("newpassword123")

// Check password
isValid := org.CheckPassword("newpassword123")
```

### Status Management
```go
org.ActivateOrganization()
org.SuspendOrganization()
org.SoftDeleteOrganization()
```