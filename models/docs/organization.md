# Go Models Package Documentation

## Overview
This package provides flexible data models that support both MongoDB and SQL databases with a unified interface for ID handling and organization management.

## Types

### IdType
```go
type IdType string
```
**Description**: Enumeration for ID type specifications.

**Constants**:
- `MongoIdType`: "mongo" - Indicates MongoDB ObjectID usage
- `SerialIdType`: "serial" - Indicates SQL auto-increment ID usage

### FlexibleID
```go
type FlexibleID struct {
    Type     IdType               `json:"type" bson:"type"`
    MongoID  *primitive.ObjectID  `json:"mongo_id,omitempty" bson:"mongo_id,omitempty"`
    SerialID *int64               `json:"serial_id,omitempty" bson:"serial_id,omitempty"`
}
```
**Description**: Unified ID structure supporting both MongoDB ObjectIDs and SQL serial IDs.

### OrgStatus
```go
type OrgStatus string
```
**Description**: Enumeration for organization status values.

**Constants**:
- `OrgStatusActive`: "active"
- `OrgStatusInactive`: "inactive"
- `OrgStatusSuspended`: "suspended"
- `OrgStatusDeleted`: "deleted"

### Organization
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
}
```
**Description**: Main organization entity with flexible ID support and database compatibility.

### OrganizationCreate
```go
type OrganizationCreate struct {
    Name        string    `json:"name" validate:"required,min=3,max=100"`
    Email       string    `json:"email" validate:"required,email"`
    Status      OrgStatus `json:"status" validate:"required,oneof=active inactive suspended deleted"`
    Password    string    `json:"password" validate:"required,min=8"`
    CreatedAt   time.Time `json:"created_at"`
    LastUpdated time.Time `json:"last_updated"`
    DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}
