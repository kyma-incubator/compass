package apptemplate_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"testing"

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

	predefinedID := "123-465-789"

	appInputJSON := fmt.Sprintf(appInputJSONWithAppTypeLabelString, testName)
	modelAppTemplate := fixModelAppTemplateWithAppInputJSON(testID, testName, appInputJSON, []*model.Webhook{})

	appTemplateInputWithWebhooks := fixModelAppTemplateInput(testName, appInputJSONString)
	appTemplateInputWithWebhooks.Webhooks = []*model.WebhookInput{
		{
			Type: model.WebhookTypeConfigurationChanged,
			URL:  str.Ptr("foourl"),
			Auth: &model.AuthInput{},
		},
	}
	appTemplateInputMatcher := func(webhooks []*model.Webhook) bool {
		return len(webhooks) == 1 && webhooks[0].ObjectID == testID && webhooks[0].Type == model.WebhookTypeConfigurationChanged && *webhooks[0].URL == "foourl"
	}

	testCases := []struct {
		Name              string
		Input             func() *model.ApplicationTemplateInput
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		WebhookRepoFn     func() *automock.WebhookRepository
		LabelUpsertSvcFn  func() *automock.LabelUpsertService
		LabelRepoFn       func() *automock.LabelRepository
		ExpectedError     error
		ExpectedOutput    string
	}{
		{
			Name: "Success",
			Input: func() *model.ApplicationTemplateInput {
				return fixModelAppTemplateInput(testName, appInputJSON)
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return([]*model.ApplicationTemplate{}, nil).Once()
				appTemplateRepo.On("Create", ctx, *modelAppTemplate).Return(nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("CreateMany", ctx, "", []*model.Webhook{}).Return(nil).Once()
				return webhookRepo
			},
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				labelUpsertService := &automock.LabelUpsertService{}
				labelUpsertService.On("UpsertMultipleLabels", ctx, "", model.AppTemplateLabelableObject, "foo", map[string]interface{}{"test": "test"}).Return(nil).Once()
				return labelUpsertService
			},
			LabelRepoFn:    UnusedLabelRepo,
			ExpectedOutput: testID,
		},
		{
			Name: "Success without app input labels",
			Input: func() *model.ApplicationTemplateInput {
				appInputJSON := `{"name":"foo","providerName":"compass","description":"Lorem ipsum","healthCheckURL":"https://foo.bar","webhooks":[{"type":"","url":"webhook1.foo.bar","auth":null},{"type":"","url":"webhook2.foo.bar","auth":null}],"integrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`
				return fixModelAppTemplateInput(testName, appInputJSON)
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return([]*model.ApplicationTemplate{}, nil).Once()
				appTemplateRepo.On("Create", ctx, mock.AnythingOfType("model.ApplicationTemplate")).Return(nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("CreateMany", ctx, "", []*model.Webhook{}).Return(nil).Once()
				return webhookRepo
			},
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				labelUpsertService := &automock.LabelUpsertService{}
				labelUpsertService.On("UpsertMultipleLabels", ctx, "", model.AppTemplateLabelableObject, "foo", map[string]interface{}{"test": "test"}).Return(nil).Once()
				return labelUpsertService
			},
			LabelRepoFn:    UnusedLabelRepo,
			ExpectedOutput: testID,
		},
		{
			Name: "Success when application template has region label",
			Input: func() *model.ApplicationTemplateInput {
				appTemplateInput := fixModelAppTemplateInput(testName, fmt.Sprintf(appInputJSONWithAppTypeLabelString, testName+" (eu-1)"))
				appTemplateInput.Labels["region"] = "eu-1"
				return appTemplateInput
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				modelAppTemplate := fixModelAppTemplateWithAppInputJSON(testID, testName, fmt.Sprintf(appInputJSONWithAppTypeLabelString, testName+" (eu-1)"), []*model.Webhook{})
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return([]*model.ApplicationTemplate{}, nil).Once()
				appTemplateRepo.On("Create", ctx, *modelAppTemplate).Return(nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("CreateMany", ctx, "", []*model.Webhook{}).Return(nil).Once()
				return webhookRepo
			},
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				labelUpsertService := &automock.LabelUpsertService{}
				labelUpsertService.On("UpsertMultipleLabels", ctx, "", model.AppTemplateLabelableObject, "foo", map[string]interface{}{"region": "eu-1", "test": "test"}).Return(nil).Once()
				return labelUpsertService
			},
			LabelRepoFn:    UnusedLabelRepo,
			ExpectedOutput: testID,
		},
		{
			Name: "Success when ID is already generated",
			Input: func() *model.ApplicationTemplateInput {
				return fixModelAppTemplateWithIDInput(testName, appInputJSON, &predefinedID)
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				modelAppTemplateWithPredefinedID := *modelAppTemplate
				modelAppTemplateWithPredefinedID.ID = predefinedID
				appTemplateRepo.On("GetByName", ctx, testName).Return([]*model.ApplicationTemplate{}, nil).Once()
				appTemplateRepo.On("Create", ctx, modelAppTemplateWithPredefinedID).Return(nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("CreateMany", ctx, "", []*model.Webhook{}).Return(nil).Once()
				return webhookRepo
			},
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				labelUpsertService := &automock.LabelUpsertService{}
				labelUpsertService.On("UpsertMultipleLabels", ctx, "", model.AppTemplateLabelableObject, predefinedID, map[string]interface{}{"test": "test"}).Return(nil).Once()
				return labelUpsertService
			},
			LabelRepoFn:    UnusedLabelRepo,
			ExpectedOutput: predefinedID,
		},
		{
			Name: "Success for Application Template with webhooks",
			Input: func() *model.ApplicationTemplateInput {
				return appTemplateInputWithWebhooks
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return([]*model.ApplicationTemplate{}, nil).Once()
				appTemplateRepo.On("Create", ctx, mock.AnythingOfType("model.ApplicationTemplate")).Return(nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("CreateMany", ctx, "", mock.MatchedBy(appTemplateInputMatcher)).Return(nil).Once()
				return webhookRepo
			},
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				labelUpsertService := &automock.LabelUpsertService{}
				labelUpsertService.On("UpsertMultipleLabels", ctx, "", model.AppTemplateLabelableObject, "foo", map[string]interface{}{"test": "test"}).Return(nil).Once()
				return labelUpsertService
			},
			LabelRepoFn:    UnusedLabelRepo,
			ExpectedOutput: testID,
		},
		{
			Name: "Error when creating application template",
			Input: func() *model.ApplicationTemplateInput {
				return fixModelAppTemplateInput(testName, appInputJSON)
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return([]*model.ApplicationTemplate{}, nil).Once()
				appTemplateRepo.On("Create", ctx, *modelAppTemplate).Return(testError).Once()
				return appTemplateRepo
			},
			WebhookRepoFn:    UnusedWebhookRepo,
			LabelUpsertSvcFn: UnusedLabelUpsertSvc,
			LabelRepoFn:      UnusedLabelRepo,
			ExpectedError:    testError,
			ExpectedOutput:   "",
		},
		{
			Name: "Error when creating application type label - region is not string",
			Input: func() *model.ApplicationTemplateInput {
				appTemplateInput := fixModelAppTemplateInput(testName, appInputJSONString)
				appTemplateInput.Labels["region"] = 123
				return appTemplateInput
			},
			AppTemplateRepoFn: UnusedAppTemplateRepo,
			WebhookRepoFn:     UnusedWebhookRepo,
			LabelUpsertSvcFn:  UnusedLabelUpsertSvc,
			LabelRepoFn:       UnusedLabelRepo,
			ExpectedError:     errors.New("\"region\" label value must be string"),
		},
		{
			Name: "Error when checking application type label - labels are not map[string]interface{}",
			Input: func() *model.ApplicationTemplateInput {
				appInputJSON := `{"name":"foo","providerName":"compass","description":"Lorem ipsum","labels":123,"healthCheckURL":"https://foo.bar","webhooks":[{"type":"","url":"webhook1.foo.bar","auth":null},{"type":"","url":"webhook2.foo.bar","auth":null}],"integrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`
				return fixModelAppTemplateInput(testName, appInputJSON)
			},
			AppTemplateRepoFn: UnusedAppTemplateRepo,
			WebhookRepoFn:     UnusedWebhookRepo,
			LabelUpsertSvcFn:  UnusedLabelUpsertSvc,
			LabelRepoFn:       UnusedLabelRepo,
			ExpectedError:     errors.New("app input json labels are type map[string]interface {} instead of map[string]interface{}"),
		},
		{
			Name: "Error when checking application type label - application type is not string",
			Input: func() *model.ApplicationTemplateInput {
				appInputJSON := `{"name":"foo","providerName":"compass","description":"Lorem ipsum","labels":{"applicationType":123,"test":["val","val2"]},"healthCheckURL":"https://foo.bar","webhooks":[{"type":"","url":"webhook1.foo.bar","auth":null},{"type":"","url":"webhook2.foo.bar","auth":null}],"integrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`
				return fixModelAppTemplateInput(testName, appInputJSON)
			},
			AppTemplateRepoFn: UnusedAppTemplateRepo,
			WebhookRepoFn:     UnusedWebhookRepo,
			LabelUpsertSvcFn:  UnusedLabelUpsertSvc,
			LabelRepoFn:       UnusedLabelRepo,
			ExpectedError:     errors.New("\"applicationType\" label value must be string"),
		},
		{
			Name: "Error when checking application type label - application type does not match <name> (<region>)",
			Input: func() *model.ApplicationTemplateInput {
				return fixModelAppTemplateInput(testName, fmt.Sprintf(appInputJSONWithAppTypeLabelString, "random-name"))
			},
			AppTemplateRepoFn: UnusedAppTemplateRepo,
			WebhookRepoFn:     UnusedWebhookRepo,
			LabelUpsertSvcFn:  UnusedLabelUpsertSvc,
			LabelRepoFn:       UnusedLabelRepo,
			ExpectedError:     errors.New("\"applicationType\" label value does not follow \"<app_template_name> (<region>)\""),
		},
		{
			Name: "Error when checking if application template already exists - GetByName returns error",
			Input: func() *model.ApplicationTemplateInput {
				return fixModelAppTemplateInput(testName, appInputJSONString)
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return(nil, testError).Once()
				return appTemplateRepo
			},
			WebhookRepoFn:    UnusedWebhookRepo,
			LabelUpsertSvcFn: UnusedLabelUpsertSvc,
			LabelRepoFn:      UnusedLabelRepo,
			ExpectedError:    testError,
		},
		{
			Name: "Error when application template already exists",
			Input: func() *model.ApplicationTemplateInput {
				return fixModelAppTemplateInput(testName, appInputJSONString)
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn:    UnusedWebhookRepo,
			LabelUpsertSvcFn: UnusedLabelUpsertSvc,
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "region").Return(nil, apperrors.NewNotFoundError(resource.Label, "id")).Once()
				return labelRepo
			},
			ExpectedError: errors.New("application template with name \"bar\" already exists"),
		},
		{
			Name: "Error when creating webhooks",
			Input: func() *model.ApplicationTemplateInput {
				return fixModelAppTemplateInput(testName, appInputJSON)
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return([]*model.ApplicationTemplate{}, nil).Once()
				appTemplateRepo.On("Create", ctx, *modelAppTemplate).Return(nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("CreateMany", ctx, "", mock.AnythingOfType("[]*model.Webhook")).Return(testError).Once()
				return webhookRepo
			},
			LabelUpsertSvcFn: UnusedLabelUpsertSvc,
			LabelRepoFn:      UnusedLabelRepo,
			ExpectedError:    testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			labelUpsertSvc := testCase.LabelUpsertSvcFn()
			labelRepo := testCase.LabelRepoFn()
			idSvc := uidSvcFn()
			svc := apptemplate.NewService(appTemplateRepo, webhookRepo, idSvc, labelUpsertSvc, labelRepo)

			// WHEN
			result, err := svc.Create(ctx, *testCase.Input())

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			appTemplateRepo.AssertExpectations(t)
			labelUpsertSvc.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
			idSvc.AssertExpectations(t)
		})
	}
}

