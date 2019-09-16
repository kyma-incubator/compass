package version

import (
	"database/sql"
)

type Version struct {
	VersionValue           sql.NullString `db:"version_value"`
	VersionDepracated      sql.NullBool   `db:"version_deprecated"`
	VersionDepracatedSince sql.NullString `db:"version_deprecated_since"`
	VersionForRemoval      sql.NullBool   `db:"version_for_removal"`
}
