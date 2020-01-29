package broker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroker_Provision(t *testing.T) {
	// given
	const clusterName = "cluster-testing"

	tb := newTestBroker(t)

	// #setup provisioner client
	fCli := provisioner.NewFakeClient()

	// #setup builder
	fixInput := schema.ProvisionRuntimeInput{}
	factoryBuilderFake := &InputBuilderForPlanFake{
		inputToReturn: fixInput,
	}

	// #create broker
	tb.
		addInputBuilder(factoryBuilderFake).
		addProvisionerClient(fCli).
		createTestBroker()
	kymaEnvBroker := tb.broker

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

	inst, err := tb.storage.Instances().GetByID(instID)
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
