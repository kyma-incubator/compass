package database

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/gocraft/dbr"

	"github.com/lib/pq"

	"github.com/pkg/errors"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const (
	retryCount       = 20
	schemaName       = "public"
	clusterTableName = "cluster"
)

// InitializeDatabase opens database connection and initializes schema if it does not exist
// This is temporary solution
func InitializeDatabase(connectionString, schemaFilePath string) (*dbr.Connection, error) {
	connection, err := waitForDatabaseAccess(connectionString, retryCount)
	if err != nil {
		return nil, err
	}

	initialized, err := checkIfDatabaseInitialized(connection)
	if err != nil {
		closeDBConnection(connection)
		return nil, errors.Wrap(err, "Failed to check if database is initialized")
	}

	if initialized {
		logrus.Info("Database already initialized")
		return connection, nil
	}

	logrus.Info("Database not initialized. Setting up schema...")

	content, err := ioutil.ReadFile(schemaFilePath)
	if err != nil {
		closeDBConnection(connection)
		return nil, errors.Wrap(err, "Failed to read schema file")
	}

	_, err = connection.Exec(string(content))
	if err != nil {
		closeDBConnection(connection)
		return nil, errors.Wrap(err, "Failed to setup database schema")
	}

	logrus.Info("Database initialized successfully")

	return connection, nil
}

func closeDBConnection(db *dbr.Connection) {
	err := db.Close()
	if err != nil {
		logrus.Warnf("Failed to close database connection: %s", err.Error())
	}
}

const TableNotExistsError = "42P01"

func checkIfDatabaseInitialized(db *dbr.Connection) (bool, error) {
	checkQuery := fmt.Sprintf(`SELECT '%s.%s'::regclass;`, schemaName, clusterTableName)

	row := db.QueryRow(checkQuery)

	var tableName string
	err := row.Scan(&tableName)

	if err != nil {
		psqlErr, converted := err.(*pq.Error)

		if converted && psqlErr.Code == TableNotExistsError {
			return false, nil
		}

		return false, errors.Wrap(err, "Failed to check if schema initialized")
	}

	return tableName == clusterTableName, nil
}

func waitForDatabaseAccess(connString string, retryCount int) (*dbr.Connection, error) {
	var connection *dbr.Connection
	var err error
	for ; retryCount > 0; retryCount-- {
		connection, err = dbr.Open("postgres", connString, nil)
		if err != nil {
			return nil, errors.Wrap(err, "Invalid connection string")
		}

		err = connection.Ping()
		if err == nil {
			return connection, nil
		}

		err = connection.Close()
		if err != nil {
			logrus.Info("Failed to close database ...")
		}

		logrus.Info("Failed to access database, waiting 5 seconds to retry...")
		time.Sleep(5 * time.Second)
	}

	return nil, errors.New("timeout waiting for database access")
}
