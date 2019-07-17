package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	// Importing the database driver (postgresql)
	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
)

// RetryCount is a number of retries when trying to open the database
const RetryCount int = 50

// Configure returns the instance of the database
func Configure(connString string) (Transactioner, func() error, error) {
	db, closeFunc, err := waitForPersistance(connString, RetryCount)

	return db, closeFunc, err
}

func SaveToContext(ctx context.Context, persistOp PersistenceOp) context.Context {
	return context.WithValue(ctx, PersistenceCtxKey, persistOp)
}

func waitForPersistance(connString string, retryCount int) (Transactioner, func() error, error) {
	var sqlxDB *sqlx.DB
	var err error
	for ; retryCount > 0; retryCount-- {
		sqlxDB, err = sqlx.Open("postgres", connString)
		if err != nil {
			return nil, nil, err
		}

		err = sqlxDB.Ping()
		if err == nil {
			break
		}

		time.Sleep(5 * time.Second)
	}

	return &db{sqlDB: sqlxDB}, sqlxDB.Close, err
}

// FromCtx extracts DatabaseOp interface from context
func FromCtx(ctx context.Context) (PersistenceOp, error) {
	dbCtx := ctx.Value(PersistenceCtxKey)

	if db, ok := dbCtx.(PersistenceOp); ok {
		return db, nil
	}

	return nil, errors.New("unable to fetch database from context")
}

//go:generate mockery -name=Transactioner -output=automock -outpkg=automock -case=underscore
type Transactioner interface {
	Begin() (PersistenceTx, error)
}

type db struct {
	sqlDB *sqlx.DB
}

func (db *db) Begin() (PersistenceTx, error) {
	tx, err := db.sqlDB.Beginx()
	return PersistenceTx(tx), err
}

func (db *db) Close() error {
	return db.sqlDB.Close()
}

//go:generate mockery -name=PersistenceTx -output=automock -outpkg=automock -case=underscore
type PersistenceTx interface {
	Commit() error
	Rollback() error
}

//go:generate mockery -name=PersistenceOp -output=automock -outpkg=automock -case=underscore
type PersistenceOp interface {
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error

	NamedExec(query string, arg interface{}) (sql.Result, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}
