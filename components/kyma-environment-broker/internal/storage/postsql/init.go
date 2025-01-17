package postsql

import (
	"fmt"
	"time"

	"github.com/gocraft/dbr"

	"github.com/pkg/errors"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const (
	schemaName             = "public"
	InstancesTableName     = "instances"
	OperationTableName     = "operations"
	OrchestrationTableName = "orchestrations"
	RuntimeStateTableName  = "runtime_states"
	CreatedAtField         = "created_at"
)

// InitializeDatabase opens database connection and initializes schema if it does not exist
func InitializeDatabase(connectionURL string, retries int, log logrus.FieldLogger) (*dbr.Connection, error) {
	connection, err := WaitForDatabaseAccess(connectionURL, retries, 100*time.Millisecond, log)
	if err != nil {
		return nil, err
	}

	initialized, err := CheckIfDatabaseInitialized(connection)
	if err != nil {
		closeDBConnection(connection, log)
		return nil, errors.Wrap(err, "Failed to check if database is initialized")
	}
	if initialized {
		log.Info("Database already initialized")
		return connection, nil
	}

	return connection, nil
}

func closeDBConnection(db *dbr.Connection, log logrus.FieldLogger) {
	err := db.Close()
	if err != nil {
		log.Warnf("Failed to close database connection: %s", err.Error())
	}
}

const TableNotExistsError = "42P01"

func CheckIfDatabaseInitialized(db *dbr.Connection) (bool, error) {
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

func WaitForDatabaseAccess(connString string, retryCount int, sleepTime time.Duration, log logrus.FieldLogger) (*dbr.Connection, error) {
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

		log.Infof("Failed to access database, waiting %v to retry...", sleepTime)
		time.Sleep(sleepTime)
	}

	return nil, errors.New("timeout waiting for database access")
}
