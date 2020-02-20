package postsql

import (
	"fmt"
	"time"

	"github.com/gocraft/dbr"

	"github.com/pkg/errors"

	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

const (
	schemaName         = "public"
	InstancesTableName = "instances"
	OperationTableName = "operations"
	connectionRetries  = 10
)

// InitializeDatabase opens database connection and initializes schema if it does not exist
func InitializeDatabase(connectionURL string) (*dbr.Connection, error) {
	connection, err := WaitForDatabaseAccess(connectionURL, connectionRetries)
	if err != nil {
		return nil, err
	}

	initialized, err := checkIfDatabaseInitialized(connection)
	if err != nil {
		closeDBConnection(connection)
		return nil, errors.Wrap(err, "Failed to check if database is initialized")
	}
	if initialized {
		log.Info("Database already initialized")
		return connection, nil
	}

	return connection, nil
}

func closeDBConnection(db *dbr.Connection) {
	err := db.Close()
	if err != nil {
		log.Warnf("Failed to close database connection: %s", err.Error())
	}
}

const TableNotExistsError = "42P01"

func checkIfDatabaseInitialized(db *dbr.Connection) (bool, error) {
	checkQuery := fmt.Sprintf(`SELECT '%s.%s'::regclass;`, schemaName, InstancesTableName)

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

	return tableName == InstancesTableName, nil
}

func WaitForDatabaseAccess(connString string, retryCount int) (*dbr.Connection, error) {
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
			log.Info("Failed to close database ...")
		}

		log.Info("Failed to access database, waiting 5 seconds to retry...")
		time.Sleep(5 * time.Second)
	}

	return nil, errors.New("timeout waiting for database access")
}
