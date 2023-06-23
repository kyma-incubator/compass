package datainputbuilder_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

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
	emptyCtx          = context.Background()
	testErr           = errors.New("test error")
	testTenantID      = "testTenantID"
	testTenantOwnerID = "testTenantOwnerID"
	testRuntimeID     = "testRuntimeID"
	testRuntimeCtxID  = "testRuntimeCtxID"

	testLabels = map[string]*model.Label{"testLabelKey": {
		ID:     "testLabelID",
		Key:    "testLabelKey",
		Tenant: &testTenantID,
		Value:  "testLabelValue",
	}}

	testTenantLabels = map[string]*model.Label{"testLabelKey": {
		ID:     "testTenantLabelID",
		Key:    "testLabelKey",
		Tenant: &testTenantOwnerID,
		Value:  "testTenantLabelValue",
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

	testRuntimeOwner = &model.BusinessTenantMapping{
		Name: "testRuntimeOwner",
	}

	testExpectedRuntimeWithLabels = &webhook.RuntimeWithLabels{
		Runtime: testRuntime,
		Labels:  convertLabels(testLabels),
		Tenant: &webhook.TenantWithLabels{
			BusinessTenantMapping: testRuntimeOwner,
			Labels:                convertLabels(testTenantLabels),
		},
	}

	testRuntimeCtx = &model.RuntimeContext{
		ID:    testRuntimeCtxID,
		Key:   "testRtmCtxKey",
		Value: "testRtmCtxValue",
	}

	testRuntimeCtxOwner = &model.BusinessTenantMapping{
		Name: "testRuntimeCtxOwner",
	}

	testExpectedRuntimeCtxWithLabels = &webhook.RuntimeContextWithLabels{
		RuntimeContext: testRuntimeCtx,
		Labels:         convertLabels(testLabels),
		Tenant: &webhook.TenantWithLabels{
			BusinessTenantMapping: testRuntimeCtxOwner,
			Labels:                convertLabels(testTenantLabels),
		},
	}

	testApplicationTenantOwner = &model.BusinessTenantMapping{
		Name: "testAppTenant",
	}
	testApplicationTemplateTenantOwner = &model.BusinessTenantMapping{
		ID:   ApplicationTemplateTenantID,
		Name: "testAppTenant",
	}

	testAppID         = "testAppID"
	testAppTemplateID = "testAppTemplateID"

	testApplication = &model.Application{
		Name:                  "testAppName",
		ApplicationTemplateID: &testAppTemplateID,
	}

	testAppTemplate = &model.ApplicationTemplate{
		ID:   testAppTemplateID,
		Name: "testAppTemplateName",
	}

	testAppTenantWithLabels = &webhook.TenantWithLabels{
		BusinessTenantMapping: testApplicationTenantOwner,
		Labels:                convertLabels(testTenantLabels),
	}

	testExpectedAppWithLabels = &webhook.ApplicationWithLabels{
		Application: testApplication,
		Labels:      convertLabels(testLabels),
		Tenant:      testAppTenantWithLabels,
	}

	testExpectedAppWithCompositeLabel = &webhook.ApplicationWithLabels{
		Application: testApplication,
		Labels:      convertLabels(testLabelsComposite),
		Tenant:      testAppTenantWithLabels,
	}

	testAppTemplateTenantWithLabels = &webhook.TenantWithLabels{
		BusinessTenantMapping: testApplicationTemplateTenantOwner,
		Labels:                convertLabels(testTenantLabels),
	}

	testExpectedAppTemplateWithLabels = &webhook.ApplicationTemplateWithLabels{
		ApplicationTemplate: testAppTemplate,
		Labels:              convertLabels(testLabels),
		Tenant:              testAppTemplateTenantWithLabels,
	}
)

func TestWebhookDataInputBuilder_PrepareApplicationAndAppTemplateWithLabels(t *testing.T) {
	testCases := []struct {
		name                          string
		appRepo                       func() *automock.ApplicationRepository
		appTemplateRepo               func() *automock.ApplicationTemplateRepository
		runtimeRepo                   func() *automock.RuntimeRepository
		runtimeCtxRepo                func() *automock.RuntimeContextRepository
		labelBuilder                  func() *automock.LabelInputBuilder
		tenantBuilder                 func() *automock.TenantInputBuilder
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
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testAppID, model.ApplicationLabelableObject).Return(testExpectedAppWithLabels.Labels, nil).Once()
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testAppTemplateID, model.AppTemplateLabelableObject).Return(testExpectedAppTemplateWithLabels.Labels, nil).Once()
				return builder
			},
			tenantBuilder: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantForObject", emptyCtx, testAppID, resource.Application).Return(testExpectedAppWithLabels.Tenant, nil).Once()
				builder.On("GetTenantForApplicationTemplate", emptyCtx, testTenantID, testExpectedAppTemplateWithLabels.Labels).Return(testExpectedAppTemplateWithLabels.Tenant, nil).Once()
				return builder
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
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testAppID, model.ApplicationLabelableObject).Return(testExpectedAppWithCompositeLabel.Labels, nil).Once()
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testAppTemplateID, model.AppTemplateLabelableObject).Return(testExpectedAppTemplateWithLabels.Labels, nil).Once()
				return builder
			},
			tenantBuilder: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantForObject", emptyCtx, testAppID, resource.Application).Return(testExpectedAppWithCompositeLabel.Tenant, nil).Once()
				builder.On("GetTenantForApplicationTemplate", emptyCtx, testTenantID, testExpectedAppTemplateWithLabels.Labels).Return(testExpectedAppTemplateWithLabels.Tenant, nil).Once()
				return builder
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
			labelBuilder:    unusedLabelBuilder,
			tenantBuilder:   unusedTenantBuilder,
			expectedErrMsg:  fmt.Sprintf("while getting application by ID: %q", testAppID),
		},
		{
			name: "Error when building application labels fail",
			appRepo: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testAppID).Return(testApplication, nil).Once()
				return appRepo
			},
			appTemplateRepo: unusedAppTemplateRepo,
			runtimeRepo:     unusedRuntimeRepo,
			runtimeCtxRepo:  unusedRuntimeCtxRepo,
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testAppID, model.ApplicationLabelableObject).Return(nil, testErr).Once()
				return builder
			},
			tenantBuilder:  unusedTenantBuilder,
			expectedErrMsg: fmt.Sprintf("while building labels for application with ID %q", testAppID),
		},
		{
			name: "Error when building application tenant with labels fails",
			appRepo: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testAppID).Return(testApplication, nil).Once()
				return appRepo
			},
			appTemplateRepo: unusedAppTemplateRepo,
			runtimeRepo:     unusedRuntimeRepo,
			runtimeCtxRepo:  unusedRuntimeCtxRepo,
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testAppID, model.ApplicationLabelableObject).Return(testExpectedAppWithCompositeLabel.Labels, nil).Once()
				return builder
			},
			tenantBuilder: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantForObject", emptyCtx, testAppID, resource.Application).Return(nil, testErr).Once()
				return builder
			},
			expectedErrMsg: fmt.Sprintf("while building tenant with labels for application with ID %q", testAppID),
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
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testAppID, model.ApplicationLabelableObject).Return(testExpectedAppWithLabels.Labels, nil).Once()
				return builder
			},
			tenantBuilder: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantForObject", emptyCtx, testAppID, resource.Application).Return(testExpectedAppWithLabels.Tenant, nil).Once()
				return builder
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
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testAppID, model.ApplicationLabelableObject).Return(testExpectedAppWithLabels.Labels, nil).Once()
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testAppTemplateID, model.AppTemplateLabelableObject).Return(nil, testErr).Once()
				return builder
			},
			tenantBuilder: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantForObject", emptyCtx, testAppID, resource.Application).Return(testExpectedAppWithLabels.Tenant, nil).Once()
				return builder
			},
			expectedErrMsg: fmt.Sprintf("while building labels for application template with ID %q", testAppTemplateID),
		},
		{
			name: "Error when building application template tenant with labels fails",
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
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testAppID, model.ApplicationLabelableObject).Return(testExpectedAppWithLabels.Labels, nil).Once()
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testAppTemplateID, model.AppTemplateLabelableObject).Return(testExpectedAppTemplateWithLabels.Labels, nil).Once()
				return builder
			},
			tenantBuilder: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantForObject", emptyCtx, testAppID, resource.Application).Return(testExpectedAppWithLabels.Tenant, nil).Once()
				builder.On("GetTenantForApplicationTemplate", emptyCtx, testTenantID, testExpectedAppTemplateWithLabels.Labels).Return(nil, testErr).Once()
				return builder
			},
			expectedErrMsg: fmt.Sprintf("while building tenant with labels for application template with ID %q", testAppTemplateID),
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			appRepo := tCase.appRepo()
			appTemplateRepo := tCase.appTemplateRepo()
			runtimeRepo := tCase.runtimeRepo()
			runtimeCtxRepo := tCase.runtimeCtxRepo()
			labelBuilder := tCase.labelBuilder()
			tenantBuilder := tCase.tenantBuilder()
			defer mock.AssertExpectationsForObjects(t, appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelBuilder, tenantBuilder)

			webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelBuilder, tenantBuilder)

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
		labelBuilder                  func() *automock.LabelInputBuilder
		tenantBuilder                 func() *automock.TenantInputBuilder
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
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testRuntimeID, model.RuntimeLabelableObject).Return(testExpectedRuntimeWithLabels.Labels, nil).Once()
				return builder
			},
			tenantBuilder: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantForObject", emptyCtx, testRuntimeID, resource.Runtime).Return(testExpectedRuntimeWithLabels.Tenant, nil).Once()
				return builder
			},
			runtimeCtxRepo:            unusedRuntimeCtxRepo,
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
			labelBuilder:   unusedLabelBuilder,
			tenantBuilder:  unusedTenantBuilder,
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
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testRuntimeID, model.RuntimeLabelableObject).Return(nil, testErr).Once()
				return builder
			},
			tenantBuilder:  unusedTenantBuilder,
			expectedErrMsg: fmt.Sprintf("while building labels for runtime with ID %q", testRuntimeID),
		},
		{
			name:            "Error when building runtime tenant with labels fail",
			appRepo:         unusedAppRepo,
			appTemplateRepo: unusedAppTemplateRepo,
			runtimeRepo: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("GetByID", emptyCtx, testTenantID, testRuntimeID).Return(testRuntime, nil).Once()
				return rtmRepo
			},
			runtimeCtxRepo: unusedRuntimeCtxRepo,
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testRuntimeID, model.RuntimeLabelableObject).Return(testExpectedRuntimeWithLabels.Labels, nil).Once()
				return builder
			},
			tenantBuilder: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantForObject", emptyCtx, testRuntimeID, resource.Runtime).Return(nil, testErr).Once()
				return builder
			},
			expectedErrMsg: fmt.Sprintf("while building tenants with labels for runtime with ID %q", testRuntimeID),
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			appRepo := tCase.appRepo()
			appTemplateRepo := tCase.appTemplateRepo()
			runtimeRepo := tCase.runtimeRepo()
			runtimeCtxRepo := tCase.runtimeCtxRepo()
			labelBuilder := tCase.labelBuilder()
			tenantBuilder := tCase.tenantBuilder()
			defer mock.AssertExpectationsForObjects(t, appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelBuilder, tenantBuilder)

			webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelBuilder, tenantBuilder)

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
		labelBuilder                  func() *automock.LabelInputBuilder
		tenantBuilder                 func() *automock.TenantInputBuilder
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
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testRuntimeCtxID, model.RuntimeContextLabelableObject).Return(testExpectedRuntimeCtxWithLabels.Labels, nil).Once()
				return builder
			},
			tenantBuilder: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantForObject", emptyCtx, testRuntimeCtxID, resource.RuntimeContext).Return(testExpectedRuntimeCtxWithLabels.Tenant, nil).Once()
				return builder
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
			labelBuilder:   unusedLabelBuilder,
			tenantBuilder:  unusedTenantBuilder,
			expectedErrMsg: fmt.Sprintf("while getting runtime context by ID: %q", testRuntimeCtxID),
		},
		{
			name:            "Error when building runtime context labels fail",
			appRepo:         unusedAppRepo,
			appTemplateRepo: unusedAppTemplateRepo,
			runtimeRepo:     unusedRuntimeRepo,
			runtimeCtxRepo: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetByID", emptyCtx, testTenantID, testRuntimeCtxID).Return(testRuntimeCtx, nil).Once()
				return rtmCtxRepo
			},
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testRuntimeCtxID, model.RuntimeContextLabelableObject).Return(nil, testErr).Once()
				return builder
			},
			tenantBuilder:  unusedTenantBuilder,
			expectedErrMsg: fmt.Sprintf("while building labels for runtime context with ID %q", testRuntimeCtxID),
		},
		{
			name:            "Error when building runtime context labels fail",
			appRepo:         unusedAppRepo,
			appTemplateRepo: unusedAppTemplateRepo,
			runtimeRepo:     unusedRuntimeRepo,
			runtimeCtxRepo: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetByID", emptyCtx, testTenantID, testRuntimeCtxID).Return(testRuntimeCtx, nil).Once()
				return rtmCtxRepo
			},
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testRuntimeCtxID, model.RuntimeContextLabelableObject).Return(testExpectedRuntimeCtxWithLabels.Labels, nil).Once()
				return builder
			},
			tenantBuilder: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantForObject", emptyCtx, testRuntimeCtxID, resource.RuntimeContext).Return(nil, testErr).Once()
				return builder
			},
			expectedErrMsg: fmt.Sprintf("while building tenant with labels for runtime context with ID %q", testRuntimeCtxID),
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			appRepo := tCase.appRepo()
			appTemplateRepo := tCase.appTemplateRepo()
			runtimeRepo := tCase.runtimeRepo()
			runtimeCtxRepo := tCase.runtimeCtxRepo()
			labelBuilder := tCase.labelBuilder()
			tenantBuilder := tCase.tenantBuilder()
			defer mock.AssertExpectationsForObjects(t, appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelBuilder, tenantBuilder)

			webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelBuilder, tenantBuilder)

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
		labelBuilder                  func() *automock.LabelInputBuilder
		tenantBuilder                 func() *automock.TenantInputBuilder
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
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testRuntimeID, model.RuntimeLabelableObject).Return(testExpectedRuntimeWithLabels.Labels, nil).Once()
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testRuntimeCtxID, model.RuntimeContextLabelableObject).Return(testExpectedRuntimeCtxWithLabels.Labels, nil).Once()
				return builder
			},
			tenantBuilder: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantForObject", emptyCtx, testRuntimeID, resource.Runtime).Return(testExpectedRuntimeWithLabels.Tenant, nil).Once()
				builder.On("GetTenantForObject", emptyCtx, testRuntimeCtxID, resource.RuntimeContext).Return(testExpectedRuntimeCtxWithLabels.Tenant, nil).Once()
				return builder
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
			labelBuilder:   unusedLabelBuilder,
			tenantBuilder:  unusedTenantBuilder,
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
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testRuntimeID, model.RuntimeLabelableObject).Return(testExpectedRuntimeWithLabels.Labels, nil).Once()
				return builder
			},
			tenantBuilder: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantForObject", emptyCtx, testRuntimeID, resource.Runtime).Return(testExpectedRuntimeWithLabels.Tenant, nil).Once()
				return builder
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
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testRuntimeID, model.RuntimeLabelableObject).Return(testExpectedRuntimeWithLabels.Labels, nil).Once()
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testRuntimeCtxID, model.RuntimeContextLabelableObject).Return(nil, testErr).Once()
				return builder
			},
			tenantBuilder: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantForObject", emptyCtx, testRuntimeID, resource.Runtime).Return(testExpectedRuntimeWithLabels.Tenant, nil).Once()
				return builder
			},
			expectedErrMsg: fmt.Sprintf("while building labels for runtime context with ID %q", testRuntimeCtxID),
		},
		{
			name:            "Error when building runtime context tenant with labels fail",
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
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testRuntimeID, model.RuntimeLabelableObject).Return(testExpectedRuntimeWithLabels.Labels, nil).Once()
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testRuntimeCtxID, model.RuntimeContextLabelableObject).Return(testExpectedRuntimeCtxWithLabels.Labels, nil).Once()
				return builder
			},
			tenantBuilder: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantForObject", emptyCtx, testRuntimeID, resource.Runtime).Return(testExpectedRuntimeWithLabels.Tenant, nil).Once()
				builder.On("GetTenantForObject", emptyCtx, testRuntimeCtxID, resource.RuntimeContext).Return(nil, testErr).Once()
				return builder
			},
			expectedErrMsg: fmt.Sprintf("while building tenant with labels for runtime context with ID %q", testRuntimeCtxID),
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			appRepo := tCase.appRepo()
			appTemplateRepo := tCase.appTemplateRepo()
			runtimeRepo := tCase.runtimeRepo()
			runtimeCtxRepo := tCase.runtimeCtxRepo()
			labelBuilder := tCase.labelBuilder()
			tenantBuilder := tCase.tenantBuilder()
			defer mock.AssertExpectationsForObjects(t, appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelBuilder, tenantBuilder)

			webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelBuilder, tenantBuilder)

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
	testRuntimeContextTenantWithLabels := &webhook.TenantWithLabels{
		BusinessTenantMapping: testRuntimeCtxOwner,
		Labels:                convertLabels(testTenantLabels),
	}
	testRuntimeTenantWithLabels := &webhook.TenantWithLabels{
		BusinessTenantMapping: testRuntimeOwner,
		Labels:                convertLabels(testTenantLabels),
	}

	testCases := []struct {
		Name                            string
		RuntimeRepoFN                   func() *automock.RuntimeRepository
		RuntimeContextRepoFN            func() *automock.RuntimeContextRepository
		LabelBuilderFN                  func() *automock.LabelInputBuilder
		TenantBuilderFN                 func() *automock.TenantInputBuilder
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
			LabelBuilderFN: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID, RuntimeID})
				}), model.RuntimeLabelableObject).Return(map[string]map[string]string{
					RuntimeID:               fixLabelsMapForRuntimeWithLabels(),
					RuntimeContextRuntimeID: fixLabelsMapForRuntimeWithLabels(),
				}, nil).Once()
				builder.On("GetLabelsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeContextID, RuntimeContext2ID})
				}), model.RuntimeContextLabelableObject).Return(map[string]map[string]string{
					RuntimeContextID:  fixLabelsMapForRuntimeContextWithLabels(),
					RuntimeContext2ID: fixLabelsMapForRuntimeContextWithLabels(),
				}, nil).Once()

				return builder
			},
			TenantBuilderFN: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID, RuntimeID})
				}), resource.Runtime).Return(map[string]*webhook.TenantWithLabels{
					RuntimeID:               testRuntimeTenantWithLabels,
					RuntimeContextRuntimeID: testRuntimeTenantWithLabels,
				}, nil)
				builder.On("GetTenantsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeContextID, RuntimeContext2ID})
				}), resource.RuntimeContext).Return(map[string]*webhook.TenantWithLabels{
					RuntimeContextID:  testRuntimeContextTenantWithLabels,
					RuntimeContext2ID: testRuntimeContextTenantWithLabels,
				}, nil)
				return builder
			},
			FormationName:                   ScenarioName,
			ExpectedRuntimesMappings:        runtimeMappings,
			ExpectedRuntimeContextsMappings: runtimeContextMappings,
			ExpectedErrMessage:              "",
		},
		{
			Name: "error when building tenant with labels for runtime context",
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
			LabelBuilderFN: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID, RuntimeID})
				}), model.RuntimeLabelableObject).Return(map[string]map[string]string{
					RuntimeID:               fixLabelsMapForRuntimeWithLabels(),
					RuntimeContextRuntimeID: fixLabelsMapForRuntimeWithLabels(),
				}, nil).Once()
				builder.On("GetLabelsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeContextID, RuntimeContext2ID})
				}), model.RuntimeContextLabelableObject).Return(map[string]map[string]string{
					RuntimeContextID:  fixLabelsMapForRuntimeContextWithLabels(),
					RuntimeContext2ID: fixLabelsMapForRuntimeContextWithLabels(),
				}, nil).Once()

				return builder
			},
			TenantBuilderFN: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID, RuntimeID})
				}), resource.Runtime).Return(map[string]*webhook.TenantWithLabels{
					RuntimeID:               testRuntimeTenantWithLabels,
					RuntimeContextRuntimeID: testRuntimeTenantWithLabels,
				}, nil)
				builder.On("GetTenantsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeContextID, RuntimeContext2ID})
				}), resource.RuntimeContext).Return(nil, testErr)
				return builder
			},
			FormationName:      ScenarioName,
			ExpectedErrMessage: "while building tenants with labels for runtime contexts",
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
			LabelBuilderFN: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID, RuntimeID})
				}), model.RuntimeLabelableObject).Return(map[string]map[string]string{
					RuntimeID:               fixLabelsMapForRuntimeWithLabels(),
					RuntimeContextRuntimeID: fixLabelsMapForRuntimeWithLabels(),
				}, nil).Once()
				builder.On("GetLabelsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeContextID, RuntimeContext2ID})
				}), model.RuntimeContextLabelableObject).Return(nil, testErr).Once()

				return builder
			},
			TenantBuilderFN: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID, RuntimeID})
				}), resource.Runtime).Return(map[string]*webhook.TenantWithLabels{
					RuntimeID:               testRuntimeTenantWithLabels,
					RuntimeContextRuntimeID: testRuntimeTenantWithLabels,
				}, nil)
				return builder
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
			LabelBuilderFN: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID, RuntimeID})
				}), model.RuntimeLabelableObject).Return(nil, testErr).Once()
				return builder
			},
			FormationName:      ScenarioName,
			ExpectedErrMessage: "while listing runtime labels",
		},
		{
			Name: "error when building tenants with labels for runtimes",
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
			LabelBuilderFN: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID, RuntimeID})
				}), model.RuntimeLabelableObject).Return(map[string]map[string]string{
					RuntimeID:               fixLabelsMapForRuntimeWithLabels(),
					RuntimeContextRuntimeID: fixLabelsMapForRuntimeWithLabels(),
				}, nil).Once()

				return builder
			},
			TenantBuilderFN: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID, RuntimeID})
				}), resource.Runtime).Return(nil, testErr)
				return builder
			},
			FormationName:      ScenarioName,
			ExpectedErrMessage: "while building tenants with labels for runtimes",
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
			labelBuilder := unusedLabelBuilder()
			if testCase.LabelBuilderFN != nil {
				labelBuilder = testCase.LabelBuilderFN()
			}
			tenantBuilder := unusedTenantBuilder()
			if testCase.TenantBuilderFN != nil {
				tenantBuilder = testCase.TenantBuilderFN()
			}

			webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(nil, nil, runtimeRepo, runtimeContextRepo, labelBuilder, tenantBuilder)

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

			mock.AssertExpectationsForObjects(t, runtimeRepo, runtimeContextRepo, labelBuilder, tenantBuilder)
		})
	}
}

