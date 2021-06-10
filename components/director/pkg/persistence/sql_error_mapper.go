package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/lib/pq"
)

func MapSQLError(ctx context.Context, err error, resourceType resource.Type, sqlOperation resource.SQLOperation, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) {
		log.C(ctx).WithError(err).Errorf("Timeout error on SQL query: %v", err)
		return apperrors.NewInternalError("Maximum processing timeout reached")
	}

	if err == sql.ErrNoRows {
		log.C(ctx).WithError(err).Errorf("SQL: no rows in result set for '%s' resource type: %v", resourceType, err)
		return apperrors.NewNotFoundErrorWithType(resourceType)
	}

	pgErr, ok := err.(*pq.Error)
	if !ok {
		log.C(ctx).WithError(err).Errorf("Error while casting to postgres error: %v", err)
		return apperrors.NewInternalError("Unexpected error while executing SQL query")
	}

	log.C(ctx).WithError(pgErr).Errorf("SQL Error. Caused by: %s. DETAILS: %s", pgErr.Message, pgErr.Detail)

	switch pgErr.Code {
	case NotNullViolation:
		return apperrors.NewNotNullViolationError(resourceType)
	case CheckViolation:
		return apperrors.NewCheckViolationError(resourceType)
	case UniqueViolation:
		return apperrors.NewNotUniqueError(resourceType)
	case ForeignKeyViolation:
		return apperrors.NewForeignKeyInvalidOperationError(sqlOperation, resourceType)
	}

	return apperrors.NewInternalError("Unexpected error while executing SQL query")
}
