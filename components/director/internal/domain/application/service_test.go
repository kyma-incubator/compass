package application_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

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

func TestService_Create(t *testing.T) {
	// GIVEN
	timestamp := time.Now()
	testErr := errors.New("Test error")
	Documents := []*model.DocumentInput{
		{Title: "foo", Description: "test", FetchRequest: &model.FetchRequestInput{URL: "doc.foo.bar"}},
		{Title: "bar", Description: "test"},
	}
	modelInput := model.ApplicationRegisterInput{
		Name: "foo.bar-not",
		Webhooks: []*model.WebhookInput{
			{URL: stringPtr("test.foo.com")},
			{URL: stringPtr("test.bar.com")},
		},

		Labels: map[string]interface{}{
			"label": "value",
		},
		IntegrationSystemID: &intSysID,
	}

	bundles := []*model.BundleCreateInput{
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
			Documents: Documents,
		},
	}
	modelInput.Bundles = bundles

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
	normalizedModelInput.Bundles = bundles

	defaultLabels := map[string]interface{}{
		model.ScenariosKey:    model.ScenariosDefaultValue,
		"integrationSystemID": intSysID,
		"label":               "value",
		"name":                "mp-foo-bar-not",
	}
	defaultNormalizedLabels := map[string]interface{}{
		model.ScenariosKey:    model.ScenariosDefaultValue,
		"integrationSystemID": intSysID,
		"label":               "value",
		"name":                "mp-foo-bar-not",
	}
	defaultLabelsWithoutIntSys := map[string]interface{}{
		model.ScenariosKey:    model.ScenariosDefaultValue,
		"integrationSystemID": "",
		"name":                "mp-test",
	}
	labelsWithoutIntSys := map[string]interface{}{
		"integrationSystemID": "",
		"name":                "mp-test",
	}
	var nilLabels map[string]interface{}

	id := "foo"

	tnt := "tenant"
	externalTnt := "external-tnt"

	appModel := modelFromInput(modelInput, tnt, id, applicationMatcher(modelInput.Name, modelInput.Description))
	normalizedAppModel := modelFromInput(normalizedModelInput, tnt, id, applicationMatcher(normalizedModelInput.Name, normalizedModelInput.Description))

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	labelScenarios := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      model.ScenariosDefaultValue,
		ObjectID:   id,
		ObjectType: model.ApplicationLabelableObject,
	}

	testCases := []struct {
		Name               string
		AppNameNormalizer  normalizer.Normalizator
		AppRepoFn          func() *automock.ApplicationRepository
		WebhookRepoFn      func() *automock.WebhookRepository
		IntSysRepoFn       func() *automock.IntegrationSystemRepository
		ScenariosServiceFn func() *automock.ScenariosService
		LabelServiceFn     func() *automock.LabelUpsertService
		BundleServiceFn    func() *automock.BundleService
		UIDServiceFn       func() *automock.UIDService
		Input              model.ApplicationRegisterInput
		ExpectedErr        error
	}{
		{
			Name:              "Success",
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
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &modelInput.Labels).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*map[string]interface{})
					require.True(t, ok)
					*arg = map[string]interface{}{
						"label":            "value",
						model.ScenariosKey: model.ScenariosDefaultValue,
					}
				}).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, defaultLabels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, id, modelInput.Bundles).Return(nil).Once()
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
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &modelInput.Labels).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*map[string]interface{})
					require.True(t, ok)
					*arg = map[string]interface{}{
						"label":            "value",
						model.ScenariosKey: model.ScenariosDefaultValue,
					}
				}).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, defaultLabels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, id, modelInput.Bundles).Return(nil).Once()
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
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &modelInput.Labels).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*map[string]interface{})
					require.True(t, ok)
					*arg = map[string]interface{}{
						"label":            "value",
						model.ScenariosKey: model.ScenariosDefaultValue,
					}
				}).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, defaultLabels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, id, modelInput.Bundles).Return(nil).Once()
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
			Name:              "Returns error when listing existing applications returns application with same name",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return([]*model.Application{{Name: modelInput.Name}}, nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
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
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			Input:       modelInput,
			ExpectedErr: apperrors.NewNotUniqueNameError(resource.Application),
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
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &normalizedModelInput.Labels).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*map[string]interface{})
					require.True(t, ok)
					*arg = map[string]interface{}{
						"label":            "value",
						model.ScenariosKey: model.ScenariosDefaultValue,
					}
				}).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, defaultNormalizedLabels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, id, normalizedModelInput.Bundles).Return(nil).Once()
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
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
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
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			Input:       normalizedModelInput,
			ExpectedErr: apperrors.NewNotUniqueNameError(resource.Application),
		},
		{
			Name:              "Success when no labels provided and default scenario assignment disabled",
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
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &nilLabels).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labelsWithoutIntSys).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, labelScenarios).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       model.ApplicationRegisterInput{Name: "test", Labels: nilLabels},
			ExpectedErr: nil,
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
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &nilLabels).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*map[string]interface{})
					require.True(t, ok)
					*arg = map[string]interface{}{
						model.ScenariosKey: model.ScenariosDefaultValue,
					}
				}).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, defaultLabelsWithoutIntSys).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, labelScenarios).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
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
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &defaultLabelsWithoutIntSys).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, defaultLabelsWithoutIntSys).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, labelScenarios).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
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
			Name:              "Returns error when application creation failed",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, nil).Once()
				repo.On("Create", ctx, tnt, mock.MatchedBy(appModel.ApplicationMatcherFn)).Return(testErr).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
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
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
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
			Name:              "Returns error when listing existing applications fails",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, testErr).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
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
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name:              "Returns error when integration system doesn't exist",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
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
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			Input:       modelInput,
			ExpectedErr: errors.New("Object not found"),
		},
		{
			Name:              "Returns error when checking for integration system fails",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
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
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
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
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &modelInput.Labels).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*map[string]interface{})
					require.True(t, ok)
					*arg = map[string]interface{}{
						"label":            "value",
						model.ScenariosKey: model.ScenariosDefaultValue,
					}
				}).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, defaultLabels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, id, modelInput.Bundles).Return(testErr).Once()
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
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appNameNormalizer := testCase.AppNameNormalizer
			appRepo := testCase.AppRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			scenariosSvc := testCase.ScenariosServiceFn()
			labelSvc := testCase.LabelServiceFn()
			uidSvc := testCase.UIDServiceFn()
			intSysRepo := testCase.IntSysRepoFn()
			bndlSvc := testCase.BundleServiceFn()
			svc := application.NewService(appNameNormalizer, nil, appRepo, webhookRepo, nil, nil, intSysRepo, labelSvc, scenariosSvc, bndlSvc, uidSvc)
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

			appRepo.AssertExpectations(t)
			intSysRepo.AssertExpectations(t)
			webhookRepo.AssertExpectations(t)
			scenariosSvc.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
			bndlSvc.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		svc := application.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := svc.Create(context.TODO(), model.ApplicationRegisterInput{})
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
}

