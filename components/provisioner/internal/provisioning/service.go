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

type ProvisioningService interface {
	ProvisionRuntime(id string, config gqlschema.ProvisionRuntimeInput) (string, error, chan interface{})
	UpgradeRuntime(id string, config *gqlschema.UpgradeRuntimeInput) (string, error)
	DeprovisionRuntime(id string, credentials gqlschema.CredentialsInput) (string, error, chan interface{})
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

func NewProvisioningService(persistenceService persistence.Service, uuidGenerator persistence.UUIDGenerator, hydroform hydroform.Service) ProvisioningService {
	return &service{
		persistenceService: persistenceService,
		hydroform:          hydroform,
		uuidGenerator:      uuidGenerator,
	}
}

func (r *service) ProvisionRuntime(id string, config gqlschema.ProvisionRuntimeInput) (string, error, chan interface{}) {
	{
		lastOperation, err := r.persistenceService.GetLastOperation(id)

		if err == nil && !lastProvisioningFailed(lastOperation) {
			return "", errors.New(fmt.Sprintf("cannot provision runtime. Runtime %s already provisioned", id)), nil
		}

		if err != nil && err.Code() != dberrors.CodeNotFound {
			return "", err, nil
		}

		if lastProvisioningFailed(lastOperation) {
			if _, dbErr := r.CleanupRuntimeData(id); dbErr != nil {
				return "", dbErr, nil
			}
		}
	}

	runtimeConfig, err := runtimeConfigFromInput(id, config, r.uuidGenerator)
	if err != nil {
		return "", err, nil
	}

	operation, err := r.persistenceService.SetProvisioningStarted(id, runtimeConfig)

	if err != nil {
		return "", err, nil
	}

	finished := make(chan interface{})

	go r.startProvisioning(operation.ID, id, runtimeConfig, config.Credentials.SecretName, finished)

	return operation.ID, nil, finished
}

func lastProvisioningFailed(operation model.Operation) bool {
	return operation.Type == model.Provision && operation.State == model.Failed
}

func (r *service) DeprovisionRuntime(id string, credentials gqlschema.CredentialsInput) (string, error, chan interface{}) {
	runtimeStatus, err := r.persistenceService.GetStatus(id)

	if err != nil {
		return "", err, nil
	}

	if runtimeStatus.LastOperationStatus.State == model.InProgress {
		return "", errors.New("cannot start new operation while previous one is in progress"), nil
	}

	operation, err := r.persistenceService.SetDeprovisioningStarted(id)

	if err != nil {
		return "", err, nil
	}

	finished := make(chan interface{})

	//TODO For now we pass secret name in parameters but we need to consider if it should be stored in the database
	go r.startDeprovisioning(operation.ID, id, runtimeStatus.RuntimeConfiguration, credentials.SecretName, finished)

	return operation.ID, nil, finished
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

func (r *service) startProvisioning(operationID, runtimeID string, config model.RuntimeConfig, secretName string, finished chan interface{}) {
	log.Infof("Provisioning runtime %s is starting", runtimeID)
	info, err := r.hydroform.ProvisionCluster(config, secretName)

	if err != nil || info.ClusterStatus != types.Provisioned {
		updateOperationStatus(func() error {
			log.Errorf("Provisioning runtime %s failed: %s", runtimeID, err.Error())
			return r.persistenceService.SetAsFailed(operationID, err.Error())
		})
	} else {
		updateOperationStatus(func() error {
			err := r.persistenceService.Update(runtimeID, info.KubeConfig, info.State)
			if err != nil {
				log.Errorf("Provisioning runtime %s failed: %s", runtimeID, err.Error())
				return r.persistenceService.SetAsFailed(operationID, err.Error())
			}
			log.Infof("Provisioning runtime %s finished successfully", runtimeID)
			return r.persistenceService.SetAsSucceeded(operationID)
		})
	}
	close(finished)
}

func (r *service) startDeprovisioning(operationID, runtimeID string, config model.RuntimeConfig, secretName string, finished chan interface{}) {
	log.Infof("Deprovisioning runtime %s is starting", runtimeID)

	cluster, dberr := r.persistenceService.GetClusterData(runtimeID)

	if dberr != nil {
		updateOperationStatus(func() error {
			log.Errorf("Deprovisioning runtime %s failed: %s", runtimeID, dberr.Error())
			return r.persistenceService.SetAsFailed(operationID, dberr.Error())
		})
	}

	err := r.hydroform.DeprovisionCluster(config, secretName, cluster.TerraformState)

	if err != nil {
		updateOperationStatus(func() error {
			log.Errorf("Deprovisioning runtime %s failed: %s", runtimeID, err.Error())
			return r.persistenceService.SetAsFailed(operationID, err.Error())
		})
	} else {
		updateOperationStatus(func() error {
			log.Infof("Deprovisioning runtime %s finished successfully", runtimeID)
			return r.persistenceService.SetAsSucceeded(operationID)
		})
	}
	close(finished)
}

func updateOperationStatus(updateFunction func() error) {
	err := retry(interval, retryCount, func() error {
		return updateFunction()
	})
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
		time.Sleep(interval)
	}

	return err
}
