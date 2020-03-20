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
}

func NewExternalEvalCreator(config avs.Config, delegator *avs.Delegator) *ExternalEvalCreator {
	return &ExternalEvalCreator{
		delegator: delegator,
		assistant: avs.NewExternalEvalAssistant(config),
		logger:    logrus.New(),
	}
}

func (eec *ExternalEvalCreator) createEval(operation internal.ProvisioningOperation, url string) (internal.ProvisioningOperation, time.Duration, error) {
	return eec.delegator.DoRun(eec.logger, operation, eec.assistant, url)
}