func TestService_CreateFromTemplate(t *testing.T) {
	// GIVEN
	timestamp := time.Now()
	testErr := errors.New("Test error")
	Documents := []*model.DocumentInput{
		{Title: "foo", Description: "test", FetchRequest: &model.FetchRequestInput{URL: "doc.foo.bar"}},
		{Title: "bar", Description: "test"},
	}
	modelInput := model.ApplicationRegisterInput{
		Name: "foo.bar-not",
		Webhooks: []*model.WebhookInput{
			{URL: stringPtr("test.foo.com")},
			{URL: stringPtr("test.bar.com")},
		},

		Labels: map[string]interface{}{
			"label": "value",
		},
		IntegrationSystemID: &intSysID,
	}

	bundles := []*model.BundleCreateInput{
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
			Documents: Documents,
		},
	}
	modelInput.Bundles = bundles

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
	normalizedModelInput.Bundles = bundles

	defaultLabels := map[string]interface{}{
		model.ScenariosKey:    model.ScenariosDefaultValue,
		"integrationSystemID": intSysID,
		"label":               "value",
		"name":                "mp-foo-bar-not",
	}
	defaultNormalizedLabels := map[string]interface{}{
		model.ScenariosKey:    model.ScenariosDefaultValue,
		"integrationSystemID": intSysID,
		"label":               "value",
		"name":                "mp-foo-bar-not",
	}
	defaultLabelsWithoutIntSys := map[string]interface{}{
		model.ScenariosKey:    model.ScenariosDefaultValue,
		"integrationSystemID": "",
		"name":                "mp-test",
	}
	labelsWithoutIntSys := map[string]interface{}{
		"integrationSystemID": "",
		"name":                "mp-test",
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

	labelScenarios := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      model.ScenariosDefaultValue,
		ObjectID:   id,
		ObjectType: model.ApplicationLabelableObject,
	}

	testCases := []struct {
		Name               string
		AppNameNormalizer  normalizer.Normalizator
		AppRepoFn          func() *automock.ApplicationRepository
		WebhookRepoFn      func() *automock.WebhookRepository
		IntSysRepoFn       func() *automock.IntegrationSystemRepository
		ScenariosServiceFn func() *automock.ScenariosService
		LabelServiceFn     func() *automock.LabelUpsertService
		BundleServiceFn    func() *automock.BundleService
		UIDServiceFn       func() *automock.UIDService
		Input              model.ApplicationRegisterInput
		ExpectedErr        error
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
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &modelInput.Labels).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*map[string]interface{})
					require.True(t, ok)
					*arg = map[string]interface{}{
						"label":            "value",
						model.ScenariosKey: model.ScenariosDefaultValue,
					}
				}).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, defaultLabels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, id, modelInput.Bundles).Return(nil).Once()
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
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &modelInput.Labels).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*map[string]interface{})
					require.True(t, ok)
					*arg = map[string]interface{}{
						"label":            "value",
						model.ScenariosKey: model.ScenariosDefaultValue,
					}
				}).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, defaultLabels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, id, modelInput.Bundles).Return(nil).Once()
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
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &modelInput.Labels).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*map[string]interface{})
					require.True(t, ok)
					*arg = map[string]interface{}{
						"label":            "value",
						model.ScenariosKey: model.ScenariosDefaultValue,
					}
				}).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, defaultLabels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, id, modelInput.Bundles).Return(nil).Once()
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
			Name:              "Returns error when listing existing applications returns application with same name",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return([]*model.Application{{Name: modelInput.Name}}, nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
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
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			Input:       modelInput,
			ExpectedErr: apperrors.NewNotUniqueNameError(resource.Application),
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
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &normalizedModelInput.Labels).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*map[string]interface{})
					require.True(t, ok)
					*arg = map[string]interface{}{
						"label":            "value",
						model.ScenariosKey: model.ScenariosDefaultValue,
					}
				}).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, defaultNormalizedLabels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, id, normalizedModelInput.Bundles).Return(nil).Once()
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
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
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
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			Input:       normalizedModelInput,
			ExpectedErr: apperrors.NewNotUniqueNameError(resource.Application),
		},
		{
			Name:              "Success when no labels provided and default scenario assignment disabled",
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
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &nilLabels).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labelsWithoutIntSys).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, labelScenarios).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       model.ApplicationRegisterInput{Name: "test", Labels: nilLabels},
			ExpectedErr: nil,
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
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &nilLabels).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*map[string]interface{})
					require.True(t, ok)
					*arg = map[string]interface{}{
						model.ScenariosKey: model.ScenariosDefaultValue,
					}
				}).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, defaultLabelsWithoutIntSys).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, labelScenarios).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
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
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &defaultLabelsWithoutIntSys).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, defaultLabelsWithoutIntSys).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, labelScenarios).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
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
			Name:              "Returns error when application creation failed",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, nil).Once()
				repo.On("Create", ctx, tnt, mock.MatchedBy(appFromTemplateModel.ApplicationMatcherFn)).Return(testErr).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
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
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
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
			Name:              "Returns error when listing existing applications fails",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, testErr).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
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
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name:              "Returns error when integration system doesn't exist",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
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
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			Input:       modelInput,
			ExpectedErr: errors.New("Object not found"),
		},
		{
			Name:              "Returns error when checking for integration system fails",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAll", ctx, mock.Anything).Return(nil, nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
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
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
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
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &modelInput.Labels).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*map[string]interface{})
					require.True(t, ok)
					*arg = map[string]interface{}{
						"label":            "value",
						model.ScenariosKey: model.ScenariosDefaultValue,
					}
				}).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, defaultLabels).Return(nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("CreateMultiple", ctx, id, modelInput.Bundles).Return(testErr).Once()
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
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appNameNormalizer := testCase.AppNameNormalizer
			appRepo := testCase.AppRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			scenariosSvc := testCase.ScenariosServiceFn()
			labelSvc := testCase.LabelServiceFn()
			uidSvc := testCase.UIDServiceFn()
			intSysRepo := testCase.IntSysRepoFn()
			bndlSvc := testCase.BundleServiceFn()
			svc := application.NewService(appNameNormalizer, nil, appRepo, webhookRepo, nil, nil, intSysRepo, labelSvc, scenariosSvc, bndlSvc, uidSvc)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// WHEN
			result, err := svc.CreateFromTemplate(ctx, testCase.Input, &appTemplteID)

			// then
			assert.IsType(t, "string", result)
			if testCase.ExpectedErr != nil {
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			appRepo.AssertExpectations(t)
			intSysRepo.AssertExpectations(t)
			webhookRepo.AssertExpectations(t)
			scenariosSvc.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
			bndlSvc.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		svc := application.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := svc.Create(context.TODO(), model.ApplicationRegisterInput{})
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
}

func TestService_Upsert_TrustedUpsert(t *testing.T) {
	// given
	timestamp := time.Now()
	testErr := errors.New("Test error")
	Documents := []*model.DocumentInput{
		{Title: "foo", Description: "test", FetchRequest: &model.FetchRequestInput{URL: "doc.foo.bar"}},
		{Title: "bar", Description: "test"},
	}
	modelInput := model.ApplicationRegisterInput{
		Name: "foo.bar-not",
		Webhooks: []*model.WebhookInput{
			{URL: stringPtr("test.foo.com")},
			{URL: stringPtr("test.bar.com")},
		},

		Labels: map[string]interface{}{
			"label": "value",
		},
		IntegrationSystemID: &intSysID,
	}

	bundles := []*model.BundleCreateInput{
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
			Documents: Documents,
		},
	}
	modelInput.Bundles = bundles

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
	normalizedModelInput.Bundles = bundles

	defaultLabels := map[string]interface{}{
		model.ScenariosKey:    model.ScenariosDefaultValue,
		"integrationSystemID": intSysID,
		"label":               "value",
		"name":                "mp-foo-bar-not",
	}
	defaultLabelsWithoutIntSys := map[string]interface{}{
		model.ScenariosKey:    model.ScenariosDefaultValue,
		"integrationSystemID": "",
		"name":                "mp-test",
	}
	labelsWithoutIntSys := map[string]interface{}{
		"integrationSystemID": "",
		"name":                "mp-test",
	}
	var nilLabels map[string]interface{}

	id := "foo"

	tnt := "tenant"
	externalTnt := "external-tnt"

	appModel := modelFromInput(modelInput, tnt, id, applicationMatcher(modelInput.Name, modelInput.Description))

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	labelScenarios := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      model.ScenariosDefaultValue,
		ObjectID:   id,
		ObjectType: model.ApplicationLabelableObject,
	}

	testCases := []struct {
		Name               string
		AppNameNormalizer  normalizer.Normalizator
		AppRepoFn          func(string) *automock.ApplicationRepository
		IntSysRepoFn       func() *automock.IntegrationSystemRepository
		ScenariosServiceFn func() *automock.ScenariosService
		LabelServiceFn     func() *automock.LabelUpsertService
		UIDServiceFn       func() *automock.UIDService
		Input              model.ApplicationRegisterInput
		ExpectedErr        error
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
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &modelInput.Labels).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*map[string]interface{})
					require.True(t, ok)
					*arg = map[string]interface{}{
						"label":            "value",
						model.ScenariosKey: model.ScenariosDefaultValue,
					}
				}).Once()
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
			Name:              "Success when no labels provided and default scenario assignment disabled",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func(upsertMethodName string) *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On(upsertMethodName, ctx, mock.Anything, mock.MatchedBy(applicationMatcher("test", nil))).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &nilLabels).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labelsWithoutIntSys).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, labelScenarios).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       model.ApplicationRegisterInput{Name: "test", Labels: nilLabels},
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
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &nilLabels).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*map[string]interface{})
					require.True(t, ok)
					*arg = map[string]interface{}{
						model.ScenariosKey: model.ScenariosDefaultValue,
					}
				}).Once()
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
			Input:       model.ApplicationRegisterInput{Name: "test", Labels: nilLabels},
			ExpectedErr: nil,
		},
		{
			Name:              "Success when scenarios label provided",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func(upsertMethodName string) *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On(upsertMethodName, ctx, mock.Anything, mock.MatchedBy(applicationMatcher("test", nil))).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &defaultLabelsWithoutIntSys).Once()
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
			Name:              "Returns error when application creation failed",
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
		t.Run(testCase.Name+"_Upsert", func(t *testing.T) {
			appNameNormalizer := testCase.AppNameNormalizer
			appRepo := testCase.AppRepoFn("Upsert")
			scenariosSvc := testCase.ScenariosServiceFn()
			labelSvc := testCase.LabelServiceFn()
			uidSvc := testCase.UIDServiceFn()
			intSysRepo := testCase.IntSysRepoFn()
			svc := application.NewService(appNameNormalizer, nil, appRepo, nil, nil, nil, intSysRepo, labelSvc, scenariosSvc, nil, uidSvc)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// when
			err := svc.Upsert(ctx, testCase.Input)

			// then
			if testCase.ExpectedErr != nil {
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			appRepo.AssertExpectations(t)
			intSysRepo.AssertExpectations(t)
			scenariosSvc.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})

		t.Run(testCase.Name+"_TrustedUpsert", func(t *testing.T) {
			appNameNormalizer := testCase.AppNameNormalizer
			appRepo := testCase.AppRepoFn("TrustedUpsert")
			scenariosSvc := testCase.ScenariosServiceFn()
			labelSvc := testCase.LabelServiceFn()
			uidSvc := testCase.UIDServiceFn()
			intSysRepo := testCase.IntSysRepoFn()
			svc := application.NewService(appNameNormalizer, nil, appRepo, nil, nil, nil, intSysRepo, labelSvc, scenariosSvc, nil, uidSvc)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// when
			err := svc.TrustedUpsert(ctx, testCase.Input)

			// then
			if testCase.ExpectedErr != nil {
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			appRepo.AssertExpectations(t)
			intSysRepo.AssertExpectations(t)
			scenariosSvc.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		svc := application.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// when
		_, err := svc.Create(context.TODO(), model.ApplicationRegisterInput{})
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
}

func TestService_TrustedUpsertFromTemplate(t *testing.T) {
	// given
	timestamp := time.Now()
	testErr := errors.New("Test error")
	Documents := []*model.DocumentInput{
		{Title: "foo", Description: "test", FetchRequest: &model.FetchRequestInput{URL: "doc.foo.bar"}},
		{Title: "bar", Description: "test"},
	}
	modelInput := model.ApplicationRegisterInput{
		Name: "foo.bar-not",
		Webhooks: []*model.WebhookInput{
			{URL: stringPtr("test.foo.com")},
			{URL: stringPtr("test.bar.com")},
		},

		Labels: map[string]interface{}{
			"label": "value",
		},
		IntegrationSystemID: &intSysID,
	}

	bundles := []*model.BundleCreateInput{
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
			Documents: Documents,
		},
	}
	modelInput.Bundles = bundles

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
	normalizedModelInput.Bundles = bundles

	defaultLabels := map[string]interface{}{
		model.ScenariosKey:    model.ScenariosDefaultValue,
		"integrationSystemID": intSysID,
		"label":               "value",
		"name":                "mp-foo-bar-not",
	}
	defaultLabelsWithoutIntSys := map[string]interface{}{
		model.ScenariosKey:    model.ScenariosDefaultValue,
		"integrationSystemID": "",
		"name":                "mp-test",
	}
	labelsWithoutIntSys := map[string]interface{}{
		"integrationSystemID": "",
		"name":                "mp-test",
	}
	var nilLabels map[string]interface{}

	id := "foo"

	tnt := "tenant"
	externalTnt := "external-tnt"
	appTemplteID := "test-app-template"

	appFromTemplateModel := modelFromInput(modelInput, tnt, id, applicationFromTemplateMatcher(modelInput.Name, modelInput.Description, &appTemplteID))

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	labelScenarios := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      model.ScenariosDefaultValue,
		ObjectID:   id,
		ObjectType: model.ApplicationLabelableObject,
	}

	testCases := []struct {
		Name               string
		AppNameNormalizer  normalizer.Normalizator
		AppRepoFn          func() *automock.ApplicationRepository
		IntSysRepoFn       func() *automock.IntegrationSystemRepository
		ScenariosServiceFn func() *automock.ScenariosService
		LabelServiceFn     func() *automock.LabelUpsertService
		UIDServiceFn       func() *automock.UIDService
		Input              model.ApplicationRegisterInput
		ExpectedErr        error
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
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &modelInput.Labels).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*map[string]interface{})
					require.True(t, ok)
					*arg = map[string]interface{}{
						"label":            "value",
						model.ScenariosKey: model.ScenariosDefaultValue,
					}
				}).Once()
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
			Name:              "Success when no labels provided and default scenario assignment disabled",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("TrustedUpsert", ctx, mock.Anything, mock.MatchedBy(applicationMatcher("test", nil))).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &nilLabels).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertMultipleLabels", ctx, tnt, model.ApplicationLabelableObject, id, labelsWithoutIntSys).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, labelScenarios).Return(nil).Once()
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       model.ApplicationRegisterInput{Name: "test", Labels: nilLabels},
			ExpectedErr: nil,
		},
		{
			Name:              "Success when no labels provided",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("TrustedUpsert", ctx, mock.Anything, mock.MatchedBy(applicationMatcher("test", nil))).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &nilLabels).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*map[string]interface{})
					require.True(t, ok)
					*arg = map[string]interface{}{
						model.ScenariosKey: model.ScenariosDefaultValue,
					}
				}).Once()
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
			Input:       model.ApplicationRegisterInput{Name: "test", Labels: nilLabels},
			ExpectedErr: nil,
		},
		{
			Name:              "Success when scenarios label provided",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("TrustedUpsert", ctx, mock.Anything, mock.MatchedBy(applicationMatcher("test", nil))).Return("foo", nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &defaultLabelsWithoutIntSys).Once()
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
			Name:              "Returns error when application creation failed",
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
			ExpectedErr: errors.New("Object not found"),
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
			appNameNormalizer := testCase.AppNameNormalizer
			appRepo := testCase.AppRepoFn()
			scenariosSvc := testCase.ScenariosServiceFn()
			labelSvc := testCase.LabelServiceFn()
			uidSvc := testCase.UIDServiceFn()
			intSysRepo := testCase.IntSysRepoFn()
			svc := application.NewService(appNameNormalizer, nil, appRepo, nil, nil, nil, intSysRepo, labelSvc, scenariosSvc, nil, uidSvc)
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

			appRepo.AssertExpectations(t)
			intSysRepo.AssertExpectations(t)
			scenariosSvc.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		svc := application.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
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

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	resetModels := func() {
		appName := "initialn"
		initialDescrription := "initald"
		initialURL := "initialu"
		updatedDescription := "updatedd"
		updatedHealthCheckURL := "updatedhcu"
		updatedBaseURL := "updatedbu"
		updateInput = fixModelApplicationUpdateInput(appName, updatedDescription, updatedHealthCheckURL, updatedBaseURL, model.ApplicationStatusConditionConnected)
		applicationModelBefore = fixModelApplicationWithAllUpdatableFields(id, appName, initialDescrription, initialURL, nil, model.ApplicationStatusConditionConnected, conditionTimestamp)
		applicationModelAfter = fixModelApplicationWithAllUpdatableFields(id, appName, updatedDescription, updatedHealthCheckURL, &updatedBaseURL, model.ApplicationStatusConditionConnected, conditionTimestamp)
		intSysLabel = fixLabelInput("integrationSystemID", intSysID, id, model.ApplicationLabelableObject)
		nameLabel = fixLabelInput("name", "mp-"+appName, id, model.ApplicationLabelableObject)
		updateInputStatusOnly = fixModelApplicationUpdateInputStatus(model.ApplicationStatusConditionConnected)
		applicationModelAfterStatusUpdate = fixModelApplicationWithAllUpdatableFields(id, appName, initialDescrription, initialURL, nil, model.ApplicationStatusConditionConnected, conditionTimestamp)
	}

	resetModels()

	testCases := []struct {
		Name               string
		AppNameNormalizer  normalizer.Normalizator
		AppRepoFn          func() *automock.ApplicationRepository
		IntSysRepoFn       func() *automock.IntegrationSystemRepository
		LabelUpsertSvcFn   func() *automock.LabelUpsertService
		Input              model.ApplicationUpdateInput
		InputID            string
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
			Name:              "Success Status Condition Update",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(applicationModelBefore, nil).Once()
				repo.On("Update", ctx, tnt, applicationModelAfterStatusUpdate).Return(nil).Once()
				repo.On("Exists", ctx, tnt, id).Return(true, nil).Once()
				return repo
			},
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				repo := &automock.IntegrationSystemRepository{}
				return repo
			},
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertLabel", ctx, tnt, nameLabel).Return(nil).Once()
				return svc
			},
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
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
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
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			InputID:            "foo",
			Input:              updateInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:              "Returns error when Integration System does not exist",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
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
			ExpectedErrMessage: errors.New("Object not found").Error(),
		},
		{
			Name:              "Returns error ensuring Integration System existence failed",
			AppNameNormalizer: &normalizer.DefaultNormalizator{},
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
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
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
			resetModels()
			appNameNormalizer := testCase.AppNameNormalizer
			appRepo := testCase.AppRepoFn()
			intSysRepo := testCase.IntSysRepoFn()
			lblUpsrtSvc := testCase.LabelUpsertSvcFn()
			svc := application.NewService(appNameNormalizer, nil, appRepo, nil, nil, nil, intSysRepo, lblUpsrtSvc, nil, nil, nil)
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

			appRepo.AssertExpectations(t)
			intSysRepo.AssertExpectations(t)
			lblUpsrtSvc.AssertExpectations(t)
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
		applicationModelBefore = fixModelApplicationWithAllUpdatableFields(id, appName, description, url, nil, model.ApplicationStatusConditionConnected, conditionTimestamp)
		applicationModelAfter = fixModelApplicationWithAllUpdatableFields(id, appName, description, url, &updatedBaseURL, model.ApplicationStatusConditionConnected, conditionTimestamp)
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
			svc := application.NewService(nil, nil, appRepo, nil, nil, nil, nil, nil, nil, nil, nil)

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

func TestService_Delete(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	formationAndRuntimeError := errors.New("The operation is not allowed [reason=System foo is still used and cannot be deleted. Unassign the system from the following formations first: Easter. Then, unassign the system from the following runtimes, too: test-runtime]")
	id := "foo"
	desc := "Lorem ipsum"
	tnt := "tenant"
	externalTnt := "external-tnt"

	scenarios := []interface{}{"Easter"}
	scenarioLabel := &model.Label{
		ID:    uuid.New().String(),
		Key:   model.ScenariosKey,
		Value: scenarios,
	}

	emptyScenarioLabel := &model.Label{
		ID:    uuid.New().String(),
		Key:   model.ScenariosKey,
		Value: []interface{}{},
	}

	applicationModel := &model.Application{
		Name:        "foo",
		Description: &desc,
		BaseEntity:  &model.BaseEntity{ID: id},
	}

	runtimeModel := &model.Runtime{
		Name: "test-runtime",
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		AppRepoFn          func() *automock.ApplicationRepository
		LabelRepoFn        func() *automock.LabelRepository
		RuntimeRepoFn      func() *automock.RuntimeRepository
		Input              model.ApplicationRegisterInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Delete", ctx, tnt, applicationModel.ID).Return(nil).Once()
				repo.On("Exists", ctx, tnt, applicationModel.ID).Return(true, nil).Once()
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, applicationModel.ID, model.ScenariosKey).Return(emptyScenarioLabel, nil)
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListAll", ctx, tnt, mock.Anything).Return(scenarioLabel).Return([]*model.Runtime{}, nil)
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when application is part of a scenario but not with runtime",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Delete", ctx, tnt, applicationModel.ID).Return(nil).Once()
				repo.AssertNotCalled(t, "Delete")
				repo.On("Exists", ctx, tnt, applicationModel.ID).Return(true, nil).Once()
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, applicationModel.ID, model.ScenariosKey).Return(scenarioLabel, nil)
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListAll", ctx, tnt, mock.Anything).Return(scenarioLabel).Return([]*model.Runtime{}, nil)
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when application deletion failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Delete", ctx, tnt, applicationModel.ID).Return(testErr).Once()
				repo.On("Exists", ctx, tnt, applicationModel.ID).Return(true, nil).Once()
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, applicationModel.ID, model.ScenariosKey).Return(emptyScenarioLabel, nil)
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListAll", ctx, tnt, mock.Anything).Return(scenarioLabel).Return([]*model.Runtime{}, nil)
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application is part of a scenario with runtime",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.AssertNotCalled(t, "Delete")
				repo.On("Exists", ctx, tnt, applicationModel.ID).Return(true, nil).Once()
				repo.On("GetByID", ctx, tnt, applicationModel.ID).Return(applicationModel, nil).Once()
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, applicationModel.ID, model.ScenariosKey).Return(scenarioLabel, nil)
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListAll", ctx, tnt, mock.Anything).Return(scenarioLabel).Return([]*model.Runtime{runtimeModel}, nil)
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: formationAndRuntimeError.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appRepo := testCase.AppRepoFn()
			labelRepo := testCase.LabelRepoFn()
			runtimeRepo := testCase.RuntimeRepoFn()
			svc := application.NewService(nil, nil, appRepo, nil, runtimeRepo, labelRepo, nil, nil, nil, nil, nil)

			// WHEN
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

