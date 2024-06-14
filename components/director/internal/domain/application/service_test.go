package application_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	tnt "github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"

	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

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

const testScenario = "test-scenario"

func TestService_Create(t *testing.T) {
	// GIVEN
	timestamp := time.Now()
	testErr := errors.New("Test error")
	modelInput := applicationRegisterInput()

	normalizedModelInput := model.ApplicationRegisterInput{
		Name: "mp-foo-bar-not",
		Webhooks: []*model.WebhookInput{
			{URL: stringPtr("test.foo.com")},
			{URL: stringPtr("test.bar.com")},
		},

		Labels: map[string]interface{}{
			"label": "value",
		},
		IntegrationSystemID: &intSysID,
	}
	normalizedModelInput.Bundles = modelInput.Bundles

	labels := map[string]interface{}{
		"integrationSystemID":  intSysID,
		"label":                "value",
		"name":                 "mp-foo-bar-not",
		"applicationType":      "test-app-with-ppms",
		"ppmsProductVersionId": "1",
	}
	normalizedLabels := map[string]interface{}{
		"integrationSystemID": intSysID,
		"label":               "value",
		"name":                "mp-foo-bar-not",
	}
	labelsWithoutIntSys := map[string]interface{}{
		"name": "mp-test",
	}
	var nilLabels map[string]interface{}

	id := "foo"

	tnt := "tenant"
	externalTnt := "external-tnt"

	appModel := modelFromInput(modelInput, tnt, id, applicationMatcher(modelInput.Name, modelInput.Description))
	normalizedAppModel := modelFromInput(normalizedModelInput, tnt, id, applicationMatcher(normalizedModelInput.Name, normalizedModelInput.Description))

	appWithScenariosLabel := applicationRegisterInput()
	appWithScenariosLabel.Labels[model.ScenariosKey] = []string{testScenario}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name              string
		AppNameNormalizer normalizer.Normalizator
		AppRepoFn         func() *automock.ApplicationRepository
		WebhookRepoFn     func() *automock.WebhookRepository
		IntSysRepoFn      func() *automock.IntegrationSystemRepository
		LabelServiceFn    func() *automock.LabelService
		BundleServiceFn   func() *automock.BundleService
		UIDServiceFn      func() *automock.UIDService
		Input             model.ApplicationRegisterInput
		ExpectedErr       error
	}{
		{
			Name:              "Returns success when listing existing applications returns empty applications",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return([]*model.Application{}, nil).Once()
				repo.On("Create", ctx, tnt, mock.MatchedBy(appModel.ApplicationMatcherFn)).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, tnt, mock.Anything).Return(nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, resource.Application, id, modelInput.Bundles).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       applicationRegisterInput(),
			ExpectedErr: nil,
		},
		{
			Name:              "Returns success when listing existing applications returns application with different name",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return([]*model.Application{{Name: modelInput.Name + "-test"}}, nil).Once()
				repo.On("Create", ctx, tnt, mock.MatchedBy(appModel.ApplicationMatcherFn)).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, tnt, mock.Anything).Return(nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, resource.Application, id, modelInput.Bundles).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       applicationRegisterInput(),
			ExpectedErr: nil,
		},
		{
			Name:              "Returns error when listing existing applications returns application with same name",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return([]*model.Application{{Name: modelInput.Name}}, nil).Once()
				return repo
			},
			WebhookRepoFn:   UnusedWebhookRepository,
			IntSysRepoFn:    UnusedIntegrationSystemRepository,
			LabelServiceFn:  UnusedLabelService,
			BundleServiceFn: UnusedBundleService,
			UIDServiceFn:    UnusedUIDService,
			Input:           applicationRegisterInput(),
			ExpectedErr:     apperrors.NewNotUniqueNameError(resource.Application),
		},
		{
			Name:              "Returns success when listing existing applications returns application with different name and incoming name is already normalized",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return([]*model.Application{{Name: normalizedModelInput.Name + "-test"}}, nil).Once()
				repo.On("Create", ctx, tnt, mock.MatchedBy(normalizedAppModel.ApplicationMatcherFn)).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, tnt, mock.Anything).Return(nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, normalizedLabels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, resource.Application, id, normalizedModelInput.Bundles).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       normalizedModelInput,
			ExpectedErr: nil,
		},
		{
			Name:              "Returns error when listing existing applications returns application with same name and incoming name is already normalized",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return([]*model.Application{{Name: normalizedModelInput.Name}}, nil).Once()
				return repo
			},
			WebhookRepoFn:   UnusedWebhookRepository,
			IntSysRepoFn:    UnusedIntegrationSystemRepository,
			LabelServiceFn:  UnusedLabelService,
			BundleServiceFn: UnusedBundleService,
			UIDServiceFn:    UnusedUIDService,
			Input:           normalizedModelInput,
			ExpectedErr:     apperrors.NewNotUniqueNameError(resource.Application),
		},
		{
			Name:              "Success when no labels provided",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, nil).Once()
				repo.On("Create", ctx, tnt, mock.MatchedBy(applicationMatcher("test", nil))).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, tnt, mock.Anything).Return(nil).Once()
				return repo
			},
			IntSysRepoFn: UnusedIntegrationSystemRepository,
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labelsWithoutIntSys).Return(nil).Once()
				return svc
			},
			BundleServiceFn: UnusedBundleService,
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       model.ApplicationRegisterInput{Name: "test", Labels: nilLabels},
			ExpectedErr: nil,
		},
		{
			Name:              "Success when scenarios label provided",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, nil).Once()
				repo.On("Create", ctx, tnt, mock.MatchedBy(applicationMatcher("test", nil))).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, tnt, mock.Anything).Return(nil).Once()
				return repo
			},
			IntSysRepoFn: UnusedIntegrationSystemRepository,
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labelsWithoutIntSys).Return(nil).Once()
				return svc
			},
			BundleServiceFn: UnusedBundleService,
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input: model.ApplicationRegisterInput{
				Name:   "test",
				Labels: labelsWithoutIntSys,
			},
			ExpectedErr: nil,
		},
		{
			Name:              "Returns error when application creation failed",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, nil).Once()
				repo.On("Create", ctx, tnt, mock.MatchedBy(appModel.ApplicationMatcherFn)).Return(testErr).Once()
				return repo
			},
			WebhookRepoFn: UnusedWebhookRepository,
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn:  UnusedLabelService,
			BundleServiceFn: UnusedBundleService,
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			Input:       applicationRegisterInput(),
			ExpectedErr: testErr,
		},
		{
			Name:              "Returns error when listing existing applications fails",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, testErr).Once()
				return repo
			},
			WebhookRepoFn:   UnusedWebhookRepository,
			IntSysRepoFn:    UnusedIntegrationSystemRepository,
			LabelServiceFn:  UnusedLabelService,
			BundleServiceFn: UnusedBundleService,
			UIDServiceFn:    UnusedUIDService,
			Input:           applicationRegisterInput(),
			ExpectedErr:     testErr,
		},
		{
			Name:              "Returns error when integration system doesn't exist",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, nil).Once()
				return repo
			},
			WebhookRepoFn: UnusedWebhookRepository,
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(false, nil).Once()
				return repo
			},
			LabelServiceFn:  UnusedLabelService,
			BundleServiceFn: UnusedBundleService,
			UIDServiceFn:    UnusedUIDService,
			Input:           applicationRegisterInput(),
			ExpectedErr:     errors.New("Object not found"),
		},
		{
			Name:              "Returns error when checking for integration system fails",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, nil).Once()
				return repo
			},
			WebhookRepoFn: UnusedWebhookRepository,
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(false, testErr).Once()
				return repo
			},
			LabelServiceFn:  UnusedLabelService,
			BundleServiceFn: UnusedBundleService,
			UIDServiceFn:    UnusedUIDService,
			Input:           applicationRegisterInput(),
			ExpectedErr:     testErr,
		},
		{
			Name:              "Returns error when creating bundles",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, nil).Once()
				repo.On("Create", ctx, tnt, mock.MatchedBy(appModel.ApplicationMatcherFn)).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, tnt, mock.Anything).Return(nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, resource.Application, id, modelInput.Bundles).Return(testErr).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       applicationRegisterInput(),
			ExpectedErr: testErr,
		},
		{
			Name:              "Returns success when trying to create application with scenarios label",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return([]*model.Application{}, nil).Once()
				repo.On("Create", ctx, tnt, mock.MatchedBy(appModel.ApplicationMatcherFn)).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: UnusedWebhookRepository,
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn:  UnusedLabelService,
			BundleServiceFn: UnusedBundleService,
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       appWithScenariosLabel,
			ExpectedErr: fmt.Errorf("label with key %s cannot be set explicitly", model.ScenariosKey),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appNameNormalizer := testCase.AppNameNormalizer
			appRepo := testCase.AppRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			labelSvc := testCase.LabelServiceFn()
			uidSvc := testCase.UIDServiceFn()
			intSysRepo := testCase.IntSysRepoFn()
			bndlSvc := testCase.BundleServiceFn()
			svc := application.NewService(appNameNormalizer, nil, appRepo, webhookRepo, nil, nil, intSysRepo, labelSvc, bndlSvc, uidSvc, nil, "", nil)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// WHEN
			result, err := svc.Create(ctx, testCase.Input)

			// THEN
			assert.IsType(t, "string", result)
			if testCase.ExpectedErr != nil {
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, appRepo, webhookRepo, labelSvc, uidSvc, intSysRepo, bndlSvc)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		svc := application.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", nil)
		// WHEN
		_, err := svc.Create(context.TODO(), model.ApplicationRegisterInput{})
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
}

func TestService_CreateFromTemplate(t *testing.T) {
	// GIVEN
	timestamp := time.Now()
	testErr := errors.New("Test error")
	modelInput := applicationRegisterInput()

	normalizedModelInput := model.ApplicationRegisterInput{
		Name: "mp-foo-bar-not",
		Webhooks: []*model.WebhookInput{
			{URL: stringPtr("test.foo.com")},
			{URL: stringPtr("test.bar.com")},
		},

		Labels: map[string]interface{}{
			"label": "value",
		},
		IntegrationSystemID: &intSysID,
	}
	normalizedModelInput.Bundles = modelInput.Bundles

	labels := map[string]interface{}{
		"integrationSystemID":  intSysID,
		"label":                "value",
		"name":                 "mp-foo-bar-not",
		"applicationType":      "test-app-with-ppms",
		"ppmsProductVersionId": "1",
	}
	normalizedLabels := map[string]interface{}{
		"integrationSystemID": intSysID,
		"label":               "value",
		"name":                "mp-foo-bar-not",
	}
	labelsWithoutIntSys := map[string]interface{}{
		"name": "mp-test",
	}
	var nilLabels map[string]interface{}

	id := "foo"
	appTemplteID := "test-app-template"
	tnt := "tenant"
	externalTnt := "external-tnt"

	appFromTemplateModel := modelFromInput(modelInput, tnt, id, applicationFromTemplateMatcher(modelInput.Name, modelInput.Description, &appTemplteID))
	normalizedAppModel := modelFromInput(normalizedModelInput, tnt, id, applicationFromTemplateMatcher(normalizedModelInput.Name, normalizedModelInput.Description, &appTemplteID))

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name              string
		AppNameNormalizer normalizer.Normalizator
		AppRepoFn         func() *automock.ApplicationRepository
		WebhookRepoFn     func() *automock.WebhookRepository
		IntSysRepoFn      func() *automock.IntegrationSystemRepository
		LabelServiceFn    func() *automock.LabelService
		BundleServiceFn   func() *automock.BundleService
		UIDServiceFn      func() *automock.UIDService
		Input             model.ApplicationRegisterInput
		ExpectedErr       error
	}{
		{
			Name:              "Success",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, nil).Once()
				repo.On("Create", ctx, tnt, mock.MatchedBy(appFromTemplateModel.ApplicationMatcherFn)).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, tnt, mock.Anything).Return(nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, resource.Application, id, modelInput.Bundles).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       applicationRegisterInput(),
			ExpectedErr: nil,
		},
		{
			Name:              "Returns success when listing existing applications returns empty applications",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return([]*model.Application{}, nil).Once()
				repo.On("Create", ctx, tnt, mock.MatchedBy(appFromTemplateModel.ApplicationMatcherFn)).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, tnt, mock.Anything).Return(nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, resource.Application, id, modelInput.Bundles).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       applicationRegisterInput(),
			ExpectedErr: nil,
		},
		{
			Name:              "Returns success when listing existing applications returns application with different name",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return([]*model.Application{{Name: modelInput.Name + "-test"}}, nil).Once()
				repo.On("Create", ctx, tnt, mock.MatchedBy(appFromTemplateModel.ApplicationMatcherFn)).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, tnt, mock.Anything).Return(nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, resource.Application, id, modelInput.Bundles).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       applicationRegisterInput(),
			ExpectedErr: nil,
		},
		{
			Name:              "Returns error when listing existing applications returns application with same name",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return([]*model.Application{{Name: modelInput.Name}}, nil).Once()
				return repo
			},
			WebhookRepoFn:   UnusedWebhookRepository,
			IntSysRepoFn:    UnusedIntegrationSystemRepository,
			LabelServiceFn:  UnusedLabelService,
			BundleServiceFn: UnusedBundleService,
			UIDServiceFn:    UnusedUIDService,
			Input:           applicationRegisterInput(),
			ExpectedErr:     apperrors.NewNotUniqueNameError(resource.Application),
		},
		{
			Name:              "Returns success when listing existing applications returns application with different name and incoming name is already normalized",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return([]*model.Application{{Name: normalizedModelInput.Name + "-test"}}, nil).Once()
				repo.On("Create", ctx, tnt, mock.MatchedBy(normalizedAppModel.ApplicationMatcherFn)).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, tnt, mock.Anything).Return(nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, normalizedLabels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, resource.Application, id, normalizedModelInput.Bundles).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       normalizedModelInput,
			ExpectedErr: nil,
		},
		{
			Name:              "Returns error when listing existing applications returns application with same name and incoming name is already normalized",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return([]*model.Application{{Name: normalizedModelInput.Name}}, nil).Once()
				return repo
			},
			WebhookRepoFn:   UnusedWebhookRepository,
			IntSysRepoFn:    UnusedIntegrationSystemRepository,
			LabelServiceFn:  UnusedLabelService,
			BundleServiceFn: UnusedBundleService,
			UIDServiceFn:    UnusedUIDService,
			Input:           normalizedModelInput,
			ExpectedErr:     apperrors.NewNotUniqueNameError(resource.Application),
		},
		{
			Name:              "Success when no labels provided",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, nil).Once()
				repo.On("Create", ctx, tnt, mock.MatchedBy(applicationMatcher("test", nil))).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, tnt, mock.Anything).Return(nil).Once()
				return repo
			},
			IntSysRepoFn: UnusedIntegrationSystemRepository,
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labelsWithoutIntSys).Return(nil).Once()
				return svc
			},
			BundleServiceFn: UnusedBundleService,
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       model.ApplicationRegisterInput{Name: "test", Labels: nilLabels},
			ExpectedErr: nil,
		},
		{
			Name:              "Returns error when application creation failed",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, nil).Once()
				repo.On("Create", ctx, tnt, mock.MatchedBy(appFromTemplateModel.ApplicationMatcherFn)).Return(testErr).Once()
				return repo
			},
			WebhookRepoFn: UnusedWebhookRepository,
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn:  UnusedLabelService,
			BundleServiceFn: UnusedBundleService,
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			Input:       applicationRegisterInput(),
			ExpectedErr: testErr,
		},
		{
			Name:              "Returns error when listing existing applications fails",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, testErr).Once()
				return repo
			},
			WebhookRepoFn:   UnusedWebhookRepository,
			IntSysRepoFn:    UnusedIntegrationSystemRepository,
			LabelServiceFn:  UnusedLabelService,
			BundleServiceFn: UnusedBundleService,
			UIDServiceFn:    UnusedUIDService,
			Input:           applicationRegisterInput(),
			ExpectedErr:     testErr,
		},
		{
			Name:              "Returns error when integration system doesn't exist",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, nil).Once()
				return repo
			},
			WebhookRepoFn: UnusedWebhookRepository,
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(false, nil).Once()
				return repo
			},
			LabelServiceFn:  UnusedLabelService,
			BundleServiceFn: UnusedBundleService,
			UIDServiceFn:    UnusedUIDService,
			Input:           applicationRegisterInput(),
			ExpectedErr:     errors.New("Object not found"),
		},
		{
			Name:              "Returns error when checking for integration system fails",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, nil).Once()
				return repo
			},
			WebhookRepoFn: UnusedWebhookRepository,
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(false, testErr).Once()
				return repo
			},
			LabelServiceFn:  UnusedLabelService,
			BundleServiceFn: UnusedBundleService,
			UIDServiceFn:    UnusedUIDService,
			Input:           applicationRegisterInput(),
			ExpectedErr:     testErr,
		},
		{
			Name:              "Returns error when creating bundles",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, nil).Once()
				repo.On("Create", ctx, tnt, mock.MatchedBy(appFromTemplateModel.ApplicationMatcherFn)).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, tnt, mock.Anything).Return(nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, resource.Application, id, modelInput.Bundles).Return(testErr).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       applicationRegisterInput(),
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appNameNormalizer := testCase.AppNameNormalizer
			appRepo := testCase.AppRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			labelSvc := testCase.LabelServiceFn()
			uidSvc := testCase.UIDServiceFn()
			intSysRepo := testCase.IntSysRepoFn()
			bndlSvc := testCase.BundleServiceFn()
			svc := application.NewService(appNameNormalizer, nil, appRepo, webhookRepo, nil, nil, intSysRepo, labelSvc, bndlSvc, uidSvc, nil, "", nil)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// WHEN
			result, err := svc.CreateFromTemplate(ctx, testCase.Input, &appTemplteID, false)

			// then
			assert.IsType(t, "string", result)
			if testCase.ExpectedErr != nil {
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, appRepo, intSysRepo, uidSvc, bndlSvc, webhookRepo, labelSvc)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		svc := application.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", nil)
		// WHEN
		_, err := svc.Create(context.TODO(), model.ApplicationRegisterInput{})
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
}

