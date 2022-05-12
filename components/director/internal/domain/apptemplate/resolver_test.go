package apptemplate_test

import (
	"context"
	"errors"
	"testing"

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
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
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
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedOutput: nil,
		},
		{
			Name: "Returns error when getting application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
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
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
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

			resolver := apptemplate.NewResolver(transact, nil, nil, appTemplateSvc, appTemplateConv, webhookSvc, webhookConverter)

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

	testCases := []struct {
		Name              string
		TxFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AppTemplateSvcFn  func() *automock.ApplicationTemplateService
		AppTemplateConvFn func() *automock.ApplicationTemplateConverter
		WebhookSvcFn      func() *automock.WebhookService
		WebhookConvFn     func() *automock.WebhookConverter
		ExpectedOutput    *graphql.ApplicationTemplatePage
		ExpectedError     error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(modelPage, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("MultipleToGraphQL", modelAppTemplates).Return(gqlAppTemplates, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedOutput: &gqlPage,
		},
		{
			Name: "Returns error when getting application templates failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(model.ApplicationTemplatePage{}, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(modelPage, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when can't convert at least one of application templates to graphql",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(modelPage, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("MultipleToGraphQL", modelAppTemplates).Return(nil, testError).Once()
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
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

			resolver := apptemplate.NewResolver(transact, nil, nil, appTemplateSvc, appTemplateConv, webhookSvc, webhookConverter)

			// WHEN
			result, err := resolver.ApplicationTemplates(ctx, &first, &gqlAfter)

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
			PersistenceFn: txtest.PersistenceContextThatDoesntExpectCommit,
			WebhookServiceFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			WebhookConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedErr: testErr,
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
			WebhookConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			webhookSvc := testCase.WebhookServiceFn()
			converter := testCase.WebhookConverterFn()

			mockPersistence := testCase.PersistenceFn()
			mockTransactioner := testCase.TransactionerFn(mockPersistence)

			resolver := apptemplate.NewResolver(mockTransactioner, nil, nil, nil, nil, webhookSvc, converter)

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

	modelAppTemplate := fixModelApplicationTemplate(testID, testName, fixModelApplicationWebhooks(testWebhookID, testID))
	modelAppTemplateInput := fixModelAppTemplateInput(testName, appInputJSONString)
	gqlAppTemplate := fixGQLAppTemplate(testID, testName, fixGQLApplicationTemplateWebhooks(testWebhookID, testID))
	gqlAppTemplateInput := fixGQLAppTemplateInputWithPlaceholder(testName)

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
				appTemplateSvc.On("Create", txtest.CtxWithDBMatcher(), *modelAppTemplateInput).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplate, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedOutput: gqlAppTemplate,
		},
		{
			Name: "Returns error when can't convert input from graphql",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(model.ApplicationTemplateInput{}, testError).Once()
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when creating application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Create", txtest.CtxWithDBMatcher(), *modelAppTemplateInput).Return("", testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when getting application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Create", txtest.CtxWithDBMatcher(), *modelAppTemplateInput).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Create", txtest.CtxWithDBMatcher(), *modelAppTemplateInput).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when can't convert application template to graphql",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Create", txtest.CtxWithDBMatcher(), *modelAppTemplateInput).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(nil, testError).Once()
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
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

			resolver := apptemplate.NewResolver(transact, nil, nil, appTemplateSvc, appTemplateConv, webhookSvc, webhookConverter)

			// WHEN
			result, err := resolver.CreateApplicationTemplate(ctx, *gqlAppTemplateInput)

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
	t.Run("Returns error when application template inputs url template has invalid method", func(t *testing.T) {
		gqlAppTemplateInputInvalid := fixGQLAppTemplateInputInvalidAppInputURLTemplateMethod(testName)
		expectedError := errors.New("failed to parse webhook url template")
		_, transact := txGen.ThatSucceeds()

		resolver := apptemplate.NewResolver(transact, nil, nil, nil, nil, nil, nil)

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

			resolver := apptemplate.NewResolver(transact, nil, nil, appTemplateSvc, nil, nil, nil)

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

	txGen := txtest.NewTransactionContextGenerator(testError)

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
				appTemplateSvc.On("GetByName", txtest.CtxWithDBMatcher(), testName).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", gqlAppFromTemplateInput).Return(modelAppFromTemplateInput).Once()
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
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedOutput: &gqlApplication,
			ExpectedError:  nil,
		},
		{
			Name: "Returns error when transaction begin fails",
			TxFn: txGen.ThatFailsOnBegin,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}

				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				return appSvc
			},
			AppConvFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				return appConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
		{
			Name: "Returns error when getting Application Template fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByName", txtest.CtxWithDBMatcher(), testName).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", gqlAppFromTemplateInput).Return(modelAppFromTemplateInput).Once()
				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				return appSvc
			},
			AppConvFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				return appConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
		{
			Name: "Returns error when preparing ApplicationCreateInputJSON fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByName", txtest.CtxWithDBMatcher(), testName).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return("", testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", gqlAppFromTemplateInput).Return(modelAppFromTemplateInput).Once()
				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				return appSvc
			},
			AppConvFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				return appConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
		{
			Name: "Returns error when CreateInputJSONToGQL fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByName", txtest.CtxWithDBMatcher(), testName).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", gqlAppFromTemplateInput).Return(modelAppFromTemplateInput).Once()
				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				return appSvc
			},
			AppConvFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateInputJSONToGQL", jsonAppCreateInput).Return(graphql.ApplicationRegisterInput{}, testError).Once()
				return appConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
		{
			Name: "Returns error when ApplicationCreateInput validation fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByName", txtest.CtxWithDBMatcher(), testName).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", gqlAppFromTemplateInput).Return(modelAppFromTemplateInput).Once()
				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				return appSvc
			},
			AppConvFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateInputJSONToGQL", jsonAppCreateInput).Return(graphql.ApplicationRegisterInput{}, nil).Once()
				return appConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedOutput: nil,
			ExpectedError:  errors.New("name=cannot be blank"),
		},
		{
			Name: "Returns error when creating Application fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByName", txtest.CtxWithDBMatcher(), testName).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", gqlAppFromTemplateInput).Return(modelAppFromTemplateInput).Once()
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
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
		{
			Name: "Returns error when getting Application fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByName", txtest.CtxWithDBMatcher(), testName).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", gqlAppFromTemplateInput).Return(modelAppFromTemplateInput).Once()
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
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
		{
			Name: "Returns error when committing transaction fails",
			TxFn: txGen.ThatFailsOnCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByName", txtest.CtxWithDBMatcher(), testName).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", gqlAppFromTemplateInput).Return(modelAppFromTemplateInput).Once()
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
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
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

			resolver := apptemplate.NewResolver(transact, appSvc, appConv, appTemplateSvc, appTemplateConv, webhookSvc, webhookConverter)

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
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedOutput: gqlAppTemplate,
		},
		{
			Name: "Returns error when can't convert input from graphql",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("UpdateInputFromGraphQL", *gqlAppTemplateUpdateInput).Return(model.ApplicationTemplateUpdateInput{}, testError).Once()
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when updating application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, *modelAppTemplateInput).Return(testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("UpdateInputFromGraphQL", *gqlAppTemplateUpdateInput).Return(*modelAppTemplateInput, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when getting application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, *modelAppTemplateInput).Return(nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("UpdateInputFromGraphQL", *gqlAppTemplateUpdateInput).Return(*modelAppTemplateInput, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, *modelAppTemplateInput).Return(nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("UpdateInputFromGraphQL", *gqlAppTemplateUpdateInput).Return(*modelAppTemplateInput, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when can't convert application template to graphql",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
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
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
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

			resolver := apptemplate.NewResolver(transact, nil, nil, appTemplateSvc, appTemplateConv, webhookSvc, webhookConverter)

			// WHEN
			result, err := resolver.UpdateApplicationTemplate(ctx, testID, *gqlAppTemplateUpdateInput)

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

	t.Run("Returns error when application template inputs url template has invalid method", func(t *testing.T) {
		gqlAppTemplateUpdateInputInvalid := fixGQLAppTemplateUpdateInputInvalidAppInput(testName)
		expectedError := errors.New("failed to parse webhook url template")
		_, transact := txGen.ThatSucceeds()

		resolver := apptemplate.NewResolver(transact, nil, nil, nil, nil, nil, nil)

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
				appTemplateSvc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplate, nil).Once()
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedOutput: gqlAppTemplate,
		},
		{
			Name: "Returns error when getting application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when deleting application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when can't convert application template to graphql",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(nil, testError).Once()
				return appTemplateConv
			},
			WebhookConvFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
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
			resolver := apptemplate.NewResolver(transact, nil, nil, appTemplateSvc, appTemplateConv, webhookSvc, webhookConverter)

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
		})
	}
}
