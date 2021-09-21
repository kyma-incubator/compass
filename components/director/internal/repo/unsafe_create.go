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

// UnsafeCreator is used to create new entities in case they do not exist.
// In case they do already exist, no action is taken, hence the provided entity ID is not guaranteed to match the ID in the Compass DB.
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

	insertStmt := fmt.Sprintf("INSERT INTO %s ( %s ) VALUES ( %s )", u.tableName, strings.Join(u.insertColumns, ", "), strings.Join(values, ", "))
	stmt := fmt.Sprintf("%s ON CONFLICT ( %s ) DO NOTHING", insertStmt, strings.Join(u.conflictingColumns, ", "))

	log.C(ctx).Debugf("Executing DB query: %s", stmt)
	_, err = persist.NamedExecContext(ctx, stmt, dbEntity)
	return persistence.MapSQLError(ctx, err, u.resourceType, resource.Upsert, "while unsafe inserting row to '%s' table", u.tableName)
}
