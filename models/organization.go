package models

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

	"config"
	"utils"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9](\.?[a-zA-Z0-9_\-+%]){0,63}@[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z]{2,})+$`)

type IdType string

const (
	MongoIdType  IdType = "mongo"
	SerialIdType IdType = "serial"
)

type FlexibleID struct {
	Type     IdType              `json:"type" bson:"type"`
	MongoID  *primitive.ObjectID `json:"mongo_id,omitempty" bson:"mongo_id,omitempty"`
	SerialID *int64              `json:"serial_id,omitempty" bson:"serial_id,omitempty"`
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
			utils.Log(zapcore.ErrorLevel, "Failed to parse ID string", []string{"prod", "minimal"}, 
				zap.String("value", v), zap.Error(err))
			return fmt.Errorf("cannot parse ID: %v", v)
		}
	case []byte:
		return f.Scan(string(v))
	default:
		utils.Log(zapcore.ErrorLevel, "Cannot scan unsupported type into FlexibleID", []string{"prod", "minimal"}, 
			zap.String("type", fmt.Sprintf("%T", value)))
		return fmt.Errorf("cannot scan %T into FlexibleID", value)
	}

	return nil
}

func (f *FlexibleID) MarshalJSON() ([]byte, error) {
	if f == nil || f.IsEmpty() {
		return json.Marshal(nil)
	}

	jsonData := map[string]interface{}{
		"type":  f.Type,
		"value": f.String(),
	}
	
	return json.Marshal(jsonData)
}

func (f *FlexibleID) UnmarshalJSON(data []byte) error {
	if f == nil {
		return errors.New("cannot unmarshal into nil FlexibleID")
	}
	
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
			utils.Log(zapcore.ErrorLevel, "Failed to parse MongoDB ObjectID from hex", []string{"prod", "minimal"}, 
				zap.String("value", value), zap.Error(err))
			return err
		}
		f.MongoID = &id
	case SerialIdType:
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			utils.Log(zapcore.ErrorLevel, "Failed to parse serial ID", []string{"prod", "minimal"}, 
				zap.String("value", value), zap.Error(err))
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
		utils.Log(zapcore.ErrorLevel, "Cannot unmarshal BSON type into FlexibleID", []string{"prod", "minimal"}, 
			zap.String("bson_type", fmt.Sprintf("%v", t)))
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
	utils.Log(zapcore.WarnLevel, "Invalid OrgStatus", []string{"prod"}, 
		zap.String("status", string(s)))
	return false
}

func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	
	isValid := emailRegex.MatchString(email)
	if !isValid {
		utils.Log(zapcore.WarnLevel, "Email validation failed", []string{"prod"}, 
			zap.String("email", email))
	}
	
	return isValid
}

func (s OrgStatus) Value() (driver.Value, error) {
	return string(s), nil
}

func (s *OrgStatus) Scan(value interface{}) error {
	if s == nil {
		utils.Log(zapcore.ErrorLevel, "Cannot scan into nil OrgStatus", []string{"prod", "minimal"})
		return errors.New("cannot scan into nil OrgStatus")
	}
	
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
		utils.Log(zapcore.ErrorLevel, "Cannot scan type into OrgStatus", []string{"prod", "minimal"}, 
			zap.String("type", fmt.Sprintf("%T", value)))
		return fmt.Errorf("cannot scan %T into OrgStatus", value)
	}

	if !s.IsValid() {
		utils.Log(zapcore.ErrorLevel, "Invalid OrgStatus after scan", []string{"prod", "minimal"}, 
			zap.String("status", string(*s)))
		return fmt.Errorf("invalid OrgStatus: %s", *s)
	}

	return nil
}

func validateOrganizationCreateRequest(req OrganizationCreate) error {
	if strings.TrimSpace(req.Name) == "" {
		utils.Log(zapcore.WarnLevel, "Organization validation failed: name is required", []string{"dev", "prod"})
		return errors.New("name is required")
	}
	
	if !IsValidEmail(req.Email) {
		utils.Log(zapcore.WarnLevel, "Organization validation failed: invalid email format", []string{"dev", "prod"}, 
			zap.String("email", req.Email))
		return fmt.Errorf("invalid email format: %s", req.Email)
	}
	
	if req.Status == "" || !req.Status.IsValid() {
		utils.Log(zapcore.WarnLevel, "Organization validation failed: invalid status", []string{"dev", "prod"}, 
			zap.String("status", string(req.Status)))
		return errors.New("status is required and must be one of: active, inactive, suspended, deleted")
	}
	
	if req.CreatedAt.IsZero() {
		utils.Log(zapcore.WarnLevel, "Organization validation failed: created_at is required", []string{"dev", "prod"})
		return errors.New("created_at is required")
	}
	
	if req.LastUpdated.IsZero() {
		utils.Log(zapcore.WarnLevel, "Organization validation failed: last_updated is required", []string{"dev", "prod"})
		return errors.New("last_updated is required")
	}
	
	if req.DeletedAt != nil && req.Status != OrgStatusDeleted {
		utils.Log(zapcore.WarnLevel, "Organization validation failed: deleted_at can only be set if status is 'deleted'", []string{"dev", "prod"})
		return errors.New("deleted_at can only be set if status is 'deleted'")
	}
	
	if len(req.Password) < 8 {
		utils.Log(zapcore.WarnLevel, "Organization validation failed: password too short", []string{"dev", "prod"}, 
			zap.Int("password_length", len(req.Password)))
		return errors.New("password must be at least 8 characters long")
	}
	
	utils.Log(zapcore.InfoLevel, "Organization create request validation completed successfully", []string{"dev", "prod"})
	return nil
}

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

type OrganizationCreate struct {
	Name        string     `json:"name" validate:"required,min=3,max=100"`
	Email       string     `json:"email" validate:"required,email"`
	Status      OrgStatus  `json:"status" validate:"required,oneof=active inactive suspended deleted"`
	Password    string     `json:"password" validate:"required,min=8"`
	CreatedAt   time.Time  `json:"created_at"`
	LastUpdated time.Time  `json:"last_updated"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type OrganizationUpdate struct {
	Name        *string    `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Email       *string    `json:"email,omitempty" validate:"omitempty,email"`
	Status      *OrgStatus `json:"status,omitempty" validate:"omitempty,oneof=active inactive suspended deleted"`
	Password    *string    `json:"password,omitempty" validate:"omitempty,min=8"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	LastUpdated *time.Time `json:"last_updated,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type OrganizationResponse struct {
	ID          interface{} `json:"id"`
	Name        string      `json:"name"`
	Email       string      `json:"email"`
	Status      OrgStatus   `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	LastUpdated time.Time   `json:"last_updated"`
	DeletedAt   *time.Time  `json:"deleted_at,omitempty"`
}

