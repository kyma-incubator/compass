package customerrors

import (
	"database/sql"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/lib/pq"
)

func MapSQLError(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	errMsg := fmt.Sprintf(format, args...)
	if err == sql.ErrNoRows {
		return newBuilder().withStatusCode(NotFound).withMessage(errMsg).build()
	}

	pgErr, ok := err.(*pq.Error)
	if !ok {
		return NewInternalError(errMsg)
	}

	switch pgErr.Code {
	case persistence.UniqueViolation:
		return newBuilder().withStatusCode(NotUnique).withMessage(errMsg).wrap(err).build()
	case persistence.ForeignKeyViolation:
		return newBuilder().withStatusCode(ConstraintViolation).withMessage(pgErr.Detail).wrap(err).build()
	}

	return newBuilder().internalError(fmt.Sprintf("SQL Error: %s", errMsg)).wrap(err).build()
}
