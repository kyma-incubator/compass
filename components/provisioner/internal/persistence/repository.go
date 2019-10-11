package persistence

import (
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"

	"github.com/gocraft/dbr"
	"github.com/gofrs/uuid"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
)

type Repository interface {
	Transaction() Transaction
	InsertCluster(runtimeID string, creationTimestamp time.Time, terraformState string) dberrors.Error
	InsertGardenerConfig(runtimeID string, config model.GardenerConfig) dberrors.Error
	InsertGCPConfig(runtimeID string, config model.GCPConfig) dberrors.Error
	InsertKymaConfig(runtimeID string, version string) (string, dberrors.Error)
	InsertKymaConfigModule(kymaConfigID string, module model.KymaModule) dberrors.Error
	InsertOperation(operation model.Operation) dberrors.Error
	DeleteCluster(runtimeID string) dberrors.Error
	GetRuntimeStatus(runtimeID string) (model.RuntimeStatus, dberrors.Error)
	GetLastOperation(runtimeID string) (model.Operation, dberrors.Error)
	GetKymaConfig(runtimeID string) (model.KymaConfig, dberrors.Error)
	GetClusterConfig(runtimeID string) (interface{}, dberrors.Error)
}

type repository struct {
	dbSession     *dbr.Session
	dbTransaction *dbr.Tx
}

type Transaction interface {
	Commit() dberrors.Error
	Rollback() dberrors.Error
}

type transaction struct {
	dbTransaction *dbr.Tx
}

func (t transaction) Commit() dberrors.Error {
	err := t.dbTransaction.Commit()

	if err != nil {
		return dberrors.Internal("Failed to commit transaction: %s", err)
	}

	return nil
}

func (t transaction) Rollback() dberrors.Error {
	err := t.dbTransaction.Rollback()

	if err != nil {
		return dberrors.Internal("Failed to rollback transaction: %s", err)
	}

	return nil
}

type RepositoryFactory interface {
	New() (Repository, dberrors.Error)
}

type repositoryFactory struct {
	dbConnection *dbr.Connection
}

func NewRepositoryFactory(dbConnection *dbr.Connection) RepositoryFactory {
	return repositoryFactory{
		dbConnection: dbConnection,
	}
}

func (rf repositoryFactory) New() (Repository, dberrors.Error) {
	dbSession := rf.dbConnection.NewSession(nil)
	dbTransaction, err := dbSession.Begin()

	if err != nil {
		return nil, dberrors.Internal("Failed to start transaction: %s", err)
	}

	return repository{
		dbSession:     dbSession,
		dbTransaction: dbTransaction,
	}, nil
}

func (r repository) InsertCluster(runtimeID string, creationTimestamp time.Time, terraformState string) dberrors.Error {
	_, err := r.dbTransaction.InsertInto("cluster").
		Pair("id", runtimeID).
		Pair("creation_timestamp", creationTimestamp).
		Pair("terraform_state", terraformState).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to insert record to Cluster table: %s", err)
	}

	return nil
}

func (r repository) InsertGardenerConfig(runtimeID string, config model.GardenerConfig) dberrors.Error {
	id, err := uuid.NewV4()
	if err != nil {
		return dberrors.Internal("Failed to generate uuid: %s.", err)
	}

	_, err = r.dbTransaction.InsertInto("gardener_config").
		Pair("id", id.String()).
		Pair("cluster_id", runtimeID).
		Pair("project_name", config.ProjectName).
		Pair("name", config.Name).
		Pair("kubernetes_version", config.KubernetesVersion).
		Pair("node_count", config.NodeCount).
		Pair("volume_size", config.VolumeSize).
		Pair("machine_type", config.MachineType).
		Pair("region", config.Region).
		Pair("zone", config.Zone).
		Pair("target_provider", config.TargetProvider).
		Pair("target_secret", config.TargetSecret).
		Pair("disk_type", config.DiskType).
		Pair("cidr", config.Cidr).
		Pair("auto_scaler_min", config.AutoScalerMin).
		Pair("auto_scaler_max", config.AutoScalerMax).
		Pair("max_surge", config.MaxSurge).
		Pair("max_unavailable", config.MaxUnavailable).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to insert record to GardenerConfig table: %s", err)
	}

	return nil
}

func (r repository) InsertGCPConfig(runtimeID string, config model.GCPConfig) dberrors.Error {
	id, err := uuid.NewV4()
	if err != nil {
		return dberrors.Internal("Failed to generate uuid: %s.", err)
	}

	_, err = r.dbTransaction.InsertInto("gcp_config").
		Pair("id", id.String()).
		Pair("cluster_id", runtimeID).
		Pair("project_name", config.ProjectName).
		Pair("kubernetes_version", config.KubernetesVersion).
		Pair("number_of_nodes", config.NumberOfNodes).
		Pair("boot_disk_size", config.BootDiskSize).
		Pair("machine_type", config.MachineType).
		Pair("zone", config.Zone).
		Pair("region", config.Region).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to insert record to GCPConfig table: %s", err)
	}

	return nil
}

func (r repository) InsertKymaConfig(runtimeID string, version string) (string, dberrors.Error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", dberrors.Internal("Failed to generate uuid: %s.", err)
	}

	_, err = r.dbTransaction.InsertInto("kyma_config").
		Pair("id", id.String()).
		Pair("version", version).
		Pair("cluster_id", runtimeID).
		Exec()

	if err != nil {
		return "", dberrors.Internal("Failed to insert record to KymaConfig table: %s", err)
	}

	return id.String(), nil
}

func (r repository) InsertKymaConfigModule(kymaConfigID string, module model.KymaModule) dberrors.Error {
	id, err := uuid.NewV4()
	if err != nil {
		return dberrors.Internal("Failed to generate uuid: %s", err)
	}

	_, err = r.dbTransaction.InsertInto("kyma_config_module").
		Pair("id", id.String()).
		Pair("module", module).
		Pair("kyma_config_id", kymaConfigID).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to insert record to KymaConfigModule table: %s", err)
	}

	return nil
}

func (r repository) InsertOperation(operation model.Operation) dberrors.Error {
	id, err := uuid.NewV4()
	if err != nil {
		return dberrors.Internal("Failed to generate uuid: %s.", err)
	}

	_, err = r.dbTransaction.InsertInto("operation").
		Pair("id", id.String()).
		Pair("type", operation.Operation).
		Pair("state", operation.State).
		Pair("message", operation.Message).
		Pair("start_timestamp", operation.Started).
		Pair("cluster_id", operation.RuntimeID).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to insert record to Operation table: %s", err)
	}

	return nil
}

func (r repository) DeleteCluster(runtimeID string) dberrors.Error {
	_, err := r.dbTransaction.DeleteFrom("cluster").
		Where(dbr.Eq("cluster_id", runtimeID)).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to delete record in Cluster table: %s", err)
	}

	return nil
}

func (r repository) GetRuntimeStatus(runtimeID string) (model.RuntimeStatus, dberrors.Error) {
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

func (r repository) GetKymaConfig(runtimeID string) (model.KymaConfig, dberrors.Error) {
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

func (r repository) GetClusterConfig(runtimeID string) (interface{}, dberrors.Error) {
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

func (r repository) GetLastOperation(runtimeID string) (model.Operation, dberrors.Error) {

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

func (r repository) Transaction() Transaction {
	return &transaction{r.dbTransaction}
}
