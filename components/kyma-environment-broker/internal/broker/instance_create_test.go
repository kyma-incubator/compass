package broker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/middleware"
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
	subAccountID    = "3cb65e5b-e455-4799-bf35-be46e8f5a533"

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
			memoryStorage.Instances(),
			queue,
			factoryBuilder,
			fixAlwaysPassJSONValidator(),
			false,
			logrus.StandardLogger(),
		)

		// when
		response, err := provisionEndpoint.Provision(fixReqCtxWithRegion(t, "req-region"), instanceID, domain.ProvisionDetails{
			ServiceID:     serviceID,
			PlanID:        planID,
			RawParameters: json.RawMessage(fmt.Sprintf(`{"name": "%s"}`, clusterName)),
			RawContext:    json.RawMessage(fmt.Sprintf(`{"globalaccount_id": "%s", "subaccount_id": "%s"}`, globalAccountID, subAccountID)),
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
		assert.Equal(t, "req-region", instanceParameters.PlatformRegion)

		instance, err := memoryStorage.Instances().GetByID(instanceID)
		require.NoError(t, err)

		assert.Equal(t, instance.ProvisioningParameters, operation.ProvisioningParameters)
		assert.Equal(t, instance.GlobalAccountID, globalAccountID)
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
			memoryStorage.Instances(),
			nil,
			factoryBuilder,
			fixAlwaysPassJSONValidator(),
			false,
			logrus.StandardLogger(),
		)

		// when
		response, err := provisionEndpoint.Provision(fixReqCtxWithRegion(t, "dummy"), instanceID, domain.ProvisionDetails{
			ServiceID:     serviceID,
			PlanID:        planID,
			RawParameters: json.RawMessage(fmt.Sprintf(`{"name": "%s"}`, clusterName)),
			RawContext:    json.RawMessage(fmt.Sprintf(`{"globalaccount_id": "%s", "subaccount_id": "%s"}`, globalAccountID, subAccountID)),
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
		err = memoryStorage.Instances().Insert(fixInstance())
		assert.NoError(t, err)

		factoryBuilder := &automock.PlanValidator{}
		factoryBuilder.On("IsPlanSupport", planID).Return(true)

		// #create provisioner endpoint
		provisionEndpoint := broker.NewProvision(
			broker.Config{EnablePlans: []string{"gcp", "azure"}},
			memoryStorage.Operations(),
			memoryStorage.Instances(),
			nil,
			factoryBuilder,
			fixAlwaysPassJSONValidator(),
			false,
			logrus.StandardLogger(),
		)

		// when
		response, err := provisionEndpoint.Provision(fixReqCtxWithRegion(t, "dummy"), instanceID, domain.ProvisionDetails{
			ServiceID:     serviceID,
			PlanID:        planID,
			RawParameters: json.RawMessage(fmt.Sprintf(`{"name": "%s"}`, clusterName)),
			RawContext:    json.RawMessage(fmt.Sprintf(`{"globalaccount_id": "%s", "subaccount_id": "%s"}`, "1cafb9c8-c8f8-478a-948a-9cb53bb76aa4", subAccountID)),
		}, true)

		// then
		assert.EqualError(t, err, "provisioning operation already exist")
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
			memoryStorage.Instances(),
			nil,
			factoryBuilder,
			fixValidator,
			false,
			logrus.StandardLogger(),
		)

		// when
		response, err := provisionEndpoint.Provision(fixReqCtxWithRegion(t, "dummy"), instanceID, domain.ProvisionDetails{
			ServiceID: serviceID,
			PlanID:    planID,
			RawParameters: json.RawMessage(fmt.Sprintf(`{
							"name": "%s", 
							"components": ["wrong component name"] 
							}`, clusterName)),
			RawContext: json.RawMessage(fmt.Sprintf(`{"globalaccount_id": "%s", "subaccount_id": "%s"}`, "1cafb9c8-c8f8-478a-948a-9cb53bb76aa4", subAccountID)),
		}, true)

		// then
		assert.EqualError(t, err, `while validating input parameters: components.0: components.0 must be one of the following: "Kiali", "Tracing"`)
		assert.False(t, response.IsAsync)
		assert.Empty(t, response.OperationData)
	})

	t.Run("return error on adding KnativeProvisionerNatss and NatssStreaming to list of components", func(t *testing.T) {
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
			memoryStorage.Instances(),
			nil,
			factoryBuilder,
			fixValidator,
			false,
			logrus.StandardLogger(),
		)

		// when
		response, err := provisionEndpoint.Provision(fixReqCtxWithRegion(t, "dummy"), instanceID, domain.ProvisionDetails{
			ServiceID: serviceID,
			PlanID:    planID,
			RawParameters: json.RawMessage(fmt.Sprintf(`{
								"name": "%s",
								"components": ["KnativeProvisionerNatss", "NatssStreaming"]
								}`, clusterName)),
			RawContext: json.RawMessage(fmt.Sprintf(`{"globalaccount_id": "%s", "subaccount_id": "%s"}`, "1cafb9c8-c8f8-478a-948a-9cb53bb76aa4", subAccountID)),
		}, true)

		// then
		assert.EqualError(t, err, `while validating input parameters: components.0: components.0 must be one of the following: "Kiali", "Tracing", components: No additional items allowed on array`)
		assert.False(t, response.IsAsync)
		assert.Empty(t, response.OperationData)
	})

	t.Run("kyma version parameters should be saved", func(t *testing.T) {
		// given
		memoryStorage := storage.NewMemoryStorage()

		factoryBuilder := &automock.PlanValidator{}
		factoryBuilder.On("IsPlanSupport", planID).Return(true)

		fixValidator, err := broker.NewPlansSchemaValidator()
		require.NoError(t, err)

		queue := &automock.Queue{}
		queue.On("Add", mock.AnythingOfType("string"))

		provisionEndpoint := broker.NewProvision(
			broker.Config{EnablePlans: []string{"gcp", "azure"}},
			memoryStorage.Operations(),
			memoryStorage.Instances(),
			queue,
			factoryBuilder,
			fixValidator,
			true,
			logrus.StandardLogger(),
		)

		// when
		response, err := provisionEndpoint.Provision(fixReqCtxWithRegion(t, "dummy"), instanceID, domain.ProvisionDetails{
			ServiceID: serviceID,
			PlanID:    planID,
			RawParameters: json.RawMessage(fmt.Sprintf(`{
								"name": "%s",
								"kymaVersion": "master-00e83e99"
								}`, clusterName)),
			RawContext: json.RawMessage(fmt.Sprintf(`{"globalaccount_id": "%s", "subaccount_id": "%s"}`, "1cafb9c8-c8f8-478a-948a-9cb53bb76aa4", subAccountID)),
		}, true)
		assert.NoError(t, err)

		// then
		operation, err := memoryStorage.Operations().GetProvisioningOperationByID(response.OperationData)
		require.NoError(t, err)

		parameters, err := operation.GetProvisioningParameters()
		assert.NoError(t, err)
		assert.Equal(t, "master-00e83e99", parameters.Parameters.KymaVersion)
	})

	t.Run("should return error when region is not specified", func(t *testing.T) {
		// given
		factoryBuilder := &automock.PlanValidator{}
		factoryBuilder.On("IsPlanSupport", planID).Return(true)

		fixValidator, err := broker.NewPlansSchemaValidator()
		require.NoError(t, err)

		provisionEndpoint := broker.NewProvision(
			broker.Config{EnablePlans: []string{"gcp", "azure"}},
			nil,
			nil,
			nil,
			factoryBuilder,
			fixValidator,
			true,
			logrus.StandardLogger(),
		)

		// when
		_, provisionErr := provisionEndpoint.Provision(context.Background(), instanceID, domain.ProvisionDetails{
			ServiceID:     serviceID,
			PlanID:        planID,
			RawParameters: json.RawMessage(fmt.Sprintf(`{"name": "%s"}`, clusterName)),
			RawContext:    json.RawMessage(fmt.Sprintf(`{"globalaccount_id": "%s", "subaccount_id": "%s"}`, "1cafb9c8-c8f8-478a-948a-9cb53bb76aa4", subAccountID)),
		}, true)

		// then
		require.EqualError(t, provisionErr, "No region specified in request.")
	})

	t.Run("kyma version parameters should NOT be saved", func(t *testing.T) {
		// given
		memoryStorage := storage.NewMemoryStorage()

		factoryBuilder := &automock.PlanValidator{}
		factoryBuilder.On("IsPlanSupport", planID).Return(true)

		fixValidator, err := broker.NewPlansSchemaValidator()
		require.NoError(t, err)

		queue := &automock.Queue{}
		queue.On("Add", mock.AnythingOfType("string"))

		provisionEndpoint := broker.NewProvision(
			broker.Config{EnablePlans: []string{"gcp", "azure"}},
			memoryStorage.Operations(),
			memoryStorage.Instances(),
			queue,
			factoryBuilder,
			fixValidator,
			false,
			logrus.StandardLogger(),
		)

		// when
		response, err := provisionEndpoint.Provision(fixReqCtxWithRegion(t, "dummy"), instanceID, domain.ProvisionDetails{
			ServiceID: serviceID,
			PlanID:    planID,
			RawParameters: json.RawMessage(fmt.Sprintf(`{
								"name": "%s",
								"kymaVersion": "master-00e83e99"
								}`, clusterName)),
			RawContext: json.RawMessage(fmt.Sprintf(`{"globalaccount_id": "%s", "subaccount_id": "%s"}`, "1cafb9c8-c8f8-478a-948a-9cb53bb76aa4", subAccountID)),
		}, true)
		assert.NoError(t, err)

		// then
		operation, err := memoryStorage.Operations().GetProvisioningOperationByID(response.OperationData)
		require.NoError(t, err)

		parameters, err := operation.GetProvisioningParameters()
		assert.NoError(t, err)
		assert.Equal(t, "", parameters.Parameters.KymaVersion)
	})
}