func NewOrganization(ctx context.Context, req OrganizationCreate, idType IdType, serialID ...int64) (*Organization, error) {
	utils.Log(zapcore.InfoLevel, "Creating new organization", []string{"dev", "prod", "minimal"}, 
		zap.String("name", req.Name), zap.String("email", req.Email), zap.String("id_type", string(idType)))
	
	select {
	case <-ctx.Done():
		utils.Log(zapcore.WarnLevel, "Organization creation cancelled due to context timeout", []string{"dev", "prod"})
		return nil, ctx.Err()
	default:
	}
	
	if err := validateOrganizationCreateRequest(req); err != nil {
		utils.Log(zapcore.ErrorLevel, "Organization creation failed validation", []string{"prod", "minimal"}, 
			zap.Error(err))
		return nil, err
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.Log(zapcore.ErrorLevel, "Failed to hash password during organization creation", []string{"prod", "minimal"}, 
			zap.Error(err))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	org := &Organization{
		Name:        req.Name,
		Email:       req.Email,
		Password:    string(hashedPassword),
		Status:      req.Status,
		CreatedAt:   req.CreatedAt,
		LastUpdated: req.LastUpdated,
		DeletedAt:   req.DeletedAt,
	}

	switch idType {
	case MongoIdType:
		org.ID = NewMongoID()
	case SerialIdType:
		if len(serialID) > 0 {
			org.ID = NewSerialID(serialID[0])
			org.SerialID = &serialID[0] // sync SQL ID field
		} else {
			utils.Log(zapcore.ErrorLevel, "Serial ID required but not provided for organization creation", []string{"dev", "prod", "minimal"})
			return nil, errors.New("serial ID required for SerialIdType")
		}
	default:
		utils.Log(zapcore.ErrorLevel, "Invalid ID type for organization creation", []string{"dev", "prod", "minimal"}, 
			zap.String("id_type", string(idType)))
		return nil, errors.New("invalid ID type")
	}

	utils.Log(zapcore.InfoLevel, "Organization created successfully", []string{"dev", "prod", "minimal"}, 
		zap.String("id", org.GetIDString()), zap.String("name", org.Name))
	return org, nil
}

func NewOrganizationForMongo(ctx context.Context, req OrganizationCreate) (*Organization, error) {
	utils.Log(zapcore.InfoLevel, "Creating organization for MongoDB", []string{"dev", "prod"})
	return NewOrganization(ctx, req, MongoIdType)
}

func NewOrganizationForSQL(ctx context.Context, req OrganizationCreate, id ...int64) (*Organization, error) {
	utils.Log(zapcore.InfoLevel, "Creating organization for SQL database", []string{"dev", "prod"})
	return NewOrganization(ctx, req, SerialIdType, id...)
}

func (o *Organization) SetMongoID(id primitive.ObjectID) {
	if o == nil {
		utils.Log(zapcore.WarnLevel, "Attempted to set MongoDB ID on nil organization", []string{"prod"})
		return
	}
	
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.ID = FromMongoID(id)
}

func (o *Organization) SetSerialID(id int64) {
	if o == nil {
		utils.Log(zapcore.WarnLevel, "Attempted to set serial ID on nil organization", []string{"prod"})
		return
	}
	
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.ID = FromSerialID(id)
}

func (o *Organization) GetIDString() string {
	if o == nil {
		utils.Log(zapcore.WarnLevel, "Attempted to get ID string from nil organization", []string{"prod"})
		return ""
	}
	
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	if o.ID == nil {
		return ""
	}
	
	return o.ID.String()
}

func (o *Organization) IsMongoID() bool {
	if o == nil {
		return false
	}
	
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	if o.ID == nil {
		return false
	}
	
	return o.ID.Type == MongoIdType && o.ID.MongoID != nil
}

func (o *Organization) IsSerialID() bool {
	if o == nil {
		return false
	}
	
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	if o.ID == nil {
		return false
	}
	
	return o.ID.Type == SerialIdType && o.ID.SerialID != nil
}

func (o *Organization) SetPassword(password string) error {
	if o == nil {
		return errors.New("organization cannot be nil")
	}
	
	if len(password) < 8 {
		orgID := o.GetIDString()
		utils.Log(zapcore.WarnLevel, "Password too short for organization", []string{"prod"}, 
			zap.String("org_id", orgID), zap.Int("password_length", len(password)))
		return errors.New("password must be at least 8 characters long")
	}

	orgID := o.GetIDString()
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		utils.Log(zapcore.ErrorLevel, "Failed to hash password for organization", []string{"prod", "minimal"}, 
			zap.String("org_id", orgID), zap.Error(err))
		return fmt.Errorf("failed to hash password: %w", err)
	}
	
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.Password = string(hashedPassword)
	o.LastUpdated = time.Now()
	
	return nil
}

