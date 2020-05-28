package customerrors

import (
	"database/sql"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/lib/pq"
)

func MapSQLError(err error, resourceType ResourceType, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	errMsg := fmt.Sprintf(format, args...)
	if err == sql.ErrNoRows {
		return newBuilder().withStatusCode(NotFound).withMessage("Object not found").with("object", string(resourceType)).build()
	}

	pgErr, ok := err.(*pq.Error)
	if !ok {
		return NewInternalError(errMsg)
	}

	switch pgErr.Code {
	case persistence.UniqueViolation:
		return NewNotUniqueErr(resourceType)
	case persistence.ForeignKeyViolation:
		return NewConstrainViolation(resourceType)
	}

	return newBuilder().internalError(fmt.Sprintf("SQL Error: %s", errMsg)).wrap(err).build()
}
