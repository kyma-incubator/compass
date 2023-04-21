package datainputbuilder_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	databuilder "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder/automock"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	emptyCtx         = context.Background()
	testErr          = errors.New("test error")
	testTenantID     = "testTenantID"
	testRuntimeID    = "testRuntimeID"
	testRuntimeCtxID = "testRuntimeCtxID"

	testLabels = map[string]*model.Label{"testLabelKey": {
		ID:     "testLabelID",
		Tenant: &testTenantID,
		Value:  "testLabelValue",
	}}

	testLabelsComposite = map[string]*model.Label{"testLabelKey": {
		ID:     "testLabelID",
		Tenant: &testTenantID,
		Value:  []string{"testLabelValue"},
	}}

	testRuntime = &model.Runtime{
		ID:   testRuntimeID,
		Name: "testRuntimeName",
	}

	testExpectedRuntimeWithLabels = &webhook.RuntimeWithLabels{
		Runtime: testRuntime,
		Labels:  convertLabels(testLabels),
	}

	testRuntimeCtx = &model.RuntimeContext{
		ID:    testRuntimeCtxID,
		Key:   "testRtmCtxKey",
		Value: "testRtmCtxValue",
	}

	testExpectedRuntimeCtxWithLabels = &webhook.RuntimeContextWithLabels{
		RuntimeContext: testRuntimeCtx,
		Labels:         convertLabels(testLabels),
	}
)

