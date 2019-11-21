package persistence

import (
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dbsession"

	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
)

//go:generate mockery -name=Service
type Service interface {
	GetStatus(runtimeID string) (model.RuntimeStatus, dberrors.Error)
	SetProvisioningStarted(runtimeID string, runtimeConfig model.RuntimeConfig) (model.Operation, dberrors.Error)
	SetDeprovisioningStarted(runtimeID string) (model.Operation, dberrors.Error)
	SetUpgradeStarted(runtimeID string) (model.Operation, dberrors.Error)
	GetLastOperation(runtimeID string) (model.Operation, dberrors.Error)
	Update(runtimeID string, kubeconfig string, terraformState string) dberrors.Error
	CleanupClusterData(runtimeID string) dberrors.Error
	GetClusterData(runtimeID string) (model.Cluster, dberrors.Error)
	Get(operationID string) (model.Operation, error)
	SetAsFailed(operationID string, message string) error
	SetAsSucceeded(operationID string) error
}

type persistenceService struct {
	dbSessionFactory dbsession.Factory
	uuidGenerator    UUIDGenerator
}

func NewService(dbSessionFactory dbsession.Factory, uuidGenerator UUIDGenerator) Service {
	return persistenceService{
		dbSessionFactory: dbSessionFactory,
		uuidGenerator:    uuidGenerator,
	}
}

func (ps persistenceService) GetStatus(runtimeID string) (model.RuntimeStatus, dberrors.Error) {
	session := ps.dbSessionFactory.NewReadSession()

	operation, err := session.GetLastOperation(runtimeID)
	if err != nil {
		return model.RuntimeStatus{}, err
	}

	clusterConfig, err := session.GetClusterConfig(runtimeID)
	if err != nil {
		return model.RuntimeStatus{}, err
	}

	kymaConfig, err := session.GetKymaConfig(runtimeID)
	if err != nil {
		return model.RuntimeStatus{}, err
	}

	cluster, err := session.GetCluster(runtimeID)
	if err != nil {
		return model.RuntimeStatus{}, err
	}

	runtimeConfiguration := model.RuntimeConfig{
		KymaConfig:            kymaConfig,
		ClusterConfig:         clusterConfig,
		Kubeconfig:            cluster.Kubeconfig,
		CredentialsSecretName: cluster.CredentialsSecretName,
	}

	return model.RuntimeStatus{
		LastOperationStatus:  operation,
		RuntimeConfiguration: runtimeConfiguration,
	}, nil
}

func (ps persistenceService) SetProvisioningStarted(runtimeID string, runtimeConfig model.RuntimeConfig) (model.Operation, dberrors.Error) {
	dbSession, err := ps.dbSessionFactory.NewSessionWithinTransaction()
	if err != nil {
		return model.Operation{}, dberrors.Internal("Failed to create repository: %s", err)
	}

	defer dbSession.RollbackUnlessCommitted()

	timestamp := time.Now()

	cluster := model.Cluster{
		ID:                    runtimeID,
		CreationTimestamp:     timestamp,
		CredentialsSecretName: runtimeConfig.CredentialsSecretName,
		TerraformState:        "{}",
	}

	err = dbSession.InsertCluster(cluster)
	if err != nil {
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	gcpConfig, isGCP := runtimeConfig.GCPConfig()
	if isGCP {

		err = dbSession.InsertGCPConfig(gcpConfig)
		if err != nil {
			return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
		}
	}

	gardenerConfig, isGardener := runtimeConfig.GardenerConfig()
	if isGardener {
		err = dbSession.InsertGardenerConfig(gardenerConfig)
		if err != nil {
			return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
		}
	}

	err = dbSession.InsertKymaConfig(runtimeConfig.KymaConfig)
	if err != nil {
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	operation, err := ps.setOperationStarted(dbSession, runtimeID, model.Provision, timestamp, "Provisioning started", "Failed to set provisioning started: %s")

	if err != nil {
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	err = dbSession.Commit()
	if err != nil {
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	return operation, nil
}

func (ps persistenceService) SetDeprovisioningStarted(runtimeID string) (model.Operation, dberrors.Error) {
	return ps.setOperationStarted(ps.dbSessionFactory.NewWriteSession(), runtimeID, model.Deprovision, time.Now(), "Deprovisioning started.", "Deprovisioning failed: %s")
}

func (ps persistenceService) SetUpgradeStarted(runtimeID string) (model.Operation, dberrors.Error) {
	return ps.setOperationStarted(ps.dbSessionFactory.NewWriteSession(), runtimeID, model.Upgrade, time.Now(), "Upgrade started.", "Upgrade failed: %s")
}

func (ps persistenceService) GetLastOperation(runtimeID string) (model.Operation, dberrors.Error) {
	session := ps.dbSessionFactory.NewReadSession()

	return session.GetLastOperation(runtimeID)
}

func (ps persistenceService) Update(runtimeID string, kubeconfig string, terraformState string) dberrors.Error {
	session := ps.dbSessionFactory.NewWriteSession()

	return session.UpdateCluster(runtimeID, kubeconfig, terraformState)
}

func (ps persistenceService) CleanupClusterData(runtimeID string) dberrors.Error {
	session := ps.dbSessionFactory.NewWriteSession()

	return session.DeleteCluster(runtimeID)
}

func (ps persistenceService) setOperationStarted(dbSession dbsession.WriteSession, runtimeID string, operationType model.OperationType, timestamp time.Time, message string, errorMessageFmt string) (model.Operation, dberrors.Error) {

	id := ps.uuidGenerator.New()

	operation := model.Operation{
		ID:             id,
		Type:           operationType,
		StartTimestamp: timestamp,
		State:          model.InProgress,
		Message:        message,
		ClusterID:      runtimeID,
	}

	err := dbSession.InsertOperation(operation)
	if err != nil {
		return model.Operation{}, dberrors.Internal(errorMessageFmt, err)
	}

	return operation, nil
}

func (ps persistenceService) GetClusterData(runtimeID string) (model.Cluster, dberrors.Error) {
	session := ps.dbSessionFactory.NewReadSession()

	return session.GetCluster(runtimeID)
}

func (ps persistenceService) Get(operationID string) (model.Operation, error) {
	session := ps.dbSessionFactory.NewReadSession()

	return session.GetOperation(operationID)
}

func (ps persistenceService) SetAsFailed(operationID string, message string) error {
	session := ps.dbSessionFactory.NewWriteSession()

	return session.UpdateOperationState(operationID, message, model.Failed)
}

func (ps persistenceService) SetAsSucceeded(operationID string) error {
	session := ps.dbSessionFactory.NewWriteSession()

	return session.UpdateOperationState(operationID, "Operation succeeded.", model.Succeeded)
}
