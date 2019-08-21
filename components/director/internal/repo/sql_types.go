package repo

import "database/sql"

func NewSqlNullString(text *string) sql.NullString {
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

func NewSqlNullBool(boolean *bool) sql.NullBool {
	var sqlBool sql.NullBool
	if boolean != nil {
		sqlBool = sql.NullBool{Valid: true, Bool: *boolean}
	}
	return sqlBool
}

func StringFromSqlNullString(sqlString *sql.NullString) *string {
	if sqlString == nil {
		return nil
	}

	if sqlString.Valid {
		return &sqlString.String
	}

	return nil
}

func BoolFromSqlNullBool(sqlBool *sql.NullBool) *bool {
	if sqlBool == nil {
		return nil
	}

	if sqlBool.Valid {
		return &sqlBool.Bool
	}
	return nil
}
