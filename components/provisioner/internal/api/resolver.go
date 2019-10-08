package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
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
	_, err := r.runtimeService.GetLastOperation(id)

	if err == nil {
		return "", errors.New(fmt.Sprintf("cannot provision runtime. Runtime %s already provisioned", id))
	}

	if err.Code() != dberrors.CodeNotFound {
		return "", err
	}

	runtimeConfig := model.RuntimeConfigFromInput(*config)

	operation, err := r.runtimeService.SetProvisioningStarted(id, runtimeConfig)

	if err != nil {
		return "", err
	}

	go r.startProvisioning(operation.OperationID, runtimeConfig, config.Credentials.SecretName)

	return operation.OperationID, nil
}

func (r *Resolver) DeprovisionRuntime(ctx context.Context, id string) (string, error) {
	runtimeStatus, err := r.runtimeService.GetStatus(id)

	if err != nil {
		return "", err
	}

	if runtimeStatus.LastOperationStatus.State == model.InProgress {
		return "", errors.New("cannot start new operation while previous one is in progress")
	}

	operation, err := r.runtimeService.SetDeprovisioningStarted(id)

	if err != nil {
		return "", err
	}

	//TODO Decide how to pass credentials
	go r.startDeprovisioning(operation.OperationID, runtimeStatus.RuntimeConfiguration, "")

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

	status := model.OperationStatusToGQLOperationStatus(operation)

	return status, nil
}

func (r *Resolver) startProvisioning(operationID string, config model.RuntimeConfig, secretName string) {
	status, err := r.hydroform.ProvisionCluster(config, secretName)

	if err != nil || status.Phase != types.Provisioned {
		r.operationService.SetAsFailed(operationID, err.Error())
		return
	}

	r.operationService.SetAsSucceeded(operationID)
}

func (r *Resolver) startDeprovisioning(operationID string, config model.RuntimeConfig, secretName string) {
	err := r.hydroform.DeprovisionCluster(config, secretName)

	if err != nil {
		r.operationService.SetAsFailed(operationID, err.Error())
		return
	}

	r.operationService.SetAsSucceeded(operationID)
}

func (r *Resolver) CleanupRuntimeData(ctx context.Context, id string) (string, error) {
	return "", nil
}
