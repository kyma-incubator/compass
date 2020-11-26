package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/lib/pq"
)

func MapSQLError(ctx context.Context, err error, resourceType resource.Type, sqlOperation resource.SQLOperation, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	if err == sql.ErrNoRows {
		log.C(ctx).Errorf("SQL: no rows in result set for '%s' resource type", resourceType)
		return apperrors.NewNotFoundErrorWithType(resourceType)
	}

	pgErr, ok := err.(*pq.Error)
	if !ok {
		log.C(ctx).Errorf("Error while casting to postgres error. Actual error: %s", err)
		return apperrors.NewInternalError("Unexpected error while executing SQL query")
	}

	log.C(ctx).Errorf("SQL Error: %s. Caused by: %s. DETAILS: %s", fmt.Sprintf(format, args...), pgErr.Message, pgErr.Detail)

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
