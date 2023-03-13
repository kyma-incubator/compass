package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
)

type pgxPool interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Ping(ctx context.Context) error
	Close()
}

type Connection struct {
	pool pgxPool
	cfg  config.Postgres
}

func NewConnection(ctx context.Context, cfg config.Postgres) (Connection, error) {
	connectCtx, connectCtxCancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer connectCtxCancel()

	pool, err := pgxpool.New(connectCtx, cfg.URI)
	if err != nil {
		return Connection{}, errors.Newf("failed to init postgres connection pool: %w", err)
	}

	return Connection{
		cfg:  cfg,
		pool: pool,
	}, nil
}

func (c Connection) Close() {
	c.pool.Close()
}

func (c Connection) Ping(ctx context.Context) error {
	return c.pool.Ping(ctx)
}