func TestWebhookDataInputBuilder_PrepareApplicationAndAppTemplateWithLabels(t *testing.T) {
	testAppID := "testAppID"
	testAppTemplateID := "testAppTemplateID"

	testApplication := &model.Application{
		Name:                  "testAppName",
		ApplicationTemplateID: &testAppTemplateID,
	}

	testAppTemplate := &model.ApplicationTemplate{
		ID:   testAppTemplateID,
		Name: "testAppTemplateName",
	}

	testExpectedAppWithLabels := &webhook.ApplicationWithLabels{
		Application: testApplication,
		Labels:      convertLabels(testLabels),
	}

	testExpectedAppWithCompositeLabel := &webhook.ApplicationWithLabels{
		Application: testApplication,
		Labels:      convertLabels(testLabelsComposite),
	}

	testExpectedAppTemplateWithLabels := &webhook.ApplicationTemplateWithLabels{
		ApplicationTemplate: testAppTemplate,
		Labels:              convertLabels(testLabels),
	}

	testCases := []struct {
		name                          string
		appRepo                       func() *automock.ApplicationRepository
		appTemplateRepo               func() *automock.ApplicationTemplateRepository
		runtimeRepo                   func() *automock.RuntimeRepository
		runtimeCtxRepo                func() *automock.RuntimeContextRepository
		labelRepo                     func() *automock.LabelRepository
		expectedAppWithLabels         *webhook.ApplicationWithLabels
		expectedAppTemplateWithLabels *webhook.ApplicationTemplateWithLabels
		expectedErrMsg                string
	}{
		{
			name: "Success",
			appRepo: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testAppID).Return(testApplication, nil).Once()
				return appRepo
			},
			appTemplateRepo: func() *automock.ApplicationTemplateRepository {
				appTmplRepo := &automock.ApplicationTemplateRepository{}
				appTmplRepo.On("Get", emptyCtx, testAppTemplateID).Return(testAppTemplate, nil).Once()
				return appTmplRepo
			},
			runtimeRepo:    unusedRuntimeRepo,
			runtimeCtxRepo: unusedRuntimeCtxRepo,
			labelRepo: func() *automock.LabelRepository {
				lblRepo := &automock.LabelRepository{}
				lblRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testAppID).Return(testLabels, nil).Once()
				lblRepo.On("ListForObject", emptyCtx, testTenantID, model.AppTemplateLabelableObject, testAppTemplateID).Return(testLabels, nil).Once()
				return lblRepo
			},
			expectedAppWithLabels:         testExpectedAppWithLabels,
			expectedAppTemplateWithLabels: testExpectedAppTemplateWithLabels,
		},
		{
			name: "Success when fails to unquote label",
			appRepo: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testAppID).Return(testApplication, nil).Once()
				return appRepo
			},
			appTemplateRepo: func() *automock.ApplicationTemplateRepository {
				appTmplRepo := &automock.ApplicationTemplateRepository{}
				appTmplRepo.On("Get", emptyCtx, testAppTemplateID).Return(testAppTemplate, nil).Once()
				return appTmplRepo
			},
			runtimeRepo:    unusedRuntimeRepo,
			runtimeCtxRepo: unusedRuntimeCtxRepo,
			labelRepo: func() *automock.LabelRepository {
				lblRepo := &automock.LabelRepository{}
				lblRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testAppID).Return(testLabelsComposite, nil).Once()
				lblRepo.On("ListForObject", emptyCtx, testTenantID, model.AppTemplateLabelableObject, testAppTemplateID).Return(testLabels, nil).Once()
				return lblRepo
			},
			expectedAppWithLabels:         testExpectedAppWithCompositeLabel,
			expectedAppTemplateWithLabels: testExpectedAppTemplateWithLabels,
		},
		{
			name: "Error when getting application fail",
			appRepo: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testAppID).Return(nil, testErr).Once()
				return appRepo
			},
			appTemplateRepo: unusedAppTemplateRepo,
			runtimeRepo:     unusedRuntimeRepo,
			runtimeCtxRepo:  unusedRuntimeCtxRepo,
			labelRepo:       unusedLabelRepo,
			expectedErrMsg:  fmt.Sprintf("while getting application by ID: %q", testAppID),
		},
		{
			name: "Error when getting application labels fail",
			appRepo: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testAppID).Return(testApplication, nil).Once()
				return appRepo
			},
			appTemplateRepo: unusedAppTemplateRepo,
			runtimeRepo:     unusedRuntimeRepo,
			runtimeCtxRepo:  unusedRuntimeCtxRepo,
			labelRepo: func() *automock.LabelRepository {
				lblRepo := &automock.LabelRepository{}
				lblRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testAppID).Return(nil, testErr).Once()
				return lblRepo
			},
			expectedErrMsg: fmt.Sprintf("while listing labels for %q with ID: %q", model.ApplicationLabelableObject, testAppID),
		},
		{
			name: "Error when getting application template fail",
			appRepo: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testAppID).Return(testApplication, nil).Once()
				return appRepo
			},
			appTemplateRepo: func() *automock.ApplicationTemplateRepository {
				appTmplRepo := &automock.ApplicationTemplateRepository{}
				appTmplRepo.On("Get", emptyCtx, testAppTemplateID).Return(nil, testErr).Once()
				return appTmplRepo
			},
			runtimeRepo:    unusedRuntimeRepo,
			runtimeCtxRepo: unusedRuntimeCtxRepo,
			labelRepo: func() *automock.LabelRepository {
				lblRepo := &automock.LabelRepository{}
				lblRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testAppID).Return(testLabels, nil).Once()
				return lblRepo
			},
			expectedErrMsg: fmt.Sprintf("while getting application template with ID: %q", testAppTemplateID),
		},
		{
			name: "Error when getting application template labels fail",
			appRepo: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testAppID).Return(testApplication, nil).Once()
				return appRepo
			},
			appTemplateRepo: func() *automock.ApplicationTemplateRepository {
				appTmplRepo := &automock.ApplicationTemplateRepository{}
				appTmplRepo.On("Get", emptyCtx, testAppTemplateID).Return(testAppTemplate, nil).Once()
				return appTmplRepo
			},
			runtimeRepo:    unusedRuntimeRepo,
			runtimeCtxRepo: unusedRuntimeCtxRepo,
			labelRepo: func() *automock.LabelRepository {
				lblRepo := &automock.LabelRepository{}
				lblRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testAppID).Return(testLabels, nil).Once()
				lblRepo.On("ListForObject", emptyCtx, testTenantID, model.AppTemplateLabelableObject, testAppTemplateID).Return(nil, testErr).Once()
				return lblRepo
			},
			expectedErrMsg: fmt.Sprintf("while listing labels for %q with ID: %q", model.AppTemplateLabelableObject, testAppTemplateID),
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			appRepo := tCase.appRepo()
			appTemplateRepo := tCase.appTemplateRepo()
			runtimeRepo := tCase.runtimeRepo()
			runtimeCtxRepo := tCase.runtimeCtxRepo()
			labelRepo := tCase.labelRepo()
			defer mock.AssertExpectationsForObjects(t, appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelRepo)

			webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelRepo)

			// WHEN
			appWithLabels, appTemplateWithLabels, err := webhookDataInputBuilder.PrepareApplicationAndAppTemplateWithLabels(emptyCtx, testTenantID, testAppID)

			// THEN
			if tCase.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tCase.expectedErrMsg)
				require.Nil(t, appWithLabels)
				require.Nil(t, appTemplateWithLabels)
			} else {
				require.NoError(t, err)
				require.Equal(t, tCase.expectedAppWithLabels, appWithLabels)
				require.Equal(t, tCase.expectedAppTemplateWithLabels, appTemplateWithLabels)
			}
		})
	}
}

