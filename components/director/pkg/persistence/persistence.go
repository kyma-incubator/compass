package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/sirupsen/logrus"

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

	return nil, errors.New("unable to fetch database from context")
}

//go:generate mockery -name=Transactioner -output=automock -outpkg=automock -case=underscore
type Transactioner interface {
	Begin() (PersistenceTx, error)
	RollbackUnlessCommited(tx PersistenceTx)
	PingContext(ctx context.Context) error
}

type db struct {
	sqlDB  *sqlx.DB
	logger *logrus.Logger
}

func (db *db) PingContext(ctx context.Context) error {
	return db.sqlDB.PingContext(ctx)
}

func (db *db) Begin() (PersistenceTx, error) {
	tx, err := db.sqlDB.Beginx()
	return PersistenceTx(tx), err
}

func (db *db) RollbackUnlessCommited(tx PersistenceTx) {
	err := tx.Rollback()
	if err == nil {
		db.logger.Warn("transaction rolled back")
	} else if err != sql.ErrTxDone {
		db.logger.Warn(err)
	}
}

//go:generate mockery -name=PersistenceTx -output=automock -outpkg=automock -case=underscore
type PersistenceTx interface {
	Commit() error
	Rollback() error
	PersistenceOp
}

//go:generate mockery -name=PersistenceOp -output=automock -outpkg=automock -case=underscore
type PersistenceOp interface {
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error

	NamedExec(query string, arg interface{}) (sql.Result, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// Configure returns the instance of the database
func Configure(logger *logrus.Logger, conf DatabaseConfig) (Transactioner, func() error, error) {
	db, closeFunc, err := waitForPersistance(logger, conf, RetryCount)

	return db, closeFunc, err
}

func waitForPersistance(logger *logrus.Logger, conf DatabaseConfig, retryCount int) (Transactioner, func() error, error) {
	var sqlxDB *sqlx.DB
	var err error
	for i := 0; i < retryCount; i++ {
		if i > 0 {
			time.Sleep(5 * time.Second)
		}
		logger.Info("Trying to connect to DB...")

		sqlxDB, err = sqlx.Open("postgres", conf.GetConnString())
		if err != nil {
			return nil, nil, err
		}
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
		err = sqlxDB.PingContext(ctx)
		cancelFunc()
		if err != nil {
			logger.Infof("Got error on pinging DB: %v", err)
			continue
		}
		logger.Infof("Configuring MaxOpenConnections: [%d] and MaxIdleConnections: [%d]", conf.MaxOpenConnections, conf.MaxIdleConnections)
		sqlxDB.SetMaxOpenConns(conf.MaxOpenConnections)
		sqlxDB.SetMaxIdleConns(conf.MaxIdleConnections)
		return &db{sqlDB: sqlxDB, logger: logger}, sqlxDB.Close, nil

	}

	return nil, nil, err
}