func TestWebhookDataInputBuilder_PrepareApplicationMappingsInFormation(t *testing.T) {
	ctx := context.TODO()

	testCases := []struct {
		Name                         string
		ApplicationRepoFN            func() *automock.ApplicationRepository
		ApplicationTemplateRepoFN    func() *automock.ApplicationTemplateRepository
		LabelBuilderFN               func() *automock.LabelInputBuilder
		TenantBuilderFN              func() *automock.TenantInputBuilder
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
			LabelBuilderFN: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{ApplicationID, Application2ID})
				}), model.ApplicationLabelableObject).Return(map[string]map[string]string{
					ApplicationID:  fixLabelsMapForApplicationWithLabels(),
					Application2ID: fixLabelsMapForApplicationWithLabels(),
				}, nil).Once()
				builder.On("GetLabelsForObjects", emptyCtx, Tnt, []string{ApplicationTemplateID}, model.AppTemplateLabelableObject).
					Return(map[string]map[string]string{
						ApplicationTemplateID: fixLabelsMapForApplicationTemplateWithLabels(),
					}, nil).Once()

				return builder
			},
			TenantBuilderFN: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{ApplicationID, Application2ID})
				}), resource.Application).Return(map[string]*webhook.TenantWithLabels{
					ApplicationID:  testAppTenantWithLabels,
					Application2ID: testAppTenantWithLabels,
				}, nil)
				builder.On("GetTenantsForApplicationTemplates", emptyCtx, Tnt, map[string]map[string]string{
					ApplicationTemplateID: fixLabelsMapForApplicationTemplateWithLabels(),
				}, []string{ApplicationTemplateID}).Return(map[string]*webhook.TenantWithLabels{
					ApplicationTemplateID: testAppTemplateTenantWithLabels,
				}, nil)
				return builder
			},
			FormationName:                ScenarioName,
			ExpectedApplicationsMappings: applicationMappings,
			ExpectedAppTemplateMappings:  applicationTemplateMappings,
			ExpectedErrMessage:           "",
		},
		{
			Name: "error when building tenant with labels for application templates",
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
			LabelBuilderFN: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{ApplicationID, Application2ID})
				}), model.ApplicationLabelableObject).Return(map[string]map[string]string{
					ApplicationID:  fixLabelsMapForApplicationWithLabels(),
					Application2ID: fixLabelsMapForApplicationWithLabels(),
				}, nil).Once()
				builder.On("GetLabelsForObjects", emptyCtx, Tnt, []string{ApplicationTemplateID}, model.AppTemplateLabelableObject).
					Return(nil, testErr).Once()
				return builder
			},
			TenantBuilderFN: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{ApplicationID, Application2ID})
				}), resource.Application).Return(map[string]*webhook.TenantWithLabels{
					ApplicationID:  testAppTenantWithLabels,
					Application2ID: testAppTenantWithLabels,
				}, nil)
				return builder
			},
			FormationName:      ScenarioName,
			ExpectedErrMessage: "while listing labels for application templates",
		},
		{
			Name: "error when building tenant with labels for application templates",
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
			LabelBuilderFN: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{ApplicationID, Application2ID})
				}), model.ApplicationLabelableObject).Return(map[string]map[string]string{
					ApplicationID:  fixLabelsMapForApplicationWithLabels(),
					Application2ID: fixLabelsMapForApplicationWithLabels(),
				}, nil).Once()
				builder.On("GetLabelsForObjects", emptyCtx, Tnt, []string{ApplicationTemplateID}, model.AppTemplateLabelableObject).
					Return(map[string]map[string]string{
						ApplicationTemplateID: fixLabelsMapForApplicationTemplateWithLabels(),
					}, nil).Once()

				return builder
			},
			TenantBuilderFN: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{ApplicationID, Application2ID})
				}), resource.Application).Return(map[string]*webhook.TenantWithLabels{
					ApplicationID:  testAppTenantWithLabels,
					Application2ID: testAppTenantWithLabels,
				}, nil)
				builder.On("GetTenantsForApplicationTemplates", emptyCtx, Tnt, map[string]map[string]string{
					ApplicationTemplateID: fixLabelsMapForApplicationTemplateWithLabels(),
				}, []string{ApplicationTemplateID}).Return(nil, testErr)
				return builder
			},
			FormationName:      ScenarioName,
			ExpectedErrMessage: "while building tenants with labels for application templates",
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
			Name: "error when listing app labels",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{ScenarioName}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			LabelBuilderFN: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{ApplicationID, Application2ID})
				}), model.ApplicationLabelableObject).Return(nil, testErr).Once()
				return builder
			},
			FormationName:      ScenarioName,
			ExpectedErrMessage: "while listing labels for applications",
		},
		{
			Name: "error when listing app labels",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{ScenarioName}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			LabelBuilderFN: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{ApplicationID, Application2ID})
				}), model.ApplicationLabelableObject).Return(map[string]map[string]string{
					ApplicationID:  fixLabelsMapForApplicationWithLabels(),
					Application2ID: fixLabelsMapForApplicationWithLabels(),
				}, nil).Once()
				return builder
			},
			TenantBuilderFN: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{ApplicationID, Application2ID})
				}), resource.Application).Return(nil, testErr)
				return builder
			},
			FormationName:      ScenarioName,
			ExpectedErrMessage: "while building tenants with labels for applications",
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
			LabelBuilderFN: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{ApplicationID, Application2ID})
				}), model.ApplicationLabelableObject).Return(map[string]map[string]string{
					ApplicationID:  fixLabelsMapForApplicationWithLabels(),
					Application2ID: fixLabelsMapForApplicationWithLabels(),
				}, nil).Once()
				return builder
			},
			TenantBuilderFN: func() *automock.TenantInputBuilder {
				builder := &automock.TenantInputBuilder{}
				builder.On("GetTenantsForObjects", emptyCtx, Tnt, mock.MatchedBy(func(ids []string) bool {
					return checkIfEqual(ids, []string{ApplicationID, Application2ID})
				}), resource.Application).Return(map[string]*webhook.TenantWithLabels{
					ApplicationID:  testAppTenantWithLabels,
					Application2ID: testAppTenantWithLabels,
				}, nil)
				return builder
			},
			FormationName:      ScenarioName,
			ExpectedErrMessage: "while listing application templates",
		},
		{
			Name: "error when listing application in formation",
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
			labelBuilder := unusedLabelBuilder()
			if testCase.LabelBuilderFN != nil {
				labelBuilder = testCase.LabelBuilderFN()
			}
			tenantBuilder := unusedTenantBuilder()
			if testCase.TenantBuilderFN != nil {
				tenantBuilder = testCase.TenantBuilderFN()
			}

			webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(applicationRepo, appTemplateRepo, nil, nil, labelBuilder, tenantBuilder)

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

			mock.AssertExpectationsForObjects(t, applicationRepo, appTemplateRepo, labelBuilder, tenantBuilder)
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
