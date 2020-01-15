package dbsession

import (
	"database/sql"
	"encoding/json"
	"fmt"

	dbr "github.com/gocraft/dbr/v2"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
)

type writeSession struct {
	session     *dbr.Session
	transaction *dbr.Tx
}

func (ws writeSession) InsertCluster(cluster model.Cluster) dberrors.Error {
	_, err := ws.insertInto("cluster").
		Pair("id", cluster.ID).
		Pair("terraform_state", cluster.TerraformState).
		Pair("credentials_secret_name", cluster.CredentialsSecretName).
		Pair("creation_timestamp", cluster.CreationTimestamp).
		Pair("tenant", cluster.Tenant).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to insert record to Cluster table: %s", err)
	}

	return nil
}

func (ws writeSession) InsertGardenerConfig(config model.GardenerConfig) dberrors.Error {
	_, err := ws.insertInto("gardener_config").
		Pair("id", config.ID).
		Pair("cluster_id", config.ClusterID).
		Pair("project_name", config.ProjectName).
		Pair("name", config.Name).
		Pair("kubernetes_version", config.KubernetesVersion).
		Pair("node_count", config.NodeCount).
		Pair("volume_size_gb", config.VolumeSizeGB).
		Pair("machine_type", config.MachineType).
		Pair("region", config.Region).
		Pair("provider", config.Provider).
		Pair("seed", config.Seed).
		Pair("target_secret", config.TargetSecret).
		Pair("disk_type", config.DiskType).
		Pair("worker_cidr", config.WorkerCidr).
		Pair("auto_scaler_min", config.AutoScalerMin).
		Pair("auto_scaler_max", config.AutoScalerMax).
		Pair("max_surge", config.MaxSurge).
		Pair("max_unavailable", config.MaxUnavailable).
		Pair("provider_specific_config", config.GardenerProviderConfig.RawJSON()).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to insert record to GardenerConfig table: %s", err)
	}

	return nil
}

func (ws writeSession) InsertGCPConfig(config model.GCPConfig) dberrors.Error {
	_, err := ws.insertInto("gcp_config").
		Columns("id", "cluster_id", "name", "project_name", "kubernetes_version", "number_of_nodes", "boot_disk_size_gb",
			"machine_type", "zone", "region").
		Record(config).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to insert record to GCPConfig table: %s", err)
	}

	return nil
}

func (ws writeSession) InsertKymaConfig(kymaConfig model.KymaConfig) dberrors.Error {
	jsonConfig, err := json.Marshal(kymaConfig.GlobalConfiguration)
	if err != nil {
		return dberrors.Internal("Failed to marshal global configuration: %s", err.Error())
	}

	_, err = ws.insertInto("kyma_config").
		Pair("id", kymaConfig.ID).
		Pair("release_id", kymaConfig.Release.Id).
		Pair("cluster_id", kymaConfig.ClusterID).
		Pair("global_configuration", jsonConfig).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to insert record to KymaConfig table: %s", err)
	}

	for _, kymaConfigModule := range kymaConfig.Components {
		err = ws.insertKymaComponentConfig(kymaConfigModule)
		if err != nil {
			return dberrors.Internal("Failed to insert record to KymaComponentConfig table: %s", err)
		}
	}

	return nil
}

func (ws writeSession) insertKymaComponentConfig(kymaConfigModule model.KymaComponentConfig) dberrors.Error {
	jsonConfig, err := json.Marshal(kymaConfigModule.Configuration)
	if err != nil {
		return dberrors.Internal("Failed to marshal %s component configuration: %s", kymaConfigModule.Component, err.Error())
	}

	_, err = ws.insertInto("kyma_component_config").
		Pair("id", kymaConfigModule.ID).
		Pair("component", kymaConfigModule.Component).
		Pair("namespace", kymaConfigModule.Namespace).
		Pair("kyma_config_id", kymaConfigModule.KymaConfigID).
		Pair("configuration", jsonConfig).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to insert record to KymaComponentConfig table: %s", err)
	}

	return nil
}

func (ws writeSession) InsertOperation(operation model.Operation) dberrors.Error {
	_, err := ws.insertInto("operation").
		Columns("id", "type", "state", "message", "start_timestamp", "cluster_id").
		Record(operation).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to insert record to Type table: %s", err)
	}

	return nil
}

func (ws writeSession) DeleteCluster(runtimeID string) dberrors.Error {
	result, err := ws.deleteFrom("cluster").
		Where(dbr.Eq("id", runtimeID)).
		Exec()

	if err != nil {
		return dberrors.Internal("Failed to delete record in Cluster table: %s", err)
	}

	val, err := result.RowsAffected()

	if err != nil {
		return dberrors.Internal("Could not fetch the number of rows affected: %s", err)
	}

	if val == 0 {
		return dberrors.NotFound("Runtime with ID %s not found", runtimeID)
	}

	return nil
}

func (ws writeSession) UpdateOperationState(operationID string, message string, state model.OperationState) dberrors.Error {
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

func (ws writeSession) UpdateCluster(runtimeID string, kubeconfig string, terraformState []byte) dberrors.Error {
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

func (ws writeSession) updateSucceeded(result sql.Result, errorMsg string) dberrors.Error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return dberrors.Internal("Failed to get number of rows affected: %s", err)
	}

	if rowsAffected == 0 {
		return dberrors.NotFound(errorMsg)
	}

	return nil
}

func (ws writeSession) Commit() dberrors.Error {
	err := ws.transaction.Commit()
	if err != nil {
		return dberrors.Internal("Failed to commit transaction: %s", err)
	}

	return nil
}

func (ws writeSession) RollbackUnlessCommitted() {
	ws.transaction.RollbackUnlessCommitted()
}

func (ws writeSession) insertInto(table string) *dbr.InsertStmt {
	if ws.transaction != nil {
		return ws.transaction.InsertInto(table)
	}

	return ws.session.InsertInto(table)
}

func (ws writeSession) deleteFrom(table string) *dbr.DeleteStmt {
	if ws.transaction != nil {
		return ws.transaction.DeleteFrom(table)
	}

	return ws.session.DeleteFrom(table)
}

func (ws writeSession) update(table string) *dbr.UpdateStmt {
	if ws.transaction != nil {
		return ws.transaction.Update(table)
	}

	return ws.session.Update(table)
}
