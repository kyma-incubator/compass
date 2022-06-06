package apptemplate_test

import (
	"context"
	"errors"
	"fmt"

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
		return len(webhooks) == 1 && webhooks[0].ObjectID == testID && webhooks[0].Type == model.WebhookTypeConfigurationChanged && *webhooks[0].URL == "foourl"
	}

	testCases := []struct {
		Name              string
		Input             *model.ApplicationTemplateInput
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		WebhookRepoFn     func() *automock.WebhookRepository
		LabelUpsertSvcFn  func() *automock.LabelUpsertService
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
				webhookRepo.On("CreateMany", ctx, "", []*model.Webhook{}).Return(nil).Once()
				return webhookRepo
			},
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				labelUpsertService := &automock.LabelUpsertService{}
				labelUpsertService.On("UpsertMultipleLabels", ctx, "", model.AppTemplateLabelableObject, "foo", map[string]interface{}{"test": "test"}).Return(nil).Once()
				return labelUpsertService
			},
			ExpectedOutput: testID,
		},
		{
			Name:  "Success when ID is already generated",
			Input: fixModelAppTemplateWithIDInput(testName, appInputJSONString, &predefinedID),
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				modelAppTemplateWithPredefinedID := *modelAppTemplate
				modelAppTemplateWithPredefinedID.ID = predefinedID
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
			ExpectedOutput: predefinedID,
		},
		{
			Name:  "Success for Application Template with webhooks",
			Input: appTemplateInputWithWebhooks,
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
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
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				return &automock.LabelUpsertService{}
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
				webhookRepo.On("CreateMany", ctx, "", mock.AnythingOfType("[]*model.Webhook")).Return(testError).Once()
				return webhookRepo
			},
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				return &automock.LabelUpsertService{}
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			labelUpsertSvc := testCase.LabelUpsertSvcFn()
			idSvc := uidSvcFn()
			svc := apptemplate.NewService(appTemplateRepo, webhookRepo, idSvc, labelUpsertSvc, nil)

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
			labelUpsertSvc.AssertExpectations(t)
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

	modelAppTemplate := fixModelApplicationTemplate(testID, testName, []*model.Webhook{})

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
			Input: fixModelAppTemplateInput(testName, appInputJSONString),
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
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
			Input: fixModelAppTemplateInput(testName, appInputJSONString),
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Create", ctx, *modelAppTemplate).Return(testError).Once()
				return appTemplateRepo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				return &automock.LabelUpsertService{}
			},
			ExpectedError:  testError,
			ExpectedOutput: "",
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
			LabelRepoFn: func() *automock.LabelRepository {
				return &automock.LabelRepository{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Error when app template does not exists",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", ctx, testID).Return(false, nil).Once()
				return appTemplateRepo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				return &automock.LabelRepository{}
			},
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
			svc := apptemplate.NewService(appTemplateRepo, webhookRepo, nil, nil, nil)

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
			svc := apptemplate.NewService(appTemplateRepo, webhookRepo, nil, nil, nil)

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
	modelAppTemplate := fixModelApplicationTemplate(testID, testName, nil)

	testCases := []struct {
		Name              string
		Input             *model.ApplicationTemplateUpdateInput
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		WebhookRepoFn     func() *automock.WebhookRepository
		ExpectedError     error
	}{
		{
			Name:  "Success",
			Input: fixModelAppTemplateUpdateInput(testName, appInputJSONString),
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
			Input: fixModelAppTemplateUpdateInput(testName, appInputJSONString),
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
			svc := apptemplate.NewService(appTemplateRepo, webhookRepo, nil, nil, nil)

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
