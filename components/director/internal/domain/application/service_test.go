package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/sirupsen/logrus"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// given
	timestamp := time.Now()
	testErr := errors.New("Test error")
	modelInput := model.ApplicationRegisterInput{
		Name: "foo.bar-not",
		Webhooks: []*model.WebhookInput{
			{URL: "test.foo.com"},
			{URL: "test.bar.com"},
		},
		Documents: []*model.DocumentInput{
			{Title: "foo", Description: "test", FetchRequest: &model.FetchRequestInput{URL: "doc.foo.bar"}},
			{Title: "bar", Description: "test"},
		},
		APIDefinitions: []*model.APIDefinitionInput{
			{
				Name: "foo",
				Spec: &model.APISpecInput{FetchRequest: &model.FetchRequestInput{URL: "api.foo.bar"}},
			}, {Name: "bar"},
		},
		EventDefinitions: []*model.EventDefinitionInput{
			{
				Name: "foo",
				Spec: &model.EventSpecInput{FetchRequest: &model.FetchRequestInput{URL: "eventapi.foo.bar"}},
			}, {Name: "bar"},
		},
		Labels: map[string]interface{}{
			"label": "value",
		},
		IntegrationSystemID: &intSysID,
	}

	defaultLabels := map[string]interface{}{
		model.ScenariosKey:      model.ScenariosDefaultValue,
		"integration-system-id": intSysID,
		"label":                 "value",
		"name":                  "foo.bar-not",
	}
	defaultLabelsWithoutIntSys := map[string]interface{}{
		model.ScenariosKey:      model.ScenariosDefaultValue,
		"integration-system-id": "",
		"name":                  "test",
	}

	id := "foo"

	tnt := "tenant"
	appModel := modelFromInput(modelInput, tnt, id)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	labelScenarios := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      model.ScenariosDefaultValue,
		ObjectID:   id,
		ObjectType: model.ApplicationLabelableObject,
	}

	testCases := []struct {
		Name               string
		AppRepoFn          func() *automock.ApplicationRepository
		WebhookRepoFn      func() *automock.WebhookRepository
		APIRepoFn          func() *automock.APIRepository
		EventAPIRepoFn     func() *automock.EventAPIRepository
		DocumentRepoFn     func() *automock.DocumentRepository
		FetchRequestRepoFn func() *automock.FetchRequestRepository
		IntSysRepoFn       func() *automock.IntegrationSystemRepository
		ScenariosServiceFn func() *automock.ScenariosService
		LabelServiceFn     func() *automock.LabelUpsertService
		UIDServiceFn       func() *automock.UIDService
		Input              model.ApplicationRegisterInput
		ExpectedErr        error
	}{
		{
			Name: "Success",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Create", ctx, mock.MatchedBy(appModel.ApplicationMatcherFn)).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, mock.Anything).Return(nil).Once()
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, &model.APIDefinition{ID: "foo", ApplicationID: "foo", Tenant: tnt, Name: "foo", Spec: &model.APISpec{}}).Return(nil).Once()
				repo.On("Create", ctx, &model.APIDefinition{ID: "foo", ApplicationID: "foo", Tenant: tnt, Name: "bar"}).Return(nil).Once()
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Create", ctx, &model.EventDefinition{ID: "foo", ApplicationID: "foo", Tenant: tnt, Name: "foo", Spec: &model.EventSpec{}}).Return(nil).Once()
				repo.On("Create", ctx, &model.EventDefinition{ID: "foo", ApplicationID: "foo", Tenant: tnt, Name: "bar"}).Return(nil).Once()
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Create", ctx, mock.Anything).Return(nil).Times(2)
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, fixFetchRequest("doc.foo.bar", model.DocumentFetchRequestReference, timestamp)).Return(nil).Once()
				repo.On("Create", ctx, fixFetchRequest("api.foo.bar", model.APIFetchRequestReference, timestamp)).Return(nil).Once()
				repo.On("Create", ctx, fixFetchRequest("eventapi.foo.bar", model.EventAPIFetchRequestReference, timestamp)).Return(nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("EnsureScenariosLabelDefinitionExists", contextThatHasTenant(tnt), tnt).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, defaultLabels).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Success when no labels provided",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Create", ctx, mock.MatchedBy(applicationMatcher("test", nil))).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, mock.Anything).Return(nil).Once()
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("EnsureScenariosLabelDefinitionExists", contextThatHasTenant(tnt), tnt).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, defaultLabelsWithoutIntSys).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, labelScenarios).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       model.ApplicationRegisterInput{Name: "test"},
			ExpectedErr: nil,
		},
		{
			Name: "Success when scenarios label provided",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Create", ctx, mock.MatchedBy(applicationMatcher("test", nil))).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, mock.Anything).Return(nil).Once()
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("EnsureScenariosLabelDefinitionExists", contextThatHasTenant(tnt), tnt).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, defaultLabelsWithoutIntSys).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, labelScenarios).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input: model.ApplicationRegisterInput{
				Name:   "test",
				Labels: defaultLabelsWithoutIntSys,
			},
			ExpectedErr: nil,
		},
		{
			Name: "Returns errors when ensuring scenarios label definition failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Create", ctx, mock.MatchedBy(appModel.ApplicationMatcherFn)).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("EnsureScenariosLabelDefinitionExists", contextThatHasTenant(tnt), tnt).Return(testErr).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, modelInput.Labels).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when application creation failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Create", ctx, mock.MatchedBy(appModel.ApplicationMatcherFn)).Return(testErr).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, modelInput.Labels).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when integration system doesn't exist",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(false, nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			Input:       modelInput,
			ExpectedErr: errors.New("does not exist"),
		},
		{
			Name: "Returns error when checking for integration system fails",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(false, testErr).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appRepo := testCase.AppRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			apiRepo := testCase.APIRepoFn()
			eventAPIRepo := testCase.EventAPIRepoFn()
			documentRepo := testCase.DocumentRepoFn()
			fetchRequestRepo := testCase.FetchRequestRepoFn()
			scenariosSvc := testCase.ScenariosServiceFn()
			labelSvc := testCase.LabelServiceFn()
			uidSvc := testCase.UIDServiceFn()
			intSysRepo := testCase.IntSysRepoFn()
			svc := application.NewService(appRepo, webhookRepo, apiRepo, eventAPIRepo, documentRepo, nil, nil, fetchRequestRepo, intSysRepo, labelSvc, scenariosSvc, uidSvc)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// when
			result, err := svc.Create(ctx, testCase.Input)

			// then
			assert.IsType(t, "string", result)
			if testCase.ExpectedErr != nil {
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			appRepo.AssertExpectations(t)
			webhookRepo.AssertExpectations(t)
			apiRepo.AssertExpectations(t)
			eventAPIRepo.AssertExpectations(t)
			documentRepo.AssertExpectations(t)
			fetchRequestRepo.AssertExpectations(t)
			scenariosSvc.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		svc := application.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// when
		_, err := svc.Create(context.TODO(), model.ApplicationRegisterInput{})
		assert.Equal(t, tenant.NoTenantError, err)
	})
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	tnt := "tenant"

	updatedName := "updatedn"
	updatedDescription := "updatedd"
	updatedURL := "updatedu"
	updateInput := fixModelApplicationUpdateInput(updatedName, updatedDescription, updatedURL)
	applicationModelBefore := fixModelApplicationWithAllUpdatableFields(id, tnt, "initialn", "initiald", "initialu")
	applicationModelAfter := fixModelApplicationWithAllUpdatableFields(id, tnt, updatedName, updatedDescription, updatedURL)
	intSysLabel := fixLabelInput("integration-system-id", intSysID, id, model.ApplicationLabelableObject)
	nameLabel := fixLabelInput("name", updatedName, id, model.ApplicationLabelableObject)
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testCases := []struct {
		Name               string
		AppRepoFn          func() *automock.ApplicationRepository
		IntSysRepoFn       func() *automock.IntegrationSystemRepository
		LabelUpsertSvcFn   func() *automock.LabelUpsertService
		Input              model.ApplicationUpdateInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, applicationModelAfter).Return(nil).Once()
				repo.On("Exists", ctx, tnt, id).Return(true, nil).Twice()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertLabel", ctx, tnt, intSysLabel).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, nameLabel).Return(nil).Once()
				return svc
			},
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when application update failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, applicationModelAfter).Return(testErr).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application retrieval failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(nil, testErr).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when Integration System does not exist",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(false, nil).Once()
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: errors.New("does not exist").Error(),
		},
		{
			Name: "Returns error ensuring Integration System existance failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(false, testErr).Once()
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when setting label fails",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, applicationModelAfter).Return(nil).Once()
				repo.On("Exists", ctx, tnt, id).Return(true, nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertLabel", ctx, tnt, intSysLabel).Return(testErr).Once()
				return svc
			},
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when app does not exist",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, applicationModelAfter).Return(nil).Once()
				repo.On("Exists", ctx, tnt, id).Return(false, nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: "doesn't exist",
		},
		{
			Name: "Returns error when ensuring app existence fails",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, applicationModelAfter).Return(nil).Once()
				repo.On("Exists", ctx, tnt, id).Return(false, testErr).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appRepo := testCase.AppRepoFn()
			intSysRepo := testCase.IntSysRepoFn()
			lblUpsrtSvc := testCase.LabelUpsertSvcFn()
			svc := application.NewService(appRepo, nil, nil, nil, nil, nil, nil, nil, intSysRepo, lblUpsrtSvc, nil, nil)

			// when
			err := svc.Update(ctx, testCase.InputID, testCase.Input)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			appRepo.AssertExpectations(t)
			intSysRepo.AssertExpectations(t)
			lblUpsrtSvc.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	// given
	testErr := errors.New("Test error")
	id := "foo"
	desc := "Lorem ipsum"
	tnt := "tenant"
	applicationModel := &model.Application{
		ID:          id,
		Name:        "foo",
		Description: &desc,
		Tenant:      tnt,
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testCases := []struct {
		Name               string
		AppRepoFn          func() *automock.ApplicationRepository
		Input              model.ApplicationRegisterInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Delete", ctx, applicationModel.Tenant, applicationModel.ID).Return(nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when application deletion failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Delete", ctx, applicationModel.Tenant, applicationModel.ID).Return(testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appRepo := testCase.AppRepoFn()
			svc := application.NewService(appRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

			// when
			err := svc.Delete(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			appRepo.AssertExpectations(t)
		})
	}
}

func TestService_Get(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"

	desc := "Lorem ipsum"

	applicationModel := &model.Application{
		ID:          "foo",
		Name:        "foo",
		Description: &desc,
	}

	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testCases := []struct {
		Name                string
		RepositoryFn        func() *automock.ApplicationRepository
		Input               model.ApplicationRegisterInput
		InputID             string
		ExpectedApplication *model.Application
		ExpectedErrMessage  string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModel, nil).Once()
				return repo
			},
			InputID:             id,
			ExpectedApplication: applicationModel,
			ExpectedErrMessage:  "",
		},
		{
			Name: "Returns error when application retrieval failed",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(nil, testErr).Once()
				return repo
			},
			InputID:             id,
			ExpectedApplication: applicationModel,
			ExpectedErrMessage:  testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

			// when
			app, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedApplication, app)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_List(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	modelApplications := []*model.Application{
		fixModelApplication("foo", "tenant-foo", "foo", "Lorem Ipsum"),
		fixModelApplication("bar", "tenant-bar", "bar", "Lorem Ipsum"),
	}
	applicationPage := &model.ApplicationPage{
		Data:       modelApplications,
		TotalCount: len(modelApplications),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}

	first := 2
	after := "test"
	filter := []*labelfilter.LabelFilter{{Key: ""}}

	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		InputLabelFilters  []*labelfilter.LabelFilter
		InputPageSize      int
		ExpectedResult     *model.ApplicationPage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("List", ctx, tnt, filter, first, after).Return(applicationPage, nil).Once()
				return repo
			},
			InputPageSize:      first,
			InputLabelFilters:  filter,
			ExpectedResult:     applicationPage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when application listing failed",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("List", ctx, tnt, filter, first, after).Return(nil, testErr).Once()
				return repo
			},
			InputPageSize:      first,
			InputLabelFilters:  filter,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when page size is less than 1",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			InputPageSize:      0,
			InputLabelFilters:  filter,
			ExpectedResult:     nil,
			ExpectedErrMessage: "page size must be between 1 and 100",
		},
		{
			Name: "Returns error when page size is bigger than 100",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			InputPageSize:      101,
			InputLabelFilters:  filter,
			ExpectedResult:     nil,
			ExpectedErrMessage: "page size must be between 1 and 100",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

			// when
			app, err := svc.List(ctx, testCase.InputLabelFilters, testCase.InputPageSize, after)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, app)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_ListByRuntimeID(t *testing.T) {
	runtimeUUID := uuid.New()
	testError := errors.New("test error")
	tenantUUID := uuid.New()
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantUUID.String())

	first := 10
	cursor := "test"
	scenarios := []interface{}{"Easter", "Christmas", "Winter-Sale"}
	scenarioLabel := model.Label{
		ID:    uuid.New().String(),
		Key:   model.ScenariosKey,
		Value: scenarios,
	}

	applications := []*model.Application{
		fixModelApplication("test1", "tenant-foo", "test1", "test1"),
		fixModelApplication("test2", "tenant-foo", "test2", "test2"),
	}
	applicationPage := fixApplicationPage(applications)
	emptyPage := model.ApplicationPage{
		TotalCount: 0,
		Data:       []*model.Application{},
		PageInfo:   &pagination.Page{StartCursor: "", EndCursor: "", HasNextPage: false}}

	testCases := []struct {
		Name                string
		Input               uuid.UUID
		RuntimeRepositoryFn func() *automock.RuntimeRepository
		LabelRepositoryFn   func() *automock.LabelRepository
		AppRepositoryFn     func() *automock.ApplicationRepository
		ExpectedResult      *model.ApplicationPage
		ExpectedError       error
	}{
		{
			Name:  "Success",
			Input: runtimeUUID,
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				runtimeRepository := &automock.RuntimeRepository{}
				runtimeRepository.On("Exists", ctx, tenantUUID.String(), runtimeUUID.String()).
					Return(true, nil).Once()
				return runtimeRepository
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				labelRepository := &automock.LabelRepository{}
				labelRepository.On("GetByKey", ctx, tenantUUID.String(), model.RuntimeLabelableObject, runtimeUUID.String(), model.ScenariosKey).
					Return(&scenarioLabel, nil).Once()
				return labelRepository
			},
			AppRepositoryFn: func() *automock.ApplicationRepository {
				appRepository := &automock.ApplicationRepository{}
				appRepository.On("ListByScenarios", ctx, tenantUUID, convertToStringArray(t, scenarios), first, cursor).
					Return(applicationPage, nil).Once()
				return appRepository
			},
			ExpectedError:  nil,
			ExpectedResult: applicationPage,
		},
		{
			Name:  "Success when scenarios label not set",
			Input: runtimeUUID,
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				runtimeRepository := &automock.RuntimeRepository{}
				runtimeRepository.On("Exists", ctx, tenantUUID.String(), runtimeUUID.String()).
					Return(true, nil).Once()
				return runtimeRepository
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				labelRepository := &automock.LabelRepository{}
				labelRepository.On("GetByKey", ctx, tenantUUID.String(), model.RuntimeLabelableObject, runtimeUUID.String(), model.ScenariosKey).
					Return(nil, apperrors.NewNotFoundError("")).Once()
				return labelRepository
			},
			AppRepositoryFn: func() *automock.ApplicationRepository {
				appRepository := &automock.ApplicationRepository{}
				return appRepository
			},
			ExpectedError: nil,
			ExpectedResult: &model.ApplicationPage{
				Data:       []*model.Application{},
				PageInfo:   &pagination.Page{},
				TotalCount: 0,
			},
		},
		{
			Name:  "Return error when checking of runtime existence failed",
			Input: runtimeUUID,
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				runtimeRepository := &automock.RuntimeRepository{}
				runtimeRepository.On("Exists", ctx, tenantUUID.String(), runtimeUUID.String()).
					Return(false, testError).Once()
				return runtimeRepository
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				labelRepository := &automock.LabelRepository{}
				return labelRepository
			},
			AppRepositoryFn: func() *automock.ApplicationRepository {
				appRepository := &automock.ApplicationRepository{}
				return appRepository
			},
			ExpectedError:  testError,
			ExpectedResult: nil,
		},
		{
			Name:  "Return error when runtime not exits",
			Input: runtimeUUID,
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				runtimeRepository := &automock.RuntimeRepository{}
				runtimeRepository.On("Exists", ctx, tenantUUID.String(), runtimeUUID.String()).
					Return(false, nil).Once()
				return runtimeRepository
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				labelRepository := &automock.LabelRepository{}
				return labelRepository
			},
			AppRepositoryFn: func() *automock.ApplicationRepository {
				appRepository := &automock.ApplicationRepository{}
				return appRepository
			},
			ExpectedError:  errors.New("runtime does not exist"),
			ExpectedResult: nil,
		},
		{
			Name:  "Return error when getting runtime scenarios by RuntimeID failed",
			Input: runtimeUUID,
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				runtimeRepository := &automock.RuntimeRepository{}
				runtimeRepository.On("Exists", ctx, tenantUUID.String(), runtimeUUID.String()).
					Return(true, nil).Once()
				return runtimeRepository
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				labelRepository := &automock.LabelRepository{}
				labelRepository.On("GetByKey", ctx, tenantUUID.String(), model.RuntimeLabelableObject, runtimeUUID.String(), model.ScenariosKey).
					Return(nil, testError).Once()
				return labelRepository
			},
			AppRepositoryFn: func() *automock.ApplicationRepository {
				appRepository := &automock.ApplicationRepository{}
				return appRepository
			},
			ExpectedError:  testError,
			ExpectedResult: nil,
		},
		{
			Name:  "Return error when listing application by scenarios failed",
			Input: runtimeUUID,
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				runtimeRepository := &automock.RuntimeRepository{}
				runtimeRepository.On("Exists", ctx, tenantUUID.String(), runtimeUUID.String()).
					Return(true, nil).Once()
				return runtimeRepository
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				labelRepository := &automock.LabelRepository{}
				labelRepository.On("GetByKey", ctx, tenantUUID.String(), model.RuntimeLabelableObject, runtimeUUID.String(), model.ScenariosKey).
					Return(&scenarioLabel, nil).Once()
				return labelRepository
			},
			AppRepositoryFn: func() *automock.ApplicationRepository {
				appRepository := &automock.ApplicationRepository{}
				appRepository.On("ListByScenarios", ctx, tenantUUID, convertToStringArray(t, scenarios), first, cursor).
					Return(nil, testError).Once()
				return appRepository
			},
			ExpectedError:  testError,
			ExpectedResult: nil,
		},
		{
			Name:  " Return empty page when runtime is not assigned to any scenario",
			Input: runtimeUUID,
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				runtimeRepository := &automock.RuntimeRepository{}
				runtimeRepository.On("Exists", ctx, tenantUUID.String(), runtimeUUID.String()).
					Return(true, nil).Once()
				return runtimeRepository
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				labelRepository := &automock.LabelRepository{}
				labelRepository.On("GetByKey", ctx, tenantUUID.String(), model.RuntimeLabelableObject, runtimeUUID.String(), model.ScenariosKey).
					Return(&model.Label{ID: uuid.New().String(), Key: model.ScenariosKey, Value: []interface{}{}}, nil).Once()
				return labelRepository
			},
			AppRepositoryFn: func() *automock.ApplicationRepository {
				appRepository := &automock.ApplicationRepository{}
				return appRepository
			},
			ExpectedError:  nil,
			ExpectedResult: &emptyPage,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			runtimeRepository := testCase.RuntimeRepositoryFn()
			labelRepository := testCase.LabelRepositoryFn()
			appRepository := testCase.AppRepositoryFn()
			svc := application.NewService(appRepository, nil, nil, nil, nil, runtimeRepository, labelRepository, nil, nil, nil, nil, nil)

			//WHEN
			results, err := svc.ListByRuntimeID(ctx, testCase.Input, first, cursor)

			//THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				logrus.Info(err)
				require.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedResult, results)
			runtimeRepository.AssertExpectations(t)
			labelRepository.AssertExpectations(t)
			appRepository.AssertExpectations(t)
		})
	}
}

