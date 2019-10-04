package api

import (
	"context"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-incubator/hydroform/types"
)

type Resolver struct {
	operationService persistence.OperationService
	runtimeService   persistence.RuntimeService
	hydroform        hydroform.Client
}

func (r *Resolver) Mutation() gqlschema.MutationResolver {
	return &Resolver{
		operationService: r.operationService,
		runtimeService:   r.runtimeService,
		hydroform:        r.hydroform,
	}
}
func (r *Resolver) Query() gqlschema.QueryResolver {
	return &Resolver{
		operationService: r.operationService,
		runtimeService:   r.runtimeService,
		hydroform:        r.hydroform,
	}
}

func NewResolver(operationService persistence.OperationService, runtimeService persistence.RuntimeService, hydroform hydroform.Client) *Resolver {
	return &Resolver{
		operationService: operationService,
		runtimeService:   runtimeService,
		hydroform:        hydroform,
	}
}

func (r *Resolver) ProvisionRuntime(ctx context.Context, id string, config *gqlschema.ProvisionRuntimeInput) (string, error) {
	clusterConfig := model.ClusterConfigFromInput(*config.ClusterConfig)
	kymaConfig := model.KymaConfigFromInput(*config.KymaConfig)
	operation, err := r.runtimeService.SetProvisioningStarted(id, clusterConfig, kymaConfig)

	if err != nil {
		return "", err
	}

	go r.startProvisioning(operation.OperationID, clusterConfig)

	return operation.OperationID, nil
}

func (r *Resolver) DeprovisionRuntime(ctx context.Context, id string) (string, error) {
	runtimeStatus, err := r.runtimeService.GetStatus(id)

	if err != nil {
		return "", err
	}

	operation, err := r.runtimeService.SetDeprovisioningStarted(id)

	if err != nil {
		return "", err
	}

	go r.startDeprovisioning(operation.OperationID, runtimeStatus.RuntimeConfiguration.ClusterConfig)

	return operation.OperationID, nil
}

func (r *Resolver) UpgradeRuntime(ctx context.Context, id string, config *gqlschema.UpgradeRuntimeInput) (string, error) {
	return "", nil
}

func (r *Resolver) ReconnectRuntimeAgent(ctx context.Context, id string) (string, error) {
	return "", nil
}

func (r *Resolver) RuntimeStatus(ctx context.Context, runtimeID string) (*gqlschema.RuntimeStatus, error) {
	return nil, nil
}

func (r *Resolver) RuntimeOperationStatus(ctx context.Context, operationID string) (*gqlschema.OperationStatus, error) {
	operation, err := r.operationService.Get(operationID)

	if err != nil {
		return &gqlschema.OperationStatus{}, err
	}

	_ = operation

	return nil, nil
}

func (r *Resolver) startProvisioning(operationID string, config model.ClusterConfig) {
	status, err := r.hydroform.ProvisionCluster(config)

	if err != nil || status.Phase != types.Provisioned {
		r.operationService.SetAsFailed(operationID, err.Error())
		return
	}

	r.operationService.SetAsSucceeded(operationID)
}

func (r *Resolver) startDeprovisioning(operationID string, config model.ClusterConfig) {
	err := r.hydroform.DeprovisionCluster(config)

	if err != nil {
		r.operationService.SetAsFailed(operationID, err.Error())
		return
	}

	r.operationService.SetAsSucceeded(operationID)
}
