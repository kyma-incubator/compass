package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/common/log"

	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-incubator/hydroform/types"
)

const (
	interval = 10 * time.Second
	timeout  = 100 * time.Second
)

type ProvisioningService interface {
	ProvisionRuntime(id string, config *gqlschema.ProvisionRuntimeInput) (string, error, chan interface{})
	UpgradeRuntime(id string, config *gqlschema.UpgradeRuntimeInput) (string, error)
	DeprovisionRuntime(id string) (string, error, chan interface{})
	CleanupRuntimeData(id string) (string, error)
	ReconnectRuntimeAgent(id string) (string, error)
	RuntimeStatus(id string) (*gqlschema.RuntimeStatus, error)
	RuntimeOperationStatus(id string) (*gqlschema.OperationStatus, error)
}

type service struct {
	operationService persistence.OperationService
	runtimeService   persistence.RuntimeService
	hydroform        hydroform.Client
}

func NewProvisioningService(operationService persistence.OperationService, runtimeService persistence.RuntimeService, hydroform hydroform.Client) ProvisioningService {
	return &service{
		operationService: operationService,
		runtimeService:   runtimeService,
		hydroform:        hydroform,
	}
}

func (r *service) ProvisionRuntime(id string, config *gqlschema.ProvisionRuntimeInput) (string, error, chan interface{}) {
	_, err := r.runtimeService.GetLastOperation(id)

	if err == nil {
		return "", errors.New(fmt.Sprintf("cannot provision runtime. Runtime %s already provisioned", id)), nil
	}

	if err.Code() != dberrors.CodeNotFound {
		return "", err, nil
	}

	runtimeConfig := model.RuntimeConfigFromInput(*config)

	operation, err := r.runtimeService.SetProvisioningStarted(id, runtimeConfig)

	if err != nil {
		return "", err, nil
	}

	finished := make(chan interface{})

	go r.startProvisioning(operation.OperationID, runtimeConfig, config.Credentials.SecretName, finished)

	return operation.OperationID, nil, finished
}

func (r *service) DeprovisionRuntime(id string) (string, error, chan interface{}) {
	runtimeStatus, err := r.runtimeService.GetStatus(id)

	if err != nil {
		return "", err, nil
	}

	if runtimeStatus.LastOperationStatus.State == model.InProgress {
		return "", errors.New("cannot start new operation while previous one is in progress"), nil
	}

	operation, err := r.runtimeService.SetDeprovisioningStarted(id)

	if err != nil {
		return "", err, nil
	}

	finished := make(chan interface{})

	//TODO Decide how to pass credentials
	go r.startDeprovisioning(operation.OperationID, runtimeStatus.RuntimeConfiguration, "", finished)

	return operation.OperationID, nil, finished
}

func (r *service) UpgradeRuntime(id string, config *gqlschema.UpgradeRuntimeInput) (string, error) {
	return "", nil
}

func (r *service) ReconnectRuntimeAgent(id string) (string, error) {
	return "", nil
}

func (r *service) RuntimeStatus(runtimeID string) (*gqlschema.RuntimeStatus, error) {
	return nil, nil
}

func (r *service) RuntimeOperationStatus(operationID string) (*gqlschema.OperationStatus, error) {
	operation, err := r.operationService.Get(operationID)

	if err != nil {
		return nil, err
	}

	status := model.OperationStatusToGQLOperationStatus(operation)

	return status, nil
}

func (r *service) CleanupRuntimeData(id string) (string, error) {
	return "", nil
}

func (r *service) startProvisioning(operationID string, config model.RuntimeConfig, secretName string, finished chan interface{}) {
	status, err := r.hydroform.ProvisionCluster(config, secretName)

	if err != nil || status.Phase != types.Provisioned {
		updateOperationStatus(func() error {
			return r.operationService.SetAsFailed(operationID, err.Error())
		})
	} else {
		updateOperationStatus(func() error {
			return r.operationService.SetAsSucceeded(operationID)
		})
	}
	close(finished)
}

func (r *service) startDeprovisioning(operationID string, config model.RuntimeConfig, secretName string, finished chan interface{}) {
	err := r.hydroform.DeprovisionCluster(config, secretName)

	if err != nil {
		updateOperationStatus(func() error {
			return r.operationService.SetAsFailed(operationID, err.Error())
		})
	} else {
		updateOperationStatus(func() error {
			return r.operationService.SetAsSucceeded(operationID)
		})
	}
	close(finished)
}

func updateOperationStatus(updateFunction func() error) {
	err := waitForFunction(interval, timeout, func() error {
		return updateFunction()
	})
	if err != nil {
		log.Errorf("failed to set operation status, %s", err.Error())
	}
}

func waitForFunction(interval, timeout time.Duration, isDone func() error) error {
	done := time.After(timeout)

	for {
		err := isDone()
		if err == nil {
			return nil
		}

		select {
		case <-done:
			return err
		default:
			time.Sleep(interval)
		}
	}
}
