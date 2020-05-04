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
	disabled  bool
}

func NewExternalEvalCreator(delegator *avs.Delegator, disabled bool, assistant *avs.ExternalEvalAssistant) *ExternalEvalCreator {
	return &ExternalEvalCreator{
		delegator: delegator,
		assistant: assistant,
		disabled:  disabled,
	}
}

func (eec *ExternalEvalCreator) createEval(operation internal.ProvisioningOperation, url string, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	if eec.disabled {
		return operation, 0, nil
	} else {
		return eec.delegator.CreateEvaluation(logger, operation, eec.assistant, url)
	}
}
