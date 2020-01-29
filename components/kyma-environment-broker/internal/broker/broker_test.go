package broker_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

const (
	serviceID       = "47c9dcbf-ff30-448e-ab36-d3bad66ba281"
	planID          = "4deee563-e5ec-4731-b9b1-53b42d855f0c"
	globalAccountID = "e8f7ec0a-0cd6-41f0-905d-5d1efa9fb6c4"

	instID = "inst-id"
)

type brokerTest struct {
	t      *testing.T
	broker *broker.KymaEnvBroker

	storage           storage.BrokerStorage
	ocnp              broker.OptionalComponentNamesProvider
	provisionerClient provisioner.Client
	directorClient    broker.DirectorClient
	builder           broker.InputBuilderForPlan
}

func newTestBroker(t *testing.T) *brokerTest {
	return &brokerTest{t: t}
}

func (bt *brokerTest) addStorage(s storage.BrokerStorage) *brokerTest {
	bt.storage = s
	bt.t.Log("storage added to test broker")
	return bt
}

func (bt *brokerTest) addProvisionerClient(pc provisioner.Client) *brokerTest {
	bt.provisionerClient = pc
	bt.t.Log("provisioner client added to test broker")
	return bt
}

func (bt *brokerTest) addDirectorClient(dc broker.DirectorClient) *brokerTest {
	bt.directorClient = dc
	bt.t.Log("director client added to test broker")
	return bt
}

func (bt *brokerTest) addOptionalComponentNamesProvider(ocnp broker.OptionalComponentNamesProvider) *brokerTest {
	bt.ocnp = ocnp
	bt.t.Log("components names provider added to test broker")
	return bt
}

func (bt *brokerTest) addInputBuilder(ib broker.InputBuilderForPlan) *brokerTest {
	bt.builder = ib
	bt.t.Log("concrete input builder added to test broker")
	return bt
}

func (bt *brokerTest) createTestBroker() {
	if bt.storage == nil {
		bt.t.Log("default MemoryStorage will be used in test broker")
		bt.storage = storage.NewMemoryStorage()
	}
	if bt.provisionerClient == nil {
		bt.t.Log("default provisioner FakeClient will be used in test broker")
		bt.provisionerClient = provisioner.NewFakeClient()
	}
	if bt.directorClient == nil {
		bt.t.Log("default director FakeClient will be used in test broker")
		bt.directorClient = director.NewFakeDirectorClient()
	}

	if bt.builder == nil {
		bt.t.Log("default InputBuilderForPlanFake will be used in test broker")
		bt.builder = InputBuilderForPlanFake{
			inputToReturn: schema.ProvisionRuntimeInput{},
		}
	}

	kymaEnvBroker, err := broker.New(
		broker.Config{EnablePlans: []string{"gcp", "azure"}},
		bt.provisionerClient,
		bt.directorClient,
		broker.ProvisioningConfig{},
		bt.storage.Instances(),
		bt.ocnp,
		bt.builder,
		&broker.DumyDumper{},
	)
	if err != nil {
		bt.t.Fatalf("cannot create broker: %s", err)
	}

	bt.broker = kymaEnvBroker
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