func TestWebhookDataInputBuilder_PrepareRuntimeWithLabels(t *testing.T) {
	testCases := []struct {
		name                          string
		appRepo                       func() *automock.ApplicationRepository
		appTemplateRepo               func() *automock.ApplicationTemplateRepository
		runtimeRepo                   func() *automock.RuntimeRepository
		runtimeCtxRepo                func() *automock.RuntimeContextRepository
		labelRepo                     func() *automock.LabelRepository
		expectedAppWithLabels         *webhook.ApplicationWithLabels
		expectedAppTemplateWithLabels *webhook.ApplicationTemplateWithLabels
		expectedRuntimeWithLabels     *webhook.RuntimeWithLabels
		expectedRuntimeCtxWithLabels  *webhook.RuntimeContextWithLabels
		expectedErrMsg                string
	}{
		{
			name:            "Success",
			appRepo:         unusedAppRepo,
			appTemplateRepo: unusedAppTemplateRepo,
			runtimeRepo: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("GetByID", emptyCtx, testTenantID, testRuntimeID).Return(testRuntime, nil).Once()
				return rtmRepo
			},
			runtimeCtxRepo: unusedRuntimeCtxRepo,
			labelRepo: func() *automock.LabelRepository {
				lblRepo := &automock.LabelRepository{}
				lblRepo.On("ListForObject", emptyCtx, testTenantID, model.RuntimeLabelableObject, testRuntimeID).Return(testLabels, nil).Once()
				return lblRepo
			},
			expectedRuntimeWithLabels: testExpectedRuntimeWithLabels,
		},
		{
			name:            "Error when getting runtime fail",
			appRepo:         unusedAppRepo,
			appTemplateRepo: unusedAppTemplateRepo,
			runtimeRepo: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("GetByID", emptyCtx, testTenantID, testRuntimeID).Return(nil, testErr).Once()
				return rtmRepo
			},
			runtimeCtxRepo: unusedRuntimeCtxRepo,
			labelRepo:      unusedLabelRepo,
			expectedErrMsg: fmt.Sprintf("while getting runtime by ID: %q", testRuntimeID),
		},
		{
			name:            "Error when getting runtime labels fail",
			appRepo:         unusedAppRepo,
			appTemplateRepo: unusedAppTemplateRepo,
			runtimeRepo: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("GetByID", emptyCtx, testTenantID, testRuntimeID).Return(testRuntime, nil).Once()
				return rtmRepo
			},
			runtimeCtxRepo: unusedRuntimeCtxRepo,
			labelRepo: func() *automock.LabelRepository {
				lblRepo := &automock.LabelRepository{}
				lblRepo.On("ListForObject", emptyCtx, testTenantID, model.RuntimeLabelableObject, testRuntimeID).Return(nil, testErr).Once()
				return lblRepo
			},
			expectedErrMsg: fmt.Sprintf("while listing labels for %q with ID: %q", model.RuntimeLabelableObject, testRuntimeID),
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			appRepo := tCase.appRepo()
			appTemplateRepo := tCase.appTemplateRepo()
			runtimeRepo := tCase.runtimeRepo()
			runtimeCtxRepo := tCase.runtimeCtxRepo()
			labelRepo := tCase.labelRepo()
			defer mock.AssertExpectationsForObjects(t, appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelRepo)

			webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelRepo)

			// WHEN
			runtimeWithLabels, err := webhookDataInputBuilder.PrepareRuntimeWithLabels(emptyCtx, testTenantID, testRuntimeID)

			// THEN
			if tCase.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tCase.expectedErrMsg)
				require.Nil(t, runtimeWithLabels)
			} else {
				require.NoError(t, err)
				require.Equal(t, tCase.expectedRuntimeWithLabels, runtimeWithLabels)
			}
		})
	}
}

