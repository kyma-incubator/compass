package dbsession

import (
	dbr "github.com/gocraft/dbr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/postsql"
	"github.com/pivotal-cf/brokerapi/v7/domain"
)

type readSession struct {
	session *dbr.Session
}

func (r readSession) GetInstanceByID(instanceID string) (internal.Instance, dberr.Error) {
	var instance internal.Instance

	err := r.session.
		Select("*").
		From(postsql.InstancesTableName).
		Where(dbr.Eq("instance_id", instanceID)).
		LoadOne(&instance)

	if err != nil {
		if err == dbr.ErrNotFound {
			return internal.Instance{}, dberr.NotFound("Cannot find Instance for instanceID:'%s'", instanceID)
		}
		return internal.Instance{}, dberr.Internal("Failed to get Instance: %s", err)
	}
	return instance, nil
}

func (r readSession) GetOperationByID(opID string) (OperationDTO, dberr.Error) {
	condition := dbr.Eq("id", opID)
	operation, err := r.getOperation(condition)
	if err != nil {
		switch {
		case dberr.IsNotFound(err):
			return OperationDTO{}, dberr.NotFound("for ID: %s %s", opID, err)
		default:
			return OperationDTO{}, err
		}
	}
	return operation, nil
}

func (r readSession) GetOperationsInProgress() ([]OperationDTO, dberr.Error) {
	condition := dbr.Eq("state", domain.InProgress)
	operations, err := r.getOperations(condition)
	if err != nil {
		switch {
		case dberr.IsNotFound(err):
			return nil, dberr.NotFound("not found operation in progress")
		default:
			return nil, err
		}
	}
	return operations, nil
}

func (r readSession) GetOperationByInstanceID(inID string) (OperationDTO, dberr.Error) {
	condition := dbr.Eq("instance_id", inID)
	operation, err := r.getOperation(condition)
	if err != nil {
		switch {
		case dberr.IsNotFound(err):
			return OperationDTO{}, dberr.NotFound("for instanceID: %s %s", inID, err)
		default:
			return OperationDTO{}, err
		}
	}
	return operation, nil
}

func (r readSession) getOperation(condition dbr.Builder) (OperationDTO, dberr.Error) {
	var operation OperationDTO

	err := r.session.
		Select("*").
		From(postsql.OperationTableName).
		Where(condition).
		LoadOne(&operation)

	if err != nil {
		if err == dbr.ErrNotFound {
			return OperationDTO{}, dberr.NotFound("cannot find operation: %s", err)
		}
		return OperationDTO{}, dberr.Internal("Failed to get operation: %s", err)
	}
	return operation, nil
}

func (r readSession) getOperations(condition dbr.Builder) ([]OperationDTO, dberr.Error) {
	var operations []OperationDTO

	_, err := r.session.
		Select("*").
		From(postsql.OperationTableName).
		Where(condition).
		Load(&operations)

	if err != nil {
		if err == dbr.ErrNotFound {
			return nil, dberr.NotFound("cannot find operations: %s", err)
		}
		return nil, dberr.Internal("Failed to get operations: %s", err)
	}
	return operations, nil
}
