package provisioning

import (
	"time"

	"github.com/kyma-project/control-plane/components/provisioner/internal/apperrors"

	"github.com/kyma-project/control-plane/components/provisioner/internal/operations/queue"

	uuid "github.com/kyma-project/control-plane/components/provisioner/internal/uuid"

	"github.com/kyma-project/control-plane/components/provisioner/internal/persistence/dberrors"

	"github.com/kyma-project/control-plane/components/provisioner/internal/provisioning/persistence/dbsession"

	"github.com/kyma-project/control-plane/components/provisioner/internal/director"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/control-plane/components/provisioner/internal/model"
	"github.com/kyma-project/control-plane/components/provisioner/pkg/gqlschema"
)

//go:generate mockery -name=Service
type Service interface {
	ProvisionRuntime(config gqlschema.ProvisionRuntimeInput, tenant, subAccount string) (*gqlschema.OperationStatus, apperrors.AppError)
	UpgradeRuntime(id string, config gqlschema.UpgradeRuntimeInput) (*gqlschema.OperationStatus, apperrors.AppError)
	DeprovisionRuntime(id, tenant string) (string, apperrors.AppError)
	ReconnectRuntimeAgent(id string) (string, apperrors.AppError)
	RuntimeStatus(id string) (*gqlschema.RuntimeStatus, apperrors.AppError)
	RuntimeOperationStatus(id string) (*gqlschema.OperationStatus, apperrors.AppError)
	RollBackLastUpgrade(runtimeID string) (*gqlschema.RuntimeStatus, apperrors.AppError)
}

//go:generate mockery -name=Provisioner
type Provisioner interface {
	ProvisionCluster(cluster model.Cluster, operationId string) apperrors.AppError
	DeprovisionCluster(cluster model.Cluster, operationId string) (model.Operation, apperrors.AppError)
}

type service struct {
	inputConverter   InputConverter
	graphQLConverter GraphQLConverter
	directorService  director.DirectorClient

	dbSessionFactory dbsession.Factory
	provisioner      Provisioner
	uuidGenerator    uuid.UUIDGenerator

	provisioningQueue   queue.OperationQueue
	deprovisioningQueue queue.OperationQueue
	upgradeQueue        queue.OperationQueue
}

func NewProvisioningService(
	inputConverter InputConverter,
	graphQLConverter GraphQLConverter,
	directorService director.DirectorClient,
	factory dbsession.Factory,
	provisioner Provisioner,
	generator uuid.UUIDGenerator,
	provisioningQueue queue.OperationQueue,
	deprovisioningQueue queue.OperationQueue,
	upgradeQueue queue.OperationQueue,
) Service {
	return &service{
		inputConverter:      inputConverter,
		graphQLConverter:    graphQLConverter,
		directorService:     directorService,
		dbSessionFactory:    factory,
		provisioner:         provisioner,
		uuidGenerator:       generator,
		provisioningQueue:   provisioningQueue,
		deprovisioningQueue: deprovisioningQueue,
		upgradeQueue:        upgradeQueue,
	}
}

func (r *service) ProvisionRuntime(config gqlschema.ProvisionRuntimeInput, tenant, subAccount string) (*gqlschema.OperationStatus, apperrors.AppError) {
	runtimeInput := config.RuntimeInput

	runtimeID, err := r.directorService.CreateRuntime(runtimeInput, tenant)
	if err != nil {
		return nil, err.Append("Failed to register Runtime")
	}

	cluster, err := r.inputConverter.ProvisioningInputToCluster(runtimeID, config, tenant, subAccount)
	if err != nil {
		r.unregisterFailedRuntime(runtimeID, tenant)
		return nil, err
	}

	dbSession, dberr := r.dbSessionFactory.NewSessionWithinTransaction()
	if dberr != nil {
		return nil, apperrors.Internal("Failed to start database transaction: %s", dberr.Error())
	}
	defer dbSession.RollbackUnlessCommitted()

	// Try to set provisioning started before triggering it (which is hard to interrupt) to verify all unique constraints
	operation, dberr := r.setProvisioningStarted(dbSession, runtimeID, cluster)
	if dberr != nil {
		r.unregisterFailedRuntime(runtimeID, tenant)
		return nil, apperrors.Internal(dberr.Error())
	}

	err = r.provisioner.ProvisionCluster(cluster, operation.ID)
	if err != nil {
		r.unregisterFailedRuntime(runtimeID, tenant)
		return nil, err.Append("Failed to start provisioning")
	}

	dberr = dbSession.Commit()
	if dberr != nil {
		r.unregisterFailedRuntime(runtimeID, tenant)
		return nil, apperrors.Internal("Failed to commit transaction: %s", dberr.Error())
	}

	r.provisioningQueue.Add(operation.ID)

	return r.graphQLConverter.OperationStatusToGQLOperationStatus(operation), nil
}

