package persistence

import "github.com/kyma-incubator/compass/components/provisioner/internal/model"

type OperationService interface {
	Get(operationID string) (model.Operation, error)
	SetAsFailed(operationID string, message string) error
	SetAsSucceeded(operationID string) error
}

type operationService struct {
	repository Repository
}

func NewOperationService(repository Repository) OperationService {
	return operationService{
		repository: repository,
	}
}

func (os operationService) Get(operationID string) (model.Operation, error) {
	return model.Operation{}, nil
}

func (os operationService) SetAsFailed(operationID string, message string) error {
	return nil
}

func (os operationService) SetAsSucceeded(operationID string) error {
	return nil
}
