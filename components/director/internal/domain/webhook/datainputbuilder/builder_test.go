package datainputbuilder_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

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

	testTrustDetails = &webhook.TrustDetails{
		Subjects: []string{"subject1", "subject2"},
	}

	testExpectedAppTemplateWithLabelsAndTrustDetails = &webhook.ApplicationTemplateWithLabels{
		ApplicationTemplate: testAppTemplate,
		Labels:              convertLabels(testLabels),
		Tenant:              testAppTemplateTenantWithLabels,
		TrustDetails:        testTrustDetails,
	}

	testExpectedRuntimeWithLabelsAndTrustDetails = &webhook.RuntimeWithLabels{
		Runtime: testRuntime,
		Labels:  convertLabels(testLabels),
		Tenant: &webhook.TenantWithLabels{
			BusinessTenantMapping: testRuntimeOwner,
			Labels:                convertLabels(testTenantLabels),
		},
		TrustDetails: testTrustDetails,
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
		certSubjectBuilder            func() *automock.CertSubjectInputBuilder
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
			certSubjectBuilder: func() *automock.CertSubjectInputBuilder {
				builder := &automock.CertSubjectInputBuilder{}
				builder.On("GetTrustDetailsForObject", emptyCtx, testAppTemplateID).Return(testTrustDetails, nil)
				return builder
			},
			expectedAppWithLabels:         testExpectedAppWithLabels,
			expectedAppTemplateWithLabels: testExpectedAppTemplateWithLabelsAndTrustDetails,
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
			certSubjectBuilder: func() *automock.CertSubjectInputBuilder {
				builder := &automock.CertSubjectInputBuilder{}
				builder.On("GetTrustDetailsForObject", emptyCtx, testAppTemplateID).Return(testTrustDetails, nil)
				return builder
			},
			expectedAppWithLabels:         testExpectedAppWithCompositeLabel,
			expectedAppTemplateWithLabels: testExpectedAppTemplateWithLabelsAndTrustDetails,
		},
		{
			name: "Error when getting application fail",
			appRepo: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testAppID).Return(nil, testErr).Once()
				return appRepo
			},
			expectedErrMsg: fmt.Sprintf("while getting application by ID: %q", testAppID),
		},
		{
			name: "Error when building application labels fail",
			appRepo: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testAppID).Return(testApplication, nil).Once()
				return appRepo
			},
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testAppID, model.ApplicationLabelableObject).Return(nil, testErr).Once()
				return builder
			},
			expectedErrMsg: fmt.Sprintf("while building labels for application with ID %q", testAppID),
		},
		{
			name: "Error when building application tenant with labels fails",
			appRepo: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testAppID).Return(testApplication, nil).Once()
				return appRepo
			},
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
		{
			name: "Error when building application template trust details fails",
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
			certSubjectBuilder: func() *automock.CertSubjectInputBuilder {
				builder := &automock.CertSubjectInputBuilder{}
				builder.On("GetTrustDetailsForObject", emptyCtx, testAppTemplateID).Return(nil, testErr)
				return builder
			},
			expectedErrMsg: fmt.Sprintf("while building trust details for application tempalate with ID %q", testAppTemplateID),
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			appRepo := &automock.ApplicationRepository{}
			if tCase.appRepo != nil {
				appRepo = tCase.appRepo()
			}
			appTemplateRepo := &automock.ApplicationTemplateRepository{}
			if tCase.appTemplateRepo != nil {
				appTemplateRepo = tCase.appTemplateRepo()
			}
			runtimeRepo := &automock.RuntimeRepository{}
			if tCase.runtimeRepo != nil {
				runtimeRepo = tCase.runtimeRepo()
			}
			runtimeCtxRepo := &automock.RuntimeContextRepository{}
			if tCase.runtimeCtxRepo != nil {
				runtimeCtxRepo = tCase.runtimeCtxRepo()
			}
			labelBuilder := &automock.LabelInputBuilder{}
			if tCase.labelBuilder != nil {
				labelBuilder = tCase.labelBuilder()
			}
			tenantBuilder := &automock.TenantInputBuilder{}
			if tCase.tenantBuilder != nil {
				tenantBuilder = tCase.tenantBuilder()
			}
			certSubjectBuilder := &automock.CertSubjectInputBuilder{}
			if tCase.certSubjectBuilder != nil {
				certSubjectBuilder = tCase.certSubjectBuilder()
			}
			defer mock.AssertExpectationsForObjects(t, appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelBuilder, tenantBuilder, certSubjectBuilder)

			webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelBuilder, tenantBuilder, certSubjectBuilder)

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
		certSubjectBuilder            func() *automock.CertSubjectInputBuilder
		expectedAppWithLabels         *webhook.ApplicationWithLabels
		expectedAppTemplateWithLabels *webhook.ApplicationTemplateWithLabels
		expectedRuntimeWithLabels     *webhook.RuntimeWithLabels
		expectedRuntimeCtxWithLabels  *webhook.RuntimeContextWithLabels
		expectedErrMsg                string
	}{
		{
			name: "Success",
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
			certSubjectBuilder: func() *automock.CertSubjectInputBuilder {
				builder := &automock.CertSubjectInputBuilder{}
				builder.On("GetTrustDetailsForObject", emptyCtx, testRuntimeID).Return(testTrustDetails, nil)
				return builder
			},
			expectedRuntimeWithLabels: testExpectedRuntimeWithLabelsAndTrustDetails,
		},
		{
			name: "Error when getting runtime fail",
			runtimeRepo: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("GetByID", emptyCtx, testTenantID, testRuntimeID).Return(nil, testErr).Once()
				return rtmRepo
			},
			expectedErrMsg: fmt.Sprintf("while getting runtime by ID: %q", testRuntimeID),
		},
		{
			name: "Error when getting runtime labels fail",
			runtimeRepo: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("GetByID", emptyCtx, testTenantID, testRuntimeID).Return(testRuntime, nil).Once()
				return rtmRepo
			},
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, testRuntimeID, model.RuntimeLabelableObject).Return(nil, testErr).Once()
				return builder
			},
			expectedErrMsg: fmt.Sprintf("while building labels for runtime with ID %q", testRuntimeID),
		},
		{
			name: "Error when building runtime tenant with labels fail",
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
				builder.On("GetTenantForObject", emptyCtx, testRuntimeID, resource.Runtime).Return(nil, testErr).Once()
				return builder
			},
			expectedErrMsg: fmt.Sprintf("while building tenants with labels for runtime with ID %q", testRuntimeID),
		},
		{
			name: "Error when building runtime trust details fails",
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
			certSubjectBuilder: func() *automock.CertSubjectInputBuilder {
				builder := &automock.CertSubjectInputBuilder{}
				builder.On("GetTrustDetailsForObject", emptyCtx, testRuntimeID).Return(nil, testErr)
				return builder
			},
			expectedErrMsg: fmt.Sprintf("while building trust details for runtime with ID %q", testRuntimeID),
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			appRepo := &automock.ApplicationRepository{}
			if tCase.appRepo != nil {
				appRepo = tCase.appRepo()
			}
			appTemplateRepo := &automock.ApplicationTemplateRepository{}
			if tCase.appTemplateRepo != nil {
				appTemplateRepo = tCase.appTemplateRepo()
			}
			runtimeRepo := &automock.RuntimeRepository{}
			if tCase.runtimeRepo != nil {
				runtimeRepo = tCase.runtimeRepo()
			}
			runtimeCtxRepo := &automock.RuntimeContextRepository{}
			if tCase.runtimeCtxRepo != nil {
				runtimeCtxRepo = tCase.runtimeCtxRepo()
			}
			labelBuilder := &automock.LabelInputBuilder{}
			if tCase.labelBuilder != nil {
				labelBuilder = tCase.labelBuilder()
			}
			tenantBuilder := &automock.TenantInputBuilder{}
			if tCase.tenantBuilder != nil {
				tenantBuilder = tCase.tenantBuilder()
			}
			certSubjectBuilder := &automock.CertSubjectInputBuilder{}
			if tCase.certSubjectBuilder != nil {
				certSubjectBuilder = tCase.certSubjectBuilder()
			}
			defer mock.AssertExpectationsForObjects(t, appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelBuilder, tenantBuilder, certSubjectBuilder)

			webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelBuilder, tenantBuilder, certSubjectBuilder)

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
			name: "Success",
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
			name: "Error when getting runtime context fail",
			runtimeCtxRepo: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetByID", emptyCtx, testTenantID, testRuntimeCtxID).Return(nil, testErr).Once()
				return rtmCtxRepo
			},
			expectedErrMsg: fmt.Sprintf("while getting runtime context by ID: %q", testRuntimeCtxID),
		},
		{
			name: "Error when building runtime context labels fail",
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
			expectedErrMsg: fmt.Sprintf("while building labels for runtime context with ID %q", testRuntimeCtxID),
		},
		{
			name: "Error when building runtime context labels fail",
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
			appRepo := &automock.ApplicationRepository{}
			if tCase.appRepo != nil {
				appRepo = tCase.appRepo()
			}
			appTemplateRepo := &automock.ApplicationTemplateRepository{}
			if tCase.appTemplateRepo != nil {
				appTemplateRepo = tCase.appTemplateRepo()
			}
			runtimeRepo := &automock.RuntimeRepository{}
			if tCase.runtimeRepo != nil {
				runtimeRepo = tCase.runtimeRepo()
			}
			runtimeCtxRepo := &automock.RuntimeContextRepository{}
			if tCase.runtimeCtxRepo != nil {
				runtimeCtxRepo = tCase.runtimeCtxRepo()
			}
			labelBuilder := &automock.LabelInputBuilder{}
			if tCase.labelBuilder != nil {
				labelBuilder = tCase.labelBuilder()
			}
			tenantBuilder := &automock.TenantInputBuilder{}
			if tCase.tenantBuilder != nil {
				tenantBuilder = tCase.tenantBuilder()
			}
			defer mock.AssertExpectationsForObjects(t, appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelBuilder, tenantBuilder)

			webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(appRepo, appTemplateRepo, runtimeRepo, runtimeCtxRepo, labelBuilder, tenantBuilder, nil)

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
