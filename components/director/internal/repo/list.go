package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/pkg/errors"
)

type PageableQuerier struct {
	tableName       string
	selectedColumns string
	tenantColumn    string
}

func NewPageableQuerier(tableName, tenantColumn, selectedColumns string) *PageableQuerier {
	return &PageableQuerier{
		tableName:       tableName,
		selectedColumns: selectedColumns,
		tenantColumn:    tenantColumn,
	}
}

type Collection interface {
	Len() int
}

func (g *PageableQuerier) List(ctx context.Context, tenant string, pageSize int, cursor string, orderByColumn string, dest Collection, additionalConditions ...string) (*pagination.Page, int, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, -1, err
	}

	offset, err := pagination.DecodeOffsetCursor(cursor)
	if err != nil {
		return nil, -1, errors.Wrap(err, "while decoding page cursor")
	}

	filterSubquery := ""
	for _, cond := range additionalConditions {
		if strings.TrimSpace(cond) != "" {
			filterSubquery += fmt.Sprintf(` AND %s`, cond)
		}
	}

	paginationSQL, err := pagination.ConvertOffsetLimitAndOrderedColumnToSQL(pageSize, offset, orderByColumn)
	if err != nil {
		return nil, -1, errors.Wrap(err, "while converting offset and limit to cursor")
	}

	stmtWithoutPagination := fmt.Sprintf("SELECT %s FROM %s WHERE %s=$1 %s", g.selectedColumns, g.tableName, g.tenantColumn, filterSubquery)
	stmtWithPagination := fmt.Sprintf("%s %s", stmtWithoutPagination, paginationSQL)

	err = persist.Select(dest, stmtWithPagination, tenant)
	if err != nil {
		return nil, -1, errors.Wrap(err, "while fetching list of objects from DB")
	}

	totalCount, err := g.getTotalCount(persist, stmtWithoutPagination, tenant)
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

func (g *PageableQuerier) getTotalCount(persist persistence.PersistenceOp, query string, tenant string) (int, error) {
	stmt := strings.Replace(query, g.selectedColumns, "COUNT(*)", 1)
	var totalCount int
	err := persist.Get(&totalCount, stmt, tenant)
	if err != nil {
		return -1, errors.Wrap(err, "while counting objects")
	}

	return totalCount, nil
}