func (o *Organization) CheckPassword(password string) bool {
    if o == nil {
        utils.Log(zapcore.WarnLevel, "Password check attempted on nil organization", []string{"dev", "prod"})
        return false
    }
    
    o.mu.RLock()
    storedPassword := o.Password
    var orgIDStr string
    if o.ID != nil {
        orgIDStr = o.ID.String()
    }
    o.mu.RUnlock()
    
    if storedPassword == "" {
        utils.Log(zapcore.WarnLevel, "Password check failed: no stored password", []string{"prod"}, 
            zap.String("org_id", orgIDStr))
        return false
    }
    
    err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
    isValid := err == nil
    
    if !isValid {
        utils.Log(zapcore.WarnLevel, "Password check failed", []string{"prod"}, 
            zap.String("org_id", orgIDStr))
    }
    
    return isValid
}

func (o *Organization) ActivateOrganization() {
	if o == nil {
		utils.Log(zapcore.WarnLevel, "Attempted to activate nil organization", []string{"dev", "prod"})
		return
	}
	
	orgID := o.GetIDString()
	utils.Log(zapcore.InfoLevel, "Activating organization", []string{"dev", "prod", "minimal"}, 
		zap.String("org_id", orgID))
	
	o.mu.Lock()
	defer o.mu.Unlock()
	
	oldStatus := o.Status
	o.Status = OrgStatusActive
	o.LastUpdated = time.Now()
	
	utils.Log(zapcore.InfoLevel, "Organization status changed", []string{"dev", "prod", "minimal"}, 
		zap.String("org_id", orgID), zap.String("old_status", string(oldStatus)), zap.String("new_status", string(o.Status)))
}

