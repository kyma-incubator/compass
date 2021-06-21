package systemfetcher_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher/automock"
	pAutomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSyncSystems(t *testing.T) {
	type testCase struct {
		name                string
		mockTransactioner   func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner)
		fixTestSystems      func() []systemfetcher.System
		fixAppInputs        func(systems []systemfetcher.System) []model.ApplicationRegisterInput
		setupTenantSvc      func() systemfetcher.TenantService
		setupSystemSvc      func(appsInputs []model.ApplicationRegisterInput) systemfetcher.SystemsService
		setupSysAPIClient   func(testSystems []systemfetcher.System) systemfetcher.SystemsAPIClient
		setupAppTemplateSvc func() systemfetcher.ApplicationTemplateService
	}
	tests := []testCase{
		{
			name: "Success with one tenant and one system",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(2)
				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateType = "type1"
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInput {
				return fixAppsInputsBySystems(systems)
			},
			setupTenantSvc: func() systemfetcher.TenantService {
				tenants := []*model.BusinessTenantMapping{
					newModelBusinessTenantMapping("t1", "tenant1"),
				}
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return tenantSvc
			},
			setupSystemSvc: func(appsInputs []model.ApplicationRegisterInput) systemfetcher.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("CreateManyIfNotExistsWithEventualTemplate", txtest.CtxWithDBMatcher(), appsInputs, []string{"appTmp1"}).Return(nil).Once()
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) systemfetcher.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return(testSystems, nil).Once()
				return sysAPIClient
			},
			setupAppTemplateSvc: func() systemfetcher.ApplicationTemplateService {
				appTmp := &model.ApplicationTemplate{
					ID: "appTmp1",
				}
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByName", txtest.CtxWithDBMatcher(), "type1").Return(appTmp, nil).Once()
				return appTemplateSvc
			},
		},
		{
			name: "Sucess with one tenant and multiple systems",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(2)
				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateType = "type1"
				systems = append(systems, systemfetcher.System{
					SystemBase: systemfetcher.SystemBase{
						DisplayName:            "System2",
						ProductDescription:     "System2 description",
						BaseURL:                "http://example2.com",
						InfrastructureProvider: "test",
					},
					TemplateType: "type2",
				})
				return systems
			},
			setupAppTemplateSvc: func() systemfetcher.ApplicationTemplateService {
				appTmp := &model.ApplicationTemplate{
					ID: "appTmp1",
				}
				appTmp2 := &model.ApplicationTemplate{
					ID: "appTmp2",
				}
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByName", txtest.CtxWithDBMatcher(), "type1").Return(appTmp, nil).Once()
				appTemplateSvc.On("GetByName", txtest.CtxWithDBMatcher(), "type2").Return(appTmp2, nil).Once()
				return appTemplateSvc
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInput {
				return fixAppsInputsBySystems(systems)
			},
			setupTenantSvc: func() systemfetcher.TenantService {
				tenants := []*model.BusinessTenantMapping{
					newModelBusinessTenantMapping("t1", "tenant1"),
				}
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return tenantSvc
			},
			setupSystemSvc: func(appsInputs []model.ApplicationRegisterInput) systemfetcher.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("CreateManyIfNotExistsWithEventualTemplate", txtest.CtxWithDBMatcher(), appsInputs, []string{"appTmp1", "appTmp2"}).Return(nil).Once()
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) systemfetcher.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return(testSystems, nil).Once()
				return sysAPIClient
			},
		},
		{
			name: "Sucess with multiple tenants with one system",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(3)
				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateType = "type1"
				systems = append(systems, systemfetcher.System{
					SystemBase: systemfetcher.SystemBase{
						DisplayName:            "System2",
						ProductDescription:     "System2 description",
						BaseURL:                "http://example2.com",
						InfrastructureProvider: "test",
					},
					TemplateType: "type2",
				})
				return systems
			},
			setupAppTemplateSvc: func() systemfetcher.ApplicationTemplateService {
				appTmp := &model.ApplicationTemplate{
					ID: "appTmp1",
				}
				appTmp2 := &model.ApplicationTemplate{
					ID: "appTmp2",
				}
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByName", txtest.CtxWithDBMatcher(), "type1").Return(appTmp, nil).Once()
				appTemplateSvc.On("GetByName", txtest.CtxWithDBMatcher(), "type2").Return(appTmp2, nil).Once()
				return appTemplateSvc
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInput {
				return fixAppsInputsBySystems(systems)
			},
			setupTenantSvc: func() systemfetcher.TenantService {
				tenants := []*model.BusinessTenantMapping{
					newModelBusinessTenantMapping("t1", "tenant1"),
					newModelBusinessTenantMapping("t2", "tenant2"),
				}
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return tenantSvc
			},
			setupSystemSvc: func(appsInputs []model.ApplicationRegisterInput) systemfetcher.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("CreateManyIfNotExistsWithEventualTemplate", txtest.CtxWithDBMatcher(), []model.ApplicationRegisterInput{appsInputs[0]}, []string{"appTmp1"}).Return(nil).Once()
				systemSvc.On("CreateManyIfNotExistsWithEventualTemplate", txtest.CtxWithDBMatcher(), []model.ApplicationRegisterInput{appsInputs[1]}, []string{"appTmp2"}).Return(nil).Once()
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) systemfetcher.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[0]}, nil).Once()
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[1]}, nil).Once()
				return sysAPIClient
			},
		},
		{
			name: "Fail when client fails to fetch systems",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()

				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				return fixSystems()
			},
			setupAppTemplateSvc: func() systemfetcher.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				return appTemplateSvc
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInput {
				return fixAppsInputsBySystems(systems)
			},
			setupTenantSvc: func() systemfetcher.TenantService {
				tenants := []*model.BusinessTenantMapping{
					newModelBusinessTenantMapping("t1", "tenant1"),
				}
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return tenantSvc
			},
			setupSystemSvc: func(appsInputs []model.ApplicationRegisterInput) systemfetcher.SystemsService {
				systemSvc := &automock.SystemsService{}
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) systemfetcher.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return(nil, errors.New("expected")).Once()
				return sysAPIClient
			},
		},
		{
			name: "Fail when service fails to save systems",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
				persistTx := &pAutomock.PersistenceTx{}

				transactioner.On("Begin").Return(persistTx, nil).Once()
				transactioner.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateType = "type1"
				return systems
			},
			setupAppTemplateSvc: func() systemfetcher.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTmp := &model.ApplicationTemplate{
					ID: "appTmp1",
				}
				appTemplateSvc.On("GetByName", txtest.CtxWithDBMatcher(), "type1").Return(appTmp, nil).Once()
				return appTemplateSvc
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInput {
				return fixAppsInputsBySystems(systems)
			},
			setupTenantSvc: func() systemfetcher.TenantService {
				tenants := []*model.BusinessTenantMapping{
					newModelBusinessTenantMapping("t1", "tenant1"),
				}
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return tenantSvc
			},
			setupSystemSvc: func(appsInputs []model.ApplicationRegisterInput) systemfetcher.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("CreateManyIfNotExistsWithEventualTemplate", txtest.CtxWithDBMatcher(), []model.ApplicationRegisterInput{appsInputs[0]}, []string{"appTmp1"}).Return(errors.New("expected")).Once()
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) systemfetcher.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[0]}, nil).Once()
				return sysAPIClient
			},
		},
		{
			name: "Succeed when client fails to fetch systems only for some tenants",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(3)

				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateType = "type1"
				systems = append(systems, systemfetcher.System{
					SystemBase: systemfetcher.SystemBase{
						DisplayName:            "System2",
						ProductDescription:     "System2 description",
						BaseURL:                "http://example2.com",
						InfrastructureProvider: "test",
					},
					TemplateType: "type2",
				})
				return systems
			},
			setupAppTemplateSvc: func() systemfetcher.ApplicationTemplateService {
				appTmp := &model.ApplicationTemplate{
					ID: "appTmp1",
				}
				appTmp2 := &model.ApplicationTemplate{
					ID: "appTmp2",
				}
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByName", txtest.CtxWithDBMatcher(), "type1").Return(appTmp, nil).Once()
				appTemplateSvc.On("GetByName", txtest.CtxWithDBMatcher(), "type2").Return(appTmp2, nil).Once()
				return appTemplateSvc
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInput {
				return fixAppsInputsBySystems(systems)
			},
			setupTenantSvc: func() systemfetcher.TenantService {
				tenants := []*model.BusinessTenantMapping{
					newModelBusinessTenantMapping("t1", "tenant1"),
					newModelBusinessTenantMapping("t2", "tenant2"),
					newModelBusinessTenantMapping("t3", "tenant3"),
				}
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return tenantSvc
			},
			setupSystemSvc: func(appsInputs []model.ApplicationRegisterInput) systemfetcher.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("CreateManyIfNotExistsWithEventualTemplate", txtest.CtxWithDBMatcher(), []model.ApplicationRegisterInput{appsInputs[0]}, []string{"appTmp1"}).Return(nil).Once()
				systemSvc.On("CreateManyIfNotExistsWithEventualTemplate", txtest.CtxWithDBMatcher(), []model.ApplicationRegisterInput{appsInputs[1]}, []string{"appTmp2"}).Return(nil).Once()
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) systemfetcher.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[0]}, nil).Once()
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return(nil, errors.New("expected")).Once()
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[1]}, nil).Once()
				return sysAPIClient
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mockedTx, transactioner := testCase.mockTransactioner()
			tenantSvc := testCase.setupTenantSvc()
			testSystems := testCase.fixTestSystems()
			appsInputs := testCase.fixAppInputs(testSystems)
			systemSvc := testCase.setupSystemSvc(appsInputs)
			sysAPIClient := testCase.setupSysAPIClient(testSystems)
			appTemplateSvc := testCase.setupAppTemplateSvc()

			svc := systemfetcher.NewSystemFetcher(transactioner, tenantSvc, systemSvc, sysAPIClient, appTemplateSvc, 30)
			err := svc.SyncSystems(context.TODO())
			require.NoError(t, err)

			mock.AssertExpectationsForObjects(t, tenantSvc, sysAPIClient, systemSvc, appTemplateSvc, mockedTx, transactioner)
		})
	}
}

func fixAppsInputsBySystems(systems []systemfetcher.System) []model.ApplicationRegisterInput {
	initStatusCond := model.ApplicationStatusConditionManaged
	result := make([]model.ApplicationRegisterInput, 0, len(systems))
	for i := range systems {
		result = append(result, model.ApplicationRegisterInput{
			Name:            systems[i].DisplayName,
			Description:     &systems[i].ProductDescription,
			BaseURL:         &systems[i].BaseURL,
			ProviderName:    &systems[i].InfrastructureProvider,
			SystemNumber:    &systems[i].SystemNumber,
			StatusCondition: &initStatusCond,
		})
	}
	return result
}
