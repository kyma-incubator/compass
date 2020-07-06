package provisioning

import (
	"time"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/avs"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal"
	"github.com/sirupsen/logrus"
)

type InternalEvaluationStep struct {
	delegator *avs.Delegator
	iec       *avs.InternalEvalAssistant
}

func NewInternalEvaluationStep(delegator *avs.Delegator, assistant *avs.InternalEvalAssistant) *InternalEvaluationStep {
	return &InternalEvaluationStep{
		delegator: delegator,
		iec:       assistant,
	}
}

func (ies *InternalEvaluationStep) Name() string {
	return "AVS_Create_Internal_Eval_Step"
}

func (ies *InternalEvaluationStep) Run(operation internal.ProvisioningOperation, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	return ies.delegator.CreateEvaluation(logger, operation, ies.iec, "")
}
