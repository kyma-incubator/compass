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

	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	serviceID       = "47c9dcbf-ff30-448e-ab36-d3bad66ba281"
	planID          = "4deee563-e5ec-4731-b9b1-53b42d855f0c"
	globalAccountID = "e8f7ec0a-0cd6-41f0-905d-5d1efa9fb6c4"

	instanceID       = "d3d5dca4-5dc8-44ee-a825-755c2a3fb839"
	existOperationID = "920cbfd9-24e9-4aa2-aa77-879e9aabe140"
	clusterName      = "cluster-testing"
)

func TestProvision_Provision(t *testing.T) {
	t.Run("new operation will be created", func(t *testing.T) {
		// given
		// #setup memory storage
		memoryStorage := storage.NewMemoryStorage()

		queue := &automock.Queue{}
		queue.On("Add", mock.AnythingOfType("string"))

		factoryBuilder := &automock.PlanValidator{}
		factoryBuilder.On("IsPlanSupport", planID).Return(true)

		// #create provisioner endpoint
		provisionEndpoint := broker.NewProvision(
			broker.Config{EnablePlans: []string{"gcp", "azure"}},
			memoryStorage.Operations(),
			queue,
			factoryBuilder,
			fixAlwaysPassJSONValidator(),
			logrus.StandardLogger(),
		)

		// when
		response, err := provisionEndpoint.Provision(context.TODO(), instanceID, domain.ProvisionDetails{
			ServiceID:     serviceID,
			PlanID:        planID,
			RawParameters: json.RawMessage(fmt.Sprintf(`{"name": "%s"}`, clusterName)),
			RawContext:    json.RawMessage(fmt.Sprintf(`{"globalaccount_id": "%s"}`, globalAccountID)),
		}, true)

		// then
		require.NoError(t, err)
		assert.Regexp(t, "^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$", response.OperationData)
		assert.NotEqual(t, instanceID, response.OperationData)

		operation, err := memoryStorage.Operations().GetProvisioningOperationByID(response.OperationData)
		require.NoError(t, err)
		assert.Equal(t, operation.InstanceID, instanceID)

		var instanceParameters internal.ProvisioningParameters
		assert.NoError(t, json.Unmarshal([]byte(operation.ProvisioningParameters), &instanceParameters))

		assert.Equal(t, globalAccountID, instanceParameters.ErsContext.GlobalAccountID)
		assert.Equal(t, clusterName, instanceParameters.Parameters.Name)
	})

	t.Run("existing operation ID will be return", func(t *testing.T) {
		// given
		// #setup memory storage
		memoryStorage := storage.NewMemoryStorage()
		err := memoryStorage.Operations().InsertProvisioningOperation(fixExistOperation())
		assert.NoError(t, err)

		factoryBuilder := &automock.PlanValidator{}
		factoryBuilder.On("IsPlanSupport", planID).Return(true)

		// #create provisioner endpoint
		provisionEndpoint := broker.NewProvision(
			broker.Config{EnablePlans: []string{"gcp", "azure"}},
			memoryStorage.Operations(),
			nil,
			factoryBuilder,
			fixAlwaysPassJSONValidator(),
			logrus.StandardLogger(),
		)

		// when
		response, err := provisionEndpoint.Provision(context.TODO(), instanceID, domain.ProvisionDetails{
			ServiceID:     serviceID,
			PlanID:        planID,
			RawParameters: json.RawMessage(fmt.Sprintf(`{"name": "%s"}`, clusterName)),
			RawContext:    json.RawMessage(fmt.Sprintf(`{"globalaccount_id": "%s"}`, globalAccountID)),
		}, true)

		// then
		require.NoError(t, err)
		assert.Equal(t, existOperationID, response.OperationData)
		assert.True(t, response.AlreadyExists)
	})

	t.Run("conflict should be handled", func(t *testing.T) {
		// given
		// #setup memory storage
		memoryStorage := storage.NewMemoryStorage()
		err := memoryStorage.Operations().InsertProvisioningOperation(fixExistOperation())
		assert.NoError(t, err)

		factoryBuilder := &automock.PlanValidator{}
		factoryBuilder.On("IsPlanSupport", planID).Return(true)

		// #create provisioner endpoint
		provisionEndpoint := broker.NewProvision(
			broker.Config{EnablePlans: []string{"gcp", "azure"}},
			memoryStorage.Operations(),
			nil,
			factoryBuilder,
			fixAlwaysPassJSONValidator(),
			logrus.StandardLogger(),
		)

		// when
		response, err := provisionEndpoint.Provision(context.TODO(), instanceID, domain.ProvisionDetails{
			ServiceID:     serviceID,
			PlanID:        planID,
			RawParameters: json.RawMessage(fmt.Sprintf(`{"name": "%s"}`, clusterName)),
			RawContext:    json.RawMessage(fmt.Sprintf(`{"globalaccount_id": "%s"}`, "1cafb9c8-c8f8-478a-948a-9cb53bb76aa4")),
		}, true)

		// then
		assert.Error(t, err)
		assert.Empty(t, response.OperationData)
	})

	t.Run("return error on wrong input parameters", func(t *testing.T) {
		// given
		// #setup memory storage
		memoryStorage := storage.NewMemoryStorage()
		err := memoryStorage.Operations().InsertProvisioningOperation(fixExistOperation())
		require.NoError(t, err)

		factoryBuilder := &automock.PlanValidator{}
		factoryBuilder.On("IsPlanSupport", planID).Return(true)

		fixValidator, err := broker.NewPlansSchemaValidator()
		require.NoError(t, err)

		// #create provisioner endpoint
		provisionEndpoint := broker.NewProvision(
			broker.Config{EnablePlans: []string{"gcp", "azure"}},
			memoryStorage.Operations(),
			nil,
			factoryBuilder,
			fixValidator,
			&broker.DumyDumper{},
		)

		// when
		response, err := provisionEndpoint.Provision(context.TODO(), instanceID, domain.ProvisionDetails{
			ServiceID: serviceID,
			PlanID:    planID,
			RawParameters: json.RawMessage(fmt.Sprintf(`{
							"name": "%s", 
							"components": ["wrong component name"] 
							}`, clusterName)),
			RawContext: json.RawMessage(fmt.Sprintf(`{"globalaccount_id": "%s"}`, "1cafb9c8-c8f8-478a-948a-9cb53bb76aa4")),
		}, true)

		// then
		assert.EqualError(t, err, `while validating input parameters: components.0: components.0 must be one of the following: "Kiali", "Jaeger"`)
		assert.False(t, response.IsAsync)
		assert.Empty(t, response.OperationData)
	})
}

func fixExistOperation() internal.ProvisioningOperation {
	return internal.ProvisioningOperation{
		Operation: internal.Operation{
			ID:         existOperationID,
			InstanceID: instanceID,
		},
		ProvisioningParameters: fmt.Sprintf(
			`{"plan_id":"%s", "service_id": "%s", "ers_context":{"globalaccount_id": "%s"}, "parameters":{"name": "%s"}}`,
			planID, serviceID, globalAccountID, clusterName),
	}
}

func fixAlwaysPassJSONValidator() broker.PlansSchemaValidator {
	validatorMock := &automock.JSONSchemaValidator{}
	validatorMock.On("ValidateString", mock.Anything).Return(jsonschema.ValidationResult{Valid: true}, nil)

	fixValidator := broker.PlansSchemaValidator{
		broker.GcpPlanID:   validatorMock,
		broker.AzurePlanID: validatorMock,
	}

	return fixValidator
}
