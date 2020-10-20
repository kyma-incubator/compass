package repo

import (
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

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

	log.Debug("Building DB query...")
	stmt := fmt.Sprintf("INSERT INTO %s ( %s ) VALUES ( %s )", c.tableName, strings.Join(c.columns, ", "), strings.Join(values, ", "))

	log.Debugf("Executing DB query: %s", stmt)
	_, err = persist.NamedExec(stmt, dbEntity)

	return persistence.MapSQLError(err, c.resourceType, "while inserting row to '%s' table", c.tableName)
}
