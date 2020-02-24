package dbsession

import (
	dbr "github.com/gocraft/dbr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/postsql"
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
	var operation OperationDTO

	err := r.session.
		Select("*").
		From(postsql.OperationTableName).
		Where(dbr.Eq("id", opID)).
		LoadOne(&operation)

	if err != nil {
		if err == dbr.ErrNotFound {
			return OperationDTO{}, dberr.NotFound("Cannot find operation for ID: '%s'", opID)
		}
		return OperationDTO{}, dberr.Internal("Failed to get operation: %s", err)
	}
	return operation, nil
}
