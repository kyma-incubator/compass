package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/pkg/errors"

	// Importing the database driver (postgresql)
	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
)

// RetryCount is a number of retries when trying to open the database
const RetryCount int = 50

func SaveToContext(ctx context.Context, persistOp PersistenceOp) context.Context {
	return context.WithValue(ctx, PersistenceCtxKey, persistOp)
}

// FromCtx extracts DatabaseOp interface from context
func FromCtx(ctx context.Context) (PersistenceOp, error) {
	dbCtx := ctx.Value(PersistenceCtxKey)

	if db, ok := dbCtx.(PersistenceOp); ok {
		return db, nil
	}

	return nil, apperrors.NewInternalError("unable to fetch database from context")
}

//go:generate mockery --name=Transactioner --output=automock --outpkg=automock --case=underscore
type Transactioner interface {
	Begin() (PersistenceTx, error)
	RollbackUnlessCommitted(ctx context.Context, tx PersistenceTx)
	PingContext(ctx context.Context) error
	Stats() sql.DBStats
}

type db struct {
	sqlDB *sqlx.DB
}

func (db *db) PingContext(ctx context.Context) error {
	return db.sqlDB.PingContext(ctx)
}

func (db *db) Stats() sql.DBStats {
	return db.sqlDB.Stats()
}

func (db *db) Begin() (PersistenceTx, error) {
	tx, err := db.sqlDB.Beginx()
	customTx := &Transaction{
		Tx:        tx,
		committed: false,
	}
	return PersistenceTx(customTx), err
}

func (db *db) RollbackUnlessCommitted(ctx context.Context, tx PersistenceTx) {
	customTx, ok := tx.(*Transaction)
	if !ok {
		log.C(ctx).Warn("State aware transaction is not in use")
		db.rollback(ctx, tx)
	}
	if customTx.committed {
		return
	}
	db.rollback(ctx, customTx)
}

func (db *db) rollback(ctx context.Context, tx PersistenceTx) {
	err := tx.Rollback()
	if err == nil {
		log.C(ctx).Warn("Transaction rolled back")
	} else if err != sql.ErrTxDone {
		log.C(ctx).Warn(err)
	}
}

type Transaction struct {
	*sqlx.Tx
	committed bool
}

func (db *Transaction) Commit() error {
	if db.committed {
		return apperrors.NewInternalError("transaction already committed")
	}
	err := db.Tx.Commit()
	if err != nil {
		return errors.Wrap(err, "while committing transaction")
	}
	db.committed = true
	return nil
}

//go:generate mockery --name=PersistenceTx --output=automock --outpkg=automock --case=underscore
type PersistenceTx interface {
	Commit() error
	Rollback() error
	PersistenceOp
}

//go:generate mockery --name=PersistenceOp --output=automock --outpkg=automock --case=underscore
type PersistenceOp interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error

	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// Configure returns the instance of the database
func Configure(context context.Context, conf DatabaseConfig) (Transactioner, func() error, error) {
	db, closeFunc, err := waitForPersistance(context, conf, RetryCount)

	return db, closeFunc, err
}

func waitForPersistance(ctx context.Context, conf DatabaseConfig, retryCount int) (Transactioner, func() error, error) {
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
