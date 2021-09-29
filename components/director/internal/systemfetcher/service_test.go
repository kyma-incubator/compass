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

const (
	appType = "type-1"
)

func TestSyncSystems(t *testing.T) {
	const (
		appTemplateID        = "appTmp1"
		appRegisterInputJSON = `{ "name": "test1"}`
	)

	testErr := errors.New("testErr")
	appTemplate := &model.ApplicationTemplate{
		ID: appTemplateID,
	}

	appTemplateSvcNoErrors := func(systems []systemfetcher.System) *automock.ApplicationTemplateService {
		svc := &automock.ApplicationTemplateService{}
		for _, s := range systems {
			inputValues := inputValuesForSystem(s)
			svc.On("Get", txtest.CtxWithDBMatcher(), s.TemplateID).Return(appTemplate, nil).Once()
			svc.On("PrepareApplicationCreateInputJSON", appTemplate, inputValues).Return(appRegisterInputJSON, nil).Once()
		}
		return svc
	}
	appConverterNoErrors := func(appsInputs []model.ApplicationRegisterInput) *automock.ApplicationConverter {
		conv := &automock.ApplicationConverter{}
		for _, in := range appsInputs {
			conv.On("CreateInputJSONToModel", txtest.CtxWithDBMatcher(), appRegisterInputJSON).Return(in, nil).Once()
		}
		return conv
	}

	type testCase struct {
		name                string
		mockTransactioner   func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner)
		fixTestSystems      func() []systemfetcher.System
		fixAppInputs        func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate
		setupTenantSvc      func() systemfetcher.TenantService
		setupAppTemplateSvc func(systems []systemfetcher.System) *automock.ApplicationTemplateService
		setupAppConverter   func(appsInputs []model.ApplicationRegisterInput) *automock.ApplicationConverter
		setupSystemSvc      func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.SystemsService
		setupSysAPIClient   func(testSystems []systemfetcher.System) systemfetcher.SystemsAPIClient
		setupDirectorClient func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.DirectorClient
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
				systems[0].TemplateID = appTemplateID
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
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
			setupAppTemplateSvc: appTemplateSvcNoErrors,
			setupAppConverter:   appConverterNoErrors,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("CreateManyIfNotExistsWithEventualTemplate", txtest.CtxWithDBMatcher(), appsInputs).Return(nil).Once()
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) systemfetcher.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return(testSystems, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Success with one tenant and multiple systems",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(2)
				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateID = appTemplateID
				systems = append(systems, systemfetcher.System{
					SystemBase: systemfetcher.SystemBase{
						DisplayName:            "System2",
						ProductDescription:     "System2 description",
						BaseURL:                "http://example2.com",
						InfrastructureProvider: "test",
					},
					TemplateID: "appTmp2",
				})
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
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
			setupAppTemplateSvc: appTemplateSvcNoErrors,
			setupAppConverter:   appConverterNoErrors,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("CreateManyIfNotExistsWithEventualTemplate", txtest.CtxWithDBMatcher(), appsInputs).Return(nil).Once()
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) systemfetcher.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return(testSystems, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Success with multiple tenants with one system",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(3)
				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateID = appTemplateID
				systems = append(systems, systemfetcher.System{
					SystemBase: systemfetcher.SystemBase{
						DisplayName:            "System2",
						ProductDescription:     "System2 description",
						BaseURL:                "http://example2.com",
						InfrastructureProvider: "test",
					},
					TemplateID: "appTmp2",
				})
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
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
			setupAppTemplateSvc: appTemplateSvcNoErrors,
			setupAppConverter:   appConverterNoErrors,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("CreateManyIfNotExistsWithEventualTemplate", txtest.CtxWithDBMatcher(), []model.ApplicationRegisterInputWithTemplate{appsInputs[0]}).Return(nil).Once()
				systemSvc.On("CreateManyIfNotExistsWithEventualTemplate", txtest.CtxWithDBMatcher(), []model.ApplicationRegisterInputWithTemplate{appsInputs[1]}).Return(nil).Once()
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) systemfetcher.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[0]}, nil).Once()
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[1]}, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.DirectorClient {
				return &automock.DirectorClient{}
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
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
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
			setupAppTemplateSvc: func(systems []systemfetcher.System) *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			setupAppConverter: func(appsInputs []model.ApplicationRegisterInput) *automock.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.SystemsService {
				return &automock.SystemsService{}
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) systemfetcher.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return(nil, errors.New("expected")).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.DirectorClient {
				return &automock.DirectorClient{}
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
				systems[0].TemplateID = appTemplateID
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
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
			setupAppTemplateSvc: appTemplateSvcNoErrors,
			setupAppConverter:   appConverterNoErrors,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("CreateManyIfNotExistsWithEventualTemplate", txtest.CtxWithDBMatcher(), []model.ApplicationRegisterInputWithTemplate{appsInputs[0]}).Return(errors.New("expected")).Once()
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) systemfetcher.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[0]}, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Fail when template cannot be fetched",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
				persistTx := &pAutomock.PersistenceTx{}

				transactioner.On("Begin").Return(persistTx, nil).Once()
				transactioner.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateID = appTemplateID
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
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
			setupAppTemplateSvc: func(systems []systemfetcher.System) *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				for _, s := range systems {
					svc.On("Get", txtest.CtxWithDBMatcher(), s.TemplateID).Return(nil, testErr).Once()
				}
				return svc
			},
			setupAppConverter: func(appsInputs []model.ApplicationRegisterInput) *automock.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.SystemsService {
				return &automock.SystemsService{}
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) systemfetcher.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[0]}, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Fail when application JSON input cannot be prepared",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
				persistTx := &pAutomock.PersistenceTx{}

				transactioner.On("Begin").Return(persistTx, nil).Once()
				transactioner.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateID = appTemplateID
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
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
			setupAppTemplateSvc: func(systems []systemfetcher.System) *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				for _, s := range systems {
					inputValues := inputValuesForSystem(s)
					svc.On("Get", txtest.CtxWithDBMatcher(), s.TemplateID).Return(appTemplate, nil).Once()
					svc.On("PrepareApplicationCreateInputJSON", appTemplate, inputValues).Return("", testErr).Once()
				}
				return svc
			},
			setupAppConverter: func(appsInputs []model.ApplicationRegisterInput) *automock.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.SystemsService {
				return &automock.SystemsService{}
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) systemfetcher.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[0]}, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Fail when application input cannot be prepared",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
				persistTx := &pAutomock.PersistenceTx{}

				transactioner.On("Begin").Return(persistTx, nil).Once()
				transactioner.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateID = appTemplateID
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
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
			setupAppTemplateSvc: func(systems []systemfetcher.System) *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				for _, s := range systems {
					inputValues := inputValuesForSystem(s)
					svc.On("Get", txtest.CtxWithDBMatcher(), s.TemplateID).Return(appTemplate, nil).Once()
					svc.On("PrepareApplicationCreateInputJSON", appTemplate, inputValues).Return(appRegisterInputJSON, nil).Once()
				}
				return svc
			},
			setupAppConverter: func(appsInputs []model.ApplicationRegisterInput) *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("CreateInputJSONToModel", txtest.CtxWithDBMatcher(), appRegisterInputJSON).Return(model.ApplicationRegisterInput{}, testErr).Once()
				return conv
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.SystemsService {
				return &automock.SystemsService{}
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) systemfetcher.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[0]}, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.DirectorClient {
				return &automock.DirectorClient{}
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
				systems[0].TemplateID = "type1"
				systems = append(systems, systemfetcher.System{
					SystemBase: systemfetcher.SystemBase{
						DisplayName:            "System2",
						ProductDescription:     "System2 description",
						BaseURL:                "http://example2.com",
						InfrastructureProvider: "test",
					},
					TemplateID: "type2",
				})
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
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
			setupAppTemplateSvc: appTemplateSvcNoErrors,
			setupAppConverter:   appConverterNoErrors,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("CreateManyIfNotExistsWithEventualTemplate", txtest.CtxWithDBMatcher(), []model.ApplicationRegisterInputWithTemplate{appsInputs[0]}).Return(nil).Once()
				systemSvc.On("CreateManyIfNotExistsWithEventualTemplate", txtest.CtxWithDBMatcher(), []model.ApplicationRegisterInputWithTemplate{appsInputs[1]}).Return(nil).Once()
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) systemfetcher.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[0]}, nil).Once()
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return(nil, errors.New("expected")).Once()
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[1]}, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Do nothing if system is already being deleted",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(3)

				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateID = "type1"
				systems = append(systems, systemfetcher.System{
					SystemBase: systemfetcher.SystemBase{
						DisplayName:            "System2",
						ProductDescription:     "System2 description",
						BaseURL:                "http://example2.com",
						InfrastructureProvider: "test",
						AdditionalAttributes: systemfetcher.AdditionalAttributes{
							systemfetcher.LifecycleAttributeName: systemfetcher.LifecycleDeleted,
						},
						SystemNumber: "sysNumber1",
					},
					TemplateID: "type2",
				})
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsBySystems(systems)
			},
			setupTenantSvc: func() systemfetcher.TenantService {
				tenants := []*model.BusinessTenantMapping{
					newModelBusinessTenantMapping("t1", "tenant1"),
					newModelBusinessTenantMapping("t3", "tenant3"),
				}
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return tenantSvc
			},
			setupAppTemplateSvc: appTemplateSvcNoErrors,
			setupAppConverter:   appConverterNoErrors,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.SystemsService {
				systemSvc := &automock.SystemsService{}

				systemSvc.On("CreateManyIfNotExistsWithEventualTemplate", txtest.CtxWithDBMatcher(), []model.ApplicationRegisterInputWithTemplate{appsInputs[0]}).Return(nil).Once()
				systemSvc.On("GetByNameAndSystemNumber", txtest.CtxWithDBMatcher(), appsInputs[1].Name, *appsInputs[1].SystemNumber).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: "id",
					},
				}, nil)
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) systemfetcher.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[0]}, nil).Once()
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[1]}, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) systemfetcher.DirectorClient {
				directorClient := &automock.DirectorClient{}
				directorClient.On("DeleteSystemAsync", mock.Anything, "id", "t1").Return(nil).Once()
				return directorClient
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mockedTx, transactioner := testCase.mockTransactioner()
			tenantSvc := testCase.setupTenantSvc()
			testSystems := testCase.fixTestSystems()
			appsInputs := testCase.fixAppInputs(testSystems)
			systemSvc := testCase.setupSystemSvc(testSystems, appsInputs)
			appTemplateSvc := testCase.setupAppTemplateSvc(testSystems)
			appInputsWithoutTemplates := make([]model.ApplicationRegisterInput, 0)
			for _, in := range appsInputs {
				appInputsWithoutTemplates = append(appInputsWithoutTemplates, in.ApplicationRegisterInput)
			}
			appConverter := testCase.setupAppConverter(appInputsWithoutTemplates)
			sysAPIClient := testCase.setupSysAPIClient(testSystems)
			directorClient := testCase.setupDirectorClient(testSystems, appsInputs)

			svc := systemfetcher.NewSystemFetcher(transactioner, tenantSvc, systemSvc, appTemplateSvc, appConverter, sysAPIClient, directorClient, systemfetcher.Config{
				SystemsQueueSize:     100,
				FetcherParallellism:  30,
				EnableSystemDeletion: true,
			})
			err := svc.SyncSystems(context.TODO())
			require.NoError(t, err)

			mock.AssertExpectationsForObjects(t, tenantSvc, sysAPIClient, systemSvc, mockedTx, transactioner)
		})
	}
}

