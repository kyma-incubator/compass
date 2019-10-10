package persistence

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"time"

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
	_, err := r.dbTransaction.InsertInto("Cluster").
		Pair("id", runtimeID).
		Pair("creationTimestamp", creationTimestamp).
		Pair("terraformState", terraformState).
		Exec()

	if err != nil {
		dberrors.Internal("Failed to insert record to Cluster table: %s", err)
	}

	return nil
}

func (r repository) InsertGardenerConfig(runtimeID string, config model.GardenerConfig) dberrors.Error {
	id, err := uuid.NewV4()
	if err != nil {
		return dberrors.Internal("Failed to generate uuid: %s.", err)
	}

	_, err = r.dbTransaction.InsertInto("GardenerConfig").
		Pair("id", id.String()).
		Pair("clusterId", runtimeID).
		Pair("name", config.Name).
		Pair("kubernetesVersion", config.KubernetesVersion).
		Pair("nodeCount", config.NodeCount).
		Pair("volumeSize", config.VolumeSize).
		Pair("machineType", config.MachineType).
		Pair("region", config.Region).
		Pair("zone", config.Zone).
		Pair("targetProvider", config.TargetProvider).
		Pair("targetSecret", config.TargetSecret).
		Pair("diskType", config.DiskType).
		Pair("cidr", config.Cidr).
		Pair("autoScalerMin", config.AutoScalerMin).
		Pair("autoScalerMax", config.AutoScalerMax).
		Pair("maxSurge", config.MaxSurge).
		Pair("maxUnavailable", config.MaxUnavailable).
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

	_, err = r.dbTransaction.InsertInto("GCPConfig").
		Pair("id", id.String()).
		Pair("clusterId", runtimeID).
		Pair("name", config.Name).
		Pair("kubernetesVersion", config.KubernetesVersion).
		Pair("numberOfNodes", config.NumberOfNodes).
		Pair("bootDiskSize", config.BootDiskSize).
		Pair("machineType", config.MachineType).
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

	_, err = r.dbTransaction.InsertInto("KymaConfig").
		Pair("id", id.String()).
		Pair("version", version).
		Pair("clusterId", runtimeID).
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

	_, err = r.dbTransaction.InsertInto("KymaConfigModule").
		Pair("id", id.String()).
		Pair("module", module).
		Pair("kymaConfigId", kymaConfigID).
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

	_, err = r.dbTransaction.InsertInto("Operation").
		Pair("id", id.String()).
		Pair("type", operation.Operation).
		Pair("state", operation.State).
		Pair("message", operation.Message).
		Pair("startTimestamp", operation.Started).
		Pair("clusterId", operation.RuntimeID).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to insert record to Operation table: %s", err)
	}

	return nil
}

func (r repository) DeleteCluster(runtimeID string) dberrors.Error {
	_, err := r.dbTransaction.DeleteFrom("Cluster").
		Where("clusterId", runtimeID).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to delete record in Cluster table: %s", err)
	}

	return nil
}

func (r repository) GetRuntimeStatus(runtimeID string) (model.RuntimeStatus, dberrors.Error) {

	//res, err := r.dbSession.
	//	Select("*").
	//	From("Cluster").
	//	Where("clusterId", runtimeID).
	//	Join("KymaConfig", "Cluster.Id=KymaConfig.clusterId").
	//	RightJoin("GCPConfig", "Cluster.Id=GCPConfig.clusterId").
	//	RightJoin("GardenerConfig", "Cluster.Id=GardenerConfig.clusterId").LoadOne()
	//

	return model.RuntimeStatus{}, nil
}

func (r repository) GetLastOperation(runtimeID string) (model.Operation, dberrors.Error) {
	return model.Operation{}, nil
}

func (r repository) Transaction() Transaction {
	return &transaction{r.dbTransaction}
}
