package api

import (
	"context"
	"errors"

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

func (r *Resolver) ProvisionRuntime(ctx context.Context, id string, config gqlschema.ProvisionRuntimeInput) (string, error) {
	err := validateInput(config)
	if err != nil {
		log.Errorf("Failed to provision Runtime %s: %s", id, err)
		return "", err
	}

	log.Infof("Requested provisioning of Runtime %s.", id)

	operationID, _, err := r.provisioning.ProvisionRuntime(id, config)
	if err != nil {
		log.Errorf("Failed to provision Runtime %s: %s", id, err)
		return "", err
	}
	log.Infof("Provisioning stared for Runtime %s. Operation id %s", id, operationID)

	return operationID, nil
}

func validateInput(config gqlschema.ProvisionRuntimeInput) error {
	if len(config.KymaConfig.Modules) == 0 {
		return errors.New("cannot provision Runtime since Kyma modules list is empty")
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

func (r *Resolver) CleanupRuntimeData(ctx context.Context, id string) (*gqlschema.CleanUpRuntimeStatus, error) {
	log.Infof("Requested cleaning up Runtime data for Runtime %s.", id)

	res, err := r.provisioning.CleanupRuntimeData(id)
	if err != nil {
		log.Errorf("Failed to cleanup data for Runtime %s: %s", id, err)
		return nil, err
	}
	log.Infof("Cleaning up Runtime data for Runtime %s succeeded.", id)

	return res, nil
}