func TestWebhookDataInputBuilder_PrepareRuntimeContextWithLabels(t *testing.T) {
	testCases := []struct {
		name                          string
		appRepo                       func() *automock.ApplicationRepository
		appTemplateRepo               func() *automock.ApplicationTemplateRepository
		runtimeRepo                   func() *automock.RuntimeRepository
		runtimeCtxRepo                func() *automock.RuntimeContextRepository
		labelRepo                     func() *automock.LabelRepository
		expectedAppWithLabels         *webhook.ApplicationWithLabels
		expectedAppTemplateWithLabels *webhook.ApplicationTemplateWithLabels
		expectedRuntimeWithLabels     *webhook.RuntimeWithLabels
		expectedRuntimeCtxWithLabels  *webhook.RuntimeContextWithLabels
		expectedErrMsg                string
	}{
		{
			name:            "Success",
			appRepo:         unusedAppRepo,
			appTemplateRepo: unusedAppTemplateRepo,
			runtimeRepo:     unusedRuntimeRepo,
			runtimeCtxRepo: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetByID", emptyCtx, testTenantID, testRuntimeCtxID).Return(testRuntimeCtx, nil).Once()
				return rtmCtxRepo
			},
			labelRepo: func() *automock.LabelRepository {
				lblRepo := &automock.LabelRepository{}
				lblRepo.On("ListForObject", emptyCtx, testTenantID, model.RuntimeContextLabelableObject, testRuntimeCtxID).Return(testLabels, nil).Once()
				return lblRepo
			},
			expectedRuntimeCtxWithLabels: testExpectedRuntimeCtxWithLabels,
		},
		{
			name:            "Error when getting runtime context fail",
			appRepo:         unusedAppRepo,
			appTemplateRepo: unusedAppTemplateRepo,
			runtimeRepo:     unusedRuntimeRepo,
			runtimeCtxRepo: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetByID", emptyCtx, testTenantID, testRuntimeCtxID).Return(nil, testErr).Once()
				return rtmCtxRepo
			},
			labelRepo:      unusedLabelRepo,
			expectedErrMsg: fmt.Sprintf("while getting runtime context by ID: %q", testRuntimeCtxID),
		},
		{
			name:            "Error when getting runtime context labels fail",
			appRepo:         unusedAppRepo,
			appTemplateRepo: unusedAppTemplateRepo,
			runtimeRepo:     unusedRuntimeRepo,
			runtimeCtxRepo: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetByID", emptyCtx, testTenantID, testRuntimeCtxID).Return(testRuntimeCtx, nil).Once()
				return rtmCtxRepo
			},
			labelRepo: func() *automock.LabelRepository {
				lblRepo := &automock.LabelRepository{}
				lblRepo.On("ListForObject", emptyCtx, testTenantID, model.RuntimeContextLabelableObject, testRuntimeCtxID).Return(nil, testErr).Once()
				return lblRepo
			},
			expectedErrMsg: fmt.Sprintf("while listing labels for %q with ID: %q", model.RuntimeContextLabelableObject, testRuntimeCtxID),
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			appRepo := tCase.appRepo()
			appTemplateRepo := tCase.appTemplateRepo()
			runtimeRepo := tCase.runtimeRepo()
			runtimeCtxRepo := tCase.runtimeCtxRepo()
			labelRepo := tCase.labelRepo()
			defer mock.AssertExpectationsForObjects(t, appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelRepo)

			webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelRepo)

			// WHEN
			runtimeCtxWithLabels, err := webhookDataInputBuilder.PrepareRuntimeContextWithLabels(emptyCtx, testTenantID, testRuntimeCtxID)

			// THEN
			if tCase.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tCase.expectedErrMsg)
				require.Nil(t, runtimeCtxWithLabels)
			} else {
				require.NoError(t, err)
				require.Equal(t, tCase.expectedRuntimeCtxWithLabels, runtimeCtxWithLabels)
			}
		})
	}
}

