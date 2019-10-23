package api

import (
	"context"
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

type Resolver struct {
	provisioning provisioning.ProvisioningService
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

func NewResolver(provisioningService provisioning.ProvisioningService) *Resolver {
	return &Resolver{
		provisioning: provisioningService,
	}
}

func (r *Resolver) ProvisionRuntime(ctx context.Context, id string, config gqlschema.ProvisionRuntimeInput) (string, error) {
	err := validateInput(config)
	if err != nil {
		log.Errorf("Failed to provision runtime: %s", err)
		return "", err
	}

	operationID, err, _ := r.provisioning.ProvisionRuntime(id, config)
	if err != nil {
		log.Errorf("Failed to provision runtime: %s", err)
	}

	return operationID, err
}

func validateInput(config gqlschema.ProvisionRuntimeInput) error {
	if len(config.KymaConfig.Modules) == 0 {
		return errors.New("cannot provision runtime since Kyma modules list is empty")
	}

	return nil
}

func (r *Resolver) DeprovisionRuntime(ctx context.Context, id string, credentials gqlschema.CredentialsInput) (string, error) {
	operationID, err, _ := r.provisioning.DeprovisionRuntime(id, credentials)
	if err != nil {
		log.Errorf("Failed to deprovision runtime: %s", err)
	}

	return operationID, err
}

func (r *Resolver) UpgradeRuntime(ctx context.Context, id string, config gqlschema.UpgradeRuntimeInput) (string, error) {
	return "", nil
}

func (r *Resolver) ReconnectRuntimeAgent(ctx context.Context, id string) (string, error) {
	return "", nil
}

func (r *Resolver) RuntimeStatus(ctx context.Context, runtimeID string) (*gqlschema.RuntimeStatus, error) {
	status, err := r.provisioning.RuntimeStatus(runtimeID)
	if err != nil {
		log.Errorf("Failed to get runtime status: %s", err)
	}

	return status, err
}

func (r *Resolver) RuntimeOperationStatus(ctx context.Context, operationID string) (*gqlschema.OperationStatus, error) {
	status, err := r.provisioning.RuntimeOperationStatus(operationID)
	if err != nil {
		log.Errorf("Failed to get runtime operation status: %s", err)
	}

	return status, err
}

func (r *Resolver) CleanupRuntimeData(ctx context.Context, id string) (string, error) {
	res, err := r.provisioning.CleanupRuntimeData(id)
	if err != nil {
		log.Errorf("Failed to cleanup runtime data: %s", err)
	}

	return res, err
}
