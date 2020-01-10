package broker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/automock"
	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	serviceID = "47c9dcbf-ff30-448e-ab36-d3bad66ba281"
	planID    = "4deee563-e5ec-4731-b9b1-53b42d855f0c"
)

func TestBroker_Services(t *testing.T) {
	// given
	instStorage := &automock.Instances{}
	defer instStorage.AssertExpectations(t)

	broker, err := broker.NewBroker(nil, broker.ProvisioningConfig{}, instStorage)
	require.NoError(t, err)

	// when
	services, err := broker.Services(context.TODO())

	// then
	require.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Len(t, services[0].Plans, 1)
}

func TestBroker_ProvisioningScenario(t *testing.T) {
	// given
	const instID = "inst-id"
	const clusterName = "cluster-testing"
	params := fmt.Sprintf(`{"name": "%s"}`, clusterName)

	instStorage := &automock.Instances{}
	defer instStorage.AssertExpectations(t)
	dashboardURL := "https://dummy.dashboard.com"
	instStorage.On("Insert", internal.Instance{
		InstanceID:             instID,
		RuntimeID:              "",
		GlobalAccountID:        "",
		ServiceID:              serviceID,
		ServicePlanID:          planID,
		DashboardURL:           dashboardURL,
		ProvisioningParameters: params,
	}).Return(nil).Once()

	fCli := provisioner.NewFakeClient()

	kymaEnvBroker, err := broker.NewBroker(fCli, broker.ProvisioningConfig{}, instStorage)
	require.NoError(t, err)

	// when
	res, err := kymaEnvBroker.Provision(context.TODO(), instID, domain.ProvisionDetails{
		ServiceID:     serviceID,
		PlanID:        planID,
		RawParameters: json.RawMessage(fmt.Sprintf(`{"name": "%s"}`, clusterName)),
		RawContext:    json.RawMessage(`{}`),
	}, true)
	require.NoError(t, err)

	// then
	assert.Equal(t, clusterName, fCli.GetProvisionRuntimeInput(0).ClusterConfig.GardenerConfig.Name)

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
