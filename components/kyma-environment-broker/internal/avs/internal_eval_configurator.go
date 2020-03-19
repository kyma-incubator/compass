package avs

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"strconv"
)

const (
	evaluationIdKey = "avs_bridge.config.evaluations.cluster.id"
	avsBridgeAPIKey = "avs_bridge.config.availabilityService.apiKey"
)

type InternalEvalConfigurator struct {
	apiKey string
}

func NewInternalEvalConfigurator(apiKey string) *InternalEvalConfigurator {
	return &InternalEvalConfigurator{apiKey: apiKey}
}

func (iec *InternalEvalConfigurator) CreateInternalBasicEvaluationRequest(operations internal.ProvisioningOperation) (*BasicEvaluationCreateRequest, error) {
	return NewInternalBasicEvaluation(operations)
}

func (iec *InternalEvalConfigurator) SetOverrides(inputCreator internal.ProvisionInputCreator, evaluationId int64) {
	inputCreator.SetOverrides("avs-bridge", []*gqlschema.ConfigEntryInput{
		{
			Key:   evaluationIdKey,
			Value: strconv.FormatInt(evaluationId, 10),
		},
		{
			Key:   avsBridgeAPIKey,
			Value: iec.apiKey,
		},
	})
}

func (iec *InternalEvalConfigurator) CheckIfAlreadyDone(operation internal.ProvisioningOperation) bool {
	return operation.AvsEvaluationInternalId != 0
}
