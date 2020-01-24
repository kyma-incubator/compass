package api

import (
	"context"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

type Resolver struct {
	provisioning provisioning.Service
}

func (r *Resolver) Mutation() gqlschema.MutationResolver {
	return &Resolver{
		provisioning: r.provisioning,
	}
}
func (r *Resolver) Query() gqlschema.QueryResolver {
	return &Resolver{
		provisioning: r.provisioning,
	}
}

func NewResolver(provisioningService provisioning.Service) *Resolver {
	return &Resolver{
		provisioning: provisioningService,
	}
}

func (r *Resolver) ProvisionRuntime(ctx context.Context, config gqlschema.ProvisionRuntimeInput) (*gqlschema.OperationStatus, error) {
	log.Infof("Requested provisioning of RUNTIME %s.", config.RuntimeInput.Name)
	err := validateInput(config)
	if err != nil {
		log.Errorf("Failed to provision Runtime %s: %s", config.RuntimeInput.Name, err)
		return nil, err
	}

	log.Infof("Requested provisioning of Runtime %s.", config.RuntimeInput.Name)
	if config.ClusterConfig.GcpConfig != nil && config.ClusterConfig.GardenerConfig == nil {
		err := fmt.Errorf("Provisioning on GCP is currently not supported, Runtime : %s", config.RuntimeInput.Name)
		strError := err.Error()
		log.Errorf(strError)
		return nil, err
	}

	operationID, runtimeID, _, err := r.provisioning.ProvisionRuntime(config)
	if err != nil {
		log.Errorf("Failed to provision Runtime %s: %s", config.RuntimeInput.Name, err)
		return nil, err
	}
	log.Infof("Provisioning started for Runtime %s. Operation id %s", config.RuntimeInput.Name, operationID)

	messageProvisioningStarted := "Provisioning started"

	return &gqlschema.OperationStatus{
		ID:        &operationID,
		Operation: gqlschema.OperationTypeProvision,
		Message:   &messageProvisioningStarted,
		RuntimeID: &runtimeID,
	}, nil
}

func validateInput(config gqlschema.ProvisionRuntimeInput) error {
	if len(config.KymaConfig.Components) == 0 {
		return errors.New("cannot provision Runtime since Kyma components list is empty")
	}

	if config.RuntimeInput == nil {
		return errors.New("cannot provision Runtime since runtime input is missing")
	}

	if config.Credentials == nil {
		return errors.New("cannot provision Runtime since credentials are missing")
	}

	return nil
}

func (r *Resolver) DeprovisionRuntime(ctx context.Context, id string) (string, error) {
	log.Infof("Requested deprovisioning of Runtime %s.", id)

	operationID, _, err := r.provisioning.DeprovisionRuntime(id)
	if err != nil {
		log.Errorf("Failed to provision Runtime %s: %s", id, err)
		return "", err
	}
	log.Infof("Deprovisioning started for Runtime %s. Operation id %s", id, operationID)

	return operationID, nil
}

func (r *Resolver) UpgradeRuntime(ctx context.Context, id string, config gqlschema.UpgradeRuntimeInput) (string, error) {
	return "", nil
}

func (r *Resolver) ReconnectRuntimeAgent(ctx context.Context, id string) (string, error) {
	return "", nil
}

func (r *Resolver) RuntimeStatus(ctx context.Context, runtimeID string) (*gqlschema.RuntimeStatus, error) {
	log.Infof("Requested to get status for Runtime %s.", runtimeID)

	status, err := r.provisioning.RuntimeStatus(runtimeID)
	if err != nil {
		log.Errorf("Failed to get status for Runtime %s: %s", runtimeID, err)
		return nil, err
	}
	log.Infof("Getting status for Runtime %s succeeded.", runtimeID)

	return status, nil
}

func (r *Resolver) RuntimeOperationStatus(ctx context.Context, operationID string) (*gqlschema.OperationStatus, error) {
	log.Infof("Requested to get Runtime operation status for Operation %s.", operationID)

	status, err := r.provisioning.RuntimeOperationStatus(operationID)
	if err != nil {
		log.Errorf("Failed to get Runtime operation status: %s Operation ID: %s", err, operationID)
		return nil, err
	}
	log.Infof("Getting Runtime operation status for Operation %s succeeded.", operationID)

	return status, nil
}
