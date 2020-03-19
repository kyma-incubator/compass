package avs

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
)

type EvalConfigurator interface {
	CreateInternalBasicEvaluationRequest(operations internal.ProvisioningOperation) (*BasicEvaluationCreateRequest, error)
	SetOverrides(inputCreator internal.ProvisionInputCreator, evaluationId int64)
	CheckIfAlreadyDone(operation internal.ProvisioningOperation) bool
}
