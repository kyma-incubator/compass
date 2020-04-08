package dbsession

import (
	"time"

	dbr "github.com/gocraft/dbr/v2"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
)

//go:generate mockery -name=Factory
type Factory interface {
	NewReadSession() ReadSession
	NewWriteSession() WriteSession
	NewReadWriteSession() ReadWriteSession
	NewSessionWithinTransaction() (WriteSessionWithinTransaction, dberrors.Error)
}

//go:generate mockery -name=ReadSession
type ReadSession interface {
	GetCluster(runtimeID string) (model.Cluster, dberrors.Error)
	GetOperation(operationID string) (model.Operation, dberrors.Error)
	GetLastOperation(runtimeID string) (model.Operation, dberrors.Error)
	GetGardenerClusterByName(name string) (model.Cluster, dberrors.Error)
	GetTenant(runtimeID string) (string, dberrors.Error)
	ListInProgressOperations() ([]model.Operation, dberrors.Error)
	GetRuntimeUpgrade(operationId string) (model.RuntimeUpgrade, dberrors.Error)
}

//go:generate mockery -name=WriteSession
type WriteSession interface {
	InsertCluster(cluster model.Cluster) dberrors.Error
	InsertGardenerConfig(config model.GardenerConfig) dberrors.Error
	InsertGCPConfig(config model.GCPConfig) dberrors.Error
	InsertKymaConfig(kymaConfig model.KymaConfig) dberrors.Error
	InsertOperation(operation model.Operation) dberrors.Error
	UpdateOperationState(operationID string, message string, state model.OperationState, endTime time.Time) dberrors.Error
	TransitionOperation(operationID string, message string, stage model.OperationStage, transitionTime time.Time) dberrors.Error
	UpdateCluster(runtimeID string, kubeconfig string, terraformState []byte) dberrors.Error
	SetActiveKymaConfig(runtimeID string, kymaConfigId string) dberrors.Error
	UpdateUpgradeState(operationID string, upgradeState model.UpgradeState) dberrors.Error
	DeleteCluster(runtimeID string) dberrors.Error
	MarkClusterAsDeleted(runtimeID string) dberrors.Error
	InsertRuntimeUpgrade(runtimeUpgrade model.RuntimeUpgrade) dberrors.Error
}

//go:generate mockery -name=ReadWriteSession
type ReadWriteSession interface {
	ReadSession
	WriteSession
}

type Transaction interface {
	Commit() dberrors.Error
	RollbackUnlessCommitted()
}

//go:generate mockery -name=WriteSessionWithinTransaction
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

func (sf *factory) NewReadWriteSession() ReadWriteSession {
	session := sf.connection.NewSession(nil)
	return readWriteSession{
		readSession:  readSession{session: session},
		writeSession: writeSession{session: session},
	}
}

type readWriteSession struct {
	readSession
	writeSession
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
