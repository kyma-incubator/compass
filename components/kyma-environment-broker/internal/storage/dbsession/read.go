package dbsession

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession/dbmodel"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/postsql"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/predicate"

	"github.com/gocraft/dbr"
	"github.com/pivotal-cf/brokerapi/v7/domain"
)

type readSession struct {
	session *dbr.Session
}

func (r readSession) FindAllInstancesJoinedWithProvisionOperation(prct ...predicate.Predicate) ([]internal.InstanceWithOperation, dberr.Error) {
	var instances []internal.InstanceWithOperation

	join := fmt.Sprintf("%s.instance_id = %s.instance_id", postsql.InstancesTableName, postsql.OperationTableName)
	stmt := r.session.
		Select("instances.instance_id, instances.runtime_id, instances.global_account_id, instances.service_id, instances.service_plan_id, instances.dashboard_url, instances.provisioning_parameters, instances.created_at, instances.updated_at, instances.delated_at, instances.sub_account_id, instances.service_name, instances.service_plan_name, operations.state, operations.description, operations.type").
		From(postsql.InstancesTableName).
		LeftJoin(postsql.OperationTableName, join)

	for _, p := range prct {
		p.ApplyToPostgres(stmt)
	}

	if _, err := stmt.Load(&instances); err != nil {
		return nil, dberr.Internal("Failed to fetch all instances: %s", err)
	}

	return instances, nil
}

func (r readSession) GetInstanceByID(instanceID string) (internal.Instance, dberr.Error) {
	var instance internal.Instance

	err := r.session.
		Select("*").
		From(postsql.InstancesTableName).
		Where(dbr.Eq("instance_id", instanceID)).
		LoadOne(&instance)

	if err != nil {
		if err == dbr.ErrNotFound {
			return internal.Instance{}, dberr.NotFound("Cannot find Instance for instanceID:'%s'", instanceID)
		}
		return internal.Instance{}, dberr.Internal("Failed to get Instance: %s", err)
	}
	return instance, nil
}

func (r readSession) GetOperationByID(opID string) (dbmodel.OperationDTO, dberr.Error) {
	condition := dbr.Eq("id", opID)
	operation, err := r.getOperation(condition)
	if err != nil {
		switch {
		case dberr.IsNotFound(err):
			return dbmodel.OperationDTO{}, dberr.NotFound("for ID: %s %s", opID, err)
		default:
			return dbmodel.OperationDTO{}, err
		}
	}
	return operation, nil
}

func (r readSession) GetOperationsInProgressByType(operationType dbmodel.OperationType) ([]dbmodel.OperationDTO, dberr.Error) {
	stateCondition := dbr.Eq("state", domain.InProgress)
	typeCondition := dbr.Eq("type", operationType)
	var operations []dbmodel.OperationDTO

	_, err := r.session.
		Select("*").
		From(postsql.OperationTableName).
		Where(stateCondition).
		Where(typeCondition).
		Load(&operations)
	if err != nil {
		return nil, dberr.Internal("Failed to get operations: %s", err)
	}
	return operations, nil
}

func (r readSession) GetOperationByTypeAndInstanceID(inID string, opType dbmodel.OperationType) (dbmodel.OperationDTO, dberr.Error) {
	idCondition := dbr.Eq("instance_id", inID)
	typeCondition := dbr.Eq("type", string(opType))
	var operation dbmodel.OperationDTO

	err := r.session.
		Select("*").
		From(postsql.OperationTableName).
		Where(idCondition).
		Where(typeCondition).
		LoadOne(&operation)

	if err != nil {
		if err == dbr.ErrNotFound {
			return dbmodel.OperationDTO{}, dberr.NotFound("cannot find operation: %s", err)
		}
		return dbmodel.OperationDTO{}, dberr.Internal("Failed to get operation: %s", err)
	}
	return operation, nil
}

func (r readSession) getOperation(condition dbr.Builder) (dbmodel.OperationDTO, dberr.Error) {
	var operation dbmodel.OperationDTO

	err := r.session.
		Select("*").
		From(postsql.OperationTableName).
		Where(condition).
		LoadOne(&operation)

	if err != nil {
		if err == dbr.ErrNotFound {
			return dbmodel.OperationDTO{}, dberr.NotFound("cannot find operation: %s", err)
		}
		return dbmodel.OperationDTO{}, dberr.Internal("Failed to get operation: %s", err)
	}
	return operation, nil
}

func (r readSession) GetLMSTenant(name, region string) (dbmodel.LMSTenantDTO, dberr.Error) {
	var dto dbmodel.LMSTenantDTO
	err := r.session.
		Select("*").
		From(postsql.LMSTenantTableName).
		Where(dbr.Eq("name", name)).
		Where(dbr.Eq("region", region)).
		LoadOne(&dto)

	if err != nil {
		if err == dbr.ErrNotFound {
			return dbmodel.LMSTenantDTO{}, dberr.NotFound("Cannot find lms tenant for name/region: '%s/%s'", name, region)
		}
		return dbmodel.LMSTenantDTO{}, dberr.Internal("Failed to get operation: %s", err)
	}
	return dto, nil
}
