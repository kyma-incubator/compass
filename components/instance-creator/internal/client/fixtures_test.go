package client_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/automock"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/types"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/config"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testID              = "test-id"
	testName            = "test-name"
	instanceSMURLPath   = "http://test-url"
	region              = "test-region"
	catalogName         = "test-catalog-name"
	subaccountID        = "test-subaccount-id"
	offeringID          = "test-offering-id"
	planName            = "test-plan-name"
	planID              = "test-plan-id"
	serviceKeyID        = "test-service-key-id"
	serviceKeyName      = "test-service-key-name"
	serviceInstanceID   = "test-service-instance-id"
	serviceInstanceName = "test-service-instance-name"

	locationHeaderKey = "Location"

	invalidURL = "test://user:abc{DEf1=ghi@example.com"
)

var (
	testError         = errors.New("test-error")
	parameters        = []byte(`{}`)
	invalidParameters = []byte(`invalid-params`)
)

func callerThatGetsCalledOnce(statusCode int, responseBody []byte) func(*testing.T, config.Config, string) *automock.ExternalSvcCallerProvider {
	return func(t *testing.T, cfg config.Config, region string) *automock.ExternalSvcCallerProvider {
		svcCaller := &automock.ExternalSvcCaller{}
		response := httptest.ResponseRecorder{
			Code: statusCode,
			Body: bytes.NewBuffer(responseBody),
		}
		svcCaller.On("Call", mock.Anything).Return(response.Result(), nil).Once()

		svcCallerProvider := &automock.ExternalSvcCallerProvider{}
		svcCallerProvider.On("GetCaller", cfg, region).Return(svcCaller, nil).Once()
		return svcCallerProvider
	}
}

func callerThatGetsCalledSeveralTimesInAsyncCase(statusCodes []int, responseBodies [][]byte, shouldSkipLocationHeader bool) func(*testing.T, config.Config, string) *automock.ExternalSvcCallerProvider {
	return func(t *testing.T, cfg config.Config, region string) *automock.ExternalSvcCallerProvider {
		require.Equal(t, len(statusCodes), len(responseBodies))

		svcCaller := &automock.ExternalSvcCaller{}

		for i := 0; i < len(statusCodes); i++ {
			response := httptest.ResponseRecorder{
				Code: statusCodes[i],
				Body: bytes.NewBuffer(responseBodies[i]),
			}

			responseResult := response.Result()
			if !shouldSkipLocationHeader {
				responseResult.Header = make(map[string][]string)
				responseResult.Header.Set(locationHeaderKey, "/location")
			}

			svcCaller.On("Call", mock.Anything).Return(responseResult, nil).Once()
		}

		svcCallerProvider := &automock.ExternalSvcCallerProvider{}
		svcCallerProvider.On("GetCaller", cfg, region).Return(svcCaller, nil).Once()
		return svcCallerProvider
	}
}

func callerThatDoesNotGetCalled(t *testing.T, _ config.Config, _ string) *automock.ExternalSvcCallerProvider {
	svcCaller := &automock.ExternalSvcCaller{}
	svcCaller.AssertNotCalled(t, "Call", mock.Anything)

	svcCallerProvider := &automock.ExternalSvcCallerProvider{}
	svcCallerProvider.AssertNotCalled(t, "GetCaller", mock.Anything, mock.Anything)
	return svcCallerProvider
}

func callerThatDoesNotGetCalledButProviderIs(t *testing.T, cfg config.Config, region string) *automock.ExternalSvcCallerProvider {
	svcCaller := &automock.ExternalSvcCaller{}
	svcCaller.AssertNotCalled(t, "Call", mock.Anything)

	svcCallerProvider := &automock.ExternalSvcCallerProvider{}
	svcCallerProvider.On("GetCaller", cfg, region).Return(svcCaller, nil).Once()
	return svcCallerProvider
}

func callerProviderThatFails(_ *testing.T, cfg config.Config, region string) *automock.ExternalSvcCallerProvider {
	svcCallerProvider := &automock.ExternalSvcCallerProvider{}
	svcCallerProvider.On("GetCaller", cfg, region).Return(nil, testError).Once()
	return svcCallerProvider
}

func callerThatDoesNotSucceed(_ *testing.T, cfg config.Config, region string) *automock.ExternalSvcCallerProvider {
	svcCaller := &automock.ExternalSvcCaller{}
	svcCaller.On("Call", mock.Anything).Return(nil, testError).Once()

	svcCallerProvider := &automock.ExternalSvcCallerProvider{}
	svcCallerProvider.On("GetCaller", cfg, region).Return(svcCaller, nil).Once()
	return svcCallerProvider
}

