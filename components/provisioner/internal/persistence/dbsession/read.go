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
		Select("id", "kubeconfig", "terraform_state", "creation_timestamp").
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
		ID           string
		KymaConfigID string
		Version      string
		Module       string
		ClusterID    string
	}

	rowsCount, err := r.session.
		Select("kyma_config_module.id", "kyma_config_id", "kyma_config.version", "kyma_config_module.module", "cluster_id").
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
			ID:     configModule.ID,
			Module: model.KymaModule(configModule.Module),
		}
		kymaModules = append(kymaModules, kymaConfigModule)
	}

	return model.KymaConfig{
		ID:      kymaConfig[0].KymaConfigID,
		Version: kymaConfig[0].Version,
		Modules: kymaModules,
	}, nil
}

func (r readSession) GetClusterConfig(runtimeID string) (interface{}, dberrors.Error) {
	var gardenerConfig model.GardenerConfig

	err := r.session.
		Select("gardener_config.id", "cluster_id", "gardener_config.name", "project_name", "kubernetes_version",
			"node_count", "volume_size", "disk_type", "machine_type", "target_provider",
			"target_secret", "cidr", "region", "zone", "auto_scaler_min", "auto_scaler_max",
			"max_surge", "max_unavailable").
		From("cluster").
		Join("gardener_config", "cluster.id=gardener_config.cluster_id").
		Where(dbr.Eq("cluster.id", runtimeID)).
		LoadOne(&gardenerConfig)

	if err == nil {
		return gardenerConfig, nil
	}

	if err != dbr.ErrNotFound {
		return model.GardenerConfig{}, dberrors.Internal("Failed to get Gardener Config: %s", err)
	}

	var gcpConfig model.GCPConfig

	err = r.session.
		Select("gcp_config.id", "cluster_id", "name", "project_name", "kubernetes_version",
			"number_of_nodes", "boot_disk_size", "machine_type", "region", "zone").
		From("cluster").
		Join("gcp_config", "cluster.id=gcp_config.cluster_id").
		Where(dbr.Eq("cluster.id", runtimeID)).
		LoadOne(&gcpConfig)

	if err != nil {
		if err == dbr.ErrNotFound {
			return model.GCPConfig{}, dberrors.NotFound("Cluster configuration not found for runtime: %s", runtimeID)
		}
		return model.GCPConfig{}, dberrors.Internal("Failed to get GCP Config: %s", err)
	}

	return gcpConfig, nil
}

func (r readSession) GetOperation(operationID string) (model.Operation, dberrors.Error) {
	var operation model.Operation

	err := r.session.
		Select("id", "type", "start_timestamp", "end_timestamp", "state", "message", "cluster_id").
		From("operation").
		Where(dbr.Eq("id", operationID)).
		LoadOne(&operation)

	if err != nil {
		if err == dbr.ErrNotFound {
			return model.Operation{}, dberrors.NotFound("Operation not found for id: %s", operationID)
		}
		return model.Operation{}, dberrors.Internal("Failed to get %s operation: %s", operationID, err)
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
		Select("id", "type", "start_timestamp", "end_timestamp", "state", "message", "cluster_id").
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
