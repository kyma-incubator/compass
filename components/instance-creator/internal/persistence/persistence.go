package persistence

import (
	"context"
	"crypto/sha256"
	"github.com/jmoiron/sqlx"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	_ "github.com/lib/pq"
	"math/big"
	"time"
)

// RetryCount is a number of retries when trying to open the database
const RetryCount int = 50

// AdvisoryLocker represents Advisory locking in Postgres
//
//go:generate mockery --name=AdvisoryLocker --output=automock --outpkg=automock --case=underscore --disable-version-string
type AdvisoryLocker interface {
	Lock(ctx context.Context, key string) error
	Unlock(ctx context.Context, key string) error
	TryLock(ctx context.Context, key string) (bool, error)
}

// DatabaseConnector returns database connection
//
//go:generate mockery --name=DatabaseConnector --output=automock --outpkg=automock --case=underscore --disable-version-string
type DatabaseConnector interface {
	GetConnection(ctx context.Context) (Connection, error)
}

// Connection represents database connection
//
//go:generate mockery --name=Connection --output=automock --outpkg=automock --case=underscore --disable-version-string
type Connection interface {
	GetAdvisoryLocker() AdvisoryLocker
	Close() error
}

func Configure(ctx context.Context, conf persistence.DatabaseConfig) (DatabaseConnector, func() error, error) {
	return waitForPersistence(ctx, conf, RetryCount)
}

func waitForPersistence(ctx context.Context, conf persistence.DatabaseConfig, retryCount int) (DatabaseConnector, func() error, error) {
	var sqlxDB *sqlx.DB
	var err error

	for i := 0; i < retryCount; i++ {
		if i > 0 {
			time.Sleep(5 * time.Second)
		}
		log.C(ctx).Info("Trying to connect to DB...")

		sqlxDB, err = sqlx.Open("postgres", conf.GetConnString())
		if err != nil {
			return nil, nil, err
		}
		ctxWithTimeout, cancelFunc := context.WithTimeout(ctx, time.Second)
		err = sqlxDB.PingContext(ctxWithTimeout)
		cancelFunc()
		if err != nil {
			log.C(ctx).Infof("Got error on pinging DB: %v", err)
			continue
		}

		log.C(ctx).Infof("Configuring MaxOpenConnections: [%d], MaxIdleConnections: [%d], ConnectionMaxLifetime: [%s]", conf.MaxOpenConnections, conf.MaxIdleConnections, conf.ConnMaxLifetime.String())
		sqlxDB.SetMaxOpenConns(conf.MaxOpenConnections)
		sqlxDB.SetMaxIdleConns(conf.MaxIdleConnections)
		sqlxDB.SetConnMaxLifetime(conf.ConnMaxLifetime)
		return &db{sqlDB: sqlxDB}, sqlxDB.Close, nil
	}

	return nil, nil, err
}

type db struct {
	sqlDB *sqlx.DB
}

func (database *db) GetConnection(ctx context.Context) (Connection, error) {
	conn, err := database.sqlDB.Connx(ctx)
	if err != nil {
		return nil, err
	}
	return &connection{sqlConn: conn}, nil
}

type connection struct {
	sqlConn *sqlx.Conn
}

func (c *connection) GetAdvisoryLocker() AdvisoryLocker {
	return &advisoryLocker{con: c}
}

func (c *connection) Close() error {
	return c.sqlConn.Close()
}

type advisoryLocker struct {
	con *connection
}

// Lock puts a named lock or waits until the resource becomes available
func (al *advisoryLocker) Lock(ctx context.Context, key string) error {
	_, err := al.con.sqlConn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", stringToAdvisoryKey(key))
	return err
}

// Unlock releases a previously-acquired exclusive session level advisory lock
func (al *advisoryLocker) Unlock(ctx context.Context, key string) error {
	_, err := al.con.sqlConn.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", stringToAdvisoryKey(key))
	return err
}

// TryLock attempts to acquire an advisory lock with the given key(name) and returns true if the lock was acquired, false otherwise
func (al *advisoryLocker) TryLock(ctx context.Context, key string) (bool, error) {
	var lockAcquired bool
	if err := al.con.sqlConn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", stringToAdvisoryKey(key)).Scan(&lockAcquired); err != nil {
		return false, err
	}
	return lockAcquired, nil
}

func stringToAdvisoryKey(key string) int64 {
	shaSum := sha256.Sum256([]byte(key))
	return new(big.Int).SetBytes(shaSum[:]).Int64()
}
