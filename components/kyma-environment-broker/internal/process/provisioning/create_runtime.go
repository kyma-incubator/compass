package provisioning

import (
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Config holds all configurations connected with Provisioner API
type Config struct {
	URL             string
	GCPSecretName   string
	AzureSecretName string
	AWSSecretName   string
}

type CreateRuntimeStep struct {
	operationStorage  storage.Operations
	instanceStorage   storage.Instances
	builderFactory    InputBuilderForPlan
	provisioningCfg   Config
	provisionerClient provisioner.Client
}

func NewCreateRuntimeStep(os storage.Operations, is storage.Instances, builderFactory InputBuilderForPlan, cfg Config, cli provisioner.Client) *CreateRuntimeStep {
	return &CreateRuntimeStep{
		operationStorage:  os,
		instanceStorage:   is,
		builderFactory:    builderFactory,
		provisioningCfg:   cfg,
		provisionerClient: cli,
	}
}

func (s *CreateRuntimeStep) Run(operation *internal.ProvisioningOperation) (error, time.Duration) {
	// TODO make sure that the process has not already been activated
	// TODO mark operation as failed if error is permanent

	pp, err := operation.GetProvisioningParameters()
	if err != nil {
		return errors.Wrap(err, "while getting provisioning parameters from operation"), 0
	}
	rawParameters, err := json.Marshal(pp.Parameters)
	if err != nil {
		return errors.Wrap(err, "while marshaling instance parameters"), 0
	}

	input, err := s.createProvisionInput(pp)
	if err != nil {
		return errors.Wrap(err, "while creating provision input"), 0
	}

	// TODO save runtimeID in operation to not ask provisioner again (what happen when I call provisioner again?)
	resp, err := s.provisionerClient.ProvisionRuntime(pp.ErsContext.GlobalAccountID, input)
	if err != nil {
		log.Errorf("[%s] Call to provision failed: %s", operation.ID, err)
		return nil, 1 * time.Minute
	}
	if resp.RuntimeID == nil {
		// TODO: return error or time.Duration ?
		return nil, 10 * time.Minute
	}

	operation.ProvisionerOperationID = *resp.ID
	_, err = s.operationStorage.UpdateProvisioningOperation(*operation)
	if err != nil {
		return nil, 1 * time.Minute
	}

	err = s.instanceStorage.Insert(internal.Instance{
		InstanceID:      operation.InstanceID,
		GlobalAccountID: pp.ErsContext.GlobalAccountID,
		RuntimeID:       *resp.RuntimeID,
		ServiceID:       pp.ServiceID,
		ServicePlanID:   pp.PlanID,
		// TODO do we need RawProvisioningParameters if we have in operation?
		ProvisioningParameters: string(rawParameters),
	})
	if err != nil {
		return nil, 1 * time.Minute
	}

	return nil, 0
}

func (s *CreateRuntimeStep) createProvisionInput(parameters internal.ProvisioningParameters) (gqlschema.ProvisionRuntimeInput, error) {
	var input gqlschema.ProvisionRuntimeInput

	inputBuilder, found := s.builderFactory.ForPlan(parameters.PlanID)
	if !found {
		return input, errors.Errorf("while finding input builder: plan %q not found.", parameters.PlanID)
	}
	inputBuilder.
		SetERSContext(parameters.ErsContext).
		SetProvisioningParameters(parameters.Parameters).
		SetProvisioningConfig(s.provisioningCfg)

	input, err := inputBuilder.Build()
	if err != nil {
		return input, errors.Wrap(err, "while building input for provisioner")
	}

	return input, nil
}
