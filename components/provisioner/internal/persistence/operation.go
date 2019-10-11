package persistence

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dbsession"
)

type OperationService interface {
	Get(operationID string) (model.Operation, error)
	SetAsFailed(operationID string, message string) error
	SetAsSucceeded(operationID string) error
}

type operationService struct {
	dbSessionFactory dbsession.DBSessionFactory
}

func NewOperationService(dbSessionFactory dbsession.DBSessionFactory) OperationService {
	return operationService{
		dbSessionFactory: dbSessionFactory,
	}
}

func (os operationService) Get(operationID string) (model.Operation, error) {
	return model.Operation{}, nil
}

func (os operationService) SetAsFailed(operationID string, message string) error {
	session := os.dbSessionFactory.NewWriteSession()

	return session.UpdateOperationState(operationID, message, model.Failed)
}

func (os operationService) SetAsSucceeded(operationID string) error {
	session := os.dbSessionFactory.NewWriteSession()

	return session.UpdateOperationState(operationID, "Operation succeeded.", model.Succeeded)
}
