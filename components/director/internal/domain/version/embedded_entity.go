package version

import (
	"database/sql"
)

type Version struct {
	Value           sql.NullString `db:"version_value"`
	Deprecated      sql.NullBool   `db:"version_deprecated"`
	DeprecatedSince sql.NullString `db:"version_deprecated_since"`
	ForRemoval      sql.NullBool   `db:"version_for_removal"`
}
