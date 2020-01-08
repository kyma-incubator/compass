package dbsession

import (
	dbr "github.com/gocraft/dbr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/schema"
)

type readSession struct {
	session *dbr.Session
}

func (r readSession) GetInstanceByID(instanceID string) (internal.Instance, dberr.Error) {
	var instance internal.Instance

	err := r.session.
		Select("*").
		From(schema.InstancesTableName).
		Where(dbr.Eq(schema.InstancesTableName+".instance_id", instanceID)).
		LoadOne(&instance)

	if err != nil {
		if err != dbr.ErrNotFound {
			return internal.Instance{}, dberr.NotFound("Cannot find Instance for instanceID:'%s", instanceID)
		}
		return internal.Instance{}, dberr.Internal("Failed to get Instance: %s", err)
	}
	return instance, nil
}
