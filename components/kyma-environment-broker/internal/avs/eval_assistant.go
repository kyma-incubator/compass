package avs

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
)

type EvalAssistant interface {
	CreateBasicEvaluationRequest(operations internal.ProvisioningOperation, url string) (*BasicEvaluationCreateRequest, error)
	AppendOverrides(inputCreator internal.ProvisionInputCreator, evaluationId int64)
	IsAlreadyCreated(lifecycleData internal.AvsLifecycleData) bool
	SetEvalId(lifecycleData *internal.AvsLifecycleData, evalId int64)
	IsAlreadyDeleted(lifecycleData internal.AvsLifecycleData) bool
	GetEvaluationId(lifecycleData internal.AvsLifecycleData) int64
	markDeleted(lifecycleData *internal.AvsLifecycleData)
}
