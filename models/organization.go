package models

import (
	"time"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"fmt"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"database/sql/driver"
	"golang.org/x/crypto/bcrypt"
)

type IdType string
const (
	MongoIdType  IdType = "mongo"
	SerialIdType IdType = "serial"
)

type FlexibleID struct {
	Type     IdType               `json:"type" bson:"type"`
	MongoID  *primitive.ObjectID  `json:"mongo_id,omitempty" bson:"mongo_id,omitempty"`
	SerialID *int64               `json:"serial_id,omitempty" bson:"serial_id,omitempty"`
}

func NewMongoID() *FlexibleID {
	id := primitive.NewObjectID()
	return &FlexibleID{
		Type:    MongoIdType,
		MongoID: &id,
	}
}

func FromMongoID(id primitive.ObjectID) *FlexibleID {
	return &FlexibleID{
		Type:    MongoIdType,
		MongoID: &id,
	}
}


func NewSerialID(id int64) *FlexibleID {
	return &FlexibleID{
		Type:     SerialIdType,
		SerialID: &id,
	}
}

func FromSerialID(id int64) *FlexibleID {
	return &FlexibleID{
		Type:     SerialIdType,
		SerialID: &id,
	}
}


func (f *FlexibleID) String() string {
	if f == nil {
		return ""
	}
	
	switch f.Type {
	case MongoIdType:
		if f.MongoID != nil {
			return f.MongoID.Hex()
		}
	case SerialIdType:
		if f.SerialID != nil {
			return strconv.FormatInt(*f.SerialID, 10)
		}
	}
	return ""
}

func (f *FlexibleID) IsEmpty() bool {
	if f == nil {
		return true
	}
	
	switch f.Type {
	case MongoIdType:
		return f.MongoID == nil || f.MongoID.IsZero()
	case SerialIdType:
		return f.SerialID == nil || *f.SerialID == 0
	}
	return true
}

func (f *FlexibleID) GetValue() interface{} {
	if f == nil {
		return nil
	}
	
	switch f.Type {
	case MongoIdType:
		if f.MongoID != nil {
			return *f.MongoID
		}
	case SerialIdType:
		if f.SerialID != nil {
			return *f.SerialID
		}
	}
	return nil
}

func (f *FlexibleID) Value() (driver.Value, error) {
	if f == nil || f.IsEmpty() {
		return nil, nil
	}
	switch f.Type {
	case SerialIdType:
		if f.SerialID != nil {
			return *f.SerialID, nil
		}
	case MongoIdType:
		if f.MongoID != nil {
			return f.MongoID.Hex(), nil
		}
	}
	return nil, nil
}

func (f *FlexibleID) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	switch v := value.(type) {
	case int64:
		f.Type = SerialIdType
		f.SerialID = &v
	case string:
		// Try to parse as MongoDB ObjectID first
		if objectID, err := primitive.ObjectIDFromHex(v); err == nil {
			f.Type = MongoIdType
			f.MongoID = &objectID
		} else if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.Type = SerialIdType
			f.SerialID = &id
		} else {
			return fmt.Errorf("cannot parse ID: %v", v)
		}
	case []byte:
		return f.Scan(string(v))
	default:
		return fmt.Errorf("cannot scan %T into FlexibleID", value)
	}
	
	return nil
}

func (f *FlexibleID) MarshalJSON() ([]byte, error) {
	if f.IsEmpty() {
		return json.Marshal(nil)
	}
	
	return json.Marshal(map[string]interface{}{
		"type":  f.Type,
		"value": f.String(),
	})
}

func (f *FlexibleID) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	
	if raw == nil {
		return nil
	}
	
	typeStr, ok := raw["type"].(string)
	if !ok {
		return fmt.Errorf("invalid type in FlexibleID")
	}
	
	value, ok := raw["value"].(string)
	if !ok {
		return fmt.Errorf("invalid value in FlexibleID")
	}
	
	f.Type = IdType(typeStr)
	
	switch f.Type {
	case MongoIdType:
		id, err := primitive.ObjectIDFromHex(value)
		if err != nil {
			return err
		}
		f.MongoID = &id
	case SerialIdType:
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		f.SerialID = &id
	}
	
	return nil
}

