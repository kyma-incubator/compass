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

// Upserter is an interface for upserting global entities without tenant or entities with tenant embedded in them.
type Upserter interface {
	Upsert(ctx context.Context, dbEntity interface{}) error
}

type universalUpserter struct {
	tableName          string
	resourceType       resource.Type
	insertColumns      []string
	conflictingColumns []string
	updateColumns      []string
}

// NewUpserter is a constructor for Upserter about entities without tenant or entities with tenant embedded in them.
func NewUpserter(resourceType resource.Type, tableName string, insertColumns []string, conflictingColumns []string, updateColumns []string) Upserter {
	return &universalUpserter{
		resourceType:       resourceType,
		tableName:          tableName,
		insertColumns:      insertColumns,
		conflictingColumns: conflictingColumns,
		updateColumns:      updateColumns,
	}
}

// Upsert adds a new entity in the Compass DB in case it does not exist. If it already exists, updates it.
// This upserter is not suitable for resources that have m2m tenant relation as it does not maintain tenant accesses.
// Use it for global scoped resources or resources with embedded tenant_id only.
func (u *universalUpserter) Upsert(ctx context.Context, dbEntity interface{}) error {
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

	update := make([]string, 0, len(u.updateColumns))
	for _, c := range u.updateColumns {
		update = append(update, fmt.Sprintf("%[1]s=EXCLUDED.%[1]s", c))
	}

	stmtWithoutUpsert := fmt.Sprintf("INSERT INTO %s ( %s ) VALUES ( %s )", u.tableName, strings.Join(u.insertColumns, ", "), strings.Join(values, ", "))
	stmtWithUpsert := fmt.Sprintf("%s ON CONFLICT ( %s ) DO UPDATE SET %s", stmtWithoutUpsert, strings.Join(u.conflictingColumns, ", "), strings.Join(update, ", "))

	log.C(ctx).Debugf("Executing DB query: %s", stmtWithUpsert)
	_, err = persist.NamedExecContext(ctx, stmtWithUpsert, dbEntity)
	return persistence.MapSQLError(ctx, err, u.resourceType, resource.Upsert, "while upserting row to '%s' table", u.tableName)
}
