package provisioning

import (
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-incubator/hydroform/types"
)

const (
	interval   = 2 * time.Second
	retryCount = 5
)

//go:generate mockery -name=Service
type Service interface {
	ProvisionRuntime(id string, config gqlschema.ProvisionRuntimeInput) (string, <-chan struct{}, error)
	UpgradeRuntime(id string, config *gqlschema.UpgradeRuntimeInput) (string, error)
	DeprovisionRuntime(id string, credentials gqlschema.CredentialsInput) (string, <-chan struct{}, error)
	CleanupRuntimeData(id string) (string, error)
	ReconnectRuntimeAgent(id string) (string, error)
	RuntimeStatus(id string) (*gqlschema.RuntimeStatus, error)
	RuntimeOperationStatus(id string) (*gqlschema.OperationStatus, error)
}

type service struct {
	persistenceService persistence.Service
	hydroform          hydroform.Service
	uuidGenerator      persistence.UUIDGenerator
}

func NewProvisioningService(persistenceService persistence.Service, uuidGenerator persistence.UUIDGenerator, hydroform hydroform.Service) Service {
	return &service{
		persistenceService: persistenceService,
		hydroform:          hydroform,
		uuidGenerator:      uuidGenerator,
	}
}

func (r *service) ProvisionRuntime(id string, config gqlschema.ProvisionRuntimeInput) (string, <-chan struct{}, error) {
	err := r.checkProvisioningRuntimeConditions(id)
	if err != nil {
		return "", nil, err
	}

	runtimeConfig := runtimeConfigFromInput(id, config, r.uuidGenerator)

	operation, err := r.persistenceService.SetProvisioningStarted(id, runtimeConfig)

	if err != nil {
		return "", nil, err
	}

	finished := make(chan struct{})

	go r.startProvisioning(operation.ID, id, runtimeConfig, config.Credentials.SecretName, finished)

	return operation.ID, finished, err
}

func (r *service) checkProvisioningRuntimeConditions(id string) error {
	lastOperation, err := r.persistenceService.GetLastOperation(id)

	if err == nil && !lastProvisioningFailed(lastOperation) {
		return errors.New(fmt.Sprintf("cannot provision runtime. Runtime %s already provisioned", id))
	}

	if err != nil && err.Code() != dberrors.CodeNotFound {
		return err
	}

	if lastProvisioningFailed(lastOperation) {
		if _, dbErr := r.CleanupRuntimeData(id); dbErr != nil {
			return dbErr
		}
	}

	return nil
}

func lastProvisioningFailed(operation model.Operation) bool {
	return operation.Type == model.Provision && operation.State == model.Failed
}

func (r *service) DeprovisionRuntime(id string, credentials gqlschema.CredentialsInput) (string, <-chan struct{}, error) {
	runtimeStatus, err := r.persistenceService.GetStatus(id)

	if err != nil {
		return "", nil, err
	}

	if runtimeStatus.LastOperationStatus.State == model.InProgress {
		return "", nil, errors.New("cannot start new operation while previous one is in progress")
	}

	operation, err := r.persistenceService.SetDeprovisioningStarted(id)

	if err != nil {
		return "", nil, err
	}

	finished := make(chan struct{})

	cluster, dberr := r.persistenceService.GetClusterData(id)

	if dberr != nil {
		return "", nil, dberr
	}

	//TODO For now we pass secret name in parameters but we need to consider if it should be stored in the database
	go r.startDeprovisioning(operation.ID, id, runtimeStatus.RuntimeConfiguration, credentials.SecretName, cluster, finished)

	return operation.ID, finished, nil
}

func (r *service) UpgradeRuntime(id string, config *gqlschema.UpgradeRuntimeInput) (string, error) {
	return "", nil
}

func (r *service) ReconnectRuntimeAgent(id string) (string, error) {
	return "", nil
}

func (r *service) RuntimeStatus(runtimeID string) (*gqlschema.RuntimeStatus, error) {
	runtimeStatus, err := r.persistenceService.GetStatus(runtimeID)

	if err != nil {
		return nil, err
	}

	status := runtimeStatusToGraphQLStatus(runtimeStatus)

	return status, nil
}

func (r *service) RuntimeOperationStatus(operationID string) (*gqlschema.OperationStatus, error) {
	operation, err := r.persistenceService.Get(operationID)

	if err != nil {
		return nil, err
	}

	status := operationStatusToGQLOperationStatus(operation)

	return status, nil
}

func (r *service) CleanupRuntimeData(id string) (string, error) {
	return id, r.persistenceService.CleanupClusterData(id)
}

func (r *service) startProvisioning(operationID, runtimeID string, config model.RuntimeConfig, secretName string, finished chan<- struct{}) {
	defer close(finished)
	log.Infof("Provisioning runtime %s is starting", runtimeID)
	info, err := r.hydroform.ProvisionCluster(config, secretName)

	if err != nil || info.ClusterStatus != types.Provisioned {
		log.Errorf("Provisioning runtime %s failed: %s", runtimeID, err.Error())
		updateOperationStatus(func() error {
			return r.persistenceService.SetAsFailed(operationID, err.Error())
		})
	} else {
		log.Infof("Provisioning runtime %s finished successfully", runtimeID)
		updateOperationStatus(func() error {
			err := r.persistenceService.Update(runtimeID, info.KubeConfig, info.State)
			if err != nil {
				return r.persistenceService.SetAsFailed(operationID, err.Error())
			}
			return r.persistenceService.SetAsSucceeded(operationID)
		})
	}
}

func (r *service) startDeprovisioning(operationID, runtimeID string, config model.RuntimeConfig, secretName string, cluster model.Cluster, finished chan<- struct{}) {
	defer close(finished)
	log.Infof("Deprovisioning runtime %s is starting", runtimeID)
	err := r.hydroform.DeprovisionCluster(config, secretName, cluster.TerraformState)

	if err != nil {
		log.Errorf("Deprovisioning runtime %s failed: %s", runtimeID, err.Error())
		updateOperationStatus(func() error {
			return r.persistenceService.SetAsFailed(operationID, err.Error())
		})
	} else {
		log.Infof("Deprovisioning runtime %s finished successfully", runtimeID)
		updateOperationStatus(func() error {
			return r.persistenceService.SetAsSucceeded(operationID)
		})
	}
}

func updateOperationStatus(updateFunction func() error) {
	err := retry(interval, retryCount, updateFunction)
	if err != nil {
		log.Errorf("Failed to set operation status, %s", err.Error())
	}
}

func retry(interval time.Duration, count int, operation func() error) error {
	var err error
	for i := 0; i < count; i++ {
		err = operation()
		if err == nil {
			return nil
		}
		log.Errorf("Error during updating operation status: %s", err.Error())
		time.Sleep(interval)
	}

	return err
}