func (r *service) unregisterFailedRuntime(id, tenant string) {
	log.Infof("Starting provisioning failed. Unregistering Runtime %s...", id)
	err := r.directorService.DeleteRuntime(id, tenant)
	if err != nil {
		log.Warnf("Failed to unregister failed Runtime %s: %s", id, err.Error())
	}
}

func (r *service) DeprovisionRuntime(id, tenant string) (string, apperrors.AppError) {
	session := r.dbSessionFactory.NewReadWriteSession()

	err := r.verifyLastOperationFinished(session, id)
	if err != nil {
		return "", err
	}

	cluster, dberr := session.GetCluster(id)
	if dberr != nil {
		return "", apperrors.Internal("Failed to get cluster: %s", dberr.Error())
	}

	operation, err := r.provisioner.DeprovisionCluster(cluster, r.uuidGenerator.New())
	if err != nil {
		return "", apperrors.Internal("Failed to start deprovisioning: %s", err.Error())
	}

	dberr = session.InsertOperation(operation)
	if dberr != nil {
		return "", apperrors.Internal("Failed to insert operation to database: %s", dberr.Error())
	}

	r.deprovisioningQueue.Add(operation.ID)

	return operation.ID, nil
}

func (r *service) verifyLastOperationFinished(session dbsession.ReadSession, runtimeId string) apperrors.AppError {
	lastOperation, dberr := session.GetLastOperation(runtimeId)
	if dberr != nil {
		return apperrors.Internal("Failed to get last operation: %s", dberr.Error())
	}

	if lastOperation.State == model.InProgress {
		return apperrors.BadRequest("cannot start new operation for %s Runtime while previous one is in progress", runtimeId)
	}

	return nil
}

func (r *service) UpgradeRuntime(runtimeId string, input gqlschema.UpgradeRuntimeInput) (*gqlschema.OperationStatus, apperrors.AppError) {
	if input.KymaConfig == nil {
		return &gqlschema.OperationStatus{}, apperrors.BadRequest("error: Kyma config is nil")
	}

	session := r.dbSessionFactory.NewReadSession()

	err := r.verifyLastOperationFinished(session, runtimeId)
	if err != nil {
		return &gqlschema.OperationStatus{}, err
	}

	kymaConfig, err := r.inputConverter.KymaConfigFromInput(runtimeId, *input.KymaConfig)
	if err != nil {
		return &gqlschema.OperationStatus{}, err.Append("failed to convert KymaConfigInput")
	}

	cluster, dberr := session.GetCluster(runtimeId)
	if dberr != nil {
		return &gqlschema.OperationStatus{}, apperrors.Internal("failed to read cluster from database: %s", dberr.Error())
	}

	txSession, dberr := r.dbSessionFactory.NewSessionWithinTransaction()
	if dberr != nil {
		return &gqlschema.OperationStatus{}, apperrors.Internal("failed to start database transaction: %s", dberr.Error())
	}
	defer txSession.RollbackUnlessCommitted()

	operation, dberr := r.setUpgradeStarted(txSession, cluster, kymaConfig)
	if dberr != nil {
		return &gqlschema.OperationStatus{}, apperrors.Internal("failed to set upgrade started: %s", dberr.Error())
	}

	dberr = txSession.Commit()
	if dberr != nil {
		return &gqlschema.OperationStatus{}, apperrors.Internal("failed to commit upgrade transaction: %s", dberr.Error())
	}

	r.upgradeQueue.Add(operation.ID)

	return r.graphQLConverter.OperationStatusToGQLOperationStatus(operation), nil
}

func (r *service) ReconnectRuntimeAgent(id string) (string, apperrors.AppError) {
	return "", nil
}

func (r *service) RuntimeStatus(runtimeID string) (*gqlschema.RuntimeStatus, apperrors.AppError) {
	runtimeStatus, dberr := r.getRuntimeStatus(runtimeID)
	if dberr != nil {
		return nil, apperrors.Internal("failed to get Runtime Status: %s", dberr.Error())
	}

	return r.graphQLConverter.RuntimeStatusToGraphQLStatus(runtimeStatus), nil
}

func (r *service) RuntimeOperationStatus(operationID string) (*gqlschema.OperationStatus, apperrors.AppError) {
	readSession := r.dbSessionFactory.NewReadSession()

	operation, dberr := readSession.GetOperation(operationID)
	if dberr != nil {
		return nil, apperrors.Internal("failed to get Runtime Operation Status: %s", dberr.Error())
	}

	return r.graphQLConverter.OperationStatusToGQLOperationStatus(operation), nil
}

