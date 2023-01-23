package apptemplate_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/tests/pkg/webhook"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/selfregmanager"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate/apptmpltest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate/automock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	RegionKey = "region"
)

func TestResolver_ApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	txGen := txtest.NewTransactionContextGenerator(testError)

	modelAppTemplate := fixModelApplicationTemplate(testID, testName, fixModelApplicationTemplateWebhooks(testWebhookID, testID))
	gqlAppTemplate := fixGQLAppTemplate(testID, testName, fixGQLApplicationTemplateWebhooks(testWebhookID, testID))

	testCases := []struct {
		Name              string
		TxFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AppTemplateSvcFn  func() *automock.ApplicationTemplateService
		AppTemplateConvFn func() *automock.ApplicationTemplateConverter
		WebhookSvcFn      func() *automock.WebhookService
		WebhookConvFn     func() *automock.WebhookConverter
		ExpectedOutput    *graphql.ApplicationTemplate
		ExpectedError     error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplate, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: gqlAppTemplate,
		},
		{
			Name: "Returns nil when application template not found",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, apperrors.NewNotFoundError(resource.ApplicationTemplate, "")).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: UnusedAppTemplateConv,
			WebhookConvFn:     UnusedWebhookConv,
			WebhookSvcFn:      UnusedWebhookSvc,
			ExpectedOutput:    nil,
		},
		{
			Name: "Returns error when getting application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: UnusedAppTemplateConv,
			WebhookConvFn:     UnusedWebhookConv,
			WebhookSvcFn:      UnusedWebhookSvc,
			ExpectedError:     testError,
		},
		{
			Name:              "Returns error when beginning transaction",
			TxFn:              txGen.ThatFailsOnBegin,
			AppTemplateSvcFn:  UnusedAppTemplateSvc,
			AppTemplateConvFn: UnusedAppTemplateConv,
			WebhookConvFn:     UnusedWebhookConv,
			WebhookSvcFn:      UnusedWebhookSvc,
			ExpectedError:     testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: UnusedAppTemplateConv,
			WebhookConvFn:     UnusedWebhookConv,
			WebhookSvcFn:      UnusedWebhookSvc,
			ExpectedError:     testError,
		},
		{
			Name: "Returns error when can't convert application template to graphql",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(nil, testError).Once()
				return appTemplateConv
			},
			WebhookConvFn: UnusedWebhookConv,
			WebhookSvcFn:  UnusedWebhookSvc,
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			appTemplateSvc := testCase.AppTemplateSvcFn()
			appTemplateConv := testCase.AppTemplateConvFn()
			webhookSvc := testCase.WebhookSvcFn()
			webhookConverter := testCase.WebhookConvFn()

			resolver := apptemplate.NewResolver(transact, nil, nil, appTemplateSvc, appTemplateConv, webhookSvc, webhookConverter, nil, nil, nil, "")

			// WHEN
			result, err := resolver.ApplicationTemplate(ctx, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			appTemplateSvc.AssertExpectations(t)
			appTemplateConv.AssertExpectations(t)
		})
	}
}

func TestResolver_ApplicationTemplates(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)
	txGen := txtest.NewTransactionContextGenerator(testError)
	modelAppTemplates := []*model.ApplicationTemplate{
		fixModelApplicationTemplate("i1", "n1", fixModelApplicationTemplateWebhooks("webhook-id-1", "i1")),
		fixModelApplicationTemplate("i2", "n2", fixModelApplicationTemplateWebhooks("webhook-id-2", "i2")),
	}
	modelPage := fixModelAppTemplatePage(modelAppTemplates)
	gqlAppTemplates := []*graphql.ApplicationTemplate{
		fixGQLAppTemplate("i1", "n1", fixGQLApplicationTemplateWebhooks("webhook-id-1", "i1")),
		fixGQLAppTemplate("i2", "n2", fixGQLApplicationTemplateWebhooks("webhook-id-2", "i2")),
	}
	gqlPage := fixGQLAppTemplatePage(gqlAppTemplates)
	first := 2
	after := "test"
	gqlAfter := graphql.PageCursor(after)

	labelFilters := []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(RegionKey, "eu-1")}
	labelFiltersEmpty := []*labelfilter.LabelFilter{}
	gqlFilter := []*graphql.LabelFilter{
		{Key: RegionKey, Query: str.Ptr("eu-1")},
	}

	testCases := []struct {
		Name              string
		LabelFilter       []*graphql.LabelFilter
		TxFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AppTemplateSvcFn  func() *automock.ApplicationTemplateService
		AppTemplateConvFn func() *automock.ApplicationTemplateConverter
		WebhookSvcFn      func() *automock.WebhookService
		WebhookConvFn     func() *automock.WebhookConverter
		ExpectedOutput    *graphql.ApplicationTemplatePage
		ExpectedError     error
	}{
		{
			Name:        "Success",
			LabelFilter: gqlFilter,
			TxFn:        txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("List", txtest.CtxWithDBMatcher(), labelFilters, first, after).Return(modelPage, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("MultipleToGraphQL", modelAppTemplates).Return(gqlAppTemplates, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: &gqlPage,
		},
		{
			Name: "Returns error when getting application templates failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("List", txtest.CtxWithDBMatcher(), labelFiltersEmpty, first, after).Return(model.ApplicationTemplatePage{}, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: UnusedAppTemplateConv,
			WebhookConvFn:     UnusedWebhookConv,
			WebhookSvcFn:      UnusedWebhookSvc,
			ExpectedError:     testError,
		},
		{
			Name:              "Returns error when beginning transaction",
			TxFn:              txGen.ThatFailsOnBegin,
			AppTemplateSvcFn:  UnusedAppTemplateSvc,
			AppTemplateConvFn: UnusedAppTemplateConv,
			WebhookConvFn:     UnusedWebhookConv,
			WebhookSvcFn:      UnusedWebhookSvc,
			ExpectedError:     testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("List", txtest.CtxWithDBMatcher(), labelFiltersEmpty, first, after).Return(modelPage, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: UnusedAppTemplateConv,
			WebhookConvFn:     UnusedWebhookConv,
			WebhookSvcFn:      UnusedWebhookSvc,
			ExpectedError:     testError,
		},
		{
			Name: "Returns error when can't convert at least one of application templates to graphql",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("List", txtest.CtxWithDBMatcher(), labelFiltersEmpty, first, after).Return(modelPage, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("MultipleToGraphQL", modelAppTemplates).Return(nil, testError).Once()
				return appTemplateConv
			},
			WebhookConvFn: UnusedWebhookConv,
			WebhookSvcFn:  UnusedWebhookSvc,
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			appTemplateSvc := testCase.AppTemplateSvcFn()
			appTemplateConv := testCase.AppTemplateConvFn()
			webhookSvc := testCase.WebhookSvcFn()
			webhookConverter := testCase.WebhookConvFn()

			resolver := apptemplate.NewResolver(transact, nil, nil, appTemplateSvc, appTemplateConv, webhookSvc, webhookConverter, nil, nil, nil, "")

			// WHEN
			result, err := resolver.ApplicationTemplates(ctx, testCase.LabelFilter, &first, &gqlAfter)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			appTemplateSvc.AssertExpectations(t)
			appTemplateConv.AssertExpectations(t)
		})
	}
}

func TestResolver_Webhooks(t *testing.T) {
	// GIVEN
	applicationTemplateID := "fooid"
	modelWebhooks := fixModelApplicationTemplateWebhooks("test-webhook-1", applicationTemplateID)
	gqlWebhooks := fixGQLApplicationTemplateWebhooks("test-webhook-1", applicationTemplateID)

	appTemplate := fixGQLAppTemplate(applicationTemplateID, "foo", gqlWebhooks)
	testErr := errors.New("Test error")

	testCases := []struct {
		Name               string
		PersistenceFn      func() *persistenceautomock.PersistenceTx
		TransactionerFn    func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		WebhookServiceFn   func() *automock.WebhookService
		WebhookConverterFn func() *automock.WebhookConverter
		ExpectedResult     []*graphql.Webhook
		ExpectedErr        error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			WebhookServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), applicationTemplateID).Return(modelWebhooks, nil).Once()
				return svc
			},
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleToGraphQL", modelWebhooks).Return(gqlWebhooks, nil).Once()
				return conv
			},
			ExpectedResult: gqlWebhooks,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when webhook listing failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			WebhookServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), applicationTemplateID).Return(nil, testErr).Once()
				return svc
			},
			WebhookConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name: "Returns error on starting transaction",
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(nil, testErr).Once()
				return transact
			},
			PersistenceFn:      txtest.PersistenceContextThatDoesntExpectCommit,
			WebhookServiceFn:   UnusedWebhookSvc,
			WebhookConverterFn: UnusedWebhookConv,
			ExpectedErr:        testErr,
		},
		{
			Name: "Returns error on committing transaction",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(testErr).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			WebhookServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), applicationTemplateID).Return(modelWebhooks, nil).Once()
				return svc
			},
			WebhookConverterFn: UnusedWebhookConv,
			ExpectedErr:        testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			webhookSvc := testCase.WebhookServiceFn()
			converter := testCase.WebhookConverterFn()

			mockPersistence := testCase.PersistenceFn()
			mockTransactioner := testCase.TransactionerFn(mockPersistence)

			resolver := apptemplate.NewResolver(mockTransactioner, nil, nil, nil, nil, webhookSvc, converter, nil, nil, nil, "")

			// WHEN
			result, err := resolver.Webhooks(context.TODO(), appTemplate)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			webhookSvc.AssertExpectations(t)
			converter.AssertExpectations(t)
			mockPersistence.AssertExpectations(t)
			mockTransactioner.AssertExpectations(t)
		})
	}
}

