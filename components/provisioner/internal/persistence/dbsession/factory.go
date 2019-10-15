package dbsession

import (
	"time"

	"github.com/gocraft/dbr"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
)

//go:generate mockery -name=Factory
type Factory interface {
	NewReadSession() ReadSession
	NewWriteSession() WriteSession
	NewSessionWithinTransaction() (WriteSessionWithinTransaction, dberrors.Error)
}

//go:generate mockery -name=ReadSession
type ReadSession interface {
	GetOperation(operationID string) (model.Operation, dberrors.Error)
	GetLastOperation(runtimeID string) (model.Operation, dberrors.Error)
	GetKymaConfig(runtimeID string) (model.KymaConfig, dberrors.Error)
	GetClusterConfig(runtimeID string) (interface{}, dberrors.Error)
}

//go:generate mockery -name=WriteSession
type WriteSession interface {
	InsertCluster(runtimeID string, creationTimestamp time.Time, terraformState string) dberrors.Error
	InsertGardenerConfig(runtimeID string, config model.GardenerConfig) dberrors.Error
	InsertGCPConfig(runtimeID string, config model.GCPConfig) dberrors.Error
	InsertKymaConfig(runtimeID string, version string) (string, dberrors.Error)
	InsertKymaConfigModule(kymaConfigID string, module model.KymaModule) dberrors.Error
	InsertOperation(operation model.Operation) (string, dberrors.Error)
	UpdateOperationState(operationID string, message string, state model.OperationState) dberrors.Error
	UpdateCluster(runtimeID string, kubeconfig string, terraformState string) dberrors.Error
	DeleteCluster(runtimeID string) dberrors.Error
}

type Transaction interface {
	Commit() dberrors.Error
	Rollback() dberrors.Error
}

type WriteSessionWithinTransaction interface {
	WriteSession
	Transaction
}

type factory struct {
	connection *dbr.Connection
}

func NewFactory(connection *dbr.Connection) Factory {
	return &factory{
		connection: connection,
	}
}

func (sf *factory) NewReadSession() ReadSession {
	return readSession{
		session: sf.connection.NewSession(nil),
	}
}

func (sf *factory) NewWriteSession() WriteSession {
	return writeSession{
		session: sf.connection.NewSession(nil),
	}
}

func (sf *factory) NewSessionWithinTransaction() (WriteSessionWithinTransaction, dberrors.Error) {
	dbSession := sf.connection.NewSession(nil)
	dbTransaction, err := dbSession.Begin()

	if err != nil {
		return nil, dberrors.Internal("Failed to start transaction: %s", err)
	}

	return writeSession{
		session:     dbSession,
		transaction: dbTransaction,
	}, nil
}
