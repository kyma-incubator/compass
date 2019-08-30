package version

import (
	"database/sql"
)

//TODO: change name to NullableVersion
type Version struct {
	VersionValue           sql.NullString `db:"version_value"`
	VersionDepracated      sql.NullBool   `db:"version_deprecated"`
	VersionDepracatedSince sql.NullString `db:"version_deprecated_since"`
	VersionForRemoval      sql.NullBool   `db:"version_for_removal"`
}

// sqlx cannot fetch value from embedded struct, thats why we need to help sqlx.
//func (v *Version) Value() (driver.Value, error) {
//	if v == nil {
//		return nil, nil
//	}
//	if v.VersionValue.Valid {
//		return v, nil
//	}
//	return nil, nil
//}
