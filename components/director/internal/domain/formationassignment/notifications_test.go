package formationassignment_test

import (
	"fmt"
	"testing"

	databuilderautomock "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder/automock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_GenerateNotification(t *testing.T) {
	testRuntimeID := "testRuntimeID"
	testNotFoundErr := apperrors.NewNotFoundError(resource.Webhook, TestTarget)
	faWithInvalidTypes := fixFormationAssignmentModel(TestConfigValueRawJSON)

	faWithSourceAppAndTargetApp := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), TestConfigValueRawJSON)
	faWithSourceAppAndTargetAppReverse := fixReverseFormationAssignment(faWithSourceAppAndTargetApp)

	faWithSourceRuntimeAndTargetApp := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), TestConfigValueRawJSON)
	faWithSourceRuntimeAndTargetAppReverse := fixReverseFormationAssignment(faWithSourceRuntimeAndTargetApp)

	faWithSourceRuntimeCtxAndTargetApp := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeRuntimeContext, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), TestConfigValueRawJSON)
	faWithSourceRuntimeCtxAndTargetAppReverse := fixReverseFormationAssignment(faWithSourceRuntimeCtxAndTargetApp)

	faWithSourceAppAndTargetRuntime := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, string(model.ReadyAssignmentState), TestConfigValueRawJSON)
	faWithSourceAppAndTargetRuntimeReverse := fixReverseFormationAssignment(faWithSourceAppAndTargetRuntime)

	faWithSourceInvalidAndTargetRuntime := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeRuntime, string(model.ReadyAssignmentState), TestConfigValueRawJSON)
	faWithSourceInvalidAndTargetRtmCtx := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeRuntimeContext, model.FormationAssignmentTypeRuntimeContext, string(model.ReadyAssignmentState), TestConfigValueRawJSON)

	faWithSourceAppCtxAndTargetRtmCtx := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntimeContext, string(model.ReadyAssignmentState), TestConfigValueRawJSON)
	faWithSourceAppCtxAndTargetRtmCtxReverse := fixReverseFormationAssignment(faWithSourceAppCtxAndTargetRtmCtx)

	testAppWebhook := &model.Webhook{
		ID:         "TestAppWebhookID",
		ObjectID:   TestSource,
		ObjectType: model.ApplicationWebhookReference,
	}

	testRuntimeWebhook := &model.Webhook{
		ID:         "TestRuntimeWebhookID",
		ObjectID:   TestSource,
		ObjectType: model.RuntimeWebhookReference,
	}

	testLabels := map[string]interface{}{"testLabelKey": "testLabelValue"}
	testAppWithLabels := &webhook.ApplicationWithLabels{
		Application: &model.Application{
			Name: "testAppName",
		},
		Labels: testLabels,
	}
	testAppTemplateWithLabels := &webhook.ApplicationTemplateWithLabels{
		ApplicationTemplate: &model.ApplicationTemplate{
			ID:   "testAppTemplateID",
			Name: "testAppTemplateName",
		},
		Labels: testLabels,
	}

	testRuntimeWithLabels := &webhook.RuntimeWithLabels{
		Runtime: &model.Runtime{
			Name: "testRuntimeName",
		},
		Labels: testLabels,
	}
	testRuntimeCtxWithLabels := &webhook.RuntimeContextWithLabels{
		RuntimeContext: &model.RuntimeContext{
			ID:        "testRuntimeCtxID",
			RuntimeID: testRuntimeID,
			Key:       "testKey",
			Value:     "testValue",
		},
		Labels: testLabels,
	}

	testGqlAppWebhook := &graphql.Webhook{
		ID:   testAppWebhook.ID,
		Type: graphql.WebhookType(testAppWebhook.Type),
	}

	testGqlRuntimeWebhook := &graphql.Webhook{
		ID:   testRuntimeWebhook.ID,
		Type: graphql.WebhookType(testRuntimeWebhook.Type),
	}

	testCustomerTenantContext := &webhook.CustomerTenantContext{
		Tenant:     TestTenantID,
		CustomerID: TntParentID,
	}

	testAppTenantMappingWebhookInput := fixAppTenantMappingWebhookInput(TestFormationID, testAppWithLabels, testAppWithLabels, testAppTemplateWithLabels, testAppTemplateWithLabels, testCustomerTenantContext, fixConvertFAFromModel(faWithSourceAppAndTargetApp), fixConvertFAFromModel(faWithSourceAppAndTargetAppReverse))
	testAppNotificationReqWithTenantMappingType := &webhookclient.NotificationRequest{
		Webhook:       *testGqlAppWebhook,
		Object:        testAppTenantMappingWebhookInput,
		CorrelationID: "",
	}

	testFormationConfigurationChangeInputWithSourceRuntimeAndTargetApp := fixFormationConfigurationChangeInput(TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, testCustomerTenantContext, fixConvertFAFromModel(faWithSourceRuntimeAndTargetApp), fixConvertFAFromModel(faWithSourceRuntimeAndTargetAppReverse))
	testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceRuntimeAndTargetApp := &webhookclient.NotificationRequest{
		Webhook:       *testGqlAppWebhook,
		Object:        testFormationConfigurationChangeInputWithSourceRuntimeAndTargetApp,
		CorrelationID: "",
	}

	testFormationConfigurationChangeInputWithSourceRtmCtxAndTargetApp := fixFormationConfigurationChangeInput(TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, testCustomerTenantContext, fixConvertFAFromModel(faWithSourceRuntimeCtxAndTargetApp), fixConvertFAFromModel(faWithSourceRuntimeCtxAndTargetAppReverse))
	testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceRtmCtxAndTargetApp := &webhookclient.NotificationRequest{
		Webhook:       *testGqlAppWebhook,
		Object:        testFormationConfigurationChangeInputWithSourceRtmCtxAndTargetApp,
		CorrelationID: "",
	}

	testFormationConfigurationChangeInputWithSourceAppAndTargetRuntime := fixFormationConfigurationChangeInput(TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, nil, testCustomerTenantContext, fixConvertFAFromModel(faWithSourceAppAndTargetRuntime), fixConvertFAFromModel(faWithSourceAppAndTargetRuntimeReverse))
	testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceAppAndTargetRuntime := &webhookclient.NotificationRequest{
		Webhook:       *testGqlRuntimeWebhook,
		Object:        testFormationConfigurationChangeInputWithSourceAppAndTargetRuntime,
		CorrelationID: "",
	}

	testFormationConfigurationChangeInputWithSourceAppAndTargetRtmCtx := fixFormationConfigurationChangeInput(TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, testCustomerTenantContext, fixConvertFAFromModel(faWithSourceAppCtxAndTargetRtmCtx), fixConvertFAFromModel(faWithSourceAppCtxAndTargetRtmCtxReverse))
	testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceAppAndTargetRtmCtx := &webhookclient.NotificationRequest{
		Webhook:       *testGqlRuntimeWebhook,
		Object:        testFormationConfigurationChangeInputWithSourceAppAndTargetRtmCtx,
		CorrelationID: "",
	}

	tenantObject := fixModelBusinessTenantMapping()

	testCases := []struct {
		name                    string
		formationAssignment     *model.FormationAssignment
		webhookRepo             func() *automock.WebhookRepository
		webhookDataInputBuilder func() *databuilderautomock.DataInputBuilder
		formationAssignmentRepo func() *automock.FormationAssignmentRepository
		webhookConverver        func() *automock.WebhookConverter
		tenantRepo              func() *automock.TenantRepository
		expectedNotification    *webhookclient.NotificationRequest
		expectedErrMsg          string
	}{
		{
			name:                "Error when formation assignment type is invalid",
			formationAssignment: faWithInvalidTypes,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithInvalidTypes.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithInvalidTypes.TenantID).Return(TntParentID, nil)
				return repo
			},
			expectedErrMsg: "Unknown formation assignment type:",
		},
		// application formation assignment notifications with source application
		{
			name:                "Successfully generate application notification when source type is application",
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(testAppWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestSource).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(faWithSourceAppAndTargetAppReverse, nil).Once()
				return faRepo
			},
			webhookConverver: func() *automock.WebhookConverter {
				webhookConv := &automock.WebhookConverter{}
				webhookConv.On("ToGraphQL", testAppWebhook).Return(testGqlAppWebhook, nil).Once()
				return webhookConv
			},
			expectedNotification: testAppNotificationReqWithTenantMappingType,
		},
		{
			name:                "Success when application webhook is not found",
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testNotFoundErr).Once()
				return webhookRepo
			},
		},
		{
			name:                "Error when getting tenant fails",
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(nil, testErr)
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting parent customer id fails",
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return("", testErr)
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting application webhook by ID and type",
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr).Once()
				return webhookRepo
			},
			expectedErrMsg: "while getting configuration changed webhook for runtime with ID:",
		},
		{
			name:                "Error when preparing app and app template with labels for source type application",
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(testAppWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(nil, nil, testErr).Once()
				return webhookDataInputBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing reverse app and app template with labels for source type application",
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(testAppWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestSource).Return(nil, nil, testErr).Once()
				return webhookDataInputBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting reverse formation assignment by source and target for source type application",
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(testAppWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestSource).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(nil, testErr).Once()
				return faRepo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when creating webhook request for source type application",
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(testAppWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestSource).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(faWithSourceAppAndTargetAppReverse, nil).Once()
				return faRepo
			},
			webhookConverver: func() *automock.WebhookConverter {
				webhookConv := &automock.WebhookConverter{}
				webhookConv.On("ToGraphQL", testAppWebhook).Return(nil, testErr).Once()
				return webhookConv
			},
			expectedErrMsg: testErr.Error(),
		},
		// application formation assignment notifications with source runtime
		{
			name:                "Successfully generate application notification when source type is runtime",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(testAppWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeAndRuntimeContextWithLabels", emptyCtx, TestTenantID, TestSource).Return(testRuntimeWithLabels, testRuntimeCtxWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(faWithSourceRuntimeAndTargetAppReverse, nil).Once()
				return faRepo
			},
			webhookConverver: func() *automock.WebhookConverter {
				webhookConv := &automock.WebhookConverter{}
				webhookConv.On("ToGraphQL", testAppWebhook).Return(testGqlAppWebhook, nil).Once()
				return webhookConv
			},
			expectedNotification: testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceRuntimeAndTargetApp,
		},
		{
			name:                "Error when getting tenant fails",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(nil, testErr)
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting parent customer id fails",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return("", testErr)
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing app and app template with labels for source type runtime",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(testAppWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(nil, nil, testErr).Once()
				return webhookDataInputBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing runtime and runtime context with labels for source type runtime",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(testAppWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeAndRuntimeContextWithLabels", emptyCtx, TestTenantID, TestSource).Return(nil, nil, testErr).Once()
				return webhookDataInputBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting reverse formation assignment by source and target for source type runtime",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(testAppWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeAndRuntimeContextWithLabels", emptyCtx, TestTenantID, TestSource).Return(testRuntimeWithLabels, testRuntimeCtxWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(nil, testErr).Once()
				return faRepo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when creating webhook request for source type runtime",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(testAppWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeAndRuntimeContextWithLabels", emptyCtx, TestTenantID, TestSource).Return(testRuntimeWithLabels, testRuntimeCtxWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(faWithSourceRuntimeAndTargetAppReverse, nil).Once()
				return faRepo
			},
			webhookConverver: func() *automock.WebhookConverter {
				webhookConv := &automock.WebhookConverter{}
				webhookConv.On("ToGraphQL", testAppWebhook).Return(nil, testErr).Once()
				return webhookConv
			},
			expectedErrMsg: testErr.Error(),
		},
		// application formation assignment notifications with source runtime context
		{
			name:                "Successfully generate application notification when source type is runtime context",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(testAppWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeContextWithLabels", emptyCtx, TestTenantID, TestSource).Return(testRuntimeCtxWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, testRuntimeID).Return(testRuntimeWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(faWithSourceRuntimeCtxAndTargetAppReverse, nil).Once()
				return faRepo
			},
			webhookConverver: func() *automock.WebhookConverter {
				webhookConv := &automock.WebhookConverter{}
				webhookConv.On("ToGraphQL", testAppWebhook).Return(testGqlAppWebhook, nil).Once()
				return webhookConv
			},
			expectedNotification: testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceRtmCtxAndTargetApp,
		},
		{
			name:                "Error when getting tenant fails",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(nil, testErr)
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting parent customer id fails",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return("", testErr)
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing app and app template with labels for source type runtime context",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(testAppWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(nil, nil, testErr).Once()
				return webhookDataInputBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing runtime context with labels for source type runtime context",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(testAppWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeContextWithLabels", emptyCtx, TestTenantID, TestSource).Return(nil, testErr).Once()
				return webhookDataInputBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing runtime with labels for source type runtime context",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(testAppWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeContextWithLabels", emptyCtx, TestTenantID, TestSource).Return(testRuntimeCtxWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, testRuntimeID).Return(nil, testErr).Once()
				return webhookDataInputBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting reverse formation assignment by source and target for source type runtime context",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(testAppWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeContextWithLabels", emptyCtx, TestTenantID, TestSource).Return(testRuntimeCtxWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, testRuntimeID).Return(testRuntimeWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(nil, testErr).Once()
				return faRepo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when creating webhook request for source type runtime",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(testAppWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeContextWithLabels", emptyCtx, TestTenantID, TestSource).Return(testRuntimeCtxWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, testRuntimeID).Return(testRuntimeWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(faWithSourceRuntimeCtxAndTargetAppReverse, nil).Once()
				return faRepo
			},
			webhookConverver: func() *automock.WebhookConverter {
				webhookConv := &automock.WebhookConverter{}
				webhookConv.On("ToGraphQL", testAppWebhook).Return(nil, testErr).Once()
				return webhookConv
			},
			expectedErrMsg: testErr.Error(),
		},
		// runtime formation assignment notifications with source application
		{
			name:                "Successfully generate runtime notification when source type is application",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(testRuntimeWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestSource).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testRuntimeWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(faWithSourceAppAndTargetRuntimeReverse, nil).Once()
				return faRepo
			},
			webhookConverver: func() *automock.WebhookConverter {
				webhookConv := &automock.WebhookConverter{}
				webhookConv.On("ToGraphQL", testRuntimeWebhook).Return(testGqlRuntimeWebhook, nil).Once()
				return webhookConv
			},
			expectedNotification: testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceAppAndTargetRuntime,
		},
		{
			name:                "Success when runtime webhook is not found",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testNotFoundErr).Once()
				return webhookRepo
			},
		},
		{
			name:                "Error when getting tenant fails",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(nil, testErr)
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting parent customer id fails",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return("", testErr)
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting runtime webhook by ID and type for runtime target",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr).Once()
				return webhookRepo
			},
			expectedErrMsg: "while getting configuration changed webhook for runtime with ID:",
		},
		{
			name:                "Error when source type is different than application for runtime target",
			formationAssignment: faWithSourceInvalidAndTargetRuntime,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceInvalidAndTargetRuntime.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceInvalidAndTargetRuntime.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(testRuntimeWebhook, nil).Once()
				return webhookRepo
			},
			expectedErrMsg: fmt.Sprintf("The formation assignmet with ID: %q and target type: %q has unsupported reverse(source) type: %q", TestID, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeRuntime),
		},
		{
			name:                "Error when preparing app and app template with labels for source type application and runtime target",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(testRuntimeWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestSource).Return(nil, nil, testErr).Once()
				return webhookDataInputBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing runtime with labels for source type application",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(testRuntimeWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestSource).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, TestTarget).Return(nil, testErr).Once()
				return webhookDataInputBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting reverse FA by application source and runtime target",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(testRuntimeWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestSource).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testRuntimeWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(nil, testErr).Once()
				return faRepo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when creating webhook request for source type application and runtime target",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(testRuntimeWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestSource).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testRuntimeWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(faWithSourceAppAndTargetRuntimeReverse, nil).Once()
				return faRepo
			},
			webhookConverver: func() *automock.WebhookConverter {
				webhookConv := &automock.WebhookConverter{}
				webhookConv.On("ToGraphQL", testRuntimeWebhook).Return(nil, testErr).Once()
				return webhookConv
			},
			expectedErrMsg: testErr.Error(),
		},
		// runtime context formation assignment notifications with source application
		{
			name:                "Successfully generate runtime context notification when source type is application",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, testRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(testRuntimeWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestSource).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, testRuntimeID).Return(testRuntimeWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeContextWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testRuntimeCtxWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(faWithSourceAppCtxAndTargetRtmCtxReverse, nil).Once()
				return faRepo
			},
			webhookConverver: func() *automock.WebhookConverter {
				webhookConv := &automock.WebhookConverter{}
				webhookConv.On("ToGraphQL", testRuntimeWebhook).Return(testGqlRuntimeWebhook, nil).Once()
				return webhookConv
			},
			expectedNotification: testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceAppAndTargetRtmCtx,
		},
		{
			name:                "Error when getting tenant fails",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(nil, testErr)
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting parent customer id fails",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return("", testErr)
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing runtime context with labels for source type application",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareRuntimeContextWithLabels", emptyCtx, TestTenantID, TestTarget).Return(nil, testErr).Once()
				return webhookDataInputBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Success when runtime webhook is not found for runtime context target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, testRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testNotFoundErr).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareRuntimeContextWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testRuntimeCtxWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
		},
		{
			name:                "Error when getting runtime webhook by ID and type for runtime context target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, testRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareRuntimeContextWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testRuntimeCtxWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			expectedErrMsg: "while getting configuration changed webhook for runtime with ID:",
		},
		{
			name:                "Error when source type is different than application for runtime context target",
			formationAssignment: faWithSourceInvalidAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceInvalidAndTargetRtmCtx.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceInvalidAndTargetRtmCtx.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, testRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(testRuntimeWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareRuntimeContextWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testRuntimeCtxWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			expectedErrMsg: fmt.Sprintf("The formation assignmet with ID: %q and target type: %q has unsupported reverse(source) type: %q", TestID, model.FormationAssignmentTypeRuntimeContext, model.FormationAssignmentTypeRuntimeContext),
		},
		{
			name:                "Error when preparing app and app template with labels for source type application and runtime ctx target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, testRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(testRuntimeWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestSource).Return(nil, nil, testErr).Once()
				webhookDataInputBuilder.On("PrepareRuntimeContextWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testRuntimeCtxWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing runtime with labels for source type application and runtime ctx target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, testRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(testRuntimeWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestSource).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, testRuntimeID).Return(nil, testErr).Once()
				webhookDataInputBuilder.On("PrepareRuntimeContextWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testRuntimeCtxWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting reverse FA by app source and runtime context target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, testRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(testRuntimeWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestSource).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, testRuntimeID).Return(testRuntimeWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeContextWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testRuntimeCtxWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(nil, testErr).Once()
				return faRepo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when creating webhook request for source type application and runtime context target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(tenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, testRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(testRuntimeWebhook, nil).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestSource).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, testRuntimeID).Return(testRuntimeWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeContextWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testRuntimeCtxWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(faWithSourceAppCtxAndTargetRtmCtxReverse, nil).Once()
				return faRepo
			},
			webhookConverver: func() *automock.WebhookConverter {
				webhookConv := &automock.WebhookConverter{}
				webhookConv.On("ToGraphQL", testRuntimeWebhook).Return(nil, testErr).Once()
				return webhookConv
			},
			expectedErrMsg: testErr.Error(),
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			faRepo := unusedFormationAssignmentRepository()
			if tCase.formationAssignmentRepo != nil {
				faRepo = tCase.formationAssignmentRepo()
			}

			webhookConv := unusedWebhookConverter()
			if tCase.webhookConverver != nil {
				webhookConv = tCase.webhookConverver()
			}

			webhookRepo := unusedWebhookRepo()
			if tCase.webhookRepo != nil {
				webhookRepo = tCase.webhookRepo()
			}

			tenantRepo := unusedTenantRepo()
			if tCase.tenantRepo != nil {
				tenantRepo = tCase.tenantRepo()
			}

			webhookDataInputBuilder := unusedWebhookDataInputBuilder()
			if tCase.webhookDataInputBuilder != nil {
				webhookDataInputBuilder = tCase.webhookDataInputBuilder()
			}

			defer mock.AssertExpectationsForObjects(t, faRepo, webhookConv, webhookRepo, tenantRepo, webhookDataInputBuilder)

			faNotificationSvc := formationassignment.NewFormationAssignmentNotificationService(faRepo, webhookConv, webhookRepo, tenantRepo, webhookDataInputBuilder)

			// WHEN
			notificationReq, err := faNotificationSvc.GenerateNotification(emptyCtx, tCase.formationAssignment)

			// THEN
			if tCase.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tCase.expectedErrMsg)
				require.Nil(t, tCase.expectedNotification)
			} else {
				require.NoError(t, err)
				require.Equal(t, tCase.expectedNotification, notificationReq)
			}
		})
	}
}
