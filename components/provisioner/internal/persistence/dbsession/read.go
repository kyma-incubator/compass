package dbsession

import (
	"github.com/gocraft/dbr"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
)

type dbReadSession struct {
	dbSession *dbr.Session
}

func (r dbReadSession) GetRuntimeStatus(runtimeID string) (model.RuntimeStatus, dberrors.Error) {
	operation, err := r.GetLastOperation(runtimeID)
	if err != nil {
		return model.RuntimeStatus{}, err
	}

	clusterConfig, err := r.GetClusterConfig(runtimeID)
	if err != nil {
		return model.RuntimeStatus{}, err
	}

	kymaConfig, err := r.GetKymaConfig(runtimeID)
	if err != nil {
		return model.RuntimeStatus{}, err
	}

	runtimeConfiguration := model.RuntimeConfig{
		KymaConfig:    kymaConfig,
		ClusterConfig: clusterConfig,
	}

	return model.RuntimeStatus{
		LastOperationStatus:  operation,
		RuntimeConfiguration: runtimeConfiguration,
	}, nil
}

func (r dbReadSession) GetKymaConfig(runtimeID string) (model.KymaConfig, dberrors.Error) {
	var kymaConfig []struct {
		Version string
		Module  string
	}

	rowsCount, err := r.dbSession.
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

	kymaModules := make([]model.KymaModule, 0)

	for _, configModule := range kymaConfig {
		kymaModules = append(kymaModules, model.KymaModule(configModule.Module))
	}

	return model.KymaConfig{
		Version: kymaConfig[0].Version,
		Modules: kymaModules,
	}, nil
}

func (r dbReadSession) GetClusterConfig(runtimeID string) (interface{}, dberrors.Error) {
	var gardenerConfig model.GardenerConfig

	rowsCount, err := r.dbSession.
		Select("*").
		From("cluster").
		LeftJoin("gardener_config", "cluster.id=gardener_config.cluster_id").
		Where(dbr.Eq("cluster.id", runtimeID)).
		Load(&gardenerConfig)

	if err != nil {
		return model.KymaConfig{}, dberrors.Internal("Failed to get Gardener Config: %s", err)
	}

	if rowsCount == 1 {
		return gardenerConfig, nil
	}

	var gcpConfig model.GardenerConfig

	err = r.dbSession.
		Select("*").
		From("cluster").
		LeftJoin("gcp_config", "cluster.id=gcpConfig.cluster_id").
		Where(dbr.Eq("cluster.id", runtimeID)).
		LoadOne(&gcpConfig)

	if err != nil {
		return model.KymaConfig{}, dberrors.Internal("Failed to get Gardener Config: %s", err)
	}

	if rowsCount == 1 {
		return gardenerConfig, nil
	}

	return model.GCPConfig{}, nil
}

func (r dbReadSession) GetLastOperation(runtimeID string) (model.Operation, dberrors.Error) {

	lastOperationDateSelect := r.dbSession.
		Select("MAX(start_timestamp)").
		From("operation").
		Where(dbr.Eq("cluster_id", runtimeID))

	var operation model.Operation

	err := r.dbSession.
		Select("*").
		From("operation").
		Where(dbr.Eq("start_timestamp", lastOperationDateSelect)).
		LoadOne(&operation)

	if err != nil {
		return model.Operation{}, dberrors.Internal("Failed to get last operation: %s", err)
	}

	return operation, nil
}
