package broker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/stretchr/testify/assert"
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

	// #setup provisioner client
	fCli := provisioner.NewFakeClient()

	// #setup builder
	fixInput := schema.ProvisionRuntimeInput{}
	factoryBuilderFake := &InputBuilderForPlanFake{
		inputToReturn: fixInput,
	}

	// #create provisioner endpoint
	provisionEndpoint := broker.NewProvision(
		broker.Config{EnablePlans: []string{"gcp", "azure"}},
		memoryStorage.Instances(),
		factoryBuilderFake,
		broker.ProvisioningConfig{},
		fCli,
		&broker.DumyDumper{},
	)

	// when
	_, err := provisionEndpoint.Provision(context.TODO(), instID, domain.ProvisionDetails{
		ServiceID:     serviceID,
		PlanID:        planID,
		RawParameters: json.RawMessage(fmt.Sprintf(`{"name": "%s"}`, clusterName)),
		RawContext:    json.RawMessage(fmt.Sprintf(`{"globalaccount_id": "%s"}`, globalAccountID)),
	}, true)
	require.NoError(t, err)

	// then
	assert.Equal(t, fixInput, fCli.GetProvisionRuntimeInput(0))

	inst, err := memoryStorage.Instances().GetByID(instID)
	require.NoError(t, err)
	assert.Equal(t, inst.InstanceID, instID)
}

type InputBuilderForPlanFake struct {
	inputToReturn schema.ProvisionRuntimeInput
}

func (i InputBuilderForPlanFake) ForPlan(planID string) (broker.ConcreteInputBuilder, bool) {
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
