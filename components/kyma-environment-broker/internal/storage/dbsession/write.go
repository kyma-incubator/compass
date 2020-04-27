package dbsession

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession/dbmodel"

	"github.com/gocraft/dbr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/postsql"
	"github.com/lib/pq"
)

const (
	UniqueViolationErrorCode = "23505"
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
		Pair("sub_account_id", instance.SubAccountID).
		Pair("service_id", instance.ServiceID).
		Pair("service_name", instance.ServiceName).
		Pair("service_plan_id", instance.ServicePlanID).
		Pair("service_plan_name", instance.ServicePlanName).
		Pair("dashboard_url", instance.DashboardURL).
		Pair("provisioning_parameters", instance.ProvisioningParameters).
		// in postgres database it will be equal to "0001-01-01 00:00:00+00"
		Pair("deleted_at", time.Time{}).
		Exec()

	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == UniqueViolationErrorCode {
				return dberr.AlreadyExists("operation with id %s already exist", instance.InstanceID)
			}
		}
		return dberr.Internal("Failed to insert record to Instance table: %s", err)
	}

	return nil
}

func (ws writeSession) DeleteInstance(instanceID string) dberr.Error {
	_, err := ws.deleteFrom(postsql.InstancesTableName).
		Where(dbr.Eq("instance_id", instanceID)).
		Exec()

	if err != nil {
		return dberr.Internal("Failed to delete record from Instance table: %s", err)
	}
	return nil
}

func (ws writeSession) UpdateInstance(instance internal.Instance) dberr.Error {
	_, err := ws.update(postsql.InstancesTableName).
		Where(dbr.Eq("instance_id", instance.InstanceID)).
		Set("instance_id", instance.InstanceID).
		Set("runtime_id", instance.RuntimeID).
		Set("global_account_id", instance.GlobalAccountID).
		Set("service_id", instance.ServiceID).
		Set("service_plan_id", instance.ServicePlanID).
		Set("dashboard_url", instance.DashboardURL).
		Set("provisioning_parameters", instance.ProvisioningParameters).
		Set("updated_at", time.Now()).
		Exec()
	if err != nil {
		return dberr.Internal("Failed to update record to Instance table: %s", err)
	}

	return nil
}

func (ws writeSession) InsertOperation(op dbmodel.OperationDTO) dberr.Error {
	_, err := ws.insertInto(postsql.OperationTableName).
		Pair("id", op.ID).
		Pair("instance_id", op.InstanceID).
		Pair("version", op.Version).
		Pair("created_at", op.CreatedAt).
		Pair("updated_at", op.UpdatedAt).
		Pair("description", op.Description).
		Pair("state", op.State).
		Pair("target_operation_id", op.TargetOperationID).
		Pair("type", op.Type).
		Pair("data", op.Data).
		Exec()

	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == UniqueViolationErrorCode {
				return dberr.AlreadyExists("operation with id %s already exist", op.ID)
			}
		}
		return dberr.Internal("Failed to insert record to operations table: %s", err)
	}

	return nil
}

func (ws writeSession) InsertLMSTenant(dto dbmodel.LMSTenantDTO) dberr.Error {
	_, err := ws.insertInto(postsql.LMSTenantTableName).
		Pair("id", dto.ID).
		Pair("name", dto.Name).
		Pair("region", dto.Region).
		Pair("created_at", dto.CreatedAt).
		Exec()

	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == UniqueViolationErrorCode {
				return dberr.AlreadyExists("lms tenant already exist")
			}
		}
		return dberr.Internal("Failed to insert record to lms tenant table: %s", err)
	}

	return nil
}

func (ws writeSession) UpdateOperation(op dbmodel.OperationDTO) dberr.Error {
	res, err := ws.update(postsql.OperationTableName).
		Where(dbr.Eq("id", op.ID)).
		Where(dbr.Eq("version", op.Version)).
		Set("instance_id", op.InstanceID).
		Set("version", op.Version+1).
		Set("created_at", op.CreatedAt).
		Set("updated_at", op.UpdatedAt).
		Set("description", op.Description).
		Set("state", op.State).
		Set("target_operation_id", op.TargetOperationID).
		Set("type", op.Type).
		Set("data", op.Data).
		Exec()

	if err != nil {
		if err == dbr.ErrNotFound {
			return dberr.NotFound("Cannot find Operation with ID:'%s'", op.ID)
		}
		return dberr.Internal("Failed to update record to Operation table: %s", err)
	}
	rAffected, e := res.RowsAffected()
	if e != nil {
		// the optimistic locking requires numbers of rows affected
		return dberr.Internal("the DB driver does not support RowsAffected operation")
	}
	if rAffected == int64(0) {
		return dberr.NotFound("Cannot find Operation with ID:'%s' Version: %v", op.ID, op.Version)
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
