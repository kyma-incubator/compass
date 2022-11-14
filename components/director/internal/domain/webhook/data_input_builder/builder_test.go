package data_input_builder_test

import (
	"context"
	"fmt"
	"testing"

	databuilder "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/data_input_builder"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/data_input_builder/automock"

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

func convertLabels(labels map[string]*model.Label) map[string]interface{} {
	convertedLabels := make(map[string]interface{}, len(labels))
	for _, l := range labels {
		convertedLabels[l.Key] = l.Value
	}
	return convertedLabels
}
