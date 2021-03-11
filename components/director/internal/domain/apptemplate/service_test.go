package apptemplate_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testID)
		return uidSvc
	}

	modelAppTemplate := fixModelApplicationTemplate(testID, testName, []*model.Webhook{})

	appTemplateInputWithWebhooks := fixModelAppTemplateInput(testName, appInputJSONString)
	appTemplateInputWithWebhooks.Webhooks = []*model.WebhookInput{
		{
			Type: model.WebhookTypeConfigurationChanged,
			URL:  str.Ptr("foourl"),
			Auth: &model.AuthInput{},
		},
	}
	appTemplateInputMatcher := func(webhooks []*model.Webhook) bool {
		return webhooks != nil && len(webhooks) == 1 && *webhooks[0].ApplicationTemplateID == testID && webhooks[0].Type == model.WebhookTypeConfigurationChanged && *webhooks[0].URL == "foourl"
	}

	testCases := []struct {
		Name              string
		Input             *model.ApplicationTemplateInput
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		WebhookRepoFn     func() *automock.WebhookRepository
		ExpectedError     error
		ExpectedOutput    string
	}{
		{
			Name:  "Success",
			Input: fixModelAppTemplateInput(testName, appInputJSONString),
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Create", ctx, *modelAppTemplate).Return(nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("CreateMany", ctx, []*model.Webhook(nil)).Return(nil).Once()
				return webhookRepo
			},
			ExpectedOutput: testID,
		},
		{
			Name:  "Success for Applicaiton Template with webhooks",
			Input: appTemplateInputWithWebhooks,
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Create", ctx, mock.AnythingOfType("model.ApplicationTemplate")).Return(nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("CreateMany", ctx, mock.MatchedBy(appTemplateInputMatcher)).Return(nil).Once()
				return webhookRepo
			},
			ExpectedOutput: testID,
		},
		{
			Name:  "Error when creating application template",
			Input: fixModelAppTemplateInput(testName, appInputJSONString),
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Create", ctx, *modelAppTemplate).Return(testError).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
			ExpectedError:  testError,
			ExpectedOutput: "",
		},
		{
			Name:  "Error when creating webhooks",
			Input: fixModelAppTemplateInput(testName, appInputJSONString),
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Create", ctx, *modelAppTemplate).Return(nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("CreateMany", ctx, mock.AnythingOfType("[]*model.Webhook")).Return(testError).Once()
				return webhookRepo
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			idSvc := uidSvcFn()
			svc := apptemplate.NewService(appTemplateRepo, webhookRepo, idSvc)

			// WHEN
			result, err := svc.Create(ctx, *testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			appTemplateRepo.AssertExpectations(t)
			idSvc.AssertExpectations(t)
		})
	}
}