func TestService_Upsert_TrustedUpsert(t *testing.T) {
	// given
	timestamp := time.Now()
	testErr := errors.New("Test error")
	modelInput := applicationRegisterInput()

	normalizedModelInput := model.ApplicationRegisterInput{
		Name: "mp-foo-bar-not",
		Webhooks: []*model.WebhookInput{
			{URL: stringPtr("test.foo.com")},
			{URL: stringPtr("test.bar.com")},
		},

		Labels: map[string]interface{}{
			"label": "value",
		},
		IntegrationSystemID: &intSysID,
	}
	normalizedModelInput.Bundles = modelInput.Bundles

	labels := map[string]interface{}{
		"integrationSystemID":  intSysID,
		"label":                "value",
		"name":                 "mp-foo-bar-not",
		"applicationType":      "test-app-with-ppms",
		"ppmsProductVersionId": "1",
	}

	labelsWithoutIntSys := map[string]interface{}{
		"name": "mp-test",
	}

	labelsWithInvalidPpmsProductVersion := map[string]interface{}{
		"integrationSystemID":  intSysID,
		"label":                "value",
		"name":                 "mp-foo-bar-not",
		"applicationType":      "test-app-with-ppms",
		"ppmsProductVersionId": "2",
	}

	var nilLabels map[string]interface{}

	modelInputWithoutApplicationType := model.ApplicationRegisterInput{
		Name: "foo.bar-not",
		Labels: map[string]interface{}{
			"label":                "value",
			"ppmsProductVersionId": "1",
		},
		IntegrationSystemID: &intSysID,
		BaseURL:             str.Ptr("http://test.com"),
	}

	id := "foo"

	tnt := "tenant"
	externalTnt := "external-tnt"

	appModel := modelFromInput(modelInput, tnt, id, applicationMatcher(modelInput.Name, modelInput.Description))

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	ordWebhookMapping := []application.ORDWebhookMapping{
		{
			Type: "test-app",
		},
		{
			Type:                "test-app-with-ppms",
			PpmsProductVersions: []string{"1"},
		},
	}

	testCases := []struct {
		Name              string
		AppNameNormalizer normalizer.Normalizator
		AppRepoFn         func(string) *automock.ApplicationRepository
		IntSysRepoFn      func() *automock.IntegrationSystemRepository
		LabelServiceFn    func() *automock.LabelService
		UIDServiceFn      func() *automock.UIDService
		WebhookRepoFn     func() *automock.WebhookRepository
		GetInput          func() model.ApplicationRegisterInput
		OrdWebhookMapping []application.ORDWebhookMapping
		ExpectedErr       error
	}{
		{
			Name:              "Success",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func(upsertMethodName string) *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On(upsertMethodName, ctx, mock.Anything, mock.MatchedBy(appModel.ApplicationMatcherFn)).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labels).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				wh := &model.Webhook{
					URL:      stringPtr("test.foo.com"),
					ObjectID: id,
				}
				webhookRepo.On("ListByReferenceObjectID", ctx, tnt, id, model.ApplicationWebhookReference).Return([]*model.Webhook{wh}, nil)
				webhookRepo.On("CreateMany", ctx, tnt, mock.AnythingOfType("[]*model.Webhook")).Return(nil)
				return webhookRepo
			},
			OrdWebhookMapping: ordWebhookMapping,
			GetInput: func() model.ApplicationRegisterInput {
				return applicationRegisterInput()
			},
			ExpectedErr: nil,
		},
		{
			Name:              "Success when no labels provided",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func(upsertMethodName string) *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On(upsertMethodName, ctx, mock.Anything, mock.MatchedBy(applicationMatcher("test", nil))).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: UnusedIntegrationSystemRepository,
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labelsWithoutIntSys).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			WebhookRepoFn: UnusedWebhookRepository,
			GetInput: func() model.ApplicationRegisterInput {
				return model.ApplicationRegisterInput{Name: "test", Labels: nilLabels}
			},
			ExpectedErr: nil,
		},
		{
			Name:              "Success when scenarios label not provided",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func(upsertMethodName string) *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On(upsertMethodName, ctx, mock.Anything, mock.MatchedBy(applicationMatcher("test", nil))).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: UnusedIntegrationSystemRepository,
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labelsWithoutIntSys).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			WebhookRepoFn: UnusedWebhookRepository,
			GetInput: func() model.ApplicationRegisterInput {
				return model.ApplicationRegisterInput{
					Name:   "test",
					Labels: labelsWithoutIntSys,
				}
			},
			ExpectedErr: nil,
		},
		{
			Name:              "Returns error when listing webhooks",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func(upsertMethodName string) *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On(upsertMethodName, ctx, mock.Anything, mock.MatchedBy(appModel.ApplicationMatcherFn)).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labels).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectID", ctx, tnt, id, model.ApplicationWebhookReference).Return(nil, testErr)
				return webhookRepo
			},
			OrdWebhookMapping: ordWebhookMapping,
			GetInput: func() model.ApplicationRegisterInput {
				return applicationRegisterInput()
			},
			ExpectedErr: testErr,
		},
		{
			Name:              "Returns error when creating webhooks",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func(upsertMethodName string) *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On(upsertMethodName, ctx, mock.Anything, mock.MatchedBy(appModel.ApplicationMatcherFn)).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labels).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectID", ctx, tnt, id, model.ApplicationWebhookReference).Return([]*model.Webhook{}, nil)
				webhookRepo.On("CreateMany", ctx, tnt, mock.AnythingOfType("[]*model.Webhook")).Return(testErr)
				return webhookRepo
			},
			OrdWebhookMapping: ordWebhookMapping,
			GetInput: func() model.ApplicationRegisterInput {
				return applicationRegisterInput()
			},
			ExpectedErr: testErr,
		},
		{
			Name:              "Returns error when application upsert failed",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func(upsertMethodName string) *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On(upsertMethodName, ctx, mock.Anything, mock.MatchedBy(appModel.ApplicationMatcherFn)).Return("", testErr).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: UnusedLabelService,
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			WebhookRepoFn: UnusedWebhookRepository,
			GetInput: func() model.ApplicationRegisterInput {
				return applicationRegisterInput()
			},
			ExpectedErr: testErr,
		},
		{
			Name:              "Returns error when integration system doesn't exist",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func(_ string) *automock.ApplicationRepository {
				return &automock.ApplicationRepository{}
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(false, nil).Once()
				return repo
			},
			LabelServiceFn: UnusedLabelService,
			UIDServiceFn:   UnusedUIDService,
			WebhookRepoFn:  UnusedWebhookRepository,
			GetInput: func() model.ApplicationRegisterInput {
				return applicationRegisterInput()
			},
			ExpectedErr: errors.New("Object not found"),
		},
		{
			Name:              "Returns error when checking for integration system fails",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func(_ string) *automock.ApplicationRepository {
				return &automock.ApplicationRepository{}
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(false, testErr).Once()
				return repo
			},
			LabelServiceFn: UnusedLabelService,
			UIDServiceFn:   UnusedUIDService,
			WebhookRepoFn:  UnusedWebhookRepository,
			GetInput: func() model.ApplicationRegisterInput {
				return applicationRegisterInput()
			},
			ExpectedErr: testErr,
		},
		{
			Name:              "Should not create webhooks when application type is missing from input labels",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func(upsertMethodName string) *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On(upsertMethodName, ctx, mock.Anything, mock.MatchedBy(appModel.ApplicationMatcherFn)).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				labels := map[string]interface{}{
					"integrationSystemID":  intSysID,
					"label":                "value",
					"name":                 "mp-foo-bar-not",
					"ppmsProductVersionId": "1",
				}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labels).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			WebhookRepoFn:     UnusedWebhookRepository,
			OrdWebhookMapping: ordWebhookMapping,
			GetInput: func() model.ApplicationRegisterInput {
				return modelInputWithoutApplicationType
			},
			ExpectedErr: nil,
		},
		{
			Name:              "Should not create webhooks when baseURL is missing from input",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func(upsertMethodName string) *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On(upsertMethodName, ctx, mock.Anything, mock.MatchedBy(appModel.ApplicationMatcherFn)).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labels).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			WebhookRepoFn:     UnusedWebhookRepository,
			OrdWebhookMapping: ordWebhookMapping,
			GetInput: func() model.ApplicationRegisterInput {
				modelInputWithoutBaseURL := applicationRegisterInput()
				modelInputWithoutBaseURL.BaseURL = str.Ptr("")
				return modelInputWithoutBaseURL
			},
			ExpectedErr: nil,
		},
		{
			Name:              "Should not create webhooks when baseURL is invalid",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func(upsertMethodName string) *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On(upsertMethodName, ctx, mock.Anything, mock.MatchedBy(appModel.ApplicationMatcherFn)).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labels).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			WebhookRepoFn:     UnusedWebhookRepository,
			OrdWebhookMapping: ordWebhookMapping,
			GetInput: func() model.ApplicationRegisterInput {
				modelInputWithInvalidBaseURL := applicationRegisterInput()
				modelInputWithInvalidBaseURL.BaseURL = str.Ptr("123://localhost")
				return modelInputWithInvalidBaseURL
			},
			ExpectedErr: nil,
		},
		{
			Name:              "Should not create webhooks when ppmsProductVersion is present in input but is not in configuration",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func(upsertMethodName string) *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On(upsertMethodName, ctx, mock.Anything, mock.MatchedBy(appModel.ApplicationMatcherFn)).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labelsWithInvalidPpmsProductVersion).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			WebhookRepoFn:     UnusedWebhookRepository,
			OrdWebhookMapping: ordWebhookMapping,
			GetInput: func() model.ApplicationRegisterInput {
				modelInputWithInvalidPpmsProductVersion := applicationRegisterInput()
				modelInputWithInvalidPpmsProductVersion.Labels["ppmsProductVersionId"] = "2"
				modelInputWithInvalidPpmsProductVersion.BaseURL = str.Ptr("123://localhost")
				return modelInputWithInvalidPpmsProductVersion
			},
			ExpectedErr: nil,
		},
		{
			Name:              "Error when scenarios label is provided explicitly",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func(upsertMethodName string) *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On(upsertMethodName, ctx, mock.Anything, mock.MatchedBy(appModel.ApplicationMatcherFn)).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: UnusedLabelService,
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			WebhookRepoFn:     UnusedWebhookRepository,
			OrdWebhookMapping: ordWebhookMapping,
			GetInput: func() model.ApplicationRegisterInput {
				app := applicationRegisterInput()
				app.Labels[model.ScenariosKey] = []string{testScenario}
				return app
			},
			ExpectedErr: fmt.Errorf("label with key %s cannot be set explicitly", model.ScenariosKey),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name+"_Upsert", func(t *testing.T) {
			appNameNormalizer := testCase.AppNameNormalizer
			appRepo := testCase.AppRepoFn("Upsert")
			labelSvc := testCase.LabelServiceFn()
			uidSvc := testCase.UIDServiceFn()
			intSysRepo := testCase.IntSysRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			svc := application.NewService(appNameNormalizer, nil, appRepo, webhookRepo, nil, nil, intSysRepo, labelSvc, nil, uidSvc, nil, "", testCase.OrdWebhookMapping)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// when
			err := svc.Upsert(ctx, testCase.GetInput())

			// then
			if testCase.ExpectedErr != nil {
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, appRepo, labelSvc, uidSvc, intSysRepo, webhookRepo)
		})

		t.Run(testCase.Name+"_TrustedUpsert", func(t *testing.T) {
			appNameNormalizer := testCase.AppNameNormalizer
			appRepo := testCase.AppRepoFn("TrustedUpsert")
			labelSvc := testCase.LabelServiceFn()
			uidSvc := testCase.UIDServiceFn()
			intSysRepo := testCase.IntSysRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			svc := application.NewService(appNameNormalizer, nil, appRepo, webhookRepo, nil, nil, intSysRepo, labelSvc, nil, uidSvc, nil, "", testCase.OrdWebhookMapping)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// when
			err := svc.TrustedUpsert(ctx, testCase.GetInput())

			// then
			if testCase.ExpectedErr != nil {
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, appRepo, labelSvc, uidSvc, intSysRepo, webhookRepo)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		svc := application.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", nil)
		// when
		_, err := svc.Create(context.TODO(), model.ApplicationRegisterInput{})
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
}

func TestService_TrustedUpsertFromTemplate(t *testing.T) {
	// given
	timestamp := time.Now()
	testErr := errors.New("Test error")
	modelInput := applicationRegisterInput()

	normalizedModelInput := model.ApplicationRegisterInput{
		Name: "mp-foo-bar-not",
		Webhooks: []*model.WebhookInput{
			{URL: stringPtr("test.foo.com")},
			{URL: stringPtr("test.bar.com")},
		},

		Labels: map[string]interface{}{
			"label": "value",
		},
		IntegrationSystemID: &intSysID,
	}
	normalizedModelInput.Bundles = modelInput.Bundles

	labels := map[string]interface{}{
		"integrationSystemID":  intSysID,
		"label":                "value",
		"name":                 "mp-foo-bar-not",
		"applicationType":      "test-app-with-ppms",
		"ppmsProductVersionId": "1",
	}
	labelsWithoutIntSys := map[string]interface{}{
		"name": "mp-test",
	}
	var nilLabels map[string]interface{}

	id := "foo"

	tnt := "tenant"
	externalTnt := "external-tnt"
	appTemplteID := "test-app-template"

	ordWebhookMapping := []application.ORDWebhookMapping{
		{
			Type: "test-app",
		},
		{
			Type:                "test-app-with-ppms",
			PpmsProductVersions: []string{"1"},
		},
	}

	appFromTemplateModel := modelFromInput(modelInput, tnt, id, applicationFromTemplateMatcher(modelInput.Name, modelInput.Description, &appTemplteID))

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name              string
		AppNameNormalizer normalizer.Normalizator
		AppRepoFn         func() *automock.ApplicationRepository
		IntSysRepoFn      func() *automock.IntegrationSystemRepository
		LabelServiceFn    func() *automock.LabelService
		UIDServiceFn      func() *automock.UIDService
		WebhookRepoFn     func() *automock.WebhookRepository
		Input             model.ApplicationRegisterInput
		OrdWebhookMapping []application.ORDWebhookMapping
		ExpectedErr       error
	}{
		{
			Name:              "Success",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("TrustedUpsert", ctx, mock.Anything, mock.MatchedBy(appFromTemplateModel.ApplicationMatcherFn)).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labels).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				wh := &model.Webhook{
					URL:      stringPtr("test.foo.com"),
					ObjectID: id,
				}
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectID", ctx, tnt, id, model.ApplicationWebhookReference).Return([]*model.Webhook{wh}, nil)
				webhookRepo.On("CreateMany", ctx, tnt, mock.AnythingOfType("[]*model.Webhook")).Return(nil)
				return webhookRepo
			},
			OrdWebhookMapping: ordWebhookMapping,
			Input:             applicationRegisterInput(),
			ExpectedErr:       nil,
		},
		{
			Name:              "Success when no labels provided",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("TrustedUpsert", ctx, mock.Anything, mock.MatchedBy(applicationMatcher("test", nil))).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: UnusedIntegrationSystemRepository,
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labelsWithoutIntSys).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			WebhookRepoFn: UnusedWebhookRepository,
			Input:         model.ApplicationRegisterInput{Name: "test", Labels: nilLabels},
			ExpectedErr:   nil,
		},
		{
			Name:              "Success when scenarios label is not provided",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("TrustedUpsert", ctx, mock.Anything, mock.MatchedBy(applicationMatcher("test", nil))).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: UnusedIntegrationSystemRepository,
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labelsWithoutIntSys).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			WebhookRepoFn: UnusedWebhookRepository,
			Input: model.ApplicationRegisterInput{
				Name:   "test",
				Labels: labelsWithoutIntSys,
			},
			ExpectedErr: nil,
		},
		{
			Name:              "Returns error when listing webhooks",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("TrustedUpsert", ctx, mock.Anything, mock.MatchedBy(appFromTemplateModel.ApplicationMatcherFn)).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labels).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectID", ctx, tnt, id, model.ApplicationWebhookReference).Return(nil, testErr)
				return webhookRepo
			},
			OrdWebhookMapping: ordWebhookMapping,
			Input:             modelInput,
			ExpectedErr:       testErr,
		},
		{
			Name:              "Returns error when creating webhooks",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("TrustedUpsert", ctx, mock.Anything, mock.MatchedBy(appFromTemplateModel.ApplicationMatcherFn)).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, map[string]interface{}{
					"integrationSystemID": intSysID,
					"label":               "value",
					"name":                "mp-foo-bar-not",
					"applicationType":     "test-app",
				}).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectID", ctx, tnt, id, model.ApplicationWebhookReference).Return([]*model.Webhook{}, nil)
				webhookRepo.On("CreateMany", ctx, tnt, mock.AnythingOfType("[]*model.Webhook")).Return(testErr)
				return webhookRepo
			},
			OrdWebhookMapping: ordWebhookMapping,
			Input: model.ApplicationRegisterInput{
				Name: "foo.bar-not",
				Labels: map[string]interface{}{
					"label":           "value",
					"applicationType": "test-app",
				},
				IntegrationSystemID: &intSysID,
				BaseURL:             str.Ptr("http://localhost.com"),
			},
			ExpectedErr: testErr,
		},
		{
			Name:              "Returns error when application trusted upsert failed",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("TrustedUpsert", ctx, mock.Anything, mock.MatchedBy(appFromTemplateModel.ApplicationMatcherFn)).Return("", testErr).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: UnusedLabelService,
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			WebhookRepoFn: UnusedWebhookRepository,
			Input:         applicationRegisterInput(),
			ExpectedErr:   testErr,
		},
		{
			Name:              "Returns error when integration system doesn't exist",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				return &automock.ApplicationRepository{}
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(false, nil).Once()
				return repo
			},
			LabelServiceFn: UnusedLabelService,
			UIDServiceFn:   UnusedUIDService,
			WebhookRepoFn:  UnusedWebhookRepository,
			Input:          applicationRegisterInput(),
			ExpectedErr:    errors.New("Object not found"),
		},
		{
			Name:              "Returns error when checking for integration system fails",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				return &automock.ApplicationRepository{}
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(false, testErr).Once()
				return repo
			},
			LabelServiceFn: UnusedLabelService,
			UIDServiceFn:   UnusedUIDService,
			WebhookRepoFn:  UnusedWebhookRepository,
			Input:          applicationRegisterInput(),
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appNameNormalizer := testCase.AppNameNormalizer
			appRepo := testCase.AppRepoFn()
			labelSvc := testCase.LabelServiceFn()
			uidSvc := testCase.UIDServiceFn()
			intSysRepo := testCase.IntSysRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			svc := application.NewService(appNameNormalizer, nil, appRepo, webhookRepo, nil, nil, intSysRepo, labelSvc, nil, uidSvc, nil, "", testCase.OrdWebhookMapping)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// when
			err := svc.TrustedUpsertFromTemplate(ctx, testCase.Input, &appTemplteID)

			// then
			if testCase.ExpectedErr != nil {
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, appRepo, labelSvc, uidSvc, intSysRepo, webhookRepo)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		svc := application.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", nil)
		// when
		_, err := svc.Create(context.TODO(), model.ApplicationRegisterInput{})
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
}

