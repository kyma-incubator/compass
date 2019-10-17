package persistence

import (
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dbsession"

	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/sirupsen/logrus"
)

//go:generate mockery -name=RuntimeService
type RuntimeService interface {
	GetStatus(runtimeID string) (model.RuntimeStatus, dberrors.Error)
	SetProvisioningStarted(runtimeID string, runtimeConfig model.RuntimeConfig) (model.Operation, dberrors.Error)
	SetDeprovisioningStarted(runtimeID string) (model.Operation, dberrors.Error)
	SetUpgradeStarted(runtimeID string) (model.Operation, dberrors.Error)
	GetLastOperation(runtimeID string) (model.Operation, dberrors.Error)
	Update(runtimeID string, kubeconfig string, terraformState string) dberrors.Error
	CleanupClusterData(runtimeID string) dberrors.Error
	GetClusterData(runtimeID string) (model.Cluster, dberrors.Error)
}

type runtimeService struct {
	dbSessionFactory dbsession.Factory
	uuidGenerator    UUIDGenerator
}

func NewRuntimeService(dbSessionFactory dbsession.Factory, uuidGenerator UUIDGenerator) RuntimeService {
	return runtimeService{
		dbSessionFactory: dbSessionFactory,
		uuidGenerator:    uuidGenerator,
	}
}

func (r runtimeService) GetStatus(runtimeID string) (model.RuntimeStatus, dberrors.Error) {
	session := r.dbSessionFactory.NewReadSession()

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
		KymaConfig:    kymaConfig,
		ClusterConfig: clusterConfig,
		Kubeconfig:    cluster.Kubeconfig,
	}

	return model.RuntimeStatus{
		LastOperationStatus:  operation,
		RuntimeConfiguration: runtimeConfiguration,
	}, nil
}

func (r runtimeService) SetProvisioningStarted(runtimeID string, runtimeConfig model.RuntimeConfig) (model.Operation, dberrors.Error) {
	dbSession, err := r.dbSessionFactory.NewSessionWithinTransaction()
	if err != nil {
		logrus.Errorf("Failed to create repository: %s", err)
	}

	defer dbSession.RollbackUnlessCommitted()

	timestamp := time.Now()

	cluster := model.Cluster{
		ID:                runtimeID,
		CreationTimestamp: timestamp,
		TerraformState:    "{}",
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

	operation, err := r.setOperationStarted(dbSession, runtimeID, model.Provision, timestamp, "Provisioning started", "Failed to set provisioning started: %s")

	if err != nil {
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	err = dbSession.Commit()
	if err != nil {
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	return operation, nil
}

func (r runtimeService) SetDeprovisioningStarted(runtimeID string) (model.Operation, dberrors.Error) {
	return r.setOperationStarted(r.dbSessionFactory.NewWriteSession(), runtimeID, model.Deprovision, time.Now(), "Deprovisioning started.", "Deprovisioning failed: %s")
}

func (r runtimeService) SetUpgradeStarted(runtimeID string) (model.Operation, dberrors.Error) {
	return r.setOperationStarted(r.dbSessionFactory.NewWriteSession(), runtimeID, model.Upgrade, time.Now(), "Upgrade started.", "Upgrade failed: %s")
}

func (r runtimeService) GetLastOperation(runtimeID string) (model.Operation, dberrors.Error) {
	session := r.dbSessionFactory.NewReadSession()

	return session.GetLastOperation(runtimeID)
}

func (r runtimeService) Update(runtimeID string, kubeconfig string, terraformState string) dberrors.Error {
	session := r.dbSessionFactory.NewWriteSession()

	return session.UpdateCluster(runtimeID, kubeconfig, terraformState)
}

func (r runtimeService) CleanupClusterData(runtimeID string) dberrors.Error {
	session := r.dbSessionFactory.NewWriteSession()

	return session.DeleteCluster(runtimeID)
}

func (r runtimeService) setOperationStarted(dbSession dbsession.WriteSession, runtimeID string, operationType model.OperationType, timestamp time.Time, message string, errorMessageFmt string) (model.Operation, dberrors.Error) {

	id, err := r.uuidGenerator.New()
	if err != nil {
		return model.Operation{}, dberrors.Internal(errorMessageFmt, err)
	}

	operation := model.Operation{
		ID:             id,
		Type:           operationType,
		StartTimestamp: timestamp,
		State:          model.InProgress,
		Message:        message,
		ClusterID:      runtimeID,
	}

	err = dbSession.InsertOperation(operation)
	if err != nil {
		return model.Operation{}, dberrors.Internal(errorMessageFmt, err)
	}

	return operation, nil
}

func (r runtimeService) GetClusterData(runtimeID string) (model.Cluster, dberrors.Error) {
	session := r.dbSessionFactory.NewReadSession()

	return session.GetCluster(runtimeID)
}
