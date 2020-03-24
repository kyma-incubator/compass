package avs

import "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"

const externalEvalCheckType = "HTTPSGET"

type ExternalEvalAssistant struct {
	avsConfig Config
}

func NewExternalEvalAssistant(avsConfig Config) *ExternalEvalAssistant {
	return &ExternalEvalAssistant{avsConfig: avsConfig}
}

func (eea *ExternalEvalAssistant) CreateBasicEvaluationRequest(operations internal.ProvisioningOperation, url string) (*BasicEvaluationCreateRequest, error) {
	return newBasicEvaluationCreateRequest(operations, eea, eea.avsConfig.GroupId, url)
}

func (eea *ExternalEvalAssistant) AppendOverrides(inputCreator internal.ProvisionInputCreator, evaluationId int64) {
	//do nothing
}

func (eea *ExternalEvalAssistant) CheckIfAlreadyDone(operation internal.ProvisioningOperation) bool {
	return operation.AVSEvaluationExternalId != 0
}

func (eea *ExternalEvalAssistant) ProvideSuffix() string {
	return "external"
}

func (eea *ExternalEvalAssistant) ProvideTesterAccessId() int64 {
	return eea.avsConfig.ExternalTesterAccessId
}

func (eea *ExternalEvalAssistant) SetEvalId(operation *internal.ProvisioningOperation, evalId int64) {
	operation.AVSEvaluationExternalId = evalId
}

func (eea *ExternalEvalAssistant) ProvideCheckType() string {
	return externalEvalCheckType
}