func TestWebhookDataInputBuilder_PrepareRuntimeAndRuntimeContextWithLabels(t *testing.T) {
	testCases := []struct {
		name                          string
		appRepo                       func() *automock.ApplicationRepository
		appTemplateRepo               func() *automock.ApplicationTemplateRepository
		runtimeRepo                   func() *automock.RuntimeRepository
		runtimeCtxRepo                func() *automock.RuntimeContextRepository
		labelRepo                     func() *automock.LabelRepository
		expectedAppWithLabels         *webhook.ApplicationWithLabels
		expectedAppTemplateWithLabels *webhook.ApplicationTemplateWithLabels
		expectedRuntimeWithLabels     *webhook.RuntimeWithLabels
		expectedRuntimeCtxWithLabels  *webhook.RuntimeContextWithLabels
		expectedErrMsg                string
	}{
		{
			name:            "Success",
			appRepo:         unusedAppRepo,
			appTemplateRepo: unusedAppTemplateRepo,
			runtimeRepo: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("GetByID", emptyCtx, testTenantID, testRuntimeID).Return(testRuntime, nil).Once()
				return rtmRepo
			},
			runtimeCtxRepo: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetByRuntimeID", emptyCtx, testTenantID, testRuntimeID).Return(testRuntimeCtx, nil).Once()
				return rtmCtxRepo
			},
			labelRepo: func() *automock.LabelRepository {
				lblRepo := &automock.LabelRepository{}
				lblRepo.On("ListForObject", emptyCtx, testTenantID, model.RuntimeLabelableObject, testRuntimeID).Return(testLabels, nil).Once()
				lblRepo.On("ListForObject", emptyCtx, testTenantID, model.RuntimeContextLabelableObject, testRuntimeCtxID).Return(testLabels, nil).Once()
				return lblRepo
			},
			expectedRuntimeWithLabels:    testExpectedRuntimeWithLabels,
			expectedRuntimeCtxWithLabels: testExpectedRuntimeCtxWithLabels,
		},
		{
			name:            "Error when preparing runtime with labels fail",
			appRepo:         unusedAppRepo,
			appTemplateRepo: unusedAppTemplateRepo,
			runtimeRepo: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("GetByID", emptyCtx, testTenantID, testRuntimeID).Return(nil, testErr).Once()
				return rtmRepo
			},
			runtimeCtxRepo: unusedRuntimeCtxRepo,
			labelRepo:      unusedLabelRepo,
			expectedErrMsg: fmt.Sprintf("while getting runtime by ID: %q", testRuntimeID),
		},
		{
			name:            "Error when getting runtime context by runtime ID fail",
			appRepo:         unusedAppRepo,
			appTemplateRepo: unusedAppTemplateRepo,
			runtimeRepo: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("GetByID", emptyCtx, testTenantID, testRuntimeID).Return(testRuntime, nil).Once()
				return rtmRepo
			},
			runtimeCtxRepo: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetByRuntimeID", emptyCtx, testTenantID, testRuntimeID).Return(nil, testErr).Once()
				return rtmCtxRepo
			},
			labelRepo: func() *automock.LabelRepository {
				lblRepo := &automock.LabelRepository{}
				lblRepo.On("ListForObject", emptyCtx, testTenantID, model.RuntimeLabelableObject, testRuntimeID).Return(testLabels, nil).Once()
				return lblRepo
			},
			expectedErrMsg: fmt.Sprintf("while getting runtime context for runtime with ID: %q", testRuntimeID),
		},
		{
			name:            "Error when getting runtime context labels fail",
			appRepo:         unusedAppRepo,
			appTemplateRepo: unusedAppTemplateRepo,
			runtimeRepo: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("GetByID", emptyCtx, testTenantID, testRuntimeID).Return(testRuntime, nil).Once()
				return rtmRepo
			},
			runtimeCtxRepo: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetByRuntimeID", emptyCtx, testTenantID, testRuntimeID).Return(testRuntimeCtx, nil).Once()
				return rtmCtxRepo
			},
			labelRepo: func() *automock.LabelRepository {
				lblRepo := &automock.LabelRepository{}
				lblRepo.On("ListForObject", emptyCtx, testTenantID, model.RuntimeLabelableObject, testRuntimeID).Return(testLabels, nil).Once()
				lblRepo.On("ListForObject", emptyCtx, testTenantID, model.RuntimeContextLabelableObject, testRuntimeCtxID).Return(nil, testErr).Once()
				return lblRepo
			},
			expectedErrMsg: fmt.Sprintf("while listing labels for %q with ID: %q", model.RuntimeContextLabelableObject, testRuntimeCtxID),
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			appRepo := tCase.appRepo()
			appTemplateRepo := tCase.appTemplateRepo()
			runtimeRepo := tCase.runtimeRepo()
			runtimeCtxRepo := tCase.runtimeCtxRepo()
			labelRepo := tCase.labelRepo()
			defer mock.AssertExpectationsForObjects(t, appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelRepo)

			webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelRepo)

			// WHEN
			runtimeWithLabels, runtimeCtxWithLabels, err := webhookDataInputBuilder.PrepareRuntimeAndRuntimeContextWithLabels(emptyCtx, testTenantID, testRuntimeID)

			// THEN
			if tCase.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tCase.expectedErrMsg)
				require.Nil(t, runtimeWithLabels)
				require.Nil(t, runtimeCtxWithLabels)
			} else {
				require.NoError(t, err)
				require.Equal(t, tCase.expectedRuntimeWithLabels, runtimeWithLabels)
				require.Equal(t, tCase.expectedRuntimeCtxWithLabels, runtimeCtxWithLabels)
			}
		})
	}
}

