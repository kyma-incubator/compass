package avs

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
)

type EvalAssistant interface {
	CreateBasicEvaluationRequest(operations internal.ProvisioningOperation, url string) (*BasicEvaluationCreateRequest, error)
	SetOverrides(inputCreator internal.ProvisionInputCreator, evaluationId int64)
	CheckIfAlreadyDone(operation internal.ProvisioningOperation) bool
	SetEvalId(operation *internal.ProvisioningOperation, evalId int64)
}
