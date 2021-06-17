package repo

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
	"strings"
)

type UnionLister interface {
	//List stores the result into dest and returns the total count of tuples for each id from ids
	List(ctx context.Context, tenant string, ids []string, idsColumn string, pageSize int, cursor string, orderBy []string, dest Collection, additionalConditions ...Condition) ([]int, error)
	SetSelectedColumns(selectedColumns []string)
	Clone() UnionLister
}

type unionLister struct {
	tableName       string
	selectedColumns string
	tenantColumn    *string
	resourceType    resource.Type
}

func NewUnionLister(resourceType resource.Type, tableName string, tenantColumn string, selectedColumns []string) UnionLister {
	return &unionLister{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		tenantColumn:    &tenantColumn,
		resourceType:    resourceType,
	}
}

func (l *unionLister) SetSelectedColumns(selectedColumns []string) {
	l.selectedColumns = strings.Join(selectedColumns, ", ")
}

func (l *unionLister) Clone() UnionLister {
	var clonedLister unionLister

	clonedLister.resourceType = l.resourceType
	clonedLister.tableName = l.tableName
	clonedLister.selectedColumns = l.selectedColumns
	clonedLister.tenantColumn = l.tenantColumn

	return &clonedLister
}

func (l *unionLister) List(ctx context.Context, tenant string, ids []string, idscolumn string, pageSize int, cursor string, orderBy []string, dest Collection, additionalConditions ...Condition) ([]int, error) {
	if tenant == "" {
		return []int{0}, apperrors.NewTenantRequiredError()
	}
	additionalConditions = append(Conditions{NewEqualCondition(*l.tenantColumn, tenant)}, additionalConditions...)
	return l.unsafeList(ctx, pageSize, cursor, orderBy, ids, idscolumn, dest, additionalConditions...)
}

type queryStruct struct {
	args      []interface{}
	statement string
}

func (l *unionLister) unsafeList(ctx context.Context, pageSize int, cursor string, orderBy []string, ids []string, idsColumn string, dest Collection, conditions ...Condition) ([]int, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, err
	}

	offset, err := pagination.DecodeOffsetCursor(cursor)
	if err != nil {
		return nil, errors.Wrap(err, "while decoding page cursor")
	}

	queries, err := l.buildQueries(ids, idsColumn, conditions, orderBy, pageSize, offset)
	if err != nil {
		return nil, err
	}

	var stmts []string
	for _, q := range queries {
		stmts = append(stmts, q.statement)
	}

	var args [][]interface{}
	for _, q := range queries {
		args = append(args, q.args)
	}

	query, arguments, err := buildUnionQuery(stmts, args)
	if err != nil {
		return nil, err
	}

	err = persist.SelectContext(ctx, dest, query, arguments...)
	if err != nil {
		return nil, persistence.MapSQLError(ctx, err, l.resourceType, resource.List, "while fetching list page of objects from '%s' table", l.tableName)
	}

	totalCount, err := l.getTotalCount(ctx, persist, []string{idsColumn}, OrderByParams{NewAscOrderBy(idsColumn)}, conditions)
	if err != nil {
		return nil, err
	}

	return totalCount, nil
}

func (l *unionLister) buildQueries(ids []string, idsColumn string, conditions []Condition, orderBy []string, limit int, offset int) ([]queryStruct, error) {
	var queries []queryStruct
	for _, id := range ids {
		c := append(conditions, NewEqualCondition(idsColumn, id))
		query, args, err := buildSelectQueryWithLimitAndOffset(l.tableName, l.selectedColumns, c, parseOrderByParams(orderBy),limit,offset, false)
		if err != nil {
			return nil, errors.Wrap(err, "while building list query")
		}

		queries = append(queries, queryStruct{
			args:      args,
			statement: query,
		})
	}
	return queries, nil
}

func parseOrderByParams(orderBy []string) OrderByParams{
	var o OrderByParams
	for _, s := range orderBy {
		o = append(o, OrderBy{
			Field: s,
			Dir:   AscOrderBy,
		})
	}
	return o
}

func (l *unionLister) getTotalCount(ctx context.Context, persist persistence.PersistenceOp, groupBy GroupByParams, orderBy OrderByParams, conditions Conditions) ([]int, error) {
	query, args, err := buildCountQuery(l.tableName, conditions, groupBy, orderBy, true)
	if err != nil {
		return nil, err
	}

	var totalCount []int
	err = persist.SelectContext(ctx, &totalCount, query, args...)
	if err != nil {
		return nil, persistence.MapSQLError(ctx, err, l.resourceType, resource.List, "while counting objects from '%s' table", l.tableName)
	}
	return totalCount, nil
}
