package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

type Creator interface {
	Create(ctx context.Context, dbEntity interface{}) error
}

type universalCreator struct {
	tableName    string
	resourceType resource.Type
	columns      []string
}

func NewCreator(resourceType resource.Type, tableName string, columns []string) Creator {
	return &universalCreator{
		resourceType: resourceType,
		tableName:    tableName,
		columns:      columns,
	}
}

func (c *universalCreator) Create(ctx context.Context, dbEntity interface{}) error {
	if dbEntity == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	var values []string
	for _, c := range c.columns {
		values = append(values, fmt.Sprintf(":%s", c))
	}

	entity, ok := dbEntity.(Entity)
	if ok && entity.GetCreatedAt().IsZero() { // This zero check is needed to mock the Create tests
		now := time.Now()
		entity.SetCreatedAt(now)
		entity.SetReady(true)

		if operation.ModeFromCtx(ctx) == graphql.OperationModeAsync {
			entity.SetReady(false)
		}

		dbEntity = entity
	}

	stmt := fmt.Sprintf("INSERT INTO %s ( %s ) VALUES ( %s )", c.tableName, strings.Join(c.columns, ", "), strings.Join(values, ", "))

	log.C(ctx).Debugf("Executing DB query: %s", stmt)
	_, err = persist.NamedExec(stmt, dbEntity)

	return persistence.MapSQLError(ctx, err, c.resourceType, resource.Create, "while inserting row to '%s' table", c.tableName)
}