func (o *Organization) DeactivateOrganization() {
	if o == nil {
		utils.Log(zapcore.WarnLevel, "Attempted to deactivate nil organization", []string{"dev", "prod"})
		return
	}
	
	orgID := o.GetIDString()
	utils.Log(zapcore.InfoLevel, "Deactivating organization", []string{"dev", "prod", "minimal"}, 
		zap.String("org_id", orgID))
	
	o.mu.Lock()
	defer o.mu.Unlock()
	
	oldStatus := o.Status
	o.Status = OrgStatusInactive
	o.LastUpdated = time.Now()
	
	utils.Log(zapcore.InfoLevel, "Organization status changed", []string{"dev", "prod", "minimal"}, 
		zap.String("org_id", orgID), zap.String("old_status", string(oldStatus)), zap.String("new_status", string(o.Status)))
}

func (o *Organization) SuspendOrganization() {
	if o == nil {
		utils.Log(zapcore.WarnLevel, "Attempted to suspend nil organization", []string{"dev", "prod"})
		return
	}
	
	orgID := o.GetIDString()
	utils.Log(zapcore.InfoLevel, "Suspending organization", []string{"dev", "prod", "minimal"}, 
		zap.String("org_id", orgID))
	
	o.mu.Lock()
	defer o.mu.Unlock()
	
	oldStatus := o.Status
	o.Status = OrgStatusSuspended
	o.LastUpdated = time.Now()
	
	utils.Log(zapcore.WarnLevel, "Organization suspended", []string{"dev", "prod", "minimal"}, 
		zap.String("org_id", orgID), zap.String("old_status", string(oldStatus)), zap.String("new_status", string(o.Status)))
}

func (o *Organization) SoftDeleteOrganization() {
	if o == nil {
		utils.Log(zapcore.WarnLevel, "Attempted to soft delete nil organization", []string{"dev", "prod"})
		return
	}
	
	orgID := o.GetIDString()
	utils.Log(zapcore.InfoLevel, "Soft deleting organization", []string{"dev", "prod", "minimal"}, 
		zap.String("org_id", orgID))
	
	o.mu.Lock()
	defer o.mu.Unlock()
	
	now := time.Now()
	oldStatus := o.Status
	o.Status = OrgStatusDeleted
	o.DeletedAt = &now
	o.LastUpdated = now
	
	utils.Log(zapcore.WarnLevel, "Organization soft deleted", []string{"dev", "prod", "minimal"}, 
		zap.String("org_id", orgID), zap.String("old_status", string(oldStatus)), zap.String("new_status", string(o.Status)))
}

func (o *Organization) ToResponse() *OrganizationResponse {
	if o == nil {
		utils.Log(zapcore.WarnLevel, "Attempted to convert nil organization to response", []string{"dev", "prod"})
		return nil
	}
	
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	orgID := o.GetIDString()
	utils.Log(zapcore.DebugLevel, "Converting organization to response format", []string{"dev"}, 
		zap.String("org_id", orgID))
	
	var id interface{}
	if o.ID != nil && !o.ID.IsEmpty() && o.ID.Type != "" {
		switch o.ID.Type {
		case MongoIdType:
			if o.ID.MongoID != nil {
				id = o.ID.MongoID.Hex()
			}
		case SerialIdType:
			if o.ID.SerialID != nil {
				id = *o.ID.SerialID
			}
		}
	}

	response := &OrganizationResponse{
		ID:          id,
		Name:        o.Name,
		Email:       o.Email,
		Status:      o.Status,
		CreatedAt:   o.CreatedAt,
		LastUpdated: o.LastUpdated,
		DeletedAt:   o.DeletedAt,
	}
	
	utils.Log(zapcore.DebugLevel, "Organization converted to response format successfully", []string{"dev"}, 
		zap.String("org_id", orgID))
	return response
}

func (Organization) TableName() string {
	utils.Log(zapcore.DebugLevel, "Getting SQL table name for Organization", []string{"dev"})
	return "organizations"
}

func (Organization) CollectionName() string {
	utils.Log(zapcore.DebugLevel, "Getting MongoDB collection name for Organization", []string{"dev"})
	return "organizations"
}