func fixExistOperation() internal.ProvisioningOperation {
	return internal.ProvisioningOperation{
		Operation: internal.Operation{
			ID:         existOperationID,
			InstanceID: instanceID,
		},
		ProvisioningParameters: fmt.Sprintf(
			`{"plan_id":"%s", "service_id": "%s", "ers_context":{"globalaccount_id": "%s", "subaccount_id": "%s"}, "parameters":{"name": "%s"}}`,
			planID, serviceID, globalAccountID, subAccountID, clusterName),
	}
}

func fixAlwaysPassJSONValidator() broker.PlansSchemaValidator {
	validatorMock := &automock.JSONSchemaValidator{}
	validatorMock.On("ValidateString", mock.Anything).Return(jsonschema.ValidationResult{Valid: true}, nil)

	fixValidator := broker.PlansSchemaValidator{
		broker.GCPPlanID:   validatorMock,
		broker.AzurePlanID: validatorMock,
	}

	return fixValidator
}

func fixInstance() internal.Instance {
	return internal.Instance{
		InstanceID:      instanceID,
		GlobalAccountID: globalAccountID,
		ServiceID:       serviceID,
		ServicePlanID:   planID,
	}
}

func fixReqCtxWithRegion(t *testing.T, region string) context.Context {
	t.Helper()

	req, err := http.NewRequest("GET", "http://url.io", nil)
	require.NoError(t, err)
	var ctx context.Context
	spyHandler := http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		ctx = req.Context()
	})

	middleware.AddRegionToContext(region).Middleware(spyHandler).ServeHTTP(httptest.NewRecorder(), req)
	return ctx
}
