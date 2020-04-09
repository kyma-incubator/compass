package deprovisioning

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/avs"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/sirupsen/logrus"
)

type AvsEvaluationRemovalStep struct {
	delegator             *avs.Delegator
	operationsStorage     storage.Operations
	externalEvalAssistant avs.EvalAssistant
	internalEvalAssistant avs.EvalAssistant
}

func NewAvsEvaluationsRemovalStep(delegator *avs.Delegator, operationsStorage storage.Operations, externalEvalAssistant, internalEvalAssistant avs.EvalAssistant) *AvsEvaluationRemovalStep {
	return &AvsEvaluationRemovalStep{
		delegator:             delegator,
		operationsStorage:     operationsStorage,
		externalEvalAssistant: externalEvalAssistant,
		internalEvalAssistant: internalEvalAssistant,
	}
}

func (ars *AvsEvaluationRemovalStep) Name() string {
	return "De-provision_AVS_Evaluations"
}

func (ars *AvsEvaluationRemovalStep) Run(deProvisioningOperation internal.DeprovisioningOperation, logger logrus.FieldLogger) (internal.DeprovisioningOperation, time.Duration, error) {
	if deProvisioningOperation.Avs.AVSExternalEvaluationDeleted && deProvisioningOperation.Avs.AVSInternalEvaluationDeleted {
		logger.Infof("Both internal and external evaluations have been deleted")
		return deProvisioningOperation, 0, nil
	}

	deProvisioningOperation, duration, _ := ars.delegator.DeleteAvsEvaluation(deProvisioningOperation, logger, ars.internalEvalAssistant)
	if duration != 0 {
		return deProvisioningOperation, duration, nil
	}

	return ars.delegator.DeleteAvsEvaluation(deProvisioningOperation, logger, ars.externalEvalAssistant)

}
