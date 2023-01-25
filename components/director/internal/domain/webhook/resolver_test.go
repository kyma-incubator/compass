package webhook_test

import (
	"context"
	"errors"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	"github.com/stretchr/testify/require"

	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"

	"github.com/stretchr/testify/assert"
)

func TestResolver_AddWebhook(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	givenAppID := "foo"
	givenAppTemplateID := "test_app_template"
	givenRuntimeID := "test_runtime"
	givenFormationTemplateID := "ftID"
	id := "bar"
	gqlWebhookInput := fixGQLWebhookInput("foo")
	modelWebhookInput := fixModelWebhookInput("foo")

	gqlWebhook := fixGQLWebhook(id, "", "")
	modelWebhook := fixApplicationModelWebhook(id, givenAppID, givenTenant(), "foo", time.Time{})

	testCases := []struct {
		Name                       string
		PersistenceFn              func() *persistenceautomock.PersistenceTx
		TransactionerFn            func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn                  func() *automock.WebhookService
		AppServiceFn               func() *automock.ApplicationService
		AppTemplateServiceFn       func() *automock.ApplicationTemplateService
		RuntimeServiceFn           func() *automock.RuntimeService
		FormationTemplateServiceFn func() *automock.FormationTemplateService
		ConverterFn                func() *automock.WebhookConverter
		ExpectedWebhook            *graphql.Webhook
		ExpectedErr                error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), givenAppID, *modelWebhookInput, model.ApplicationWebhookReference).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id, model.ApplicationWebhookReference).Return(modelWebhook, nil).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), givenAppID).Return(true, nil).Once()
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook, nil).Once()
				return conv
			},
			ExpectedWebhook: gqlWebhook,
			ExpectedErr:     nil,
		},
		{
			Name:          "Returns error on starting transaction",
			PersistenceFn: txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, givenError()).Once()
				return transact
			},
			ServiceFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			AppServiceFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			ConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedErr: givenError(),
		},
		{
			Name:            "Returns error on webhook conversion",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatDoesARollback,
			ServiceFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			AppServiceFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			ConverterFn: func() *automock.WebhookConverter {
				converter := &automock.WebhookConverter{}
				converter.Mock.On("InputFromGraphQL", gqlWebhookInput).Return(nil, testErr)
				return converter
			},
			ExpectedErr: errors.New("while converting the WebhookInput: Test error"),
		},
		{
			Name: "Returns error on committing transaction",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(givenError()).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), givenAppID, *modelWebhookInput, model.ApplicationWebhookReference).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id, model.ApplicationWebhookReference).Return(modelWebhook, nil).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), givenAppID).Return(true, nil).Once()
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedErr: givenError(),
		},
		{
			Name:            "Returns error when application does not exist",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), givenAppID).Return(false, nil)
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     errors.New("cannot add ApplicationWebhook due to not existing reference entity"),
		},
		{
			Name:            "Returns error when application template does not exist",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Exists", txtest.CtxWithDBMatcher(), givenAppTemplateID).Return(false, nil)
				return appTemplateSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     errors.New("cannot add ApplicationTemplateWebhook due to not existing reference entity"),
		},
		{
			Name:            "Returns error when runtime does not exist",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				return svc
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				runtimeSvc := &automock.RuntimeService{}
				runtimeSvc.On("Exist", txtest.CtxWithDBMatcher(), givenRuntimeID).Return(false, nil)
				return runtimeSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     errors.New("cannot add RuntimeWebhook due to not existing reference entity"),
		},
		{
			Name:            "Returns error when formation template does not exist",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				return svc
			},
			FormationTemplateServiceFn: func() *automock.FormationTemplateService {
				ftSvc := &automock.FormationTemplateService{}
				ftSvc.On("Exist", txtest.CtxWithDBMatcher(), givenFormationTemplateID).Return(false, nil)
				return ftSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     errors.New("cannot add FormationTemplateWebhook due to not existing reference entity"),
		},
		{
			Name:            "Returns error when application existence check failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), givenAppID).Return(false, testErr).Once()
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when webhook creation failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), givenAppID, *modelWebhookInput, model.ApplicationWebhookReference).Return("", testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), givenAppID).Return(true, nil).Once()
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when webhook retrieval failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), givenAppID, *modelWebhookInput, model.ApplicationWebhookReference).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id, model.ApplicationWebhookReference).Return(nil, testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), givenAppID).Return(true, nil).Once()
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var appSvc *automock.ApplicationService
			var appTemplateSvc *automock.ApplicationTemplateService
			var runtimeSvc *automock.RuntimeService
			var formationTemplateSvc *automock.FormationTemplateService
			svc := testCase.ServiceFn()
			if testCase.AppServiceFn != nil {
				appSvc = testCase.AppServiceFn()
			}
			if testCase.AppTemplateServiceFn != nil {
				appTemplateSvc = testCase.AppTemplateServiceFn()
			}
			if testCase.RuntimeServiceFn != nil {
				runtimeSvc = testCase.RuntimeServiceFn()
			}
			if testCase.FormationTemplateServiceFn != nil {
				formationTemplateSvc = testCase.FormationTemplateServiceFn()
			}

			converter := testCase.ConverterFn()

			persistTxMock := testCase.PersistenceFn()
			transactionerMock := testCase.TransactionerFn(persistTxMock)

			resolver := webhook.NewResolver(transactionerMock, svc, appSvc, appTemplateSvc, runtimeSvc, formationTemplateSvc, converter)

			// WHEN
			var err error
			var result *graphql.Webhook
			if testCase.AppServiceFn != nil {
				result, err = resolver.AddWebhook(context.TODO(), stringPtr(givenAppID), nil, nil, nil, *gqlWebhookInput)
			}
			if testCase.AppTemplateServiceFn != nil {
				result, err = resolver.AddWebhook(context.TODO(), nil, stringPtr(givenAppTemplateID), nil, nil, *gqlWebhookInput)
			}
			if testCase.RuntimeServiceFn != nil {
				result, err = resolver.AddWebhook(context.TODO(), nil, nil, stringPtr(givenRuntimeID), nil, *gqlWebhookInput)
			}
			if testCase.FormationTemplateServiceFn != nil {
				result, err = resolver.AddWebhook(context.TODO(), nil, nil, nil, stringPtr(givenFormationTemplateID), *gqlWebhookInput)
			}

			// THEN
			assert.Equal(t, testCase.ExpectedWebhook, result)
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			svc.AssertExpectations(t)
			if testCase.AppServiceFn != nil {
				appSvc.AssertExpectations(t)
			}
			if testCase.AppTemplateServiceFn != nil {
				appTemplateSvc.AssertExpectations(t)
			}
			if testCase.RuntimeServiceFn != nil {
				runtimeSvc.AssertExpectations(t)
			}
			if testCase.FormationTemplateServiceFn != nil {
				formationTemplateSvc.AssertExpectations(t)
			}
			converter.AssertExpectations(t)
			persistTxMock.AssertExpectations(t)
			transactionerMock.AssertExpectations(t)
		})
		t.Run("Error when more than one of application, application_template, runtime, formation_template is specified", func(t *testing.T) {
			persistTxMock := txtest.PersistenceContextThatDoesntExpectCommit()
			transactionerMock := txtest.TransactionerThatSucceeds(persistTxMock)

			resolver := webhook.NewResolver(transactionerMock, nil, nil, nil, nil, nil, nil)
			_, err := resolver.AddWebhook(context.TODO(), stringPtr("app"), stringPtr("app_template"), stringPtr("runtime"), stringPtr("formation_template"), graphql.WebhookInput{})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "exactly one of applicationID, applicationTemplateID, runtimeID or formationTemplateID should be specified")
		})
		t.Run("Error when none of application, application_template, runtime, formation_template is specified", func(t *testing.T) {
			persistTxMock := txtest.PersistenceContextThatDoesntExpectCommit()
			transactionerMock := txtest.TransactionerThatSucceeds(persistTxMock)

			resolver := webhook.NewResolver(transactionerMock, nil, nil, nil, nil, nil, nil)
			_, err := resolver.AddWebhook(context.TODO(), nil, nil, nil, nil, graphql.WebhookInput{})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "exactly one of applicationID, applicationTemplateID, runtimeID or formationTemplateID should be specified")
		})
	}
}