func TestService_Update(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	tnt := "tenant"
	externalTnt := "external-tnt"
	conditionTimestamp := time.Now()
	timestampGenFunc := func() time.Time { return conditionTimestamp }

	var updateInput model.ApplicationUpdateInput
	var applicationModelBefore *model.Application
	var applicationModelAfter *model.Application
	var intSysLabel *model.LabelInput
	var nameLabel *model.LabelInput
	var updateInputStatusOnly model.ApplicationUpdateInput
	var applicationModelAfterStatusUpdate *model.Application

	appTypeLabel := &model.Label{
		Key:   "applicationType",
		Value: "test-app",
	}
	ppmsVersionIDLabel := &model.Label{
		Key:   "ppmsProductVersionId",
		Value: "1",
	}
	ordWebhookMapping := []application.ORDWebhookMapping{
		{
			Type:                "test-app",
			PpmsProductVersions: []string{"1"},
			SubdomainSuffix:     "-test",
		},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	resetModels := func() {
		appName := "initialn"
		initialDescrription := "initald"
		initialURL := "initialu"
		updatedDescription := "updatedd"
		updatedHealthCheckURL := "updatedhcu"
		updatedBaseURL := "updatedbu"
		updatedApplicationNamespace := "updatedappns"
		updateInput = fixModelApplicationUpdateInput(appName, updatedDescription, updatedHealthCheckURL, updatedBaseURL, updatedApplicationNamespace, model.ApplicationStatusConditionConnected)
		applicationModelBefore = fixModelApplicationWithAllUpdatableFields(id, appName, initialDescrription, initialURL, nil, nil, model.ApplicationStatusConditionConnected, conditionTimestamp)
		applicationModelAfter = fixModelApplicationWithAllUpdatableFields(id, appName, updatedDescription, updatedHealthCheckURL, &updatedBaseURL, &updatedApplicationNamespace, model.ApplicationStatusConditionConnected, conditionTimestamp)
		intSysLabel = fixLabelInput("integrationSystemID", intSysID, id, model.ApplicationLabelableObject)
		intSysLabel.Version = 0
		nameLabel = fixLabelInput("name", "mp-"+appName, id, model.ApplicationLabelableObject)
		updateInputStatusOnly = fixModelApplicationUpdateInputStatus(model.ApplicationStatusConditionConnected)
		applicationModelAfterStatusUpdate = fixModelApplicationWithAllUpdatableFields(id, appName, initialDescrription, initialURL, nil, nil, model.ApplicationStatusConditionConnected, conditionTimestamp)
	}

	resetModels()

	testCases := []struct {
		Name               string
		AppNameNormalizer  normalizer.Normalizator
		AppRepoFn          func() *automock.ApplicationRepository
		IntSysRepoFn       func() *automock.IntegrationSystemRepository
		LabelSvcFn         func() *automock.LabelService
		WebhookRepoFn      func() *automock.WebhookRepository
		UIDServiceFn       func() *automock.UIDService
		Input              model.ApplicationUpdateInput
		InputID            string
		ORDWebhookMapping  []application.ORDWebhookMapping
		ExpectedErrMessage string
	}{
		{
			Name:              "Success",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, tnt, applicationModelAfter).Return(nil).Once()
				repo.On("Exists", ctx, tnt, id).Return(true, nil).Twice()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertLabel", ctx, tnt, intSysLabel).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, nameLabel).Return(nil).Once()
				svc.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, id, "applicationType").Return(appTypeLabel, nil).Once()
				svc.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, id, "ppmsProductVersionId").Return(ppmsVersionIDLabel, nil).Once()
				return svc
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				wh := &model.Webhook{
					URL:      stringPtr("test.foo.com"),
					ObjectID: id,
				}
				webhookRepo.On("ListByReferenceObjectID", ctx, tnt, id, model.ApplicationWebhookReference).Return([]*model.Webhook{wh}, nil)
				webhookRepo.On("CreateMany", ctx, tnt, mock.AnythingOfType("[]*model.Webhook")).Return(nil)
				return webhookRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			InputID:            "foo",
			ORDWebhookMapping:  ordWebhookMapping,
			Input:              updateInput,
			ExpectedErrMessage: "",
		},
		{
			Name:              "Success Status Condition Update",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, tnt, applicationModelAfterStatusUpdate).Return(nil).Once()
				repo.On("Exists", ctx, tnt, id).Return(true, nil).Once()
				return repo
			},
			IntSysRepoFn: UnusedIntegrationSystemRepository,
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertLabel", ctx, tnt, nameLabel).Return(nil).Once()
				svc.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, id, "applicationType").Return(appTypeLabel, nil).Once()
				svc.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, id, "ppmsProductVersionId").Return(ppmsVersionIDLabel, nil).Once()
				return svc
			},
			WebhookRepoFn:      UnusedWebhookRepository,
			UIDServiceFn:       UnusedUIDService,
			InputID:            "foo",
			Input:              updateInputStatusOnly,
			ExpectedErrMessage: "",
		},
		{
			Name:              "Returns error when application update failed",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, tnt, applicationModelAfter).Return(testErr).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelSvcFn:         UnusedLabelUpsertService,
			WebhookRepoFn:      UnusedWebhookRepository,
			UIDServiceFn:       UnusedUIDService,
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:              "Returns error when application retrieval failed",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
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
			LabelSvcFn:         UnusedLabelUpsertService,
			WebhookRepoFn:      UnusedWebhookRepository,
			UIDServiceFn:       UnusedUIDService,
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:              "Returns error when Integration System does not exist",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn:         UnusedApplicationRepository,
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(false, nil).Once()
				return repo
			},
			LabelSvcFn:         UnusedLabelUpsertService,
			WebhookRepoFn:      UnusedWebhookRepository,
			UIDServiceFn:       UnusedUIDService,
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: errors.New("Object not found").Error(),
		},
		{
			Name:              "Returns error ensuring Integration System existence failed",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn:         UnusedApplicationRepository,
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(false, testErr).Once()
				return repo
			},
			LabelSvcFn:         UnusedLabelUpsertService,
			WebhookRepoFn:      UnusedWebhookRepository,
			UIDServiceFn:       UnusedUIDService,
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:              "Returns error when setting label fails",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, tnt, applicationModelAfter).Return(nil).Once()
				repo.On("Exists", ctx, tnt, id).Return(true, nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertLabel", ctx, tnt, intSysLabel).Return(testErr).Once()
				return svc
			},
			WebhookRepoFn:      UnusedWebhookRepository,
			UIDServiceFn:       UnusedUIDService,
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:              "Returns error when app does not exist",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, tnt, applicationModelAfter).Return(nil).Once()
				repo.On("Exists", ctx, tnt, id).Return(false, nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelSvcFn:         UnusedLabelUpsertService,
			WebhookRepoFn:      UnusedWebhookRepository,
			UIDServiceFn:       UnusedUIDService,
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: "Object not found",
		},
		{
			Name:              "Returns error when ensuring app existence fails",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, tnt, applicationModelAfter).Return(nil).Once()
				repo.On("Exists", ctx, tnt, id).Return(false, testErr).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelSvcFn:         UnusedLabelUpsertService,
			WebhookRepoFn:      UnusedWebhookRepository,
			UIDServiceFn:       UnusedUIDService,
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:              "Should return error when fetching applicationType label",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, tnt, applicationModelAfter).Return(nil).Once()
				repo.On("Exists", ctx, tnt, id).Return(true, nil).Twice()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertLabel", ctx, tnt, intSysLabel).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, nameLabel).Return(nil).Once()
				svc.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, id, "applicationType").Return(nil, testErr).Once()
				return svc
			},
			WebhookRepoFn:      UnusedWebhookRepository,
			UIDServiceFn:       UnusedUIDService,
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:              "Should not create webhooks when applicationType label is missing",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, tnt, applicationModelAfter).Return(nil).Once()
				repo.On("Exists", ctx, tnt, id).Return(true, nil).Twice()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertLabel", ctx, tnt, intSysLabel).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, nameLabel).Return(nil).Once()
				svc.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, id, "applicationType").Return(nil, apperrors.NewNotFoundErrorWithType(resource.Label)).Once()
				return svc
			},
			WebhookRepoFn:      UnusedWebhookRepository,
			UIDServiceFn:       UnusedUIDService,
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: "",
		},
		{
			Name:              "Should return error when fetching applicationType label",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, tnt, applicationModelAfter).Return(nil).Once()
				repo.On("Exists", ctx, tnt, id).Return(true, nil).Twice()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertLabel", ctx, tnt, intSysLabel).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, nameLabel).Return(nil).Once()
				svc.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, id, "applicationType").Return(appTypeLabel, nil).Once()
				svc.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, id, "ppmsProductVersionId").Return(nil, testErr).Once()
				return svc
			},
			WebhookRepoFn:      UnusedWebhookRepository,
			UIDServiceFn:       UnusedUIDService,
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:              "Should return error while creating webhook",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, tnt, applicationModelAfter).Return(nil).Once()
				repo.On("Exists", ctx, tnt, id).Return(true, nil).Twice()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertLabel", ctx, tnt, intSysLabel).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, nameLabel).Return(nil).Once()
				svc.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, id, "applicationType").Return(appTypeLabel, nil).Once()
				svc.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, id, "ppmsProductVersionId").Return(ppmsVersionIDLabel, nil).Once()
				return svc
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectID", ctx, tnt, id, model.ApplicationWebhookReference).Return(nil, testErr)
				return webhookRepo
			},
			ORDWebhookMapping:  ordWebhookMapping,
			UIDServiceFn:       UnusedUIDService,
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:              "Not matching ord mapping configuration",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, tnt, applicationModelAfter).Return(nil).Once()
				repo.On("Exists", ctx, tnt, id).Return(true, nil).Twice()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				repo.On("Exists", ctx, intSysID).Return(true, nil).Once()
				return repo
			},
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertLabel", ctx, tnt, intSysLabel).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, nameLabel).Return(nil).Once()
				svc.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, id, "applicationType").Return(appTypeLabel, nil).Once()
				svc.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, id, "ppmsProductVersionId").Return(ppmsVersionIDLabel, nil).Once()
				return svc
			},
			WebhookRepoFn: UnusedWebhookRepository,
			ORDWebhookMapping: []application.ORDWebhookMapping{
				{
					Type:                "test-app-not-match",
					PpmsProductVersions: []string{"1"},
				},
			},
			UIDServiceFn:       UnusedUIDService,
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			resetModels()
			appNameNormalizer := testCase.AppNameNormalizer
			appRepo := testCase.AppRepoFn()
			intSysRepo := testCase.IntSysRepoFn()
			lblSvc := testCase.LabelSvcFn()
			webhookRepo := testCase.WebhookRepoFn()
			uidSvc := testCase.UIDServiceFn()
			svc := application.NewService(appNameNormalizer, nil, appRepo, webhookRepo, nil, nil, intSysRepo, lblSvc, nil, uidSvc, nil, "", testCase.ORDWebhookMapping)
			svc.SetTimestampGen(timestampGenFunc)

			// WHEN
			err := svc.Update(ctx, testCase.InputID, testCase.Input)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, uidSvc, webhookRepo, appRepo, intSysRepo, lblSvc)
		})
	}
}