func TestWebhookDataInputBuilder_PrepareRuntimesAndRuntimeContextsMappingsInFormation(t *testing.T) {
	ctx := context.TODO()

	testCases := []struct {
		Name                            string
		RuntimeRepoFN                   func() *automock.RuntimeRepository
		RuntimeContextRepoFN            func() *automock.RuntimeContextRepository
		LabelRepoFN                     func() *automock.LabelRepository
		FormationName                   string
		ExpectedRuntimesMappings        map[string]*webhook.RuntimeWithLabels
		ExpectedRuntimeContextsMappings map[string]*webhook.RuntimeContextWithLabels
		ExpectedErrMessage              string
	}{
		{
			Name: "success",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID})
				})).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil).Once()
				repo.On("ListByScenarios", ctx, Tnt, []string{ScenarioName}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil).Once()
				return repo
			},
			RuntimeContextRepoFN: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{ScenarioName}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID, RuntimeID})
				})).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
				return repo
			},
			FormationName:                   ScenarioName,
			ExpectedRuntimesMappings:        runtimeMappings,
			ExpectedRuntimeContextsMappings: runtimeContextMappings,
			ExpectedErrMessage:              "",
		},
		{
			Name: "error when listing runtime contexts labels",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID})
				})).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil).Once()
				repo.On("ListByScenarios", ctx, Tnt, []string{ScenarioName}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil).Once()
				return repo
			},
			RuntimeContextRepoFN: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{ScenarioName}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID, RuntimeID})
				})).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(nil, testErr).Once()
				return repo
			},
			FormationName:      ScenarioName,
			ExpectedErrMessage: "while listing labels for runtime contexts",
		},
		{
			Name: "error when listing runtimes labels",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID})
				})).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil).Once()
				repo.On("ListByScenarios", ctx, Tnt, []string{ScenarioName}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil).Once()
				return repo
			},
			RuntimeContextRepoFN: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{ScenarioName}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID, RuntimeID})
				})).Return(nil, testErr).Once()
				return repo
			},
			FormationName:      ScenarioName,
			ExpectedErrMessage: "while listing runtime labels",
		},
		{
			Name: "error when listing parent runtimes",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID})
				})).Return(nil, testErr).Once()
				repo.On("ListByScenarios", ctx, Tnt, []string{ScenarioName}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil).Once()
				return repo
			},
			RuntimeContextRepoFN: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{ScenarioName}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil).Once()
				return repo
			},
			FormationName:      ScenarioName,
			ExpectedErrMessage: "while listing parent runtimes of runtime contexts in scenario",
		},
		{
			Name: "error when listing runtime contexts",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{ScenarioName}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil).Once()
				return repo
			},
			RuntimeContextRepoFN: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{ScenarioName}).Return(nil, testErr)
				return repo
			},
			FormationName:      ScenarioName,
			ExpectedErrMessage: "while listing runtime contexts in scenario",
		},
		{
			Name: "error when listing runtimes",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{ScenarioName}).Return(nil, testErr).Once()
				return repo
			},
			FormationName:      ScenarioName,
			ExpectedErrMessage: "while listing runtimes in scenario",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			runtimeRepo := unusedRuntimeRepo()
			if testCase.RuntimeRepoFN != nil {
				runtimeRepo = testCase.RuntimeRepoFN()
			}
			runtimeContextRepo := unusedRuntimeCtxRepo()
			if testCase.RuntimeContextRepoFN != nil {
				runtimeContextRepo = testCase.RuntimeContextRepoFN()
			}
			labelRepo := unusedLabelRepo()
			if testCase.LabelRepoFN != nil {
				labelRepo = testCase.LabelRepoFN()
			}

			webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(nil, nil, runtimeRepo, runtimeContextRepo, labelRepo)

			// WHEN
			runtimeMappings, runtimeContextMappings, err := webhookDataInputBuilder.PrepareRuntimesAndRuntimeContextsMappingsInFormation(emptyCtx, Tnt, testCase.FormationName)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedRuntimesMappings, runtimeMappings)
				assert.Equal(t, testCase.ExpectedRuntimeContextsMappings, runtimeContextMappings)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, runtimeMappings, runtimeContextMappings)
			}

			mock.AssertExpectationsForObjects(t, runtimeRepo, runtimeContextRepo, labelRepo)
		})
	}
}

