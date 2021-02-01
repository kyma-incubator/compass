package repo

import (
	"database/sql"
	"time"
)

// TimestampableEntity denotes an DB-layer entity which can be timestamp with created_at and updated_at values
type TimestampableEntity interface {
	GetCreatedAt() time.Time
	SetCreatedAt(t time.Time)

	GetUpdatedAt() time.Time
	SetUpdatedAt(t time.Time)
}

func NewNullableString(text *string) sql.NullString {
	nullString := sql.NullString{}
	if text != nil {
		nullString.String = *text
		nullString.Valid = true
	}

	return nullString
}

func NewValidNullableString(text string) sql.NullString {
	return sql.NullString{
		String: text,
		Valid:  true,
	}
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

func BoolPtrFromNullableBool(sqlBool sql.NullBool) *bool {
	if sqlBool.Valid {
		return &sqlBool.Bool
	}
	return nil
}