func TestService_Exist(t *testing.T) {
	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)
	testError := errors.New("Test error")

	applicationID := "id"

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		InputApplicationID string
		ExptectedValue     bool
		ExpectedError      error
	}{
		{
			Name: "Application exits",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(true, nil)
				return repo
			},
			InputApplicationID: applicationID,
			ExptectedValue:     true,
			ExpectedError:      nil,
		},
		{
			Name: "Application not exits",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(false, nil)
				return repo
			},
			InputApplicationID: applicationID,
			ExptectedValue:     false,
			ExpectedError:      nil,
		},
		{
			Name: "Returns error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(false, testError)
				return repo
			},
			InputApplicationID: applicationID,
			ExptectedValue:     false,
			ExpectedError:      testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			appRepo := testCase.RepositoryFn()
			svc := application.NewService(appRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

			// WHEN
			value, err := svc.Exist(ctx, testCase.InputApplicationID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.Nil(t, err)
			}

			assert.Equal(t, testCase.ExptectedValue, value)
			appRepo.AssertExpectations(t)
		})
	}
}

func TestService_SetLabel(t *testing.T) {
	// given
	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testErr := errors.New("Test error")

	applicationID := "foo"

	label := &model.LabelInput{
		Key:        "key",
		Value:      []string{"value1"},
		ObjectID:   applicationID,
		ObjectType: model.ApplicationLabelableObject,
	}

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		LabelServiceFn     func() *automock.LabelUpsertService
		InputApplicationID string
		InputLabel         *model.LabelInput
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(true, nil).Once()

				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertLabel", ctx, tnt, label).Return(nil).Once()
				return svc
			},
			InputApplicationID: applicationID,
			InputLabel:         label,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when label set failed",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(true, nil).Once()

				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertLabel", ctx, tnt, label).Return(testErr).Once()
				return svc
			},
			InputApplicationID: applicationID,
			InputLabel:         label,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application retrieval failed",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(false, testErr).Once()

				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			InputApplicationID: applicationID,
			InputLabel:         label,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelSvc := testCase.LabelServiceFn()
			svc := application.NewService(repo, nil, nil, nil, nil, nil, nil, nil, nil, labelSvc, nil, nil)

			// when
			err := svc.SetLabel(ctx, testCase.InputLabel)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_GetLabel(t *testing.T) {
	// given
	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testErr := errors.New("Test error")

	applicationID := "foo"
	labelKey := "key"
	labelValue := []string{"value1"}

	label := &model.LabelInput{
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   applicationID,
		ObjectType: model.ApplicationLabelableObject,
	}

	modelLabel := &model.Label{
		ID:         "5d23d9d9-3d04-4fa9-95e6-d22e1ae62c11",
		Tenant:     tnt,
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   applicationID,
		ObjectType: model.ApplicationLabelableObject,
	}

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		LabelRepositoryFn  func() *automock.LabelRepository
		InputApplicationID string
		InputLabel         *model.LabelInput
		ExpectedLabel      *model.Label
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, applicationID, labelKey).Return(modelLabel, nil).Once()
				return repo
			},
			InputApplicationID: applicationID,
			InputLabel:         label,
			ExpectedLabel:      modelLabel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when label receiving failed",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(true, nil).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, applicationID, labelKey).Return(nil, testErr).Once()
				return repo
			},
			InputApplicationID: applicationID,
			InputLabel:         label,
			ExpectedLabel:      nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application doesn't exist",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(false, testErr).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			InputApplicationID: applicationID,
			InputLabel:         label,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := application.NewService(repo, nil, nil, nil, nil, nil, labelRepo, nil, nil, nil, nil, nil)

			// when
			l, err := svc.GetLabel(ctx, testCase.InputApplicationID, testCase.InputLabel.Key)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, l, testCase.ExpectedLabel)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
		})
	}
}