func TestResolver_CreateApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	txGen := txtest.NewTransactionContextGenerator(testError)

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testUUID)
		return uidSvc
	}

	modelAppTemplate := fixModelApplicationTemplate(testID, testName, fixModelApplicationWebhooks(testWebhookID, testID))
	modelAppTemplateInput := fixModelAppTemplateInput(testName, appInputJSONString)
	modelAppTemplateInput.ID = &testUUID
	gqlAppTemplate := fixGQLAppTemplate(testID, testName, fixGQLApplicationTemplateWebhooks(testWebhookID, testID))
	gqlAppTemplateInput := fixGQLAppTemplateInputWithPlaceholder(testName)
	gqlAppTemplateInputWithProvider := fixGQLAppTemplateInputWithPlaceholderAndProvider("SAP " + testName)

	modelAppTemplateInputWithSelRegLabels := fixModelAppTemplateInput(testName, appInputJSONString)
	modelAppTemplateInputWithSelRegLabels.ID = &testUUID
	modelAppTemplateInputWithSelRegLabels.Labels = graphql.Labels{
		apptmpltest.TestDistinguishLabel: fmt.Sprintf("\"%s\"", "selfRegVal"),
		selfregmanager.RegionLabel:       fmt.Sprintf("\"%s\"", "region"),
	}
	gqlAppTemplateWithSelfRegLabels := fixGQLAppTemplate(testID, testName, fixGQLApplicationTemplateWebhooks(testWebhookID, testID))
	gqlAppTemplateWithSelfRegLabels.Labels = graphql.Labels{
		apptmpltest.TestDistinguishLabel: fmt.Sprintf("\"%s\"", "selfRegVal"),
		selfregmanager.RegionLabel:       fmt.Sprintf("\"%s\"", "region"),
	}
	gqlAppTemplateInputWithSelfRegLabels := fixGQLAppTemplateInputWithPlaceholder(testName)
	gqlAppTemplateInputWithSelfRegLabels.Labels = graphql.Labels{
		apptmpltest.TestDistinguishLabel: fmt.Sprintf("\"%s\"", "selfRegVal"),
		selfregmanager.RegionLabel:       fmt.Sprintf("\"%s\"", "region"),
	}

	gqlAppTemplateInputWithProviderAndWebhook := fixGQLAppTemplateInputWithPlaceholderAndProvider("SAP " + testName)
	gqlAppTemplateInputWithProviderAndWebhook.Webhooks = []*graphql.WebhookInput{
		{
			Type:    graphql.WebhookTypeConfigurationChanged,
			URL:     &testURL,
			Auth:    nil,
			Mode:    webhook.WebhookModePtr(graphql.WebhookModeSync),
			Version: str.Ptr("v1.0"),
		},
		{
			Type: graphql.WebhookTypeOpenResourceDiscovery,
			URL:  &testURL,
			Auth: nil,
		},
	}
	gqlAppTemplateInputWithProviderAndWebhookWithAsyncCallback := fixGQLAppTemplateInputWithPlaceholderAndProvider("SAP " + testName)
	gqlAppTemplateInputWithProviderAndWebhookWithAsyncCallback.Webhooks = []*graphql.WebhookInput{
		{
			Type:    graphql.WebhookTypeConfigurationChanged,
			URL:     &testURL,
			Auth:    nil,
			Mode:    webhook.WebhookModePtr(graphql.WebhookModeAsyncCallback),
			Version: str.Ptr("v1.0"),
		},
		{
			Type: graphql.WebhookTypeOpenResourceDiscovery,
			URL:  &testURL,
			Auth: nil,
		},
	}

	labels := map[string]interface{}{"cloneLabel": "clone"}
	labelsContainingSelfRegistration := map[string]interface{}{apptmpltest.TestDistinguishLabel: "selfRegVal", RegionKey: "region"}
	distinguishLabel := map[string]interface{}{apptmpltest.TestDistinguishLabel: "selfRegVal"}
	regionLabel := map[string]interface{}{RegionKey: "region"}
	badValueLabel := map[string]interface{}{RegionKey: 1}

	getAppTemplateFilters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(apptmpltest.TestDistinguishLabel, fmt.Sprintf("\"%s\"", "selfRegVal")),
		labelfilter.NewForKeyWithQuery(selfregmanager.RegionLabel, fmt.Sprintf("\"%s\"", "region")),
	}

	testCases := []struct {
		Name                  string
		TxFn                  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AppTemplateSvcFn      func() *automock.ApplicationTemplateService
		AppTemplateConvFn     func() *automock.ApplicationTemplateConverter
		WebhookSvcFn          func() *automock.WebhookService
		WebhookConvFn         func() *automock.WebhookConverter
		SelfRegManagerFn      func() *automock.SelfRegisterManager
		TenantMappingConfigFn func() map[string]interface{}
		Input                 *graphql.ApplicationTemplateInput
		ExpectedOutput        *graphql.ApplicationTemplate
		ExpectedError         error
	}{
		{
			Name: "Success",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false)

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInput, labels).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInputWithProvider).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplate, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn:         UnusedWebhookConv,
			WebhookSvcFn:          UnusedWebhookSvc,
			SelfRegManagerFn:      apptmpltest.SelfRegManagerThatDoesNotCleanupFunc(labels),
			TenantMappingConfigFn: EmptyTenantMappingConfig,
			Input:                 gqlAppTemplateInputWithProvider,
			ExpectedOutput:        gqlAppTemplate,
		},
		{
			Name: "Success with tenant mapping configuration",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false)

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInput, labels).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				expectedGqlAppTemplateInputWithProviderAndWebhook := fixGQLAppTemplateInputWithPlaceholderAndProvider("SAP " + testName)
				expectedGqlAppTemplateInputWithProviderAndWebhook.Webhooks = []*graphql.WebhookInput{
					{
						Type:           graphql.WebhookTypeConfigurationChanged,
						Auth:           nil,
						Mode:           webhook.WebhookModePtr(graphql.WebhookModeSync),
						URLTemplate:    &testURL,
						InputTemplate:  str.Ptr("input template"),
						HeaderTemplate: str.Ptr("header template"),
						OutputTemplate: str.Ptr("output template"),
					},
					{
						Type: graphql.WebhookTypeOpenResourceDiscovery,
						URL:  &testURL,
						Auth: nil,
					},
				}
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *expectedGqlAppTemplateInputWithProviderAndWebhook).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplate, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     UnusedWebhookSvc,
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatDoesNotCleanupFunc(labels),
			TenantMappingConfigFn: func() map[string]interface{} {
				tenantMappingJSON := "{\"SYNC\": {\"v1.0\": [{ \"type\": \"CONFIGURATION_CHANGED\",\"urlTemplate\": \"%s\",\"inputTemplate\": \"input template\",\"headerTemplate\": \"header template\",\"outputTemplate\": \"output template\"}]}}"
				return GetTenantMappingConfig(tenantMappingJSON)
			},
			Input:          gqlAppTemplateInputWithProviderAndWebhook,
			ExpectedOutput: gqlAppTemplate,
		},
		{
			Name: "Success with tenant mapping configuration with ASYNC_CALLBACK mode",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false)

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInput, labels).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				expectedGqlAppTemplateInputWithProviderAndWebhook := fixGQLAppTemplateInputWithPlaceholderAndProvider("SAP " + testName)
				expectedGqlAppTemplateInputWithProviderAndWebhook.Webhooks = []*graphql.WebhookInput{
					{
						Type:           graphql.WebhookTypeConfigurationChanged,
						Auth:           nil,
						Mode:           webhook.WebhookModePtr(graphql.WebhookModeAsyncCallback),
						URLTemplate:    &testURL,
						InputTemplate:  str.Ptr("input template"),
						HeaderTemplate: &testURL,
						OutputTemplate: str.Ptr("output template"),
					},
					{
						Type: graphql.WebhookTypeOpenResourceDiscovery,
						URL:  &testURL,
						Auth: nil,
					},
				}
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *expectedGqlAppTemplateInputWithProviderAndWebhook).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplate, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     UnusedWebhookSvc,
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatDoesNotCleanupFunc(labels),
			TenantMappingConfigFn: func() map[string]interface{} {
				tenantMappingJSON := "{\"ASYNC_CALLBACK\": {\"v1.0\": [{ \"type\": \"CONFIGURATION_CHANGED\",\"urlTemplate\": \"%s\",\"inputTemplate\": \"input template\",\"headerTemplate\": \"%s\",\"outputTemplate\": \"output template\"}]}}"
				return GetTenantMappingConfig(tenantMappingJSON)
			},
			Input:          gqlAppTemplateInputWithProviderAndWebhookWithAsyncCallback,
			ExpectedOutput: gqlAppTemplate,
		},
		{
			Name: "Success when self registered app template still does not exists",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false)

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInputWithSelRegLabels, labelsContainingSelfRegistration).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("GetByFilters", txtest.CtxWithDBMatcher(), getAppTemplateFilters).Return(nil, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInputWithSelfRegLabels).Return(*modelAppTemplateInputWithSelRegLabels, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplateWithSelfRegLabels, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn:         UnusedWebhookConv,
			WebhookSvcFn:          UnusedWebhookSvc,
			SelfRegManagerFn:      apptmpltest.SelfRegManagerThatDoesNotCleanupFunc(labelsContainingSelfRegistration),
			TenantMappingConfigFn: EmptyTenantMappingConfig,
			Input:                 gqlAppTemplateInputWithSelfRegLabels,
			ExpectedOutput:        gqlAppTemplateWithSelfRegLabels,
		},
		{
			Name: "Error when self registered app template already exists for the given labels",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "CreateWithLabels")
				appTemplateSvc.AssertNotCalled(t, "Get")
				appTemplateSvc.On("GetByFilters", txtest.CtxWithDBMatcher(), getAppTemplateFilters).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInputWithSelfRegLabels).Return(*modelAppTemplateInputWithSelRegLabels, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:         UnusedWebhookConv,
			WebhookSvcFn:          UnusedWebhookSvc,
			SelfRegManagerFn:      apptmpltest.SelfRegManagerThatDoesCleanup(labelsContainingSelfRegistration),
			TenantMappingConfigFn: EmptyTenantMappingConfig,
			Input:                 gqlAppTemplateInputWithSelfRegLabels,
			ExpectedError:         errors.New("Cannot have more than one application template with labels \"region\": \"region\" and \"test-distinguish-label\": \"selfRegVal\""),
		},
		{
			Name: "Error when missing tenant mapping configuration for mode XXXX",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}

				transact := &persistenceautomock.Transactioner{}

				return persistTx, transact
			},
			AppTemplateSvcFn:      UnusedAppTemplateSvc,
			AppTemplateConvFn:     UnusedAppTemplateConv,
			WebhookConvFn:         UnusedWebhookConv,
			WebhookSvcFn:          UnusedWebhookSvc,
			SelfRegManagerFn:      UnusedSelfRegManager,
			TenantMappingConfigFn: EmptyTenantMappingConfig,
			Input:                 gqlAppTemplateInputWithProviderAndWebhook,
			ExpectedError:         errors.New("missing tenant mapping configuration for mode SYNC"),
		},
		{
			Name: "Error when missing tenant mapping configuration for mode XXX and version XXX",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				transact := &persistenceautomock.Transactioner{}
				return persistTx, transact
			},
			AppTemplateSvcFn:  UnusedAppTemplateSvc,
			AppTemplateConvFn: UnusedAppTemplateConv,
			WebhookConvFn:     UnusedWebhookConv,
			WebhookSvcFn:      UnusedWebhookSvc,
			SelfRegManagerFn:  UnusedSelfRegManager,
			TenantMappingConfigFn: func() map[string]interface{} {
				return GetTenantMappingConfig("{\"SYNC\": {}}")
			},
			Input:         gqlAppTemplateInputWithProviderAndWebhook,
			ExpectedError: errors.New("missing tenant mapping configuration for mode SYNC and version v1.0"),
		},
		{
			Name: "Error when unexpected mode type",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				transact := &persistenceautomock.Transactioner{}
				return persistTx, transact
			},
			AppTemplateSvcFn:  UnusedAppTemplateSvc,
			AppTemplateConvFn: UnusedAppTemplateConv,
			WebhookConvFn:     UnusedWebhookConv,
			WebhookSvcFn:      UnusedWebhookSvc,
			SelfRegManagerFn:  UnusedSelfRegManager,
			TenantMappingConfigFn: func() map[string]interface{} {
				return GetTenantMappingConfig("{\"SYNC\": \"\"}")
			},
			Input:         gqlAppTemplateInputWithProviderAndWebhook,
			ExpectedError: errors.New("unexpected mode type, should be a map, but was string"),
		},
		{
			Name: "Error when checking if self registered app template already exists for the given labels",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "CreateWithLabels")
				appTemplateSvc.AssertNotCalled(t, "Get")
				appTemplateSvc.On("GetByFilters", txtest.CtxWithDBMatcher(), getAppTemplateFilters).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInputWithSelfRegLabels).Return(*modelAppTemplateInputWithSelRegLabels, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:         UnusedWebhookConv,
			WebhookSvcFn:          UnusedWebhookSvc,
			SelfRegManagerFn:      apptmpltest.SelfRegManagerThatDoesCleanup(labelsContainingSelfRegistration),
			TenantMappingConfigFn: EmptyTenantMappingConfig,
			Input:                 gqlAppTemplateInputWithSelfRegLabels,
			ExpectedError:         testError,
		},
		{
			Name: "Returns error when can't convert input from graphql",
			TxFn: txGen.ThatDoesntStartTransaction,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "CreateWithLabels")
				appTemplateSvc.AssertNotCalled(t, "Get")
				appTemplateSvc.AssertNotCalled(t, "GetByFilters")
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(model.ApplicationTemplateInput{}, testError).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:         UnusedWebhookConv,
			WebhookSvcFn:          UnusedWebhookSvc,
			SelfRegManagerFn:      apptmpltest.NoopSelfRegManager,
			TenantMappingConfigFn: EmptyTenantMappingConfig,
			Input:                 gqlAppTemplateInput,
			ExpectedError:         testError,
		},
		{
			Name: "Returns error when creating application template failed",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.AssertNotCalled(t, "Commit")

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInput, labels).Return("", testError).Once()
				appTemplateSvc.AssertNotCalled(t, "Get")
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:         UnusedWebhookConv,
			WebhookSvcFn:          UnusedWebhookSvc,
			SelfRegManagerFn:      apptmpltest.SelfRegManagerThatDoesNotCleanupFunc(labels),
			TenantMappingConfigFn: EmptyTenantMappingConfig,
			Input:                 gqlAppTemplateInput,
			ExpectedError:         testError,
		},
		{
			Name: "Returns error when getting application template failed",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.AssertNotCalled(t, "Commit")

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInput, labels).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:         UnusedWebhookConv,
			WebhookSvcFn:          UnusedWebhookSvc,
			SelfRegManagerFn:      apptmpltest.SelfRegManagerThatDoesNotCleanupFunc(labels),
			TenantMappingConfigFn: EmptyTenantMappingConfig,
			Input:                 gqlAppTemplateInput,
			ExpectedError:         testError,
		},
		{
			Name: "Returns error when beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "CreateWithLabels")
				appTemplateSvc.AssertNotCalled(t, "Get")
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:         UnusedWebhookConv,
			WebhookSvcFn:          UnusedWebhookSvc,
			Input:                 gqlAppTemplateInput,
			SelfRegManagerFn:      apptmpltest.SelfRegManagerThatDoesPrepWithNoErrors(labels),
			TenantMappingConfigFn: EmptyTenantMappingConfig,
			ExpectedError:         testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(testError).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInput, labels).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:         UnusedWebhookConv,
			WebhookSvcFn:          UnusedWebhookSvc,
			SelfRegManagerFn:      apptmpltest.SelfRegManagerThatDoesNotCleanupFunc(labels),
			TenantMappingConfigFn: EmptyTenantMappingConfig,
			Input:                 gqlAppTemplateInput,
			ExpectedError:         testError,
		},
		{
			Name: "Returns error when can't convert application template to graphql",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInput, labels).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(nil, testError).Once()
				return appTemplateConv
			},
			WebhookConvFn:         UnusedWebhookConv,
			WebhookSvcFn:          UnusedWebhookSvc,
			SelfRegManagerFn:      apptmpltest.SelfRegManagerThatDoesNotCleanupFunc(labels),
			TenantMappingConfigFn: EmptyTenantMappingConfig,
			Input:                 gqlAppTemplateInput,
			ExpectedError:         testError,
		},
		{
			Name: "Success when labels are nil after converting gql AppTemplateInput",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				modelAppTemplateInput.Labels = map[string]interface{}{}
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInput, labels).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				modelAppTemplateInput.Labels = nil
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplate, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn:         UnusedWebhookConv,
			WebhookSvcFn:          UnusedWebhookSvc,
			SelfRegManagerFn:      apptmpltest.SelfRegManagerThatDoesNotCleanupFunc(labels),
			TenantMappingConfigFn: EmptyTenantMappingConfig,
			Input:                 gqlAppTemplateInput,
			ExpectedOutput:        gqlAppTemplate,
		},
		{
			Name: "Returns error when app template self registration fails",
			TxFn: txGen.ThatDoesntStartTransaction,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "CreateWithLabels")
				appTemplateSvc.AssertNotCalled(t, "Get")
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:         UnusedWebhookConv,
			WebhookSvcFn:          UnusedWebhookSvc,
			SelfRegManagerFn:      apptmpltest.SelfRegManagerThatReturnsErrorOnPrep,
			TenantMappingConfigFn: EmptyTenantMappingConfig,
			Input:                 gqlAppTemplateInput,
			ExpectedError:         errors.New(apptmpltest.SelfRegErrorMsg),
		},
		{
			Name: "Returns error when self registered app template fails on create",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.AssertNotCalled(t, "Commit")

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				modelAppTemplateInput.Labels = distinguishLabel
				appTemplateSvc.On("GetByFilters", txtest.CtxWithDBMatcher(), getAppTemplateFilters).Return(nil, nil).Once()
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInput, labelsContainingSelfRegistration).Return("", testError).Once()
				appTemplateSvc.AssertNotCalled(t, "Get")
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				gqlAppTemplateInput.Labels = regionLabel
				modelAppTemplateInput.Labels = distinguishLabel
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:         UnusedWebhookConv,
			WebhookSvcFn:          UnusedWebhookSvc,
			SelfRegManagerFn:      apptmpltest.SelfRegManagerThatDoesCleanup(labelsContainingSelfRegistration),
			TenantMappingConfigFn: EmptyTenantMappingConfig,
			Input:                 gqlAppTemplateInput,
			ExpectedError:         testError,
		},
		{
			Name: "Success but couldn't cast region label value to string",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInput, labels).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				gqlAppTemplateInput.Labels = badValueLabel
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplate, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn:         UnusedWebhookConv,
			WebhookSvcFn:          UnusedWebhookSvc,
			SelfRegManagerFn:      apptmpltest.SelfRegManagerThatInitiatesCleanupButNotFinishIt(labels),
			TenantMappingConfigFn: EmptyTenantMappingConfig,
			Input:                 gqlAppTemplateInput,
			ExpectedOutput:        gqlAppTemplate,
		},
		{
			Name:                  "Returns error when validating app template name",
			TxFn:                  txGen.ThatDoesntStartTransaction,
			AppTemplateSvcFn:      UnusedAppTemplateSvc,
			AppTemplateConvFn:     UnusedAppTemplateConv,
			WebhookConvFn:         UnusedWebhookConv,
			WebhookSvcFn:          UnusedWebhookSvc,
			SelfRegManagerFn:      UnusedSelfRegManager,
			TenantMappingConfigFn: EmptyTenantMappingConfig,
			Input:                 fixGQLAppTemplateInputWithPlaceholderAndProvider(testName),
			ExpectedError:         errors.New("application template name \"bar\" does not comply with the following naming convention: \"SAP <product name>\""),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			appTemplateSvc := testCase.AppTemplateSvcFn()
			appTemplateConv := testCase.AppTemplateConvFn()
			webhookSvc := testCase.WebhookSvcFn()
			webhookConverter := testCase.WebhookConvFn()
			selfRegManager := testCase.SelfRegManagerFn()
			tenantMappingConfig := testCase.TenantMappingConfigFn()
			uuidSvc := uidSvcFn()

			resolver := apptemplate.NewResolver(transact, nil, nil, appTemplateSvc, appTemplateConv, webhookSvc, webhookConverter, selfRegManager, uuidSvc, tenantMappingConfig, testURL)

			// WHEN
			result, err := resolver.CreateApplicationTemplate(ctx, *testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			appTemplateSvc.AssertExpectations(t)
			appTemplateConv.AssertExpectations(t)
			selfRegManager.AssertExpectations(t)
		})
	}
	t.Run("Returns error when application template inputs url template has invalid method", func(t *testing.T) {
		gqlAppTemplateInputInvalid := fixGQLAppTemplateInputInvalidAppInputURLTemplateMethod(testName)
		expectedError := errors.New("failed to parse webhook url template")
		_, transact := txGen.ThatSucceeds()

		resolver := apptemplate.NewResolver(transact, nil, nil, nil, nil, nil, nil, nil, nil, nil, "")

		// WHEN
		_, err := resolver.CreateApplicationTemplate(ctx, *gqlAppTemplateInputInvalid)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), expectedError.Error())
	})
}

