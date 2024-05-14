package systemfielddiscoveryengine_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	systemfielddiscoveryengine "github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine"
	"github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine/automock"
	"github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine/config"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_ProcessSaasRegistryApplication(t *testing.T) {
	applicationID := "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	applicationTemplateID := "application-template-id"
	tenantID := "bbbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	internalTenantID := "internal-tenant-id"
	externalTenantID := "external-tenant-id"
	subscriptionsResponse := `{"subscriptions":[
						{"url":""},
						{"url":"app-url.com"}
					]}`
	applicationURL := "app-url.com"

	regionLabelKey := "region"
	regionLabelValue := "eu1"
	regionLabel := &model.Label{
		Key:   regionLabelKey,
		Value: regionLabelValue,
	}

	tnt := &model.BusinessTenantMapping{
		ID:             internalTenantID,
		ExternalTenant: externalTenantID,
	}

	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                               string
		TransactionerFn                    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ApplicationSvcFn                   func() *automock.ApplicationService
		ApplicationTemplateSvcFn           func() *automock.ApplicationTemplateService
		TenantSvcFn                        func() *automock.TenantService
		SystemFieldDiscoveryEngineConfigFn func() *automock.SystemFieldDiscoveryEngineConfig
		HTTPClientFn                       func() *automock.Client
		ExpectedErrorOutput                string
	}{
		{
			Name: "Success - app url was updated",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			ApplicationSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", mock.Anything, applicationID).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: applicationID,
					},
					ApplicationTemplateID: &applicationTemplateID,
				}, nil).Once()
				appSvc.On("UpdateBaseURLAndReadyState", mock.Anything, applicationID, applicationURL, true).Return(nil).Once()
				return appSvc
			},
			ApplicationTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetLabel", mock.Anything, applicationTemplateID, regionLabelKey).Return(regionLabel, nil).Once()
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", mock.Anything, resource.Application, applicationID).Return(internalTenantID, nil).Once()
				tenantSvc.On("GetTenantByID", mock.Anything, internalTenantID).Return(tnt, nil).Once()
				return tenantSvc
			},
			SystemFieldDiscoveryEngineConfigFn: func() *automock.SystemFieldDiscoveryEngineConfig {
				sfdConf := &automock.SystemFieldDiscoveryEngineConfig{}
				sfdConf.On("PrepareConfiguration").Return(&config.SystemFieldDiscoveryEngineConfig{
					RegionToSaasRegConfig: map[string]config.SaasRegConfig{regionLabelValue: {}},
				}, nil).Once()
				return sfdConf
			},
			HTTPClientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("Do", mock.Anything).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(subscriptionsResponse))),
				}, nil)
				return client
			},
		},
		{
			Name: "Success - app url was not found",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			ApplicationSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", mock.Anything, applicationID).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: applicationID,
					},
					ApplicationTemplateID: &applicationTemplateID,
				}, nil).Once()
				return appSvc
			},
			ApplicationTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetLabel", mock.Anything, applicationTemplateID, regionLabelKey).Return(regionLabel, nil).Once()
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", mock.Anything, resource.Application, applicationID).Return(internalTenantID, nil).Once()
				tenantSvc.On("GetTenantByID", mock.Anything, internalTenantID).Return(tnt, nil).Once()
				return tenantSvc
			},
			SystemFieldDiscoveryEngineConfigFn: func() *automock.SystemFieldDiscoveryEngineConfig {
				sfdConf := &automock.SystemFieldDiscoveryEngineConfig{}
				sfdConf.On("PrepareConfiguration").Return(&config.SystemFieldDiscoveryEngineConfig{
					RegionToSaasRegConfig: map[string]config.SaasRegConfig{regionLabelValue: {}},
				}, nil).Once()
				return sfdConf
			},
			HTTPClientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("Do", mock.Anything).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"subscriptions":[]}`))),
				}, nil)
				return client
			},
		},
		{
			Name:            "Error - GetLowestOwnerForResource return error",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ApplicationSvcFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			ApplicationTemplateSvcFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", mock.Anything, resource.Application, applicationID).Return("", testErr).Once()
				return tenantSvc
			},
			SystemFieldDiscoveryEngineConfigFn: func() *automock.SystemFieldDiscoveryEngineConfig {
				sfdConf := &automock.SystemFieldDiscoveryEngineConfig{}
				sfdConf.On("PrepareConfiguration").Return(&config.SystemFieldDiscoveryEngineConfig{}, nil).Once()
				return sfdConf
			},
			HTTPClientFn: func() *automock.Client {
				return &automock.Client{}
			},
			ExpectedErrorOutput: testErr.Error(),
		},
		{
			Name:            "Error - GetTenantByID return error",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ApplicationSvcFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			ApplicationTemplateSvcFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", mock.Anything, resource.Application, applicationID).Return(internalTenantID, nil).Once()
				tenantSvc.On("GetTenantByID", mock.Anything, internalTenantID).Return(nil, testErr).Once()
				return tenantSvc
			},
			SystemFieldDiscoveryEngineConfigFn: func() *automock.SystemFieldDiscoveryEngineConfig {
				sfdConf := &automock.SystemFieldDiscoveryEngineConfig{}
				sfdConf.On("PrepareConfiguration").Return(&config.SystemFieldDiscoveryEngineConfig{}, nil).Once()
				return sfdConf
			},
			HTTPClientFn: func() *automock.Client {
				return &automock.Client{}
			},
			ExpectedErrorOutput: testErr.Error(),
		},
		{
			Name: "Error - fail on getting application",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndThenDoesntExpectCommit(1)
			},
			ApplicationSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", mock.Anything, applicationID).Return(nil, testErr).Once()
				return appSvc
			},
			ApplicationTemplateSvcFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", mock.Anything, resource.Application, applicationID).Return(internalTenantID, nil).Once()
				tenantSvc.On("GetTenantByID", mock.Anything, internalTenantID).Return(tnt, nil).Once()
				return tenantSvc
			},
			SystemFieldDiscoveryEngineConfigFn: func() *automock.SystemFieldDiscoveryEngineConfig {
				sfdConf := &automock.SystemFieldDiscoveryEngineConfig{}
				sfdConf.On("PrepareConfiguration").Return(&config.SystemFieldDiscoveryEngineConfig{}, nil).Once()
				return sfdConf
			},
			HTTPClientFn: func() *automock.Client {
				return &automock.Client{}
			},
			ExpectedErrorOutput: testErr.Error(),
		},
		{
			Name: "Error - fail when the application does not have application template id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndThenDoesntExpectCommit(1)
			},
			ApplicationSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", mock.Anything, applicationID).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: applicationID,
					},
				}, nil).Once()
				return appSvc
			},
			ApplicationTemplateSvcFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", mock.Anything, resource.Application, applicationID).Return(internalTenantID, nil).Once()
				tenantSvc.On("GetTenantByID", mock.Anything, internalTenantID).Return(tnt, nil).Once()
				return tenantSvc
			},
			SystemFieldDiscoveryEngineConfigFn: func() *automock.SystemFieldDiscoveryEngineConfig {
				sfdConf := &automock.SystemFieldDiscoveryEngineConfig{}
				sfdConf.On("PrepareConfiguration").Return(&config.SystemFieldDiscoveryEngineConfig{}, nil).Once()
				return sfdConf
			},
			HTTPClientFn: func() *automock.Client {
				return &automock.Client{}
			},
			ExpectedErrorOutput: "application with id aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa does not have application template id",
		},
		{
			Name: "Error - fail on getting label",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndThenDoesntExpectCommit(1)
			},
			ApplicationSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", mock.Anything, applicationID).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: applicationID,
					},
					ApplicationTemplateID: &applicationTemplateID,
				}, nil).Once()
				return appSvc
			},
			ApplicationTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetLabel", mock.Anything, applicationTemplateID, regionLabelKey).Return(nil, testErr).Once()
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", mock.Anything, resource.Application, applicationID).Return(internalTenantID, nil).Once()
				tenantSvc.On("GetTenantByID", mock.Anything, internalTenantID).Return(tnt, nil).Once()
				return tenantSvc
			},
			SystemFieldDiscoveryEngineConfigFn: func() *automock.SystemFieldDiscoveryEngineConfig {
				sfdConf := &automock.SystemFieldDiscoveryEngineConfig{}
				sfdConf.On("PrepareConfiguration").Return(&config.SystemFieldDiscoveryEngineConfig{}, nil).Once()
				return sfdConf
			},
			HTTPClientFn: func() *automock.Client {
				return &automock.Client{}
			},
			ExpectedErrorOutput: testErr.Error(),
		},
		{
			Name: "Error - label value is not string",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndThenDoesntExpectCommit(1)
			},
			ApplicationSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", mock.Anything, applicationID).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: applicationID,
					},
					ApplicationTemplateID: &applicationTemplateID,
				}, nil).Once()
				return appSvc
			},
			ApplicationTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				label := &model.Label{
					Key:   regionLabelKey,
					Value: map[string]string{},
				}
				appTemplateSvc.On("GetLabel", mock.Anything, applicationTemplateID, regionLabelKey).Return(label, nil).Once()
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", mock.Anything, resource.Application, applicationID).Return(internalTenantID, nil).Once()
				tenantSvc.On("GetTenantByID", mock.Anything, internalTenantID).Return(tnt, nil).Once()
				return tenantSvc
			},
			SystemFieldDiscoveryEngineConfigFn: func() *automock.SystemFieldDiscoveryEngineConfig {
				sfdConf := &automock.SystemFieldDiscoveryEngineConfig{}
				sfdConf.On("PrepareConfiguration").Return(&config.SystemFieldDiscoveryEngineConfig{}, nil).Once()
				return sfdConf
			},
			HTTPClientFn: func() *automock.Client {
				return &automock.Client{}
			},
			ExpectedErrorOutput: "is not a string",
		},
		{
			Name: "Error - region is not present into the saas reg configuration",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			ApplicationSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", mock.Anything, applicationID).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: applicationID,
					},
					ApplicationTemplateID: &applicationTemplateID,
				}, nil).Once()
				return appSvc
			},
			ApplicationTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetLabel", mock.Anything, applicationTemplateID, regionLabelKey).Return(regionLabel, nil).Once()
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", mock.Anything, resource.Application, applicationID).Return(internalTenantID, nil).Once()
				tenantSvc.On("GetTenantByID", mock.Anything, internalTenantID).Return(tnt, nil).Once()
				return tenantSvc
			},
			SystemFieldDiscoveryEngineConfigFn: func() *automock.SystemFieldDiscoveryEngineConfig {
				sfdConf := &automock.SystemFieldDiscoveryEngineConfig{}
				sfdConf.On("PrepareConfiguration").Return(&config.SystemFieldDiscoveryEngineConfig{}, nil).Once()
				return sfdConf
			},
			HTTPClientFn: func() *automock.Client {
				return &automock.Client{}
			},
			ExpectedErrorOutput: "region \"eu1\" is not present into the saas reg configuration",
		},
		{
			Name: "Error - http client returns error",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			ApplicationSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", mock.Anything, applicationID).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: applicationID,
					},
					ApplicationTemplateID: &applicationTemplateID,
				}, nil).Once()
				return appSvc
			},
			ApplicationTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetLabel", mock.Anything, applicationTemplateID, regionLabelKey).Return(regionLabel, nil).Once()
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", mock.Anything, resource.Application, applicationID).Return(internalTenantID, nil).Once()
				tenantSvc.On("GetTenantByID", mock.Anything, internalTenantID).Return(tnt, nil).Once()
				return tenantSvc
			},
			SystemFieldDiscoveryEngineConfigFn: func() *automock.SystemFieldDiscoveryEngineConfig {
				sfdConf := &automock.SystemFieldDiscoveryEngineConfig{}
				sfdConf.On("PrepareConfiguration").Return(&config.SystemFieldDiscoveryEngineConfig{
					RegionToSaasRegConfig: map[string]config.SaasRegConfig{regionLabelValue: {}},
				}, nil).Once()
				return sfdConf
			},
			HTTPClientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("Do", mock.Anything).Return(nil, testErr)
				return client
			},
			ExpectedErrorOutput: "failed executing request",
		},
		{
			Name: "Error - http client returns non ok response code",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			ApplicationSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", mock.Anything, applicationID).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: applicationID,
					},
					ApplicationTemplateID: &applicationTemplateID,
				}, nil).Once()
				return appSvc
			},
			ApplicationTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetLabel", mock.Anything, applicationTemplateID, regionLabelKey).Return(regionLabel, nil).Once()
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", mock.Anything, resource.Application, applicationID).Return(internalTenantID, nil).Once()
				tenantSvc.On("GetTenantByID", mock.Anything, internalTenantID).Return(tnt, nil).Once()
				return tenantSvc
			},
			SystemFieldDiscoveryEngineConfigFn: func() *automock.SystemFieldDiscoveryEngineConfig {
				sfdConf := &automock.SystemFieldDiscoveryEngineConfig{}
				sfdConf.On("PrepareConfiguration").Return(&config.SystemFieldDiscoveryEngineConfig{
					RegionToSaasRegConfig: map[string]config.SaasRegConfig{regionLabelValue: {}},
				}, nil).Once()
				return sfdConf
			},
			HTTPClientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("Do", mock.Anything).Return(&http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
				}, nil)
				return client
			},
			ExpectedErrorOutput: "unexpected status code",
		},
		{
			Name: "Error - UpdateBaseURLAndReadyState returns error",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndThenDoesntExpectCommit(2)
			},
			ApplicationSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", mock.Anything, applicationID).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: applicationID,
					},
					ApplicationTemplateID: &applicationTemplateID,
				}, nil).Once()
				appSvc.On("UpdateBaseURLAndReadyState", mock.Anything, applicationID, applicationURL, true).Return(testErr).Once()
				return appSvc
			},
			ApplicationTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetLabel", mock.Anything, applicationTemplateID, regionLabelKey).Return(regionLabel, nil).Once()
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", mock.Anything, resource.Application, applicationID).Return(internalTenantID, nil).Once()
				tenantSvc.On("GetTenantByID", mock.Anything, internalTenantID).Return(tnt, nil).Once()
				return tenantSvc
			},
			SystemFieldDiscoveryEngineConfigFn: func() *automock.SystemFieldDiscoveryEngineConfig {
				sfdConf := &automock.SystemFieldDiscoveryEngineConfig{}
				sfdConf.On("PrepareConfiguration").Return(&config.SystemFieldDiscoveryEngineConfig{
					RegionToSaasRegConfig: map[string]config.SaasRegConfig{regionLabelValue: {}},
				}, nil).Once()
				return sfdConf
			},
			HTTPClientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("Do", mock.Anything).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(subscriptionsResponse))),
				}, nil)
				return client
			},
			ExpectedErrorOutput: testErr.Error(),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, tx := testCase.TransactionerFn()
			appSvc := testCase.ApplicationSvcFn()
			appTemplateSvc := testCase.ApplicationTemplateSvcFn()
			tenantSvc := testCase.TenantSvcFn()
			sfdConf := testCase.SystemFieldDiscoveryEngineConfigFn()
			client := testCase.HTTPClientFn()
			defer mock.AssertExpectationsForObjects(t, persist, tx, appSvc, appTemplateSvc, tenantSvc, sfdConf)

			svc, err := systemfielddiscoveryengine.NewSystemFieldDiscoverEngineService(sfdConf, client, tx, appSvc, appTemplateSvc, tenantSvc)
			assert.NoError(t, err)

			// WHEN
			err = svc.ProcessSaasRegistryApplication(context.TODO(), applicationID, tenantID)

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
