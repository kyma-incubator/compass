package deprovisioning

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"

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
	deProvisioningManager *process.DeprovisionOperationManager
}

func NewAvsEvaluationsRemovalStep(delegator *avs.Delegator, operationsStorage storage.Operations, externalEvalAssistant, internalEvalAssistant avs.EvalAssistant) *AvsEvaluationRemovalStep {
	return &AvsEvaluationRemovalStep{
		delegator:             delegator,
		operationsStorage:     operationsStorage,
		externalEvalAssistant: externalEvalAssistant,
		internalEvalAssistant: internalEvalAssistant,
		deProvisioningManager: process.NewDeprovisionOperationManager(operationsStorage),
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

	deProvisioningOperation, err := ars.delegator.DeleteAvsEvaluation(deProvisioningOperation, logger, ars.internalEvalAssistant)
	if err != nil {
		return ars.deProvisioningManager.RetryOperation(deProvisioningOperation, err.Error(), 10*time.Second, 1*time.Hour, logger)
	}

	deProvisioningOperation, err = ars.delegator.DeleteAvsEvaluation(deProvisioningOperation, logger, ars.externalEvalAssistant)
	if err != nil {
		return ars.deProvisioningManager.RetryOperation(deProvisioningOperation, err.Error(), 10*time.Second, 1*time.Hour, logger)
	}
	return deProvisioningOperation, 0, nil

}