func TestResolver_UpdateWebhook(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	applicationID := "foo"
	givenWebhookID := "bar"
	gqlWebhookInput := fixGQLWebhookInput("foo")
	modelWebhookInput := fixModelWebhookInput("foo")
	gqlWebhook := fixGQLWebhook(givenWebhookID, "", "")
	modelWebhook := fixApplicationModelWebhook(givenWebhookID, applicationID, givenTenant(), "foo", time.Time{})

	testCases := []struct {
		Name            string
		ServiceFn       func() *automock.WebhookService
		ConverterFn     func() *automock.WebhookConverter
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ExpectedWebhook *graphql.Webhook
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), givenWebhookID, *modelWebhookInput, model.UnknownWebhookReference).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), givenWebhookID, model.UnknownWebhookReference).Return(modelWebhook, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook, nil).Once()
				return conv
			},
			ExpectedWebhook: gqlWebhook,
			ExpectedErr:     nil,
		},
		{
			Name:          "Returns error on starting transaction",
			PersistenceFn: txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, givenError()).Once()
				return transact
			},
			ServiceFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedErr: givenError(),
		},
		{
			Name: "Returns error on committing transaction",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(givenError()).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), givenWebhookID, *modelWebhookInput, model.UnknownWebhookReference).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), givenWebhookID, model.UnknownWebhookReference).Return(modelWebhook, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedErr: givenError(),
		},
		{
			Name:            "Returns error when webhook conversion failed",
			TransactionerFn: txtest.TransactionerThatSucceeds,
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			ServiceFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(nil, testErr).Once()
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     errors.New("while converting the WebhookInput"),
		},
		{
			Name:            "Returns error when webhook update failed",
			TransactionerFn: txtest.TransactionerThatSucceeds,
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), givenWebhookID, *modelWebhookInput, model.UnknownWebhookReference).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when webhook retrieval failed",
			TransactionerFn: txtest.TransactionerThatSucceeds,
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), givenWebhookID, *modelWebhookInput, model.UnknownWebhookReference).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), givenWebhookID, model.UnknownWebhookReference).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			persistTxMock := testCase.PersistenceFn()
			transactionerMock := testCase.TransactionerFn(persistTxMock)

			resolver := webhook.NewResolver(transactionerMock, svc, nil, nil, nil, nil, converter)

			// WHEN
			result, err := resolver.UpdateWebhook(context.TODO(), givenWebhookID, *gqlWebhookInput)

			// THEN
			assert.Equal(t, testCase.ExpectedWebhook, result)
			if testCase.ExpectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transactionerMock.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteWebhook(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	applicationID := "foo"
	givenWebhookID := "bar"

	gqlWebhook := fixGQLWebhook(givenWebhookID, "", "")
	modelWebhook := fixApplicationModelWebhook(givenWebhookID, applicationID, givenTenant(), "foo", time.Time{})

	testCases := []struct {
		Name            string
		ServiceFn       func() *automock.WebhookService
		ConverterFn     func() *automock.WebhookConverter
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ExpectedWebhook *graphql.Webhook
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txtest.TransactionerThatSucceeds,
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), givenWebhookID, model.UnknownWebhookReference).Return(modelWebhook, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), givenWebhookID, model.UnknownWebhookReference).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook, nil).Once()
				return conv
			},
			ExpectedWebhook: gqlWebhook,
			ExpectedErr:     nil,
		},
		{
			Name:          "Returns error on starting transaction",
			PersistenceFn: txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, givenError()).Once()
				return transact
			},
			ServiceFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},

			ConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedErr: givenError(),
		},
		{
			Name: "Returns error on committing transaction",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(givenError()).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), givenWebhookID, model.UnknownWebhookReference).Return(modelWebhook, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), givenWebhookID, model.UnknownWebhookReference).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook, nil).Once()
				return conv
			},
			ExpectedErr: givenError(),
		},
		{
			Name:            "Returns error when webhook retrieval failed",
			TransactionerFn: txtest.TransactionerThatSucceeds,
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), givenWebhookID, model.UnknownWebhookReference).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when webhook deletion failed",
			TransactionerFn: txtest.TransactionerThatSucceeds,
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), givenWebhookID, model.UnknownWebhookReference).Return(modelWebhook, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), givenWebhookID, model.UnknownWebhookReference).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook, nil).Once()
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			persistTxMock := testCase.PersistenceFn()
			transactionerMock := testCase.TransactionerFn(persistTxMock)

			resolver := webhook.NewResolver(transactionerMock, svc, nil, nil, nil, nil, converter)

			// WHEN
			result, err := resolver.DeleteWebhook(context.TODO(), givenWebhookID)

			// THEN
			assert.Equal(t, testCase.ExpectedWebhook, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transactionerMock.AssertExpectations(t)
		})
	}
}
