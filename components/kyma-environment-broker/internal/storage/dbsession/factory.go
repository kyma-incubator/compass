package dbsession

import (
	dbr "github.com/gocraft/dbr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession/dbmodel"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/predicate"
)

//go:generate mockery -name=Factory
type Factory interface {
	NewReadSession() ReadSession
	NewWriteSession() WriteSession
	NewSessionWithinTransaction() (WriteSessionWithinTransaction, dberr.Error)
}

//go:generate mockery -name=ReadSession
type ReadSession interface {
	FindAllInstancesJoinedWithProvisionOperation(prct ...predicate.Predicate) ([]internal.InstanceWithOperation, dberr.Error)
	GetInstanceByID(instanceID string) (internal.Instance, dberr.Error)
	GetOperationByID(opID string) (dbmodel.OperationDTO, dberr.Error)
	GetOperationsInProgressByType(operationType dbmodel.OperationType) ([]dbmodel.OperationDTO, dberr.Error)
	GetOperationByTypeAndInstanceID(inID string, opType dbmodel.OperationType) (dbmodel.OperationDTO, dberr.Error)
	GetLMSTenant(name, region string) (dbmodel.LMSTenantDTO, dberr.Error)
}

//go:generate mockery -name=WriteSession
type WriteSession interface {
	InsertInstance(instance internal.Instance) dberr.Error
	InsertOperation(dto dbmodel.OperationDTO) dberr.Error
	UpdateInstance(instance internal.Instance) dberr.Error
	UpdateOperation(instance dbmodel.OperationDTO) dberr.Error
	InsertLMSTenant(dto dbmodel.LMSTenantDTO) dberr.Error
	DeleteInstance(instanceID string) dberr.Error
}

type Transaction interface {
	Commit() dberr.Error
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

func (sf *factory) NewSessionWithinTransaction() (WriteSessionWithinTransaction, dberr.Error) {
	dbSession := sf.connection.NewSession(nil)
	dbTransaction, err := dbSession.Begin()

	if err != nil {
		return nil, dberr.Internal("Failed to start transaction: %s", err)
	}

	return writeSession{
		session:     dbSession,
		transaction: dbTransaction,
	}, nil
}