func TestResolver_Labels(t *testing.T) {
	// GIVEN

	id := "foo"
	tenant := "tenant"
	labelKey1 := "key1"
	labelValue1 := "val1"
	labelKey2 := "key2"
	labelValue2 := "val2"

	gqlAppTemplate := fixGQLAppTemplate(testID, testName, fixGQLApplicationTemplateWebhooks(testWebhookID, testID))

	modelLabels := map[string]*model.Label{
		"abc": {
			ID:         "abc",
			Tenant:     str.Ptr(tenant),
			Key:        labelKey1,
			Value:      labelValue1,
			ObjectID:   id,
			ObjectType: model.AppTemplateLabelableObject,
		},
		"def": {
			ID:         "def",
			Tenant:     str.Ptr(tenant),
			Key:        labelKey2,
			Value:      labelValue2,
			ObjectID:   id,
			ObjectType: model.AppTemplateLabelableObject,
		},
	}

	gqlLabels := graphql.Labels{
		labelKey1: labelValue1,
		labelKey2: labelValue2,
	}

	gqlLabels1 := graphql.Labels{
		labelKey1: labelValue1,
	}

	testErr := errors.New("Test error")

	testCases := []struct {
		Name             string
		PersistenceFn    func() *persistenceautomock.PersistenceTx
		TransactionerFn  func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		AppTemplateSvcFn func() *automock.ApplicationTemplateService
		InputKey         *string
		ExpectedResult   graphql.Labels
		ExpectedErr      error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("ListLabels", txtest.CtxWithDBMatcher(), id).Return(modelLabels, nil).Once()
				return svc
			},
			InputKey:       nil,
			ExpectedResult: gqlLabels,
			ExpectedErr:    nil,
		},
		{
			Name:            "Success when labels are filtered",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("ListLabels", txtest.CtxWithDBMatcher(), id).Return(modelLabels, nil).Once()
				return svc
			},
			InputKey:       &labelKey1,
			ExpectedResult: gqlLabels1,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when label listing failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("ListLabels", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			InputKey:       &labelKey1,
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)

			//persist, transact := testCase.TxFn()
			appTemplateSvc := testCase.AppTemplateSvcFn()

			resolver := apptemplate.NewResolver(transact, nil, nil, appTemplateSvc, nil, nil, nil, nil, nil, nil, "")

			// WHEN
			result, err := resolver.Labels(context.TODO(), gqlAppTemplate, testCase.InputKey)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			appTemplateSvc.AssertExpectations(t)
			transact.AssertExpectations(t)
			persistTx.AssertExpectations(t)
		})
	}
}

