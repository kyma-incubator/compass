package avs

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/sirupsen/logrus"
)

type InternalEvaluationStep struct {
	delegator *delegator
}

func NewInternalEvaluationStep(avsConfig Config, operationsStorage storage.Operations) *InternalEvaluationStep {
	return &InternalEvaluationStep{
		delegator: newDelegator(avsConfig, operationsStorage),
	}
}

func (ies *InternalEvaluationStep) Name() string {
	return "AVS_Configuration_Step"
}

func (ies *InternalEvaluationStep) Run(operation internal.ProvisioningOperation, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	stepName := ies.Name()
	return ies.delegator.doRun(logger, stepName, operation, createInternalBasicEvaluationRequest)
}

func createInternalBasicEvaluationRequest(operations internal.ProvisioningOperation) (*basicEvaluationCreateRequest, error) {
	return newInternalBasicEvaluation(operations)
}