func TestService_CreateWithLabels(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testID)
		return uidSvc
	}

	appInputJSON := fmt.Sprintf(appInputJSONWithAppTypeLabelString, testName)
	modelAppTemplate := fixModelAppTemplateWithAppInputJSON(testID, testName, appInputJSON, []*model.Webhook{})

	testCases := []struct {
		Name              string
		Input             *model.ApplicationTemplateInput
		AppTemplateID     string
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		WebhookRepoFn     func() *automock.WebhookRepository
		LabelUpsertSvcFn  func() *automock.LabelUpsertService
		ExpectedError     error
		ExpectedOutput    string
	}{
		{
			Name:  "Success",
			Input: fixModelAppTemplateInput(testName, appInputJSON),
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return([]*model.ApplicationTemplate{}, nil).Once()
				appTemplateRepo.On("Create", ctx, *modelAppTemplate).Return(nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("CreateMany", ctx, "", []*model.Webhook{}).Return(nil).Once()
				return webhookRepo
			},
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				labelUpsertService := &automock.LabelUpsertService{}
				labelUpsertService.On("UpsertMultipleLabels", ctx, "", model.AppTemplateLabelableObject, "foo", map[string]interface{}{"createWithLabels": "OK", "test": "test"}).Return(nil).Once()
				return labelUpsertService
			},
			ExpectedError:  nil,
			ExpectedOutput: testID,
		},
		{
			Name:  "Error when creating application template",
			Input: fixModelAppTemplateInput(testName, appInputJSON),
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return([]*model.ApplicationTemplate{}, nil).Once()
				appTemplateRepo.On("Create", ctx, *modelAppTemplate).Return(testError).Once()
				return appTemplateRepo
			},
			WebhookRepoFn:    UnusedWebhookRepo,
			LabelUpsertSvcFn: UnusedLabelUpsertSvc,
			ExpectedError:    testError,
			ExpectedOutput:   "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			labelUpsertSvc := testCase.LabelUpsertSvcFn()
			idSvc := uidSvcFn()
			svc := apptemplate.NewService(appTemplateRepo, webhookRepo, idSvc, labelUpsertSvc, nil)

			defer mock.AssertExpectationsForObjects(t, appTemplateRepo, labelUpsertSvc, idSvc)

			// WHEN
			result, err := svc.CreateWithLabels(ctx, *testCase.Input, map[string]interface{}{"createWithLabels": "OK"})

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)
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
			WebhookRepoFn:  UnusedWebhookRepo,
			ExpectedOutput: modelAppTemplate,
		},
		{
			Name: "Error when getting application template",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Get", ctx, testID).Return(nil, testError).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: UnusedWebhookRepo,
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, webhookRepo, nil, nil, nil)

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

