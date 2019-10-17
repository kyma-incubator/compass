package dbsession

import (
	"github.com/gocraft/dbr"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
)

type readSession struct {
	session *dbr.Session
}

func (r readSession) GetCluster(runtimeID string) (model.Cluster, dberrors.Error) {
	var cluster model.Cluster

	err := r.session.
		Select("*").
		From("cluster").
		Where(dbr.Eq("cluster.id", runtimeID)).
		LoadOne(&cluster)

	if err != nil {
		if err != dbr.ErrNotFound {
			return model.Cluster{}, dberrors.NotFound("Cannot find Cluster for runtimeID:'%s", runtimeID)
		}

		return model.Cluster{}, dberrors.Internal("Failed to get Cluster: %s", err)
	}

	return cluster, nil
}

func (r readSession) GetKymaConfig(runtimeID string) (model.KymaConfig, dberrors.Error) {
	var kymaConfig []struct {
		Version string
		Module  string
	}

	rowsCount, err := r.session.
		Select("*").
		From("cluster").
		Join("kyma_config", "cluster.id=kyma_config.cluster_id").
		Join("kyma_config_module", "kyma_config.id=kyma_config_module.kyma_config_id").
		Where(dbr.Eq("cluster.id", runtimeID)).
		Load(&kymaConfig)

	if err != nil {
		return model.KymaConfig{}, dberrors.Internal("Failed to get Kyma Config: %s", err)
	}

	if rowsCount == 0 {
		return model.KymaConfig{}, dberrors.NotFound("Cannot find Kyma Config for runtimeID:'%s", runtimeID)
	}

	kymaModules := make([]model.KymaConfigModule, 0)

	for _, configModule := range kymaConfig {
		kymaConfigModule := model.KymaConfigModule{
			Module: model.KymaModule(configModule.Module),
		}
		kymaModules = append(kymaModules, kymaConfigModule)
	}

	return model.KymaConfig{
		Version: kymaConfig[0].Version,
		Modules: kymaModules,
	}, nil
}

func (r readSession) GetClusterConfig(runtimeID string) (interface{}, dberrors.Error) {
	var gardenerConfig model.GardenerConfig

	err := r.session.
		Select("*").
		From("cluster").
		Join("gardener_config", "cluster.id=gardener_config.cluster_id").
		Where(dbr.Eq("cluster.id", runtimeID)).
		LoadOne(&gardenerConfig)

	if err == nil {
		return gardenerConfig, nil
	}

	if err != nil && err != dbr.ErrNotFound {
		return model.KymaConfig{}, dberrors.Internal("Failed to get Gardener Config: %s", err)
	}

	var gcpConfig model.GardenerConfig

	err = r.session.
		Select("*").
		From("cluster").
		Join("gcp_config", "cluster.id=gcp_config.cluster_id").
		Where(dbr.Eq("cluster.id", runtimeID)).
		LoadOne(&gcpConfig)

	if err != nil {
		if err == dbr.ErrNotFound {
			return model.Operation{}, dberrors.NotFound("Operation not found for runtime: %s", runtimeID)
		}
		return model.KymaConfig{}, dberrors.Internal("Failed to get Gardener Config: %s", err)
	}

	return gardenerConfig, nil
}

func (r readSession) GetOperation(operationID string) (model.Operation, dberrors.Error) {
	var operation model.Operation

	err := r.session.
		Select("*").
		From("operation").
		Where(dbr.Eq("id", operationID)).
		LoadOne(&operation)

	if err != nil {
		if err == dbr.ErrNotFound {
			return model.Operation{}, dberrors.NotFound("Operation not found for id: %s", operationID)
		}
		return model.Operation{}, dberrors.Internal("Failed to get last operation: %s", err)
	}

	return operation, nil
}

func (r readSession) GetLastOperation(runtimeID string) (model.Operation, dberrors.Error) {
	lastOperationDateSelect := r.session.
		Select("MAX(start_timestamp)").
		From("operation").
		Where(dbr.Eq("cluster_id", runtimeID))

	var operation model.Operation

	err := r.session.
		Select("*").
		From("operation").
		Where(dbr.Eq("start_timestamp", lastOperationDateSelect)).
		LoadOne(&operation)

	if err != nil {
		if err == dbr.ErrNotFound {
			return model.Operation{}, dberrors.NotFound("Last operation not found for runtime: %s", runtimeID)
		}
		return model.Operation{}, dberrors.Internal("Failed to get last operation: %s", err)
	}

	return operation, nil
}