func TestWebhookDataInputBuilder_PrepareApplicationMappingsInFormation(t *testing.T) {
	ctx := context.TODO()

	testCases := []struct {
		Name                         string
		ApplicationRepoFN            func() *automock.ApplicationRepository
		ApplicationTemplateRepoFN    func() *automock.ApplicationTemplateRepository
		LabelRepoFN                  func() *automock.LabelRepository
		FormationName                string
		ExpectedApplicationsMappings map[string]*webhook.ApplicationWithLabels
		ExpectedAppTemplateMappings  map[string]*webhook.ApplicationTemplateWithLabels
		ExpectedErrMessage           string
	}{
		{
			Name: "success",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{ScenarioName}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			FormationName:                ScenarioName,
			ExpectedApplicationsMappings: applicationMappings,
			ExpectedAppTemplateMappings:  applicationTemplateMappings,
			ExpectedErrMessage:           "",
		},
		{
			Name: "success when fails to unquote label",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{ScenarioName}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMapWithUnquotableLabels(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			FormationName:                ScenarioName,
			ExpectedApplicationsMappings: applicationMappingsWithCompositeLabel,
			ExpectedAppTemplateMappings:  applicationTemplateMappings,
			ExpectedErrMessage:           "",
		},
		{
			Name: "success when there are no applications in scenario",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{ScenarioName}).Return([]*model.Application{}, nil).Once()
				return repo
			},
			FormationName:                ScenarioName,
			ExpectedApplicationsMappings: make(map[string]*webhook.ApplicationWithLabels, 0),
			ExpectedAppTemplateMappings:  make(map[string]*webhook.ApplicationTemplateWithLabels, 0),
			ExpectedErrMessage:           "",
		},
		{
			Name: "error when listing app template labels",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{ScenarioName}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(nil, testErr).Once()
				return repo
			},
			FormationName:      ScenarioName,
			ExpectedErrMessage: "while listing labels for application templates",
		},
		{
			Name: "error when listing app templates",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{ScenarioName}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return(nil, testErr).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				return repo
			},
			FormationName:      ScenarioName,
			ExpectedErrMessage: "while listing application templates",
		},
		{
			Name: "error when listing application labels",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{ScenarioName}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(nil, testErr).Once()
				return repo
			},
			FormationName:      ScenarioName,
			ExpectedErrMessage: "while listing labels for applications",
		},
		{
			Name: "error when listing application labels",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{ScenarioName}).Return(nil, testErr).Once()
				return repo
			},
			FormationName:      ScenarioName,
			ExpectedErrMessage: "while listing applications in formation",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			applicationRepo := unusedAppRepo()
			if testCase.ApplicationRepoFN != nil {
				applicationRepo = testCase.ApplicationRepoFN()
			}
			appTemplateRepo := unusedAppTemplateRepo()
			if testCase.ApplicationTemplateRepoFN != nil {
				appTemplateRepo = testCase.ApplicationTemplateRepoFN()
			}
			labelRepo := unusedLabelRepo()
			if testCase.LabelRepoFN != nil {
				labelRepo = testCase.LabelRepoFN()
			}

			webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(applicationRepo, appTemplateRepo, nil, nil, labelRepo)

			// WHEN
			appMappings, appTemplateMappings, err := webhookDataInputBuilder.PrepareApplicationMappingsInFormation(emptyCtx, Tnt, testCase.FormationName)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedApplicationsMappings, appMappings)
				assert.Equal(t, testCase.ExpectedAppTemplateMappings, appTemplateMappings)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, appMappings, appTemplateMappings)
			}

			mock.AssertExpectationsForObjects(t, applicationRepo, appTemplateRepo, labelRepo)
		})
	}
}

func convertLabels(labels map[string]*model.Label) map[string]string {
	convertedLabels := make(map[string]string, len(labels))
	for _, l := range labels {
		stringLabel, ok := l.Value.(string)
		if !ok {
			marshalled, err := json.Marshal(l.Value)
			if err != nil {
				return nil
			}
			stringLabel = string(marshalled)
		}

		convertedLabels[l.Key] = stringLabel
	}
	return convertedLabels
}

// helper func that checks if the elements of two slices are the same no matter their order
func checkIfEqual(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	exists := make(map[string]bool)
	for _, value := range first {
		exists[value] = true
	}
	for _, value := range second {
		if !exists[value] {
			return false
		}
	}
	return true
}

func checkIfIDInSet(wh *model.Webhook, ids []string) bool {
	for _, id := range ids {
		if wh.ID == id {
			return true
		}
	}
	return false
}