func callerThatDoesNotSucceedTheLastTimeWhenCalled(statusCodes []int, responseBodies [][]byte) func(*testing.T, config.Config, string) *automock.ExternalSvcCallerProvider {
	return func(t *testing.T, cfg config.Config, region string) *automock.ExternalSvcCallerProvider {
		require.Equal(t, len(statusCodes), len(responseBodies))

		svcCaller := &automock.ExternalSvcCaller{}

		for i := 0; i <= len(statusCodes); i++ {
			if i != len(statusCodes) {
				response := httptest.ResponseRecorder{
					Code: statusCodes[i],
					Body: bytes.NewBuffer(responseBodies[i]),
				}

				responseResult := response.Result()
				responseResult.Header = make(map[string][]string)
				responseResult.Header.Set(locationHeaderKey, "/location")

				svcCaller.On("Call", mock.Anything).Return(responseResult, nil).Once()
			} else {
				svcCaller.On("Call", mock.Anything).Return(nil, testError).Once()
			}
		}

		svcCallerProvider := &automock.ExternalSvcCallerProvider{}
		svcCallerProvider.On("GetCaller", cfg, region).Return(svcCaller, nil).Once()
		return svcCallerProvider
	}
}

func callerThatReturnsBadStatus(_ *testing.T, cfg config.Config, region string) *automock.ExternalSvcCallerProvider {
	svcCaller := &automock.ExternalSvcCaller{}
	response := httptest.ResponseRecorder{
		Code: http.StatusBadRequest,
		Body: bytes.NewBufferString(""),
	}
	svcCaller.On("Call", mock.Anything).Return(response.Result(), nil).Once()

	svcCallerProvider := &automock.ExternalSvcCallerProvider{}
	svcCallerProvider.On("GetCaller", cfg, region).Return(svcCaller, nil).Once()
	return svcCallerProvider
}

func fixConfig() config.Config {
	return config.Config{
		InstanceSMURLPath: instanceSMURLPath,
		Ticker:            time.Millisecond * 10,
		Timeout:           time.Second,
	}
}

func fixConfigWithInvalidURL() config.Config {
	return config.Config{
		InstanceSMURLPath: invalidURL,
	}
}

func fixServiceOffering(id, catalogName string) types.ServiceOffering {
	return types.ServiceOffering{
		ID:          id,
		CatalogName: catalogName,
	}
}

func fixServiceOfferings() types.ServiceOfferings {
	return types.ServiceOfferings{
		NumItems: 2,
		Items:    []types.ServiceOffering{fixServiceOffering(offeringID, catalogName), fixServiceOffering(testID, testName)},
	}
}

func fixServiceOfferingsWithNoMatchingCatalogName() types.ServiceOfferings {
	return types.ServiceOfferings{
		NumItems: 2,
		Items:    []types.ServiceOffering{fixServiceOffering(offeringID, testName), fixServiceOffering(testID, testName)},
	}
}

func fixServicePlan(id, catalogName, offeringID string) types.ServicePlan {
	return types.ServicePlan{
		ID:                id,
		CatalogName:       catalogName,
		ServiceOfferingId: offeringID,
	}
}

func fixServicePlans() types.ServicePlans {
	return types.ServicePlans{
		NumItems: 2,
		Items:    []types.ServicePlan{fixServicePlan(planID, planName, offeringID), fixServicePlan(testID, testName, testID)},
	}
}

func fixServicePlansWithNoMatchingCatalogNameAndOfferingID() types.ServicePlans {
	return types.ServicePlans{
		NumItems: 2,
		Items:    []types.ServicePlan{fixServicePlan(testID, testName, testID), fixServicePlan(testID, testName, testID)},
	}
}

func fixServiceKey() *types.ServiceKey {
	return &types.ServiceKey{
		ID:                serviceKeyID,
		ServiceInstanceID: serviceInstanceID,
		Credentials:       json.RawMessage("{}"),
	}
}

func fixServiceKeys() *types.ServiceKeys {
	return &types.ServiceKeys{
		NumItems: 1,
		Items:    []types.ServiceKey{*fixServiceKey()},
	}
}

func fixServiceInstance(id, name string) types.ServiceInstance {
	return types.ServiceInstance{
		ID:   id,
		Name: name,
	}
}

func fixServiceInstances() types.ServiceInstances {
	return types.ServiceInstances{
		NumItems: 2,
		Items:    []types.ServiceInstance{fixServiceInstance(serviceInstanceID, serviceInstanceName), fixServiceInstance(testID, testName)},
	}
}

func fixServiceInstancesWithNoMatchingName() types.ServiceInstances {
	return types.ServiceInstances{
		NumItems: 2,
		Items:    []types.ServiceInstance{fixServiceInstance(serviceInstanceID, testName), fixServiceInstance(testID, testName)},
	}
}

func fixOperationStatus(state types.OperationState, resourceID string) types.OperationStatus {
	return types.OperationStatus{
		State:      state,
		ResourceID: resourceID,
	}
}

func fixOperationStatusWithEmptyResourceID(state types.OperationState) types.OperationStatus {
	return types.OperationStatus{
		State: state,
	}
}
