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
	iec       *avs.InternalEvalConfigurator
}

func NewInternalEvaluationStep(avsConfig avs.Config, operationsStorage storage.Operations) *InternalEvaluationStep {
	return &InternalEvaluationStep{
		delegator: avs.NewDelegator(avsConfig, operationsStorage),
		iec:       avs.NewInternalEvalConfigurator(avsConfig.ApiKey),
	}
}

func (ies *InternalEvaluationStep) Name() string {
	return "AVS_Configuration_Step"
}

func (ies *InternalEvaluationStep) Run(operation internal.ProvisioningOperation, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	return ies.delegator.DoRun(logger, operation, ies.iec)
}
