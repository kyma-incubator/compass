package provisioning

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/avs"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/sirupsen/logrus"
)

type InternalEvaluationStep struct {
	delegator *avs.Delegator
}

func NewInternalEvaluationStep(avsConfig avs.Config, operationsStorage storage.Operations) *InternalEvaluationStep {
	return &InternalEvaluationStep{
		delegator: avs.NewDelegator(avsConfig, operationsStorage),
	}
}

func (ies *InternalEvaluationStep) Name() string {
	return "AVS_Configuration_Step"
}

func (ies *InternalEvaluationStep) Run(operation internal.ProvisioningOperation, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	return ies.delegator.DoRun(logger, operation, createInternalBasicEvaluationRequest)
}

func createInternalBasicEvaluationRequest(operations internal.ProvisioningOperation) (*avs.BasicEvaluationCreateRequest, error) {
	return avs.NewInternalBasicEvaluation(operations)
}
