package formationassignment_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	databuilderautomock "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder/automock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	tnt "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_GenerateNotification(t *testing.T) {
	testRuntimeID := "testRuntimeID"

	ctx := context.TODO()

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
		CustomerID: TntParentID,
		AccountID:  str.Ptr(TestTenantID),
		Path:       nil,
	}

	testCustomerTenantContextWithPath := &webhook.CustomerTenantContext{
		CustomerID: TntParentID,
		AccountID:  nil,
		Path:       str.Ptr(TestTenantID),
	}

	testAppTenantMappingWebhookInputWithTenantPath := fixAppTenantMappingWebhookInput(TestFormationID, testAppWithLabels, testAppWithLabels, testAppTemplateWithLabels, testAppTemplateWithLabels, testCustomerTenantContextWithPath, fixConvertFAFromModel(faWithSourceAppAndTargetApp), fixConvertFAFromModel(faWithSourceAppAndTargetAppReverse))
	testAppNotificationReqWithTenantMappingTypeWithTenantPath := &webhookclient.NotificationRequest{
		Webhook:       *testGqlAppWebhook,
		Object:        testAppTenantMappingWebhookInputWithTenantPath,
		CorrelationID: "",
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

	formation := &model.Formation{FormationTemplateID: TestFormationTemplateID}
	details := &formationconstraint.GenerateNotificationOperationDetails{}

	var emptyRuntimeCtx *webhook.RuntimeContextWithLabels
	gaTenantObject := fixModelBusinessTenantMappingWithType(tnt.Account)
	rgTenantObject := fixModelBusinessTenantMappingWithType(tnt.ResourceGroup)

	testCases := []struct {
		name                    string
		formationAssignment     *model.FormationAssignment
		webhookRepo             func() *automock.WebhookRepository
		webhookDataInputBuilder func() *databuilderautomock.DataInputBuilder
		formationAssignmentRepo func() *automock.FormationAssignmentRepository
		formationRepo           func() *automock.FormationRepository
		notificationBuilder     func() *automock.NotificationBuilder
		tenantRepo              func() *automock.TenantRepository
		expectedNotification    *webhookclient.NotificationRequest
		expectedErrMsg          string
	}{
		{
			name: "Error when formation assignment type is invalid",
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(nil, nil).Once()
				return repo
			},
			formationAssignment: faWithInvalidTypes,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithInvalidTypes.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithInvalidTypes.TenantID).Return(TntParentID, nil)
				return repo
			},
			expectedErrMsg: "Unknown formation assignment type:",
		},
		// application formation assignment notifications with source application
		{
			name:                "Successfully generate application notification when source type is application and tenant is global account",
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			notificationBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testAppTemplateWithLabels, testAppWithLabels, convertFormationAssignmentFromModel(faWithSourceAppAndTargetApp), convertFormationAssignmentFromModel(faWithSourceAppAndTargetAppReverse), testCustomerTenantContext).Return(details, nil).Once()
				notificationsBuilder.On("BuildNotificationRequest", ctx, TestFormationTemplateID, details, testAppWebhook).Return(testAppNotificationReqWithTenantMappingType, nil).Once()

				return notificationsBuilder
			},
			expectedNotification: testAppNotificationReqWithTenantMappingType,
		},
		{
			name:                "Successfully generate application notification when source type is application and tenant is resource group",
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(rgTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			notificationBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testAppTemplateWithLabels, testAppWithLabels, convertFormationAssignmentFromModel(faWithSourceAppAndTargetApp), convertFormationAssignmentFromModel(faWithSourceAppAndTargetAppReverse), testCustomerTenantContextWithPath).Return(details, nil).Once()
				notificationsBuilder.On("BuildNotificationRequest", ctx, TestFormationTemplateID, details, testAppWebhook).Return(testAppNotificationReqWithTenantMappingTypeWithTenantPath, nil).Once()

				return notificationsBuilder
			},
			expectedNotification: testAppNotificationReqWithTenantMappingTypeWithTenantPath,
		},
		{
			name:                "Success when application webhook is not found",
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testNotFoundErr).Once()
				return webhookRepo
			},
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
		},
		{
			name:                "Error when getting formation by ID",
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(nil, testErr).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
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
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr).Once()
				return webhookRepo
			},
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: "while getting configuration changed webhook for runtime with ID:",
		},
		{
			name:                "Error when preparing app and app template with labels for source type application",
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing reverse app and app template with labels for source type application",
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting reverse formation assignment by source and target for source type application",
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing details",
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			notificationBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testAppTemplateWithLabels, testAppWithLabels, convertFormationAssignmentFromModel(faWithSourceAppAndTargetApp), convertFormationAssignmentFromModel(faWithSourceAppAndTargetAppReverse), testCustomerTenantContext).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error while building notification request",
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			notificationBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testAppTemplateWithLabels, testAppWithLabels, convertFormationAssignmentFromModel(faWithSourceAppAndTargetApp), convertFormationAssignmentFromModel(faWithSourceAppAndTargetAppReverse), testCustomerTenantContext).Return(details, nil).Once()
				notificationsBuilder.On("BuildNotificationRequest", ctx, TestFormationTemplateID, details, testAppWebhook).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		// application formation assignment notifications with source runtime
		{
			name:                "Successfully generate application notification when source type is runtime",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			notificationBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, convertFormationAssignmentFromModel(faWithSourceRuntimeAndTargetApp), convertFormationAssignmentFromModel(faWithSourceRuntimeAndTargetAppReverse), model.ApplicationResourceType, testCustomerTenantContext).Return(details, nil).Once()
				notificationsBuilder.On("BuildNotificationRequest", ctx, TestFormationTemplateID, details, testAppWebhook).Return(testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceRuntimeAndTargetApp, nil).Once()

				return notificationsBuilder
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
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing runtime and runtime context with labels for source type runtime",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting reverse formation assignment by source and target for source type runtime",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing details for source type runtime",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			notificationBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, convertFormationAssignmentFromModel(faWithSourceRuntimeAndTargetApp), convertFormationAssignmentFromModel(faWithSourceRuntimeAndTargetAppReverse), model.ApplicationResourceType, testCustomerTenantContext).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when building notification request for source type runtime",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			notificationBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, convertFormationAssignmentFromModel(faWithSourceRuntimeAndTargetApp), convertFormationAssignmentFromModel(faWithSourceRuntimeAndTargetAppReverse), model.ApplicationResourceType, testCustomerTenantContext).Return(details, nil).Once()
				notificationsBuilder.On("BuildNotificationRequest", ctx, TestFormationTemplateID, details, testAppWebhook).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		// application formation assignment notifications with source runtime context
		{
			name:                "Successfully generate application notification when source type is runtime context",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			notificationBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, convertFormationAssignmentFromModel(faWithSourceRuntimeCtxAndTargetApp), convertFormationAssignmentFromModel(faWithSourceRuntimeCtxAndTargetAppReverse), model.ApplicationResourceType, testCustomerTenantContext).Return(details, nil).Once()
				notificationsBuilder.On("BuildNotificationRequest", ctx, TestFormationTemplateID, details, testAppWebhook).Return(testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceRtmCtxAndTargetApp, nil).Once()

				return notificationsBuilder
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
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing runtime context with labels for source type runtime context",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing runtime with labels for source type runtime context",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting reverse formation assignment by source and target for source type runtime context",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing details for source type runtime context",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			notificationBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, convertFormationAssignmentFromModel(faWithSourceRuntimeCtxAndTargetApp), convertFormationAssignmentFromModel(faWithSourceRuntimeCtxAndTargetAppReverse), model.ApplicationResourceType, testCustomerTenantContext).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when building notification request for source type runtime context",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(gaTenantObject, nil)
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
				webhookDataInputBuilder.On("PrepareRuntimeContextWithLabels", emptyCtx, TestTenantID, TestSource).Return(testRuntimeCtxWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, testRuntimeID).Return(testRuntimeWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(faWithSourceRuntimeCtxAndTargetAppReverse, nil).Once()
				return faRepo
			},
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			notificationBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, convertFormationAssignmentFromModel(faWithSourceRuntimeCtxAndTargetApp), convertFormationAssignmentFromModel(faWithSourceRuntimeCtxAndTargetAppReverse), model.ApplicationResourceType, testCustomerTenantContext).Return(details, nil).Once()
				notificationsBuilder.On("BuildNotificationRequest", ctx, TestFormationTemplateID, details, testAppWebhook).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		// runtime formation assignment notifications with source application
		{
			name:                "Successfully generate runtime notification when source type is application",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			notificationBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, emptyRuntimeCtx, convertFormationAssignmentFromModel(faWithSourceAppAndTargetRuntime), convertFormationAssignmentFromModel(faWithSourceAppAndTargetRuntimeReverse), model.RuntimeResourceType, testCustomerTenantContext).Return(details, nil).Once()
				notificationsBuilder.On("BuildNotificationRequest", ctx, TestFormationTemplateID, details, testRuntimeWebhook).Return(testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceAppAndTargetRuntime, nil).Once()

				return notificationsBuilder
			},
			expectedNotification: testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceAppAndTargetRuntime,
		},
		{
			name:                "Success when runtime webhook is not found",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testNotFoundErr).Once()
				return webhookRepo
			},
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
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
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(gaTenantObject, nil)
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
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr).Once()
				return webhookRepo
			},
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: "while getting configuration changed webhook for runtime with ID:",
		},
		{
			name:                "Error when source type is different than application for runtime target",
			formationAssignment: faWithSourceInvalidAndTargetRuntime,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceInvalidAndTargetRuntime.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceInvalidAndTargetRuntime.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(testRuntimeWebhook, nil).Once()
				return webhookRepo
			},
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: fmt.Sprintf("The formation assignmet with ID: %q and target type: %q has unsupported reverse(source) type: %q", TestID, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeRuntime),
		},
		{
			name:                "Error when preparing app and app template with labels for source type application and runtime target",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing runtime with labels for source type application",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting reverse FA by application source and runtime target",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing details for source type application and runtime target",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetRuntime.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			notificationBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, emptyRuntimeCtx, convertFormationAssignmentFromModel(faWithSourceAppAndTargetRuntime), convertFormationAssignmentFromModel(faWithSourceAppAndTargetRuntimeReverse), model.RuntimeResourceType, testCustomerTenantContext).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when building notification request for source type application and runtime target",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(TntParentID, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			notificationBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, emptyRuntimeCtx, convertFormationAssignmentFromModel(faWithSourceAppAndTargetRuntime), convertFormationAssignmentFromModel(faWithSourceAppAndTargetRuntimeReverse), model.RuntimeResourceType, testCustomerTenantContext).Return(details, nil).Once()
				notificationsBuilder.On("BuildNotificationRequest", ctx, TestFormationTemplateID, details, testRuntimeWebhook).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		// runtime context formation assignment notifications with source application
		{
			name:                "Successfully generate runtime context notification when source type is application",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			notificationBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, convertFormationAssignmentFromModel(faWithSourceAppCtxAndTargetRtmCtx), convertFormationAssignmentFromModel(faWithSourceAppCtxAndTargetRtmCtxReverse), model.RuntimeContextResourceType, testCustomerTenantContext).Return(details, nil).Once()
				notificationsBuilder.On("BuildNotificationRequest", ctx, TestFormationTemplateID, details, testRuntimeWebhook).Return(testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceAppAndTargetRtmCtx, nil).Once()

				return notificationsBuilder
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
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(gaTenantObject, nil)
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
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareRuntimeContextWithLabels", emptyCtx, TestTenantID, TestTarget).Return(nil, testErr).Once()
				return webhookDataInputBuilder
			},
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Success when runtime webhook is not found for runtime context target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
		},
		{
			name:                "Error when getting runtime webhook by ID and type for runtime context target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: "while getting configuration changed webhook for runtime with ID:",
		},
		{
			name:                "Error when source type is different than application for runtime context target",
			formationAssignment: faWithSourceInvalidAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceInvalidAndTargetRtmCtx.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: fmt.Sprintf("The formation assignmet with ID: %q and target type: %q has unsupported reverse(source) type: %q", TestID, model.FormationAssignmentTypeRuntimeContext, model.FormationAssignmentTypeRuntimeContext),
		},
		{
			name:                "Error when preparing app and app template with labels for source type application and runtime ctx target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing runtime with labels for source type application and runtime ctx target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting reverse FA by app source and runtime context target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when building details source type application and runtime context target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppCtxAndTargetRtmCtx.TenantID).Return(gaTenantObject, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			notificationBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, convertFormationAssignmentFromModel(faWithSourceAppCtxAndTargetRtmCtx), convertFormationAssignmentFromModel(faWithSourceAppCtxAndTargetRtmCtxReverse), model.RuntimeContextResourceType, testCustomerTenantContext).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when building notification request for source type application and runtime context target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(TntParentID, nil)
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
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return repo
			},
			notificationBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, convertFormationAssignmentFromModel(faWithSourceAppCtxAndTargetRtmCtx), convertFormationAssignmentFromModel(faWithSourceAppCtxAndTargetRtmCtxReverse), model.RuntimeContextResourceType, testCustomerTenantContext).Return(details, nil).Once()
				notificationsBuilder.On("BuildNotificationRequest", ctx, TestFormationTemplateID, details, testRuntimeWebhook).Return(nil, testErr).Once()

				return notificationsBuilder
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

			formationRepo := unusedFormationRepo()
			if tCase.formationRepo != nil {
				formationRepo = tCase.formationRepo()
			}

			notificationBuilder := unusedNotificationBuilder()
			if tCase.notificationBuilder != nil {
				notificationBuilder = tCase.notificationBuilder()
			}

			defer mock.AssertExpectationsForObjects(t, faRepo, webhookRepo, webhookDataInputBuilder, formationRepo, notificationBuilder)

			faNotificationSvc := formationassignment.NewFormationAssignmentNotificationService(faRepo, nil, webhookRepo, tenantRepo, webhookDataInputBuilder, formationRepo, notificationBuilder)
			defer mock.AssertExpectationsForObjects(t, faRepo, webhookRepo, tenantRepo, webhookDataInputBuilder)

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
