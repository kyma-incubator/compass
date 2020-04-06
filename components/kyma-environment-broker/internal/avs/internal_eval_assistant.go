package avs

import (
	"strconv"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

const (
	evaluationIdKey = "avs_bridge.config.evaluations.cluster.id"
	avsBridgeAPIKey = "avs_bridge.config.availabilityService.apiKey"
)

type InternalEvalAssistant struct {
	avsConfig Config
}

func NewInternalEvalAssistant(avsConfig Config) *InternalEvalAssistant {
	return &InternalEvalAssistant{avsConfig: avsConfig}
}

func (iec *InternalEvalAssistant) CreateBasicEvaluationRequest(operations internal.ProvisioningOperation, url string) (*BasicEvaluationCreateRequest, error) {
	return newBasicEvaluationCreateRequest(operations, iec, iec.avsConfig.GroupId, url)
}

func (iec *InternalEvalAssistant) AppendOverrides(inputCreator internal.ProvisionInputCreator, evaluationId int64) {
	inputCreator.AppendOverrides("avs-bridge", []*gqlschema.ConfigEntryInput{
		{
			Key:   evaluationIdKey,
			Value: strconv.FormatInt(evaluationId, 10),
		},
		{
			Key:   avsBridgeAPIKey,
			Value: iec.avsConfig.ApiKey,
		},
	})
}

func (iec *InternalEvalAssistant) CheckIfAlreadyDone(operation internal.ProvisioningOperation) bool {
	return operation.AvsEvaluationInternalId != 0
}

func (iec *InternalEvalAssistant) ProvideSuffix() string {
	return "int"
}

func (iec *InternalEvalAssistant) ProvideTesterAccessId() int64 {
	return iec.avsConfig.InternalTesterAccessId
}

func (iec *InternalEvalAssistant) ProvideCheckType() string {
	return ""
}

func (iec *InternalEvalAssistant) SetEvalId(operation *internal.ProvisioningOperation, evalId int64) {
	operation.AvsEvaluationInternalId = evalId
}
