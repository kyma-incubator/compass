package repo

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Entity denotes an DB-layer entity which can be timestamped with created_at, updated_at, deleted_at and ready values
type Entity interface {
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

type BaseEntity struct {
	ID        string         `db:"id"`
	Ready     bool           `db:"ready"`
	CreatedAt *time.Time     `db:"created_at"`
	UpdatedAt *time.Time     `db:"updated_at"`
	DeletedAt *time.Time     `db:"deleted_at"`
	Error     sql.NullString `db:"error"`
}

func (e *BaseEntity) GetReady() bool {
	return e.Ready
}

func (e *BaseEntity) SetReady(ready bool) {
	e.Ready = ready
}

func (e *BaseEntity) GetCreatedAt() time.Time {
	if e.CreatedAt == nil {
		return time.Time{}
	}
	return *e.CreatedAt
}

func (e *BaseEntity) SetCreatedAt(t time.Time) {
	e.CreatedAt = &t
}

func (e *BaseEntity) GetUpdatedAt() time.Time {
	if e.UpdatedAt == nil {
		return time.Time{}
	}
	return *e.UpdatedAt
}

func (e *BaseEntity) SetUpdatedAt(t time.Time) {
	e.UpdatedAt = &t
}

func (e *BaseEntity) GetDeletedAt() time.Time {
	if e.DeletedAt == nil {
		return time.Time{}
	}
	return *e.DeletedAt
}

func (e *BaseEntity) SetDeletedAt(t time.Time) {
	e.DeletedAt = &t
}

func (e *BaseEntity) GetError() sql.NullString {
	return e.Error
}

func (e *BaseEntity) SetError(err sql.NullString) {
	e.Error = err
}

func NewNullableString(text *string) sql.NullString {
	nullString := sql.NullString{}
	if text != nil {
		nullString.String = *text
		nullString.Valid = true
	}

	return nullString
}

func NewNullableInt(i *int) sql.NullInt32 {
	nullInt := sql.NullInt32{}
	if i != nil {
		nullInt.Int32 = int32(*i)
		nullInt.Valid = true
	}

	return nullInt
}

func NewValidNullableString(text string) sql.NullString {
	return sql.NullString{
		String: text,
		Valid:  true,
	}
}

func NewNullableStringFromJSONRawMessage(json json.RawMessage) sql.NullString {
	nullString := sql.NullString{}
	if json != nil {
		nullString.String = string(json)
		nullString.Valid = true
	}
	return nullString
}

func NewNullableBool(boolean *bool) sql.NullBool {
	var sqlBool sql.NullBool
	if boolean != nil {
		sqlBool = sql.NullBool{Valid: true, Bool: *boolean}
	}

	return sqlBool
}

func NewValidNullableBool(boolean bool) sql.NullBool {
	return sql.NullBool{
		Valid: true,
		Bool:  boolean,
	}
}

func StringPtrFromNullableString(sqlString sql.NullString) *string {
	if sqlString.Valid {
		return &sqlString.String
	}

	return nil
}

func JSONRawMessageFromNullableString(sqlString sql.NullString) json.RawMessage {
	if sqlString.Valid {
		return json.RawMessage(sqlString.String)
	}
	return nil
}

func IntPtrFromNullableInt(i sql.NullInt32) *int {
	if i.Valid {
		val := int(i.Int32)
		return &val
	}

	return nil
}

func BoolPtrFromNullableBool(sqlBool sql.NullBool) *bool {
	if sqlBool.Valid {
		return &sqlBool.Bool
	}
	return nil
}
