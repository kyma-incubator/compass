// +build integration

package broker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/runtime"
	"net/http"
	"path"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

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
//   go test ./components/kyma-environment-broker/internal/broker/... -run=TestBrokerProvisioningScenario  -test.update-golden
func TestBrokerProvisioningScenario(t *testing.T) {
	// given
	const (
		clusterName = "cluster-testing"
		instID      = "inst-id"
		kymaVersion = "1.9.0"
	)

	fCli := provisioner.NewFakeClient()
	fdCli := director.NewFakeDirectorClient()
	memStorage := storage.NewMemoryStorage()
	graphqlizer := provisioner.Graphqlizer{}

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
	runtimeProvider := runtime.NewComponentsListProvider(http.DefaultClient, kymaVersion, path.Join("testdata", "managed-runtime-components.yaml"))

	fullRuntimeComponentList, err := runtimeProvider.AllComponents()
	require.NoError(t, err)

	inputFactory := broker.NewInputBuilderFactory(optComponentsSvc, fullRuntimeComponentList, kymaVersion, internal.ServiceManagerOverride{})

	kymaEnvBroker, err := broker.New(broker.Config{EnablePlans: []string{"gcp", "azure"}}, fCli, fdCli, broker.ProvisioningConfig{}, memStorage.Instances(), optComponentsSvc, inputFactory, &broker.DumyDumper{})
	require.NoError(t, err)

	// when
	_, err = kymaEnvBroker.Provision(context.TODO(), instID, domain.ProvisionDetails{
		ServiceID:     serviceID,
		PlanID:        planID,
		RawParameters: json.RawMessage(fmt.Sprintf(`{"name": "%s"}`, clusterName)),
		RawContext:    json.RawMessage(fmt.Sprintf(`{"globalaccount_id": "%s"}`, globalAccountID)),
	}, true)

	// then
	require.NoError(t, err)

	gotGQLInput, err := graphqlizer.ProvisionRuntimeInputToGraphQL(fCli.GetProvisionRuntimeInput(0))
	require.NoError(t, err)
	golden.Assert(t, gotGQLInput, t.Name()+".golden.graphql")
}
