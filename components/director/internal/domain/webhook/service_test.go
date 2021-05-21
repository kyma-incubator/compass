package webhook_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testCaseErrorOnLoadingTenant = "Returns error on loading tenant"

func TestService_Create(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	modelInput := fixModelWebhookInput("foo")

	webhookModel := mock.MatchedBy(func(webhook *model.Webhook) bool {
		return webhook.Type == modelInput.Type && webhook.URL == modelInput.URL
	})

	applicationConvererFunc := (*model.WebhookInput).ToApplicationTemplateWebhook

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, givenTenant(), givenExternalTenant())
	ctxNoTenant := context.TODO()
	ctxNoTenant = tenant.SaveToContext(ctx, "", "")

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.WebhookRepository
		UIDServiceFn func() *automock.UIDService
		ExpectedErr  error
		Context      context.Context
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("Create", ctx, webhookModel).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return("foo").Once()
				return svc
			},
			ExpectedErr: nil,
			Context:     ctx,
		},
		{
			Name: "Success when tenant is missing",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("Create", ctxNoTenant, webhookModel).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return("foo").Once()
				return svc
			},
			ExpectedErr: nil,
			Context:     ctxNoTenant,
		},
		{
			Name: "Returns error when webhook creation failed",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("Create", ctx, webhookModel).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return("").Once()
				return svc
			},
			ExpectedErr: testErr,
			Context:     ctx,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			uidSvc := testCase.UIDServiceFn()

			svc := webhook.NewService(repo, nil, uidSvc)

			// when
			result, err := svc.Create(testCase.Context, givenApplicationID(), *modelInput, applicationConvererFunc)

			// then

			if testCase.ExpectedErr == nil {
				assert.NotEmpty(t, result)
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}

	t.Run(testCaseErrorOnLoadingTenant, func(t *testing.T) {
		svc := webhook.NewService(nil, nil, nil)
		// when
		_, err := svc.Create(context.TODO(), givenApplicationID(), *modelInput, applicationConvererFunc)
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
}

func TestService_Get(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	url := "bar"

	webhookModel := fixApplicationModelWebhook("1", id, givenTenant(), url)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, givenTenant(), givenExternalTenant())

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.WebhookRepository
		ExpectedWebhook    *model.Webhook
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", ctx, givenTenant(), id).Return(webhookModel, nil).Once()
				return repo
			},
			ExpectedWebhook:    webhookModel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when webhook retrieval failed",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", ctx, givenTenant(), id).Return(nil, testErr).Once()
				return repo
			},
			ExpectedWebhook:    webhookModel,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := webhook.NewService(repo, nil, nil)

			// when
			actual, err := svc.Get(ctx, id)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedWebhook, actual)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_ListForApplication(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	modelWebhooks := []*model.Webhook{
		fixApplicationModelWebhook("1", "foo", givenTenant(), "Foo"),
		fixApplicationModelWebhook("2", "bar", givenTenant(), "Bar"),
	}
	applicationID := "foo"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, givenTenant(), givenExternalTenant())

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.WebhookRepository
		ExpectedResult     []*model.Webhook
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByApplicationID", ctx, givenTenant(), applicationID).Return(modelWebhooks, nil).Once()
				return repo
			},
			ExpectedResult:     modelWebhooks,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when webhook listing failed",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByApplicationID", ctx, givenTenant(), applicationID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := webhook.NewService(repo, nil, nil)

			// when
			webhooks, err := svc.ListForApplication(ctx, applicationID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, webhooks)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run(testCaseErrorOnLoadingTenant, func(t *testing.T) {
		svc := webhook.NewService(nil, nil, nil)
		// when
		_, err := svc.ListForApplication(context.TODO(), givenApplicationID())
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
}

func TestService_ListForApplicationTemplate(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	modelWebhooks := []*model.Webhook{
		fixApplicationTemplateModelWebhook("1", "foo", "Foo"),
		fixApplicationTemplateModelWebhook("2", "bar", "Bar"),
	}
	applicationTemplateID := "foo"

	ctx := context.TODO()

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.WebhookRepository
		ExpectedResult     []*model.Webhook
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByApplicationTemplateID", ctx, applicationTemplateID).Return(modelWebhooks, nil).Once()
				return repo
			},
			ExpectedResult:     modelWebhooks,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when webhook listing failed",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByApplicationTemplateID", ctx, applicationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := webhook.NewService(repo, nil, nil)

			// when
			webhooks, err := svc.ListForApplicationTemplate(ctx, applicationTemplateID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, webhooks)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_ListAllApplicationWebhooks(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	application := &model.Application{
		Tenant: givenTenant(),
		Name:   "app-1",
		BaseEntity: &model.BaseEntity{
			ID: "app-1-id",
		},
	}
	applicationFromTemplate := &model.Application{
		Tenant:                givenTenant(),
		Name:                  "app-1",
		ApplicationTemplateID: str.Ptr("app-template-id-1"),
		BaseEntity: &model.BaseEntity{
			ID: "app-1-id",
		},
	}

	appModelWebhooks := []*model.Webhook{
		fixApplicationModelWebhookWithType("app-webhook-1", "app-1", givenTenant(), "test-url-1.com", model.WebhookTypeRegisterApplication),
		fixApplicationModelWebhookWithType("app-webhook-2", "app-1", givenTenant(), "test-url-2.com", model.WebhookTypeDeleteApplication),
	}
	appTemplateModelWebhooks := []*model.Webhook{
		fixApplicationTemplateModelWebhookWithType("app-template-webhook-1", "app-template-1", "test-url-1.com", model.WebhookTypeRegisterApplication),
		fixApplicationTemplateModelWebhookWithType("app-template-webhook-2", "app-template-2", "test-url-2.com", model.WebhookTypeDeleteApplication),
	}
	mergedWebhooks := []*model.Webhook{
		appTemplateModelWebhooks[1],
		appModelWebhooks[0],
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, givenTenant(), givenExternalTenant())

	testCases := []struct {
		Name                    string
		WebhookRepositoryFn     func() *automock.WebhookRepository
		ApplicationRepositoryFn func() *automock.ApplicationRepository
		ExpectedResult          []*model.Webhook
		ExpectedErrMessage      string
	}{
		{
			Name: "Success when only application has webhooks",
			ApplicationRepositoryFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", ctx, application.ID).Return(application, nil).Once()
				return appRepo
			},
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByApplicationID", ctx, givenTenant(), application.ID).Return(appModelWebhooks, nil).Once()
				return webhookRepo
			},
			ExpectedResult:     appModelWebhooks,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when only application template has webhooks",
			ApplicationRepositoryFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", ctx, applicationFromTemplate.ID).Return(applicationFromTemplate, nil).Once()
				return appRepo
			},
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByApplicationID", ctx, givenTenant(), applicationFromTemplate.ID).Return(nil, nil).Once()
				webhookRepo.On("ListByApplicationTemplateID", ctx, *applicationFromTemplate.ApplicationTemplateID).Return(appTemplateModelWebhooks, nil).Once()
				return webhookRepo
			},
			ExpectedResult:     appTemplateModelWebhooks,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when application template webhooks have to be overwritten",
			ApplicationRepositoryFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", ctx, applicationFromTemplate.ID).Return(applicationFromTemplate, nil).Once()
				return appRepo
			},
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByApplicationID", ctx, givenTenant(), applicationFromTemplate.ID).Return(appModelWebhooks, nil).Once()
				webhookRepo.On("ListByApplicationTemplateID", ctx, *applicationFromTemplate.ApplicationTemplateID).Return(appTemplateModelWebhooks, nil).Once()
				return webhookRepo
			},
			ExpectedResult:     appModelWebhooks,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when webhooks have to be merged from both app and template",
			ApplicationRepositoryFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", ctx, applicationFromTemplate.ID).Return(applicationFromTemplate, nil).Once()
				return appRepo
			},
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByApplicationID", ctx, givenTenant(), applicationFromTemplate.ID).Return([]*model.Webhook{appModelWebhooks[0]}, nil).Once()
				webhookRepo.On("ListByApplicationTemplateID", ctx, *applicationFromTemplate.ApplicationTemplateID).Return([]*model.Webhook{appTemplateModelWebhooks[1]}, nil).Once()
				return webhookRepo
			},
			ExpectedResult:     mergedWebhooks,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when webhook listing for application failed",
			ApplicationRepositoryFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", ctx, application.ID).Return(application, nil).Once()
				return appRepo
			},
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByApplicationID", ctx, givenTenant(), application.ID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when webhook listing for application template failed",
			ApplicationRepositoryFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", ctx, applicationFromTemplate.ID).Return(applicationFromTemplate, nil).Once()
				return appRepo
			},
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByApplicationID", ctx, givenTenant(), applicationFromTemplate.ID).Return(appModelWebhooks, nil).Once()
				webhookRepo.On("ListByApplicationTemplateID", ctx, *applicationFromTemplate.ApplicationTemplateID).Return(nil, testErr).Once()

				return webhookRepo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application repository returns an error",
			ApplicationRepositoryFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", ctx, applicationFromTemplate.ID).Return(nil, testErr).Once()
				return appRepo
			},
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			webhookRepo := testCase.WebhookRepositoryFn()
			applicationRepo := testCase.ApplicationRepositoryFn()
			svc := webhook.NewService(webhookRepo, applicationRepo, nil)

			// when
			webhooks, err := svc.ListAllApplicationWebhooks(ctx, application.ID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				for _, expected := range testCase.ExpectedResult {
					assert.Contains(t, webhooks, expected)
				}
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			webhookRepo.AssertExpectations(t)
		})
	}

	t.Run(testCaseErrorOnLoadingTenant, func(t *testing.T) {
		svc := webhook.NewService(nil, nil, nil)
		// when
		_, err := svc.ListForApplication(context.TODO(), givenApplicationID())
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	url := "foo"
	id := "bar"
	modelInput := fixModelWebhookInput(url)

	inputWebhookModel := mock.MatchedBy(func(webhook *model.Webhook) bool {
		return webhook.URL == modelInput.URL
	})

	applicationWebhookModel := fixApplicationModelWebhook("1", id, givenTenant(), url)
	applicationTemplateWebhookModel := fixApplicationTemplateModelWebhook("1", id, url)
	noIDWebhookModel := &model.Webhook{}
	*noIDWebhookModel = *applicationWebhookModel
	noIDWebhookModel.ApplicationID = nil

	tenantCtx := context.TODO()
	tenantCtx = tenant.SaveToContext(tenantCtx, givenTenant(), givenExternalTenant())

	testCases := []struct {
		Name                string
		WebhookRepositoryFn func() *automock.WebhookRepository
		ExpectedErrMessage  string
		Context             context.Context
	}{
		{
			Name: "Success when applicationID is present",
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", tenantCtx, givenTenant(), id).Return(applicationWebhookModel, nil).Once()
				repo.On("Update", tenantCtx, inputWebhookModel).Return(nil).Once()
				return repo
			},
			ExpectedErrMessage: "",
			Context:            tenantCtx,
		},
		{
			Name: "Success when applicationTemplateID is present",
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				ctx := context.TODO()
				repo.On("GetByIDGlobal", ctx, id).Return(applicationTemplateWebhookModel, nil).Once()
				repo.On("Update", ctx, inputWebhookModel).Return(nil).Once()
				return repo
			},
			ExpectedErrMessage: "",
			Context:            context.TODO(),
		},
		{
			Name: "Returns error when webhook update failed",
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", tenantCtx, givenTenant(), id).Return(applicationWebhookModel, nil).Once()
				repo.On("Update", tenantCtx, inputWebhookModel).Return(testErr).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
			Context:            tenantCtx,
		},
		{
			Name: "Returns error when webhook retrieval failed",
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", tenantCtx, givenTenant(), id).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
			Context:            tenantCtx,
		},
		{
			Name: "Returns error application doesn't have any ids",
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", tenantCtx, givenTenant(), id).Return(noIDWebhookModel, nil).Once()
				return repo
			},
			ExpectedErrMessage: "webhook doesn't have neither of application_id and application_template_id",
			Context:            tenantCtx,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.WebhookRepositoryFn()
			svc := webhook.NewService(repo, nil, nil)

			// when
			err := svc.Update(testCase.Context, id, *modelInput)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

}

func TestService_Delete(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	url := "bar"

	webhookModel := fixApplicationModelWebhook("1", id, givenTenant(), url)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, givenTenant(), givenExternalTenant())

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.WebhookRepository
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", ctx, givenTenant(), id).Return(webhookModel, nil).Once()
				repo.On("Delete", ctx, webhookModel.ID).Return(nil).Once()
				return repo
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when webhook deletion failed",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", ctx, givenTenant(), id).Return(webhookModel, nil).Once()
				repo.On("Delete", ctx, webhookModel.ID).Return(testErr).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when webhook retrieval failed",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", ctx, givenTenant(), id).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := webhook.NewService(repo, nil, nil)

			// when
			err := svc.Delete(ctx, id)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

}