func (f *FlexibleID) MarshalBSONValue() (bson.ValueType, []byte, error) {
	if f == nil || f.IsEmpty() {
		return bson.TypeNull, nil, nil
	}
	
	switch f.Type {
	case MongoIdType:
		if f.MongoID != nil {
			return bson.MarshalValue(*f.MongoID)
		}
	case SerialIdType:
		if f.SerialID != nil {
			return bson.MarshalValue(*f.SerialID)
		}
	}
	
	return bson.TypeNull, nil, nil
}

func (f *FlexibleID) UnmarshalBSONValue(t bson.ValueType, data []byte) error {
	switch t {
	case bson.TypeObjectID:
		var id primitive.ObjectID
		if err := bson.UnmarshalValue(t, data, &id); err != nil {
			return err
		}
		f.Type = MongoIdType
		f.MongoID = &id
	case bson.TypeInt64:
		var id int64
		if err := bson.UnmarshalValue(t, data, &id); err != nil {
			return err
		}
		f.Type = SerialIdType
		f.SerialID = &id
	case bson.TypeInt32:
		var id int32
		if err := bson.UnmarshalValue(t, data, &id); err != nil {
			return err
		}
		id64 := int64(id)
		f.Type = SerialIdType
		f.SerialID = &id64
	default:
		return fmt.Errorf("cannot unmarshal %v into FlexibleID", t)
	}
	
	return nil
}

type OrgStatus string
const (
	OrgStatusActive    OrgStatus = "active"
	OrgStatusInactive  OrgStatus = "inactive"
	OrgStatusSuspended OrgStatus = "suspended"
	OrgStatusDeleted   OrgStatus = "deleted"
)

func (s OrgStatus) IsValid() bool {
	switch s {
	case OrgStatusActive, OrgStatusInactive, OrgStatusSuspended, OrgStatusDeleted:
		return true
	}
	return false
}

func (email *string) IsValidEmail() bool {
	if email == nil || *email == "" {
		return false
	}
	re := regexp.MustCompile(`^[a-zA-Z0-9](\.?[a-zA-Z0-9_\-+%]){0,63}@[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z]{2,})+$`)
	return re.MatchString(*email)
}

func (s OrgStatus) Value() (driver.Value, error) {
	return string(s), nil
}

func (s *OrgStatus) Scan(value interface{}) error {
	if value == nil {
		*s = OrgStatusInactive
		return nil
	}
	
	switch v := value.(type) {
	case string:
		*s = OrgStatus(v)
	case []byte:
		*s = OrgStatus(v)
	default:
		return fmt.Errorf("cannot scan %T into OrgStatus", value)
	}
	
	if !s.IsValid() {
		return fmt.Errorf("invalid OrgStatus: %s", *s)
	}
	
	return nil
}

func validateOrganizationCreateRequest(req OrganizationCreate) error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("name is required")
	}
	if !(&req.Email).IsValidEmail() {
		return errors.New("invalid email format")
	}
	if req.Status == "" || !req.Status.IsValid() {
		return errors.New("status is required and must be one of: active, inactive, suspended, deleted")
	}
	if req.CreatedAt.IsZero() {
		return errors.New("created_at is required")
	}
	if req.LastUpdated.IsZero() {
		return errors.New("last_updated is required")
	}
	if req.DeletedAt != nil && req.Status != OrgStatusDeleted {
		return errors.New("deleted_at can only be set if status is 'deleted'")
	}
	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}

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

