package schema

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const (
	schemaVersionColumn = "version"
	schemaDirtyColumn   = "dirty"
	tableName           = `"public"."schema_migrations"`
)

// PgRepository represents a repository for schema migration operations
type PgRepository struct {
	singleGetter repo.SingleGetterGlobal
}

// NewRepository creates a new instance of PgRepository
func NewRepository() *PgRepository {
	return &PgRepository{
		singleGetter: repo.NewSingleGetterGlobal(resource.Schema, tableName, []string{schemaVersionColumn, schemaDirtyColumn}),
	}
}

type schemaVersion struct {
	Version string `db:"version"`
	Dirty   bool   `db:"dirty"`
}

// GetVersion returns the current schema version
func (r *PgRepository) GetVersion(ctx context.Context) (string, bool, error) {
	var version schemaVersion
	err := r.singleGetter.GetGlobal(ctx, repo.Conditions{}, repo.NoOrderBy, &version)
	if err != nil {
		return "", false, errors.Wrap(err, "while getting schema version")
	}

	return version.Version, version.Dirty, nil
}
