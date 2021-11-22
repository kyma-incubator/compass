package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// ExistQuerier missing godoc
type ExistQuerier interface {
	Exists(ctx context.Context, tenant string, conditions Conditions) (bool, error)
}

// ExistQuerierGlobal missing godoc
type ExistQuerierGlobal interface {
	ExistsGlobal(ctx context.Context, conditions Conditions) (bool, error)
}

type universalExistQuerier struct {
	tableName    string
	tenantColumn *string
	resourceType resource.Type
}

// NewExistQuerier missing godoc
func NewExistQuerier(resourceType resource.Type, tableName string, tenantColumn string) ExistQuerier {
	return &universalExistQuerier{resourceType: resourceType, tableName: tableName, tenantColumn: &tenantColumn}
}

// NewExistQuerierGlobal missing godoc
func NewExistQuerierGlobal(resourceType resource.Type, tableName string) ExistQuerierGlobal {
	return &universalExistQuerier{tableName: tableName, resourceType: resourceType}
}

// Exists missing godoc
func (g *universalExistQuerier) Exists(ctx context.Context, tenantID string, conditions Conditions) (bool, error) {
	if tenantID == "" {
		return false, apperrors.NewTenantRequiredError()
	}

	switch tenant.LoadIsolationTypeFromContext(ctx) {
	case tenant.RecursiveIsolationType:
		conditions = append(Conditions{NewTenantIsolationCondition(*g.tenantColumn, tenantID)}, conditions...)
	case tenant.SimpleIsolationType:
		conditions = append(Conditions{NewEqualCondition(*g.tenantColumn, tenantID)}, conditions...)
	}

	return g.unsafeExists(ctx, conditions)
}

// ExistsGlobal missing godoc
func (g *universalExistQuerier) ExistsGlobal(ctx context.Context, conditions Conditions) (bool, error) {
	return g.unsafeExists(ctx, conditions)
}

func (g *universalExistQuerier) unsafeExists(ctx context.Context, conditions Conditions) (bool, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return false, err
	}

	var stmtBuilder strings.Builder

	stmtBuilder.WriteString(fmt.Sprintf("SELECT 1 FROM %s", g.tableName))
	if len(conditions) > 0 {
		stmtBuilder.WriteString(" WHERE")
	}

	err = writeEnumeratedConditions(&stmtBuilder, conditions)
	if err != nil {
		return false, errors.Wrap(err, "while writing enumerated conditions")
	}
	allArgs := getAllArgs(conditions)

	query := getQueryFromBuilder(stmtBuilder)

	log.C(ctx).Debugf("Executing DB query: %s", query)
	var count int
	err = persist.GetContext(ctx, &count, query, allArgs...)
	err = persistence.MapSQLError(ctx, err, g.resourceType, resource.Exists, "while getting object from '%s' table", g.tableName)

	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
