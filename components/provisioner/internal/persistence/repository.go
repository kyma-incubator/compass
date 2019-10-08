package persistence

import (
	"time"

	"github.com/gocraft/dbr"
	"github.com/gofrs/uuid"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
)

type Repository interface {
	BeginTransaction() (*dbr.Tx, error)
	InsertCluster(runtimeID string, creationTimestamp time.Time, terraformState string) error
	InsertGardenerConfig(config model.GardenerConfig) error
	InsertGCPConfig(config model.GCPConfig) error
	InsertKymaConfig(runtimeID string, config model.KymaConfig) (string, error)
	InsertKymaConfigModule(kymaConfigID string, kymaModule model.KymaModule) error
	InsertOperation(operation model.Operation) error
	DeleteCluster(runtimeID string) error
	GetRuntimeStatus(runtimeID string) (model.RuntimeStatus, error)
	GetLastOperation(runtimeID string) (model.Operation, error)
}

const (
	GardenerConfigTable = "GardenerConfig"
)

type repository struct {
	dbSession *dbr.Session
}

func NewRepository(dbSession *dbr.Session) Repository {
	return repository{
		dbSession: dbSession,
	}
}

func (r repository) InsertCluster(runtimeID string, creationTimestamp time.Time, terraformState string) error {
	_, err := r.dbSession.InsertInto("Cluster").
		Pair("id", runtimeID).
		Pair("creation_timestamp", creationTimestamp).
		Pair("terraform_state", terraformState).
		Exec()

	return err
}

func (r repository) InsertGardenerConfig(config model.GardenerConfig) error {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}

	_, err = r.dbSession.InsertInto(GardenerConfigTable).
		Pair("id", id.String()).
		Pair("name", config.Name).
		Pair("kubernetesVersion", config.KubernetesVersion).
		Pair("nodeCount", config.NodeCount).
		Pair("volumeSize", config.VolumeSize).
		Pair("machineType", config.MachineType).
		Pair("region", config.Region).
		Pair("zone", config.Zone).
		Pair("targetProvider", config.TargetProvider).
		Pair("targetSecret", config.TargetSecret).
		Pair("cidr", config.Cidr).
		Pair("autoScalerMin", config.AutoScalerMin).
		Pair("autoScalerMax", config.AutoScalerMax).
		Pair("maxSurge", config.MaxSurge).
		Pair("maxUnavailable", config.MaxUnavailable).
		Exec()

	return err
}

func (r repository) InsertGCPConfig(config model.GCPConfig) error {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}

	_, err = r.dbSession.InsertInto("GCPConfig").
		Pair("id", id.String()).
		Pair("name", config.Name).
		Pair("kubernetesVersion", config.KubernetesVersion).
		Pair("numberOfNodes", config.NumberOfNodes).
		Pair("bootDiskSize", config.BootDiskSize).
		Pair("machineType", config.MachineType).
		Pair("region", config.Region).
		Exec()

	return err
}

func (r repository) InsertKymaConfig(runtimeID string, config model.KymaConfig) (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	_, err = r.dbSession.InsertInto("KymaConfig").
		Pair("id", id.String()).
		Pair("version", config.Version).
		Pair("cluster_id", runtimeID).
		Exec()

	return id.String(), err
}

func (r repository) InsertKymaConfigModule(kymaConfigID string, kymaModule model.KymaModule) error {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}

	_, err = r.dbSession.InsertInto("KymaConfigModule").
		Pair("id", id.String()).
		Pair("module", kymaModule).
		Pair("kyma_config_id", kymaConfigID).
		Exec()

	return err
}

func (r repository) InsertOperation(operation model.Operation) error {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}

	_, err = r.dbSession.InsertInto("KymaConfigModule").
		Pair("id", id.String()).
		Pair("type", operation.Operation).
		Pair("state", operation.State).
		Pair("message", "").
		Pair("start_timestamp", operation.Started).
		Pair("cluster_id", operation.RuntimeID).
		Exec()

	return nil
}

func (r repository) DeleteCluster(runtimeID string) error {
	return nil
}

func (r repository) GetRuntimeStatus(runtimeID string) (model.RuntimeStatus, error) {
	return model.RuntimeStatus{}, nil
}

func (r repository) BeginTransaction() (*dbr.Tx, error) {
	return r.dbSession.Begin()
}

func (r repository) GetLastOperation(runtimeID string) (model.Operation, error) {
	return model.Operation{}, nil
}
