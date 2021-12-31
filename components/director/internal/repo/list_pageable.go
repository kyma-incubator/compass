package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// PageableQuerier is an interface for listing with paging of tenant scoped entities with either externally managed tenant accesses (m2m table or view) or embedded tenant in them.
type PageableQuerier interface {
	List(ctx context.Context, resourceType resource.Type, tenant string, pageSize int, cursor string, orderByColumn string, dest Collection, additionalConditions ...Condition) (*pagination.Page, int, error)
}

// PageableQuerierGlobal is an interface for listing with paging of global entities.
type PageableQuerierGlobal interface {
	ListGlobal(ctx context.Context, pageSize int, cursor string, orderByColumn string, dest Collection, additionalConditions ...Condition) (*pagination.Page, int, error)
	ListWithQueryGlobal(ctx context.Context, query string, args []interface{}, orderByColumn string, pageSize int, cursor string, dest Collection) (*pagination.Page, int, error)
}

type universalPageableQuerier struct {
	tableName       string
	selectedColumns string
	tenantColumn    *string
	resourceType    resource.Type
}

// NewPageableQuerierWithEmbeddedTenant is a constructor for PageableQuerier about entities with tenant embedded in them.
func NewPageableQuerierWithEmbeddedTenant(tableName string, tenantColumn string, selectedColumns []string) PageableQuerier {
	return &universalPageableQuerier{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		tenantColumn:    &tenantColumn,
	}
}

// NewPageableQuerier is a constructor for PageableQuerier about entities with externally managed tenant accesses (m2m table or view)
func NewPageableQuerier(tableName string, selectedColumns []string) PageableQuerier {
	return &universalPageableQuerier{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
	}
}

// NewPageableQuerierGlobal is a constructor for PageableQuerierGlobal about global entities.
func NewPageableQuerierGlobal(resourceType resource.Type, tableName string, selectedColumns []string) PageableQuerierGlobal {
	return &universalPageableQuerier{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		resourceType:    resourceType,
	}
}

// Collection is an interface for a collection of entities.
type Collection interface {
	Len() int
}

// List lists a page of tenant scoped entities with tenant isolation subquery.
// If the tenantColumn is configured the isolation is based on equal condition on tenantColumn.
// If the tenantColumn is not configured an entity with externally managed tenant accesses in m2m table / view is assumed.
func (g *universalPageableQuerier) List(ctx context.Context, resourceType resource.Type, tenant string, pageSize int, cursor string, orderByColumn string, dest Collection, additionalConditions ...Condition) (*pagination.Page, int, error) {
	if tenant == "" {
		return nil, -1, apperrors.NewTenantRequiredError()
	}

	if g.tenantColumn != nil {
		additionalConditions = append(Conditions{NewEqualCondition(*g.tenantColumn, tenant)}, additionalConditions...)
		return g.list(ctx, resourceType, pageSize, cursor, orderByColumn, dest, additionalConditions...)
	}

	tenantIsolation, err := NewTenantIsolationCondition(resourceType, tenant, false)
	if err != nil {
		return nil, -1, err
	}

	additionalConditions = append(additionalConditions, tenantIsolation)

	return g.list(ctx, resourceType, pageSize, cursor, orderByColumn, dest, additionalConditions...)
}

// ListGlobal lists a page of global entities without tenant isolation.
func (g *universalPageableQuerier) ListGlobal(ctx context.Context, pageSize int, cursor string, orderByColumn string, dest Collection, additionalConditions ...Condition) (*pagination.Page, int, error) {
	return g.list(ctx, g.resourceType, pageSize, cursor, orderByColumn, dest, additionalConditions...)
}

// ListWithQueryGlobal lists a page of global entities with a given sql query but without tenant isolation.
func (g *universalPageableQuerier) ListWithQueryGlobal(ctx context.Context, query string, args []interface{}, orderByColumn string, pageSize int, cursor string, dest Collection) (*pagination.Page, int, error) {
	return g.listWithQuery(ctx, query, args, orderByColumn, g.resourceType, pageSize, cursor, dest)
}

func (g *universalPageableQuerier) list(ctx context.Context, resourceType resource.Type, pageSize int, cursor string, orderByColumn string, dest Collection, conditions ...Condition) (*pagination.Page, int, error) {
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

	query, args, err := buildSelectQuery(g.tableName, g.selectedColumns, conditions, OrderByParams{}, true)
	if err != nil {
		return nil, -1, errors.Wrap(err, "while building list query")
	}

	// TODO: Refactor query builder
	stmtWithPagination := fmt.Sprintf("%s %s", query, paginationSQL)

	err = persist.SelectContext(ctx, dest, stmtWithPagination, args...)
	if err != nil {
		return nil, -1, persistence.MapSQLError(ctx, err, resourceType, resource.List, "while fetching list page of objects from '%s' table", g.tableName)
	}

	totalCount, err := g.getTotalCount(ctx, resourceType, persist, query, args)
	if err != nil {
		return nil, -1, err
	}

	hasNextPage, endCursor := g.getNextPageAndCursor(totalCount, offset, pageSize, dest.Len())
	return &pagination.Page{
		StartCursor: cursor,
		EndCursor:   endCursor,
		HasNextPage: hasNextPage,
	}, totalCount, nil
}

func (g *universalPageableQuerier) listWithQuery(ctx context.Context, query string, args []interface{}, orderedColumn string, resourceType resource.Type, pageSize int, cursor string, dest Collection) (*pagination.Page, int, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, -1, err
	}

	offset, err := pagination.DecodeOffsetCursor(cursor)
	if err != nil {
		return nil, -1, errors.Wrap(err, "while decoding page cursor")
	}

	paginationSQL, err := pagination.ConvertOffsetLimitAndOrderedColumnToSQL(pageSize, offset, orderedColumn)
	if err != nil {
		return nil, -1, errors.Wrap(err, "while converting offset and limit to cursor")
	}

	stmtWithPagination := fmt.Sprintf("%s %s", query, paginationSQL)
	err = persist.SelectContext(ctx, dest, stmtWithPagination, args...)
	if err != nil {
		return nil, -1, persistence.MapSQLError(ctx, err, resourceType, resource.List, "while fetching list page of objects from '%s' table", g.tableName)
	}

	totalCount, err := g.getTotalCount(ctx, resourceType, persist, query, args)
	if err != nil {
		return nil, -1, err
	}

	hasNextPage, endCursor := g.getNextPageAndCursor(totalCount, offset, pageSize, dest.Len())
	return &pagination.Page{
		StartCursor: cursor,
		EndCursor:   endCursor,
		HasNextPage: hasNextPage,
	}, totalCount, nil
}

func (g *universalPageableQuerier) getNextPageAndCursor(totalCount, offset, pageSize, totalLen int) (bool, string) {
	hasNextPage := false
	endCursor := ""
	if totalCount > offset+totalLen {
		hasNextPage = true
		endCursor = pagination.EncodeNextOffsetCursor(offset, pageSize)
	}
	return hasNextPage, endCursor
}

func (g *universalPageableQuerier) getTotalCount(ctx context.Context, resourceType resource.Type, persist persistence.PersistenceOp, query string, args []interface{}) (int, error) {
	stmt := strings.Replace(query, g.selectedColumns, "COUNT(*)", 1)
	var totalCount int
	err := persist.GetContext(ctx, &totalCount, stmt, args...)
	if err != nil {
		return -1, persistence.MapSQLError(ctx, err, resourceType, resource.List, "while counting objects from '%s' table", g.tableName)
	}
	return totalCount, nil
}
