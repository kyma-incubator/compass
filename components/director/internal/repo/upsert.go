package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/lib/pq"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/pkg/errors"
)

type Upserter interface {
	Upsert(ctx context.Context, dbEntity interface{}) error
}

type universalUpserter struct {
	tableName          string
	insertColumns      []string
	conflictingColumns []string
	updateColumns      []string
}

func NewUpserter(tableName string, insertColumns []string, conflictingColumns []string, updateColumns []string) Upserter {
	return &universalUpserter{
		tableName:          tableName,
		insertColumns:      insertColumns,
		conflictingColumns: conflictingColumns,
		updateColumns:      updateColumns,
	}
}

func (u *universalUpserter) Upsert(ctx context.Context, dbEntity interface{}) error {
	if dbEntity == nil {
		return errors.New("item cannot be nil")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	var values []string
	for _, c := range u.insertColumns {
		values = append(values, fmt.Sprintf(":%s", c))
	}

	var update []string
	for _, c := range u.updateColumns {
		update = append(update, fmt.Sprintf("%[1]s=EXCLUDED.%[1]s", c))
	}

	stmtWithoutUpsert := fmt.Sprintf("INSERT INTO %s ( %s ) VALUES ( %s )", u.tableName, strings.Join(u.insertColumns, ", "), strings.Join(values, ", "))
	stmtWithUpsert := fmt.Sprintf("%s ON CONFLICT ( %s ) DO UPDATE SET %s", stmtWithoutUpsert, strings.Join(u.conflictingColumns, ", "), strings.Join(update, ", "))

	_, err = persist.NamedExec(stmtWithUpsert, dbEntity)
	if pqerr, ok := err.(*pq.Error); ok {
		if pqerr.Code == persistence.UniqueViolation {
			return apperrors.NewNotUniqueError("")
		}
	}

	return errors.Wrapf(err, "while upserting row to '%s' table", u.tableName)
}