func TestService_UpdateBaseURL(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	ctxErr := errors.New("while loading tenant from context: cannot read tenant from context")

	id := "foo"
	tnt := "tenant"
	externalTnt := "external-tnt"
	conditionTimestamp := time.Now()
	targetURL := "http://compass.kyma.local/api/event?myEvent=true"

	var applicationModelBefore *model.Application
	var applicationModelAfter *model.Application

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	resetModels := func() {
		appName := "initial"
		description := "description"
		updatedBaseURL := "http://compass.kyma.local"
		url := "url.com"
		applicationModelBefore = fixModelApplicationWithAllUpdatableFields(id, appName, description, url, nil, nil, model.ApplicationStatusConditionConnected, conditionTimestamp)
		applicationModelAfter = fixModelApplicationWithAllUpdatableFields(id, appName, description, url, &updatedBaseURL, nil, model.ApplicationStatusConditionConnected, conditionTimestamp)
	}

	resetModels()

	testCases := []struct {
		Name               string
		AppRepoFn          func() *automock.ApplicationRepository
		Input              model.ApplicationUpdateInput
		InputID            string
		TargetURL          string
		Context            context.Context
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, tnt, applicationModelAfter).Return(nil).Once()
				return repo
			},
			InputID:            id,
			TargetURL:          targetURL,
			Context:            ctx,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when application update failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, tnt, applicationModelAfter).Return(testErr).Once()
				return repo
			},
			InputID:            id,
			TargetURL:          targetURL,
			Context:            ctx,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application retrieval failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(nil, testErr).Once()
				repo.AssertNotCalled(t, "Update")
				return repo
			},
			InputID:            id,
			TargetURL:          targetURL,
			Context:            ctx,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when tenant is not in the context",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.AssertNotCalled(t, "Update")
				repo.AssertNotCalled(t, "GetByID")

				return repo
			},
			InputID:            id,
			TargetURL:          targetURL,
			Context:            context.Background(),
			ExpectedErrMessage: ctxErr.Error(),
		},
		{
			Name: "Does not update Application when BaseURL is already set",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.AssertNotCalled(t, "Update")
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelAfter, nil).Once()

				return repo
			},
			InputID:            id,
			TargetURL:          targetURL,
			Context:            ctx,
			ExpectedErrMessage: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			resetModels()
			appRepo := testCase.AppRepoFn()
			svc := application.NewService(nil, nil, appRepo, nil, nil, nil, nil, nil, nil, nil, nil, "", nil)

			// WHEN
			err := svc.UpdateBaseURL(testCase.Context, testCase.InputID, testCase.TargetURL)

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

func TestService_UpdateBaseURLAndReadyState(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	ctxErr := errors.New("while loading tenant from context: cannot read tenant from context")

	id := "foo"
	tnt := "tenant"
	externalTnt := "external-tnt"
	conditionTimestamp := time.Now()
	updatedBaseURL := "http://compass.kyma.local"

	var applicationModelBefore *model.Application
	var applicationModelAfter *model.Application

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	resetModels := func() {
		appName := "initial"
		description := "description"
		url := "url.com"
		applicationModelBefore = fixModelApplicationWithAllUpdatableFields(id, appName, description, url, nil, nil, model.ApplicationStatusConditionConnected, conditionTimestamp)
		applicationModelBefore.Ready = false
		applicationModelAfter = fixModelApplicationWithAllUpdatableFields(id, appName, description, url, &updatedBaseURL, nil, model.ApplicationStatusConditionConnected, conditionTimestamp)
		applicationModelAfter.Ready = true
	}

	resetModels()

	testCases := []struct {
		Name               string
		AppRepoFn          func() *automock.ApplicationRepository
		Input              model.ApplicationUpdateInput
		InputID            string
		TargetURL          string
		Context            context.Context
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, tnt, applicationModelAfter).Return(nil).Once()
				return repo
			},
			InputID:            id,
			Context:            ctx,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when application update failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, tnt, applicationModelAfter).Return(testErr).Once()
				return repo
			},
			InputID:            id,
			Context:            ctx,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application retrieval failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(nil, testErr).Once()
				repo.AssertNotCalled(t, "Update")
				return repo
			},
			InputID:            id,
			Context:            ctx,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when tenant is not in the context",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.AssertNotCalled(t, "Update")
				repo.AssertNotCalled(t, "GetByID")

				return repo
			},
			InputID:            id,
			Context:            context.Background(),
			ExpectedErrMessage: ctxErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			resetModels()
			appRepo := testCase.AppRepoFn()
			svc := application.NewService(nil, nil, appRepo, nil, nil, nil, nil, nil, nil, nil, nil, "", nil)

			// WHEN
			err := svc.UpdateBaseURLAndReadyState(testCase.Context, testCase.InputID, updatedBaseURL, true)

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

func TestService_Delete(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	id := "foo"
	desc := "Lorem ipsum"
	tnt := "tenant"
	externalTnt := "external-tnt"

	formations := []*model.Formation{
		{
			Name: testScenario,
		},
	}

	applicationModel := &model.Application{
		Name:        "foo",
		Description: &desc,
		BaseEntity:  &model.BaseEntity{ID: id},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		AppRepoFn          func() *automock.ApplicationRepository
		FormationServiceFn func() *automock.FormationService
		Input              model.ApplicationRegisterInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Delete", ctx, tnt, applicationModel.ID).Return(nil).Once()
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListFormationsForObjectGlobal", ctx, id).Return([]*model.Formation{}, nil).Once()
				return svc
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name: "Return error when application is part of a scenario",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, applicationModel.ID).Return(applicationModel, nil).Once()
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListFormationsForObjectGlobal", ctx, id).Return(formations, nil).Once()
				return svc
			},
			InputID:            id,
			ExpectedErrMessage: fmt.Sprintf("System foo is part of the following formations : %s", testScenario),
		},
		{
			Name: "Return error when application is part of a scenario",
			FormationServiceFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("ListFormationsForObjectGlobal", ctx, applicationModel.ID).Return(nil, testErr).Once()
				return formationSvc
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Return error when fails to get application by ID",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.AssertNotCalled(t, "Delete")
				repo.On("GetByID", ctx, tnt, applicationModel.ID).Return(nil, testErr).Once()
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListFormationsForObjectGlobal", ctx, id).Return(formations, nil).Once()
				return svc
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application deletion failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Delete", ctx, tnt, applicationModel.ID).Return(testErr).Once()
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListFormationsForObjectGlobal", ctx, id).Return([]*model.Formation{}, nil).Once()
				return svc
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appRepo := &automock.ApplicationRepository{}
			if testCase.AppRepoFn != nil {
				appRepo = testCase.AppRepoFn()
			}
			formationSvc := &automock.FormationService{}
			if testCase.FormationServiceFn != nil {
				formationSvc = testCase.FormationServiceFn()
			}
			svc := application.NewService(nil, nil, appRepo, nil, nil, nil, nil, nil, nil, nil, formationSvc, "", nil)

			// WHEN
			err := svc.Delete(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, appRepo, formationSvc)
		})
	}
}

func TestService_Unpair(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	formationAndRuntimeError := errors.New("The operation is not allowed [reason=System foo is still used and cannot be deleted. Unassign the system from the following formations first: test-scenario. Then, unassign the system from the following runtimes, too: test-runtime]")
	id := "foo"
	rtmID := "bar"
	desc := "Lorem ipsum"
	tnt := "tenant"
	externalTnt := "external-tnt"

	scenarios := []string{testScenario}

	formations := []*model.Formation{
		{
			Name: testScenario,
		},
	}

	timestamp := time.Now()

	applicationModel := &model.Application{
		Name:        "foo",
		Description: &desc,
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusConditionConnected,
			Timestamp: timestamp,
		},
		BaseEntity: &model.BaseEntity{ID: id},
	}

	applicationModelWithInitialStatus := &model.Application{
		Name:        "foo",
		Description: &desc,
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusConditionInitial,
			Timestamp: timestamp,
		},
		BaseEntity: &model.BaseEntity{ID: id},
	}

	runtimeModel := &model.Runtime{
		Name: "test-runtime",
	}

	ctx := context.Background()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		AppRepoFn          func() *automock.ApplicationRepository
		RuntimeRepoFn      func() *automock.RuntimeRepository
		FormationServiceFn func() *automock.FormationService
		Input              model.ApplicationRegisterInput
		ContextFn          func() context.Context
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Update", ctx, tnt, applicationModelWithInitialStatus).Return(nil).Once()
				repo.On("GetByID", ctx, tnt, applicationModelWithInitialStatus.ID).Return(applicationModelWithInitialStatus, nil).Once()
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, tnt, []string{}).Return([]*model.Runtime{}, nil).Once()
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListObjectIDsOfTypeForFormations", ctx, tnt, scenarios, model.FormationAssignmentTypeRuntime).Return([]string{}, nil).Once()
				svc.On("ListFormationsForObjectGlobal", ctx, id).Return(formations, nil).Once()
				return svc
			},
			ContextFn: func() context.Context {
				ctx := context.Background()

				return ctx
			},
			InputID: id,
		},
		{
			Name: "Success when application is part of a scenario but not with runtime",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Update", ctx, tnt, applicationModel).Return(nil).Once()
				repo.On("GetByID", ctx, tnt, applicationModel.ID).Return(applicationModel, nil).Once()
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, tnt, []string{}).Return([]*model.Runtime{}, nil).Once()
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListObjectIDsOfTypeForFormations", ctx, tnt, scenarios, model.FormationAssignmentTypeRuntime).Return([]string{}, nil).Once()
				svc.On("ListFormationsForObjectGlobal", ctx, id).Return(formations, nil).Once()
				return svc
			},
			ContextFn: func() context.Context {
				ctx := context.Background()

				return ctx
			},
			InputID: id,
		},
		{
			Name: "Success when operation type is SYNC and sets the application status to INITIAL",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Update", mock.Anything, tnt, applicationModelWithInitialStatus).Return(nil).Once()
				repo.On("GetByID", ctx, tnt, applicationModel.ID).Return(applicationModel, nil).Once()
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, tnt, []string{}).Return([]*model.Runtime{}, nil).Once()
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListObjectIDsOfTypeForFormations", ctx, tnt, scenarios, model.FormationAssignmentTypeRuntime).Return([]string{}, nil).Once()
				svc.On("ListFormationsForObjectGlobal", ctx, id).Return(formations, nil).Once()
				return svc
			},
			InputID: id,
			ContextFn: func() context.Context {
				backgroundCtx := context.Background()
				return backgroundCtx
			},
		},
		{
			Name: "Success when operation type is ASYNC and does not change the application status",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Update", mock.Anything, tnt, applicationModel).Return(nil).Once()
				repo.On("GetByID", mock.Anything, tnt, applicationModel.ID).Return(applicationModel, nil).Once()
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", mock.Anything, tnt, []string{}).Return([]*model.Runtime{}, nil).Once()
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListObjectIDsOfTypeForFormations", mock.Anything, tnt, scenarios, model.FormationAssignmentTypeRuntime).Return([]string{}, nil).Once()
				svc.On("ListFormationsForObjectGlobal", mock.Anything, id).Return(formations, nil).Once()
				return svc
			},
			InputID: id,
			ContextFn: func() context.Context {
				backgroundCtx := context.Background()
				backgroundCtx = operation.SaveModeToContext(backgroundCtx, graphql.OperationModeAsync)
				return backgroundCtx
			},
		},
		{
			Name: "Returns error when listing formations for object failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, applicationModel.ID).Return(nil, testErr).Once()
				repo.AssertNotCalled(t, "Update")
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", mock.Anything, tnt, []string{}).Return([]*model.Runtime{}, nil).Once()
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListObjectIDsOfTypeForFormations", mock.Anything, tnt, scenarios, model.FormationAssignmentTypeRuntime).Return([]string{}, nil).Once()
				svc.On("ListFormationsForObjectGlobal", mock.Anything, id).Return(formations, nil).Once()
				return svc
			},
			ContextFn: func() context.Context {
				ctx := context.Background()

				return ctx
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application is part of a scenario with runtime",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.AssertNotCalled(t, "Delete")
				repo.On("GetByID", ctx, tnt, applicationModel.ID).Return(applicationModel, nil).Once()
				repo.AssertNotCalled(t, "Update")
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", mock.Anything, tnt, []string{rtmID}).Return([]*model.Runtime{runtimeModel}, nil).Once()
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListObjectIDsOfTypeForFormations", mock.Anything, tnt, scenarios, model.FormationAssignmentTypeRuntime).Return([]string{rtmID}, nil).Once()
				svc.On("ListFormationsForObjectGlobal", mock.Anything, id).Return(formations, nil).Once()
				return svc
			},
			ContextFn: func() context.Context {
				ctx := context.Background()

				return ctx
			},
			InputID:            id,
			ExpectedErrMessage: formationAndRuntimeError.Error(),
		},
		{
			Name: "Returns error when update fails",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.AssertNotCalled(t, "Delete")
				repo.On("Update", ctx, tnt, applicationModel).Return(testErr).Once()
				repo.On("GetByID", ctx, tnt, applicationModel.ID).Return(applicationModel, nil).Once()
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", mock.Anything, tnt, []string{}).Return([]*model.Runtime{}, nil).Once()
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListObjectIDsOfTypeForFormations", mock.Anything, tnt, scenarios, model.FormationAssignmentTypeRuntime).Return([]string{}, nil).Once()
				svc.On("ListFormationsForObjectGlobal", mock.Anything, id).Return(formations, nil).Once()
				return svc
			},
			ContextFn: func() context.Context {
				ctx := context.Background()

				return ctx
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appRepo := testCase.AppRepoFn()
			runtimeRepo := testCase.RuntimeRepoFn()
			formationSvc := &automock.FormationService{}
			if testCase.FormationServiceFn != nil {
				formationSvc = testCase.FormationServiceFn()
			}
			ctx := testCase.ContextFn()
			ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
			svc := application.NewService(nil, nil, appRepo, nil, runtimeRepo, nil, nil, nil, nil, nil, formationSvc, "", nil)
			svc.SetTimestampGen(func() time.Time { return timestamp })
			// WHEN
			err := svc.Unpair(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, appRepo, runtimeRepo, formationSvc)
		})
	}
}