func TestResolver_RegisterApplicationFromTemplate(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)
	ctx = consumer.SaveToContext(ctx, consumer.Consumer{ConsumerID: testConsumerID})

	txGen := txtest.NewTransactionContextGenerator(testError)

	globalSubaccountIDLabelKey := "global_subaccount_id"
	filters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(globalSubaccountIDLabelKey, fmt.Sprintf("\"%s\"", "consumer-id")),
	}

	jsonAppCreateInput := fixJSONApplicationCreateInput(testName)
	modelAppCreateInput := fixModelApplicationCreateInput(testName)
	modelAppWithLabelCreateInput := fixModelApplicationWithLabelCreateInput(testName)
	gqlAppCreateInput := fixGQLApplicationCreateInput(testName)

	modelAppTemplate := fixModelAppTemplateWithAppInputJSON(testID, testName, jsonAppCreateInput, fixModelApplicationTemplateWebhooks(testWebhookID, testID))

	modelApplication := fixModelApplication(testID, testName)
	gqlApplication := fixGQLApplication(testID, testName)

	gqlAppFromTemplateInput := fixGQLApplicationFromTemplateInput(testName)
	modelAppFromTemplateInput := fixModelApplicationFromTemplateInput(testName)

	testCases := []struct {
		Name              string
		TxFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AppTemplateSvcFn  func() *automock.ApplicationTemplateService
		AppTemplateConvFn func() *automock.ApplicationTemplateConverter
		WebhookSvcFn      func() *automock.WebhookService
		WebhookConvFn     func() *automock.WebhookConverter
		AppSvcFn          func() *automock.ApplicationService
		AppConvFn         func() *automock.ApplicationConverter
		ExpectedOutput    *graphql.Application
		ExpectedError     error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{}, nil).Once()
				appTemplateSvc.On("ListByName", txtest.CtxWithDBMatcher(), testName).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, globalSubaccountIDLabelKey).Return(nil, apperrors.NewNotFoundError(resource.Label, "id")).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateInput, nil).Once()
				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("CreateFromTemplate", txtest.CtxWithDBMatcher(), modelAppWithLabelCreateInput, str.Ptr(testID)).Return(testID, nil).Once()
				appSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&modelApplication, nil).Once()
				return appSvc
			},
			AppConvFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				appConv.On("CreateInputFromGraphQL", mock.Anything, gqlAppCreateInput).Return(modelAppCreateInput, nil).Once()
				appConv.On("ToGraphQL", &modelApplication).Return(&gqlApplication).Once()
				return appConv
			},
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: &gqlApplication,
			ExpectedError:  nil,
		},
		{
			Name:              "Returns error when transaction begin fails",
			TxFn:              txGen.ThatFailsOnBegin,
			AppTemplateSvcFn:  UnusedAppTemplateSvc,
			AppTemplateConvFn: UnusedAppTemplateConv,
			AppSvcFn:          UnusedAppSvc,
			AppConvFn:         UnusedAppConv,
			WebhookConvFn:     UnusedWebhookConv,
			WebhookSvcFn:      UnusedWebhookSvc,
			ExpectedOutput:    nil,
			ExpectedError:     testError,
		},
		{
			Name: "Returns error when list application templates by filters fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			AppSvcFn:       UnusedAppSvc,
			AppConvFn:      UnusedAppConv,
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
		{
			Name: "Returns error when list application templates by name fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{}, nil).Once()
				appTemplateSvc.On("ListByName", txtest.CtxWithDBMatcher(), testName).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			AppSvcFn:       UnusedAppSvc,
			AppConvFn:      UnusedAppConv,
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
		{
			Name: "Returns error when get application template label fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{}, nil).Once()
				appTemplateSvc.On("ListByName", txtest.CtxWithDBMatcher(), testName).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, globalSubaccountIDLabelKey).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			AppSvcFn:       UnusedAppSvc,
			AppConvFn:      UnusedAppConv,
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
		{
			Name: "Returns error when list application templates by name cannot find application template",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{}, nil).Once()
				appTemplateSvc.On("ListByName", txtest.CtxWithDBMatcher(), testName).Return([]*model.ApplicationTemplate{}, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			AppSvcFn:       UnusedAppSvc,
			AppConvFn:      UnusedAppConv,
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: nil,
			ExpectedError:  errors.New("application template with name \"bar\" and consumer id \"consumer-id\" not found"),
		},
		{
			Name: "Returns error when list application templates by name return more than one application template",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{}, nil).Once()
				appTemplateSvc.On("ListByName", txtest.CtxWithDBMatcher(), testName).Return([]*model.ApplicationTemplate{modelAppTemplate, modelAppTemplate}, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, globalSubaccountIDLabelKey).Return(nil, apperrors.NewNotFoundError(resource.Label, "id")).Twice()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			AppSvcFn:       UnusedAppSvc,
			AppConvFn:      UnusedAppConv,
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: nil,
			ExpectedError:  errors.New("unexpected number of application templates. found 2"),
		},
		{
			Name: "Returns error when preparing ApplicationCreateInputJSON fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return("", testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateInput, nil).Once()
				return appTemplateConv
			},
			AppSvcFn:       UnusedAppSvc,
			AppConvFn:      UnusedAppConv,
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
		{
			Name: "Returns error when CreateInputJSONToGQL fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateInput, nil).Once()
				return appTemplateConv
			},
			AppSvcFn: UnusedAppSvc,
			AppConvFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateInputJSONToGQL", jsonAppCreateInput).Return(graphql.ApplicationRegisterInput{}, testError).Once()
				return appConv
			},
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
		{
			Name: "Returns error when ApplicationCreateInput validation fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateInput, nil).Once()
				return appTemplateConv
			},
			AppSvcFn: UnusedAppSvc,
			AppConvFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateInputJSONToGQL", jsonAppCreateInput).Return(graphql.ApplicationRegisterInput{}, nil).Once()
				return appConv
			},
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: nil,
			ExpectedError:  errors.New("name=cannot be blank"),
		},
		{
			Name: "Returns error when creating Application fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateInput, nil).Once()
				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("CreateFromTemplate", txtest.CtxWithDBMatcher(), modelAppWithLabelCreateInput, str.Ptr(testID)).Return("", testError).Once()
				return appSvc
			},
			AppConvFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateInputFromGraphQL", mock.Anything, gqlAppCreateInput).Return(modelAppCreateInput, nil).Once()
				appConv.On("CreateInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				return appConv
			},
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
		{
			Name: "Returns error when getting Application fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateInput, nil).Once()
				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("CreateFromTemplate", txtest.CtxWithDBMatcher(), modelAppWithLabelCreateInput, str.Ptr(testID)).Return(testID, nil).Once()
				appSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return appSvc
			},
			AppConvFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateInputFromGraphQL", mock.Anything, gqlAppCreateInput).Return(modelAppCreateInput, nil).Once()
				appConv.On("CreateInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				return appConv
			},
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
		{
			Name: "Returns error when committing transaction fails",
			TxFn: txGen.ThatFailsOnCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateInput, nil).Once()
				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("CreateFromTemplate", txtest.CtxWithDBMatcher(), modelAppWithLabelCreateInput, str.Ptr(testID)).Return(testID, nil).Once()
				appSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&modelApplication, nil).Once()
				return appSvc
			},
			AppConvFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateInputFromGraphQL", mock.Anything, gqlAppCreateInput).Return(modelAppCreateInput, nil).Once()
				appConv.On("CreateInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				return appConv
			},
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			appTemplateSvc := testCase.AppTemplateSvcFn()
			appTemplateConv := testCase.AppTemplateConvFn()
			webhookSvc := testCase.WebhookSvcFn()
			webhookConverter := testCase.WebhookConvFn()
			appSvc := testCase.AppSvcFn()
			appConv := testCase.AppConvFn()

			resolver := apptemplate.NewResolver(transact, appSvc, appConv, appTemplateSvc, appTemplateConv, webhookSvc, webhookConverter, nil, nil, nil, "")

			// WHEN
			result, err := resolver.RegisterApplicationFromTemplate(ctx, gqlAppFromTemplateInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			appTemplateSvc.AssertExpectations(t)
			appTemplateConv.AssertExpectations(t)
			appSvc.AssertExpectations(t)
			appConv.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	txGen := txtest.NewTransactionContextGenerator(testError)

	modelAppTemplate := fixModelApplicationTemplate(testID, testName, fixModelApplicationTemplateWebhooks(testWebhookID, testID))
	modelAppTemplateInput := fixModelAppTemplateUpdateInput(testName, appInputJSONString)
	gqlAppTemplate := fixGQLAppTemplate(testID, testName, fixGQLApplicationTemplateWebhooks(testWebhookID, testID))
	gqlAppTemplateUpdateInput := fixGQLAppTemplateUpdateInputWithPlaceholder(testName)
	gqlAppTemplateUpdateInputWithoutNameProperty := fixGQLAppTemplateUpdateInputWithPlaceholder(testName)
	gqlAppTemplateUpdateInputWithoutNameProperty.ApplicationInput.Name = ""
	gqlAppTemplateUpdateInputWithoutDisplayNameLabel := fixGQLAppTemplateUpdateInputWithPlaceholder(testName)
	gqlAppTemplateUpdateInputWithoutDisplayNameLabel.ApplicationInput.Labels = graphql.Labels{}
	gqlAppTemplateUpdateInputWithNonStringDisplayLabel := fixGQLAppTemplateUpdateInputWithPlaceholder(testName)
	gqlAppTemplateUpdateInputWithNonStringDisplayLabel.ApplicationInput.Labels = graphql.Labels{
		"displayName": false,
	}
	gqlAppTemplateUpdateInputWithProvider := fixGQLAppTemplateUpdateInputWithPlaceholderAndProvider(testName)

	labels := map[string]*model.Label{
		"test": {
			Key:   "test key",
			Value: "test value",
		},
	}
	resultLabels := map[string]interface{}{
		"test key": "test value",
	}
	testCases := []struct {
		Name              string
		TxFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AppTemplateSvcFn  func() *automock.ApplicationTemplateService
		AppTemplateConvFn func() *automock.ApplicationTemplateConverter
		SelfRegManagerFn  func() *automock.SelfRegisterManager
		WebhookSvcFn      func() *automock.WebhookService
		WebhookConvFn     func() *automock.WebhookConverter
		Input             *graphql.ApplicationTemplateUpdateInput
		ExpectedOutput    *graphql.ApplicationTemplate
		ExpectedError     error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testID).Return(labels, nil).Once()
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, *modelAppTemplateInput).Return(nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("UpdateInputFromGraphQL", *gqlAppTemplateUpdateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplate, nil).Once()
				return appTemplateConv
			},
			SelfRegManagerFn: func() *automock.SelfRegisterManager {
				srm := &automock.SelfRegisterManager{}
				srm.On("IsSelfRegistrationFlow", txtest.CtxWithDBMatcher(), resultLabels).Return(false, nil).Once()
				return srm
			},
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			Input:          gqlAppTemplateUpdateInput,
			ExpectedOutput: gqlAppTemplate,
		},
		{
			Name: "Success with self reg flow",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				placeholders := []model.ApplicationTemplatePlaceholder{
					{
						Name:        "name",
						Description: &testDescription,
						JSONPath:    &testJSONPath,
					},
					{
						Name:        "display-name",
						Description: &testDescription,
						JSONPath:    &testJSONPath,
					},
				}
				modelAppTemplate := fixModelAppTemplateWithAppInputJSONAndPlaceholders(testID, "SAP app-template", appInputJSONString, fixModelApplicationTemplateWebhooks(testWebhookID, testID), placeholders)
				modelAppTemplateInput := fixModelAppTemplateUpdateInputWithPlaceholders("SAP app-template", appInputJSONString, placeholders)

				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testID).Return(labels, nil).Once()
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, *modelAppTemplateInput).Return(nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				placeholders := []model.ApplicationTemplatePlaceholder{
					{
						Name:        "name",
						Description: &testDescription,
						JSONPath:    &testJSONPath,
					},
					{
						Name:        "display-name",
						Description: &testDescription,
						JSONPath:    &testJSONPath,
					},
				}
				modelAppTemplate := fixModelAppTemplateWithAppInputJSONAndPlaceholders(testID, "SAP app-template", appInputJSONString, fixModelApplicationTemplateWebhooks(testWebhookID, testID), placeholders)
				modelAppTemplateInput := fixModelAppTemplateUpdateInputWithPlaceholders("SAP app-template", appInputJSONString, placeholders)

				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("UpdateInputFromGraphQL", *gqlAppTemplateUpdateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplate, nil).Once()
				return appTemplateConv
			},
			SelfRegManagerFn: func() *automock.SelfRegisterManager {
				srm := &automock.SelfRegisterManager{}
				srm.On("IsSelfRegistrationFlow", txtest.CtxWithDBMatcher(), resultLabels).Return(true, nil).Once()
				return srm
			},
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			Input:          gqlAppTemplateUpdateInput,
			ExpectedOutput: gqlAppTemplate,
		},
		{
			Name:             "Returns error when can't convert input from graphql",
			TxFn:             txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: UnusedAppTemplateSvc,
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("UpdateInputFromGraphQL", *gqlAppTemplateUpdateInput).Return(model.ApplicationTemplateUpdateInput{}, testError).Once()
				return appTemplateConv
			},
			SelfRegManagerFn: UnusedSelfRegManager,
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     UnusedWebhookSvc,
			Input:            gqlAppTemplateUpdateInput,
			ExpectedError:    testError,
		},
		{
			Name: "Returns error when updating application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testID).Return(labels, nil).Once()
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, *modelAppTemplateInput).Return(testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("UpdateInputFromGraphQL", *gqlAppTemplateUpdateInput).Return(*modelAppTemplateInput, nil).Once()
				return appTemplateConv
			},
			SelfRegManagerFn: func() *automock.SelfRegisterManager {
				srm := &automock.SelfRegisterManager{}
				srm.On("IsSelfRegistrationFlow", txtest.CtxWithDBMatcher(), resultLabels).Return(false, nil).Once()
				return srm
			},
			WebhookConvFn: UnusedWebhookConv,
			WebhookSvcFn:  UnusedWebhookSvc,
			Input:         gqlAppTemplateUpdateInput,
			ExpectedError: testError,
		},
		{
			Name: "Returns error when getting application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testID).Return(labels, nil).Once()
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, *modelAppTemplateInput).Return(nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("UpdateInputFromGraphQL", *gqlAppTemplateUpdateInput).Return(*modelAppTemplateInput, nil).Once()
				return appTemplateConv
			},
			SelfRegManagerFn: func() *automock.SelfRegisterManager {
				srm := &automock.SelfRegisterManager{}
				srm.On("IsSelfRegistrationFlow", txtest.CtxWithDBMatcher(), resultLabels).Return(false, nil).Once()
				return srm
			},
			WebhookConvFn: UnusedWebhookConv,
			WebhookSvcFn:  UnusedWebhookSvc,
			Input:         gqlAppTemplateUpdateInput,
			ExpectedError: testError,
		},
		{
			Name: "Returns error when list application template labels failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("UpdateInputFromGraphQL", *gqlAppTemplateUpdateInput).Return(*modelAppTemplateInput, nil).Once()
				return appTemplateConv
			},
			SelfRegManagerFn: UnusedSelfRegManager,
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     UnusedWebhookSvc,
			Input:            gqlAppTemplateUpdateInput,
			ExpectedError:    testError,
		},
		{
			Name: "Returns error when self registration flow check failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testID).Return(labels, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("UpdateInputFromGraphQL", *gqlAppTemplateUpdateInput).Return(*modelAppTemplateInput, nil).Once()
				return appTemplateConv
			},
			SelfRegManagerFn: func() *automock.SelfRegisterManager {
				srm := &automock.SelfRegisterManager{}
				srm.On("IsSelfRegistrationFlow", txtest.CtxWithDBMatcher(), resultLabels).Return(false, testError).Once()
				return srm
			},
			WebhookConvFn: UnusedWebhookConv,
			WebhookSvcFn:  UnusedWebhookSvc,
			Input:         gqlAppTemplateUpdateInput,
			ExpectedError: testError,
		},
		{
			Name: "Returns error when appInputJSON name property is missing",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			SelfRegManagerFn: func() *automock.SelfRegisterManager {
				srm := &automock.SelfRegisterManager{}
				return srm
			},
			WebhookConvFn: UnusedWebhookConv,
			WebhookSvcFn:  UnusedWebhookSvc,
			Input:         gqlAppTemplateUpdateInputWithoutNameProperty,
			ExpectedError: errors.New("appInput: (name: cannot be blank.)"),
		},
		{
			Name: "Returns error when appInputJSON displayName label is missing",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testID).Return(labels, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("UpdateInputFromGraphQL", *gqlAppTemplateUpdateInputWithoutDisplayNameLabel).Return(*fixModelAppTemplateUpdateInput(testName, appInputJSONWithoutDisplayNameLabelString), nil).Once()
				return appTemplateConv
			},
			SelfRegManagerFn: func() *automock.SelfRegisterManager {
				srm := &automock.SelfRegisterManager{}
				srm.On("IsSelfRegistrationFlow", txtest.CtxWithDBMatcher(), resultLabels).Return(true, nil).Once()
				return srm
			},
			WebhookConvFn: UnusedWebhookConv,
			WebhookSvcFn:  UnusedWebhookSvc,
			Input:         gqlAppTemplateUpdateInputWithoutDisplayNameLabel,
			ExpectedError: errors.New("applicationInputJSON name property or applicationInputJSON displayName label is missing. They must be present in order to proceed."),
		},
		{
			Name: "Returns error when appInputJSON displayName label is not string",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testID).Return(labels, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("UpdateInputFromGraphQL", *gqlAppTemplateUpdateInputWithNonStringDisplayLabel).Return(*fixModelAppTemplateUpdateInput(testName, appInputJSONNonStringDisplayNameLabelString), nil).Once()
				return appTemplateConv
			},
			SelfRegManagerFn: func() *automock.SelfRegisterManager {
				srm := &automock.SelfRegisterManager{}
				srm.On("IsSelfRegistrationFlow", txtest.CtxWithDBMatcher(), resultLabels).Return(true, nil).Once()
				return srm
			},
			WebhookConvFn: UnusedWebhookConv,
			WebhookSvcFn:  UnusedWebhookSvc,
			Input:         gqlAppTemplateUpdateInputWithNonStringDisplayLabel,
			ExpectedError: errors.New("\"displayName\" label value must be string"),
		},
		{
			Name:              "Returns error when validating app template name",
			TxFn:              txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn:  UnusedAppTemplateSvc,
			AppTemplateConvFn: UnusedAppTemplateConv,
			SelfRegManagerFn:  UnusedSelfRegManager,
			WebhookConvFn:     UnusedWebhookConv,
			WebhookSvcFn:      UnusedWebhookSvc,
			Input:             gqlAppTemplateUpdateInputWithProvider,
			ExpectedError:     errors.New("application template name \"bar\" does not comply with the following naming convention: \"SAP <product name>\""),
		},
		{
			Name:              "Returns error when beginning transaction",
			TxFn:              txGen.ThatFailsOnBegin,
			AppTemplateSvcFn:  UnusedAppTemplateSvc,
			AppTemplateConvFn: UnusedAppTemplateConv,
			SelfRegManagerFn:  UnusedSelfRegManager,
			WebhookConvFn:     UnusedWebhookConv,
			WebhookSvcFn:      UnusedWebhookSvc,
			Input:             gqlAppTemplateUpdateInput,
			ExpectedError:     testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testID).Return(labels, nil).Once()
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, *modelAppTemplateInput).Return(nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("UpdateInputFromGraphQL", *gqlAppTemplateUpdateInput).Return(*modelAppTemplateInput, nil).Once()
				return appTemplateConv
			},
			SelfRegManagerFn: func() *automock.SelfRegisterManager {
				srm := &automock.SelfRegisterManager{}
				srm.On("IsSelfRegistrationFlow", txtest.CtxWithDBMatcher(), resultLabels).Return(false, nil).Once()
				return srm
			},
			WebhookConvFn: UnusedWebhookConv,
			WebhookSvcFn:  UnusedWebhookSvc,
			Input:         gqlAppTemplateUpdateInput,
			ExpectedError: testError,
		},
		{
			Name: "Returns error when can't convert application template to graphql",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testID).Return(labels, nil).Once()
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, *modelAppTemplateInput).Return(nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("UpdateInputFromGraphQL", *gqlAppTemplateUpdateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(nil, testError).Once()
				return appTemplateConv
			},
			SelfRegManagerFn: func() *automock.SelfRegisterManager {
				srm := &automock.SelfRegisterManager{}
				srm.On("IsSelfRegistrationFlow", txtest.CtxWithDBMatcher(), resultLabels).Return(false, nil).Once()
				return srm
			},
			WebhookConvFn: UnusedWebhookConv,
			WebhookSvcFn:  UnusedWebhookSvc,
			Input:         gqlAppTemplateUpdateInput,
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			appTemplateSvc := testCase.AppTemplateSvcFn()
			appTemplateConv := testCase.AppTemplateConvFn()
			selfRegManager := testCase.SelfRegManagerFn()
			webhookSvc := testCase.WebhookSvcFn()
			webhookConverter := testCase.WebhookConvFn()

			resolver := apptemplate.NewResolver(transact, nil, nil, appTemplateSvc, appTemplateConv, webhookSvc, webhookConverter, selfRegManager, nil, nil, "")

			// WHEN
			result, err := resolver.UpdateApplicationTemplate(ctx, testID, *testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, transact, appTemplateSvc, appTemplateConv, selfRegManager)
		})
	}

	t.Run("Returns error when application template inputs url template has invalid method", func(t *testing.T) {
		gqlAppTemplateUpdateInputInvalid := fixGQLAppTemplateUpdateInputInvalidAppInput(testName)
		expectedError := errors.New("failed to parse webhook url template")
		_, transact := txGen.ThatSucceeds()

		resolver := apptemplate.NewResolver(transact, nil, nil, nil, nil, nil, nil, nil, nil, nil, "")

		// WHEN
		_, err := resolver.UpdateApplicationTemplate(ctx, testID, *gqlAppTemplateUpdateInputInvalid)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), expectedError.Error())
	})
}

