package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/pkg/errors"
	"strings"
)

type SingleGetter struct {
	Query string
}

type NotFoundError struct {
	ObjectType string

}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("")
}

func (g *SingleGetter) Get(ctx context.Context, dest interface{}, args ...interface{}) error {
	if dest == nil {
		return errors.New("missing destination")
	}
	persist, err := FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching DB from context")
	}

	err = persist.Get(dest, g.Query, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return errors.Wrap(err, "while getting runtime from DB")
		}

		return &NotFoundError{}
	}

	return nil
}

type Deleter struct {
	Query string
}

func (g *Deleter) Delete(ctx context.Context, id string) error {
	persist, err := FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching persistence from context")
	}

	_, err = persist.Exec(g.Query, id)

	return errors.Wrap(err, "while deleting the runtime entity from database")
}

type PageableQuerier struct {
	Query        string
	Columns      string
	RelationName string // TODO
}

type Collection interface {
	Len() int
}

func (g *PageableQuerier) List(ctx context.Context, tenant string, pageSize int, cursor string, dest Collection, additionalConditions ...string) (*pagination.Page, int, error) {
	persist, err := FromCtx(ctx)
	if err != nil {
		return nil, -1, errors.Wrap(err, "while fetching DB from context")
	}

	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, -1, errors.Wrap(err, "while parsing tenant as UUID")
	}

	offset, err := pagination.DecodeOffsetCursor(cursor)
	if err != nil {
		return nil, -1, errors.Wrap(err, "while decoding page cursor")
	}

	filterSubquery := ""
	for _, cond := range additionalConditions {
		filterSubquery = fmt.Sprintf(` AND %s`, cond)
	}

	paginationSQL, err := pagination.ConvertOffsetLimitAndOrderedColumnToSQL(pageSize, offset, "id")
	if err != nil {
		return nil, -1, errors.Wrap(err, "while converting offset and limit to cursor")
	}

	stmt := fmt.Sprintf("%s %s %s", g.Query, filterSubquery, paginationSQL)

	err = persist.Select(dest, stmt, tenant)
	if err != nil {
		return nil, -1, errors.Wrap(err, "while fetching runtimes from DB")
	}

	totalCount, err := g.getTotalCount(persist, stmt, tenantID)
	if err != nil {
		return nil, -1, errors.Wrap(err, "while getting total count of runtimes")
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

func (g *PageableQuerier) getTotalCount(persist PersistenceOp, query string, tenantUUID uuid.UUID) (int, error) {
	stmt := strings.Replace(query, g.Columns, "COUNT (*)", 1)
	var totalCount int
	err := persist.Get(&totalCount, stmt, tenantUUID.String())
	if err != nil {
		return -1, errors.Wrap(err, "while counting runtimes")
	}

	return totalCount, nil
}
