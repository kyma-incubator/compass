package broker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProvision_Provision(t *testing.T) {
	// given
	const (
		serviceID       = "47c9dcbf-ff30-448e-ab36-d3bad66ba281"
		planID          = "4deee563-e5ec-4731-b9b1-53b42d855f0c"
		globalAccountID = "e8f7ec0a-0cd6-41f0-905d-5d1efa9fb6c4"

		instID      = "inst-id"
		clusterName = "cluster-testing"
	)

	// #setup memory storage
	memoryStorage := storage.NewMemoryStorage()

	queue := &automock.Queue{}
	queue.On("Add", mock.AnythingOfType("string"))

	factoryBuilder := &automock.InputBuilderForPlan{}
	factoryBuilder.On("IsPlanSupport", planID).Return(true)

	// #create provisioner endpoint
	provisionEndpoint := broker.NewProvision(
		broker.Config{EnablePlans: []string{"gcp", "azure"}},
		memoryStorage.Operations(),
		queue,
		factoryBuilder,
		&broker.DumyDumper{},
	)

	// when
	response, err := provisionEndpoint.Provision(context.TODO(), instID, domain.ProvisionDetails{
		ServiceID:     serviceID,
		PlanID:        planID,
		RawParameters: json.RawMessage(fmt.Sprintf(`{"name": "%s"}`, clusterName)),
		RawContext:    json.RawMessage(fmt.Sprintf(`{"globalaccount_id": "%s"}`, globalAccountID)),
	}, true)
	require.NoError(t, err)

	// then
	assert.Regexp(t, "^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$", response.OperationData)
	assert.NotEqual(t, instID, response.OperationData)

	operation, err := memoryStorage.Operations().GetProvisioningOperationByID(response.OperationData)
	require.NoError(t, err)
	assert.Equal(t, operation.InstanceID, instID)

	var instanceParameters internal.ProvisioningParameters
	assert.NoError(t, json.Unmarshal([]byte(operation.ProvisioningParameters), &instanceParameters))

	assert.Equal(t, globalAccountID, instanceParameters.ErsContext.GlobalAccountID)
	assert.Equal(t, clusterName, instanceParameters.Parameters.Name)
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

func (c *ConcreteInputBuilderFake) SetInstanceID(instanceID string) broker.ConcreteInputBuilder {
	return c
}

func (c *ConcreteInputBuilderFake) Build() (schema.ProvisionRuntimeInput, error) {
	return c.input, nil
}
