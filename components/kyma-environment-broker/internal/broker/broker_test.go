package broker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	serviceID       = "47c9dcbf-ff30-448e-ab36-d3bad66ba281"
	planID          = "4deee563-e5ec-4731-b9b1-53b42d855f0c"
	globalAccountID = "e8f7ec0a-0cd6-41f0-905d-5d1efa9fb6c4"
)

func TestBroker_Services(t *testing.T) {
	// given
	memStorage := storage.NewMemoryStorage()

	optComponentsProviderMock := &automock.OptionalComponentNamesProvider{}
	defer optComponentsProviderMock.AssertExpectations(t)

	optComponentsNames := []string{"monitoring", "kiali", "loki", "jaeger"}
	optComponentsProviderMock.On("GetAllOptionalComponentsNames").Return(optComponentsNames)

	kymaEnvBroker, err := broker.New(broker.Config{EnablePlans: []string{"gcp", "azure"}}, nil, nil, broker.ProvisioningConfig{}, memStorage.Instances(), optComponentsProviderMock, nil, &broker.DumyDumper{})
	require.NoError(t, err)

	// when
	services, err := kymaEnvBroker.Services(context.TODO())

	// then
	require.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Len(t, services[0].Plans, 2)

	// assert provisioning schema
	componentItem := services[0].Plans[0].Schemas.Instance.Create.Parameters["properties"].(map[string]interface{})["components"]
	componentJSON, err := json.Marshal(componentItem)
	require.NoError(t, err)
	assert.JSONEq(t, fmt.Sprintf(`
		{
		  "type": "array",
		  "items": {
			  "type": "string",
			  "enum": %s
		  }
		}`, toJSONList(optComponentsNames)), string(componentJSON))
}

func toJSONList(in []string) string {
	return fmt.Sprintf(`["%s"]`, strings.Join(in, `", "`))
}

func TestBroker_ProvisioningScenario(t *testing.T) {
	// given
	const instID = "inst-id"
	const clusterName = "cluster-testing"

	fCli := provisioner.NewFakeClient()
	fdCli := director.NewFakeDirectorClient()
	memStorage := storage.NewMemoryStorage()

	fixInput := schema.ProvisionRuntimeInput{}
	factoryBuilderFake := &InputBuilderForPlanFake{
		inputToReturn: fixInput,
	}

	kymaEnvBroker, err := broker.New(broker.Config{EnablePlans: []string{"gcp", "azure"}}, fCli, fdCli, broker.ProvisioningConfig{}, memStorage.Instances(), nil, factoryBuilderFake, &broker.DumyDumper{})
	require.NoError(t, err)

	// when
	res, err := kymaEnvBroker.Provision(context.TODO(), instID, domain.ProvisionDetails{
		ServiceID:     serviceID,
		PlanID:        planID,
		RawParameters: json.RawMessage(fmt.Sprintf(`{"name": "%s"}`, clusterName)),
		RawContext:    json.RawMessage(fmt.Sprintf(`{"globalaccount_id": "%s"}`, globalAccountID)),
	}, true)
	require.NoError(t, err)

	// then
	assert.Equal(t, fixInput, fCli.GetProvisionRuntimeInput(0))

	inst, err := memStorage.Instances().GetByID(instID)
	require.NoError(t, err)
	assert.Equal(t, inst.InstanceID, instID)

	// when
	op, err := kymaEnvBroker.LastOperation(context.TODO(), instID, domain.PollDetails{
		ServiceID:     serviceID,
		PlanID:        planID,
		OperationData: res.OperationData,
	})

	// then
	require.NoError(t, err)
	assert.Equal(t, domain.InProgress, op.State)

	// when
	fCli.FinishProvisionerOperation(res.OperationData, schema.OperationStateSucceeded)
	op, err = kymaEnvBroker.LastOperation(context.TODO(), instID, domain.PollDetails{
		ServiceID:     serviceID,
		PlanID:        planID,
		OperationData: res.OperationData,
	})

	// then
	require.NoError(t, err)
	assert.Equal(t, domain.Succeeded, op.State)
}

type InputBuilderForPlanFake struct {
	inputToReturn schema.ProvisionRuntimeInput
}

func (i *InputBuilderForPlanFake) ForPlan(planID string) (broker.ConcreteInputBuilder, bool) {
	return &ConcreteInputBuilderFake{
		input: i.inputToReturn,
	}, true
}

type ConcreteInputBuilderFake struct {
	input schema.ProvisionRuntimeInput
}

func (c *ConcreteInputBuilderFake) SetProvisioningParameters(params internal.ProvisioningParametersDTO) broker.ConcreteInputBuilder {
	return c
}

func (c *ConcreteInputBuilderFake) SetERSContext(ersCtx internal.ERSContext) broker.ConcreteInputBuilder {
	return c
}

func (c *ConcreteInputBuilderFake) SetProvisioningConfig(brokerConfig broker.ProvisioningConfig) broker.ConcreteInputBuilder {
	return c
}

func (c *ConcreteInputBuilderFake) Build() (schema.ProvisionRuntimeInput, error) {
	return c.input, nil
}
