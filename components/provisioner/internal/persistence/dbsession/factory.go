package dbsession

import (
	"github.com/gocraft/dbr"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"time"
)

type DBSessionFactory interface {
	NewReadSession() ReadSession
	NewWriteSession() WriteSession
	NewWriteSessionInTransaction() (WriteSessionInTransaction, dberrors.Error)
}

type ReadSession interface {
	GetRuntimeStatus(runtimeID string) (model.RuntimeStatus, dberrors.Error)
	GetLastOperation(runtimeID string) (model.Operation, dberrors.Error)
	GetKymaConfig(runtimeID string) (model.KymaConfig, dberrors.Error)
	GetClusterConfig(runtimeID string) (interface{}, dberrors.Error)
}

type WriteSession interface {
	InsertCluster(runtimeID string, creationTimestamp time.Time, terraformState string) dberrors.Error
	InsertGardenerConfig(runtimeID string, config model.GardenerConfig) dberrors.Error
	InsertGCPConfig(runtimeID string, config model.GCPConfig) dberrors.Error
	InsertKymaConfig(runtimeID string, version string) (string, dberrors.Error)
	InsertKymaConfigModule(kymaConfigID string, module model.KymaModule) dberrors.Error
	InsertOperation(operation model.Operation) dberrors.Error
	UpdateOperationState(operationID string, message string, state model.OperationState) dberrors.Error
	UpdateCluster(runtimeID string, kubeconfig string, terraformState string) dberrors.Error
	DeleteCluster(runtimeID string) dberrors.Error
}

type Transaction interface {
	Commit() dberrors.Error
	Rollback() dberrors.Error
}

type WriteSessionInTransaction interface {
	WriteSession
	Transaction
}

type dbSessionFactory struct {
	connection *dbr.Connection
}

func NewDBSessionFactory(connection *dbr.Connection) DBSessionFactory {
	return &dbSessionFactory{
		connection: connection,
	}
}

func (sf *dbSessionFactory) NewReadSession() ReadSession {
	return dbReadSession{
		dbSession: sf.connection.NewSession(nil),
	}
}

func (sf *dbSessionFactory) NewWriteSession() WriteSession {
	return dbWriteSession{
		dbSession: sf.connection.NewSession(nil),
	}
}

func (sf *dbSessionFactory) NewWriteSessionInTransaction() (WriteSessionInTransaction, dberrors.Error) {
	dbSession := sf.connection.NewSession(nil)
	dbTransaction, err := dbSession.Begin()

	if err != nil {
		return nil, dberrors.Internal("Failed to start transaction: %s", err)
	}

	return dbWriteSession{
		dbSession:     dbSession,
		dbTransaction: dbTransaction,
	}, nil
}
