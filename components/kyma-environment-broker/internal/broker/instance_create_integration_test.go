package broker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/mock"
	"path"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/runtime"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	hyperscalerMocks "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/mocks"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/stretchr/testify/require"
	"gotest.tools/golden"
)

// TestBrokerProvisioningScenario tests that we are sending proper provisioner input.
//
// This test is based on golden file.
// If the `-test.update-golden` flag is set then the actual content is written
// to the golden file.
//
// Example:
//   go test ./components/kyma-environment-broker/internal/broker/... -run=TestBrokerProvisioningScenario  -test.update-golden -tags=integration
func TestBrokerProvisioningScenario(t *testing.T) {
	// given
	const (
		clusterName     = "cluster-testing"
		instID          = "inst-id"
		kymaVersion     = "1.9.0"
		directorUrl     = "https://compass-gateway-auth-oauth.kyma.local/director/graphql"
		serviceID       = "47c9dcbf-ff30-448e-ab36-d3bad66ba281"
		planID          = "4deee563-e5ec-4731-b9b1-53b42d855f0c"
		globalAccountID = "e8f7ec0a-0cd6-41f0-905d-5d1efa9fb6c4"
	)

	var (
		fakeProvisionCli = provisioner.NewFakeClient()
		fakeDirectorCli  = director.NewFakeDirectorClient()
		memStorage       = storage.NewMemoryStorage()
		graphqlizer      = provisioner.Graphqlizer{}

		provisioningConfig = broker.ProvisioningConfig{}
		sm                 = internal.ServiceManagerOverride{
			CredentialsOverride: true,
			URL:                 "sm-url",
			Password:            "sm-pass",
			Username:            "sm-user",
		}
	)

	// this is the components configuration copied from the main.go file
	// TODO: after adding support for new Kyma version this can be
	// 	refactored and extracted to some dedicated service
	optionalComponentsDisablers := runtime.ComponentsDisablers{
		"Loki":       runtime.NewLokiDisabler(),
		"Kiali":      runtime.NewGenericComponentDisabler("kiali", "kyma-system"),
		"Jaeger":     runtime.NewGenericComponentDisabler("jaeger", "kyma-system"),
		"Monitoring": runtime.NewGenericComponentDisabler("monitoring", "kyma-system"),
	}
	optComponentsSvc := runtime.NewOptionalComponentsService(optionalComponentsDisablers)
	runtimeProvider := runtime.NewComponentsListProvider(kymaVersion, path.Join("testdata", "managed-runtime-components.yaml"))

	fullRuntimeComponentList, err := runtimeProvider.AllComponents()
	require.NoError(t, err)

	// ask Gophers about it
	accountProviderMock := &hyperscalerMocks.AccountProvider{}
	accountProviderMock.On("GardenerSecretName", mock.Anything, mock.Anything).Return("", nil)

	inputFactory := broker.NewInputBuilderFactory(optComponentsSvc, fullRuntimeComponentList, kymaVersion, sm, accountProviderMock, directorUrl)

	brokerCfg := broker.Config{EnablePlans: []string{"gcp", "azure"}}
	dumper := &broker.DumyDumper{}
	kymaEnvBroker := &broker.KymaEnvironmentBroker{
		ServicesEndpoint:             broker.NewServices(brokerCfg, optComponentsSvc, dumper),
		ProvisionEndpoint:            broker.NewProvision(brokerCfg, memStorage.Instances(), inputFactory, provisioningConfig, fakeProvisionCli, dumper),
		DeprovisionEndpoint:          broker.NewDeprovision(memStorage.Instances(), fakeProvisionCli, dumper),
		UpdateEndpoint:               broker.NewUpdate(dumper),
		GetInstanceEndpoint:          broker.NewGetInstance(memStorage.Instances(), dumper),
		LastOperationEndpoint:        broker.NewLastOperation(memStorage.Instances(), fakeProvisionCli, fakeDirectorCli, dumper),
		BindEndpoint:                 broker.NewBind(dumper),
		UnbindEndpoint:               broker.NewUnbind(dumper),
		GetBindingEndpoint:           broker.NewGetBinding(dumper),
		LastBindingOperationEndpoint: broker.NewLastBindingOperation(dumper),
	}

	// when
	_, err = kymaEnvBroker.Provision(context.TODO(), instID, domain.ProvisionDetails{
		ServiceID:     serviceID,
		PlanID:        planID,
		RawParameters: json.RawMessage(fmt.Sprintf(`{"name": "%s"}`, clusterName)),
		RawContext:    json.RawMessage(fmt.Sprintf(`{"globalaccount_id": "%s"}`, globalAccountID)),
	}, true)

	// then
	require.NoError(t, err)

	gotGQLInput, err := graphqlizer.ProvisionRuntimeInputToGraphQL(fakeProvisionCli.GetProvisionRuntimeInput(0))
	require.NoError(t, err)
	golden.Assert(t, gotGQLInput, t.Name()+".golden.graphql")
}