func inputValuesForSystem(s systemfetcher.System) model.ApplicationFromTemplateInputValues {
	return model.ApplicationFromTemplateInputValues{
		{
			Placeholder: "name",
			Value:       s.DisplayName,
		},
		{
			Placeholder: "display-name",
			Value:       s.DisplayName,
		},
	}
}

func fixAppsInputsBySystems(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
	initStatusCond := model.ApplicationStatusConditionInitial
	result := make([]model.ApplicationRegisterInputWithTemplate, 0, len(systems))
	for i := range systems {
		input := model.ApplicationRegisterInputWithTemplate{
			ApplicationRegisterInput: model.ApplicationRegisterInput{
				Name:            systems[i].DisplayName,
				Description:     &systems[i].ProductDescription,
				BaseURL:         &systems[i].BaseURL,
				ProviderName:    &systems[i].InfrastructureProvider,
				SystemNumber:    &systems[i].SystemNumber,
				StatusCondition: &initStatusCond,
				Labels: map[string]interface{}{
					"managed": "true",
				},
			},
			TemplateID: systems[i].TemplateID,
		}
		if len(input.TemplateID) > 0 {
			input.Labels["applicationType"] = appType
		}
		result = append(result, input)
	}
	return result
}

type asserter interface {
	assert()
}
