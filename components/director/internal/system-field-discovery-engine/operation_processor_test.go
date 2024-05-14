package systemfielddiscoveryengine_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	systemfielddiscoveryengine "github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine"
	"github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOperationProcessor_Process(t *testing.T) {
	applicationID := "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	tenantID := "bbbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	testErr := errors.New("test error")

	opData := []byte(`
			{
				"applicationID": "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"tenantID": "bbbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
			}`)

	testCases := []struct {
		Name                      string
		SystemFieldDiscoverySvcFn func() *automock.SystemFieldDiscoveryService
		Operation                 *model.Operation
		ExpectedError             string
	}{
		{
			Name: "Success",
			SystemFieldDiscoverySvcFn: func() *automock.SystemFieldDiscoveryService {
				sfdSvc := &automock.SystemFieldDiscoveryService{}
				sfdSvc.On("ProcessSaasRegistryApplication", mock.Anything, applicationID, tenantID).Return(nil).Once()
				return sfdSvc
			},
			Operation: &model.Operation{
				OpType: model.OperationTypeSaasRegistryDiscovery,
				Data:   opData,
			},
		},
		{
			Name: "Success - empty application id",
			SystemFieldDiscoverySvcFn: func() *automock.SystemFieldDiscoveryService {
				return &automock.SystemFieldDiscoveryService{}
			},
			Operation: &model.Operation{
				OpType: model.OperationTypeSaasRegistryDiscovery,
				Data: []byte(`
					{
						"applicationID": "",
						"tenantID": "bbbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
					}`),
			},
		},
		{
			Name: "Success - empty tenant id",
			SystemFieldDiscoverySvcFn: func() *automock.SystemFieldDiscoveryService {
				return &automock.SystemFieldDiscoveryService{}
			},
			Operation: &model.Operation{
				OpType: model.OperationTypeSaasRegistryDiscovery,
				Data: []byte(`
					{
						"applicationID": "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
						"tenantID": ""
					}`),
			},
		},
		{
			Name: "Error - ProcessSaasRegistryApplication returns error",
			SystemFieldDiscoverySvcFn: func() *automock.SystemFieldDiscoveryService {
				sfdSvc := &automock.SystemFieldDiscoveryService{}
				sfdSvc.On("ProcessSaasRegistryApplication", mock.Anything, applicationID, tenantID).Return(testErr).Once()
				return sfdSvc
			},
			Operation: &model.Operation{
				OpType: model.OperationTypeSaasRegistryDiscovery,
				Data:   opData,
			},
			ExpectedError: testErr.Error(),
		},
		{
			Name: "Error - unsupported operation type",
			SystemFieldDiscoverySvcFn: func() *automock.SystemFieldDiscoveryService {
				return &automock.SystemFieldDiscoveryService{}
			},
			Operation: &model.Operation{
				OpType: model.OperationTypeOrdAggregation,
			},
			ExpectedError: "unsupported operation type",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			systemFieldDiscoverySvc := testCase.SystemFieldDiscoverySvcFn()
			defer mock.AssertExpectationsForObjects(t, systemFieldDiscoverySvc)

			operationProcessor := systemfielddiscoveryengine.NewOperationProcessor(systemFieldDiscoverySvc)
			err := operationProcessor.Process(context.TODO(), testCase.Operation)

			if len(testCase.ExpectedError) > 0 {
				assert.Contains(t, err.Error(), testCase.ExpectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
