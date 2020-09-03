package repo

import (
	"database/sql"
	"encoding/json"
)

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

func NewNullableRawJSON(json json.RawMessage) sql.NullString {
	nullString := sql.NullString{}
	if json != nil {
		nullString.String = string([]byte(json))
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

func RawJSONFromNullableString(sqlString sql.NullString) json.RawMessage {
	if sqlString.Valid {
		return []byte(sqlString.String)
	}

	return nil
}

func BoolPtrFromNullableBool(sqlBool sql.NullBool) *bool {
	if sqlBool.Valid {
		return &sqlBool.Bool
	}
	return nil
}
