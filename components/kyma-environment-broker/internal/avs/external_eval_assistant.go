package avs

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
)

const externalEvalCheckType = "HTTPSGET"

type ExternalEvalAssistant struct {
	avsConfig   Config
	retryConfig *RetryConfig
}

func NewExternalEvalAssistant(avsConfig Config) *ExternalEvalAssistant {
	return &ExternalEvalAssistant{
		avsConfig:   avsConfig,
		retryConfig: &RetryConfig{maxTime: 90 * time.Minute, retryInterval: 1 * time.Minute},
	}
}

func (eea *ExternalEvalAssistant) CreateBasicEvaluationRequest(operations internal.ProvisioningOperation, url string) (*BasicEvaluationCreateRequest, error) {
	return newBasicEvaluationCreateRequest(operations, eea, eea.avsConfig.GroupId, url)
}

func (eea *ExternalEvalAssistant) AppendOverrides(inputCreator internal.ProvisionInputCreator, evaluationId int64) {
	//do nothing
}

func (eea *ExternalEvalAssistant) IsAlreadyCreated(lifecycleData internal.AvsLifecycleData) bool {
	return lifecycleData.AVSEvaluationExternalId != 0
}

func (eea *ExternalEvalAssistant) ProvideSuffix() string {
	return "ext"
}

func (eea *ExternalEvalAssistant) ProvideTesterAccessId() int64 {
	return eea.avsConfig.ExternalTesterAccessId
}

func (eea *ExternalEvalAssistant) SetEvalId(lifecycleData *internal.AvsLifecycleData, evalId int64) {
	lifecycleData.AVSEvaluationExternalId = evalId
}

func (eea *ExternalEvalAssistant) ProvideCheckType() string {
	return externalEvalCheckType
}

func (eea *ExternalEvalAssistant) IsAlreadyDeleted(lifecycleData internal.AvsLifecycleData) bool {
	return lifecycleData.AVSExternalEvaluationDeleted
}
func (eea *ExternalEvalAssistant) GetEvaluationId(lifecycleData internal.AvsLifecycleData) int64 {
	return lifecycleData.AVSEvaluationExternalId
}

func (eea *ExternalEvalAssistant) markDeleted(lifecycleData *internal.AvsLifecycleData) {
	lifecycleData.AVSExternalEvaluationDeleted = true
}

func (eea *ExternalEvalAssistant) provideRetryConfig() *RetryConfig {
	return eea.retryConfig
}