func TestService_Unpair(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	formationAndRuntimeError := errors.New("The operation is not allowed [reason=System foo is still used and cannot be deleted. Unassign the system from the following formations first: Easter. Then, unassign the system from the following runtimes, too: test-runtime]")
	id := "foo"
	desc := "Lorem ipsum"
	tnt := "tenant"
	externalTnt := "external-tnt"

	scenarios := []interface{}{"Easter"}
	scenarioLabel := &model.Label{
		ID:    uuid.New().String(),
		Key:   model.ScenariosKey,
		Value: scenarios,
	}

	emptyScenarioLabel := &model.Label{
		ID:    uuid.New().String(),
		Key:   model.ScenariosKey,
		Value: []interface{}{},
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
		LabelRepoFn        func() *automock.LabelRepository
		RuntimeRepoFn      func() *automock.RuntimeRepository
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
				repo.On("Exists", ctx, tnt, applicationModelWithInitialStatus.ID).Return(true, nil).Once()
				repo.On("GetByID", ctx, tnt, applicationModelWithInitialStatus.ID).Return(applicationModelWithInitialStatus, nil).Once()
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, applicationModel.ID, model.ScenariosKey).Return(emptyScenarioLabel, nil)
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.AssertNotCalled(t, "ListAll")
				return repo
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
				repo.On("Exists", ctx, tnt, applicationModel.ID).Return(true, nil).Once()
				repo.On("GetByID", ctx, tnt, applicationModel.ID).Return(applicationModel, nil).Once()
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, applicationModel.ID, model.ScenariosKey).Return(scenarioLabel, nil)
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListAll", ctx, tnt, mock.Anything).Return(scenarioLabel).Return([]*model.Runtime{}, nil)
				return repo
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
				repo.On("Exists", ctx, tnt, applicationModel.ID).Return(true, nil).Once()
				repo.On("GetByID", ctx, tnt, applicationModel.ID).Return(applicationModel, nil).Once()
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, applicationModel.ID, model.ScenariosKey).Return(scenarioLabel, nil)
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListAll", ctx, tnt, mock.Anything).Return(scenarioLabel).Return([]*model.Runtime{}, nil)
				return repo
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
				repo.On("Exists", mock.Anything, tnt, applicationModel.ID).Return(true, nil).Once()
				repo.On("GetByID", mock.Anything, tnt, applicationModel.ID).Return(applicationModel, nil).Once()
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", mock.Anything, tnt, model.ApplicationLabelableObject, applicationModel.ID, model.ScenariosKey).Return(scenarioLabel, nil)
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListAll", mock.Anything, tnt, mock.Anything).Return(scenarioLabel).Return([]*model.Runtime{}, nil)
				return repo
			},
			InputID: id,
			ContextFn: func() context.Context {
				backgroundCtx := context.Background()
				backgroundCtx = operation.SaveModeToContext(backgroundCtx, graphql.OperationModeAsync)
				return backgroundCtx
			},
		},
		{
			Name: "Returns error when application fetch failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, tnt, applicationModel.ID).Return(nil, testErr).Once()
				repo.On("Exists", ctx, tnt, applicationModel.ID).Return(true, nil).Once()
				repo.AssertNotCalled(t, "Update")
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, applicationModel.ID, model.ScenariosKey).Return(emptyScenarioLabel, nil)
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.AssertNotCalled(t, "ListAll")
				return repo
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
				repo.On("Exists", ctx, tnt, applicationModel.ID).Return(true, nil).Once()
				repo.On("GetByID", ctx, tnt, applicationModel.ID).Return(applicationModel, nil).Once()
				repo.AssertNotCalled(t, "Update")
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, applicationModel.ID, model.ScenariosKey).Return(scenarioLabel, nil)
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListAll", ctx, tnt, mock.Anything).Return(scenarioLabel).Return([]*model.Runtime{runtimeModel}, nil)
				return repo
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
				repo.On("Exists", ctx, tnt, applicationModel.ID).Return(true, nil).Once()
				repo.On("GetByID", ctx, tnt, applicationModel.ID).Return(applicationModel, nil).Once()
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, applicationModel.ID, model.ScenariosKey).Return(emptyScenarioLabel, nil)
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.AssertNotCalled(t, "ListAll")
				return repo
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
			labelRepo := testCase.LabelRepoFn()
			runtimeRepo := testCase.RuntimeRepoFn()
			ctx := testCase.ContextFn()
			ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
			svc := application.NewService(nil, nil, appRepo, nil, runtimeRepo, labelRepo, nil, nil, nil, nil, nil)
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

			appRepo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
			runtimeRepo.AssertExpectations(t)
		})
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

			svc := application.NewService(nil, nil, repo, nil, nil, nil, nil, nil, nil, nil, nil)

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
			Name: "Returns error when extracting tenant from context",
			Ctx:  context.TODO(),
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			ExpectedErrMessage: "while loading tenant from context:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(nil, nil, repo, nil, nil, nil, nil, nil, nil, nil, nil)

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
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

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
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Returns error when page size is bigger than 200",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			InputPageSize:      201,
			InputLabelFilters:  filter,
			ExpectedResult:     nil,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(nil, nil, repo, nil, nil, nil, nil, nil, nil, nil, nil)

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

			repo.AssertExpectations(t)
		})
	}
}

