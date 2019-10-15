package api

import (
	"context"

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

func (r *Resolver) ProvisionRuntime(ctx context.Context, id string, config *gqlschema.ProvisionRuntimeInput) (string, error) {
	operationID, err, _ := r.provisioning.ProvisionRuntime(id, config)
	return operationID, err
}

func (r *Resolver) DeprovisionRuntime(ctx context.Context, id string) (string, error) {
	operationID, err, _ := r.provisioning.DeprovisionRuntime(id)
	return operationID, err
}

func (r *Resolver) UpgradeRuntime(ctx context.Context, id string, config *gqlschema.UpgradeRuntimeInput) (string, error) {
	return "", nil
}

func (r *Resolver) ReconnectRuntimeAgent(ctx context.Context, id string) (string, error) {
	return "", nil
}

func (r *Resolver) RuntimeStatus(ctx context.Context, runtimeID string) (*gqlschema.RuntimeStatus, error) {
	return r.provisioning.RuntimeStatus(runtimeID)
}

func (r *Resolver) RuntimeOperationStatus(ctx context.Context, operationID string) (*gqlschema.OperationStatus, error) {
	return r.provisioning.RuntimeOperationStatus(operationID)
}

func (r *Resolver) CleanupRuntimeData(ctx context.Context, id string) (string, error) {
	return r.provisioning.CleanupRuntimeData(id)
}