func TestService_Merge(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	destID := "foo"
	srcID := "bar"
	tnt := "tenant"
	externalTnt := "external-tnt"
	baseURL := "http://localhost.com"
	templateID := "12346789"
	otherTemplateID := "qwerty"
	destName := "dest app"
	srcName := "src app"
	srcDescription := "Long src description"
	selfRegDistLabelKey := "subscriptionProviderId"

	labelKey1 := model.ScenariosKey
	labelKey2 := "managed"
	var labelValue1 []interface{}
	labelValue2 := []interface{}{"Easter", "Bunny"}

	upsertLabelValues := make(map[string]interface{})
	upsertLabelValues[labelKey1] = []string{"Easter", "Bunny"}
	upsertLabelValues[labelKey2] = "true"

	upsertLabelValuesWithManagedFalse := make(map[string]interface{})
	upsertLabelValuesWithManagedFalse[labelKey1] = []string{"Easter", "Bunny"}
	upsertLabelValuesWithManagedFalse[labelKey2] = "false"

	srcAppLabels := fixApplicationLabels(srcID, labelKey1, labelKey2, labelValue1, "true")
	destAppLabels := fixApplicationLabels(srcID, labelKey1, labelKey2, labelValue2, "false")
	srcAppLabelsWithFalseManaged := fixApplicationLabels(srcID, labelKey1, labelKey2, labelValue1, "false")
	appTemplateLabelsWithSelfRegDistLabelKey := map[string]*model.Label{
		selfRegDistLabelKey: {
			ID:         "abc",
			Tenant:     str.Ptr(tnt),
			Key:        selfRegDistLabelKey,
			Value:      labelValue1,
			ObjectID:   templateID,
			ObjectType: model.AppTemplateLabelableObject,
		},
	}

	srcModel := fixDetailedModelApplication(t, srcID, tnt, srcName, srcDescription)
	srcModel.ApplicationTemplateID = &templateID
	srcModel.Status.Timestamp = time.Time{}

	destModel := fixModelApplication(destID, tnt, destName, "")
	destModel.ApplicationTemplateID = &templateID
	destModel.BaseURL = srcModel.BaseURL
	destModel.Status.Timestamp = time.Time{}

	mergedDestModel := fixDetailedModelApplication(t, destID, tnt, destName, srcDescription)
	mergedDestModel.ApplicationTemplateID = &templateID
	mergedDestModel.Status.Timestamp = time.Time{}

	srcModelWithoutBaseURL := fixDetailedModelApplication(t, srcID, tnt, srcName, srcDescription)
	srcModelWithoutBaseURL.BaseURL = nil

	srcModelWithDifferentBaseURL := fixDetailedModelApplication(t, srcID, tnt, srcName, srcDescription)
	srcModelWithDifferentBaseURL.BaseURL = &baseURL

	srcModelWithDifferentTemplateID := fixDetailedModelApplication(t, srcID, tnt, srcName, srcDescription)
	srcModelWithDifferentTemplateID.ApplicationTemplateID = &otherTemplateID

	srcModelConnected := fixDetailedModelApplication(t, srcID, tnt, srcName, srcDescription)
	srcModelConnected.Status.Condition = model.ApplicationStatusConditionConnected
	srcModelConnected.ApplicationTemplateID = &templateID

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name                           string
		AppRepoFn                      func() *automock.ApplicationRepository
		LabelRepoFn                    func() *automock.LabelRepository
		LabelUpsertSvcFn               func() *automock.LabelService
		FormationServiceFn             func() *automock.FormationService
		ExpectedDestinationApplication *model.Application
		Ctx                            context.Context
		SourceID                       string
		DestinationID                  string
		ExpectedErrMessage             string
	}{
		{
			Name: "Success",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, destModel.ID).Return(destModel, nil).Once()
				repo.On("GetByID", ctx, tnt, destModel.ID).Return(destModel, nil).Once()
				repo.On("GetByID", ctx, tnt, srcModel.ID).Return(srcModel, nil).Once()
				repo.On("Update", ctx, tnt, destModel).Return(nil).Once()
				repo.On("Delete", ctx, tnt, srcModel.ID).Return(nil).Once()
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, srcModel.ID).Return(srcAppLabels, nil)
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, destModel.ID).Return(destAppLabels, nil)
				repo.On("ListForObject", ctx, tnt, model.AppTemplateLabelableObject, *srcModel.ApplicationTemplateID).Return(map[string]*model.Label{}, nil)
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, destModel.ID, upsertLabelValues).Return(nil)
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListFormationsForObjectGlobal", ctx, srcID).Return([]*model.Formation{}, nil).Twice()
				return svc
			},
			Ctx:                            ctx,
			DestinationID:                  destID,
			SourceID:                       srcID,
			ExpectedDestinationApplication: mergedDestModel,
			ExpectedErrMessage:             "",
		},
		{
			Name: "Success with managed \"false\" label when both labels are \"false\"",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, destModel.ID).Return(destModel, nil).Once()
				repo.On("GetByID", ctx, tnt, destModel.ID).Return(destModel, nil).Once()
				repo.On("GetByID", ctx, tnt, srcModel.ID).Return(srcModel, nil).Once()
				repo.On("Update", ctx, tnt, destModel).Return(nil).Once()
				repo.On("Delete", ctx, tnt, srcModel.ID).Return(nil).Once()
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, srcModel.ID).Return(srcAppLabelsWithFalseManaged, nil)
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, destModel.ID).Return(destAppLabels, nil)
				repo.On("ListForObject", ctx, tnt, model.AppTemplateLabelableObject, *srcModel.ApplicationTemplateID).Return(map[string]*model.Label{}, nil)
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, destModel.ID, upsertLabelValuesWithManagedFalse).Return(nil)
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListFormationsForObjectGlobal", ctx, srcID).Return([]*model.Formation{}, nil).Twice()
				return svc
			},
			Ctx:                            ctx,
			DestinationID:                  destID,
			SourceID:                       srcID,
			ExpectedDestinationApplication: mergedDestModel,
			ExpectedErrMessage:             "",
		},
		{
			Name: "Error when tenant is not in context",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.AssertNotCalled(t, "GetByID")
				repo.AssertNotCalled(t, "Exists")
				repo.AssertNotCalled(t, "Update")
				repo.AssertNotCalled(t, "Delete")
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.AssertNotCalled(t, "ListForObject")
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.AssertNotCalled(t, "UpsertMultipleLabels")
				return svc
			},
			Ctx:                context.Background(),
			DestinationID:      destID,
			SourceID:           srcID,
			ExpectedErrMessage: "while loading tenant from context: cannot read tenant from context",
		},
		{
			Name: "Error when cannot get destination application",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, srcModel.ID).Return(srcModel, nil).Once()
				repo.On("GetByID", ctx, tnt, destModel.ID).Return(nil, testErr).Once()
				repo.AssertNotCalled(t, "Update")
				repo.AssertNotCalled(t, "Delete")
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.AssertNotCalled(t, "ListForObject")
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.AssertNotCalled(t, "UpsertMultipleLabels")
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListFormationsForObjectGlobal", ctx, srcID).Return([]*model.Formation{}, nil).Once()
				return svc
			},
			Ctx:                ctx,
			DestinationID:      destID,
			SourceID:           srcID,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Error when cannot get source application",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, srcModel.ID).Return(nil, testErr).Once()
				repo.AssertNotCalled(t, "Exists")
				repo.AssertNotCalled(t, "Update")
				repo.AssertNotCalled(t, "Delete")
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.AssertNotCalled(t, "ListForObject")
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.AssertNotCalled(t, "UpsertMultipleLabels")
				return svc
			},
			Ctx:                ctx,
			DestinationID:      destID,
			SourceID:           srcID,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Error when source app and destination app templates do not match",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, destModel.ID).Return(destModel, nil).Once()
				repo.On("GetByID", ctx, tnt, srcModel.ID).Return(srcModelWithDifferentTemplateID, nil).Once()
				repo.AssertNotCalled(t, "Update")
				repo.AssertNotCalled(t, "Delete")
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, srcModel.ID).Return(srcAppLabels, nil)
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, destModel.ID).Return(destAppLabels, nil)
				repo.AssertNotCalled(t, "ListForObject")
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.AssertNotCalled(t, "UpsertMultipleLabels")
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListFormationsForObjectGlobal", ctx, srcID).Return([]*model.Formation{}, nil).Once()
				return svc
			},
			Ctx:                ctx,
			DestinationID:      destID,
			SourceID:           srcID,
			ExpectedErrMessage: "Application templates are not the same. Destination app template: 12346789. Source app template: qwerty",
		},
		{
			Name: "Error when source app and destination app base url do not match",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, destModel.ID).Return(destModel, nil).Once()
				repo.On("GetByID", ctx, tnt, srcModel.ID).Return(srcModelWithDifferentBaseURL, nil).Once()
				repo.AssertNotCalled(t, "Update")
				repo.AssertNotCalled(t, "Delete")
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, srcModel.ID).Return(srcAppLabels, nil)
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, destModel.ID).Return(destAppLabels, nil)
				repo.AssertNotCalled(t, "ListForObject")
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.AssertNotCalled(t, "UpsertMultipleLabels")
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListFormationsForObjectGlobal", ctx, srcID).Return([]*model.Formation{}, nil).Once()
				return svc
			},
			Ctx:                ctx,
			DestinationID:      destID,
			SourceID:           srcID,
			ExpectedErrMessage: "BaseURL for applications foo and bar are not the same.",
		},
		{
			Name: "Error when source app is in CONNECTED status",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, destModel.ID).Return(destModel, nil).Once()
				repo.On("GetByID", ctx, tnt, srcModel.ID).Return(srcModelConnected, nil).Once()
				repo.AssertNotCalled(t, "Update")
				repo.AssertNotCalled(t, "Delete")
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, srcModel.ID).Return(srcAppLabels, nil)
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, destModel.ID).Return(destAppLabels, nil)
				repo.On("ListForObject", ctx, tnt, model.AppTemplateLabelableObject, *srcModel.ApplicationTemplateID).Return(map[string]*model.Label{}, nil)
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.AssertNotCalled(t, "UpsertMultipleLabels")
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListFormationsForObjectGlobal", ctx, srcID).Return([]*model.Formation{}, nil).Once()
				return svc
			},
			Ctx:                ctx,
			DestinationID:      destID,
			SourceID:           srcID,
			ExpectedErrMessage: "Cannot merge application with id bar, because it is in a CONNECTED status",
		},
		{
			Name: "Error when source deletion fails",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, destModel.ID).Return(destModel, nil).Once()
				repo.On("GetByID", ctx, tnt, srcModel.ID).Return(srcModel, nil).Once()
				repo.AssertNotCalled(t, "Update")
				repo.On("Delete", ctx, tnt, srcModel.ID).Return(testErr).Once()
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, srcModel.ID).Return(srcAppLabels, nil)
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, destModel.ID).Return(destAppLabels, nil)
				repo.On("ListForObject", ctx, tnt, model.AppTemplateLabelableObject, *srcModel.ApplicationTemplateID).Return(map[string]*model.Label{}, nil)
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.AssertNotCalled(t, "UpsertMultipleLabels")
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListFormationsForObjectGlobal", ctx, srcID).Return([]*model.Formation{}, nil).Twice()
				return svc
			},
			Ctx:                ctx,
			DestinationID:      destID,
			SourceID:           srcID,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Error when destination app update fails",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, destModel.ID).Return(destModel, nil).Once()
				repo.On("GetByID", ctx, tnt, srcModel.ID).Return(srcModel, nil).Once()
				repo.On("Update", ctx, tnt, mergedDestModel).Return(testErr)
				repo.On("Delete", ctx, tnt, srcModel.ID).Return(nil).Once()
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, srcModel.ID).Return(srcAppLabels, nil)
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, destModel.ID).Return(destAppLabels, nil)
				repo.On("ListForObject", ctx, tnt, model.AppTemplateLabelableObject, *srcModel.ApplicationTemplateID).Return(map[string]*model.Label{}, nil)
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.AssertNotCalled(t, "UpsertMultipleLabels")
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListFormationsForObjectGlobal", ctx, srcID).Return([]*model.Formation{}, nil).Twice()
				return svc
			},
			Ctx:                ctx,
			DestinationID:      destID,
			SourceID:           srcID,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Error when update labels fails",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, destModel.ID).Return(destModel, nil).Once()
				repo.On("GetByID", ctx, tnt, srcModel.ID).Return(srcModel, nil).Once()
				repo.On("Update", ctx, tnt, mergedDestModel).Return(nil)
				repo.On("Delete", ctx, tnt, srcModel.ID).Return(nil).Once()
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, srcModel.ID).Return(srcAppLabels, nil)
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, destModel.ID).Return(destAppLabels, nil)
				repo.On("ListForObject", ctx, tnt, model.AppTemplateLabelableObject, *srcModel.ApplicationTemplateID).Return(map[string]*model.Label{}, nil)
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, destModel.ID, upsertLabelValues).Return(testErr)
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListFormationsForObjectGlobal", ctx, srcID).Return([]*model.Formation{}, nil).Twice()
				return svc
			},
			Ctx:                ctx,
			DestinationID:      destID,
			SourceID:           srcID,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Error when app template has label subscriptionProviderId",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, destModel.ID).Return(destModel, nil).Once()
				repo.On("GetByID", ctx, tnt, srcModel.ID).Return(srcModel, nil).Once()
				repo.AssertNotCalled(t, "Update")
				repo.AssertNotCalled(t, "Delete")
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, srcModel.ID).Return(srcAppLabels, nil)
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, destModel.ID).Return(destAppLabels, nil)
				repo.On("ListForObject", ctx, tnt, model.AppTemplateLabelableObject, *srcModel.ApplicationTemplateID).Return(appTemplateLabelsWithSelfRegDistLabelKey, nil)

				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListFormationsForObjectGlobal", ctx, srcID).Return([]*model.Formation{}, nil).Once()
				return svc
			},
			LabelUpsertSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.AssertNotCalled(t, "UpsertMultipleLabels")
				return svc
			},
			Ctx:                ctx,
			DestinationID:      destID,
			SourceID:           srcID,
			ExpectedErrMessage: "app template: 12346789 has label subscriptionProviderId",
		},
		{
			Name: "Error when cannot get application template labels",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, destModel.ID).Return(destModel, nil).Once()
				repo.On("GetByID", ctx, tnt, srcModel.ID).Return(srcModel, nil).Once()
				repo.AssertNotCalled(t, "Update")
				repo.AssertNotCalled(t, "Delete")
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, srcModel.ID).Return(srcAppLabels, nil)
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, destModel.ID).Return(destAppLabels, nil)
				repo.On("ListForObject", ctx, tnt, model.AppTemplateLabelableObject, *srcModel.ApplicationTemplateID).Return(nil, testErr)

				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.AssertNotCalled(t, "UpsertMultipleLabels")
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListFormationsForObjectGlobal", ctx, srcID).Return([]*model.Formation{}, nil).Once()
				return svc
			},
			Ctx:                ctx,
			DestinationID:      destID,
			SourceID:           srcID,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Error while listing formations for application",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, srcModel.ID).Return(srcModel, nil).Once()
				repo.AssertNotCalled(t, "Update")
				repo.AssertNotCalled(t, "Delete")
				return repo
			},
			LabelRepoFn: UnusedLabelRepository,
			LabelUpsertSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.AssertNotCalled(t, "UpsertMultipleLabels")
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListFormationsForObjectGlobal", ctx, srcID).Return(nil, testErr).Once()
				return svc
			},
			Ctx:                ctx,
			DestinationID:      destID,
			SourceID:           srcID,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Error while listing source app labels",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, destModel.ID).Return(destModel, nil).Once()
				repo.On("GetByID", ctx, tnt, srcModel.ID).Return(srcModel, nil).Once()
				repo.AssertNotCalled(t, "Update")
				repo.AssertNotCalled(t, "Delete")
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, srcModel.ID).Return(nil, testErr)
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, destModel.ID).Return(destAppLabels, nil).Once()
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.AssertNotCalled(t, "UpsertMultipleLabels")
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListFormationsForObjectGlobal", ctx, srcID).Return([]*model.Formation{}, nil).Once()
				return svc
			},
			Ctx:                ctx,
			DestinationID:      destID,
			SourceID:           srcID,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Error while listing destination app labels",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, destModel.ID).Return(destModel, nil).Once()
				repo.On("GetByID", ctx, tnt, srcModel.ID).Return(srcModel, nil).Once()
				repo.AssertNotCalled(t, "Update")
				repo.AssertNotCalled(t, "Delete")
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.ApplicationLabelableObject, destModel.ID).Return(nil, testErr).Once()
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.AssertNotCalled(t, "UpsertMultipleLabels")
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListFormationsForObjectGlobal", ctx, srcID).Return([]*model.Formation{}, nil).Once()
				return svc
			},
			Ctx:                ctx,
			DestinationID:      destID,
			SourceID:           srcID,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appRepo := testCase.AppRepoFn()
			labelRepo := testCase.LabelRepoFn()
			labelUpserSvc := testCase.LabelUpsertSvcFn()
			fomationService := &automock.FormationService{}
			if testCase.FormationServiceFn != nil {
				fomationService = testCase.FormationServiceFn()
			}
			svc := application.NewService(nil, nil, appRepo, nil, nil, labelRepo, nil, labelUpserSvc, nil, nil, fomationService, selfRegDistLabelKey, nil)

			// WHEN
			destApp, err := svc.Merge(testCase.Ctx, testCase.DestinationID, testCase.SourceID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedDestinationApplication, destApp)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, appRepo, labelRepo, labelUpserSvc, fomationService)
		})

		srcAppLabels = fixApplicationLabels(srcID, labelKey1, labelKey2, labelValue1, "true")
		destAppLabels = fixApplicationLabels(srcID, labelKey1, labelKey2, labelValue2, "false")
	}
}

