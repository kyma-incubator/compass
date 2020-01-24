package dbsession

import (
	dbr "github.com/gocraft/dbr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/postsql"
)

type writeSession struct {
	session     *dbr.Session
	transaction *dbr.Tx
}

func (ws writeSession) InsertInstance(instance internal.Instance) dberr.Error {
	_, err := ws.insertInto(postsql.InstancesTableName).
		Pair("instance_id", instance.InstanceID).
		Pair("runtime_id", instance.RuntimeID).
		Pair("global_account_id", instance.GlobalAccountID).
		Pair("service_id", instance.ServiceID).
		Pair("service_plan_id", instance.ServicePlanID).
		Pair("dashboard_url", instance.DashboardURL).
		Pair("provisioning_parameters", instance.ProvisioningParameters).
		Exec()
	if err != nil {
		return dberr.Internal("Failed to insert record to Instance table: %s", err)
	}

	return nil
}

func (ws writeSession) UpdateInstance(instance internal.Instance) dberr.Error {
	_, err := ws.update(postsql.InstancesTableName).
		Where(dbr.Eq(postsql.InstancesTableName+".instance_id", instance.InstanceID)).
		Set("instance_id", instance.InstanceID).
		Set("runtime_id", instance.RuntimeID).
		Set("global_account_id", instance.GlobalAccountID).
		Set("service_id", instance.ServiceID).
		Set("service_plan_id", instance.ServicePlanID).
		Set("dashboard_url", instance.DashboardURL).
		Set("provisioning_parameters", instance.ProvisioningParameters).
		Exec()
	if err != nil {
		return dberr.Internal("Failed to update record to Instance table: %s", err)
	}

	return nil
}

func (ws writeSession) Commit() dberr.Error {
	err := ws.transaction.Commit()
	if err != nil {
		return dberr.Internal("Failed to commit transaction: %s", err)
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
