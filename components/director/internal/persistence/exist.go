package persistence

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
)

type ExistQuerier struct {
	Query string
}

func (g *ExistQuerier) Exists(ctx context.Context, tenant, id string) (bool, error) {
	persist, err := FromCtx(ctx)
	if err != nil {
		return false, errors.Wrap(err, "while fetching DB from context")
	}

	var count int
	err = persist.Get(&count, g.Query, id, tenant)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.Wrap(err, "while getting runtime from DB")
	}
	return true, nil
}