func TestService_Get(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"

	desc := "Lorem ipsum"

	applicationModel := &model.Application{
		Name:        "foo",
		Description: &desc,
		BaseEntity:  &model.BaseEntity{ID: "foo"},
	}

	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name                string
		RepositoryFn        func() *automock.ApplicationRepository
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

			svc := application.NewService(nil, nil, repo, nil, nil, nil, nil, nil, nil, nil, nil, "", nil)

			// WHEN
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

func TestService_GetGlobalByID(t *testing.T) {
	// GIVEN
	testErr := errors.New("test error")

	id := "foo"

	desc := "Lorem ipsum"

	applicationModel := &model.Application{
		Name:        "foo",
		Description: &desc,
		BaseEntity:  &model.BaseEntity{ID: "foo"},
	}

	ctx := context.TODO()

	testCases := []struct {
		Name                string
		RepositoryFn        func() *automock.ApplicationRepository
		InputID             string
		ExpectedApplication *model.Application
		ExpectedErrMessage  string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetGlobalByID", ctx, id).Return(applicationModel, nil).Once()
				return repo
			},
			InputID:             id,
			ExpectedApplication: applicationModel,
		},
		{
			Name: "Returns error when application retrieval failed",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetGlobalByID", ctx, id).Return(nil, testErr).Once()
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

			svc := application.NewService(nil, nil, repo, nil, nil, nil, nil, nil, nil, nil, nil, "", nil)

			// WHEN
			app, err := svc.GetGlobalByID(ctx, testCase.InputID)

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

func TestService_GetSystem(t *testing.T) {
	// GIVEN
	testErr := errors.New("test error")

	tnt := "id"
	locationID := "loc_id"
	virtualHost := "vhost"
	filter := labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"Host\":\"%s\",\"Subaccount\":\"%s\",\"LocationID\":\"%s\"}", virtualHost, tnt, locationID))

	desc := "Lorem ipsum"

	applicationModel := &model.Application{
		Name:        "foo",
		Description: &desc,
		BaseEntity:  &model.BaseEntity{ID: "foo"},
	}

	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name                string
		Ctx                 context.Context
		RepositoryFn        func() *automock.ApplicationRepository
		ExpectedApplication *model.Application
		ExpectedErrMessage  string
	}{
		{
			Name: "Success",
			Ctx:  ctx,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByFilter", ctx, tnt, []*labelfilter.LabelFilter{filter}).Return(applicationModel, nil).Once()
				return repo
			},
			ExpectedApplication: applicationModel,
		},
		{
			Name: "Returns error when application retrieval failed",
			Ctx:  ctx,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByFilter", ctx, tnt, []*labelfilter.LabelFilter{filter}).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:               "Returns error when extracting tenant from context",
			Ctx:                context.TODO(),
			RepositoryFn:       UnusedApplicationRepository,
			ExpectedErrMessage: "while loading tenant from context:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(nil, nil, repo, nil, nil, nil, nil, nil, nil, nil, nil, "", nil)

			// WHEN
			app, err := svc.GetSccSystem(testCase.Ctx, "id", locationID, virtualHost)

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
	// GIVEN
	testErr := errors.New("Test error")
	appID := "foo"
	appID2 := "bar"

	modelApplications := []*model.Application{
		fixModelApplication(appID, "tenant-foo", "foo", "Lorem Ipsum"),
		fixModelApplication(appID2, "tenant-bar", "bar", "Lorem Ipsum"),
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

	emptyApplicationPage := &model.ApplicationPage{
		Data:       []*model.Application{},
		TotalCount: 0,
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "",
			StartCursor: "test",
		},
	}

	first := 2
	after := "test"
	scenarios := []string{"DEFAULT"}
	filter := []*labelfilter.LabelFilter{{Key: ""}}
	scenariosFilter := []*labelfilter.LabelFilter{{Key: model.ScenariosKey, Query: stringPtr("$[*] ? (@ == \"DEFAULT\")")}}

	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		FormationServiceFn func() *automock.FormationService
		InputLabelFilters  []*labelfilter.LabelFilter
		InputPageSize      int
		ExpectedResult     *model.ApplicationPage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByIDsAndFilters", ctx, tnt, []string{}, filter, first, after).Return(applicationPage, nil).Once()
				return repo
			},
			InputPageSize:      first,
			InputLabelFilters:  filter,
			ExpectedResult:     applicationPage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success with scenario label filter - there are applications in the formation",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByIDsAndFilters", ctx, tnt, []string{appID, appID2}, []*labelfilter.LabelFilter{}, first, after).Return(applicationPage, nil).Once()
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListObjectIDsOfTypeForFormations", ctx, tnt, scenarios, model.FormationAssignmentTypeApplication).Return([]string{appID, appID2}, nil).Once()
				return svc
			},
			InputPageSize:      first,
			InputLabelFilters:  scenariosFilter,
			ExpectedResult:     applicationPage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success with scenario label filter - there are no applications in the formation",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListObjectIDsOfTypeForFormations", ctx, tnt, scenarios, model.FormationAssignmentTypeApplication).Return([]string{}, nil).Once()
				return svc
			},
			InputPageSize:      first,
			InputLabelFilters:  scenariosFilter,
			ExpectedResult:     emptyApplicationPage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when listing object IDs for formations failed",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListObjectIDsOfTypeForFormations", ctx, tnt, scenarios, model.FormationAssignmentTypeApplication).Return(nil, testErr).Once()
				return svc
			},
			InputPageSize:      first,
			InputLabelFilters:  scenariosFilter,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application listing failed",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByIDsAndFilters", ctx, tnt, []string{}, filter, first, after).Return(nil, testErr).Once()
				return repo
			},
			InputPageSize:      first,
			InputLabelFilters:  filter,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "Returns error when page size is less than 1",
			RepositoryFn: UnusedApplicationRepository,

			InputPageSize:      0,
			InputLabelFilters:  filter,
			ExpectedResult:     nil,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name:         "Returns error when page size is bigger than 200",
			RepositoryFn: UnusedApplicationRepository,

			InputPageSize:      201,
			InputLabelFilters:  filter,
			ExpectedResult:     nil,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			formationSvc := &automock.FormationService{}
			if testCase.FormationServiceFn != nil {
				formationSvc = testCase.FormationServiceFn()
			}

			svc := application.NewService(nil, nil, repo, nil, nil, nil, nil, nil, nil, nil, formationSvc, "", nil)

			// WHEN
			app, err := svc.List(ctx, testCase.InputLabelFilters, testCase.InputPageSize, after)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, app)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, repo, formationSvc)
		})
	}
}

func TestService_ListAll(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	modelApplications := []*model.Application{
		fixModelApplication("foo", "tenant-foo", "foo", "Lorem Ipsum"),
		fixModelApplication("bar", "tenant-bar", "bar", "Lorem Ipsum"),
	}
	tnt := "tenant"
	externalTnt := "external-tnt"

	ctxEmpty := context.TODO()
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		Context            context.Context
		RepositoryFn       func() *automock.ApplicationRepository
		ExpectedResult     []*model.Application
		ExpectedErrMessage string
	}{
		{
			Name:    "Success",
			Context: ctx,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, tnt).Return(modelApplications, nil).Once()
				return repo
			},

			ExpectedResult:     modelApplications,
			ExpectedErrMessage: "",
		},
		{
			Name:    "Returns error when application listing failed",
			Context: ctx,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, tnt).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:    "Returns error when tenant is not in the context",
			Context: ctxEmpty,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.AssertNotCalled(t, "ListAll")
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: "cannot read tenant from context",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			defer mock.AssertExpectationsForObjects(t, repo)

			svc := application.NewService(nil, nil, repo, nil, nil, nil, nil, nil, nil, nil, nil, "", nil)

			// WHEN
			app, err := svc.ListAll(testCase.Context)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, app)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}
		})
	}
}

