package schema

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const (
	schemaVersionColumn = "version"
	tableName           = `"public"."schema_migrations"`
)

// PgRepository missing godoc
type PgRepository struct {
	singleGetter repo.SingleGetterGlobal
}

// NewRepository missing godoc
func NewRepository() *PgRepository {
	return &PgRepository{
		singleGetter: repo.NewSingleGetterGlobal(resource.Schema, tableName, []string{schemaVersionColumn}),
	}
}

// GetVersion missing godoc
func (r *PgRepository) GetVersion(ctx context.Context) (string, error) {
	var version string
	err := r.singleGetter.GetGlobal(ctx, repo.Conditions{}, repo.NoOrderBy, &version)
	if err != nil {
		return "", errors.Wrap(err, "while getting schema version")
	}

	return version, nil
}