func TestService_ListLabel(t *testing.T) {
	// given
	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testErr := errors.New("Test error")

	applicationID := "foo"
	labelKey := "key"
	labelValue := []string{"value1"}

	label := &model.LabelInput{
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   applicationID,
		ObjectType: model.ApplicationLabelableObject,
	}

	modelLabel := &model.Label{
		ID:         "5d23d9d9-3d04-4fa9-95e6-d22e1ae62c11",
		Tenant:     tnt,
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   applicationID,
		ObjectType: model.ApplicationLabelableObject,
	}

	labels := map[string]*model.Label{"first": modelLabel, "second": modelLabel}
	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		LabelRepositoryFn  func() *automock.LabelRepository
		InputApplicationID string
		InputLabel         *model.LabelInput
		ExpectedOutput     map[string]*model.Label
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, applicationID).Return(labels, nil).Once()
				return repo
			},
			InputApplicationID: applicationID,
			InputLabel:         label,
			ExpectedOutput:     labels,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when labels receiving failed",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(true, nil).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, applicationID).Return(nil, testErr).Once()
				return repo
			},
			InputApplicationID: applicationID,
			InputLabel:         label,
			ExpectedOutput:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application doesn't exist",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(false, testErr).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			InputApplicationID: applicationID,
			InputLabel:         label,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := application.NewService(repo, nil, nil, nil, nil, nil, labelRepo, nil, nil, nil, nil, nil)

			// when
			l, err := svc.ListLabels(ctx, testCase.InputApplicationID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, l, testCase.ExpectedOutput)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
		})
	}
}

