package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type Creator struct {
	tableName string
	columns   []string
	statement string
}

func NewCreator(tableName string, fields []string) *Creator {
	return &Creator{
		tableName: tableName,
		columns:   fields,
	}
}

func (c *Creator) Create(ctx context.Context, dbEntity interface{}) error {
	if dbEntity == nil {
		return errors.New("item cannot be nil")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	var fields []string
	for _, c := range c.columns {
		fields = append(fields, fmt.Sprintf(":%s", c))
	}

	stmt := fmt.Sprintf("INSERT INTO %s ( %s ) VALUES ( %s )", c.tableName, strings.Join(c.columns, ", "), strings.Join(fields, ", "))

	_, err = persist.NamedExec(stmt, dbEntity)
	if pqerr, ok := err.(*pq.Error); ok {
		if pqerr.Code == persistence.UniqueViolation {
			return &notUniqueError{}
		}
	}

	return errors.Wrapf(err, "while inserting row to '%s' table", c.tableName)
}
