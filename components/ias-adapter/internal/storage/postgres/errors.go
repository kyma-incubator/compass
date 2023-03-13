package postgres

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
)

const uniqueViolation = "23505"

func dbError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.Newf("%s: %w", err, errors.EntityNotFound)
	}

	var pgxErr *pgconn.PgError
	if ok := errors.As(err, &pgxErr); ok {
		switch pgxErr.Code {
		case uniqueViolation:
			return errors.Newf("%s: %w", err, errors.EntityAlreadyExists)
		}
	}

	return err
}