func TestService_DeleteLabel(t *testing.T) {
	// given
	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testErr := errors.New("Test error")

	applicationID := "foo"

	labelKey := "key"

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		LabelRepositoryFn  func() *automock.LabelRepository
		InputApplicationID string
		InputKey           string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, tnt, model.ApplicationLabelableObject, applicationID, labelKey).Return(nil).Once()
				return repo
			},
			InputApplicationID: applicationID,
			InputKey:           labelKey,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when label delete failed",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, tnt, model.ApplicationLabelableObject, applicationID, labelKey).Return(testErr).Once()
				return repo
			},
			InputApplicationID: applicationID,
			InputKey:           labelKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application retrieval failed",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(false, testErr).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			InputApplicationID: applicationID,
			InputKey:           labelKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when trying to delete scenarios label",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			InputApplicationID: applicationID,
			InputKey:           model.ScenariosKey,
			ExpectedErrMessage: "can not be deleted from application",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := application.NewService(repo, nil, nil, nil, nil, nil, labelRepo, nil, nil, nil, nil, nil)

			// when
			err := svc.DeleteLabel(ctx, testCase.InputApplicationID, testCase.InputKey)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

type testModel struct {
	ApplicationMatcherFn func(app *model.Application) bool
	Webhooks             []*model.Webhook
	Apis                 []*model.APIDefinition
	EventAPIs            []*model.EventDefinition
	Documents            []*model.Document
}

