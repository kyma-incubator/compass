package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/pkg/errors"
)

type PageableQuerier interface {
	List(ctx context.Context, tenant string, pageSize int, cursor string, orderByColumn string, dest Collection, additionalConditions ...string) (*pagination.Page, int, error)
}

type PageableQuerierGlobal interface {
	ListGlobal(ctx context.Context, pageSize int, cursor string, orderByColumn string, dest Collection, additionalConditions ...string) (*pagination.Page, int, error)
}

type universalPageableQuerier struct {
	tableName       string
	selectedColumns string
	tenantColumn    *string
}

func NewPageableQuerier(tableName string, tenantColumn string, selectedColumns []string) PageableQuerier {
	return &universalPageableQuerier{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		tenantColumn:    &tenantColumn,
	}
}

func NewPageableQuerierGlobal(tableName string, selectedColumns []string) PageableQuerierGlobal {
	return &universalPageableQuerier{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
	}
}

type Collection interface {
	Len() int
}

// List returns Page, TotalCount or error
func (g *universalPageableQuerier) List(ctx context.Context, tenant string, pageSize int, cursor string, orderByColumn string, dest Collection, additionalConditions ...string) (*pagination.Page, int, error) {
	return g.unsafeList(ctx, str.Ptr(tenant), pageSize, cursor, orderByColumn, dest, additionalConditions...)
}

func (g *universalPageableQuerier) ListGlobal(ctx context.Context, pageSize int, cursor string, orderByColumn string, dest Collection, additionalConditions ...string) (*pagination.Page, int, error) {
	return g.unsafeList(ctx, nil, pageSize, cursor, orderByColumn, dest, additionalConditions...)
}

func (g *universalPageableQuerier) unsafeList(ctx context.Context, tenant *string, pageSize int, cursor string, orderByColumn string, dest Collection, additionalConditions ...string) (*pagination.Page, int, error) {
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

	stmtWithoutPagination := buildSelectStatement(g.selectedColumns, g.tableName, g.tenantColumn, additionalConditions)
	stmtWithPagination := fmt.Sprintf("%s %s", stmtWithoutPagination, paginationSQL)

	var args []interface{}
	if tenant != nil {
		args = append(args, *tenant)
	}

	err = persist.Select(dest, stmtWithPagination, args...)
	if err != nil {
		return nil, -1, errors.Wrap(err, "while fetching list of objects from DB")
	}

	totalCount, err := g.getTotalCount(persist, stmtWithoutPagination, args)
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

func (g *universalPageableQuerier) getTotalCount(persist persistence.PersistenceOp, query string, args []interface{}) (int, error) {
	stmt := strings.Replace(query, g.selectedColumns, "COUNT(*)", 1)
	var totalCount int
	err := persist.Get(&totalCount, stmt, args...)
	if err != nil {
		return -1, errors.Wrap(err, "while counting objects")
	}

	return totalCount, nil
}