func (o *Organization) BeforeUpdate(tx *gorm.DB) error {
    ctx := tx.Statement.Context
    if ctx == nil {
        ctx = context.Background()
    }
    
    select {
    	case <-ctx.Done():
        	return ctx.Err()
    	default:
    }
	if o == nil {
		utils.Log(zapcore.WarnLevel, "BeforeUpdate called on nil organization", []string{"dev", "prod"})
		return errors.New("organization cannot be nil")
	}
	
	orgID := o.GetIDString()
	utils.Log(zapcore.DebugLevel, "Running BeforeUpdate hook for organization", []string{"dev"}, 
		zap.String("org_id", orgID))
	
	o.mu.Lock()
	defer o.mu.Unlock()
	
	now := time.Now()
	o.LastUpdated = now
	
	if o.Status == OrgStatusDeleted {
		o.DeletedAt = &now
		utils.Log(zapcore.InfoLevel, "Organization marked for deletion in BeforeUpdate", []string{"dev", "prod"}, 
			zap.String("org_id", orgID))
	} else {
		o.DeletedAt = nil
		utils.Log(zapcore.DebugLevel, "Organization deletion timestamp cleared in BeforeUpdate", []string{"dev"}, 
			zap.String("org_id", orgID))
	}
	return nil
}

func (o *Organization) AfterCreate(tx *gorm.DB) error {
	if o == nil {
		utils.Log(zapcore.ErrorLevel, "AfterCreate called on nil organization", []string{"dev", "prod", "minimal"})
		return errors.New("organization cannot be nil")
	}
	
    ctx := tx.Statement.Context
    if ctx == nil {
        ctx = context.Background()
    }
    
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
	orgID := o.GetIDString()
	utils.Log(zapcore.DebugLevel, "Running AfterCreate hook for organization", []string{"dev"}, 
		zap.String("org_id", orgID))
	
	if o.SerialID != nil && (o.ID == nil || o.ID.IsEmpty()) {
		utils.Log(zapcore.DebugLevel, "Syncing FlexibleID with SerialID in AfterCreate", []string{"dev"}, 
			zap.Int64("serial_id", *o.SerialID))
		o.ID = NewSerialID(*o.SerialID)
		utils.Log(zapcore.InfoLevel, "Organization FlexibleID synced after create", []string{"dev", "prod"}, 
			zap.String("org_id", o.GetIDString()))
	}
	
	return nil
}

func CreateOrganizationMongo(ctx context.Context, req OrganizationCreate) (*Organization, error) {
    utils.Log(zapcore.InfoLevel, "Creating MongoDB organization", []string{"dev", "prod"},
        zap.String("name", req.Name), zap.String("email", req.Email))

    org, err := NewOrganizationForMongo(ctx, req)
    if err != nil {
        utils.Log(zapcore.ErrorLevel, "Failed to create MongoDB organization", []string{"dev", "prod", "minimal"}, 
            zap.Error(err))
        return nil, err
    }
    
    collection := config.MongoDB.Collection("organizations")
    _, err = collection.InsertOne(ctx, org)
    if err != nil {
        utils.Log(zapcore.ErrorLevel, "Failed to insert organization into MongoDB", []string{"dev", "prod", "minimal"}, 
            zap.Error(err))
        return nil, err
    }
    
    utils.Log(zapcore.InfoLevel, "MongoDB organization created successfully", []string{"dev", "prod"}, 
        zap.String("id", org.GetIDString()), zap.String("name", org.Name))
    return org, nil
}
func CreateOrganizationSQL(ctx context.Context, req OrganizationCreate) (*Organization, error) {
    org, err := NewOrganizationForSQL(ctx, req)
    if err != nil {
        return nil, err
    }
    
    err = config.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(org).Error; err != nil {
            return err
        }
        return nil
    })
    
    if err != nil {
        return nil, err
    }
    return org, nil
}

