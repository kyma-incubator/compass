package systemfetcher_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher/automock"
	pAutomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	appType              = "type-1"
	mainURLKey           = "mainUrl"
	testTenantID         = "t1"
	testTenantExternalID = "et1"
	testTenantName       = "tenant1"
	testProductID        = "product1"
	appTemplateID        = "appTmp1"
)

var (
	testErr                           = errors.New("testErr")
	testTenant                        = newModelBusinessTenantMapping(testTenantID, testTenantExternalID, testTenantName)
	sfSystemSynchronizationTimestamps = map[string]systemfetcher.SystemSynchronizationTimestamp{
		"product1": {
			ID:                "time",
			LastSyncTimestamp: time.Date(2023, 5, 2, 20, 30, 0, 0, time.UTC).UTC(),
		},
	}
	modelSystemSynchronizationTimestamps = []*model.SystemSynchronizationTimestamp{
		{
			ID:                "time",
			TenantID:          testTenantID,
			ProductID:         testProductID,
			LastSyncTimestamp: time.Date(2023, 5, 2, 20, 30, 0, 0, time.UTC).UTC(),
		},
	}
)

func TestSyncSystems(t *testing.T) {
	type testCase struct {
		name                     string
		mockTransactioner        func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner)
		fixTestSystems           func() []systemfetcher.System
		fixAppInputs             func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate
		setupTenantSvc           func() *automock.TenantService
		setupTemplateRendererSvc func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInput) *automock.TemplateRenderer
		setupSystemSvc           func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService
		setupSystemsSyncSvc      func() *automock.SystemsSyncService
		setupSysAPIClient        func(testSystems []systemfetcher.System) *automock.SystemsAPIClient
		setupDirectorClient      func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient
		verificationTenant       string
		expectedErr              error
	}
	tests := []testCase{
		{
			name: "Success with one tenant and one system",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(4)
				return mockedTx, transactioner
			},
			fixTestSystems: fixSingleTestSystems,
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(t, systems)
			},
			setupTenantSvc:           okTenantSvc,
			setupTemplateRendererSvc: okTenantTemplateRendererSvc,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), appsInputs[0].ApplicationRegisterInput, mock.Anything).Return(nil).Once()
				systemSvc.On("GetBySystemNumber", txtest.CtxWithDBMatcher(), *appsInputs[0].SystemNumber).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: "id",
					},
				}, nil)
				return systemSvc
			},
			setupSystemsSyncSvc: okTenantSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, testTenant, sfSystemSynchronizationTimestamps).Return(testSystems, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Success with one tenant and one system which has tbt and tbt does not exist in the db",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(4)
				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystemsWithTbt()
				systems[0].TemplateID = appTemplateID
				systems[0].SystemPayload["productId"] = testProductID
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(t, systems)
			},
			setupTenantSvc:           okTenantSvc,
			setupTemplateRendererSvc: okTenantTemplateRendererSvc,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), appsInputs[0].ApplicationRegisterInput, mock.Anything).Return(nil).Once()
				systemSvc.On("GetBySystemNumber", txtest.CtxWithDBMatcher(), *appsInputs[0].SystemNumber).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: "id",
					},
				}, nil)
				return systemSvc
			},
			setupSystemsSyncSvc: okTenantSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, testTenant, sfSystemSynchronizationTimestamps).Return(testSystems, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Success with one tenant and one system which has tbt and tbt exists in the db",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(4)
				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystemsWithTbt()
				systems[0].TemplateID = appTemplateID
				systems[0].SystemPayload["productId"] = testProductID
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(t, systems)
			},
			setupTenantSvc:           okTenantSvc,
			setupTemplateRendererSvc: okTenantTemplateRendererSvc,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), appsInputs[0].ApplicationRegisterInput, mock.Anything).Return(nil).Once()
				systemSvc.On("GetBySystemNumber", txtest.CtxWithDBMatcher(), *appsInputs[0].SystemNumber).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: "id",
					},
				}, nil)
				return systemSvc
			},
			setupSystemsSyncSvc: okTenantSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, testTenant, sfSystemSynchronizationTimestamps).Return(testSystems, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Success when in verification mode",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(4)
				return mockedTx, transactioner
			},
			fixTestSystems: fixSingleTestSystems,
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(t, systems)
			},
			setupTenantSvc:           okTenantSvc,
			setupTemplateRendererSvc: okTenantTemplateRendererSvc,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), appsInputs[0].ApplicationRegisterInput, mock.Anything).Return(nil).Once()
				systemSvc.On("GetBySystemNumber", txtest.CtxWithDBMatcher(), *appsInputs[0].SystemNumber).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: "id",
					},
				}, nil)
				return systemSvc
			},
			setupSystemsSyncSvc: okTenantSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, testTenant, sfSystemSynchronizationTimestamps).Return(testSystems, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
			verificationTenant: "t1",
		},
		{
			name: "Success with one tenant and one system that has already been in the database and will not have it's status condition changed",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(4)
				return mockedTx, transactioner
			},
			fixTestSystems: fixSingleTestSystems,
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(t, systems)
			},
			setupTenantSvc: okTenantSvc,
			setupTemplateRendererSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInput) *automock.TemplateRenderer {
				svc := &automock.TemplateRenderer{}
				connectedStatus := model.ApplicationStatusConditionConnected

				for i := range appsInputs {
					input := systems[i]
					input.StatusCondition = connectedStatus

					result := appsInputs[i]
					result.StatusCondition = &connectedStatus

					svc.On("ApplicationRegisterInputFromTemplate", txtest.CtxWithDBMatcher(), input).Return(&result, nil).Once()
				}
				return svc
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}

				connectedStatus := model.ApplicationStatusConditionConnected
				appInput := appsInputs[0].ApplicationRegisterInput
				appInput.StatusCondition = &connectedStatus

				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), appInput, mock.Anything).Return(nil).Once()
				systemSvc.On("GetBySystemNumber", txtest.CtxWithDBMatcher(), *appsInputs[0].SystemNumber).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: "id",
					},
					Status: &model.ApplicationStatus{
						Condition: model.ApplicationStatusConditionConnected,
					},
				}, nil)
				return systemSvc
			},
			setupSystemsSyncSvc: okTenantSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, testTenant, sfSystemSynchronizationTimestamps).Return(testSystems, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Success with one tenant and one system with null base url",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(4)
				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := []systemfetcher.System{
					{
						SystemPayload: map[string]interface{}{
							"displayName":            "System1",
							"productDescription":     "System1 description",
							"infrastructureProvider": "test",
						},
						StatusCondition: model.ApplicationStatusConditionInitial,
					},
				}
				systems[0].TemplateID = "type1"
				systems[0].SystemPayload["productId"] = testProductID
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(t, systems)
			},
			setupTenantSvc:           okTenantSvc,
			setupTemplateRendererSvc: okTenantTemplateRendererSvc,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), mock.AnythingOfType("model.ApplicationRegisterInput"), mock.Anything).Return(nil).Once()
				systemSvc.On("GetBySystemNumber", txtest.CtxWithDBMatcher(), *appsInputs[0].SystemNumber).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: "id",
					},
				}, nil)
				return systemSvc
			},
			setupSystemsSyncSvc: okTenantSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, testTenant, sfSystemSynchronizationTimestamps).Return(testSystems, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Success with one tenant and one system without template",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(4)
				return mockedTx, transactioner
			},
			fixTestSystems: fixSystems,
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(t, systems)
			},
			setupTenantSvc: okTenantSvc,
			setupTemplateRendererSvc: func(_ []systemfetcher.System, _ []model.ApplicationRegisterInput) *automock.TemplateRenderer {
				return &automock.TemplateRenderer{}
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("TrustedUpsert", txtest.CtxWithDBMatcher(), appsInputs[0].ApplicationRegisterInput, mock.Anything).Return(nil).Once()
				systemSvc.On("GetBySystemNumber", txtest.CtxWithDBMatcher(), *appsInputs[0].SystemNumber).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: "id",
					},
				}, nil)
				return systemSvc
			},
			setupSystemsSyncSvc: okTenantSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, testTenant, sfSystemSynchronizationTimestamps).Return(testSystems, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Success with one tenant and multiple systems",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(5)
				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateID = appTemplateID
				systems = append(systems, systemfetcher.System{
					SystemPayload: map[string]interface{}{
						"displayName":            "System2",
						"productDescription":     "System2 description",
						"baseUrl":                "http://example2.com",
						"infrastructureProvider": "test",
					},
					TemplateID:      "appTmp2",
					StatusCondition: model.ApplicationStatusConditionInitial,
				})
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(t, systems)
			},
			setupTenantSvc:           okTenantSvc,
			setupTemplateRendererSvc: okTenantTemplateRendererSvc,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), appsInputs[0].ApplicationRegisterInput, mock.Anything).Return(nil).Once()
				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), appsInputs[1].ApplicationRegisterInput, mock.Anything).Return(nil).Once()
				systemSvc.On("GetBySystemNumber", txtest.CtxWithDBMatcher(), *appsInputs[0].SystemNumber).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: "id",
					},
				}, nil)
				systemSvc.On("GetBySystemNumber", txtest.CtxWithDBMatcher(), *appsInputs[1].SystemNumber).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: "id",
					},
				}, nil)
				return systemSvc
			},
			setupSystemsSyncSvc: okTenantSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, testTenant, sfSystemSynchronizationTimestamps).Return(testSystems, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Fail when transaction cannot be started",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(testErr).ThatFailsOnBegin()

				return mockedTx, transactioner
			},
			fixTestSystems: fixSystems,
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(t, systems)
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
			setupSystemsSyncSvc: okTenantSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				return &automock.SystemsAPIClient{}
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
			expectedErr: testErr,
		},
		{
			name: "Fail when Timestamps for tenant cannot be fetched",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				persistTx := &pAutomock.PersistenceTx{}

				transact := &pAutomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return persistTx, transact
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateID = appTemplateID
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(t, systems)
			},
			setupTenantSvc: func() *automock.TenantService {
				return &automock.TenantService{}
			},
			setupTemplateRendererSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInput) *automock.TemplateRenderer {
				return &automock.TemplateRenderer{}
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				return &automock.SystemsService{}
			},
			setupSystemsSyncSvc: errListByTenantSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				return &automock.SystemsAPIClient{}
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
			expectedErr: testErr,
		},
		{
			name: "Fail when Tenant cannot be fetched",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				persistTx := &pAutomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &pAutomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateID = appTemplateID
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(t, systems)
			},
			setupTenantSvc: errTenantSvc,
			setupTemplateRendererSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInput) *automock.TemplateRenderer {
				return &automock.TemplateRenderer{}
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				return &automock.SystemsService{}
			},
			setupSystemsSyncSvc: okTenantSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				return &automock.SystemsAPIClient{}
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
			expectedErr: testErr,
		},
		{
			name: "Succeed when Tenant cannot be found",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				persistTx := &pAutomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Twice()

				transact := &pAutomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return persistTx, transact
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateID = appTemplateID
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(t, systems)
			},
			setupTenantSvc: okNilTenantSvc,
			setupTemplateRendererSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInput) *automock.TemplateRenderer {
				return &automock.TemplateRenderer{}
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				return &automock.SystemsService{}
			},
			setupSystemsSyncSvc: okTenantSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				return &automock.SystemsAPIClient{}
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Fail when client fails to fetch systems",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsTwice()

				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				return fixSystems()
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(t, systems)
			},
			setupTenantSvc: okTenantSvc,
			setupTemplateRendererSvc: func(_ []systemfetcher.System, _ []model.ApplicationRegisterInput) *automock.TemplateRenderer {
				return &automock.TemplateRenderer{}
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				return &automock.SystemsService{}
			},
			setupSystemsSyncSvc: okTenantSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, testTenant, sfSystemSynchronizationTimestamps).Return(nil, testErr).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
			expectedErr: testErr,
		},
		{
			name: "Succeed when Upsert Timestamps fails",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(2)
				persistTx := &pAutomock.PersistenceTx{}

				transactioner.On("Begin").Return(persistTx, nil).Twice()
				persistTx.On("Commit").Return(nil).Once()
				persistTx.On("Commit").Return(testErr).Once()
				transactioner.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				transactioner.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return mockedTx, transactioner
			},
			fixTestSystems: fixSingleTestSystems,
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(t, systems)
			},
			setupTenantSvc:           okTenantSvc,
			setupTemplateRendererSvc: okTenantTemplateRendererSvc,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), appsInputs[0].ApplicationRegisterInput, mock.Anything).Return(nil).Once()
				systemSvc.On("GetBySystemNumber", txtest.CtxWithDBMatcher(), *appsInputs[0].SystemNumber).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: "id",
					},
				}, nil)
				return systemSvc
			},
			setupSystemsSyncSvc: errUpsertTenantSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, testTenant, sfSystemSynchronizationTimestamps).Return(testSystems, nil).Once()
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

				transactioner.On("Begin").Return(persistTx, nil).Twice()
				persistTx.On("Commit").Return(nil).Twice()
				transactioner.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Twice()

				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateID = appTemplateID
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(t, systems)
			},
			setupTenantSvc:           okTenantSvc,
			setupTemplateRendererSvc: okTenantTemplateRendererSvc,
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), appsInputs[0].ApplicationRegisterInput, mock.Anything).Return(testErr).Once()
				systemSvc.On("GetBySystemNumber", txtest.CtxWithDBMatcher(), *appsInputs[0].SystemNumber).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: "id",
					},
				}, nil)
				return systemSvc
			},
			setupSystemsSyncSvc: okTenantSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, testTenant, sfSystemSynchronizationTimestamps).Return([]systemfetcher.System{testSystems[0]}, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
			expectedErr: testErr,
		},
		{
			name: "Fail when application from template cannot be rendered",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(1)
				persistTx := &pAutomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transactioner.On("Begin").Return(persistTx, nil).Twice()
				transactioner.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				transactioner.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				systems := fixSystems()
				systems[0].TemplateID = appTemplateID
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(t, systems)
			},
			setupTenantSvc: okTenantSvc,
			setupTemplateRendererSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInput) *automock.TemplateRenderer {
				svc := &automock.TemplateRenderer{}
				for i := range appsInputs {
					svc.On("ApplicationRegisterInputFromTemplate", txtest.CtxWithDBMatcher(), systems[i]).Return(nil, testErr).Once()
				}
				return svc
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}
				systemSvc.On("GetBySystemNumber", txtest.CtxWithDBMatcher(), *appsInputs[0].SystemNumber).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: "id",
					},
				}, nil)

				return systemSvc
			},
			setupSystemsSyncSvc: okTenantSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, testTenant, sfSystemSynchronizationTimestamps).Return([]systemfetcher.System{testSystems[0]}, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
			expectedErr: testErr,
		},
		{
			name: "Do nothing if system is already being deleted",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(3)
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
					SystemPayload: map[string]interface{}{
						"displayName":            "System2",
						"productDescription":     "System2 description",
						"baseUrl":                "http://example2.com",
						"infrastructureProvider": "test",
						"additionalAttributes": map[string]string{
							systemfetcher.LifecycleAttributeName: systemfetcher.LifecycleDeleted,
						},
						"systemNumber": "sysNumber1",
					},
					TemplateID:      "type2",
					StatusCondition: model.ApplicationStatusConditionInitial,
				})
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(t, systems)
			},
			setupTenantSvc: okTenantSvc,
			setupTemplateRendererSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInput) *automock.TemplateRenderer {
				svc := &automock.TemplateRenderer{}
				appInput := appsInputs[0] // appsInputs[1] belongs to a system with status "DELETED"
				svc.On("ApplicationRegisterInputFromTemplate", txtest.CtxWithDBMatcher(), systems[0]).Return(&appInput, nil)
				return svc
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}

				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), mock.AnythingOfType("model.ApplicationRegisterInput"), mock.Anything).Return(nil).Once()
				systemSvc.On("GetBySystemNumber", txtest.CtxWithDBMatcher(), *appsInputs[0].SystemNumber).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: "id",
					},
				}, nil)
				systemSvc.On("GetBySystemNumber", txtest.CtxWithDBMatcher(), *appsInputs[1].SystemNumber).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: "id",
					},
				}, nil)
				return systemSvc
			},
			setupSystemsSyncSvc: okTenantSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, testTenant, sfSystemSynchronizationTimestamps).Return([]systemfetcher.System{testSystems[0], testSystems[1]}, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				directorClient := &automock.DirectorClient{}
				directorClient.On("DeleteSystemAsync", mock.Anything, "id", testTenantID).Return(nil).Once()
				return directorClient
			},
		},
		{
			name: "Do nothing if system has already been deleted",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(3)
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
					SystemPayload: map[string]interface{}{
						"displayName":            "System2",
						"productDescription":     "System2 description",
						"baseUrl":                "http://example2.com",
						"infrastructureProvider": "test",
						"additionalAttributes": map[string]string{
							systemfetcher.LifecycleAttributeName: systemfetcher.LifecycleDeleted,
						},
						"systemNumber": "sysNumber1",
					},
					TemplateID:      "type2",
					StatusCondition: model.ApplicationStatusConditionInitial,
				})
				return systems
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return fixAppsInputsWithTemplatesBySystems(t, systems)
			},
			setupTenantSvc: okTenantSvc,
			setupTemplateRendererSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInput) *automock.TemplateRenderer {
				svc := &automock.TemplateRenderer{}
				appInput := appsInputs[0] // appsInputs[1] belongs to a system with status "DELETED"
				svc.On("ApplicationRegisterInputFromTemplate", txtest.CtxWithDBMatcher(), systems[0]).Return(&appInput, nil)
				return svc
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				systemSvc := &automock.SystemsService{}

				systemSvc.On("TrustedUpsertFromTemplate", txtest.CtxWithDBMatcher(), mock.AnythingOfType("model.ApplicationRegisterInput"), mock.Anything).Return(nil).Once()
				systemSvc.On("GetBySystemNumber", txtest.CtxWithDBMatcher(), *appsInputs[0].SystemNumber).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID: "id",
					},
				}, nil)
				systemSvc.On("GetBySystemNumber", txtest.CtxWithDBMatcher(), *appsInputs[1].SystemNumber).Return(nil, nil)
				return systemSvc
			},
			setupSystemsSyncSvc: okTenantSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				sysAPIClient := &automock.SystemsAPIClient{}
				sysAPIClient.On("FetchSystemsForTenant", mock.Anything, testTenant, sfSystemSynchronizationTimestamps).Return([]systemfetcher.System{testSystems[0], testSystems[1]}, nil).Once()
				return sysAPIClient
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				directorClient := &automock.DirectorClient{}
				directorClient.On("DeleteSystemAsync", mock.Anything, "id", testTenantID).Return(nil).Once()
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
			systemsSyncSvc := testCase.setupSystemsSyncSvc()
			appInputsWithoutTemplates := make([]model.ApplicationRegisterInput, 0)
			for _, in := range appsInputs {
				appInputsWithoutTemplates = append(appInputsWithoutTemplates, in.ApplicationRegisterInput)
			}
			templateAppResolver := testCase.setupTemplateRendererSvc(testSystems, appInputsWithoutTemplates)
			sysAPIClient := testCase.setupSysAPIClient(testSystems)
			directorClient := testCase.setupDirectorClient(testSystems, appsInputs)
			defer mock.AssertExpectationsForObjects(t, tenantSvc, sysAPIClient, systemSvc, templateAppResolver, mockedTx, transactioner)

			svc := systemfetcher.NewSystemFetcher(transactioner, tenantSvc, systemSvc, systemsSyncSvc, templateAppResolver, sysAPIClient, directorClient, systemfetcher.Config{
				EnableSystemDeletion: true,
				VerifyTenant:         testCase.verificationTenant,
			})

			err := svc.ProcessTenant(context.TODO(), testTenantID)
			if testCase.expectedErr != nil {
				require.ErrorIs(t, err, testCase.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpsertSystemsSyncTimestamps(t *testing.T) {
	type testCase struct {
		name                     string
		mockTransactioner        func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner)
		fixTestSystems           func() []systemfetcher.System
		fixAppInputs             func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate
		setupTenantSvc           func() *automock.TenantService
		setupTemplateRendererSvc func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInput) *automock.TemplateRenderer
		setupSystemSvc           func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService
		setupSystemsSyncSvc      func() *automock.SystemsSyncService
		setupSysAPIClient        func(testSystems []systemfetcher.System) *automock.SystemsAPIClient
		setupDirectorClient      func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient
		verificationTenant       string
		expectedErr              error
	}

	tests := []testCase{
		{
			name: "Success",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(1)
				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				return []systemfetcher.System{}
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return []model.ApplicationRegisterInputWithTemplate{}
			},
			setupTenantSvc: func() *automock.TenantService {
				return &automock.TenantService{}
			},
			setupTemplateRendererSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInput) *automock.TemplateRenderer {
				return &automock.TemplateRenderer{}
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				return &automock.SystemsService{}
			},
			setupSystemsSyncSvc: okSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				return &automock.SystemsAPIClient{}
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
		},
		{
			name: "Error while upserting",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
				return mockedTx, transactioner
			},
			fixTestSystems: func() []systemfetcher.System {
				return []systemfetcher.System{}
			},
			fixAppInputs: func(systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
				return []model.ApplicationRegisterInputWithTemplate{}
			},
			setupTenantSvc: func() *automock.TenantService {
				return &automock.TenantService{}
			},
			setupTemplateRendererSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInput) *automock.TemplateRenderer {
				return &automock.TemplateRenderer{}
			},
			setupSystemSvc: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.SystemsService {
				return &automock.SystemsService{}
			},
			setupSystemsSyncSvc: errSystemsSyncSvc,
			setupSysAPIClient: func(testSystems []systemfetcher.System) *automock.SystemsAPIClient {
				return &automock.SystemsAPIClient{}
			},
			setupDirectorClient: func(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInputWithTemplate) *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
			expectedErr: testErr,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mockedTx, transactioner := testCase.mockTransactioner()
			tenantSvc := testCase.setupTenantSvc()
			testSystems := testCase.fixTestSystems()
			appsInputs := testCase.fixAppInputs(testSystems)
			systemSvc := testCase.setupSystemSvc(testSystems, appsInputs)
			systemsSyncSvc := testCase.setupSystemsSyncSvc()
			appInputsWithoutTemplates := make([]model.ApplicationRegisterInput, 0)
			for _, in := range appsInputs {
				appInputsWithoutTemplates = append(appInputsWithoutTemplates, in.ApplicationRegisterInput)
			}
			templateAppResolver := testCase.setupTemplateRendererSvc(testSystems, appInputsWithoutTemplates)
			sysAPIClient := testCase.setupSysAPIClient(testSystems)
			directorClient := testCase.setupDirectorClient(testSystems, appsInputs)
			defer mock.AssertExpectationsForObjects(t, tenantSvc, sysAPIClient, systemSvc, templateAppResolver, mockedTx, transactioner)

			svc := systemfetcher.NewSystemFetcher(transactioner, tenantSvc, systemSvc, systemsSyncSvc, templateAppResolver, sysAPIClient, directorClient, systemfetcher.Config{
				EnableSystemDeletion: true,
				VerifyTenant:         testCase.verificationTenant,
			})

			err := svc.UpsertSystemsSyncTimestampsForTenant(context.TODO(), testTenantID, sfSystemSynchronizationTimestamps)
			if testCase.expectedErr != nil {
				require.ErrorIs(t, err, testCase.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func fixAppsInputsWithTemplatesBySystems(t *testing.T, systems []systemfetcher.System) []model.ApplicationRegisterInputWithTemplate {
	initStatusCond := model.ApplicationStatusConditionInitial
	result := make([]model.ApplicationRegisterInputWithTemplate, 0, len(systems))
	for _, s := range systems {
		systemPayload, err := json.Marshal(s.SystemPayload)
		require.NoError(t, err)
		input := model.ApplicationRegisterInputWithTemplate{
			ApplicationRegisterInput: model.ApplicationRegisterInput{
				Name:            gjson.GetBytes(systemPayload, "displayName").String(),
				Description:     str.Ptr(gjson.GetBytes(systemPayload, "productDescription").String()),
				BaseURL:         str.Ptr(gjson.GetBytes(systemPayload, "additionalUrls"+"."+mainURLKey).String()),
				ProviderName:    str.Ptr(gjson.GetBytes(systemPayload, "infrastructureProvider").String()),
				SystemNumber:    str.Ptr(gjson.GetBytes(systemPayload, "systemNumber").String()),
				StatusCondition: &initStatusCond,
				Labels: map[string]interface{}{
					"managed":              "true",
					"productId":            str.Ptr(gjson.GetBytes(systemPayload, "productId").String()),
					"ppmsProductVersionId": str.Ptr(gjson.GetBytes(systemPayload, "ppmsProductVersionId").String()),
				},
			},
			TemplateID: s.TemplateID,
		}
		if len(input.TemplateID) > 0 {
			input.Labels["applicationType"] = appType
		}
		result = append(result, input)
	}
	return result
}

func fixSingleTestSystems() []systemfetcher.System {
	systems := fixSystems()
	systems[0].TemplateID = appTemplateID
	systems[0].SystemPayload["productId"] = testProductID
	return systems
}

func okTenantSystemsSyncSvc() *automock.SystemsSyncService {
	syncMock := &automock.SystemsSyncService{}
	syncMock.On("ListByTenant", mock.Anything, testTenantID).Return(modelSystemSynchronizationTimestamps, nil)
	syncMock.On("Upsert", mock.Anything, mock.Anything).Return(nil)
	return syncMock
}

func errListByTenantSystemsSyncSvc() *automock.SystemsSyncService {
	syncMock := &automock.SystemsSyncService{}
	syncMock.On("ListByTenant", mock.Anything, testTenantID).Return(nil, testErr)
	return syncMock
}
func errUpsertTenantSystemsSyncSvc() *automock.SystemsSyncService {
	syncMock := &automock.SystemsSyncService{}
	syncMock.On("ListByTenant", mock.Anything, testTenantID).Return(modelSystemSynchronizationTimestamps, nil)
	syncMock.On("Upsert", mock.Anything, mock.Anything).Return(testErr)
	return syncMock
}

func okTenantSvc() *automock.TenantService {
	tenantSvc := &automock.TenantService{}
	tenantSvc.On("GetTenantByID", mock.Anything, testTenantID).Return(testTenant, nil).Once()
	return tenantSvc
}

func errTenantSvc() *automock.TenantService {
	tenantSvc := &automock.TenantService{}
	tenantSvc.On("GetTenantByID", mock.Anything, testTenantID).Return(nil, testErr).Once()
	return tenantSvc
}

func okNilTenantSvc() *automock.TenantService {
	tenantSvc := &automock.TenantService{}
	tenantSvc.On("GetTenantByID", mock.Anything, testTenantID).Return(nil, nil).Once()
	return tenantSvc
}

func okTenantTemplateRendererSvc(systems []systemfetcher.System, appsInputs []model.ApplicationRegisterInput) *automock.TemplateRenderer {
	ttrSvc := &automock.TemplateRenderer{}
	for i := range appsInputs {
		ttrSvc.On("ApplicationRegisterInputFromTemplate", mock.Anything, systems[i]).Return(&appsInputs[i], nil).Once()
	}
	return ttrSvc
}

func okSystemsSyncSvc() *automock.SystemsSyncService {
	syncMock := &automock.SystemsSyncService{}
	syncMock.On("Upsert", mock.Anything, mock.Anything).Return(nil)
	return syncMock
}

func errSystemsSyncSvc() *automock.SystemsSyncService {
	syncMock := &automock.SystemsSyncService{}
	syncMock.On("Upsert", mock.Anything, mock.Anything).Return(testErr)
	return syncMock
}
