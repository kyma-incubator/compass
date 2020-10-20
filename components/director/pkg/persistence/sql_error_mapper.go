package persistence

import (
	"database/sql"
	"fmt"
	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/lib/pq"
)

func MapSQLError(err error, resourceType resource.Type, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	if err == sql.ErrNoRows {
		log.Warnf("SQL: no rows in result set for '%s' table", resourceType)
		return apperrors.NewNotFoundErrorWithType(resourceType)
	}

	pgErr, ok := err.(*pq.Error)
	if !ok {
		log.Warn("Error while getting postgres error code")
		return apperrors.InternalErrorFrom(err, format, args...)
	}

	log.Debug("Checking Postgres error code...")
	switch pgErr.Code {
	case UniqueViolation:
		log.Warn("Postgres unique violation error code found")
		return apperrors.NewNotUniqueError(resourceType)
	case ForeignKeyViolation:
		log.Warn("Postgres foreign key violation error found")
		return apperrors.NewInvalidDataError("Object already exist")
	}

	return apperrors.InternalErrorFrom(err, "SQL Error: %s", fmt.Sprintf(format, args...))
}