func applyOrganizationUpdates(ctx context.Context, req OrganizationUpdate, org *Organization) error {
    orgID := org.GetIDString()

    if req.Password != nil {
        if err := org.SetPassword(*req.Password); err != nil {
            return err
        }
    }

    org.mu.Lock()
    defer org.mu.Unlock()

    if req.Name != nil {
        org.Name = *req.Name
    }
    
    if req.Email != nil {
        if !IsValidEmail(*req.Email) {
            return fmt.Errorf("invalid email format: %s", *req.Email)
        }
        org.Email = *req.Email
    }
    
    if req.Status != nil {
        if !req.Status.IsValid() {
            return errors.New("invalid status")
        }
        org.Status = *req.Status
    }
    
    if req.CreatedAt != nil {
        org.CreatedAt = *req.CreatedAt
    }
    
    if req.LastUpdated != nil {
        org.LastUpdated = *req.LastUpdated
    }
    
    if req.DeletedAt != nil {
        org.DeletedAt = req.DeletedAt
    }
    now := time.Now()
    org.LastUpdated = now
    
    if org.Status == OrgStatusDeleted {
        org.DeletedAt = &now
    } else if req.DeletedAt == nil {
        org.DeletedAt = nil
    }

    return nil
}


func UpdateOrganizationSQL(ctx context.Context, req OrganizationUpdate, org *Organization) error {
    if org == nil {
        return errors.New("organization cannot be nil")
    }

    return config.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        orgCopy := *org
        
        if err := applyOrganizationUpdates(ctx, req, &orgCopy); err != nil {
            return err
        }
        
        if err := tx.Save(&orgCopy).Error; err != nil {
            return err
        }
        
        *org = orgCopy
        return nil
    })
}
func UpdateOrganizationMongo(ctx context.Context, req OrganizationUpdate, org *Organization) error {
    if org == nil {
        return errors.New("organization cannot be nil")
    }

    if err := applyOrganizationUpdates(ctx, req, org); err != nil {
        return err
    }
    
    collection := config.MongoDB.Collection("organizations")
    filter := bson.M{"_id": org.ID.MongoID}
    update := bson.M{"$set": org}
    
    _, err := collection.UpdateOne(ctx, filter, update)
    if err != nil {
        utils.Log(zapcore.ErrorLevel, "Failed to update organization in MongoDB", []string{"dev", "prod"}, 
            zap.String("org_id", org.GetIDString()), zap.Error(err))
        return err
    }
    
    return nil
}


func PatchOrganizationSQL(ctx context.Context, req OrganizationUpdate, org *Organization) error {
    return UpdateOrganizationSQL(ctx, req, org)
}
func PatchOrganizationMongo(ctx context.Context, req OrganizationUpdate, org *Organization) error {
    return UpdateOrganizationMongo(ctx, req, org)
}


func GetOrganizationByIDSQL(ctx context.Context, id *FlexibleID) (*Organization, error) {
    if id == nil {
        return nil, errors.New("id cannot be nil")
    }
    if id.Type != SerialIdType || id.SerialID == nil {
        return nil, errors.New("invalid ID type for SQL query")
    }

    var org Organization
    if err := config.DB.WithContext(ctx).First(&org, "id = ?", *id.SerialID).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("organization not found with id: %d", *id.SerialID)
        }
        utils.Log(zapcore.ErrorLevel, "Failed to find organization by ID", []string{"prod"}, 
            zap.Int64("id", *id.SerialID), zap.Error(err))
        return nil, err
    }
    return &org, nil
}
func GetOrganizationByIDMongo(ctx context.Context, id *FlexibleID) (*Organization, error) {
    if id == nil {
        return nil, errors.New("id cannot be nil")
    }

    collection := config.MongoDB.Collection("organizations")
    var org Organization
    err := collection.FindOne(ctx, bson.M{"_id": id.MongoID}).Decode(&org)
    
    if err == mongo.ErrNoDocuments {
        return nil, fmt.Errorf("organization not found with id: %s", id.String())
    }
    
    if err != nil {
        utils.Log(zapcore.ErrorLevel, "Failed to find organization by ID", []string{"prod"}, 
            zap.String("id", id.String()), zap.Error(err))
        return nil, err
    }
    
    return &org, nil
}

func validatePaginationParams(limit, offset int) error {
    if limit < 0 {
        return errors.New("limit cannot be negative")
    }
    if offset < 0 {
        return errors.New("offset cannot be negative")
    }
    if limit > 1000 {
        return errors.New("limit cannot exceed 1000")
    }
    return nil
}

func GetOrganizationsSQL(ctx context.Context, limit, offset int) ([]*Organization, error) {
	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}
	var orgs []*Organization
	if err := config.DB.WithContext(ctx).Limit(limit).Offset(offset).Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}