func TestService_ListGlobal(t *testing.T) {
	// GIVEN
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
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

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
				repo.On("ListGlobal", ctx, first, after).Return(applicationPage, nil).Once()
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
				repo.On("ListGlobal", ctx, first, after).Return(nil, testErr).Once()
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
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Returns error when page size is bigger than 200",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			InputPageSize:      201,
			InputLabelFilters:  filter,
			ExpectedResult:     nil,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(nil, nil, repo, nil, nil, nil, nil, nil, nil, nil, nil)

			// WHEN
			app, err := svc.ListGlobal(ctx, testCase.InputPageSize, after)

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
	externalTenantUUID := uuid.New()
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantUUID.String(), externalTenantUUID.String())

	first := 10
	cursor := "test"
	scenarios := []interface{}{"Easter", "Christmas", "Winter-Sale"}
	scenarioLabel := model.Label{
		ID:    uuid.New().String(),
		Key:   model.ScenariosKey,
		Value: scenarios,
	}
	hidingSelectors := map[string][]string{"foo": {"bar", "baz"}}

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
		ConfigProviderFn    func() *automock.ApplicationHideCfgProvider
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
				appRepository.On("ListByScenarios", ctx, tenantUUID, convertToStringArray(t, scenarios), first, cursor, hidingSelectors).
					Return(applicationPage, nil).Once()
				return appRepository
			},
			ConfigProviderFn: func() *automock.ApplicationHideCfgProvider {
				cfgProvider := &automock.ApplicationHideCfgProvider{}
				cfgProvider.On("GetApplicationHideSelectors").Return(hidingSelectors, nil).Once()
				return cfgProvider
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
					Return(nil, apperrors.NewNotFoundError(resource.Application, "")).Once()
				return labelRepository
			},
			AppRepositoryFn: func() *automock.ApplicationRepository {
				appRepository := &automock.ApplicationRepository{}
				return appRepository
			},
			ConfigProviderFn: func() *automock.ApplicationHideCfgProvider {
				cfgProvider := &automock.ApplicationHideCfgProvider{}
				return cfgProvider
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
			ConfigProviderFn: func() *automock.ApplicationHideCfgProvider {
				cfgProvider := &automock.ApplicationHideCfgProvider{}
				return cfgProvider
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
			ConfigProviderFn: func() *automock.ApplicationHideCfgProvider {
				cfgProvider := &automock.ApplicationHideCfgProvider{}
				return cfgProvider
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
			ConfigProviderFn: func() *automock.ApplicationHideCfgProvider {
				cfgProvider := &automock.ApplicationHideCfgProvider{}
				return cfgProvider
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
				appRepository.On("ListByScenarios", ctx, tenantUUID, convertToStringArray(t, scenarios), first, cursor, hidingSelectors).
					Return(nil, testError).Once()
				return appRepository
			},
			ConfigProviderFn: func() *automock.ApplicationHideCfgProvider {
				cfgProvider := &automock.ApplicationHideCfgProvider{}
				cfgProvider.On("GetApplicationHideSelectors").Return(hidingSelectors, nil).Once()
				return cfgProvider
			},
			ExpectedError:  testError,
			ExpectedResult: nil,
		},
		{
			Name:  "Return empty page when runtime is not assigned to any scenario",
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
			ConfigProviderFn: func() *automock.ApplicationHideCfgProvider {
				cfgProvider := &automock.ApplicationHideCfgProvider{}
				return cfgProvider
			},
			ExpectedError:  nil,
			ExpectedResult: &emptyPage,
		},
		{
			Name:  "Return error when config provider returns error",
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
				return appRepository
			},
			ConfigProviderFn: func() *automock.ApplicationHideCfgProvider {
				cfgProvider := &automock.ApplicationHideCfgProvider{}
				cfgProvider.On("GetApplicationHideSelectors").Return(nil, testError).Once()
				return cfgProvider
			},
			ExpectedError:  testError,
			ExpectedResult: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			runtimeRepository := testCase.RuntimeRepositoryFn()
			labelRepository := testCase.LabelRepositoryFn()
			appRepository := testCase.AppRepositoryFn()
			cfgProvider := testCase.ConfigProviderFn()
			svc := application.NewService(nil, cfgProvider, appRepository, nil, runtimeRepository, labelRepository, nil, nil, nil, nil, nil)

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
			runtimeRepository.AssertExpectations(t)
			labelRepository.AssertExpectations(t)
			appRepository.AssertExpectations(t)
			cfgProvider.AssertExpectations(t)
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
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
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
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			InputLabelFilter:   filter,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when extracting tenant from context",
			Ctx:  context.TODO(),
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			ExpectedErrMessage: "while loading tenant from context:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appRepo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := application.NewService(nil, nil, appRepo, nil, nil, labelRepo, nil, nil, nil, nil, nil)

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

			svc := application.NewService(nil, nil, nil, nil, nil, repo, nil, nil, nil, nil, nil)

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
			svc := application.NewService(nil, nil, appRepo, nil, nil, nil, nil, nil, nil, nil, nil)

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
			svc := application.NewService(nil, nil, repo, nil, nil, nil, nil, labelSvc, nil, nil, nil)

			// WHEN
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
			svc := application.NewService(nil, nil, repo, nil, nil, labelRepo, nil, nil, nil, nil, nil)

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
			svc := application.NewService(nil, nil, repo, nil, nil, labelRepo, nil, nil, nil, nil, nil)

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
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := application.NewService(nil, nil, repo, nil, nil, labelRepo, nil, nil, nil, nil, nil)

			// WHEN
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

func TestService_GetByNameAndSystemNumber(t *testing.T) {
	tnt := "tenant"
	externalTnt := "external-tnt"

	modelApp := fixModelApplication("foo", "tenant-foo", "foo", "Lorem Ipsum")
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
	testError := errors.New("Test error")

	applicationName := "name"
	systemNumber := "1"

	testCases := []struct {
		Name                 string
		RepositoryFn         func() *automock.ApplicationRepository
		InputApplicationName string
		InputSystemNumber    string
		ExptectedValue       *model.Application
		ExpectedError        error
	}{
		{
			Name: "Application found",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByNameAndSystemNumber", ctx, tnt, applicationName, systemNumber).Return(modelApp, nil)
				return repo
			},
			InputApplicationName: applicationName,
			InputSystemNumber:    systemNumber,
			ExptectedValue:       modelApp,
			ExpectedError:        nil,
		},
		{
			Name: "Returns error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByNameAndSystemNumber", ctx, tnt, applicationName, systemNumber).Return(nil, testError)
				return repo
			},
			InputApplicationName: applicationName,
			InputSystemNumber:    systemNumber,
			ExptectedValue:       nil,
			ExpectedError:        testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			appRepo := testCase.RepositoryFn()
			svc := application.NewService(nil, nil, appRepo, nil, nil, nil, nil, nil, nil, nil, nil)

			// WHEN
			value, err := svc.GetByNameAndSystemNumber(ctx, testCase.InputApplicationName, testCase.InputSystemNumber)

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