func TestResolver_DeleteApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	txGen := txtest.NewTransactionContextGenerator(testError)

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testUUID)
		return uidSvc
	}
	modelAppTemplate := fixModelApplicationTemplate(testID, testName, fixModelApplicationTemplateWebhooks(testWebhookID, testID))
	gqlAppTemplate := fixGQLAppTemplate(testID, testName, fixGQLApplicationTemplateWebhooks(testWebhookID, testID))

	label := &model.Label{Key: RegionKey, Value: "region-0"}
	badValueLabel := &model.Label{Key: RegionKey, Value: 1}

	testCases := []struct {
		Name              string
		TxFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AppTemplateSvcFn  func() *automock.ApplicationTemplateService
		AppTemplateConvFn func() *automock.ApplicationTemplateConverter
		WebhookSvcFn      func() *automock.WebhookService
		WebhookConvFn     func() *automock.WebhookConverter
		SelfRegManagerFn  func() *automock.SelfRegisterManager
		ExpectedOutput    *graphql.ApplicationTemplate
		ExpectedError     error
	}{
		{
			Name: "Success",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(2)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(2)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, apptmpltest.TestDistinguishLabel).Return(label, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, RegionKey).Return(label, nil).Once()
				appTemplateSvc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplate, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     UnusedWebhookSvc,
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatDoesCleanupWithNoErrors,
			ExpectedOutput:   gqlAppTemplate,
		},
		{
			Name: "Returns error when getting application template failed",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.AssertNotCalled(t, "Commit")

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				appTemplateSvc.AssertNotCalled(t, "GetLabel")
				appTemplateSvc.AssertNotCalled(t, "Delete")
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     UnusedWebhookSvc,
			SelfRegManagerFn: apptmpltest.NoopSelfRegManager,
			ExpectedError:    testError,
		},
		{
			Name: "Returns error when deleting application template failed",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(2)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, apptmpltest.TestDistinguishLabel).Return(label, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, RegionKey).Return(label, nil).Once()
				appTemplateSvc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     UnusedWebhookSvc,
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatDoesCleanupWithNoErrors,
			ExpectedError:    testError,
		},
		{
			Name: "Returns error when beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "Get")
				appTemplateSvc.AssertNotCalled(t, "GetLabel")
				appTemplateSvc.AssertNotCalled(t, "Delete")
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     UnusedWebhookSvc,
			SelfRegManagerFn: apptmpltest.NoopSelfRegManager,
			ExpectedError:    testError,
		},
		{
			Name: "Returns error when committing transaction for first time",
			TxFn: txGen.ThatFailsOnCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, apptmpltest.TestDistinguishLabel).Return(label, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, RegionKey).Return(label, nil).Once()
				appTemplateSvc.AssertNotCalled(t, "Delete")
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     UnusedWebhookSvc,
			SelfRegManagerFn: apptmpltest.SelfRegManagerReturnsDistinguishingLabel,
			ExpectedError:    testError,
		},
		{
			Name: "Returns error when committing transaction for second time",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				persistTx.On("Commit").Return(testError).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(2)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, apptmpltest.TestDistinguishLabel).Return(label, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, RegionKey).Return(label, nil).Once()
				appTemplateSvc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     UnusedWebhookSvc,
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatDoesCleanupWithNoErrors,
			ExpectedError:    testError,
		},
		{
			Name: "Returns error when can't convert application template to graphql",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(2)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(2)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, apptmpltest.TestDistinguishLabel).Return(label, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, RegionKey).Return(label, nil).Once()
				appTemplateSvc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(nil, testError).Once()
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     UnusedWebhookSvc,
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatDoesCleanupWithNoErrors,
			ExpectedError:    testError,
		},
		{
			Name: "Returns error when getting label for first time",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, apptmpltest.TestDistinguishLabel).Return(nil, testError).Once()
				appTemplateSvc.AssertNotCalled(t, "GetLabel")
				appTemplateSvc.AssertNotCalled(t, "Delete")
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     UnusedWebhookSvc,
			SelfRegManagerFn: apptmpltest.SelfRegManagerReturnsDistinguishingLabel,
			ExpectedError:    testError,
		},
		{
			Name: "Returns error when getting label for second time",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, apptmpltest.TestDistinguishLabel).Return(label, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, RegionKey).Return(nil, testError).Once()
				appTemplateSvc.AssertNotCalled(t, "Delete")
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     UnusedWebhookSvc,
			SelfRegManagerFn: apptmpltest.SelfRegManagerReturnsDistinguishingLabel,
			ExpectedError:    testError,
		},
		{
			Name: "Success but couldn't cast region label value to string",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				persistTx.AssertNotCalled(t, "Commit")

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.AssertNotCalled(t, "Begin")
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, apptmpltest.TestDistinguishLabel).Return(label, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, RegionKey).Return(badValueLabel, nil).Once()
				appTemplateSvc.AssertNotCalled(t, "Delete")
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     UnusedWebhookSvc,
			SelfRegManagerFn: apptmpltest.SelfRegManagerReturnsDistinguishingLabel,
			ExpectedOutput:   nil,
		},
		{
			Name: "Returns error when CleanUpSelfRegistration fails",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				persistTx.AssertNotCalled(t, "Commit")

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.AssertNotCalled(t, "Begin")
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, apptmpltest.TestDistinguishLabel).Return(label, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, RegionKey).Return(label, nil).Once()
				appTemplateSvc.AssertNotCalled(t, "Delete")
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     UnusedWebhookSvc,
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatReturnsErrorOnCleanup,
			ExpectedError:    errors.New(apptmpltest.SelfRegErrorMsg),
		},
		{
			Name: "Returns error when beginning transaction for second time",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				persistTx.AssertNotCalled(t, "Commit")

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("Begin").Return(persistTx, testError).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, apptmpltest.TestDistinguishLabel).Return(label, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, RegionKey).Return(label, nil).Once()
				appTemplateSvc.AssertNotCalled(t, "Delete")
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     UnusedWebhookSvc,
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatDoesCleanupWithNoErrors,
			ExpectedError:    testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			appTemplateSvc := testCase.AppTemplateSvcFn()
			appTemplateConv := testCase.AppTemplateConvFn()
			webhookSvc := testCase.WebhookSvcFn()
			webhookConverter := testCase.WebhookConvFn()
			selfRegManager := testCase.SelfRegManagerFn()
			uuidSvc := uidSvcFn()

			resolver := apptemplate.NewResolver(transact, nil, nil, appTemplateSvc, appTemplateConv, webhookSvc, webhookConverter, selfRegManager, uuidSvc, nil, "")

			// WHEN
			result, err := resolver.DeleteApplicationTemplate(ctx, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			appTemplateSvc.AssertExpectations(t)
			appTemplateConv.AssertExpectations(t)
			selfRegManager.AssertExpectations(t)
		})
	}
}

func UnusedAppTemplateSvc() *automock.ApplicationTemplateService {
	return &automock.ApplicationTemplateService{}
}

func UnusedAppTemplateConv() *automock.ApplicationTemplateConverter {
	return &automock.ApplicationTemplateConverter{}
}

func UnusedAppSvc() *automock.ApplicationService {
	return &automock.ApplicationService{}
}

func UnusedAppConv() *automock.ApplicationConverter {
	return &automock.ApplicationConverter{}
}

func UnusedSelfRegManager() *automock.SelfRegisterManager {
	return &automock.SelfRegisterManager{}
}

func UnusedWebhookConv() *automock.WebhookConverter {
	return &automock.WebhookConverter{}
}

func UnusedWebhookSvc() *automock.WebhookService {
	return &automock.WebhookService{}
}

func EmptyTenantMappingConfig() map[string]interface{} {
	return nil
}

func GetTenantMappingConfig(config string) map[string]interface{} {
	var tenantMappingConfig map[string]interface{}
	json.Unmarshal([]byte(config), &tenantMappingConfig)
	return tenantMappingConfig
}
