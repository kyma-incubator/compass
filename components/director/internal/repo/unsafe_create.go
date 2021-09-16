package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// UnsafeCreator missing godoc
type UnsafeCreator interface {
	UnsafeCreate(ctx context.Context, dbEntity interface{}) error
}

type unsafeCreator struct {
	tableName          string
	resourceType       resource.Type
	insertColumns      []string
	conflictingColumns []string
}

// NewUnsafeCreator returns a new Creator which supports creation with conflicts.
func NewUnsafeCreator(resourceType resource.Type, tableName string, insertColumns []string, conflictingColumns []string) UnsafeCreator {
	return &unsafeCreator{
		resourceType:       resourceType,
		tableName:          tableName,
		insertColumns:      insertColumns,
		conflictingColumns: conflictingColumns,
	}
}

// UnsafeCreate adds a new entity in the Compass DB in case it does not exist. If it already exists, no action is taken.
func (u *unsafeCreator) UnsafeCreate(ctx context.Context, dbEntity interface{}) error {
	if dbEntity == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	values := make([]string, 0, len(u.insertColumns))
	for _, c := range u.insertColumns {
		values = append(values, fmt.Sprintf(":%s", c))
	}

	stmtWithoutUpsert := fmt.Sprintf("INSERT INTO %s ( %s ) VALUES ( %s )", u.tableName, strings.Join(u.insertColumns, ", "), strings.Join(values, ", "))
	stmtWithUpsert := fmt.Sprintf("%s ON CONFLICT ( %s ) DO NOTHING", stmtWithoutUpsert, strings.Join(u.conflictingColumns, ", "))

	log.C(ctx).Debugf("Executing DB query: %s", stmtWithUpsert)
	_, err = persist.NamedExecContext(ctx, stmtWithUpsert, dbEntity)
	return persistence.MapSQLError(ctx, err, u.resourceType, resource.Upsert, "while upserting row to '%s' table", u.tableName)
}
