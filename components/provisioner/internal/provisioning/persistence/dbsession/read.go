package dbsession

import (
	"encoding/json"
	"fmt"
	"sort"

	dbr "github.com/gocraft/dbr/v2"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
)

type readSession struct {
	session *dbr.Session
}

func (r readSession) GetTenant(runtimeID string) (string, dberrors.Error) {
	var tenant string

	err := r.session.
		Select("tenant").
		From("cluster").
		Where(dbr.Eq("cluster.id", runtimeID)).
		LoadOne(&tenant)

	if err != nil {
		if err != dbr.ErrNotFound {
			return "", dberrors.NotFound("Cannot find Tenant for runtimeID:'%s", runtimeID)
		}

		return "", dberrors.Internal("Failed to get Tenant: %s", err)
	}
	return tenant, nil
}

func (r readSession) GetCluster(runtimeID string) (model.Cluster, dberrors.Error) {
	var cluster model.Cluster

	err := r.session.
		Select(
			"id", "kubeconfig", "terraform_state", "tenant",
			"credentials_secret_name", "creation_timestamp", "deleted", "sub_account_id").
		From("cluster").
		Where(dbr.Eq("cluster.id", runtimeID)).
		LoadOne(&cluster)

	if err != nil {
		if err != dbr.ErrNotFound {
			return model.Cluster{}, dberrors.NotFound("Cannot find Cluster for runtimeID: %s", runtimeID)
		}
		return model.Cluster{}, dberrors.Internal("Failed to get Cluster: %s", err)
	}

	providerConfig, err := r.getProviderConfig(runtimeID)
	if err != nil {
		return model.Cluster{}, dberrors.NotFound("Cannot find Provider config for runtimeID: %s", runtimeID)
	}
	cluster.ClusterConfig = providerConfig

	kymaConfig, err := r.getKymaConfig(runtimeID)
	if err != nil {
		return model.Cluster{}, dberrors.NotFound("Cannot find Kyma config for runtimeID: %s", runtimeID)
	}
	cluster.KymaConfig = kymaConfig

	return cluster, nil
}

func (r readSession) GetGardenerClusterByName(name string) (model.Cluster, dberrors.Error) {
	var clusterWithProvider = struct {
		model.Cluster
		gardenerConfigRead
	}{}

	err := r.session.
		Select(
			"cluster.id", "cluster.kubeconfig", "cluster.tenant",
			"cluster.credentials_secret_name", "cluster.creation_timestamp", "cluster.deleted",
			"name", "project_name", "kubernetes_version",
			"volume_size_gb", "disk_type", "machine_type", "provider", "seed",
			"target_secret", "worker_cidr", "region", "auto_scaler_min", "auto_scaler_max",
			"max_surge", "max_unavailable", "provider_specific_config").
		From("gardener_config").
		Join("cluster", "gardener_config.cluster_id=cluster.id").
		Where(dbr.Eq("name", name)).
		LoadOne(&clusterWithProvider)

	if err != nil {
		if err == dbr.ErrNotFound {
			return model.Cluster{}, dberrors.NotFound("Cannot find Gardener Cluster with name: %s", name)
		}

		return model.Cluster{}, dberrors.Internal("Failed to get Gardener Cluster with name: %s, error: %s", name, err)
	}
	cluster := clusterWithProvider.Cluster

	err = clusterWithProvider.gardenerConfigRead.DecodeProviderConfig()
	if err != nil {
		return model.Cluster{}, dberrors.Internal("Failed to decode Gardener provider config fetched from database: %s", err.Error())
	}
	cluster.ClusterConfig = clusterWithProvider.gardenerConfigRead.GardenerConfig

	kymaConfig, err := r.getKymaConfig(clusterWithProvider.Cluster.ID)
	if err != nil {
		return model.Cluster{}, dberrors.NotFound("Cannot find Kyma config for runtimeID: %s", clusterWithProvider.Cluster.ID)
	}
	cluster.KymaConfig = kymaConfig

	return cluster, nil
}

