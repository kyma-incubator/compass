package provisioning

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/avs"
	"github.com/sirupsen/logrus"
)

type ExternalEvalCreator struct {
	delegator *avs.Delegator
	assistant *avs.ExternalEvalAssistant
	logger    logrus.FieldLogger
	disabled  bool
}

func NewExternalEvalCreator(config avs.Config, delegator *avs.Delegator, disabled bool, assistant *avs.ExternalEvalAssistant) *ExternalEvalCreator {
	return &ExternalEvalCreator{
		delegator: delegator,
		assistant: assistant,
		logger:    logrus.New(),
		disabled:  disabled,
	}
}

func (eec *ExternalEvalCreator) createEval(operation internal.ProvisioningOperation, url string) (internal.ProvisioningOperation, time.Duration, error) {
	if eec.disabled {
		return operation, 0, nil
	} else {
		return eec.delegator.CreateEvaluation(eec.logger, operation, eec.assistant, url)
	}
}
