package apptemplate_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/apiclient"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"

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
	RegionKey               = "region"
	AppTemplateProductLabel = "systemRole"
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
			Name: "Returns NotFoundError when application template not found",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, apperrors.NewNotFoundError(resource.ApplicationTemplate, "")).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: UnusedAppTemplateConv,
			WebhookConvFn:     UnusedWebhookConv,
			WebhookSvcFn:      UnusedWebhookSvc,
			ExpectedError:     apperrors.NewNotFoundError(resource.ApplicationTemplate, ""),
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

			resolver := apptemplate.NewResolver(transact, nil, nil, appTemplateSvc, appTemplateConv, webhookSvc, webhookConverter, nil, nil, nil, nil, nil, "", apiclient.OrdAggregatorClientConfig{})

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

			resolver := apptemplate.NewResolver(transact, nil, nil, appTemplateSvc, appTemplateConv, webhookSvc, webhookConverter, nil, nil, nil, nil, nil, "", apiclient.OrdAggregatorClientConfig{})

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

			resolver := apptemplate.NewResolver(mockTransactioner, nil, nil, nil, nil, webhookSvc, converter, nil, nil, nil, nil, nil, "", apiclient.OrdAggregatorClientConfig{})

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
	tokenConsumer := consumer.Consumer{
		ConsumerID: testTenant,
		Flow:       oathkeeper.OAuth2Flow,
		Region:     "region",
	}
	certConsumer := consumer.Consumer{
		ConsumerID: testTenant,
		Flow:       oathkeeper.CertificateFlow,
		Region:     "region",
	}

	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)
	ctxWithCertConsumer := consumer.SaveToContext(ctx, certConsumer)
	ctxWithTokenConsumer := consumer.SaveToContext(ctx, tokenConsumer)

	txGen := txtest.NewTransactionContextGenerator(testError)

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testUUID)
		return uidSvc
	}

	modelAppTemplate := fixModelApplicationTemplate(testID, testName, fixModelApplicationWebhooks(testWebhookID, testID))
	modelAppTemplateWithRegionPlaceholder := fixModelApplicationTemplateWithRegionPlaceholders(testID, testName, fixModelApplicationWebhooks(testWebhookID, testID))

	modelAppTemplateWithOrdWebhookInput := fixModelAppTemplateInputWithOrdWebhook(testName, appInputJSONString)
	modelAppTemplateWithOrdWebhookInput.ID = &testUUID
	modelAppTemplateInput := fixModelAppTemplateInput(testName, appInputJSONString)
	modelAppTemplateInput.ID = &testUUID

	gqlAppTemplate := fixGQLAppTemplate(testID, testName, fixGQLApplicationTemplateWebhooks(testWebhookID, testID))
	gqlAppTemplateInput := fixGQLAppTemplateInputWithPlaceholder(testName)
	gqlAppTemplateInputWithProvider := fixGQLAppTemplateInputWithPlaceholderAndProvider("SAP " + testName)

	modelAppTemplateInputWithSelRegLabels := fixModelAppTemplateInput(testName, appInputJSONString)
	modelAppTemplateInputWithSelRegLabels.ID = &testUUID
	modelAppTemplateInputWithSelRegLabels.Labels = graphql.Labels{
		apptmpltest.TestDistinguishLabel: "selfRegVal",
		selfregmanager.RegionLabel:       region,
	}

	modelGlobalAppTemplateInputWithProductLabels := fixGlobalModelAppTemplateInputWithProductLabel(testName, appInputJSONWithRegionString)
	modelGlobalAppTemplateInputWithProductLabels.ID = &testUUID

	modelRegionalAppTemplateInputWithProductLabelsWithoutRegionPlaceholder := fixModelAppTemplateInput(testName, appInputJSONWithRegionString)
	modelRegionalAppTemplateInputWithProductLabelsWithoutRegionPlaceholder.ID = &testUUID
	modelRegionalAppTemplateInputWithProductLabelsWithoutRegionPlaceholder.Labels = graphql.Labels{
		AppTemplateProductLabel:    []interface{}{"role"},
		selfregmanager.RegionLabel: region,
	}

	modelRegionalAppTemplateInputWithProductLabels := fixRegionalModelAppTemplateInputWithProductLabel(testName, appInputJSONWithRegionString, region)
	modelRegionalAppTemplateInputWithProductLabels.ID = &testUUID

	modelRegionalAppTemplateInputWithProductLabelsAndDifferentPlaceholders := fixModelAppTemplateInputWithRegionLabelAndDifferentPlaceholders(testName, appInputJSONWithRegionString, region)
	modelRegionalAppTemplateInputWithProductLabelsAndDifferentPlaceholders.ID = &testUUID

	modelRegionalAppTemplateInputWithProductLabelsAndNoPlaceholders := fixRegionalModelAppTemplateInputWithProductLabel(testName, appInputJSONWithRegionString, region)
	modelRegionalAppTemplateInputWithProductLabelsAndNoPlaceholders.ID = &testUUID
	modelRegionalAppTemplateInputWithProductLabelsAndNoPlaceholders.Placeholders = []model.ApplicationTemplatePlaceholder{}

	modelRegionalAppTemplateInputWithProductLabelsNoAppInputJSONRegion := fixRegionalModelAppTemplateInputWithProductLabel(testName, appInputJSONStringNoRegionString, region)
	modelRegionalAppTemplateInputWithProductLabelsNoAppInputJSONRegion.ID = &testUUID

	modelAppTemplateInputWithProductAndSelfRegLabels := fixModelAppTemplateInput(testName, appInputJSONString)
	modelAppTemplateInputWithProductAndSelfRegLabels.ID = &testUUID
	modelAppTemplateInputWithProductAndSelfRegLabels.Labels = graphql.Labels{
		AppTemplateProductLabel:          []interface{}{"role"},
		selfregmanager.RegionLabel:       region,
		apptmpltest.TestDistinguishLabel: "selfRegVal",
	}

	gqlAppTemplateWithSelfRegLabels := fixGQLAppTemplate(testID, testName, fixGQLApplicationTemplateWebhooks(testWebhookID, testID))
	gqlAppTemplateWithSelfRegLabels.Labels = graphql.Labels{
		apptmpltest.TestDistinguishLabel: "selfRegVal",
		selfregmanager.RegionLabel:       region,
	}
	gqlAppTemplateInputWithSelfRegLabels := fixGQLAppTemplateInputWithPlaceholder(testName)
	gqlAppTemplateInputWithSelfRegLabels.Labels = graphql.Labels{
		apptmpltest.TestDistinguishLabel: "selfRegVal",
		selfregmanager.RegionLabel:       region,
	}
	gqlGlobalAppTemplateInputWithProductLabels := fixGlobalGQLAppTemplateInputWithProductLabel(testName)

	gqlRegionalAppTemplateInputWithProductLabelsWithoutRegionPlaceholder := fixRegionalGQLAppTemplateInputWithProductLabel(testName, region)
	gqlRegionalAppTemplateInputWithProductLabelsWithoutRegionPlaceholder.Placeholders = fixGQLPlaceholderDefinitionInput()

	gqlRegionalAppTemplateInputWithProductLabels := fixRegionalGQLAppTemplateInputWithProductLabel(testName, region)

	gqlAppTemplateInputWithDifferentPlaceholdersProductLabels := fixRegionalGQLAppTemplateInputWithDifferentRegionPlaceholder(testName)

	gqlRegionalAppTemplateInputWithProductLabelsNoPlaceholders := fixRegionalGQLAppTemplateInputWithProductLabel(testName, region)
	gqlRegionalAppTemplateInputWithProductLabelsNoPlaceholders.Placeholders = []*graphql.PlaceholderDefinitionInput{}

	gqlRegionalAppTemplateInputWithProductLabelsNoAppInputJSONRegion := fixRegionalGQLAppTemplateInputWithProductLabel(testName, region)
	gqlRegionalAppTemplateInputWithProductLabelsNoAppInputJSONRegion.ApplicationInput = &graphql.ApplicationJSONInput{
		Name:   "foo",
		Labels: map[string]interface{}{"otherKey": "{{region}}", "another": "{{test}}"},
	}

	syncMode := graphql.WebhookModeSync
	asyncCallbackMode := graphql.WebhookModeAsyncCallback

	gqlAppTemplateInputWithProviderAndWebhook := fixGQLAppTemplateInputWithPlaceholderAndProvider("SAP " + testName)
	gqlAppTemplateInputWithProviderAndWebhook.Webhooks = []*graphql.WebhookInput{
		{
			Type:    graphql.WebhookTypeConfigurationChanged,
			URL:     &testURL,
			Auth:    nil,
			Mode:    &syncMode,
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
			Mode:    &asyncCallbackMode,
			Version: str.Ptr("v1.0"),
		},
		{
			Type: graphql.WebhookTypeOpenResourceDiscovery,
			URL:  &testURL,
			Auth: nil,
		},
	}

	labelsContainingSelfRegistrationAndInvaidRegion := map[string]interface{}{apptmpltest.TestDistinguishLabel: "selfRegVal", RegionKey: 1}
	labelsContainingSelfRegistration := map[string]interface{}{apptmpltest.TestDistinguishLabel: "selfRegVal", RegionKey: region}
	labelsContainingSelfRegAndSubaccount := map[string]interface{}{
		apptmpltest.TestDistinguishLabel:   "selfRegVal",
		RegionKey:                          region,
		scenarioassignment.SubaccountIDKey: testTenant,
	}
	distinguishLabel := map[string]interface{}{apptmpltest.TestDistinguishLabel: "selfRegVal"}
	regionLabel := map[string]interface{}{RegionKey: region}

	getAppTemplateFiltersForSelfReg := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(apptmpltest.TestDistinguishLabel, fmt.Sprintf("\"%s\"", "selfRegVal")),
		labelfilter.NewForKeyWithQuery(selfregmanager.RegionLabel, fmt.Sprintf("\"%s\"", region)),
	}

	getAppTemplateFiltersForProduct := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(AppTemplateProductLabel, fmt.Sprintf(`$[*] ? (@ == "%s")`, "role")),
	}

	sameRegionLabelModel := &model.Label{
		Key:   RegionKey,
		Value: region,
	}
	differentRegionLabelModel := &model.Label{
		Key:   RegionKey,
		Value: "totally-different",
	}

	testCases := []struct {
		Name              string
		TxFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AppTemplateSvcFn  func() *automock.ApplicationTemplateService
		AppTemplateConvFn func() *automock.ApplicationTemplateConverter
		WebhookSvcFn      func() *automock.WebhookService
		WebhookConvFn     func() *automock.WebhookConverter
		SelfRegManagerFn  func() *automock.SelfRegisterManager
		LabelSvcFn        func() *automock.LabelService
		Ctx               context.Context
		Input             *graphql.ApplicationTemplateInput
		ExpectedOutput    *graphql.ApplicationTemplate
		ExpectedError     error
	}{
		{
			Name: "Success - no self reg flow",
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
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInput, modelAppTemplateInput.Labels).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInputWithProvider).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplate, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInputWithProvider.Webhooks, gqlAppTemplateInputWithProvider.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerOnlyGetDistinguishedLabelKeyTwice(),
			LabelSvcFn:       UnusedLabelService,
			Ctx:              ctxWithTokenConsumer,
			Input:            gqlAppTemplateInputWithProvider,
			ExpectedOutput:   gqlAppTemplate,
		},
		{
			Name: "Success - no self reg flow and app template has ord webhook",
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
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateWithOrdWebhookInput, modelAppTemplateWithOrdWebhookInput.Labels).Return(*modelAppTemplateWithOrdWebhookInput.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testUUID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInputWithProvider).Return(*modelAppTemplateWithOrdWebhookInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplate, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInputWithProvider.Webhooks, gqlAppTemplateInputWithProvider.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerOnlyGetDistinguishedLabelKeyTwice(),
			LabelSvcFn:       UnusedLabelService,
			Ctx:              ctxWithTokenConsumer,
			Input:            gqlAppTemplateInputWithProvider,
			ExpectedOutput:   gqlAppTemplate,
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
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInput, modelAppTemplateInput.Labels).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				expectedGqlAppTemplateInputWithProviderAndWebhook := fixGQLAppTemplateInputWithPlaceholderAndProvider("SAP " + testName)
				expectedGqlAppTemplateInputWithProviderAndWebhook.Webhooks = fixEnrichedTenantMappedWebhooks()
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *expectedGqlAppTemplateInputWithProviderAndWebhook).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplate, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInputWithProviderAndWebhook.Webhooks, fixEnrichedTenantMappedWebhooks()),
			SelfRegManagerFn: apptmpltest.SelfRegManagerOnlyGetDistinguishedLabelKeyTwice(),
			LabelSvcFn:       UnusedLabelService,
			Ctx:              ctxWithTokenConsumer,
			Input:            gqlAppTemplateInputWithProviderAndWebhook,
			ExpectedOutput:   gqlAppTemplate,
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
				appTemplateSvc.On("GetByFilters", txtest.CtxWithDBMatcher(), getAppTemplateFiltersForSelfReg).Return(nil, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInputWithSelfRegLabels).Return(*modelAppTemplateInputWithSelRegLabels, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplateWithSelfRegLabels, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInputWithSelfRegLabels.Webhooks, gqlAppTemplateInputWithSelfRegLabels.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatDoesNotCleanupFunc(labelsContainingSelfRegistration),
			LabelSvcFn:       UnusedLabelService,
			Ctx:              ctxWithCertConsumer,
			Input:            gqlAppTemplateInputWithSelfRegLabels,
			ExpectedOutput:   gqlAppTemplateWithSelfRegLabels,
		},
		{
			Name: "Error when self registered app template already exists for the given self reg labels",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "CreateWithLabels")
				appTemplateSvc.AssertNotCalled(t, "Get")
				appTemplateSvc.On("GetByFilters", txtest.CtxWithDBMatcher(), getAppTemplateFiltersForSelfReg).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInputWithSelfRegLabels).Return(*modelAppTemplateInputWithSelRegLabels, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInputWithSelfRegLabels.Webhooks, gqlAppTemplateInputWithSelfRegLabels.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatDoesCleanup(labelsContainingSelfRegistration),
			LabelSvcFn:       UnusedLabelService,
			Input:            gqlAppTemplateInputWithSelfRegLabels,
			Ctx:              ctxWithCertConsumer,
			ExpectedError:    errors.New("Cannot have more than one application template with labels \"test-distinguish-label\": \"selfRegVal\" and \"region\": \"region-1\""),
		},
		{
			Name: "Error when app template is regional and the region already exists for the given product labels",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "CreateWithLabels")
				appTemplateSvc.AssertNotCalled(t, "Get")
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), getAppTemplateFiltersForProduct).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlRegionalAppTemplateInputWithProductLabels).Return(*modelRegionalAppTemplateInputWithProductLabels, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlRegionalAppTemplateInputWithProductLabels.Webhooks, gqlRegionalAppTemplateInputWithProductLabels.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatInitiatesCleanupButNotFinishIt(),
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", txtest.CtxWithDBMatcher(), "", model.AppTemplateLabelableObject, testID, RegionKey).Return(sameRegionLabelModel, nil).Once()
				return svc
			},
			Ctx:           ctxWithCertConsumer,
			Input:         gqlRegionalAppTemplateInputWithProductLabels,
			ExpectedError: errors.New(`Regional Application Template with "systemRole" label and "region": "region-1" already exists`),
		},
		{
			Name: "Error when app template is regional and the listing Application Templates fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "CreateWithLabels")
				appTemplateSvc.AssertNotCalled(t, "Get")
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), getAppTemplateFiltersForProduct).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlRegionalAppTemplateInputWithProductLabels).Return(*modelRegionalAppTemplateInputWithProductLabels, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlRegionalAppTemplateInputWithProductLabels.Webhooks, gqlRegionalAppTemplateInputWithProductLabels.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatInitiatesCleanupButNotFinishIt(),
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", txtest.CtxWithDBMatcher(), "", model.AppTemplateLabelableObject, testID, RegionKey).Return(sameRegionLabelModel, nil).Once()
				return svc
			},
			Ctx:           ctxWithCertConsumer,
			Input:         gqlRegionalAppTemplateInputWithProductLabels,
			ExpectedError: errors.New(`while getting Application Template for labels "systemRole": ["role"]`),
		},
		{
			Name: "Error when app template is regional and getting labels fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "CreateWithLabels")
				appTemplateSvc.AssertNotCalled(t, "Get")
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), getAppTemplateFiltersForProduct).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlRegionalAppTemplateInputWithProductLabels).Return(*modelRegionalAppTemplateInputWithProductLabels, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlRegionalAppTemplateInputWithProductLabels.Webhooks, gqlRegionalAppTemplateInputWithProductLabels.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatInitiatesCleanupButNotFinishIt(),
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", txtest.CtxWithDBMatcher(), "", model.AppTemplateLabelableObject, testID, RegionKey).Return(nil, testError).Once()
				return svc
			},
			Ctx:           ctxWithCertConsumer,
			Input:         gqlRegionalAppTemplateInputWithProductLabels,
			ExpectedError: errors.New(`while getting "region" label for Application Template`),
		},
		{
			Name: "Error when app template is regional and the region placeholder is different from the existing regional app templates",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "CreateWithLabels")
				appTemplateSvc.AssertNotCalled(t, "Get")
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), getAppTemplateFiltersForProduct).Return([]*model.ApplicationTemplate{modelAppTemplateWithRegionPlaceholder}, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInputWithDifferentPlaceholdersProductLabels).Return(*modelRegionalAppTemplateInputWithProductLabelsAndDifferentPlaceholders, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInputWithDifferentPlaceholdersProductLabels.Webhooks, gqlAppTemplateInputWithDifferentPlaceholdersProductLabels.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatInitiatesCleanupButNotFinishIt(),
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", txtest.CtxWithDBMatcher(), "", model.AppTemplateLabelableObject, testID, RegionKey).Return(differentRegionLabelModel, nil).Once()
				return svc
			},
			Ctx:           ctxWithCertConsumer,
			Input:         gqlAppTemplateInputWithDifferentPlaceholdersProductLabels,
			ExpectedError: errors.New(`Regional Application Template input with "systemRole" label has a different "region" placeholder from the other Application Templates with the same label`),
		},
		{
			Name: "Error when app template is regional but it does not have a region placeholder",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "CreateWithLabels")
				appTemplateSvc.AssertNotCalled(t, "Get")
				appTemplateSvc.AssertNotCalled(t, "ListByFilters")
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlRegionalAppTemplateInputWithProductLabelsNoPlaceholders).Return(*modelRegionalAppTemplateInputWithProductLabelsAndNoPlaceholders, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlRegionalAppTemplateInputWithProductLabelsNoPlaceholders.Webhooks, gqlRegionalAppTemplateInputWithProductLabelsNoPlaceholders.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatInitiatesCleanupButNotFinishIt(),
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", txtest.CtxWithDBMatcher(), "", model.AppTemplateLabelableObject, testID, RegionKey).Return(differentRegionLabelModel, nil).Once()
				return svc
			},
			Ctx:           ctxWithCertConsumer,
			Input:         gqlRegionalAppTemplateInputWithProductLabelsNoPlaceholders,
			ExpectedError: errors.New(`"region" placeholder should be present for regional Application Templates`),
		},
		{
			Name: "Error when app template is regional and the existing application template does not have a region placeholder",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "CreateWithLabels")
				appTemplateSvc.AssertNotCalled(t, "Get")
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), getAppTemplateFiltersForProduct).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlRegionalAppTemplateInputWithProductLabels).Return(*modelRegionalAppTemplateInputWithProductLabels, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlRegionalAppTemplateInputWithProductLabels.Webhooks, gqlRegionalAppTemplateInputWithProductLabels.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatInitiatesCleanupButNotFinishIt(),
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", txtest.CtxWithDBMatcher(), "", model.AppTemplateLabelableObject, testID, RegionKey).Return(differentRegionLabelModel, nil).Once()
				return svc
			},
			Ctx:           ctxWithCertConsumer,
			Input:         gqlRegionalAppTemplateInputWithProductLabels,
			ExpectedError: errors.New(`"region" placeholder should be present for regional Application Templates`),
		},
		{
			Name: "Error when app template is regional but does not have a region label in the Application Input JSON",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "CreateWithLabels")
				appTemplateSvc.AssertNotCalled(t, "Get")
				appTemplateSvc.AssertNotCalled(t, "ListByFilters")
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlRegionalAppTemplateInputWithProductLabelsNoAppInputJSONRegion).Return(*modelRegionalAppTemplateInputWithProductLabelsNoAppInputJSONRegion, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlRegionalAppTemplateInputWithProductLabelsNoAppInputJSONRegion.Webhooks, gqlRegionalAppTemplateInputWithProductLabelsNoAppInputJSONRegion.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatInitiatesCleanupButNotFinishIt(),
			LabelSvcFn:       UnusedLabelService,
			Ctx:              ctxWithCertConsumer,
			Input:            gqlRegionalAppTemplateInputWithProductLabelsNoAppInputJSONRegion,
			ExpectedError:    errors.New(`App Template with "region" label has a missing "region" label in the applicationInput`),
		},
		{
			Name: "Error when a global Application Template already exists",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "CreateWithLabels")
				appTemplateSvc.AssertNotCalled(t, "Get")
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), getAppTemplateFiltersForProduct).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlGlobalAppTemplateInputWithProductLabels).Return(*modelRegionalAppTemplateInputWithProductLabels, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlGlobalAppTemplateInputWithProductLabels.Webhooks, gqlGlobalAppTemplateInputWithProductLabels.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatInitiatesCleanupButNotFinishIt(),
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", txtest.CtxWithDBMatcher(), "", model.AppTemplateLabelableObject, testID, RegionKey).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				return svc
			},
			Ctx:           ctxWithCertConsumer,
			Input:         gqlGlobalAppTemplateInputWithProductLabels,
			ExpectedError: errors.New(`Application Template with "systemRole" label is global and already exists`),
		},
		{
			Name: "Error when there are regional Application Templates but the new Application Template is global",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "CreateWithLabels")
				appTemplateSvc.AssertNotCalled(t, "Get")
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), getAppTemplateFiltersForProduct).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlGlobalAppTemplateInputWithProductLabels).Return(*modelGlobalAppTemplateInputWithProductLabels, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlGlobalAppTemplateInputWithProductLabels.Webhooks, gqlGlobalAppTemplateInputWithProductLabels.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatInitiatesCleanupButNotFinishIt(),
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", txtest.CtxWithDBMatcher(), "", model.AppTemplateLabelableObject, testID, RegionKey).Return(differentRegionLabelModel, nil).Once()
				return svc
			},
			Ctx:           ctxWithCertConsumer,
			Input:         gqlGlobalAppTemplateInputWithProductLabels,
			ExpectedError: errors.New(`Existing application template with "systemRole" label is regional. The input application template should contain a "region" label`),
		},
		{
			Name: "Error when the new Application Template is regional but does not have a region placeholder",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "CreateWithLabels")
				appTemplateSvc.AssertNotCalled(t, "Get")
				appTemplateSvc.AssertNotCalled(t, "ListByFilters")
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlRegionalAppTemplateInputWithProductLabelsWithoutRegionPlaceholder).Return(*modelRegionalAppTemplateInputWithProductLabelsWithoutRegionPlaceholder, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlRegionalAppTemplateInputWithProductLabelsWithoutRegionPlaceholder.Webhooks, gqlRegionalAppTemplateInputWithProductLabelsWithoutRegionPlaceholder.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatInitiatesCleanupButNotFinishIt(),
			LabelSvcFn:       UnusedLabelService,
			Ctx:              ctxWithCertConsumer,
			Input:            gqlRegionalAppTemplateInputWithProductLabelsWithoutRegionPlaceholder,
			ExpectedError:    errors.New(`"region" placeholder should be present for regional Application Templates`),
		},
		{
			Name: "Error when checking if self registered app template already exists for the given labels",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "CreateWithLabels")
				appTemplateSvc.AssertNotCalled(t, "Get")
				appTemplateSvc.On("GetByFilters", txtest.CtxWithDBMatcher(), getAppTemplateFiltersForSelfReg).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInputWithSelfRegLabels).Return(*modelAppTemplateInputWithSelRegLabels, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInputWithSelfRegLabels.Webhooks, gqlAppTemplateInputWithSelfRegLabels.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatDoesCleanup(labelsContainingSelfRegistration),
			LabelSvcFn:       UnusedLabelService,
			Ctx:              ctxWithCertConsumer,
			Input:            gqlAppTemplateInputWithSelfRegLabels,
			ExpectedError:    testError,
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
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInput.Webhooks, gqlAppTemplateInput.Webhooks),
			SelfRegManagerFn: apptmpltest.NoopSelfRegManager,
			LabelSvcFn:       UnusedLabelService,
			Ctx:              ctxWithTokenConsumer,
			Input:            gqlAppTemplateInput,
			ExpectedError:    testError,
		},
		{
			Name: "Returns error when loading consumer info",
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
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInput.Webhooks, gqlAppTemplateInput.Webhooks),
			SelfRegManagerFn: apptmpltest.NoopSelfRegManager,
			LabelSvcFn:       UnusedLabelService,
			Ctx:              context.Background(),
			Input:            gqlAppTemplateInput,
			ExpectedError:    errors.New("while loading consumer"),
		},
		{
			Name: "Returns error when flow is cert and self reg label and product label are present",
			TxFn: txGen.ThatDoesntStartTransaction,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInputWithProductAndSelfRegLabels, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInput.Webhooks, gqlAppTemplateInput.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerOnlyGetDistinguishedLabelKeyOnce(),
			LabelSvcFn:       UnusedLabelService,
			Ctx:              ctxWithCertConsumer,
			Input:            gqlAppTemplateInput,
			ExpectedError:    fmt.Errorf("should provide either %q or %q label", apptmpltest.TestDistinguishLabel, AppTemplateProductLabel),
		},
		{
			Name: "Returns error when flow is cert and self reg label or product label is not present",
			TxFn: txGen.ThatDoesntStartTransaction,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInput.Webhooks, gqlAppTemplateInput.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerOnlyGetDistinguishedLabelKeyOnce(),
			LabelSvcFn:       UnusedLabelService,
			Ctx:              ctxWithCertConsumer,
			Input:            gqlAppTemplateInput,
			ExpectedError:    fmt.Errorf("missing %q or %q label", apptmpltest.TestDistinguishLabel, AppTemplateProductLabel),
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
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInput, modelAppTemplateInput.Labels).Return("", testError).Once()
				appTemplateSvc.AssertNotCalled(t, "Get")
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInput.Webhooks, gqlAppTemplateInput.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerOnlyGetDistinguishedLabelKeyTwice(),
			LabelSvcFn:       UnusedLabelService,
			Ctx:              ctxWithTokenConsumer,
			Input:            gqlAppTemplateInput,
			ExpectedError:    testError,
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
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInput, modelAppTemplateInput.Labels).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInput.Webhooks, gqlAppTemplateInput.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerOnlyGetDistinguishedLabelKeyTwice(),
			LabelSvcFn:       UnusedLabelService,
			Input:            gqlAppTemplateInput,
			Ctx:              ctxWithTokenConsumer,
			ExpectedError:    testError,
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
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInput.Webhooks, gqlAppTemplateInput.Webhooks),
			Input:            gqlAppTemplateInput,
			SelfRegManagerFn: apptmpltest.SelfRegManagerOnlyGetDistinguishedLabelKeyOnce(),
			LabelSvcFn:       UnusedLabelService,
			Ctx:              ctxWithTokenConsumer,
			ExpectedError:    testError,
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
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInput, modelAppTemplateInput.Labels).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInput.Webhooks, gqlAppTemplateInput.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerOnlyGetDistinguishedLabelKeyTwice(),
			LabelSvcFn:       UnusedLabelService,
			Ctx:              ctxWithTokenConsumer,
			Input:            gqlAppTemplateInput,
			ExpectedError:    testError,
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
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInput, modelAppTemplateInput.Labels).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(nil, testError).Once()
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInput.Webhooks, gqlAppTemplateInput.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerOnlyGetDistinguishedLabelKeyTwice(),
			LabelSvcFn:       UnusedLabelService,
			Input:            gqlAppTemplateInput,
			Ctx:              ctxWithTokenConsumer,
			ExpectedError:    testError,
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
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInput, modelAppTemplateInput.Labels).Return(modelAppTemplate.ID, nil).Once()
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
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInput.Webhooks, gqlAppTemplateInput.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerOnlyGetDistinguishedLabelKeyTwice(),
			LabelSvcFn:       UnusedLabelService,
			Ctx:              ctxWithTokenConsumer,
			Input:            gqlAppTemplateInput,
			ExpectedOutput:   gqlAppTemplate,
		},
		{
			Name: "Returns error when flow is not cert but self register label has been provided",
			TxFn: txGen.ThatDoesntStartTransaction,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "CreateWithLabels")
				appTemplateSvc.AssertNotCalled(t, "Get")
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInputWithSelfRegLabels).Return(*modelAppTemplateInputWithSelRegLabels, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInputWithSelfRegLabels.Webhooks, gqlAppTemplateInputWithSelfRegLabels.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerOnlyGetDistinguishedLabelKeyOnce(),
			LabelSvcFn:       UnusedLabelService,
			Input:            gqlAppTemplateInputWithSelfRegLabels,
			Ctx:              ctxWithTokenConsumer,
			ExpectedError:    errors.New(apptmpltest.NonSelfRegFlowErrorMsg),
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
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInputWithSelRegLabels, nil).Once()
				appTemplateConv.AssertNotCalled(t, "ToGraphQL")
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInput.Webhooks, gqlAppTemplateInput.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatReturnsErrorOnPrep,
			LabelSvcFn:       UnusedLabelService,
			Input:            gqlAppTemplateInput,
			Ctx:              ctxWithCertConsumer,
			ExpectedError:    errors.New(apptmpltest.SelfRegErrorMsg),
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
				appTemplateSvc.On("GetByFilters", txtest.CtxWithDBMatcher(), getAppTemplateFiltersForSelfReg).Return(nil, nil).Once()
				appTemplateSvc.On("CreateWithLabels", txtest.CtxWithDBMatcher(), *modelAppTemplateInput, labelsContainingSelfRegAndSubaccount).Return("", testError).Once()
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
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInput.Webhooks, gqlAppTemplateInput.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatDoesCleanup(labelsContainingSelfRegistration),
			LabelSvcFn:       UnusedLabelService,
			Ctx:              ctxWithCertConsumer,
			Input:            gqlAppTemplateInput,
			ExpectedError:    testError,
		},
		{
			Name: "Error when couldn't cast region label value to string",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInputWithSelRegLabels, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn:    UnusedWebhookConv,
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateInput.Webhooks, gqlAppTemplateInput.Webhooks),
			SelfRegManagerFn: apptmpltest.SelfRegManagerThatDoesPrepAndInitiatesCleanupButNotFinishIt(labelsContainingSelfRegistrationAndInvaidRegion),
			LabelSvcFn:       UnusedLabelService,
			Ctx:              ctxWithCertConsumer,
			Input:            gqlAppTemplateInput,
			ExpectedError:    errors.New("region label should be string"),
		},
		{
			Name:              "Returns error when validating app template name",
			TxFn:              txGen.ThatDoesntStartTransaction,
			AppTemplateSvcFn:  UnusedAppTemplateSvc,
			AppTemplateConvFn: UnusedAppTemplateConv,
			WebhookConvFn:     UnusedWebhookConv,
			WebhookSvcFn:      UnusedWebhookSvc,
			SelfRegManagerFn:  UnusedSelfRegManager,
			LabelSvcFn:        UnusedLabelService,
			Input:             fixGQLAppTemplateInputWithPlaceholderAndProvider(testName),
			ExpectedError:     errors.New("application template name \"bar\" does not comply with the following naming convention: \"SAP <product name>\""),
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
			labelService := testCase.LabelSvcFn()
			uuidSvc := uidSvcFn()
			ctx := ctxWithCertConsumer

			if testCase.Ctx != nil {
				ctx = testCase.Ctx
			}

			resolver := apptemplate.NewResolver(transact, nil, nil, appTemplateSvc, appTemplateConv, webhookSvc, webhookConverter, labelService, selfRegManager, uuidSvc, nil, nil, AppTemplateProductLabel, apiclient.OrdAggregatorClientConfig{})

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

		resolver := apptemplate.NewResolver(transact, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", apiclient.OrdAggregatorClientConfig{})

		// WHEN
		_, err := resolver.CreateApplicationTemplate(ctxWithCertConsumer, *gqlAppTemplateInputInvalid)

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

			resolver := apptemplate.NewResolver(transact, nil, nil, appTemplateSvc, nil, nil, nil, nil, nil, nil, nil, nil, "", apiclient.OrdAggregatorClientConfig{})

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
	region := "cf-eu-1"
	subaccountID := "consumer-id"

	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)
	ctx = consumer.SaveToContext(ctx, consumer.Consumer{ConsumerID: testConsumerID})

	txGen := txtest.NewTransactionContextGenerator(testError)

	globalSubaccountIDLabelKey := "global_subaccount_id"
	filters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(globalSubaccountIDLabelKey, fmt.Sprintf("\"%s\"", "consumer-id")),
	}

	jsonAppCreateInput := fixJSONApplicationCreateInput(testName)
	modelAppCreateInput := fixModelApplicationCreateInput(testName)
	modelAppWithLabelCreateInput := fixModelApplicationWithManagedLabelCreateInput(testName, "false")
	modelAppWithManagedTrueLabelCreateInput := fixModelApplicationWithManagedLabelCreateInput(testName, "true")
	gqlAppCreateInput := fixGQLApplicationCreateInput(testName)
	gqlAppCreateWithManagedTrueLabelInput := fixGQLApplicationCreateWithManagedTrueLabelInput(testName, "true")

	customID := "customTemplateID"
	modelAppTemplate := fixModelAppTemplateWithAppInputJSON(testID, testName, jsonAppCreateInput, fixModelApplicationTemplateWebhooks(testWebhookID, testID))
	modelAppTemplateCustomID := fixModelAppTemplateWithAppInputJSON(customID, testName, jsonAppCreateInput, fixModelApplicationTemplateWebhooks(testWebhookID, customID))

	appTemplateLabels := map[string]*model.Label{
		"global_subaccount_id": {
			Value: subaccountID,
		},
		"systemFieldDiscovery": {
			Value: true,
		},
		"region": {
			Value: region,
		},
	}
	modelApplication := fixModelApplication(testID, testName)
	gqlApplication := fixGQLApplication(testID, testName)

	gqlAppFromTemplateInput := fixGQLApplicationFromTemplateInput(testName)
	gqlAppFromTemplateWithManagedLabelInput := fixGQLApplicationFromTemplateWithManagedLabelInput(testName, "true")
	gqlAppFromTemplateWithIDInput := fixGQLApplicationFromTemplateInput(testName)
	gqlAppFromTemplateWithIDInput.ID = &customID
	modelAppFromTemplateInput := fixModelApplicationFromTemplateInput(testName)
	modelAppFromTemplateWithManagedLabelInput := fixModelApplicationFromTemplateWithManagedLabelInput(testName, "true")
	modelAppFromTemplateWithIDInput := fixModelApplicationFromTemplateInput(testName)
	modelAppFromTemplateWithIDInput.ID = &customID

	systemFieldDiscoveryLabelIsTrue := true
	testCases := []struct {
		Name                       string
		AppFromTemplateInput       graphql.ApplicationFromTemplateInput
		TxFn                       func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AppTemplateSvcFn           func() *automock.ApplicationTemplateService
		AppTemplateConvFn          func() *automock.ApplicationTemplateConverter
		WebhookSvcFn               func() *automock.WebhookService
		WebhookConvFn              func() *automock.WebhookConverter
		AppSvcFn                   func() *automock.ApplicationService
		AppConvFn                  func() *automock.ApplicationConverter
		SystemFieldDiscoveryEngine func() *automock.SystemFieldDiscoveryEngine
		ExpectedOutput             *graphql.Application
		ExpectedError              error
	}{
		{
			Name:                 "Success",
			TxFn:                 txGen.ThatSucceeds,
			AppFromTemplateInput: gqlAppFromTemplateInput,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{}, nil).Once()
				appTemplateSvc.On("ListByName", txtest.CtxWithDBMatcher(), testName).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, globalSubaccountIDLabelKey).Return(nil, apperrors.NewNotFoundError(resource.Label, "id")).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), modelAppTemplate.ID).Return(appTemplateLabels, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateInput, nil).Once()
				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				modelAppWithLabelCreateInputWithWebhook := modelAppWithLabelCreateInput
				modelAppWithLabelCreateInputWithWebhook.Webhooks = append(modelAppWithLabelCreateInput.Webhooks, &model.WebhookInput{
					Type: model.WebhookTypeSystemFieldDiscovery})
				appSvc.On("CreateFromTemplate", txtest.CtxWithDBMatcher(), modelAppWithLabelCreateInputWithWebhook, str.Ptr(testID), systemFieldDiscoveryLabelIsTrue).Return(testID, nil).Once()
				appSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&modelApplication, nil).Once()
				return appSvc
			},
			AppConvFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateRegisterInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				appConv.On("CreateInputFromGraphQL", mock.Anything, gqlAppCreateInput).Return(modelAppCreateInput, nil).Once()
				appConv.On("ToGraphQL", &modelApplication).Return(&gqlApplication).Once()
				return appConv
			},
			SystemFieldDiscoveryEngine: func() *automock.SystemFieldDiscoveryEngine {
				systemFieldDiscoveryEngine := &automock.SystemFieldDiscoveryEngine{}
				newWebhooks := modelAppCreateInput.Webhooks
				newWebhooks = append(newWebhooks, &model.WebhookInput{
					Type: model.WebhookTypeSystemFieldDiscovery,
				})
				systemFieldDiscoveryEngine.On("EnrichApplicationWebhookIfNeeded", txtest.CtxWithDBMatcher(), modelAppWithLabelCreateInput, systemFieldDiscoveryLabelIsTrue, region, subaccountID, modelAppTemplate.Name, testName).Return(newWebhooks, systemFieldDiscoveryLabelIsTrue).Once()
				systemFieldDiscoveryEngine.On("CreateLabelForApplicationWebhook", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return systemFieldDiscoveryEngine
			},
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: &gqlApplication,
			ExpectedError:  nil,
		},
		{
			Name:                 "SuccessWithIDField",
			AppFromTemplateInput: gqlAppFromTemplateWithIDInput,
			TxFn:                 txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{}, nil).Once()
				appTemplateSvc.On("ListByName", txtest.CtxWithDBMatcher(), testName).Return([]*model.ApplicationTemplate{modelAppTemplate, modelAppTemplateCustomID}, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), customID, globalSubaccountIDLabelKey).Return(nil, apperrors.NewNotFoundError(resource.Label, "id")).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, globalSubaccountIDLabelKey).Return(nil, apperrors.NewNotFoundError(resource.Label, "id")).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplateCustomID, modelAppFromTemplateWithIDInput.Values).Return(jsonAppCreateInput, nil).Once()
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), modelAppTemplateCustomID.ID).Return(map[string]*model.Label{}, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplateCustomID, gqlAppFromTemplateWithIDInput).Return(modelAppFromTemplateWithIDInput, nil).Once()
				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("CreateFromTemplate", txtest.CtxWithDBMatcher(), modelAppWithLabelCreateInput, str.Ptr(customID), !systemFieldDiscoveryLabelIsTrue).Return(testID, nil).Once()
				appSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&modelApplication, nil).Once()
				return appSvc
			},
			AppConvFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateRegisterInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				appConv.On("CreateInputFromGraphQL", mock.Anything, gqlAppCreateInput).Return(modelAppWithLabelCreateInput, nil).Once()
				appConv.On("ToGraphQL", &modelApplication).Return(&gqlApplication).Once()
				return appConv
			},
			SystemFieldDiscoveryEngine: func() *automock.SystemFieldDiscoveryEngine {
				systemFieldDiscoveryEngine := &automock.SystemFieldDiscoveryEngine{}
				systemFieldDiscoveryEngine.AssertNotCalled(t, "EnrichApplicationWebhookIfNeeded")
				systemFieldDiscoveryEngine.AssertNotCalled(t, "CreateLabelForApplicationWebhook")
				return systemFieldDiscoveryEngine
			},
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: &gqlApplication,
			ExpectedError:  nil,
		},
		{
			Name:                 "Success when managed label is present",
			TxFn:                 txGen.ThatSucceeds,
			AppFromTemplateInput: gqlAppFromTemplateWithManagedLabelInput,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{}, nil).Once()
				appTemplateSvc.On("ListByName", txtest.CtxWithDBMatcher(), testName).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, globalSubaccountIDLabelKey).Return(nil, apperrors.NewNotFoundError(resource.Label, "id")).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), modelAppTemplate.ID).Return(appTemplateLabels, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateWithManagedLabelInput).Return(modelAppFromTemplateWithManagedLabelInput, nil).Once()
				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				newModelInput := modelAppWithManagedTrueLabelCreateInput
				newModelInput.Webhooks = append(newModelInput.Webhooks, &model.WebhookInput{Type: model.WebhookTypeSystemFieldDiscovery})
				appSvc.On("CreateFromTemplate", txtest.CtxWithDBMatcher(), newModelInput, str.Ptr(testID), systemFieldDiscoveryLabelIsTrue).Return(testID, nil).Once()
				appSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&modelApplication, nil).Once()
				return appSvc
			},
			AppConvFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateRegisterInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateWithManagedTrueLabelInput, nil).Once()
				appConv.On("CreateInputFromGraphQL", mock.Anything, gqlAppCreateWithManagedTrueLabelInput).Return(modelAppWithManagedTrueLabelCreateInput, nil).Once()
				appConv.On("ToGraphQL", &modelApplication).Return(&gqlApplication).Once()
				return appConv
			},
			SystemFieldDiscoveryEngine: func() *automock.SystemFieldDiscoveryEngine {
				systemFieldDiscoveryEngine := &automock.SystemFieldDiscoveryEngine{}
				newWebhooks := modelAppWithManagedTrueLabelCreateInput.Webhooks
				newWebhooks = append(newWebhooks, &model.WebhookInput{
					Type: model.WebhookTypeSystemFieldDiscovery,
				})
				systemFieldDiscoveryEngine.On("EnrichApplicationWebhookIfNeeded", txtest.CtxWithDBMatcher(), modelAppWithManagedTrueLabelCreateInput, systemFieldDiscoveryLabelIsTrue, region, subaccountID, modelAppTemplate.Name, testName).Return(newWebhooks, systemFieldDiscoveryLabelIsTrue).Once()
				systemFieldDiscoveryEngine.On("CreateLabelForApplicationWebhook", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return systemFieldDiscoveryEngine
			},
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: &gqlApplication,
			ExpectedError:  nil,
		},
		{
			Name:                       "Returns error when transaction begin fails",
			AppFromTemplateInput:       gqlAppFromTemplateInput,
			TxFn:                       txGen.ThatFailsOnBegin,
			AppTemplateSvcFn:           UnusedAppTemplateSvc,
			AppTemplateConvFn:          UnusedAppTemplateConv,
			AppSvcFn:                   UnusedAppSvc,
			AppConvFn:                  UnusedAppConv,
			WebhookConvFn:              UnusedWebhookConv,
			WebhookSvcFn:               UnusedWebhookSvc,
			SystemFieldDiscoveryEngine: UnusedSystemFieldDiscoveryEngine,
			ExpectedOutput:             nil,
			ExpectedError:              testError,
		},
		{
			Name:                 "Returns error when list application templates by filters fails",
			AppFromTemplateInput: gqlAppFromTemplateInput,
			TxFn:                 txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			AppSvcFn:                   UnusedAppSvc,
			AppConvFn:                  UnusedAppConv,
			WebhookConvFn:              UnusedWebhookConv,
			WebhookSvcFn:               UnusedWebhookSvc,
			SystemFieldDiscoveryEngine: UnusedSystemFieldDiscoveryEngine,
			ExpectedOutput:             nil,
			ExpectedError:              testError,
		},
		{
			Name:                 "Returns error when list application templates by name fails",
			AppFromTemplateInput: gqlAppFromTemplateInput,
			TxFn:                 txGen.ThatDoesntExpectCommit,
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
			AppSvcFn:                   UnusedAppSvc,
			AppConvFn:                  UnusedAppConv,
			WebhookConvFn:              UnusedWebhookConv,
			WebhookSvcFn:               UnusedWebhookSvc,
			SystemFieldDiscoveryEngine: UnusedSystemFieldDiscoveryEngine,
			ExpectedOutput:             nil,
			ExpectedError:              testError,
		},
		{
			Name:                 "Returns error when get application template label fails",
			AppFromTemplateInput: gqlAppFromTemplateInput,
			TxFn:                 txGen.ThatDoesntExpectCommit,
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
			AppSvcFn:                   UnusedAppSvc,
			AppConvFn:                  UnusedAppConv,
			WebhookConvFn:              UnusedWebhookConv,
			WebhookSvcFn:               UnusedWebhookSvc,
			SystemFieldDiscoveryEngine: UnusedSystemFieldDiscoveryEngine,
			ExpectedOutput:             nil,
			ExpectedError:              testError,
		},
		{
			Name:                 "Returns error when list application templates by name cannot find application template",
			AppFromTemplateInput: gqlAppFromTemplateInput,
			TxFn:                 txGen.ThatDoesntExpectCommit,
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
			AppSvcFn:                   UnusedAppSvc,
			AppConvFn:                  UnusedAppConv,
			WebhookConvFn:              UnusedWebhookConv,
			WebhookSvcFn:               UnusedWebhookSvc,
			SystemFieldDiscoveryEngine: UnusedSystemFieldDiscoveryEngine,
			ExpectedOutput:             nil,
			ExpectedError:              errors.New("application template with name \"bar\" and consumer id \"consumer-id\" not found"),
		},
		{
			Name:                 "Returns error when list application templates by name return more than one application template",
			AppFromTemplateInput: gqlAppFromTemplateInput,
			TxFn:                 txGen.ThatDoesntExpectCommit,
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
			AppSvcFn:                   UnusedAppSvc,
			AppConvFn:                  UnusedAppConv,
			WebhookConvFn:              UnusedWebhookConv,
			WebhookSvcFn:               UnusedWebhookSvc,
			SystemFieldDiscoveryEngine: UnusedSystemFieldDiscoveryEngine,
			ExpectedOutput:             nil,
			ExpectedError:              errors.New("unexpected number of application templates. found 2"),
		},
		{
			Name:                 "Returns error when preparing ApplicationCreateInputJSON fails",
			AppFromTemplateInput: gqlAppFromTemplateInput,
			TxFn:                 txGen.ThatDoesntExpectCommit,
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
			AppSvcFn:                   UnusedAppSvc,
			AppConvFn:                  UnusedAppConv,
			WebhookConvFn:              UnusedWebhookConv,
			WebhookSvcFn:               UnusedWebhookSvc,
			SystemFieldDiscoveryEngine: UnusedSystemFieldDiscoveryEngine,
			ExpectedOutput:             nil,
			ExpectedError:              testError,
		},
		{
			Name:                 "Returns error when CreateRegisterInputJSONToGQL fails",
			AppFromTemplateInput: gqlAppFromTemplateInput,
			TxFn:                 txGen.ThatDoesntExpectCommit,
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
				appConv.On("CreateRegisterInputJSONToGQL", jsonAppCreateInput).Return(graphql.ApplicationRegisterInput{}, testError).Once()
				return appConv
			},
			WebhookConvFn:              UnusedWebhookConv,
			WebhookSvcFn:               UnusedWebhookSvc,
			SystemFieldDiscoveryEngine: UnusedSystemFieldDiscoveryEngine,
			ExpectedOutput:             nil,
			ExpectedError:              testError,
		},
		{
			Name:                 "Returns error when ApplicationCreateInput validation fails",
			AppFromTemplateInput: gqlAppFromTemplateInput,
			TxFn:                 txGen.ThatDoesntExpectCommit,
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
				appConv.On("CreateRegisterInputJSONToGQL", jsonAppCreateInput).Return(graphql.ApplicationRegisterInput{}, nil).Once()
				return appConv
			},
			WebhookConvFn:              UnusedWebhookConv,
			WebhookSvcFn:               UnusedWebhookSvc,
			SystemFieldDiscoveryEngine: UnusedSystemFieldDiscoveryEngine,
			ExpectedOutput:             nil,
			ExpectedError:              errors.New("name=cannot be blank"),
		},
		{
			Name:                 "Returns error when listing labels for Application Template fails",
			AppFromTemplateInput: gqlAppFromTemplateInput,
			TxFn:                 txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), modelAppTemplate.ID).Return(nil, testError).Once()
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
				appConv.On("CreateInputFromGraphQL", mock.Anything, gqlAppCreateInput).Return(modelAppCreateInput, nil).Once()
				appConv.On("CreateRegisterInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				return appConv
			},
			SystemFieldDiscoveryEngine: func() *automock.SystemFieldDiscoveryEngine {
				systemFieldDiscoveryEngine := &automock.SystemFieldDiscoveryEngine{}
				newWebhooks := modelAppCreateInput.Webhooks
				newWebhooks = append(newWebhooks, &model.WebhookInput{
					Type: model.WebhookTypeSystemFieldDiscovery,
				})
				systemFieldDiscoveryEngine.On("EnrichApplicationWebhookIfNeeded", txtest.CtxWithDBMatcher(), modelAppCreateInput, systemFieldDiscoveryLabelIsTrue, region, subaccountID, modelAppTemplate.Name, testName).Return(newWebhooks, systemFieldDiscoveryLabelIsTrue).Once()
				systemFieldDiscoveryEngine.AssertNotCalled(t, "CreateLabelForApplicationWebhook")
				return systemFieldDiscoveryEngine
			},
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: nil,
			ExpectedError:  errors.New("error while listing labels for Application Template"),
		},
		{
			Name:                 "Returns error when creating Application fails",
			AppFromTemplateInput: gqlAppFromTemplateInput,
			TxFn:                 txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), modelAppTemplate.ID).Return(appTemplateLabels, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateInput, nil).Once()
				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				modelAppWithLabelCreateInputWithWebhook := modelAppWithLabelCreateInput
				modelAppWithLabelCreateInputWithWebhook.Webhooks = append(modelAppWithLabelCreateInput.Webhooks, &model.WebhookInput{
					Type: model.WebhookTypeSystemFieldDiscovery})
				appSvc.On("CreateFromTemplate", txtest.CtxWithDBMatcher(), modelAppWithLabelCreateInputWithWebhook, str.Ptr(testID), systemFieldDiscoveryLabelIsTrue).Return("", testError).Once()
				return appSvc
			},
			AppConvFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateInputFromGraphQL", mock.Anything, gqlAppCreateInput).Return(modelAppCreateInput, nil).Once()
				appConv.On("CreateRegisterInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				return appConv
			},
			SystemFieldDiscoveryEngine: func() *automock.SystemFieldDiscoveryEngine {
				systemFieldDiscoveryEngine := &automock.SystemFieldDiscoveryEngine{}
				newWebhooks := modelAppCreateInput.Webhooks
				newWebhooks = append(newWebhooks, &model.WebhookInput{
					Type: model.WebhookTypeSystemFieldDiscovery,
				})
				systemFieldDiscoveryEngine.On("EnrichApplicationWebhookIfNeeded", txtest.CtxWithDBMatcher(), modelAppWithLabelCreateInput, systemFieldDiscoveryLabelIsTrue, region, subaccountID, modelAppTemplate.Name, testName).Return(newWebhooks, systemFieldDiscoveryLabelIsTrue).Once()
				systemFieldDiscoveryEngine.AssertNotCalled(t, "CreateLabelForApplicationWebhook")
				return systemFieldDiscoveryEngine
			},
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
		{
			Name:                 "Returns error when getting Application fails",
			AppFromTemplateInput: gqlAppFromTemplateInput,
			TxFn:                 txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), modelAppTemplate.ID).Return(appTemplateLabels, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateInput, nil).Once()
				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				modelAppWithLabelCreateInputWithWebhook := modelAppWithLabelCreateInput
				modelAppWithLabelCreateInputWithWebhook.Webhooks = append(modelAppWithLabelCreateInput.Webhooks, &model.WebhookInput{
					Type: model.WebhookTypeSystemFieldDiscovery})
				appSvc.On("CreateFromTemplate", txtest.CtxWithDBMatcher(), modelAppWithLabelCreateInputWithWebhook, str.Ptr(testID), systemFieldDiscoveryLabelIsTrue).Return(testID, nil).Once()
				appSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return appSvc
			},
			AppConvFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateInputFromGraphQL", mock.Anything, gqlAppCreateInput).Return(modelAppCreateInput, nil).Once()
				appConv.On("CreateRegisterInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				return appConv
			},
			SystemFieldDiscoveryEngine: func() *automock.SystemFieldDiscoveryEngine {
				systemFieldDiscoveryEngine := &automock.SystemFieldDiscoveryEngine{}
				newWebhooks := modelAppCreateInput.Webhooks
				newWebhooks = append(newWebhooks, &model.WebhookInput{
					Type: model.WebhookTypeSystemFieldDiscovery,
				})
				systemFieldDiscoveryEngine.On("EnrichApplicationWebhookIfNeeded", txtest.CtxWithDBMatcher(), modelAppWithLabelCreateInput, systemFieldDiscoveryLabelIsTrue, region, subaccountID, modelAppTemplate.Name, testName).Return(newWebhooks, systemFieldDiscoveryLabelIsTrue).Once()
				systemFieldDiscoveryEngine.On("CreateLabelForApplicationWebhook", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return systemFieldDiscoveryEngine
			},
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
		{
			Name:                 "Returns error when creating label for application webhook fails",
			AppFromTemplateInput: gqlAppFromTemplateInput,
			TxFn:                 txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), modelAppTemplate.ID).Return(appTemplateLabels, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateInput, nil).Once()
				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				modelAppWithLabelCreateInputWithWebhook := modelAppWithLabelCreateInput
				modelAppWithLabelCreateInputWithWebhook.Webhooks = append(modelAppWithLabelCreateInput.Webhooks, &model.WebhookInput{
					Type: model.WebhookTypeSystemFieldDiscovery})
				appSvc.On("CreateFromTemplate", txtest.CtxWithDBMatcher(), modelAppWithLabelCreateInputWithWebhook, str.Ptr(testID), systemFieldDiscoveryLabelIsTrue).Return(testID, nil).Once()
				return appSvc
			},
			AppConvFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateInputFromGraphQL", mock.Anything, gqlAppCreateInput).Return(modelAppCreateInput, nil).Once()
				appConv.On("CreateRegisterInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				return appConv
			},
			SystemFieldDiscoveryEngine: func() *automock.SystemFieldDiscoveryEngine {
				systemFieldDiscoveryEngine := &automock.SystemFieldDiscoveryEngine{}
				newWebhooks := modelAppCreateInput.Webhooks
				newWebhooks = append(newWebhooks, &model.WebhookInput{
					Type: model.WebhookTypeSystemFieldDiscovery,
				})
				systemFieldDiscoveryEngine.On("EnrichApplicationWebhookIfNeeded", txtest.CtxWithDBMatcher(), modelAppWithLabelCreateInput, systemFieldDiscoveryLabelIsTrue, region, subaccountID, modelAppTemplate.Name, testName).Return(newWebhooks, systemFieldDiscoveryLabelIsTrue).Once()
				systemFieldDiscoveryEngine.On("CreateLabelForApplicationWebhook", txtest.CtxWithDBMatcher(), testID).Return(testError).Once()
				return systemFieldDiscoveryEngine
			},
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
		{
			Name:                 "Returns error when committing transaction fails",
			AppFromTemplateInput: gqlAppFromTemplateInput,
			TxFn:                 txGen.ThatFailsOnCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), modelAppTemplate.ID).Return(appTemplateLabels, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateInput, nil).Once()
				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				modelAppWithLabelCreateInputWithWebhook := modelAppWithLabelCreateInput
				modelAppWithLabelCreateInputWithWebhook.Webhooks = append(modelAppWithLabelCreateInput.Webhooks, &model.WebhookInput{
					Type: model.WebhookTypeSystemFieldDiscovery})
				appSvc.On("CreateFromTemplate", txtest.CtxWithDBMatcher(), modelAppWithLabelCreateInputWithWebhook, str.Ptr(testID), systemFieldDiscoveryLabelIsTrue).Return(testID, nil).Once()
				appSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&modelApplication, nil).Once()
				return appSvc
			},
			AppConvFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateInputFromGraphQL", mock.Anything, gqlAppCreateInput).Return(modelAppCreateInput, nil).Once()
				appConv.On("CreateRegisterInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				return appConv
			},
			SystemFieldDiscoveryEngine: func() *automock.SystemFieldDiscoveryEngine {
				systemFieldDiscoveryEngine := &automock.SystemFieldDiscoveryEngine{}
				newWebhooks := modelAppCreateInput.Webhooks
				newWebhooks = append(newWebhooks, &model.WebhookInput{
					Type: model.WebhookTypeSystemFieldDiscovery,
				})
				systemFieldDiscoveryEngine.On("EnrichApplicationWebhookIfNeeded", txtest.CtxWithDBMatcher(), modelAppWithLabelCreateInput, systemFieldDiscoveryLabelIsTrue, region, subaccountID, modelAppTemplate.Name, testName).Return(newWebhooks, systemFieldDiscoveryLabelIsTrue).Once()
				systemFieldDiscoveryEngine.On("CreateLabelForApplicationWebhook", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return systemFieldDiscoveryEngine
			},
			WebhookConvFn:  UnusedWebhookConv,
			WebhookSvcFn:   UnusedWebhookSvc,
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
		{
			Name:                 "ErrorWhenNoTemplatesWithGivenIDFound",
			AppFromTemplateInput: gqlAppFromTemplateWithIDInput,
			TxFn:                 txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListByFilters", txtest.CtxWithDBMatcher(), filters).Return([]*model.ApplicationTemplate{}, nil).Once()
				appTemplateSvc.On("ListByName", txtest.CtxWithDBMatcher(), testName).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				appTemplateSvc.On("GetLabel", txtest.CtxWithDBMatcher(), testID, globalSubaccountIDLabelKey).Return(nil, apperrors.NewNotFoundError(resource.Label, "id")).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn:          UnusedAppTemplateConv,
			AppSvcFn:                   UnusedAppSvc,
			AppConvFn:                  UnusedAppConv,
			WebhookConvFn:              UnusedWebhookConv,
			WebhookSvcFn:               UnusedWebhookSvc,
			SystemFieldDiscoveryEngine: UnusedSystemFieldDiscoveryEngine,
			ExpectedError:              errors.New("application template with id customTemplateID and consumer id \"consumer-id\" not found"),
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
			systemFieldDiscoveryEngine := testCase.SystemFieldDiscoveryEngine()

			resolver := apptemplate.NewResolver(transact, appSvc, appConv, appTemplateSvc, appTemplateConv, webhookSvc, webhookConverter, nil, nil, nil, systemFieldDiscoveryEngine, nil, "", apiclient.OrdAggregatorClientConfig{})

			// WHEN
			result, err := resolver.RegisterApplicationFromTemplate(ctx, testCase.AppFromTemplateInput)

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
	shouldOverride := true
	testCases := []struct {
		Name              string
		TxFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AppTemplateSvcFn  func() *automock.ApplicationTemplateService
		AppTemplateConvFn func() *automock.ApplicationTemplateConverter
		SelfRegManagerFn  func() *automock.SelfRegisterManager
		WebhookSvcFn      func() *automock.WebhookService
		WebhookConvFn     func() *automock.WebhookConverter
		Input             *graphql.ApplicationTemplateUpdateInput
		InputOverride     *bool
		ExpectedOutput    *graphql.ApplicationTemplate
		ExpectedError     error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testID).Return(labels, nil).Once()
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, false, *modelAppTemplateInput).Return(nil).Once()
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
			WebhookSvcFn:   SuccessfulWebhookSvc(gqlAppTemplateUpdateInputWithProvider.Webhooks, gqlAppTemplateUpdateInputWithProvider.Webhooks),
			Input:          gqlAppTemplateUpdateInput,
			ExpectedOutput: gqlAppTemplate,
		},
		{
			Name: "Success with override",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testID).Return(labels, nil).Once()
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, shouldOverride, *modelAppTemplateInput).Return(nil).Once()
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
			WebhookSvcFn:   SuccessfulWebhookSvc(gqlAppTemplateUpdateInputWithProvider.Webhooks, gqlAppTemplateUpdateInputWithProvider.Webhooks),
			Input:          gqlAppTemplateUpdateInput,
			InputOverride:  &shouldOverride,
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
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, false, *modelAppTemplateInput).Return(nil).Once()
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
			WebhookSvcFn:   SuccessfulWebhookSvc(gqlAppTemplateUpdateInputWithProvider.Webhooks, gqlAppTemplateUpdateInputWithProvider.Webhooks),
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
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateUpdateInputWithProvider.Webhooks, gqlAppTemplateUpdateInputWithProvider.Webhooks),
			Input:            gqlAppTemplateUpdateInput,
			ExpectedError:    testError,
		},
		{
			Name: "Returns error when updating application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testID).Return(labels, nil).Once()
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, false, *modelAppTemplateInput).Return(testError).Once()
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
			WebhookSvcFn:  SuccessfulWebhookSvc(gqlAppTemplateUpdateInputWithProvider.Webhooks, gqlAppTemplateUpdateInputWithProvider.Webhooks),
			Input:         gqlAppTemplateUpdateInput,
			ExpectedError: testError,
		},
		{
			Name: "Returns error when getting application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testID).Return(labels, nil).Once()
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, false, *modelAppTemplateInput).Return(nil).Once()
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
			WebhookSvcFn:  SuccessfulWebhookSvc(gqlAppTemplateUpdateInputWithProvider.Webhooks, gqlAppTemplateUpdateInputWithProvider.Webhooks),
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
			WebhookSvcFn:     SuccessfulWebhookSvc(gqlAppTemplateUpdateInputWithProvider.Webhooks, gqlAppTemplateUpdateInputWithProvider.Webhooks),
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
			WebhookSvcFn:  SuccessfulWebhookSvc(gqlAppTemplateUpdateInputWithProvider.Webhooks, gqlAppTemplateUpdateInputWithProvider.Webhooks),
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
			WebhookSvcFn:  SuccessfulWebhookSvc(gqlAppTemplateUpdateInputWithProvider.Webhooks, gqlAppTemplateUpdateInputWithProvider.Webhooks),
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
			WebhookSvcFn:  SuccessfulWebhookSvc(gqlAppTemplateUpdateInputWithProvider.Webhooks, gqlAppTemplateUpdateInputWithProvider.Webhooks),
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
			WebhookSvcFn:  SuccessfulWebhookSvc(gqlAppTemplateUpdateInputWithProvider.Webhooks, gqlAppTemplateUpdateInputWithProvider.Webhooks),
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
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, false, *modelAppTemplateInput).Return(nil).Once()
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
			WebhookSvcFn:  SuccessfulWebhookSvc(gqlAppTemplateUpdateInputWithProvider.Webhooks, gqlAppTemplateUpdateInputWithProvider.Webhooks),
			Input:         gqlAppTemplateUpdateInput,
			ExpectedError: testError,
		},
		{
			Name: "Returns error when can't convert application template to graphql",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testID).Return(labels, nil).Once()
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, false, *modelAppTemplateInput).Return(nil).Once()
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
			WebhookSvcFn:  SuccessfulWebhookSvc(gqlAppTemplateUpdateInputWithProvider.Webhooks, gqlAppTemplateUpdateInputWithProvider.Webhooks),
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

			resolver := apptemplate.NewResolver(transact, nil, nil, appTemplateSvc, appTemplateConv, webhookSvc, webhookConverter, nil, selfRegManager, nil, nil, nil, "", apiclient.OrdAggregatorClientConfig{})

			// WHEN
			result, err := resolver.UpdateApplicationTemplate(ctx, testID, testCase.InputOverride, *testCase.Input)

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

		resolver := apptemplate.NewResolver(transact, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", apiclient.OrdAggregatorClientConfig{})

		// WHEN
		override := false
		_, err := resolver.UpdateApplicationTemplate(ctx, testID, &override, *gqlAppTemplateUpdateInputInvalid)

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
		Name                 string
		TxFn                 func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AppTemplateSvcFn     func() *automock.ApplicationTemplateService
		AppTemplateConvFn    func() *automock.ApplicationTemplateConverter
		WebhookSvcFn         func() *automock.WebhookService
		WebhookConvFn        func() *automock.WebhookConverter
		SelfRegManagerFn     func() *automock.SelfRegisterManager
		CertSubjMappingSvcFn func() *automock.CertSubjectMappingService
		ExpectedOutput       *graphql.ApplicationTemplate
		ExpectedError        error
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
			CertSubjMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjMappingSvc := &automock.CertSubjectMappingService{}
				certSubjMappingSvc.On("DeleteByConsumerID", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return certSubjMappingSvc
			},
			ExpectedOutput: gqlAppTemplate,
		},
		{
			Name: "Returns error when deleting cert subject mappings by consumer id failed",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Twice()
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
			AppTemplateConvFn: UnusedAppTemplateConv,
			WebhookConvFn:     UnusedWebhookConv,
			WebhookSvcFn:      UnusedWebhookSvc,
			SelfRegManagerFn:  apptmpltest.SelfRegManagerThatDoesCleanupWithNoErrors,
			CertSubjMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjMappingSvc := &automock.CertSubjectMappingService{}
				certSubjMappingSvc.On("DeleteByConsumerID", txtest.CtxWithDBMatcher(), testID).Return(testError).Once()
				return certSubjMappingSvc
			},
			ExpectedError: testError,
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
			WebhookConvFn:        UnusedWebhookConv,
			WebhookSvcFn:         UnusedWebhookSvc,
			SelfRegManagerFn:     apptmpltest.NoopSelfRegManager,
			CertSubjMappingSvcFn: UnusedCertSubjMappingSvc,
			ExpectedError:        testError,
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
			WebhookConvFn:        UnusedWebhookConv,
			WebhookSvcFn:         UnusedWebhookSvc,
			SelfRegManagerFn:     apptmpltest.SelfRegManagerThatDoesCleanupWithNoErrors,
			CertSubjMappingSvcFn: UnusedCertSubjMappingSvc,
			ExpectedError:        testError,
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
			WebhookConvFn:        UnusedWebhookConv,
			WebhookSvcFn:         UnusedWebhookSvc,
			SelfRegManagerFn:     apptmpltest.NoopSelfRegManager,
			CertSubjMappingSvcFn: UnusedCertSubjMappingSvc,
			ExpectedError:        testError,
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
			WebhookConvFn:        UnusedWebhookConv,
			WebhookSvcFn:         UnusedWebhookSvc,
			SelfRegManagerFn:     apptmpltest.SelfRegManagerReturnsDistinguishingLabel,
			CertSubjMappingSvcFn: UnusedCertSubjMappingSvc,
			ExpectedError:        testError,
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
			CertSubjMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjMappingSvc := &automock.CertSubjectMappingService{}
				certSubjMappingSvc.On("DeleteByConsumerID", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return certSubjMappingSvc
			},
			ExpectedError: testError,
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
			CertSubjMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjMappingSvc := &automock.CertSubjectMappingService{}
				certSubjMappingSvc.On("DeleteByConsumerID", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return certSubjMappingSvc
			},
			ExpectedError: testError,
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
			WebhookConvFn:        UnusedWebhookConv,
			WebhookSvcFn:         UnusedWebhookSvc,
			SelfRegManagerFn:     apptmpltest.SelfRegManagerReturnsDistinguishingLabel,
			CertSubjMappingSvcFn: UnusedCertSubjMappingSvc,
			ExpectedError:        testError,
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
			WebhookConvFn:        UnusedWebhookConv,
			WebhookSvcFn:         UnusedWebhookSvc,
			SelfRegManagerFn:     apptmpltest.SelfRegManagerReturnsDistinguishingLabel,
			CertSubjMappingSvcFn: UnusedCertSubjMappingSvc,
			ExpectedError:        testError,
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
			WebhookConvFn:        UnusedWebhookConv,
			WebhookSvcFn:         UnusedWebhookSvc,
			SelfRegManagerFn:     apptmpltest.SelfRegManagerReturnsDistinguishingLabel,
			CertSubjMappingSvcFn: UnusedCertSubjMappingSvc,
			ExpectedOutput:       nil,
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
			WebhookConvFn:        UnusedWebhookConv,
			WebhookSvcFn:         UnusedWebhookSvc,
			SelfRegManagerFn:     apptmpltest.SelfRegManagerThatReturnsErrorOnCleanup,
			CertSubjMappingSvcFn: UnusedCertSubjMappingSvc,
			ExpectedError:        errors.New(apptmpltest.SelfRegErrorMsg),
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
			WebhookConvFn:        UnusedWebhookConv,
			WebhookSvcFn:         UnusedWebhookSvc,
			SelfRegManagerFn:     apptmpltest.SelfRegManagerThatDoesCleanupWithNoErrors,
			CertSubjMappingSvcFn: UnusedCertSubjMappingSvc,
			ExpectedError:        testError,
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
			certSubjMappingSvc := testCase.CertSubjMappingSvcFn()

			resolver := apptemplate.NewResolver(transact, nil, nil, appTemplateSvc, appTemplateConv, webhookSvc, webhookConverter, nil, selfRegManager, uuidSvc, nil, certSubjMappingSvc, "", apiclient.OrdAggregatorClientConfig{})

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

			mock.AssertExpectationsForObjects(t, persist, transact, appTemplateSvc, appTemplateConv, selfRegManager, certSubjMappingSvc)
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

func UnusedCertSubjMappingSvc() *automock.CertSubjectMappingService {
	return &automock.CertSubjectMappingService{}
}

func UnusedWebhookSvc() *automock.WebhookService {
	return &automock.WebhookService{}
}

func UnusedLabelService() *automock.LabelService {
	return &automock.LabelService{}
}
func UnusedSystemFieldDiscoveryEngine() *automock.SystemFieldDiscoveryEngine {
	return &automock.SystemFieldDiscoveryEngine{}
}
func SuccessfulWebhookSvc(webhooksInput, enriched []*graphql.WebhookInput) func() *automock.WebhookService {
	return func() *automock.WebhookService {
		svc := &automock.WebhookService{}
		svc.On("EnrichWebhooksWithTenantMappingWebhooks", webhooksInput).Return(enriched, nil)
		return svc
	}
}
