package dbsession

import (
	"database/sql"
	"fmt"
	"github.com/gocraft/dbr"
	"github.com/gofrs/uuid"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"time"
)

type dbWriteSession struct {
	dbSession     *dbr.Session
	dbTransaction *dbr.Tx
}

func (ws dbWriteSession) InsertCluster(runtimeID string, creationTimestamp time.Time, terraformState string) dberrors.Error {
	_, err := ws.insertInto("cluster").
		Pair("id", runtimeID).
		Pair("creation_timestamp", creationTimestamp).
		Pair("terraform_state", terraformState).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to insert record to Cluster table: %s", err)
	}

	return nil
}

func (ws dbWriteSession) InsertGardenerConfig(runtimeID string, config model.GardenerConfig) dberrors.Error {
	id, err := uuid.NewV4()
	if err != nil {
		return dberrors.Internal("Failed to generate uuid: %s.", err)
	}

	_, err = ws.insertInto("gardener_config").
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

func (ws dbWriteSession) InsertGCPConfig(runtimeID string, config model.GCPConfig) dberrors.Error {
	id, err := uuid.NewV4()
	if err != nil {
		return dberrors.Internal("Failed to generate uuid: %s.", err)
	}

	_, err = ws.insertInto("gcp_config").
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

func (ws dbWriteSession) InsertKymaConfig(runtimeID string, version string) (string, dberrors.Error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", dberrors.Internal("Failed to generate uuid: %s.", err)
	}

	_, err = ws.insertInto("kyma_config").
		Pair("id", id.String()).
		Pair("version", version).
		Pair("cluster_id", runtimeID).
		Exec()

	if err != nil {
		return "", dberrors.Internal("Failed to insert record to KymaConfig table: %s", err)
	}

	return id.String(), nil
}

func (ws dbWriteSession) InsertKymaConfigModule(kymaConfigID string, module model.KymaModule) dberrors.Error {
	id, err := uuid.NewV4()
	if err != nil {
		return dberrors.Internal("Failed to generate uuid: %s", err)
	}

	_, err = ws.insertInto("kyma_config_module").
		Pair("id", id.String()).
		Pair("module", module).
		Pair("kyma_config_id", kymaConfigID).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to insert record to KymaConfigModule table: %s", err)
	}

	return nil
}

func (ws dbWriteSession) InsertOperation(operation model.Operation) (string, dberrors.Error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", dberrors.Internal("Failed to generate uuid: %s.", err)
	}

	_, err = ws.insertInto("operation").
		Pair("id", id.String()).
		Pair("type", operation.Operation).
		Pair("state", operation.State).
		Pair("message", operation.Message).
		Pair("start_timestamp", operation.Started).
		Pair("cluster_id", operation.RuntimeID).
		Exec()

	if err != nil {
		return "", dberrors.Internal("Failed to insert record to Operation table: %s", err)
	}

	return id.String(), nil
}

func (ws dbWriteSession) CleanupData(runtimeID string) dberrors.Error {
	_, err := ws.deleteFrom("cluster").
		Where(dbr.Eq("id", runtimeID)).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to delete record in Cluster table: %s", err)
	}

	return nil
}

func (ws dbWriteSession) UpdateOperationState(operationID string, message string, state model.OperationState) dberrors.Error {
	res, err := ws.update("operation").
		Where(dbr.Eq("id", operationID)).
		Set("state", state).
		Set("message", message).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to update operation %s state: %s", operationID, err)
	}

	return ws.updateSucceeded(res, fmt.Sprintf("Failed to update operation %s state: %s", operationID, err))
}

func (ws dbWriteSession) UpdateCluster(runtimeID string, kubeconfig string, terraformState string) dberrors.Error {
	res, err := ws.update("cluster").
		Where(dbr.Eq("id", runtimeID)).
		Set("kubeconfig", kubeconfig).
		Set("terraform_state", terraformState).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to update cluster %s state: %s", runtimeID, err)
	}

	return ws.updateSucceeded(res, fmt.Sprintf("Failed to update cluster %s data: %s", runtimeID, err))
}

func (ws dbWriteSession) updateSucceeded(result sql.Result, errorMsg string) dberrors.Error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return dberrors.Internal("Failed to get number of rows affected: %s", err)
	}

	if rowsAffected == 0 {
		return dberrors.NotFound(errorMsg)
	}

	return nil
}

func (ws dbWriteSession) Commit() dberrors.Error {
	err := ws.dbTransaction.Commit()
	if err != nil {
		return dberrors.Internal("Failed to commit transaction: %s", err)
	}

	return nil
}

func (ws dbWriteSession) Rollback() dberrors.Error {
	err := ws.dbTransaction.Rollback()
	if err != nil {
		return dberrors.Internal("Failed to rollback transaction: %s", err)
	}

	return nil
}

func (ws dbWriteSession) insertInto(table string) *dbr.InsertStmt {
	if ws.dbTransaction != nil {
		return ws.dbTransaction.InsertInto(table)
	}

	return ws.dbSession.InsertInto(table)
}

func (ws dbWriteSession) deleteFrom(table string) *dbr.DeleteStmt {
	if ws.dbTransaction != nil {
		return ws.dbTransaction.DeleteFrom(table)
	}

	return ws.dbSession.DeleteFrom(table)
}

func (ws dbWriteSession) update(table string) *dbr.UpdateStmt {
	if ws.dbTransaction != nil {
		return ws.dbTransaction.Update(table)
	}

	return ws.dbSession.Update(table)
}