func TestService_ListAllGlobalByFilter(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	appID := "foo"

	modelApplication := fixModelApplication(appID, "tenant-foo", "Foo", "Lorem Ipsum")
	modelTenant := fixBusinessTenantMappingModel("customer", tnt.Customer)
	modelApplicationsWithTenants := []*model.ApplicationWithTenants{
		{
			Application: *modelApplication,
			Tenants: []*model.BusinessTenantMapping{
				modelTenant,
			},
		},
	}
	applicationWithTenantsPage := fixApplicationWithTenantsPage(modelApplicationsWithTenants)

	testErr := errors.New("test error")
	pageSize := 2
	cursor := ""
	query := "foo"

	emptyPage := &model.ApplicationWithTenantsPage{
		Data:       []*model.ApplicationWithTenants{},
		TotalCount: 0,
		PageInfo: &pagination.Page{
			StartCursor: cursor,
			EndCursor:   "",
			HasNextPage: false,
		},
	}

	labelFilter := []*labelfilter.LabelFilter{
		{Key: "", Query: &query},
	}
	scenarios := []string{"DEFAULT"}
	scenariosFilter := []*labelfilter.LabelFilter{{Key: model.ScenariosKey, Query: stringPtr("$[*] ? (@ == \"DEFAULT\")")}}

	testCases := []struct {
		Name               string
		LabelFilter        []*labelfilter.LabelFilter
		RepositoryFn       func() *automock.ApplicationRepository
		FormationServiceFn func() *automock.FormationService
		ExpectedResult     *model.ApplicationWithTenantsPage
		ExpectedErrMessage string
	}{
		{
			Name:        "Success",
			LabelFilter: labelFilter,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAllGlobalByFilter", ctx, []string{}, labelFilter, pageSize, cursor).Return(applicationWithTenantsPage, nil).Once()
				return repo
			},
			ExpectedResult:     applicationWithTenantsPage,
			ExpectedErrMessage: "",
		},
		{
			Name:        "Success with sccenarios filter - there are no applications in formation",
			LabelFilter: scenariosFilter,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListObjectIDsOfTypeForFormationsGlobal", ctx, scenarios, model.FormationAssignmentTypeApplication).Return([]string{}, nil).Once()
				return svc
			},
			ExpectedResult:     emptyPage,
			ExpectedErrMessage: "",
		},
		{
			Name:        "Returns erorr while listing formations",
			LabelFilter: scenariosFilter,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListObjectIDsOfTypeForFormationsGlobal", ctx, scenarios, model.FormationAssignmentTypeApplication).Return(nil, testErr).Once()
				return svc
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:        "Success with sccenarios filter - there are  applications in formation",
			LabelFilter: scenariosFilter,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAllGlobalByFilter", ctx, []string{appID}, []*labelfilter.LabelFilter{}, pageSize, cursor).Return(applicationWithTenantsPage, nil).Once()
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListObjectIDsOfTypeForFormationsGlobal", ctx, scenarios, model.FormationAssignmentTypeApplication).Return([]string{appID}, nil).Once()
				return svc
			},
			ExpectedResult:     applicationWithTenantsPage,
			ExpectedErrMessage: "",
		},
		{
			Name:        "Returns error when ListAllGlobalByFilter fail",
			LabelFilter: labelFilter,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAllGlobalByFilter", ctx, []string{}, labelFilter, pageSize, cursor).Return(applicationWithTenantsPage, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			formationSvc := &automock.FormationService{}
			if testCase.FormationServiceFn != nil {
				formationSvc = testCase.FormationServiceFn()
			}

			svc := application.NewService(nil, nil, repo, nil, nil, nil, nil, nil, nil, nil, formationSvc, "", nil)

			// WHEN
			app, err := svc.ListAllGlobalByFilter(ctx, testCase.LabelFilter, pageSize, cursor)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, app)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, repo, formationSvc)
		})
	}
}

func TestService_ListByRuntimeID(t *testing.T) {
	runtimeUUID := uuid.New()
	testError := errors.New("test error")
	tenantUUID := uuid.New()
	externalTenantUUID := uuid.New()
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantUUID.String(), externalTenantUUID.String())

	first := 10
	cursor := "test"
	scenarios := []string{"Easter", "Christmas", "Winter-Sale"}

	formations := []*model.Formation{{Name: "Easter"}, {Name: "Christmas"}, {Name: "Winter-Sale"}}

	applicationIDs := []string{"test1", "test2"}
	applications := []*model.Application{
		fixModelApplication("test1", "tenant-foo", "test1", "test1"),
		fixModelApplication("test2", "tenant-foo", "test2", "test2"),
	}
	applicationPage := fixApplicationPage(applications)

	testCases := []struct {
		Name                string
		Input               uuid.UUID
		RuntimeRepositoryFn func() *automock.RuntimeRepository
		LabelRepositoryFn   func() *automock.LabelRepository
		AppRepositoryFn     func() *automock.ApplicationRepository
		FormationServiceFn  func() *automock.FormationService
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
			AppRepositoryFn: func() *automock.ApplicationRepository {
				appRepository := &automock.ApplicationRepository{}
				appRepository.On("ListByIDs", ctx, tenantUUID, applicationIDs, first, cursor).
					Return(applicationPage, nil).Once()
				return appRepository
			},
			FormationServiceFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("ListFormationsForObject", ctx, runtimeUUID.String()).Return(formations, nil).Once()
				formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantUUID.String(), scenarios, model.FormationAssignmentTypeApplication).Return(applicationIDs, nil).Once()
				return formationSvc
			},
			ExpectedError:  nil,
			ExpectedResult: applicationPage,
		},
		{
			Name:  "Success when runtime is not part of any formations",
			Input: runtimeUUID,
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				runtimeRepository := &automock.RuntimeRepository{}
				runtimeRepository.On("Exists", ctx, tenantUUID.String(), runtimeUUID.String()).
					Return(true, nil).Once()
				return runtimeRepository
			},
			FormationServiceFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("ListFormationsForObject", ctx, runtimeUUID.String()).Return(nil, nil).Once()
				formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantUUID.String(), scenarios, model.FormationAssignmentTypeApplication).Return(applicationIDs, nil).Once()
				return formationSvc
			},
			AppRepositoryFn: UnusedApplicationRepository,
			ExpectedError:   nil,
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
			AppRepositoryFn: UnusedApplicationRepository,
			ExpectedError:   testError,
			ExpectedResult:  nil,
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
			AppRepositoryFn: UnusedApplicationRepository,
			ExpectedError:   errors.New("runtime does not exist"),
			ExpectedResult:  nil,
		},
		{
			Name:  "Return error when listing formations for runtime failed",
			Input: runtimeUUID,
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				runtimeRepository := &automock.RuntimeRepository{}
				runtimeRepository.On("Exists", ctx, tenantUUID.String(), runtimeUUID.String()).
					Return(true, nil).Once()
				return runtimeRepository
			},
			FormationServiceFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("ListFormationsForObject", ctx, runtimeUUID.String()).Return(nil, testError).Once()
				return formationSvc
			},
			AppRepositoryFn: UnusedApplicationRepository,
			ExpectedError:   testError,
			ExpectedResult:  nil,
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
			FormationServiceFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("ListFormationsForObject", ctx, runtimeUUID.String()).Return(formations, nil).Once()
				formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantUUID.String(), scenarios, model.FormationAssignmentTypeApplication).Return(applicationIDs, nil).Once()
				return formationSvc
			},
			AppRepositoryFn: func() *automock.ApplicationRepository {
				appRepository := &automock.ApplicationRepository{}
				appRepository.On("ListByIDs", ctx, tenantUUID, applicationIDs, first, cursor).
					Return(nil, testError).Once()
				return appRepository
			},
			ExpectedError:  testError,
			ExpectedResult: nil,
		},
		{
			Name:  "Return error when listing object IDs for formation returns error",
			Input: runtimeUUID,
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				runtimeRepository := &automock.RuntimeRepository{}
				runtimeRepository.On("Exists", ctx, tenantUUID.String(), runtimeUUID.String()).
					Return(true, nil).Once()
				return runtimeRepository
			},
			FormationServiceFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("ListFormationsForObject", ctx, runtimeUUID.String()).Return(formations, nil).Once()
				formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantUUID.String(), scenarios, model.FormationAssignmentTypeApplication).Return(nil, testError).Once()
				return formationSvc
			},
			AppRepositoryFn: UnusedApplicationRepository,
			ExpectedError:   testError,
			ExpectedResult:  nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			runtimeRepository := testCase.RuntimeRepositoryFn()
			labelRepository := UnusedLabelRepository()
			if testCase.LabelRepositoryFn != nil {
				labelRepository = testCase.LabelRepositoryFn()
			}
			appRepository := testCase.AppRepositoryFn()
			formationSvc := UnusedFormationService()
			if testCase.FormationServiceFn != nil {
				formationSvc = testCase.FormationServiceFn()
			}
			svc := application.NewService(nil, nil, appRepository, nil, runtimeRepository, labelRepository, nil, nil, nil, nil, formationSvc, "", nil)

			// WHEN
			results, err := svc.ListByRuntimeID(ctx, testCase.Input, first, cursor)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				logrus.Info(err)
				require.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedResult, results)
			mock.AssertExpectationsForObjects(t, runtimeRepository, labelRepository, appRepository)
		})
	}
}

func TestService_ListBySCC(t *testing.T) {
	// GIVEN
	testErr := errors.New("test error")

	tnt := "tenant"

	app1ID := "foo"
	app2ID := "bar"

	app1 := fixModelApplication(app1ID, tnt, "foo", "Lorem Ipsum")
	app2 := fixModelApplication(app2ID, tnt, "bar", "Lorem Ipsum")
	applications := []*model.Application{app1, app2}

	labelValue := stringPtr("{\"locationId\":\"locationId\", \"subaccount\":\"tenant\"}")

	label1 := &model.Label{
		ObjectID: app1ID,
		Value:    labelValue,
	}

	label2 := &model.Label{
		ObjectID: app2ID,
		Value:    labelValue,
	}

	applicationsWitLabel := []*model.ApplicationWithLabel{
		{
			App:      app1,
			SccLabel: label1,
		},
		{
			App:      app2,
			SccLabel: label2,
		},
	}

	filter := &labelfilter.LabelFilter{Key: "scc", Query: stringPtr("{\"locationId\":\"locationId\", \"subaccount\":\"tenant\"}")}

	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		Ctx                context.Context
		RepositoryFn       func() *automock.ApplicationRepository
		LabelRepositoryFn  func() *automock.LabelRepository
		InputLabelFilter   *labelfilter.LabelFilter
		ExpectedResult     []*model.ApplicationWithLabel
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			Ctx:  ctx,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAllByFilter", ctx, tnt, []*labelfilter.LabelFilter{filter}).Return(applications, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListGlobalByKeyAndObjects", ctx, model.ApplicationLabelableObject, []string{app1ID, app2ID}, "scc").Return([]*model.Label{label1, label2}, nil)
				return repo
			},
			InputLabelFilter:   filter,
			ExpectedResult:     applicationsWitLabel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when no apps matching the filter are found",
			Ctx:  ctx,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAllByFilter", ctx, tnt, []*labelfilter.LabelFilter{filter}).Return([]*model.Application{}, nil).Once()
				return repo
			},
			LabelRepositoryFn:  UnusedLabelRepository,
			InputLabelFilter:   filter,
			ExpectedResult:     []*model.ApplicationWithLabel{},
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when failed to list labels for applications",
			Ctx:  ctx,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAllByFilter", ctx, tnt, []*labelfilter.LabelFilter{filter}).Return(applications, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListGlobalByKeyAndObjects", ctx, model.ApplicationLabelableObject, []string{app1ID, app2ID}, "scc").Return(nil, testErr)
				return repo
			},
			InputLabelFilter:   filter,
			ExpectedResult:     nil,
			ExpectedErrMessage: "while getting labels with key scc for applications with IDs:",
		},
		{
			Name: "Returns error when application listing failed",
			Ctx:  ctx,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAllByFilter", ctx, tnt, []*labelfilter.LabelFilter{filter}).Return(nil, testErr).Once()
				return repo
			},
			LabelRepositoryFn:  UnusedLabelRepository,
			InputLabelFilter:   filter,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "Returns error when extracting tenant from context",
			Ctx:          context.TODO(),
			RepositoryFn: UnusedApplicationRepository,

			LabelRepositoryFn:  UnusedLabelRepository,
			ExpectedErrMessage: "while loading tenant from context:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appRepo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := application.NewService(nil, nil, appRepo, nil, nil, labelRepo, nil, nil, nil, nil, nil, "", nil)

			// WHEN
			app, err := svc.ListBySCC(testCase.Ctx, filter)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, app)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			appRepo.AssertExpectations(t)
		})
	}
}

func TestService_ListSCCs(t *testing.T) {
	// GIVEN
	testErr := errors.New("test error")

	tnt := "tenant"

	key := "scc"

	locationID1 := "locationID1"
	locationID2 := "locationID2"
	subaccount1 := "subaccount1"
	subaccount2 := "subaccount2"

	labelValue1 := map[string]interface{}{"LocationID": locationID1, "Subaccount": subaccount1}
	labelValue2 := map[string]interface{}{"LocationID": locationID2, "Subaccount": subaccount2}

	labels := []*model.Label{
		{Value: labelValue1},
		{Value: labelValue2},
	}

	sccs := []*model.SccMetadata{
		{
			Subaccount: subaccount1,
			LocationID: locationID1,
		},
		{
			Subaccount: subaccount2,
			LocationID: locationID2,
		},
	}

	filter := &labelfilter.LabelFilter{Key: "scc", Query: stringPtr("{\"locationId\":\"locationId\", \"subaccount\":\"tenant\"}")}

	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		Ctx                context.Context
		RepositoryFn       func() *automock.LabelRepository
		InputLabelFilter   *labelfilter.LabelFilter
		ExpectedResult     []*model.SccMetadata
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			Ctx:  ctx,
			RepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListGlobalByKey", ctx, key).Return(labels, nil).Once()
				return repo
			},
			InputLabelFilter:   filter,
			ExpectedResult:     sccs,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when labels listing failed",
			Ctx:  ctx,
			RepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListGlobalByKey", ctx, key).Return(nil, testErr).Once()
				return repo
			},
			InputLabelFilter:   filter,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(nil, nil, nil, nil, nil, repo, nil, nil, nil, nil, nil, "", nil)

			// WHEN
			app, err := svc.ListSCCs(testCase.Ctx)

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