func TestService_Get(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	modelAppTemplate := fixModelApplicationTemplate(testID, testName, fixModelApplicationTemplateWebhooks(testWebhookID, testID))

	testCases := []struct {
		Name              string
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		WebhookRepoFn     func() *automock.WebhookRepository
		ExpectedError     error
		ExpectedOutput    *model.ApplicationTemplate
	}{
		{
			Name: "Success",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Get", ctx, testID).Return(modelAppTemplate, nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
			ExpectedOutput: modelAppTemplate,
		},
		{
			Name: "Error when getting application template",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Get", ctx, testID).Return(nil, testError).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, webhookRepo, nil)

			// WHEN
			result, err := svc.Get(ctx, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			appTemplateRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetByName(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	modelAppTemplate := fixModelApplicationTemplate(testID, testName, fixModelApplicationTemplateWebhooks(testWebhookID, testID))

	testCases := []struct {
		Name              string
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		WebhookRepoFn     func() *automock.WebhookRepository
		ExpectedError     error
		ExpectedOutput    *model.ApplicationTemplate
	}{
		{
			Name: "Success",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return(modelAppTemplate, nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
			ExpectedOutput: modelAppTemplate,
		},
		{
			Name: "Error when getting application template",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return(nil, testError).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, webhookRepo, nil)

			// WHEN
			result, err := svc.GetByName(ctx, testName)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			appTemplateRepo.AssertExpectations(t)
		})
	}
}

func TestService_Exists(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	testCases := []struct {
		Name              string
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		WebhookRepoFn     func() *automock.WebhookRepository
		ExpectedError     error
		ExpectedOutput    bool
	}{
		{
			Name: "Success",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", ctx, testID).Return(true, nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
			ExpectedOutput: true,
		},
		{
			Name: "Error when getting application template",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", ctx, testID).Return(false, testError).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
			ExpectedError:  testError,
			ExpectedOutput: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, webhookRepo, nil)

			// WHEN
			result, err := svc.Exists(ctx, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			appTemplateRepo.AssertExpectations(t)
		})
	}
}

func TestService_List(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)
	modelAppTemplate := fixModelAppTemplatePage([]*model.ApplicationTemplate{
		fixModelApplicationTemplate("foo1", "bar1", fixModelApplicationTemplateWebhooks("webhook-id-1", "foo1")),
		fixModelApplicationTemplate("foo2", "bar2", fixModelApplicationTemplateWebhooks("webhook-id-2", "foo2")),
	})

	testCases := []struct {
		Name              string
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		WebhookRepoFn     func() *automock.WebhookRepository
		InputPageSize     int
		ExpectedError     error
		ExpectedOutput    model.ApplicationTemplatePage
	}{
		{
			Name: "Success",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("List", ctx, 50, testCursor).Return(modelAppTemplate, nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
			InputPageSize:  50,
			ExpectedOutput: modelAppTemplate,
		},
		{
			Name: "Error when listing application template",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("List", ctx, 50, testCursor).Return(model.ApplicationTemplatePage{}, testError).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
			InputPageSize:  50,
			ExpectedError:  testError,
			ExpectedOutput: model.ApplicationTemplatePage{},
		},
		{
			Name: "Error when page size too small",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
			InputPageSize:  0,
			ExpectedError:  errors.New("page size must be between 1 and 200"),
			ExpectedOutput: model.ApplicationTemplatePage{},
		},
		{
			Name: "Error when page size too big",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
			InputPageSize:  201,
			ExpectedError:  errors.New("page size must be between 1 and 200"),
			ExpectedOutput: model.ApplicationTemplatePage{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, webhookRepo, nil)

			// WHEN
			result, err := svc.List(ctx, testCase.InputPageSize, testCursor)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			appTemplateRepo.AssertExpectations(t)
		})
	}
}

func TestService_Update(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)
	modelAppTemplate := fixModelApplicationTemplate(testID, testName, []*model.Webhook{})

	testCases := []struct {
		Name              string
		Input             *model.ApplicationTemplateInput
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		WebhookRepoFn     func() *automock.WebhookRepository
		ExpectedError     error
	}{
		{
			Name:  "Success",
			Input: fixModelAppTemplateInput(testName, appInputJSONString),
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Update", ctx, *modelAppTemplate).Return(nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
		},
		{
			Name:  "Error when updating application template",
			Input: fixModelAppTemplateInput(testName, appInputJSONString),
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Update", ctx, *modelAppTemplate).Return(testError).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, webhookRepo, nil)

			// WHEN
			err := svc.Update(ctx, testID, *testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			appTemplateRepo.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	testCases := []struct {
		Name              string
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		WebhookRepoFn     func() *automock.WebhookRepository
		ExpectedError     error
	}{
		{
			Name: "Success",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Delete", ctx, testID).Return(nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
		},
		{
			Name: "Error when deleting application template",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Delete", ctx, testID).Return(testError).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, webhookRepo, nil)

			// WHEN
			err := svc.Delete(ctx, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			appTemplateRepo.AssertExpectations(t)
		})
	}
}

func TestService_PrepareApplicationCreateInputJSON(t *testing.T) {
	// GIVEN
	svc := apptemplate.NewService(nil, nil, nil)

	testCases := []struct {
		Name             string
		InputAppTemplate *model.ApplicationTemplate
		InputValues      model.ApplicationFromTemplateInputValues
		ExpectedOutput   string
		ExpectedError    error
	}{
		{
			Name: "Success when no placeholders",
			InputAppTemplate: &model.ApplicationTemplate{
				ApplicationInputJSON: `{"Name": "my-app", "Description": "Lorem ipsum"}`,
				Placeholders:         nil,
			},
			InputValues:    nil,
			ExpectedOutput: `{"Name": "my-app", "Description": "Lorem ipsum"}`,
			ExpectedError:  nil,
		},
		{
			Name: "Success when with placeholders",
			InputAppTemplate: &model.ApplicationTemplate{
				ApplicationInputJSON: `{"Name": "{{name}}", "Description": "Lorem ipsum"}`,
				Placeholders: []model.ApplicationTemplatePlaceholder{
					{Name: "name", Description: str.Ptr("Application name")},
				},
			},
			InputValues: []*model.ApplicationTemplateValueInput{
				{Placeholder: "name", Value: "my-application"},
			},
			ExpectedOutput: `{"Name": "my-application", "Description": "Lorem ipsum"}`,
			ExpectedError:  nil,
		},
		{
			Name: "Returns error when required placeholder value not provided",
			InputAppTemplate: &model.ApplicationTemplate{
				ApplicationInputJSON: `{"Name": "{{name}}", "Description": "Lorem ipsum"}`,
				Placeholders: []model.ApplicationTemplatePlaceholder{
					{Name: "name", Description: str.Ptr("Application name")},
				},
			},
			InputValues:    []*model.ApplicationTemplateValueInput{},
			ExpectedOutput: "",
			ExpectedError:  errors.New("required placeholder not provided: value for placeholder name 'name' not found"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			result, err := svc.PrepareApplicationCreateInputJSON(testCase.InputAppTemplate, testCase.InputValues)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Empty(t, result)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, result)
			}
		})
	}
}
