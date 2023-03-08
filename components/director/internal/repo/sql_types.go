package repo

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// ChildEntity is an interface for a child entity that can be used to obtain its parent ID.
type ChildEntity interface {
	GetParent(resource.Type) (resource.Type, string)
}

// Identifiable is an interface that can be used to identify an object.
type Identifiable interface {
	GetID() string
}

// EntityWithExternalTenant is an interface that can be used for object with an external tenant to add tenant_id to the struct.
// This is needed for update operations when we want to use named arguments in the SQL queries.
type EntityWithExternalTenant interface {
	DecorateWithTenantID(tenant string) interface{}
}

// Entity denotes an DB-layer entity which can be timestamped with created_at, updated_at, deleted_at and ready values
type Entity interface {
	Identifiable

	GetReady() bool
	SetReady(ready bool)

	GetCreatedAt() time.Time
	SetCreatedAt(t time.Time)

	GetUpdatedAt() time.Time
	SetUpdatedAt(t time.Time)

	GetDeletedAt() time.Time
	SetDeletedAt(t time.Time)

	GetError() sql.NullString
	SetError(err sql.NullString)
}

// BaseEntity represents a base implementation of Entity
type BaseEntity struct {
	ID        string         `db:"id"`
	Ready     bool           `db:"ready"`
	CreatedAt *time.Time     `db:"created_at"`
	UpdatedAt *time.Time     `db:"updated_at"`
	DeletedAt *time.Time     `db:"deleted_at"`
	Error     sql.NullString `db:"error"`
}

// GetID returns the ID of the entity
func (e *BaseEntity) GetID() string {
	return e.ID
}

// GetReady returns the ready value of the entity
func (e *BaseEntity) GetReady() bool {
	return e.Ready
}

// SetReady sets the ready value of the entity
func (e *BaseEntity) SetReady(ready bool) {
	e.Ready = ready
}

// GetCreatedAt returns the created_at value of the entity
func (e *BaseEntity) GetCreatedAt() time.Time {
	if e.CreatedAt == nil {
		return time.Time{}
	}
	return *e.CreatedAt
}

// SetCreatedAt sets the created_at value of the entity
func (e *BaseEntity) SetCreatedAt(t time.Time) {
	e.CreatedAt = &t
}

// GetUpdatedAt returns the updated_at value of the entity
func (e *BaseEntity) GetUpdatedAt() time.Time {
	if e.UpdatedAt == nil {
		return time.Time{}
	}
	return *e.UpdatedAt
}

// SetUpdatedAt sets the updated_at value of the entity
func (e *BaseEntity) SetUpdatedAt(t time.Time) {
	e.UpdatedAt = &t
}

// GetDeletedAt returns the deleted_at value of the entity
func (e *BaseEntity) GetDeletedAt() time.Time {
	if e.DeletedAt == nil {
		return time.Time{}
	}
	return *e.DeletedAt
}

// SetDeletedAt sets the deleted_at value of the entity
func (e *BaseEntity) SetDeletedAt(t time.Time) {
	e.DeletedAt = &t
}

// GetError returns the error value of the entity
func (e *BaseEntity) GetError() sql.NullString {
	return e.Error
}

// SetError sets the error value of the entity
func (e *BaseEntity) SetError(err sql.NullString) {
	e.Error = err
}

// NewNullableString returns a new sql.NullString based on the given string pointer
func NewNullableString(text *string) sql.NullString {
	nullString := sql.NullString{}
	if text != nil {
		nullString = NewValidNullableString(*text)
	}

	return nullString
}

// NewNullableInt returns a new sql.NullInt32 based on the given int pointer
func NewNullableInt(i *int) sql.NullInt32 {
	nullInt := sql.NullInt32{}
	if i != nil {
		nullInt.Int32 = int32(*i)
		nullInt.Valid = true
	}

	return nullInt
}

// NewValidNullableString returns a new sql.NullString based on the given string
func NewValidNullableString(text string) sql.NullString {
	if text == "" || text == "null" {
		return sql.NullString{}
	}

	return sql.NullString{
		String: text,
		Valid:  true,
	}
}

// NewNullableStringFromJSONRawMessage returns a new sql.NullString based on the given json.RawMessage
func NewNullableStringFromJSONRawMessage(json json.RawMessage) sql.NullString {
	nullString := sql.NullString{}
	if json != nil && string(json) != "null" {
		nullString.String = string(json)
		nullString.Valid = true
	}
	return nullString
}

// NewNullableBool returns a new sql.NullBool based on the given bool pointer
func NewNullableBool(boolean *bool) sql.NullBool {
	var sqlBool sql.NullBool
	if boolean != nil {
		sqlBool = sql.NullBool{Valid: true, Bool: *boolean}
	}

	return sqlBool
}

// NewValidNullableBool returns a new sql.NullBool based on the given bool
func NewValidNullableBool(boolean bool) sql.NullBool {
	return sql.NullBool{
		Valid: true,
		Bool:  boolean,
	}
}

// StringPtrFromNullableString returns a string pointer based on the given sql.NullString
func StringPtrFromNullableString(sqlString sql.NullString) *string {
	if sqlString.Valid {
		return &sqlString.String
	}

	return nil
}

// JSONRawMessageFromNullableString returns a json.RawMessage based on the given sql.NullString
func JSONRawMessageFromNullableString(sqlString sql.NullString) json.RawMessage {
	if sqlString.Valid {
		return json.RawMessage(sqlString.String)
	}
	return nil
}

// IntPtrFromNullableInt returns an int pointer based on the given sql.NullInt32
func IntPtrFromNullableInt(i sql.NullInt32) *int {
	if i.Valid {
		val := int(i.Int32)
		return &val
	}

	return nil
}

// BoolPtrFromNullableBool returns a bool pointer based on the given sql.NullBool
func BoolPtrFromNullableBool(sqlBool sql.NullBool) *bool {
	if sqlBool.Valid {
		return &sqlBool.Bool
	}
	return nil
}