func TestService_Exist(t *testing.T) {
	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
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
			// GIVEN
			appRepo := testCase.RepositoryFn()
			svc := application.NewService(nil, nil, appRepo, nil, nil, nil, nil, nil, nil, nil, nil, "", nil)

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
	// GIVEN
	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testErr := errors.New("Test error")

	applicationID := "foo"

	label := &model.LabelInput{
		Key:        "key",
		Value:      []string{"value1"},
		ObjectID:   applicationID,
		ObjectType: model.ApplicationLabelableObject,
	}

	scenariosLabel := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{"value1"},
		ObjectID:   applicationID,
		ObjectType: model.ApplicationLabelableObject,
	}

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		LabelServiceFn     func() *automock.LabelService
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
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertLabel", ctx, tnt, label).Return(nil).Once()
				return svc
			},
			InputApplicationID: applicationID,
			InputLabel:         label,
			ExpectedErrMessage: "",
		},
		{
			Name: "Error when trying to set scenarios label explicitly",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(true, nil).Once()

				return repo
			},
			LabelServiceFn:     UnusedLabelUpsertService,
			InputApplicationID: applicationID,
			InputLabel:         scenariosLabel,
			ExpectedErrMessage: fmt.Sprintf("label with key %s cannot be set explicitly", model.ScenariosKey),
		},
		{
			Name: "Returns error when label set failed",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(true, nil).Once()

				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
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
			LabelServiceFn:     UnusedLabelService,
			InputApplicationID: applicationID,
			InputLabel:         label,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelSvc := testCase.LabelServiceFn()

			svc := application.NewService(nil, nil, repo, nil, nil, nil, nil, labelSvc, nil, nil, nil, "", nil)

			// WHEN
			err := svc.SetLabel(ctx, testCase.InputLabel)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, repo, labelSvc)
		})
	}
}

func TestService_GetLabel(t *testing.T) {
	// GIVEN
	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

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
		Tenant:     str.Ptr(tnt),
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
			LabelRepositoryFn:  UnusedLabelRepository,
			InputApplicationID: applicationID,
			InputLabel:         label,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := application.NewService(nil, nil, repo, nil, nil, labelRepo, nil, nil, nil, nil, nil, "", nil)

			// WHEN
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
	// GIVEN
	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

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
		Tenant:     str.Ptr(tnt),
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
			LabelRepositoryFn:  UnusedLabelRepository,
			InputApplicationID: applicationID,
			InputLabel:         label,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := application.NewService(nil, nil, repo, nil, nil, labelRepo, nil, nil, nil, nil, nil, "", nil)

			// WHEN
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

func TestService_ListLabelsGlobal(t *testing.T) {
	// GIVEN
	internalTenant := "tenant"
	ctx := context.TODO()
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
		Tenant:     str.Ptr(internalTenant),
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
				repo.On("ExistsGlobal", ctx, applicationID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForGlobalObject", ctx, model.ApplicationLabelableObject, applicationID).Return(labels, nil).Once()
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
				repo.On("ExistsGlobal", ctx, applicationID).Return(true, nil).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForGlobalObject", ctx, model.ApplicationLabelableObject, applicationID).Return(nil, testErr).Once()
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
				repo.On("ExistsGlobal", ctx, applicationID).Return(false, testErr).Once()

				return repo
			},
			LabelRepositoryFn:  UnusedLabelRepository,
			InputApplicationID: applicationID,
			InputLabel:         label,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := application.NewService(nil, nil, repo, nil, nil, labelRepo, nil, nil, nil, nil, nil, "", nil)

			// WHEN
			l, err := svc.ListLabelsGlobal(ctx, testCase.InputApplicationID)

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
	// GIVEN
	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

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
			Name: "Error when trying to delete scenarios label explicitly",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn:  UnusedLabelRepository,
			InputApplicationID: applicationID,
			InputKey:           model.ScenariosKey,
			ExpectedErrMessage: fmt.Sprintf("label with key %s cannot be deleted explicitly", model.ScenariosKey),
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
			LabelRepositoryFn:  UnusedLabelRepository,
			InputApplicationID: applicationID,
			InputKey:           labelKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application does not exist",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exists", ctx, tnt, applicationID).Return(false, nil).Once()
				return repo
			},
			LabelRepositoryFn:  UnusedLabelRepository,
			InputApplicationID: applicationID,
			InputKey:           labelKey,
			ExpectedErrMessage: fmt.Sprintf("application with ID %s doesn't exist", applicationID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := application.NewService(nil, nil, repo, nil, nil, labelRepo, nil, nil, nil, nil, nil, "", nil)

			// WHEN
			err := svc.DeleteLabel(ctx, testCase.InputApplicationID, testCase.InputKey)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}
			mock.AssertExpectationsForObjects(t, repo, labelRepo)
		})
	}
}

func TestService_GetBySystemNumber(t *testing.T) {
	tnt := "tenant"
	externalTnt := "external-tnt"

	modelApp := fixModelApplication("foo", "tenant-foo", "foo", "Lorem Ipsum")
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
	testError := errors.New("Test error")
	systemNumber := "1"

	testCases := []struct {
		Name              string
		RepositoryFn      func() *automock.ApplicationRepository
		InputSystemNumber string
		ExptectedValue    *model.Application
		ExpectedError     error
	}{
		{
			Name: "Application found",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetBySystemNumber", ctx, tnt, systemNumber).Return(modelApp, nil)
				return repo
			},
			InputSystemNumber: systemNumber,
			ExptectedValue:    modelApp,
			ExpectedError:     nil,
		},
		{
			Name: "Returns error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetBySystemNumber", ctx, tnt, systemNumber).Return(nil, testError)
				return repo
			},
			InputSystemNumber: systemNumber,
			ExptectedValue:    nil,
			ExpectedError:     testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			appRepo := testCase.RepositoryFn()
			svc := application.NewService(nil, nil, appRepo, nil, nil, nil, nil, nil, nil, nil, nil, "", nil)

			// WHEN
			value, err := svc.GetBySystemNumber(ctx, testCase.InputSystemNumber)

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

func TestService_ListByLocalTenantID(t *testing.T) {
	tnt := "tenant"
	externalTnt := "external-tnt"
	appID := "foo"

	modelApplications := fixApplicationPage([]*model.Application{fixModelApplication(appID, "tenant-foo", "foo", "Lorem Ipsum")})
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
	testError := errors.New("Test error")
	filter := labelfilter.MultipleFromGraphQL([]*graphql.LabelFilter{{Key: "key", Query: str.Ptr("query")}})
	first := 200
	cursor := "cursor"

	emptyApplicationPage := &model.ApplicationPage{
		Data:       []*model.Application{},
		TotalCount: 0,
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "",
			StartCursor: cursor,
		},
	}

	scenarios := []string{"DEFAULT"}
	scenariosFilter := []*labelfilter.LabelFilter{{Key: model.ScenariosKey, Query: stringPtr("$[*] ? (@ == \"DEFAULT\")")}}

	testCases := []struct {
		Name               string
		Ctx                context.Context
		RepositoryFn       func() *automock.ApplicationRepository
		FormationServiceFn func() *automock.FormationService
		LocalTenantID      string
		Filter             []*labelfilter.LabelFilter
		First              int
		Cursor             string
		ExptectedValue     *model.ApplicationPage
		ExpectedError      error
	}{
		{
			Name: "Getting tenant from context fails",
			Ctx:  context.TODO(),
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.AssertNotCalled(t, "ListByLocalTenantID")
				return repo
			},
			LocalTenantID:  localTenantID,
			Filter:         filter,
			First:          first,
			Cursor:         cursor,
			ExptectedValue: nil,
			ExpectedError:  errors.New("cannot read tenant from context"),
		},
		{
			Name: "Repository operation fails",
			Ctx:  ctx,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByLocalTenantID", ctx, tnt, localTenantID, []string{}, filter, first, cursor).Return(nil, testError).Once()
				return repo
			},
			LocalTenantID:  localTenantID,
			Filter:         filter,
			First:          first,
			Cursor:         cursor,
			ExptectedValue: nil,
			ExpectedError:  testError,
		},
		{
			Name: "Success",
			Ctx:  ctx,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByLocalTenantID", ctx, tnt, localTenantID, []string{}, filter, first, cursor).Return(modelApplications, nil).Once()
				return repo
			},
			LocalTenantID:  localTenantID,
			Filter:         filter,
			First:          first,
			Cursor:         cursor,
			ExptectedValue: modelApplications,
			ExpectedError:  nil,
		},
		{
			Name: "Success - there are no applications in the formation",
			Ctx:  ctx,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListObjectIDsOfTypeForFormations", ctx, tnt, scenarios, model.FormationAssignmentTypeApplication).Return([]string{}, nil).Once()
				return svc
			},
			LocalTenantID:  localTenantID,
			Filter:         scenariosFilter,
			First:          first,
			Cursor:         cursor,
			ExptectedValue: emptyApplicationPage,
			ExpectedError:  nil,
		},
		{
			Name: "Success - there are applications in the formation",
			Ctx:  ctx,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByLocalTenantID", ctx, tnt, localTenantID, []string{appID}, []*labelfilter.LabelFilter{}, first, cursor).Return(modelApplications, nil).Once()
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListObjectIDsOfTypeForFormations", ctx, tnt, scenarios, model.FormationAssignmentTypeApplication).Return([]string{appID}, nil).Once()
				return svc
			},
			LocalTenantID:  localTenantID,
			Filter:         scenariosFilter,
			First:          first,
			Cursor:         cursor,
			ExptectedValue: modelApplications,
			ExpectedError:  nil,
		},
		{
			Name: "Error when lisintg objects for formation fails",
			Ctx:  ctx,
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("ListObjectIDsOfTypeForFormations", ctx, tnt, scenarios, model.FormationAssignmentTypeApplication).Return(nil, testError).Once()
				return svc
			},
			LocalTenantID: localTenantID,
			Filter:        scenariosFilter,
			First:         first,
			Cursor:        cursor,
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			appRepo := testCase.RepositoryFn()
			formationSvc := &automock.FormationService{}
			if testCase.FormationServiceFn != nil {
				formationSvc = testCase.FormationServiceFn()
			}
			svc := application.NewService(nil, nil, appRepo, nil, nil, nil, nil, nil, nil, nil, formationSvc, "", nil)

			// WHEN
			value, err := svc.ListByLocalTenantID(testCase.Ctx, testCase.LocalTenantID, testCase.Filter, testCase.First, testCase.Cursor)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.Nil(t, err)
			}

			assert.Equal(t, testCase.ExptectedValue, value)
			mock.AssertExpectationsForObjects(t, appRepo, formationSvc)
		})
	}
}

func TestService_GetByLocalTenantIDAndAppTemplateID(t *testing.T) {
	tnt := "tenant"
	externalTnt := "external-tnt"

	modelApp := fixModelApplication("foo", "tenant-foo", "foo", "Lorem Ipsum")
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
	testError := errors.New("Test error")

	testCases := []struct {
		Name           string
		RepositoryFn   func() *automock.ApplicationRepository
		LocalTenantID  string
		AppTemplateID  string
		ExptectedValue *model.Application
		ExpectedError  error
	}{
		{
			Name: "Application found",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByLocalTenantIDAndAppTemplateID", ctx, tnt, localTenantID, appTemplateID).Return(modelApp, nil)
				return repo
			},
			LocalTenantID:  localTenantID,
			AppTemplateID:  appTemplateID,
			ExptectedValue: modelApp,
			ExpectedError:  nil,
		},
		{
			Name: "Returns error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByLocalTenantIDAndAppTemplateID", ctx, tnt, localTenantID, appTemplateID).Return(nil, testError)
				return repo
			},
			LocalTenantID:  localTenantID,
			AppTemplateID:  appTemplateID,
			ExptectedValue: nil,
			ExpectedError:  testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			appRepo := testCase.RepositoryFn()
			svc := application.NewService(nil, nil, appRepo, nil, nil, nil, nil, nil, nil, nil, nil, "", nil)

			// WHEN
			value, err := svc.GetByLocalTenantIDAndAppTemplateID(ctx, testCase.LocalTenantID, testCase.AppTemplateID)

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

type testModel struct {
	ApplicationMatcherFn func(app *model.Application) bool
	Webhooks             []*model.Webhook
	APIs                 []*model.APIDefinition
	EventAPIs            []*model.EventDefinition
	Documents            []*model.Document
}

type MatcherFunc func(app *model.Application) bool

func modelFromInput(in model.ApplicationRegisterInput, tenant, applicationID string, applicationModelMatcherFn MatcherFunc) testModel {
	webhooksModel := make([]*model.Webhook, 0, len(in.Webhooks))
	for _, item := range in.Webhooks {
		webhooksModel = append(webhooksModel, item.ToWebhook(uuid.New().String(), applicationID, model.ApplicationWebhookReference))
	}

	return testModel{
		ApplicationMatcherFn: applicationModelMatcherFn,
		Webhooks:             webhooksModel,
	}
}

func convertToStringArray(t *testing.T, array []interface{}) []string {
	stringArray := make([]string, 0, len(array))
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

func applicationFromTemplateMatcher(name string, description *string, appTemplateID *string) func(app *model.Application) bool {
	return func(app *model.Application) bool {
		return applicationMatcher(name, description)(app) && app.ApplicationTemplateID == appTemplateID
	}
}

func applicationRegisterInput() model.ApplicationRegisterInput {
	return model.ApplicationRegisterInput{
		Name:    "foo.bar-not",
		BaseURL: str.Ptr("http://test.com"),
		Webhooks: []*model.WebhookInput{
			{URL: stringPtr("test.foo.com")},
			{URL: stringPtr("test.bar.com")},
		},
		Labels: map[string]interface{}{
			"label":                "value",
			"applicationType":      "test-app-with-ppms",
			"ppmsProductVersionId": "1",
		},
		IntegrationSystemID: &intSysID,
		Bundles: []*model.BundleCreateInput{
			{
				Name: "bndl1",
				APIDefinitions: []*model.APIDefinitionInput{
					{
						Name: "foo",
					},
					{
						Name: "bar",
					},
				},
				APISpecs: []*model.SpecInput{
					{
						FetchRequest: &model.FetchRequestInput{URL: "api.foo.bar"},
					},
					nil,
				},
				EventDefinitions: []*model.EventDefinitionInput{
					{
						Name: "foo",
					},
					{
						Name: "bar",
					},
				},
				EventSpecs: []*model.SpecInput{
					{
						FetchRequest: &model.FetchRequestInput{URL: "eventapi.foo.bar"},
					},
					nil,
				},
				Documents: []*model.DocumentInput{
					{Title: "foo", Description: "test", FetchRequest: &model.FetchRequestInput{URL: "doc.foo.bar"}},
					{Title: "bar", Description: "test"},
				},
			},
		},
	}
}

func UnusedLabelService() *automock.LabelService {
	return &automock.LabelService{}
}
func UnusedBundleService() *automock.BundleService {
	return &automock.BundleService{}
}
func UnusedUIDService() *automock.UIDService {
	return &automock.UIDService{}
}

func UnusedFormationService() *automock.FormationService {
	return &automock.FormationService{}
}

func UnusedWebhookRepository() *automock.WebhookRepository {
	return &automock.WebhookRepository{}
}

func UnusedIntegrationSystemRepository() *automock.IntegrationSystemRepository {
	return &automock.IntegrationSystemRepository{}
}

func UnusedLabelRepository() *automock.LabelRepository {
	return &automock.LabelRepository{}
}

func UnusedApplicationRepository() *automock.ApplicationRepository {
	return &automock.ApplicationRepository{}
}

func UnusedLabelUpsertService() *automock.LabelService {
	return &automock.LabelService{}
}
