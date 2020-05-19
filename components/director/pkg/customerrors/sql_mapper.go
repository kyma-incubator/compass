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
		return NewErrorBuilder(NotFound).Build()
	}

	pqerr, ok := err.(*pq.Error)
	if !ok {
		return NewErrorBuilder(InternalError).Wrap(err).Build()
	}

	switch pqerr.Code {
	case persistence.UniqueViolation:
		return NewErrorBuilder(NotUnique).Wrap(err).Build()
	case persistence.ForeignKeyViolation:
		return NewErrorBuilder(ConstaintVolation).WithMessage(pqerr.Detail).Wrap(err).Build()
	}

	return NewErrorBuilder(InternalError).Wrap(err).Build()
}
