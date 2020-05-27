package customerrors

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/lib/pq"
)

func MapSQLError(err error) error {
	if err == nil {
		return nil
	}
	if err == sql.ErrNoRows {
		return NewBuilder().WithStatusCode(NotFound).Build()
	}

	pqerr, ok := err.(*pq.Error)
	if !ok {
		return NewBuilder().InternalError("").Wrap(err).Build()
	}

	switch pqerr.Code {
	case persistence.UniqueViolation:
		return NewBuilder().WithStatusCode(NotUnique).Wrap(err).Build()
	case persistence.ForeignKeyViolation:
		return NewBuilder().WithStatusCode(ConstraintVolation).WithMessage(pqerr.Detail).Wrap(err).Build()
	}

	return NewBuilder().InternalError("").Wrap(err).Build()
}
