package persistence

import (
	"database/sql"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/lib/pq"
)

func MapSQLError(err error, resourceType resource.Type, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	if err == sql.ErrNoRows {
		return apperrors.NewNotFoundErrorWithType(resourceType)
	}

	pgErr, ok := err.(*pq.Error)
	if !ok {
		return apperrors.InternalErrorFrom(err, format, args...)
	}

	switch pgErr.Code {
	case UniqueViolation:
		return apperrors.NewNotUniqueError(resourceType)
	case ForeignKeyViolation:
		return apperrors.NewInvalidDataError("Object already exist")
	}

	return apperrors.InternalErrorFrom(err, "SQL Error: %s", fmt.Sprintf(format, args...))
}