```
**Description**: Data transfer object for creating new organizations.

### OrganizationUpdate
```go
type OrganizationUpdate struct {
    Name        *string   `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
    Email       *string   `json:"email,omitempty" validate:"omitempty,email"`
    Status      *OrgStatus `json:"status,omitempty" validate:"omitempty,oneof=active inactive suspended deleted"`
    Password    *string   `json:"password,omitempty" validate:"omitempty,min=8"`
    CreatedAt   *time.Time `json:"created_at,omitempty"`
    LastUpdated *time.Time `json:"last_updated,omitempty"`
    DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}
```
**Description**: Data transfer object for updating existing organizations with optional fields.

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
**Description**: Data transfer object for API responses, excludes sensitive data like passwords.

## FlexibleID Functions

### NewMongoID
```go
func NewMongoID() *FlexibleID
```
**Description**: Creates a new FlexibleID with a generated MongoDB ObjectID.

**Arguments**: None

**Returns**: 
- `*FlexibleID`: New FlexibleID instance with MongoIdType and generated ObjectID

### FromMongoID
```go
func FromMongoID(id primitive.ObjectID) *FlexibleID
```
**Description**: Creates a FlexibleID from an existing MongoDB ObjectID.

**Arguments**:
- `id primitive.ObjectID`: Existing MongoDB ObjectID

**Returns**: 
- `*FlexibleID`: FlexibleID instance containing the provided ObjectID

### NewSerialID
```go
func NewSerialID(id int64) *FlexibleID
```
**Description**: Creates a FlexibleID with a serial ID value.

**Arguments**:
- `id int64`: Serial ID value

**Returns**: 
- `*FlexibleID`: FlexibleID instance with SerialIdType and provided ID

### FromSerialID
```go
func FromSerialID(id int64) *FlexibleID
```
**Description**: Creates a FlexibleID from an existing serial ID.

**Arguments**:
- `id int64`: Existing serial ID value

**Returns**: 
- `*FlexibleID`: FlexibleID instance containing the provided serial ID

## FlexibleID Methods

### String
```go
func (f *FlexibleID) String() string
```
**Description**: Returns string representation of the ID.

**Arguments**: None (receiver method)

**Returns**: 
- `string`: Hex string for MongoDB ObjectID, decimal string for serial ID, empty string if nil/empty

### IsEmpty
```go
func (f *FlexibleID) IsEmpty() bool
```
**Description**: Checks if the FlexibleID is empty or nil.

**Arguments**: None (receiver method)

**Returns**: 
- `bool`: true if ID is nil, zero ObjectID, or zero serial ID

### GetValue
```go
func (f *FlexibleID) GetValue() interface{}
```
**Description**: Returns the underlying ID value as interface{}.

**Arguments**: None (receiver method)

**Returns**: 
- `interface{}`: ObjectID for mongo type, int64 for serial type, nil if empty

### Value
```go
func (f *FlexibleID) Value() (driver.Value, error)
```
**Description**: SQL driver.Valuer interface implementation for database storage.

**Arguments**: None (receiver method)

**Returns**: 
- `driver.Value`: Database-compatible value (int64 for serial, hex string for mongo)
- `error`: Error if conversion fails

### Scan
```go
func (f *FlexibleID) Scan(value interface{}) error
```
**Description**: SQL driver.Scanner interface implementation for database retrieval.

**Arguments**:
- `value interface{}`: Value from database (int64, string, or []byte)

**Returns**: 
- `error`: Error if scanning fails or value is invalid

### MarshalJSON
```go
func (f *FlexibleID) MarshalJSON() ([]byte, error)
```
**Description**: JSON marshaling implementation.

**Arguments**: None (receiver method)

**Returns**: 
- `[]byte`: JSON bytes with type and value fields
- `error`: Error if marshaling fails

### UnmarshalJSON
```go
func (f *FlexibleID) UnmarshalJSON(data []byte) error
```
**Description**: JSON unmarshaling implementation.

**Arguments**:
- `data []byte`: JSON data to unmarshal

**Returns**: 
- `error`: Error if unmarshaling fails or data is invalid

### MarshalBSONValue
```go
func (f *FlexibleID) MarshalBSONValue() (bson.ValueType, []byte, error)
```
**Description**: BSON marshaling implementation for MongoDB storage.

**Arguments**: None (receiver method)

**Returns**: 
- `bson.ValueType`: BSON type (ObjectID or Int64)
- `[]byte`: BSON-encoded bytes
- `error`: Error if marshaling fails

### UnmarshalBSONValue
```go
func (f *FlexibleID) UnmarshalBSONValue(t bson.ValueType, data []byte) error
```
**Description**: BSON unmarshaling implementation for MongoDB retrieval.

**Arguments**:
- `t bson.ValueType`: BSON value type
- `data []byte`: BSON data to unmarshal

**Returns**: 
- `error`: Error if unmarshaling fails or type is unsupported

## OrgStatus Methods

### IsValid
```go
func (s OrgStatus) IsValid() bool
```
**Description**: Validates if the OrgStatus value is one of the defined constants.

**Arguments**: None (receiver method)

**Returns**: 
- `bool`: true if status is valid (active, inactive, suspended, deleted)

### Value
```go
func (s OrgStatus) Value() (driver.Value, error)
```
**Description**: SQL driver.Valuer interface implementation.

**Arguments**: None (receiver method)

**Returns**: 
- `driver.Value`: String representation of the status
- `error`: Always nil for this implementation

### Scan
```go
func (s *OrgStatus) Scan(value interface{}) error
```
**Description**: SQL driver.Scanner interface implementation.

**Arguments**:
- `value interface{}`: Database value (string or []byte)

**Returns**: 
- `error`: Error if value is invalid or cannot be scanned

## Validation Functions

### IsValidEmail
```go
func (email *string) IsValidEmail() bool
```
**Description**: Validates email format using regex pattern.

**Arguments**: None (receiver method on *string)

**Returns**: 
- `bool`: true if email matches valid email regex pattern

### validateOrganizationCreateRequest
```go
func validateOrganizationCreateRequest(req OrganizationCreate) error
```
**Description**: Validates all fields in OrganizationCreate request.

**Arguments**:
- `req OrganizationCreate`: Organization creation request to validate

**Returns**: 
- `error`: Error describing validation failure, nil if all validations pass

**Validation Rules**:
- Name: Must not be empty after trimming
- Email: Must match valid email format
- Status: Must be non-empty and valid
- CreatedAt: Must not be zero time
- LastUpdated: Must not be zero time
- DeletedAt: Can only be set if status is 'deleted'
- Password: Must be at least 8 characters

## Organization Constructor Functions

### NewOrganization
```go
func NewOrganization(req OrganizationCreate, idType IdType, serialID ...int64) (*Organization, error)
```
**Description**: Creates a new Organization instance with specified ID type.

**Arguments**:
- `req OrganizationCreate`: Organization data
- `idType IdType`: Type of ID to generate (MongoIdType or SerialIdType)
- `serialID ...int64`: Optional serial ID (required for SerialIdType)

**Returns**: 
- `*Organization`: New organization instance with hashed password
- `error`: Error if validation fails, password hashing fails, or invalid parameters

### NewOrganizationForMongo
```go
func NewOrganizationForMongo(req OrganizationCreate) (*Organization, error)
```
**Description**: Convenience function to create Organization with MongoDB ObjectID.

**Arguments**:
- `req OrganizationCreate`: Organization data

**Returns**: 
- `*Organization`: New organization with MongoDB ObjectID
- `error`: Error if creation fails

### NewOrganizationForSQL
```go
func NewOrganizationForSQL(req OrganizationCreate, id ...int64) (*Organization, error)
```
**Description**: Convenience function to create Organization with serial ID.

**Arguments**:
- `req OrganizationCreate`: Organization data
- `id ...int64`: Optional serial ID values

**Returns**: 
- `*Organization`: New organization with serial ID
- `error`: Error if creation fails

## Organization Methods

### SetMongoID
```go
func (o *Organization) SetMongoID(id primitive.ObjectID)
```
**Description**: Sets the organization's ID to a MongoDB ObjectID.

**Arguments**:
- `id primitive.ObjectID`: MongoDB ObjectID to set

**Returns**: None

### SetSerialID
```go
func (o *Organization) SetSerialID(id int64)
```
**Description**: Sets the organization's ID to a serial ID.

**Arguments**:
- `id int64`: Serial ID to set

**Returns**: None

### GetIDString
```go
func (o *Organization) GetIDString() string
```
**Description**: Returns string representation of the organization's ID.

**Arguments**: None (receiver method)

**Returns**: 
- `string`: ID as string, empty if organization or ID is nil

### IsMongoID
```go
func (o *Organization) IsMongoID() bool
```
**Description**: Checks if organization uses MongoDB ObjectID.

**Arguments**: None (receiver method)

**Returns**: 
- `bool`: true if organization has valid MongoDB ObjectID

### IsSerialID
```go
func (o *Organization) IsSerialID() bool
```
**Description**: Checks if organization uses serial ID.

**Arguments**: None (receiver method)

**Returns**: 
- `bool`: true if organization has valid serial ID

### SetPassword
```go
func (o *Organization) SetPassword(password string) error
```
**Description**: Updates organization password with bcrypt hashing and updates LastUpdated timestamp.

**Arguments**:
- `password string`: New plaintext password (minimum 8 characters)

**Returns**: 
- `error`: Error if organization is nil, password too short, or hashing fails

### CheckPassword
```go
func (o *Organization) CheckPassword(password string) bool
```
**Description**: Verifies plaintext password against stored hash.

**Arguments**:
- `password string`: Plaintext password to verify

**Returns**: 
- `bool`: true if password matches stored hash

### ActivateOrganization
```go
func (o *Organization) ActivateOrganization()
```
**Description**: Sets organization status to active and updates timestamp.

**Arguments**: None (receiver method)

**Returns**: None

### DeactivateOrganization
```go
func (o *Organization) DeactivateOrganization()
```
**Description**: Sets organization status to inactive and updates timestamp.

**Arguments**: None (receiver method)

**Returns**: None

### SuspendOrganization
```go
func (o *Organization) SuspendOrganization()
```
**Description**: Sets organization status to suspended and updates timestamp.

**Arguments**: None (receiver method)

**Returns**: None

### SoftDeleteOrganization
```go
func (o *Organization) SoftDeleteOrganization()
```
**Description**: Sets organization status to deleted, sets DeletedAt timestamp, and updates LastUpdated.

**Arguments**: None (receiver method)

**Returns**: None

### ToResponse
```go
func (o *Organization) ToResponse() *OrganizationResponse
```
**Description**: Converts Organization to OrganizationResponse, excluding sensitive data.

**Arguments**: None (receiver method)

**Returns**: 
- `*OrganizationResponse`: Response object with public fields, nil if organization is nil

### TableName
```go
func (Organization) TableName() string
```
**Description**: Returns SQL table name for GORM.

**Arguments**: None

**Returns**: 
- `string`: "organizations"

### CollectionName
```go
func (Organization) CollectionName() string
```
**Description**: Returns MongoDB collection name.

**Arguments**: None

**Returns**: 
- `string`: "organizations"

### BeforeUpdate
```go
func (o *Organization) BeforeUpdate()
```
**Description**: Hook called before database updates. Updates LastUpdated timestamp and manages DeletedAt field based on status.

**Arguments**: None (receiver method)

**Returns**: None

## Database Interfaces

The package implements several Go interfaces for database compatibility:

- **driver.Valuer**: FlexibleID, OrgStatus
- **driver.Scanner**: FlexibleID, OrgStatus  
- **json.Marshaler/Unmarshaler**: FlexibleID
- **bson.ValueMarshaler/ValueUnmarshaler**: FlexibleID

## Usage Notes

1. **FlexibleID**: Automatically handles MongoDB ObjectIDs and SQL serial IDs with transparent conversion
2. **Password Security**: All passwords are automatically hashed using bcrypt with default cost
3. **Timestamps**: CreatedAt and LastUpdated are managed automatically by database ORMs
4. **Soft Deletion**: DeletedAt field supports soft delete patterns
5. **Validation**: Built-in validation for email format, password strength, and required fields
6. **Database Agnostic**: Same models work with both MongoDB and SQL databases