package avs

import (
	"time"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal"
)

type EvalAssistant interface {
	CreateBasicEvaluationRequest(operations internal.ProvisioningOperation, configForModel *configForModel, url string) (*BasicEvaluationCreateRequest, error)
	AppendOverrides(inputCreator internal.ProvisionInputCreator, evaluationId int64)
	IsAlreadyCreated(lifecycleData internal.AvsLifecycleData) bool
	SetEvalId(lifecycleData *internal.AvsLifecycleData, evalId int64)
	IsAlreadyDeleted(lifecycleData internal.AvsLifecycleData) bool
	GetEvaluationId(lifecycleData internal.AvsLifecycleData) int64
	markDeleted(lifecycleData *internal.AvsLifecycleData)
	provideRetryConfig() *RetryConfig
}

type RetryConfig struct {
	retryInterval time.Duration
	maxTime       time.Duration
}
