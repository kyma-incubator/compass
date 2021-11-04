package repo

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// PageableQuerier missing godoc
type PageableQuerier interface {
	List(ctx context.Context, tenant string, pageSize int, cursor string, orderByColumn string, dest Collection, additionalConditions ...Condition) (*pagination.Page, int, error)
}

// PageableQuerierGlobal missing godoc
type PageableQuerierGlobal interface {
	ListGlobal(ctx context.Context, pageSize int, cursor string, orderByColumn string, dest Collection, additionalConditions ...Condition) (*pagination.Page, int, error)
}

type universalPageableQuerier struct {
	tableName       string
	selectedColumns string
	tenantColumn    *string
	resourceType    resource.Type
}

// NewPageableQuerierWithEmbeddedTenant missing godoc
func NewPageableQuerierWithEmbeddedTenant(resourceType resource.Type, tableName string, tenantColumn string, selectedColumns []string) PageableQuerier {
	return &universalPageableQuerier{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		tenantColumn:    &tenantColumn,
		resourceType:    resourceType,
	}
}

// NewPageableQuerier missing godoc
func NewPageableQuerier(resourceType resource.Type, tableName string, selectedColumns []string) PageableQuerier {
	return &universalPageableQuerier{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		resourceType:    resourceType,
	}
}

// NewPageableQuerierGlobal missing godoc
func NewPageableQuerierGlobal(resourceType resource.Type, tableName string, selectedColumns []string) PageableQuerierGlobal {
	return &universalPageableQuerier{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		resourceType:    resourceType,
	}
}

// Collection missing godoc
type Collection interface {
	Len() int
}

// List returns Page, TotalCount or error
func (g *universalPageableQuerier) List(ctx context.Context, tenant string, pageSize int, cursor string, orderByColumn string, dest Collection, additionalConditions ...Condition) (*pagination.Page, int, error) {
	if tenant == "" {
		return nil, -1, apperrors.NewTenantRequiredError()
	}

	if g.tenantColumn != nil {
		additionalConditions = append(Conditions{NewEqualCondition(*g.tenantColumn, tenant)}, additionalConditions...)
		return g.unsafeList(ctx, tenant, pageSize, cursor, orderByColumn, dest, additionalConditions...)
	}

	tenantIsolation, err := NewTenantIsolationCondition(g.resourceType, tenant, false)
	if err != nil {
		return nil, -1, err
	}

	additionalConditions = append(additionalConditions, tenantIsolation)

	return g.unsafeList(ctx, tenant, pageSize, cursor, orderByColumn, dest, additionalConditions...)
}

// ListGlobal missing godoc
func (g *universalPageableQuerier) ListGlobal(ctx context.Context, pageSize int, cursor string, orderByColumn string, dest Collection, additionalConditions ...Condition) (*pagination.Page, int, error) {
	return g.unsafeList(ctx, "", pageSize, cursor, orderByColumn, dest, additionalConditions...)
}

func (g *universalPageableQuerier) unsafeList(ctx context.Context, tenant string, pageSize int, cursor string, orderByColumn string, dest Collection, conditions ...Condition) (*pagination.Page, int, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, -1, err
	}

	offset, err := pagination.DecodeOffsetCursor(cursor)
	if err != nil {
		return nil, -1, errors.Wrap(err, "while decoding page cursor")
	}

	paginationSQL, err := pagination.ConvertOffsetLimitAndOrderedColumnToSQL(pageSize, offset, orderByColumn)
	if err != nil {
		return nil, -1, errors.Wrap(err, "while converting offset and limit to cursor")
	}

	query, args, err := buildSelectQuery(g.resourceType, g.tableName, g.selectedColumns, tenant, conditions, OrderByParams{}, true)
	if err != nil {
		return nil, -1, errors.Wrap(err, "while building list query")
	}

	// TODO: Refactor query builder
	stmtWithPagination := fmt.Sprintf("%s %s", query, paginationSQL)

	err = persist.SelectContext(ctx, dest, stmtWithPagination, args...)
	if err != nil {
		return nil, -1, persistence.MapSQLError(ctx, err, g.resourceType, resource.List, "while fetching list page of objects from '%s' table", g.tableName)
	}

	totalCount, err := g.getTotalCount(ctx, persist, query, args)
	if err != nil {
		return nil, -1, err
	}

	hasNextPage := false
	endCursor := ""
	if totalCount > offset+dest.Len() {
		hasNextPage = true
		endCursor = pagination.EncodeNextOffsetCursor(offset, pageSize)
	}
	return &pagination.Page{
		StartCursor: cursor,
		EndCursor:   endCursor,
		HasNextPage: hasNextPage,
	}, totalCount, nil
}

func (g *universalPageableQuerier) getTotalCount(ctx context.Context, persist persistence.PersistenceOp, query string, args []interface{}) (int, error) {
	stmt := strings.Replace(query, g.selectedColumns, "COUNT(*)", 1)
	var totalCount int
	err := persist.GetContext(ctx, &totalCount, stmt, args...)
	if err != nil {
		return -1, errors.Wrap(err, "while counting objects")
	}
	return totalCount, nil
}
