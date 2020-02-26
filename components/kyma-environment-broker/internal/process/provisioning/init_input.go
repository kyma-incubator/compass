package provisioning

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/input"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/sirupsen/logrus"
)

type InputInitialisationStep struct {
	operationManager *process.OperationManager
	directorURL      string
	inputBuilder     input.CreatorForPlan
}

func NewInputInitialisationStep(os storage.Operations, b input.CreatorForPlan, dURL string) *InputInitialisationStep {
	return &InputInitialisationStep{
		operationManager: process.NewOperationManager(os),
		directorURL:      dURL,
		inputBuilder:     b,
	}
}

func (s *InputInitialisationStep) Name() string {
	return "Initialize runtime input request"
}

func (s *InputInitialisationStep) Run(operation internal.ProvisioningOperation, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	pp, err := operation.GetProvisioningParameters()
	if err != nil {
		log.Errorf("cannot fetch provisioning parameters from operation: %s", err)
		return s.operationManager.OperationFailed(operation, "invalid operation provisioning parameters")
	}

	log.Infof("create input creator for %q plan ID", pp.PlanID)
	creator, found := s.inputBuilder.ForPlan(pp.PlanID)
	if !found {
		log.Error("input creator does not exist")
		return s.operationManager.OperationFailed(operation, "cannot create provisioning input creator")
	}

	log.Info("set overrides for 'core' and 'compass-runtime-agent' components")
	creator.SetOverrides("core", []*gqlschema.ConfigEntryInput{
		{
			Key:   "console.managementPlane.url",
			Value: s.directorURL,
		},
	})
	creator.SetOverrides("compass-runtime-agent", []*gqlschema.ConfigEntryInput{
		{
			Key:   "managementPlane.url",
			Value: s.directorURL,
		},
	})

	operation.InputCreator = creator
	return operation, 0, nil
}