type OrganizationCreate struct {
	Name        string    `json:"name" validate:"required,min=3,max=100"`
	Email       string    `json:"email" validate:"required,email"`
	Status      OrgStatus `json:"status" validate:"required,oneof=active inactive suspended deleted"`
	Password    string    `json:"password" validate:"required,min=8"`
	CreatedAt   time.Time `json:"created_at"`
	LastUpdated time.Time `json:"last_updated"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type OrganizationUpdate struct {
	Name        *string   `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Email       *string   `json:"email,omitempty" validate:"omitempty,email"`
	Status      *OrgStatus `json:"status,omitempty" validate:"omitempty,oneof=active inactive suspended deleted"`
	Password    *string   `json:"password,omitempty" validate:"omitempty,min=8"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	LastUpdated *time.Time `json:"last_updated,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type OrganizationResponse struct {
	ID          interface{} `json:"id"`
	Name        string      `json:"name"`
	Email       string      `json:"email"`
	// Password    string      `json:"password"`
	Status      OrgStatus   `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	LastUpdated time.Time   `json:"last_updated"`
	DeletedAt   *time.Time  `json:"deleted_at,omitempty"`
}


func NewOrganization(req OrganizationCreate, idType IdType, serialID ...int64) (*Organization, error) {
    if err := validateOrganizationCreateRequest(req); err != nil {
        return nil, err
    }
    
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %w", err)
    }

    now := time.Now()
    org := &Organization{
        Name:        req.Name,
        Email:       req.Email,
        Password:    string(hashedPassword),
        Status:      req.Status,
        CreatedAt:   now,
        LastUpdated: now,
    }
    
    switch idType {
    case MongoIdType:
        org.ID = NewMongoID()
    case SerialIdType:
        if len(serialID) > 0 {
            org.ID = NewSerialID(serialID[0])
        } else {
            return nil, errors.New("serial ID required for SerialIdType")
        }
    default:
        return nil, errors.New("invalid ID type")
    }
    
    return org, nil
}

func NewOrganizationForMongo(req OrganizationCreate) (*Organization, error) {
	return NewOrganization(req, MongoIdType)
}

func NewOrganizationForSQL(req OrganizationCreate, id ...int64) (*Organization, error) {
    return NewOrganization(req, SerialIdType, id...)
}

func (o *Organization) SetMongoID(id primitive.ObjectID) {
	o.ID = FromMongoID(id)
}

func (o *Organization) SetSerialID(id int64) {
	o.ID = FromSerialID(id)
}

func (o *Organization) GetIDString() string {
	if o.ID == nil || o.ID.IsEmpty() {
		return ""
	}
	return o.ID.String()
}

func (o *Organization) IsMongoID() bool {
	return o.ID.Type == MongoIdType && o.ID.MongoID != nil
}

func (o *Organization) IsSerialID() bool {
	return o.ID.Type == SerialIdType && o.ID.SerialID != nil
}

func (o *Organization) SetPassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	o.Password = string(hashedPassword)
	o.LastUpdated = time.Now()
	return nil
}

func (o *Organization) CheckPassword(password string) bool {
	if o.Password == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(o.Password), []byte(password))
	return err == nil
}

func (o *Organization) ActivateOrganization() {
	o.Status = OrgStatusActive
	o.LastUpdated = time.Now()
}

func (o *Organization) DeactivateOrganization() {
	o.Status = OrgStatusInactive
	o.LastUpdated = time.Now()
}

func (o *Organization) SuspendOrganization() {
	o.Status = OrgStatusSuspended
	o.LastUpdated = time.Now()
}

func (o *Organization) SoftDeleteOrganization() {
	now := time.Now()
	o.Status = OrgStatusDeleted
	o.DeletedAt = &now
	o.LastUpdated = now
}


func (o *Organization) ToResponse() *OrganizationResponse {
	var id FlexibleID
	if o.ID != nil && !o.ID.IsEmpty() {
		id = o.ID.GetValue()
		if id.Type == MongoIdType && id.MongoID != nil {
			id = id.MongoID.Hex()
		} else if id.Type == SerialIdType && id.SerialID != nil {
			id = *id.SerialID
		}
	}

	return &OrganizationResponse{
		ID:          id,
		Name:        o.Name,
		Email:       o.Email,
		// Password:    o.Password, 
		Status:      o.Status,
		CreatedAt:   o.CreatedAt,
		LastUpdated: o.LastUpdated,
		DeletedAt:   o.DeletedAt,
	}
}

func (Organization) TableName() string {
	return "organizations"
}
func (Organization) CollectionName() string {
	return "organizations"
}
func (o *Organization) BeforeUpdate() {
	o.LastUpdated = time.Now()
	if o.Status == OrgStatusDeleted {
		now := time.Now()
		o.DeletedAt = &now
	} else {
		o.DeletedAt = nil
	}
}