func GetOrganizationsMongo(ctx context.Context, limit, offset int) ([]*Organization, error) {
	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}
	collection := config.MongoDB.Collection("organizations")
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset))
	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orgs []*Organization
	for cursor.Next(ctx) {
		var org Organization
		if err := cursor.Decode(&org); err != nil {
			return nil, err
		}
		orgs = append(orgs, &org)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return orgs, nil
}


func DeleteOrganizationSQL(ctx context.Context, org *Organization) error {
	if org == nil {
		return errors.New("organization cannot be nil")
	}

	orgID := org.GetIDString()
	utils.Log(zapcore.InfoLevel, "Deleting organization", []string{"dev", "prod"}, 
		zap.String("org_id", orgID))

	return config.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(org).Error; err != nil {
			utils.Log(zapcore.ErrorLevel, "Failed to delete organization", []string{"dev", "prod"}, 
				zap.String("org_id", orgID), zap.Error(err))
			return err
		}
		utils.Log(zapcore.InfoLevel, "Organization deleted successfully", []string{"dev", "prod"}, 
			zap.String("org_id", orgID))
		return nil
	})
}
func DeleteOrganizationMongo(ctx context.Context, org *Organization) error {
	if org == nil {
		return errors.New("organization cannot be nil")
	}

	orgID := org.GetIDString()
	utils.Log(zapcore.InfoLevel, "Deleting organization from MongoDB", []string{"dev", "prod"}, 
		zap.String("org_id", orgID))

	collection := config.MongoDB.Collection("organizations")
	_, err := collection.DeleteOne(ctx, bson.M{"_id": org.ID.MongoID})
	if err != nil {
		utils.Log(zapcore.ErrorLevel, "Failed to delete organization from MongoDB", []string{"dev", "prod"}, 
			zap.String("org_id", orgID), zap.Error(err))
		return err
	}
	
	utils.Log(zapcore.InfoLevel, "Organization deleted successfully from MongoDB", []string{"dev", "prod"}, 
		zap.String("org_id", orgID))
	return nil
}

func GetOrganizationByEmailMongo(ctx context.Context, email string) (*Organization, error) {
    if email == "" {
        return nil, errors.New("email cannot be empty")
    }

    collection := config.MongoDB.Collection("organizations")
    var org Organization
    err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&org)
    
    if err == mongo.ErrNoDocuments {
        utils.Log(zapcore.WarnLevel, "No organization found with given email", []string{"dev", "prod"}, 
            zap.String("email", email))
        return nil, fmt.Errorf("organization not found with email: %s", email)
    }
    
    if err != nil {
        utils.Log(zapcore.ErrorLevel, "Failed to find organization by email", []string{"dev", "prod"}, 
            zap.String("email", email), zap.Error(err))
        return nil, err
    }
    
    return &org, nil
}
func GetOrganizationByEmailSQL(ctx context.Context, email string) (*Organization, error) {
    if email == "" {
        return nil, errors.New("email cannot be empty")
    }

    var org Organization
    if err := config.DB.WithContext(ctx).First(&org, "email = ?", email).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("organization not found with email: %s", email)
        }
        utils.Log(zapcore.ErrorLevel, "Failed to find organization by email", []string{"prod"}, 
            zap.String("email", email), zap.Error(err))
        return nil, err
    }
    return &org, nil
}

func GetOrganizationByNameMongo(ctx context.Context, name string) (*Organization, error) {
    if name == "" {
        return nil, errors.New("name cannot be empty")
    }

    collection := config.MongoDB.Collection("organizations")
    var org Organization
    err := collection.FindOne(ctx, bson.M{"name": name}).Decode(&org)
    
    if err == mongo.ErrNoDocuments {
        utils.Log(zapcore.WarnLevel, "No organization found with given name", []string{"dev", "prod"}, 
            zap.String("name", name))
        return nil, fmt.Errorf("organization not found with name: %s", name)
    }
    
    if err != nil {
        utils.Log(zapcore.ErrorLevel, "Failed to find organization by name", []string{"dev", "prod"}, 
            zap.String("name", name), zap.Error(err))
        return nil, err
    }
    
    return &org, nil
}
func GetOrganizationByNameSQL(ctx context.Context, name string) (*Organization, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	var org Organization
	if err := config.DB.WithContext(ctx).First(&org, "name = ?", name).Error; err != nil {
		utils.Log(zapcore.ErrorLevel, "Failed to find organization by name", []string{"dev", "prod"}, 
			zap.String("name", name), zap.Error(err))
		return nil, err
	}
	return &org, nil
}