func modelFromInput(in model.ApplicationRegisterInput, tenant, applicationID string) testModel {
	applicationModelMatcherFn := applicationMatcher(in.Name, in.Description)

	var webhooksModel []*model.Webhook
	for _, item := range in.Webhooks {
		webhooksModel = append(webhooksModel, item.ToWebhook(uuid.New().String(), tenant, applicationID))
	}

	var apisModel []*model.APIDefinition
	for _, item := range in.APIDefinitions {
		apisModel = append(apisModel, item.ToAPIDefinition(uuid.New().String(), applicationID, tenant))
	}

	var eventAPIsModel []*model.EventDefinition
	for _, item := range in.EventDefinitions {
		eventAPIsModel = append(eventAPIsModel, item.ToEventDefinition(uuid.New().String(), tenant, applicationID))
	}

	var documentsModel []*model.Document
	for _, item := range in.Documents {
		documentsModel = append(documentsModel, item.ToDocument(uuid.New().String(), tenant, applicationID))
	}

	return testModel{
		ApplicationMatcherFn: applicationModelMatcherFn,
		Documents:            documentsModel,
		Apis:                 apisModel,
		EventAPIs:            eventAPIsModel,
		Webhooks:             webhooksModel,
	}
}

func convertToStringArray(t *testing.T, array []interface{}) []string {
	var stringArray []string
	for _, value := range array {
		convertedValue, ok := value.(string)
		require.True(t, ok, "Cannot convert array of interface{} to array of string in test method")
		stringArray = append(stringArray, convertedValue)
	}
	return stringArray
}

func applicationMatcher(name string, description *string) func(app *model.Application) bool {
	return func(app *model.Application) bool {
		return app.Name == name && app.Description == description
	}
}

func contextThatHasTenant(expectedTenant string) interface{} {
	return mock.MatchedBy(func(actual context.Context) bool {
		actualTenant, err := tenant.LoadFromContext(actual)
		if err != nil {
			return false
		}
		return actualTenant == expectedTenant
	})
}
