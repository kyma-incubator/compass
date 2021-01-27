package repo

import (
	"context"
	"fmt"
	"strings"

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

	sqlQuery := fmt.Sprintf("INSERT INTO %s ( %s ) VALUES ( %s ) RETURNING id;", c.tableName, strings.Join(c.columns, ", "), strings.Join(values, ", "))

	stmt, err := persist.PrepareNamedContext(ctx, sqlQuery)
	if err != nil {
		return err
	}

	resultDto := &struct {
		ID string `db:"id"`
	}{}

	log.C(ctx).Debugf("Executing DB query: %s", sqlQuery)
	err = stmt.GetContext(ctx, resultDto, dbEntity)

	opMode := operation.ModeFromCtx(ctx)
	if opMode == graphql.OperationModeAsync {
		operations, exists := operation.FromCtx(ctx)
		if !exists {
			return apperrors.NewInternalError("unable to fetch operations from context")
		}

		op := (*operations)[len(*operations)-1]

		relatedResource := operation.RelatedResource{
			ResourceType: c.tableName,
			ResourceID:   resultDto.ID,
		}

		op.RelatedResources = append(op.RelatedResources, relatedResource)
	}

	return persistence.MapSQLError(ctx, err, c.resourceType, resource.Create, "while inserting row to '%s' table", c.tableName)
}
