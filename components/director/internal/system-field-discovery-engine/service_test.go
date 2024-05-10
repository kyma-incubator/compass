package systemfielddiscoveryengine_test

import (
	"context"
	"errors"
	systemfielddiscoveryengine "github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine"
	"github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine/automock"
	"github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine/config"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_ProcessSaasRegistryApplication(t *testing.T) {
	applicationID := "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	tenantID := "bbbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	//operationID := "ccccccccc-cccc-cccc-cccc-cccccccccccc"
	//webhookID := "ddddddddd-dddd-dddd-dddd-dddddddddddd"
	//application := &model.Application{
	//	ApplicationTemplateID: &appTemplateID,
	//	BaseEntity:            &model.BaseEntity{ID: applicationID},
	//}

	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                     string
		TransactionerFn          func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ApplicationSvcFn         func() *automock.ApplicationService
		ApplicationTemplateSvcFn func() *automock.ApplicationTemplateService
		TenantSvcFn              func() *automock.TenantService
		ExpectedErrorOutput      string
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,

			ApplicationSvcFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			ApplicationTemplateSvcFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			TenantSvcFn: func() *automock.TenantService {
				tnt := &model.BusinessTenantMapping{
					ID:             "internal-tenant-id",
					ExternalTenant: "external-tenant-id",
				}

				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", mock.Anything, resource.Application, applicationID).Return("internal-tenant-id", nil).Once()
				tenantSvc.On("GetTenantByID", mock.Anything, "internal-tenant-id").Return(tnt, nil).Once()
				return tenantSvc
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, tx := testCase.TransactionerFn()
			appSvc := testCase.ApplicationSvcFn()
			appTemplateSvc := testCase.ApplicationTemplateSvcFn()
			tenantSvc := testCase.TenantSvcFn()
			defer mock.AssertExpectationsForObjects(t, persist, tx, appSvc, appTemplateSvc, tenantSvc)

			svc, err := systemfielddiscoveryengine.NewSystemFieldDiscoverEngineService(config.SystemFieldDiscoveryEngineConfig{
				SaasRegSecretPath:       "1",
				OauthTokenPath:          "2",
				SaasRegClientIDPath:     "3",
				SaasRegClientSecretPath: "4",
				SaasRegTokenURLPath:     "5",
				SaasRegURLPath:          "6",
				RegionToSaasRegConfig:   nil,
			}, nil, tx, appSvc, appTemplateSvc, tenantSvc)
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
