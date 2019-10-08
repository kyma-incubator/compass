package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/pkg/errors"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const (
	retryCount       = 20
	schemaName       = "public"
	clusterTableName = `"Cluster"`
)

// InitializeDatabase opens database connection and initializes schema if it does not exist
// This is temporary solution
func InitializeDatabase(connectionString, schemaFilePath string) (*sql.DB, error) {
	sqlDatabase, err := waitForDatabaseAccess(connectionString, retryCount)
	if err != nil {
		return nil, err
	}

	initialized, err := checkIfDatabaseInitialized(sqlDatabase)
	if err != nil {
		closeDBConnection(sqlDatabase)
		return nil, errors.Wrap(err, "Failed to check if database is initialized")
	}

	if initialized {
		logrus.Info("Database already initialized")
		return sqlDatabase, nil
	}

	logrus.Info("Database not initialized. Setting up schema...")

	content, err := ioutil.ReadFile(schemaFilePath)
	if err != nil {
		closeDBConnection(sqlDatabase)
		return nil, errors.Wrap(err, "Failed to read schema file")
	}

	_, err = sqlDatabase.Exec(string(content))
	if err != nil {
		closeDBConnection(sqlDatabase)
		return nil, errors.Wrap(err, "Failed to setup database schema")
	}

	logrus.Info("Database initialized successfully")

	return sqlDatabase, nil
}

func closeDBConnection(db *sql.DB) {
	err := db.Close()
	if err != nil {
		logrus.Warnf("Failed to close database connection: %s", err.Error())
	}
}

func checkIfDatabaseInitialized(db *sql.DB) (bool, error) {
	checkQuery := fmt.Sprintf(`SELECT '%s.%s'::regclass;`, schemaName, clusterTableName)

	row := db.QueryRow(checkQuery)

	var tableName string
	err := row.Scan(&tableName)
	if err != nil {
		logrus.Infof("Database not initialized: %s", err.Error())
		return false, nil
	}

	return tableName == clusterTableName, nil
}

func waitForDatabaseAccess(connString string, retryCount int) (*sql.DB, error) {
	var sqlDB *sql.DB
	var err error
	for ; retryCount > 0; retryCount-- {
		sqlDB, err = sql.Open("postgres", connString)
		if err != nil {
			return nil, errors.Wrap(err, "Invalid connection string")
		}

		err = sqlDB.Ping()
		if err == nil {
			return sqlDB, nil
		}

		logrus.Info("Failed to access database, waiting 5 seconds to retry...")
		time.Sleep(5 * time.Second)
	}

	return nil, errors.New("timeout waiting for database access")
}