func TestService_ListLabels(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	modelAppTemplate := fixModelApplicationTemplate(testID, testName, fixModelApplicationTemplateWebhooks(testWebhookID, testID))

	id := "foo"
	labelKey1 := "key1"
	labelValue1 := "val1"
	labelKey2 := "key2"
	labelValue2 := "val2"
	labels := map[string]*model.Label{
		"abc": {
			ID:         "abc",
			Tenant:     str.Ptr(testTenant),
			Key:        labelKey1,
			Value:      labelValue1,
			ObjectID:   id,
			ObjectType: model.AppTemplateLabelableObject,
		},
		"def": {
			ID:         "def",
			Tenant:     str.Ptr(testTenant),
			Key:        labelKey2,
			Value:      labelValue2,
			ObjectID:   id,
			ObjectType: model.AppTemplateLabelableObject,
		},
	}

	testCases := []struct {
		Name              string
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		LabelRepoFn       func() *automock.LabelRepository
		ExpectedError     error
		ExpectedOutput    *model.ApplicationTemplate
	}{
		{
			Name: "Success",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", ctx, testID).Return(true, nil).Once()
				return appTemplateRepo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctx, testTenant, model.AppTemplateLabelableObject, testID).Return(labels, nil).Once()
				return labelRepo
			},
			ExpectedOutput: modelAppTemplate,
		},
		{
			Name: "Error when listing labels",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", ctx, testID).Return(true, nil).Once()
				return appTemplateRepo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctx, testTenant, model.AppTemplateLabelableObject, testID).Return(nil, testError).Once()
				return labelRepo
			},
			ExpectedError: testError,
		},
		{
			Name: "Error when checking app template existence",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", ctx, testID).Return(false, testError).Once()
				return appTemplateRepo
			},
			LabelRepoFn:   UnusedLabelRepo,
			ExpectedError: testError,
		},
		{
			Name: "Error when app template does not exists",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", ctx, testID).Return(false, nil).Once()
				return appTemplateRepo
			},
			LabelRepoFn:   UnusedLabelRepo,
			ExpectedError: errors.New("application template with ID foo doesn't exist"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			labelRepo := testCase.LabelRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, nil, nil, nil, labelRepo)

			// WHEN
			result, err := svc.ListLabels(ctx, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, labels, result)
			}

			appTemplateRepo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetLabel(t *testing.T) {
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	id := "foo"
	labelKey1 := "key1"
	labelValue1 := "val1"
	labels := map[string]*model.Label{
		"abc": {
			ID:         "abc",
			Tenant:     str.Ptr(testTenant),
			Key:        labelKey1,
			Value:      labelValue1,
			ObjectID:   id,
			ObjectType: model.AppTemplateLabelableObject,
		},
	}

	testCases := []struct {
		Name              string
		Key               string
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		LabelRepoFn       func() *automock.LabelRepository
		ExpectedError     error
		ExpectedOutput    *model.Label
	}{
		{
			Name: "Success",
			Key:  "abc",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", ctx, testID).Return(true, nil).Once()
				return appTemplateRepo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctx, testTenant, model.AppTemplateLabelableObject, testID).Return(labels, nil).Once()
				return labelRepo
			},
			ExpectedOutput: labels["abc"],
			ExpectedError:  nil,
		},
		{
			Name: "Error when listing labels",
			Key:  "abc",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", ctx, testID).Return(true, nil).Once()
				return appTemplateRepo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctx, testTenant, model.AppTemplateLabelableObject, testID).Return(nil, testError).Once()
				return labelRepo
			},
			ExpectedOutput: nil,
			ExpectedError:  testError,
		},
		{
			Name: "Error when checking label with key",
			Key:  "fake-key",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", ctx, testID).Return(true, nil).Once()
				return appTemplateRepo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctx, testTenant, model.AppTemplateLabelableObject, testID).Return(labels, nil).Once()
				return labelRepo
			},
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("label fake-key for application template with ID %s doesn't exist", testID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			labelRepo := testCase.LabelRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, nil, nil, nil, labelRepo)

			// WHEN
			result, err := svc.GetLabel(ctx, testID, testCase.Key)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, result)
			}

			appTemplateRepo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetByName(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	modelAppTemplates := []*model.ApplicationTemplate{fixModelApplicationTemplate(testID, testName, fixModelApplicationTemplateWebhooks(testWebhookID, testID))}

	testCases := []struct {
		Name              string
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		ExpectedError     error
		ExpectedOutput    []*model.ApplicationTemplate
	}{
		{
			Name: "Success",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return(modelAppTemplates, nil).Once()
				return appTemplateRepo
			},
			ExpectedOutput: modelAppTemplates,
		},
		{
			Name: "Error when getting application template",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return(nil, testError).Once()
				return appTemplateRepo
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, nil, nil, nil, nil)

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

func TestService_GetByNameAndRegion(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)
	modelAppTemplate := fixModelApplicationTemplate(testID, testName, fixModelApplicationTemplateWebhooks(testWebhookID, testID))
	modelAppTemplates := []*model.ApplicationTemplate{modelAppTemplate}

	testCases := []struct {
		Name              string
		Region            interface{}
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		LabelRepoFn       func() *automock.LabelRepository
		ExpectedError     error
		ExpectedOutput    *model.ApplicationTemplate
	}{
		{
			Name:   "Success",
			Region: nil,
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return(modelAppTemplates, nil).Once()
				return appTemplateRepo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "region").Return(nil, apperrors.NewNotFoundError(resource.Label, "id")).Once()
				return labelRepo
			},
			ExpectedOutput: modelAppTemplate,
		},
		{
			Name:   "Success with region",
			Region: "eu-1",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return(modelAppTemplates, nil).Once()
				return appTemplateRepo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "region").Return(&model.Label{Value: "eu-1"}, nil).Once()
				return labelRepo
			},
			ExpectedOutput: modelAppTemplate,
		},
		{
			Name:   "Error when getting application templates by name",
			Region: nil,
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return(nil, testError).Once()
				return appTemplateRepo
			},
			LabelRepoFn:   UnusedLabelRepo,
			ExpectedError: testError,
		},
		{
			Name:   "Error when retrieving region label",
			Region: nil,
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return(modelAppTemplates, nil).Once()
				return appTemplateRepo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "region").Return(nil, testError).Once()
				return labelRepo
			},
			ExpectedError: testError,
		},
		{
			Name:   "Error when application template not found",
			Region: nil,
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return([]*model.ApplicationTemplate{}, nil).Once()
				return appTemplateRepo
			},
			LabelRepoFn:   UnusedLabelRepo,
			ExpectedError: apperrors.NewNotFoundErrorWithType(resource.ApplicationTemplate),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			labelRepo := testCase.LabelRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, nil, nil, nil, labelRepo)

			// WHEN
			result, err := svc.GetByNameAndRegion(ctx, testName, testCase.Region)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			appTemplateRepo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetByNameAndSubaccount(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)
	modelAppTemplate := fixModelApplicationTemplate(testID, testName, fixModelApplicationTemplateWebhooks(testWebhookID, testID))
	modelAppTemplates := []*model.ApplicationTemplate{modelAppTemplate}

	testCases := []struct {
		Name              string
		Subaccount        string
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		LabelRepoFn       func() *automock.LabelRepository
		ExpectedError     error
		ExpectedOutput    *model.ApplicationTemplate
	}{
		{
			Name:       "Success",
			Subaccount: "",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return(modelAppTemplates, nil).Once()
				return appTemplateRepo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "global_subaccount_id").Return(nil, apperrors.NewNotFoundError(resource.Label, "id")).Once()
				return labelRepo
			},
			ExpectedOutput: modelAppTemplate,
		},
		{
			Name:       "Success matching subaccount",
			Subaccount: testTenant,
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return(modelAppTemplates, nil).Once()
				return appTemplateRepo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "global_subaccount_id").Return(&model.Label{Value: testTenant}, nil).Once()
				return labelRepo
			},
			ExpectedOutput: modelAppTemplate,
		},
		{
			Name:       "Error when getting application templates by name",
			Subaccount: "",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return(nil, testError).Once()
				return appTemplateRepo
			},
			LabelRepoFn:   UnusedLabelRepo,
			ExpectedError: testError,
		},
		{
			Name:       "Error when retrieving subaccount label",
			Subaccount: "",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return(modelAppTemplates, nil).Once()
				return appTemplateRepo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "global_subaccount_id").Return(nil, testError).Once()
				return labelRepo
			},
			ExpectedError: testError,
		},
		{
			Name:       "Error when application template not found",
			Subaccount: "",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName).Return([]*model.ApplicationTemplate{}, nil).Once()
				return appTemplateRepo
			},
			LabelRepoFn:   UnusedLabelRepo,
			ExpectedError: apperrors.NewNotFoundErrorWithType(resource.ApplicationTemplate),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			labelRepo := testCase.LabelRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, nil, nil, nil, labelRepo)

			// WHEN
			result, err := svc.GetByNameAndSubaccount(ctx, testName, testCase.Subaccount)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			appTemplateRepo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetByFilters(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	modelAppTemplate := fixModelApplicationTemplate(testID, testName, fixModelApplicationTemplateWebhooks(testWebhookID, testID))
	filters := []*labelfilter.LabelFilter{labelfilter.NewForKey("someKey")}

	testCases := []struct {
		Name              string
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		ExpectedError     error
		ExpectedOutput    *model.ApplicationTemplate
	}{
		{
			Name: "Success",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByFilters", ctx, filters).Return(modelAppTemplate, nil).Once()
				return appTemplateRepo
			},
			ExpectedOutput: modelAppTemplate,
		},
		{
			Name: "Error when getting application template",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByFilters", ctx, filters).Return(nil, testError).Once()
				return appTemplateRepo
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, nil, nil, nil, nil)

			// WHEN
			result, err := svc.GetByFilters(ctx, filters)

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
			WebhookRepoFn:  UnusedWebhookRepo,
			ExpectedOutput: true,
		},
		{
			Name: "Error when getting application template",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", ctx, testID).Return(false, testError).Once()
				return appTemplateRepo
			},
			WebhookRepoFn:  UnusedWebhookRepo,
			ExpectedError:  testError,
			ExpectedOutput: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, webhookRepo, nil, nil, nil)

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
	labelFilters := []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(RegionKey, "eu-1")}

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
				appTemplateRepo.On("List", ctx, labelFilters, 50, testCursor).Return(modelAppTemplate, nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn:  UnusedWebhookRepo,
			InputPageSize:  50,
			ExpectedOutput: modelAppTemplate,
		},
		{
			Name: "Error when listing application template",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("List", ctx, labelFilters, 50, testCursor).Return(model.ApplicationTemplatePage{}, testError).Once()
				return appTemplateRepo
			},
			WebhookRepoFn:  UnusedWebhookRepo,
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
			WebhookRepoFn:  UnusedWebhookRepo,
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
			WebhookRepoFn:  UnusedWebhookRepo,
			InputPageSize:  201,
			ExpectedError:  errors.New("page size must be between 1 and 200"),
			ExpectedOutput: model.ApplicationTemplatePage{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, webhookRepo, nil, nil, nil)

			// WHEN
			result, err := svc.List(ctx, labelFilters, testCase.InputPageSize, testCursor)

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
	appInputJSON := fmt.Sprintf(appInputJSONWithAppTypeLabelString, testName)
	modelAppTemplate := fixModelAppTemplateWithAppInputJSON(testID, testName, appInputJSON, nil)

	testCases := []struct {
		Name              string
		Input             func() *model.ApplicationTemplateUpdateInput
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		WebhookRepoFn     func() *automock.WebhookRepository
		LabelRepoFn       func() *automock.LabelRepository
		ExpectedError     error
	}{
		{
			Name: "Success",
			Input: func() *model.ApplicationTemplateUpdateInput {
				return fixModelAppTemplateUpdateInput(testName, appInputJSON)
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Get", ctx, modelAppTemplate.ID).Return(modelAppTemplate, nil).Once()
				appTemplateRepo.On("Update", ctx, *modelAppTemplate).Return(nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: UnusedWebhookRepo,
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "region").Return(nil, apperrors.NewNotFoundError(resource.Label, "id")).Once()
				return labelRepo
			},
		},
		{
			Name: "Success - app input json without labels",
			Input: func() *model.ApplicationTemplateUpdateInput {
				appInputJSON := `{"name":"foo","providerName":"compass","description":"Lorem ipsum","healthCheckURL":"https://foo.bar","webhooks":[{"type":"","url":"webhook1.foo.bar","auth":null},{"type":"","url":"webhook2.foo.bar","auth":null}],"integrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`
				return fixModelAppTemplateUpdateInput(testName, appInputJSON)
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Get", ctx, modelAppTemplate.ID).Return(modelAppTemplate, nil).Once()
				appTemplateRepo.On("Update", ctx, mock.AnythingOfType("model.ApplicationTemplate")).Return(nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: UnusedWebhookRepo,
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "region").Return(nil, apperrors.NewNotFoundError(resource.Label, "id")).Once()
				return labelRepo
			},
		},
		{
			Name: "Error when getting application template",
			Input: func() *model.ApplicationTemplateUpdateInput {
				return fixModelAppTemplateUpdateInput(testName, appInputJSON)
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Get", ctx, modelAppTemplate.ID).Return(nil, testError).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: UnusedWebhookRepo,
			LabelRepoFn:   UnusedLabelRepo,
			ExpectedError: testError,
		},
		{
			Name: "Error when creating applicationType from region - region is not string",
			Input: func() *model.ApplicationTemplateUpdateInput {
				return fixModelAppTemplateUpdateInput(testName, appInputJSON)
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Get", ctx, modelAppTemplate.ID).Return(modelAppTemplate, nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: UnusedWebhookRepo,
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "region").Return(&model.Label{Value: 123}, nil).Once()
				return labelRepo
			},
			ExpectedError: errors.New("\"region\" label value must be string"),
		},
		{
			Name: "Error when func enriching app input json with applicationType label - labels are not map[string]interface{}",
			Input: func() *model.ApplicationTemplateUpdateInput {
				appInputJSON := `{"name":"foo","providerName":"compass","description":"Lorem ipsum","labels":123,"healthCheckURL":"https://foo.bar","webhooks":[{"type":"","url":"webhook1.foo.bar","auth":null},{"type":"","url":"webhook2.foo.bar","auth":null}],"integrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`
				return fixModelAppTemplateUpdateInput(testName, appInputJSON)
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Get", ctx, modelAppTemplate.ID).Return(modelAppTemplate, nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: UnusedWebhookRepo,
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "region").Return(&model.Label{Value: "eu-1"}, nil).Once()
				return labelRepo
			},
			ExpectedError: errors.New("app input json labels are type map[string]interface {} instead of map[string]interface{}"),
		},
		{
			Name: "Error when func enriching app input json with applicationType label - application type is not string",
			Input: func() *model.ApplicationTemplateUpdateInput {
				appInputJSON := `{"name":"foo","providerName":"compass","description":"Lorem ipsum","labels":{"applicationType":123,"test":["val","val2"]},"healthCheckURL":"https://foo.bar","webhooks":[{"type":"","url":"webhook1.foo.bar","auth":null},{"type":"","url":"webhook2.foo.bar","auth":null}],"integrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`
				return fixModelAppTemplateUpdateInput(testName, appInputJSON)
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Get", ctx, modelAppTemplate.ID).Return(modelAppTemplate, nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: UnusedWebhookRepo,
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "region").Return(&model.Label{Value: "eu-1"}, nil).Once()
				return labelRepo
			},
			ExpectedError: errors.New("\"applicationType\" label value must be string"),
		},
		{
			Name: "Error when func enriching app input json with applicationType label - application type value does not follow <app_template_name> (<region>) schema",
			Input: func() *model.ApplicationTemplateUpdateInput {
				appInputJSON := `{"name":"foo","providerName":"compass","description":"Lorem ipsum","labels":{"applicationType":"random-text","test":["val","val2"]},"healthCheckURL":"https://foo.bar","webhooks":[{"type":"","url":"webhook1.foo.bar","auth":null},{"type":"","url":"webhook2.foo.bar","auth":null}],"integrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`
				return fixModelAppTemplateUpdateInput(testName, appInputJSON)
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Get", ctx, modelAppTemplate.ID).Return(modelAppTemplate, nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: UnusedWebhookRepo,
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "region").Return(&model.Label{Value: "eu-1"}, nil).Once()
				return labelRepo
			},
			ExpectedError: errors.New("\"applicationType\" label value does not follow \"<app_template_name> (<region>)\""),
		},
		{
			Name: "Error when updating application template - retrieve region failed",
			Input: func() *model.ApplicationTemplateUpdateInput {
				return fixModelAppTemplateUpdateInput(testName+"test", appInputJSON)
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Get", ctx, modelAppTemplate.ID).Return(modelAppTemplate, nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: UnusedWebhookRepo,
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "region").Return(nil, testError).Once()
				return labelRepo
			},
			ExpectedError: testError,
		},
		{
			Name: "Error when updating application template - exists check failed",
			Input: func() *model.ApplicationTemplateUpdateInput {
				return fixModelAppTemplateUpdateInput(testName+"test", appInputJSONString)
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName+"test").Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				appTemplateRepo.On("Get", ctx, modelAppTemplate.ID).Return(modelAppTemplate, nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: UnusedWebhookRepo,
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "region").Return(nil, apperrors.NewNotFoundError(resource.Label, "id")).Once()
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "region").Return(nil, testError).Once()
				return labelRepo
			},
			ExpectedError: testError,
		},
		{
			Name: "Error when application template already exists",
			Input: func() *model.ApplicationTemplateUpdateInput {
				return fixModelAppTemplateUpdateInput(testName+"test", appInputJSONString)
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("GetByName", ctx, testName+"test").Return([]*model.ApplicationTemplate{modelAppTemplate}, nil).Once()
				appTemplateRepo.On("Get", ctx, modelAppTemplate.ID).Return(modelAppTemplate, nil).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: UnusedWebhookRepo,
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "region").Return(nil, apperrors.NewNotFoundError(resource.Label, "id")).Once()
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "region").Return(nil, apperrors.NewNotFoundError(resource.Label, "id")).Once()
				return labelRepo
			},
			ExpectedError: errors.New("application template with name \"bartest\" already exists"),
		},
		{
			Name: "Error when updating application template - update failed",
			Input: func() *model.ApplicationTemplateUpdateInput {
				return fixModelAppTemplateUpdateInput(testName, appInputJSON)
			},
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Get", ctx, modelAppTemplate.ID).Return(modelAppTemplate, nil).Once()
				appTemplateRepo.On("Update", ctx, *modelAppTemplate).Return(testError).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: UnusedWebhookRepo,
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", ctx, "", model.AppTemplateLabelableObject, modelAppTemplate.ID, "region").Return(nil, apperrors.NewNotFoundError(resource.Label, "id")).Once()
				return labelRepo
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			labelRepo := testCase.LabelRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, webhookRepo, nil, nil, labelRepo)

			// WHEN
			err := svc.Update(ctx, testID, *testCase.Input())

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			appTemplateRepo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
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
			WebhookRepoFn: UnusedWebhookRepo,
		},
		{
			Name: "Error when deleting application template",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Delete", ctx, testID).Return(testError).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: UnusedWebhookRepo,
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, webhookRepo, nil, nil, nil)

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
	svc := apptemplate.NewService(nil, nil, nil, nil, nil)

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

func UnusedLabelRepo() *automock.LabelRepository {
	return &automock.LabelRepository{}
}

func UnusedWebhookRepo() *automock.WebhookRepository {
	return &automock.WebhookRepository{}
}

func UnusedLabelUpsertSvc() *automock.LabelUpsertService {
	return &automock.LabelUpsertService{}
}

func UnusedAppTemplateRepo() *automock.ApplicationTemplateRepository {
	return &automock.ApplicationTemplateRepository{}
}
