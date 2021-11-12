package repo

import (
	"database/sql"
	"encoding/json"
	"time"
)

type ChildEntity interface {
	GetParentID() string
}

type Identifiable interface {
	GetID() string
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

// BaseEntity missing godoc
type BaseEntity struct {
	ID        string         `db:"id"`
	Ready     bool           `db:"ready"`
	CreatedAt *time.Time     `db:"created_at"`
	UpdatedAt *time.Time     `db:"updated_at"`
	DeletedAt *time.Time     `db:"deleted_at"`
	Error     sql.NullString `db:"error"`
}

func (e *BaseEntity) GetID() string {
	return e.ID
}

// GetReady missing godoc
func (e *BaseEntity) GetReady() bool {
	return e.Ready
}

// SetReady missing godoc
func (e *BaseEntity) SetReady(ready bool) {
	e.Ready = ready
}

// GetCreatedAt missing godoc
func (e *BaseEntity) GetCreatedAt() time.Time {
	if e.CreatedAt == nil {
		return time.Time{}
	}
	return *e.CreatedAt
}

// SetCreatedAt missing godoc
func (e *BaseEntity) SetCreatedAt(t time.Time) {
	e.CreatedAt = &t
}

// GetUpdatedAt missing godoc
func (e *BaseEntity) GetUpdatedAt() time.Time {
	if e.UpdatedAt == nil {
		return time.Time{}
	}
	return *e.UpdatedAt
}

// SetUpdatedAt missing godoc
func (e *BaseEntity) SetUpdatedAt(t time.Time) {
	e.UpdatedAt = &t
}

// GetDeletedAt missing godoc
func (e *BaseEntity) GetDeletedAt() time.Time {
	if e.DeletedAt == nil {
		return time.Time{}
	}
	return *e.DeletedAt
}

// SetDeletedAt missing godoc
func (e *BaseEntity) SetDeletedAt(t time.Time) {
	e.DeletedAt = &t
}

// GetError missing godoc
func (e *BaseEntity) GetError() sql.NullString {
	return e.Error
}

// SetError missing godoc
func (e *BaseEntity) SetError(err sql.NullString) {
	e.Error = err
}

// NewNullableString missing godoc
func NewNullableString(text *string) sql.NullString {
	nullString := sql.NullString{}
	if text != nil {
		nullString.String = *text
		nullString.Valid = true
	}

	return nullString
}

// NewNullableInt missing godoc
func NewNullableInt(i *int) sql.NullInt32 {
	nullInt := sql.NullInt32{}
	if i != nil {
		nullInt.Int32 = int32(*i)
		nullInt.Valid = true
	}

	return nullInt
}

// NewValidNullableString missing godoc
func NewValidNullableString(text string) sql.NullString {
	if text == "" {
		return sql.NullString{}
	}

	return sql.NullString{
		String: text,
		Valid:  true,
	}
}

// NewNullableStringFromJSONRawMessage missing godoc
func NewNullableStringFromJSONRawMessage(json json.RawMessage) sql.NullString {
	nullString := sql.NullString{}
	if json != nil {
		nullString.String = string(json)
		nullString.Valid = true
	}
	return nullString
}

// NewNullableBool missing godoc
func NewNullableBool(boolean *bool) sql.NullBool {
	var sqlBool sql.NullBool
	if boolean != nil {
		sqlBool = sql.NullBool{Valid: true, Bool: *boolean}
	}

	return sqlBool
}

// NewValidNullableBool missing godoc
func NewValidNullableBool(boolean bool) sql.NullBool {
	return sql.NullBool{
		Valid: true,
		Bool:  boolean,
	}
}

// StringPtrFromNullableString missing godoc
func StringPtrFromNullableString(sqlString sql.NullString) *string {
	if sqlString.Valid {
		return &sqlString.String
	}

	return nil
}

// JSONRawMessageFromNullableString missing godoc
func JSONRawMessageFromNullableString(sqlString sql.NullString) json.RawMessage {
	if sqlString.Valid {
		return json.RawMessage(sqlString.String)
	}
	return nil
}

// IntPtrFromNullableInt missing godoc
func IntPtrFromNullableInt(i sql.NullInt32) *int {
	if i.Valid {
		val := int(i.Int32)
		return &val
	}

	return nil
}

// BoolPtrFromNullableBool missing godoc
func BoolPtrFromNullableBool(sqlBool sql.NullBool) *bool {
	if sqlBool.Valid {
		return &sqlBool.Bool
	}
	return nil
}
