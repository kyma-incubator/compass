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
	appType    = "type-1"
	mainURLKey = "mainUrl"
)

func TestSyncSystems(t *testing.T) {
	const appTemplateID = "appTmp1"
	testErr := errors.New("testErr")
	setupSuccessfulTemplateRenderer := func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInput) *automock.TemplateRenderer {
		svc := &automock.TemplateRenderer{}
		for i := range appsInputs {
			svc.On("ApplicationRegisterInputFromTemplate", txtest.CtxWithDBMatcher(), systems[i]).Return(&appsInputs[i], nil).Once()
		}
		return svc
	}

	type testCase struct {
		name                     string
		mockTransactioner        func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner)
		fixTestSystems           func() []systemfetcher.System
		fixAppInputs             func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate
		setupTenantSvc           func() *automock.TenantService
		setupTemplateRendererSvc func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInput) *automock.TemplateRenderer
		setupSystemSvc           func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService
		setupSysAPIClient        func(testSystems []systemfetcher.System) *automock.SystemsAPIClient
		setupDirectorClient      func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient
		expectedErr              error
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
				systems[0].ProductID = "TEST"
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(systems)
			},
			setupTenantSvc: func() *automock.TenantService {
				tenants := []*model.BusinessTenantMapping{
					newModelBusinessTenantMapping("t1", "tenant1"),
				}
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return tenantSvc
			},
			setupTemplateRendererSvc: setupSuccessfulTemplateRenderer,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), appsInputs[0].ApplicationRegisterInput, mock.Anything).Return(nil).Once()
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return(testSystems, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Success with one tenant and one system with null base url",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(2)
				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := []systemfetcher.System{
					{
						SystemBase: systemfetcher.SystemBase{
							DisplayName:            "System1",
							ProductDescription:     "System1 description",
							InfrastructureProvider: "test",
						},
					},
				}
				systems[0].TemplateID = "type1"
				systems[0].ProductID = "TEST"
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(systems)
			},
			setupTenantSvc: func() *automock.TenantService {
				tenants := []*model.BusinessTenantMapping{
					newModelBusinessTenantMapping("t1", "tenant1"),
				}
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return tenantSvc
			},
			setupTemplateRendererSvc: setupSuccessfulTemplateRenderer,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), mock.AnythingOfType("model.ApplicationRegisterInput"), mock.Anything).Return(nil).Once()
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return(testSystems, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Success with one tenant and one system without template",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(2)
				return mockedTx, transactioner
			},
			fixTestSystems: fixSystems,
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(systems)
			},
			setupTenantSvc: func() *automock.TenantService {
				tenants := []*model.BusinessTenantMapping{
					newModelBusinessTenantMapping("t1", "tenant1"),
				}
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return tenantSvc
			},
			setupTemplateRendererSvc: func(_ []systemfetcher.System, _ []model.ApplicationRegisterInput) *automock.TemplateRenderer {
				return &automock.TemplateRenderer{}
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("TrustedUpsert", txtest.CtxWithDBMatcher(), appsInputs[0].ApplicationRegisterInput, mock.Anything).Return(nil).Once()
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return(testSystems, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Success with one tenant and multiple systems",
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
				return fixAppsInputsWithTemplatesBySystems(systems)
			},
			setupTenantSvc: func() *automock.TenantService {
				tenants := []*model.BusinessTenantMapping{
					newModelBusinessTenantMapping("t1", "tenant1"),
				}
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return tenantSvc
			},
			setupTemplateRendererSvc: setupSuccessfulTemplateRenderer,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), appsInputs[0].ApplicationRegisterInput, mock.Anything).Return(nil).Once()
				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), appsInputs[1].ApplicationRegisterInput, mock.Anything).Return(nil).Once()
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return(testSystems, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
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
				return fixAppsInputsWithTemplatesBySystems(systems)
			},
			setupTenantSvc: func() *automock.TenantService {
				firstTenant := newModelBusinessTenantMapping("t1", "tenant1")
				firstTenant.ExternalTenant = "t1"
				secondTenant := newModelBusinessTenantMapping("t2", "tenant2")
				secondTenant.ExternalTenant = "t2"
				tenants := []*model.BusinessTenantMapping{firstTenant, secondTenant}
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return tenantSvc
			},
			setupTemplateRendererSvc: setupSuccessfulTemplateRenderer,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), appsInputs[0].ApplicationRegisterInput, mock.Anything).Return(nil).Once()
				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), appsInputs[1].ApplicationRegisterInput, mock.Anything).Return(nil).Once()
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "t1").Return([]systemfetcher.System{testSystems[0]}, nil).Once()
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "t2").Return([]systemfetcher.System{testSystems[1]}, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Fail when tenant fetching fails",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()

				return mockedTx, transactioner
			},
			fixTestSystems: fixSystems,
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(systems)
			},
			setupTenantSvc: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(nil, testErr).Once()
				return tenantSvc
			},
			setupTemplateRendererSvc: func(_ []systemfetcher.System, _ []model.ApplicationRegisterInput) *automock.TemplateRenderer {
				return &automock.TemplateRenderer{}
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				return &automock.SystemsService{}
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				return &automock.SystemsAPIClient{}
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
			expectedErr: testErr,
		},
		{
			name: "Fail when transaction cannot be started",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(testErr).ThatFailsOnBegin()

				return mockedTx, transactioner
			},
			fixTestSystems: fixSystems,
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(systems)
			},
			setupTenantSvc: func() *automock.TenantService {
				return &automock.TenantService{}
			},
			setupTemplateRendererSvc: func(_ []systemfetcher.System, _ []model.ApplicationRegisterInput) *automock.TemplateRenderer {
				return &automock.TemplateRenderer{}
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				return &automock.SystemsService{}
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				return &automock.SystemsAPIClient{}
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
			expectedErr: testErr,
		},
		{
			name: "Fail when commit fails",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(testErr).ThatFailsOnCommit()
				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateID = appTemplateID
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(systems)
			},
			setupTenantSvc: func() *automock.TenantService {
				tenants := []*model.BusinessTenantMapping{
					newModelBusinessTenantMapping("t1", "tenant1"),
				}
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return tenantSvc
			},
			setupTemplateRendererSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInput) *automock.TemplateRenderer {
				return &automock.TemplateRenderer{}
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				return &automock.SystemsService{}
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				return &automock.SystemsAPIClient{}
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
			expectedErr: testErr,
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
				return fixAppsInputsWithTemplatesBySystems(systems)
			},
			setupTenantSvc: func() *automock.TenantService {
				tenants := []*model.BusinessTenantMapping{
					newModelBusinessTenantMapping("t1", "tenant1"),
				}
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return tenantSvc
			},
			setupTemplateRendererSvc: func(_ []systemfetcher.System, _ []model.ApplicationRegisterInput) *automock.TemplateRenderer {
				return &automock.TemplateRenderer{}
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				return &automock.SystemsService{}
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return(nil, errors.New("expected")).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Fail when service fails to save systems",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(1)
				persistTx := &pAutomock.PersistenceTx{}

				transactioner.On("Begin").Return(persistTx, nil).Once()
				persistTx.On("Commit").Return(nil).Once()
				transactioner.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateID = appTemplateID
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(systems)
			},
			setupTenantSvc: func() *automock.TenantService {
				tenants := []*model.BusinessTenantMapping{
					newModelBusinessTenantMapping("t1", "tenant1"),
				}
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return tenantSvc
			},
			setupTemplateRendererSvc: setupSuccessfulTemplateRenderer,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), appsInputs[0].ApplicationRegisterInput, mock.Anything).Return(errors.New("expected")).Once()
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[0]}, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Fail when application from template cannot be rendered",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
				persistTx := &pAutomock.PersistenceTx{}

				transactioner.On("Begin").Return(persistTx, nil).Once()
				transactioner.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateID = appTemplateID
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(systems)
			},
			setupTenantSvc: func() *automock.TenantService {
				tenants := []*model.BusinessTenantMapping{
					newModelBusinessTenantMapping("t1", "tenant1"),
				}
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return tenantSvc
			},
			setupTemplateRendererSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInput) *automock.TemplateRenderer {
				svc := &automock.TemplateRenderer{}
				for i := range appsInputs {
					svc.On("ApplicationRegisterInputFromTemplate", txtest.CtxWithDBMatcher(), systems[i]).Return(nil, testErr).Once()
				}
				return svc
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				return &automock.SystemsService{}
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[0]}, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
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
				return fixAppsInputsWithTemplatesBySystems(systems)
			},
			setupTenantSvc: func() *automock.TenantService {
				tenants := []*model.BusinessTenantMapping{
					newModelBusinessTenantMapping("t1", "tenant1"),
					newModelBusinessTenantMapping("t2", "tenant2"),
					newModelBusinessTenantMapping("t3", "tenant3"),
				}
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return tenantSvc
			},
			setupTemplateRendererSvc: setupSuccessfulTemplateRenderer,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), mock.AnythingOfType("model.ApplicationRegisterInput"), mock.Anything).Return(nil).Once()
				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), mock.AnythingOfType("model.ApplicationRegisterInput"), mock.Anything).Return(nil).Once()
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[0]}, nil).Once()
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return(nil, errors.New("expected")).Once()
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "external").Return([]systemfetcher.System{testSystems[1]}, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Do nothing if system is already being deleted",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
				persistTx := &pAutomock.PersistenceTx{}

				transactioner.On("Begin").Return(persistTx, nil).Twice()
				persistTx.On("Commit").Return(nil).Once()
				transactioner.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Twice()

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
				return fixAppsInputsWithTemplatesBySystems(systems)
			},
			setupTenantSvc: func() *automock.TenantService {
				firstTenant := newModelBusinessTenantMapping("t1", "tenant1")
				firstTenant.ExternalTenant = "t1"
				secondTenant := newModelBusinessTenantMapping("t2", "tenant2")
				secondTenant.ExternalTenant = "t2"
				tenants := []*model.BusinessTenantMapping{firstTenant, secondTenant}
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return tenantSvc
			},
			setupTemplateRendererSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInput) *automock.TemplateRenderer {
				svc := &automock.TemplateRenderer{}
				appInput := appsInputs[0] // appsInputs[1] belongs to a system with status "DELETED"
				svc.On("ApplicationRegisterInputFromTemplate", txtest.CtxWithDBMatcher(), systems[0]).Return(&appInput, nil)
				return svc
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}

				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), mock.AnythingOfType("model.ApplicationRegisterInput"), mock.Anything).Return(nil).Once()
				systemSvc.On("GetByNameAndSystemNumber", txtest.CtxWithDBMatcher(), appsInputs[1].Name, *appsInputs[1].SystemNumber).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: "id",
					},
				}, nil)
				return systemSvc
			},
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "t1").Return([]systemfetcher.System{testSystems[0]}, nil).Once()
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, "t2").Return([]systemfetcher.System{testSystems[1]}, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				directorClient := &automock.DirectorClient{}
				directorClient.On("DeleteSystemAsync", mock.Anything, "id", "t2").Return(nil).Once()
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
			appInputsWithoutTemplates := make([]model.ApplicationRegisterInput, 0)
			for _, in := range appsInputs {
				appInputsWithoutTemplates = append(appInputsWithoutTemplates, in.ApplicationRegisterInput)
			}
			templateAppResolver := testCase.setupTemplateRendererSvc(testSystems, appInputsWithoutTemplates)
			sysAPIClient := testCase.setupSysAPIClient(testSystems)
			directorClient := testCase.setupDirectorClient(testSystems, appsInputs)
			defer mock.AssertExpectationsForObjects(t, tenantSvc, sysAPIClient, systemSvc, templateAppResolver, mockedTx, transactioner)

			svc := systemfetcher.NewSystemFetcher(transactioner, tenantSvc, systemSvc, templateAppResolver, sysAPIClient, directorClient, systemfetcher.Config{
				SystemsQueueSize:     100,
				FetcherParallellism:  30,
				EnableSystemDeletion: true,
			})
			err := svc.SyncSystems(context.TODO())
			if testCase.expectedErr != nil {
				require.ErrorIs(t, err, testCase.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func fixAppsInputsWithTemplatesBySystems(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
	initStatusCond := model.ApplicationStatusConditionInitial
	result := make([]model.ApplicationRegisterInputWithTemplate, 0, len(systems))
	for i := range systems {
		baseURL := systems[i].AdditionalURLs[mainURLKey]
		input := model.ApplicationRegisterInputWithTemplate{
			ApplicationRegisterInput: model.ApplicationRegisterInput{
				Name:            systems[i].DisplayName,
				Description:     &systems[i].ProductDescription,
				BaseURL:         &baseURL,
				ProviderName:    &systems[i].InfrastructureProvider,
				SystemNumber:    &systems[i].SystemNumber,
				StatusCondition: &initStatusCond,
				Labels: map[string]interface{}{
					"managed":              "true",
					"productId":            &systems[i].ProductID,
					"ppmsProductVersionId": &systems[i].PpmsProductVersionID,
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
