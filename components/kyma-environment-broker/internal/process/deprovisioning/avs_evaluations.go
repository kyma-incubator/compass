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
	if deProvisioningOperation.AVSExternalEvaluationDeleted && deProvisioningOperation.AVSInternalEvaluationDeleted {
		logger.Infof("Both internal and external evaluations have been deleted")
		return deProvisioningOperation, 0, nil
	}

	provisioningOperation, err := ars.operationsStorage.GetProvisioningOperationByInstanceID(deProvisioningOperation.InstanceID)
	if err != nil {
		logger.Errorf("error while getting provisioning deProvisioningOperation from storage")
		return deProvisioningOperation, time.Second * 10, nil
	}

	deProvisioningOperation, duration, err := ars.delegator.DeleteAvsEvaluation(deProvisioningOperation, provisioningOperation, logger, ars.internalEvalAssistant)
	if duration != 0 {
		return deProvisioningOperation, duration, nil
	}

	return ars.delegator.DeleteAvsEvaluation(deProvisioningOperation, provisioningOperation, logger, ars.externalEvalAssistant)

}