func (r readSession) getKymaConfig(runtimeID string) (model.KymaConfig, dberrors.Error) {
	var kymaConfig []struct {
		ID                  string
		KymaConfigID        string
		GlobalConfiguration []byte
		ReleaseID           string
		Version             string
		TillerYAML          string
		InstallerYAML       string
		Component           string
		Namespace           string
		SourceURL           string
		Configuration       []byte
		ComponentOrder      int
		ClusterID           string
	}

	rowsCount, err := r.session.
		Select("kyma_config_id", "kyma_config.release_id", "kyma_config.global_configuration",
			"kyma_component_config.id", "kyma_component_config.component", "kyma_component_config.namespace",
			"kyma_component_config.source_url", "kyma_component_config.configuration",
			"kyma_component_config.component_order",
			"cluster_id",
			"kyma_release.version", "kyma_release.tiller_yaml", "kyma_release.installer_yaml").
		From("cluster").
		Join("kyma_config", "cluster.id=kyma_config.cluster_id").
		Join("kyma_component_config", "kyma_config.id=kyma_component_config.kyma_config_id").
		Join("kyma_release", "kyma_config.release_id=kyma_release.id").
		Where(dbr.Eq("cluster.id", runtimeID)).
		Load(&kymaConfig)

	if err != nil {
		return model.KymaConfig{}, dberrors.Internal("Failed to get Kyma Config: %s", err)
	}

	if rowsCount == 0 {
		return model.KymaConfig{}, dberrors.NotFound("Cannot find Kyma Config for runtimeID: %s", runtimeID)
	}

	kymaModulesOrdered := make(map[int][]model.KymaComponentConfig, 0)

	for _, componentCfg := range kymaConfig {
		var configuration model.Configuration
		err := json.Unmarshal(componentCfg.Configuration, &configuration)
		if err != nil {
			return model.KymaConfig{}, dberrors.Internal("Failed to unmarshal configuration for %s component: %s", componentCfg.Component, err.Error())
		}

		kymaComponentConfig := model.KymaComponentConfig{
			ID:             componentCfg.ID,
			Component:      model.KymaComponent(componentCfg.Component),
			Namespace:      componentCfg.Namespace,
			SourceURL:      componentCfg.SourceURL,
			Configuration:  configuration,
			KymaConfigID:   componentCfg.KymaConfigID,
			ComponentOrder: componentCfg.ComponentOrder,
		}

		// In case order is 0 for all components map stores slice (it is the case for Runtimes created before migration)
		kymaModulesOrdered[componentCfg.ComponentOrder] = append(kymaModulesOrdered[componentCfg.ComponentOrder], kymaComponentConfig)
	}

	keys := make([]int, 0, len(kymaModulesOrdered))
	for k, _ := range kymaModulesOrdered {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	orderedComponents := make([]model.KymaComponentConfig, 0, len(kymaConfig))
	for _, k := range keys {
		orderedComponents = append(orderedComponents, kymaModulesOrdered[k]...)
	}

	var globalConfiguration model.Configuration
	err = json.Unmarshal(kymaConfig[0].GlobalConfiguration, &globalConfiguration)
	if err != nil {
		return model.KymaConfig{}, dberrors.Internal("Failed to unmarshal global configuration: %s", err.Error())
	}

	return model.KymaConfig{
		ID: kymaConfig[0].KymaConfigID,
		Release: model.Release{
			Id:            kymaConfig[0].ReleaseID,
			Version:       kymaConfig[0].Version,
			TillerYAML:    kymaConfig[0].TillerYAML,
			InstallerYAML: kymaConfig[0].InstallerYAML,
		},
		Components:          orderedComponents,
		GlobalConfiguration: globalConfiguration,
		ClusterID:           runtimeID,
	}, nil
}

type gardenerConfigRead struct {
	model.GardenerConfig
	ProviderSpecificConfig string `db:"provider_specific_config"`
}

func (gcr *gardenerConfigRead) DecodeProviderConfig() error {
	gardenerConfigProviderConfig, err := model.NewGardenerProviderConfigFromJSON(gcr.ProviderSpecificConfig)
	if err != nil {
		return fmt.Errorf("error decoding Gardener provider config: %s", err.Error())
	}

	gcr.GardenerProviderConfig = gardenerConfigProviderConfig
	return nil
}

func (r readSession) getProviderConfig(runtimeID string) (model.ProviderConfiguration, dberrors.Error) {
	gardenerConfig := gardenerConfigRead{}

	err := r.session.
		Select("gardener_config.id", "cluster_id", "gardener_config.name", "project_name", "kubernetes_version",
			"volume_size_gb", "disk_type", "machine_type", "provider", "seed",
			"target_secret", "worker_cidr", "region", "auto_scaler_min", "auto_scaler_max",
			"max_surge", "max_unavailable", "provider_specific_config").
		From("cluster").
		Join("gardener_config", "cluster.id=gardener_config.cluster_id").
		Where(dbr.Eq("cluster.id", runtimeID)).
		LoadOne(&gardenerConfig)

	if err == nil {
		err := gardenerConfig.DecodeProviderConfig()
		if err != nil {
			return nil, dberrors.Internal("Failed to decode Gardener provider config fetched from database: %s", err.Error())
		}

		return gardenerConfig.GardenerConfig, nil
	}

	if err != dbr.ErrNotFound {
		return model.GardenerConfig{}, dberrors.Internal("Failed to get Gardener Config: %s", err)
	}

	var gcpConfig model.GCPConfig

	err = r.session.
		Select("gcp_config.id", "cluster_id", "name", "project_name", "kubernetes_version",
			"number_of_nodes", "boot_disk_size_gb", "machine_type", "region", "zone").
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