func (r *service) RollBackLastUpgrade(runtimeID string) (*gqlschema.RuntimeStatus, apperrors.AppError) {

	readSession := r.dbSessionFactory.NewReadSession()

	lastOp, err := readSession.GetLastOperation(runtimeID)
	if err != nil {
		return nil, apperrors.Internal("error rolling back last upgrade: %s", err.Error())
	}

	if lastOp.Type != model.Upgrade || lastOp.State == model.InProgress {
		return nil, apperrors.BadRequest("error: upgrade can be rolled back only if it is the last operation that is already finished")
	}

	runtimeUpgrade, err := readSession.GetRuntimeUpgrade(lastOp.ID)
	if err != nil {
		return nil, apperrors.Internal("error rolling back last upgrade: %s", err.Error())
	}

	txSession, err := r.dbSessionFactory.NewSessionWithinTransaction()
	if err != nil {
		return nil, apperrors.Internal("error rolling back last upgrade: %s", err.Error())
	}
	defer txSession.RollbackUnlessCommitted()

	err = txSession.SetActiveKymaConfig(runtimeID, runtimeUpgrade.PreUpgradeKymaConfigId)
	if err != nil {
		return nil, apperrors.Internal("error rolling back last upgrade: %s", err.Error())
	}

	err = txSession.UpdateUpgradeState(lastOp.ID, model.UpgradeRolledBack)
	if err != nil {
		return nil, apperrors.Internal("error rolling back last upgrade: %s", err.Error())
	}

	err = txSession.Commit()
	if err != nil {
		return nil, apperrors.Internal("error rolling back last upgrade: %s", err.Error())
	}

	return r.RuntimeStatus(runtimeID)
}

func (r *service) getRuntimeStatus(runtimeID string) (model.RuntimeStatus, dberrors.Error) {
	session := r.dbSessionFactory.NewReadSession()

	operation, err := session.GetLastOperation(runtimeID)
	if err != nil {
		return model.RuntimeStatus{}, err
	}

	cluster, err := session.GetCluster(runtimeID)
	if err != nil {
		return model.RuntimeStatus{}, err
	}

	return model.RuntimeStatus{
		LastOperationStatus:  operation,
		RuntimeConfiguration: cluster,
	}, nil
}

func (r *service) setProvisioningStarted(dbSession dbsession.WriteSession, runtimeID string, cluster model.Cluster) (model.Operation, dberrors.Error) {
	timestamp := time.Now()

	cluster.CreationTimestamp = timestamp

	err := dbSession.InsertCluster(cluster)
	if err != nil {
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	err = dbSession.InsertGardenerConfig(cluster.ClusterConfig)
	if err != nil {
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	err = dbSession.InsertKymaConfig(cluster.KymaConfig)
	if err != nil {
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	operation, err := r.setOperationStarted(dbSession, runtimeID, model.Provision, model.WaitingForClusterDomain, timestamp, "Provisioning started")
	if err != nil {
		return model.Operation{}, err.Append("Failed to set provisioning started: %s")
	}

	return operation, nil
}

func (r *service) setUpgradeStarted(txSession dbsession.WriteSession, cluster model.Cluster, kymaConfig model.KymaConfig) (model.Operation, dberrors.Error) {

	err := txSession.InsertKymaConfig(kymaConfig)
	if err != nil {
		return model.Operation{}, err.Append("Failed to insert Kyma Config")
	}

	operation, err := r.setOperationStarted(txSession, cluster.ID, model.Upgrade, model.StartingUpgrade, time.Now(), "Starting Kyma upgrade")
	if err != nil {
		return model.Operation{}, err.Append("Failed to set operation started")
	}

	runtimeUpgrade := model.RuntimeUpgrade{
		Id:                      r.uuidGenerator.New(),
		State:                   model.UpgradeInProgress,
		OperationId:             operation.ID,
		PreUpgradeKymaConfigId:  cluster.KymaConfig.ID,
		PostUpgradeKymaConfigId: kymaConfig.ID,
	}

	err = txSession.InsertRuntimeUpgrade(runtimeUpgrade)
	if err != nil {
		return model.Operation{}, err.Append("Failed to insert Runtime Upgrade")
	}

	err = txSession.SetActiveKymaConfig(cluster.ID, kymaConfig.ID)
	if err != nil {
		return model.Operation{}, err.Append("Failed to update Kyma config in cluster")
	}

	return operation, nil
}

func (r *service) setOperationStarted(
	dbSession dbsession.WriteSession,
	runtimeID string,
	operationType model.OperationType,
	operationStage model.OperationStage,
	timestamp time.Time,
	message string) (model.Operation, dberrors.Error) {
	id := r.uuidGenerator.New()

	operation := model.Operation{
		ID:             id,
		Type:           operationType,
		StartTimestamp: timestamp,
		State:          model.InProgress,
		Message:        message,
		ClusterID:      runtimeID,
		Stage:          operationStage,
		LastTransition: &timestamp,
	}

	err := dbSession.InsertOperation(operation)
	if err != nil {
		return model.Operation{}, err.Append("failed to insert operation")
	}

	return operation, nil
}
