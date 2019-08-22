package repo

import "database/sql"

func NewNullableString(text *string) sql.NullString {
	nullString := sql.NullString{
		String: "",
		Valid:  false,
	}
	if text != nil && len(*text) > 0 {
		nullString.String = *text
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

func StringFromSqlNullString(sqlString sql.NullString) *string {
	if sqlString.Valid {
		return &sqlString.String
	}

	return nil
}

func BoolFromSqlNullBool(sqlBool sql.NullBool) *bool {
	if sqlBool.Valid {
		return &sqlBool.Bool
	}
	return nil
}
