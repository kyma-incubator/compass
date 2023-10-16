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

var (
	testRuntimeID     = "testRuntimeID"
	testAppTemplateID = "testAppTemplateID"

	testAppWebhook = &model.Webhook{
		ID:         "TestAppWebhookID",
		ObjectID:   TestSource,
		ObjectType: model.ApplicationWebhookReference,
	}

	testRuntimeWebhook = &model.Webhook{
		ID:         "TestRuntimeWebhookID",
		ObjectID:   TestSource,
		ObjectType: model.RuntimeWebhookReference,
	}

	testRuntimeWithLabels = &webhook.RuntimeWithLabels{
		Runtime: &model.Runtime{
			Name: "testRuntimeName",
		},
		Labels: testLabels,
	}
	testRuntimeCtxWithLabels = &webhook.RuntimeContextWithLabels{
		RuntimeContext: &model.RuntimeContext{
			ID:        "testRuntimeCtxID",
			RuntimeID: testRuntimeID,
			Key:       "testKey",
			Value:     "testValue",
		},
		Labels: testLabels,
	}

	testLabels        = map[string]string{"testLabelKey": "testLabelValue"}
	testAppWithLabels = &webhook.ApplicationWithLabels{
		Application: &model.Application{
			Name:                  "testAppName",
			ApplicationTemplateID: str.Ptr(testAppTemplateID),
		},
		Labels: testLabels,
	}
	testAppTemplateWithLabels = &webhook.ApplicationTemplateWithLabels{
		ApplicationTemplate: &model.ApplicationTemplate{
			ID:   testAppTemplateID,
			Name: "testAppTemplateName",
		},
		Labels: testLabels,
	}

	testGqlAppWebhook = &graphql.Webhook{
		ID:   testAppWebhook.ID,
		Type: graphql.WebhookType(testAppWebhook.Type),
	}

	testCustomerTenantContext = &webhook.CustomerTenantContext{
		CustomerID: TntParentID,
		AccountID:  str.Ptr(TestTenantID),
		Path:       nil,
	}

	faWithSourceAppAndTargetApp               = fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), TestConfigValueRawJSON, TestEmptyErrorValueRawJSON)
	faWithSourceAppAndTargetAppReverse        = fixReverseFormationAssignment(faWithSourceAppAndTargetApp)
	webhookFaWithSourceAppAndTargetApp        = convertFormationAssignmentFromModel(faWithSourceAppAndTargetApp)
	webhookFaWithSourceAppAndTargetAppReverse = convertFormationAssignmentFromModel(faWithSourceAppAndTargetAppReverse)

	testAppTenantMappingWebhookInput            = fixAppTenantMappingWebhookInput(TestFormationID, testAppWithLabels, testAppWithLabels, testAppTemplateWithLabels, testAppTemplateWithLabels, testCustomerTenantContext, fixConvertFAFromModel(faWithSourceAppAndTargetApp), fixConvertFAFromModel(faWithSourceAppAndTargetAppReverse))
	testAppNotificationReqWithTenantMappingType = &webhookclient.FormationAssignmentNotificationRequest{
		Webhook:       testGqlAppWebhook,
		Object:        testAppTenantMappingWebhookInput,
		CorrelationID: "",
	}

	applicationLabelInput = &model.LabelInput{
		Key:        appTypeLabelKey,
		ObjectID:   TestTarget,
		ObjectType: model.ApplicationLabelableObject,
	}
	applicationTypeLabel = &model.Label{Value: appSubtype}
	runtimeTypeLabel     = &model.Label{Value: rtmSubtype}
)

func Test_GenerateFormationAssignmentNotification(t *testing.T) {
	ctx := context.TODO()

	testNotFoundErr := apperrors.NewNotFoundError(resource.Webhook, TestTarget)
	faWithInvalidTypes := fixFormationAssignmentModel(TestConfigValueRawJSON)

	faWithSourceRuntimeAndTargetApp := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), TestConfigValueRawJSON, TestEmptyErrorValueRawJSON)
	faWithSourceRuntimeAndTargetAppReverse := fixReverseFormationAssignment(faWithSourceRuntimeAndTargetApp)

	faWithSourceRuntimeCtxAndTargetApp := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeRuntimeContext, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), TestConfigValueRawJSON, TestEmptyErrorValueRawJSON)
	faWithSourceRuntimeCtxAndTargetAppReverse := fixReverseFormationAssignment(faWithSourceRuntimeCtxAndTargetApp)

	faWithSourceAppAndTargetRuntime := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, string(model.ReadyAssignmentState), TestConfigValueRawJSON, TestEmptyErrorValueRawJSON)
	faWithSourceAppAndTargetRuntimeReverse := fixReverseFormationAssignment(faWithSourceAppAndTargetRuntime)

	faWithSourceInvalidAndTargetRuntime := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeRuntime, string(model.ReadyAssignmentState), TestConfigValueRawJSON, TestEmptyErrorValueRawJSON)
	faWithSourceInvalidAndTargetRtmCtx := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeRuntimeContext, model.FormationAssignmentTypeRuntimeContext, string(model.ReadyAssignmentState), TestConfigValueRawJSON, TestEmptyErrorValueRawJSON)

	faWithSourceAppCtxAndTargetRtmCtx := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntimeContext, string(model.ReadyAssignmentState), TestConfigValueRawJSON, TestEmptyErrorValueRawJSON)
	faWithSourceAppCtxAndTargetRtmCtxReverse := fixReverseFormationAssignment(faWithSourceAppCtxAndTargetRtmCtx)

	testGqlRuntimeWebhook := &graphql.Webhook{
		ID:   testRuntimeWebhook.ID,
		Type: graphql.WebhookType(testRuntimeWebhook.Type),
	}

	testCustomerTenantContextWithPath := &webhook.CustomerTenantContext{
		CustomerID: TntParentID,
		AccountID:  nil,
		Path:       str.Ptr(TestTenantID),
	}

	testAppTenantMappingWebhookInputWithTenantPath := fixAppTenantMappingWebhookInput(TestFormationID, testAppWithLabels, testAppWithLabels, testAppTemplateWithLabels, testAppTemplateWithLabels, testCustomerTenantContextWithPath, fixConvertFAFromModel(faWithSourceAppAndTargetApp), fixConvertFAFromModel(faWithSourceAppAndTargetAppReverse))
	testAppNotificationReqWithTenantMappingTypeWithTenantPath := &webhookclient.FormationAssignmentNotificationRequest{
		Webhook:       testGqlAppWebhook,
		Object:        testAppTenantMappingWebhookInputWithTenantPath,
		CorrelationID: "",
	}

	testFormationConfigurationChangeInputWithSourceRuntimeAndTargetApp := fixFormationConfigurationChangeInput(TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, testCustomerTenantContext, fixConvertFAFromModel(faWithSourceRuntimeAndTargetApp), fixConvertFAFromModel(faWithSourceRuntimeAndTargetAppReverse))
	testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceRuntimeAndTargetApp := &webhookclient.FormationAssignmentNotificationRequest{
		Webhook:       testGqlAppWebhook,
		Object:        testFormationConfigurationChangeInputWithSourceRuntimeAndTargetApp,
		CorrelationID: "",
	}

	testFormationConfigurationChangeInputWithSourceRtmCtxAndTargetApp := fixFormationConfigurationChangeInput(TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, testCustomerTenantContext, fixConvertFAFromModel(faWithSourceRuntimeCtxAndTargetApp), fixConvertFAFromModel(faWithSourceRuntimeCtxAndTargetAppReverse))
	testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceRtmCtxAndTargetApp := &webhookclient.FormationAssignmentNotificationRequest{
		Webhook:       testGqlAppWebhook,
		Object:        testFormationConfigurationChangeInputWithSourceRtmCtxAndTargetApp,
		CorrelationID: "",
	}

	testFormationConfigurationChangeInputWithSourceAppAndTargetRuntime := fixFormationConfigurationChangeInput(TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, nil, testCustomerTenantContext, fixConvertFAFromModel(faWithSourceAppAndTargetRuntime), fixConvertFAFromModel(faWithSourceAppAndTargetRuntimeReverse))
	testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceAppAndTargetRuntime := &webhookclient.FormationAssignmentNotificationRequest{
		Webhook:       testGqlRuntimeWebhook,
		Object:        testFormationConfigurationChangeInputWithSourceAppAndTargetRuntime,
		CorrelationID: "",
	}

	testFormationConfigurationChangeInputWithSourceAppAndTargetRtmCtx := fixFormationConfigurationChangeInput(TestFormationID, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, testCustomerTenantContext, fixConvertFAFromModel(faWithSourceAppCtxAndTargetRtmCtx), fixConvertFAFromModel(faWithSourceAppCtxAndTargetRtmCtxReverse))
	testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceAppAndTargetRtmCtx := &webhookclient.FormationAssignmentNotificationRequest{
		Webhook:       testGqlRuntimeWebhook,
		Object:        testFormationConfigurationChangeInputWithSourceAppAndTargetRtmCtx,
		CorrelationID: "",
	}

	formationOnlyWithFormationTemplateID := &model.Formation{FormationTemplateID: TestFormationTemplateID}
	details := &formationconstraint.GenerateFormationAssignmentNotificationOperationDetails{}

	var emptyRuntimeCtx *webhook.RuntimeContextWithLabels
	gaTenantObject := fixModelBusinessTenantMappingWithType(tnt.Account)
	rgTenantObject := fixModelBusinessTenantMappingWithType(tnt.ResourceGroup)

	testCases := []struct {
		name                    string
		formationAssignment     *model.FormationAssignment
		formationOperation      model.FormationOperation
		webhookRepo             func() *automock.WebhookRepository
		webhookDataInputBuilder func() *databuilderautomock.DataInputBuilder
		formationAssignmentRepo func() *automock.FormationAssignmentRepository
		formationRepo           func() *automock.FormationRepository
		notificationBuilder     func() *automock.NotificationBuilder
		tenantRepo              func() *automock.TenantRepository
		expectedNotification    *webhookclient.FormationAssignmentNotificationRequest
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
			formationOperation:  model.AssignFormation,
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
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeApplicationTenantMapping).Return(testAppWebhook, nil).Once()
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

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testAppTemplateWithLabels, testAppWithLabels, webhookFaWithSourceAppAndTargetApp, webhookFaWithSourceAppAndTargetAppReverse, testCustomerTenantContext, TestTenantID).Return(details, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, details, testAppWebhook).Return(testAppNotificationReqWithTenantMappingType, nil).Once()

				return notificationsBuilder
			},
			expectedNotification: testAppNotificationReqWithTenantMappingType,
		},
		{
			name:                "Successfully generate application notification when source type is application and tenant is resource group",
			formationAssignment: faWithSourceAppAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(rgTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeApplicationTenantMapping).Return(testAppWebhook, nil).Once()
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

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testAppTemplateWithLabels, testAppWithLabels, webhookFaWithSourceAppAndTargetApp, webhookFaWithSourceAppAndTargetAppReverse, testCustomerTenantContextWithPath, TestTenantID).Return(details, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, details, testAppWebhook).Return(testAppNotificationReqWithTenantMappingTypeWithTenantPath, nil).Once()

				return notificationsBuilder
			},
			expectedNotification: testAppNotificationReqWithTenantMappingTypeWithTenantPath,
		},
		{
			name:                "Success when application webhook is not found",
			formationAssignment: faWithSourceAppAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeApplicationTenantMapping).Return(nil, testNotFoundErr).Once()
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, testAppTemplateID, model.ApplicationTemplateWebhookReference, model.WebhookTypeApplicationTenantMapping).Return(nil, testNotFoundErr).Once()
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

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testAppTemplateWithLabels, testAppWithLabels, webhookFaWithSourceAppAndTargetApp, webhookFaWithSourceAppAndTargetAppReverse, testCustomerTenantContext, TestTenantID).Return(details, nil).Once()

				return notificationsBuilder
			},
		},
		{
			name:                "Error when getting formation by ID",
			formationAssignment: faWithSourceAppAndTargetApp,
			formationOperation:  model.AssignFormation,
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
			formationOperation:  model.AssignFormation,
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
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return("", testErr)
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting application template webhook by ID and type",
			formationAssignment: faWithSourceAppAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeApplicationTenantMapping).Return(nil, testNotFoundErr).Once()
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, testAppTemplateID, model.ApplicationTemplateWebhookReference, model.WebhookTypeApplicationTenantMapping).Return(nil, testErr).Once()

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

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testAppTemplateWithLabels, testAppWithLabels, webhookFaWithSourceAppAndTargetApp, webhookFaWithSourceAppAndTargetAppReverse, testCustomerTenantContext, TestTenantID).Return(details, nil).Once()

				return notificationsBuilder
			},
			expectedErrMsg: fmt.Sprintf("while listing %q webhooks for application template with ID: %q on behalf of application with ID: %q", model.WebhookTypeApplicationTenantMapping, testAppTemplateID, TestTarget),
		},
		{
			name:                "Error when preparing app and app template with labels for source type application",
			formationAssignment: faWithSourceAppAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(nil, nil, testErr).Once()
				return webhookDataInputBuilder
			},
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing reverse app and app template with labels for source type application",
			formationAssignment: faWithSourceAppAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestSource).Return(nil, nil, testErr).Once()
				return webhookDataInputBuilder
			},
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting reverse formation assignment by source and target for source type application",
			formationAssignment: faWithSourceAppAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
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
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing details",
			formationAssignment: faWithSourceAppAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
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

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testAppTemplateWithLabels, testAppWithLabels, webhookFaWithSourceAppAndTargetApp, webhookFaWithSourceAppAndTargetAppReverse, testCustomerTenantContext, TestTenantID).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error while building notification request",
			formationAssignment: faWithSourceAppAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeApplicationTenantMapping).Return(testAppWebhook, nil).Once()
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

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testAppTemplateWithLabels, testAppWithLabels, webhookFaWithSourceAppAndTargetApp, webhookFaWithSourceAppAndTargetAppReverse, testCustomerTenantContext, TestTenantID).Return(details, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, details, testAppWebhook).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting application webhook by ID and type",
			formationAssignment: faWithSourceAppAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceAppAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeApplicationTenantMapping).Return(nil, testErr).Once()
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

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testAppTemplateWithLabels, testAppWithLabels, webhookFaWithSourceAppAndTargetApp, webhookFaWithSourceAppAndTargetAppReverse, testCustomerTenantContext, TestTenantID).Return(details, nil).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		// application formation assignment notifications with source runtime
		{
			name:                "Successfully generate application notification when source type is runtime",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			formationOperation:  model.AssignFormation,
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
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, TestSource).Return(testRuntimeWithLabels, nil).Once()
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

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, emptyRuntimeCtx, convertFormationAssignmentFromModel(faWithSourceRuntimeAndTargetApp), convertFormationAssignmentFromModel(faWithSourceRuntimeAndTargetAppReverse), model.ApplicationResourceType, testCustomerTenantContext, TestTenantID).Return(details, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, details, testAppWebhook).Return(testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceRuntimeAndTargetApp, nil).Once()

				return notificationsBuilder
			},
			expectedNotification: testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceRuntimeAndTargetApp,
		},
		{
			name:                "Error when getting tenant fails",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			formationOperation:  model.AssignFormation,
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
			formationOperation:  model.AssignFormation,
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
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(nil, nil, testErr).Once()
				return webhookDataInputBuilder
			},
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing runtime and runtime context with labels for source type runtime",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, TestSource).Return(nil, testErr).Once()
				return webhookDataInputBuilder
			},
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting reverse formation assignment by source and target for source type runtime",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, TestSource).Return(testRuntimeWithLabels, nil).Once()
				return webhookDataInputBuilder
			},
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(nil, testErr).Once()
				return faRepo
			},
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing details for source type runtime",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, TestSource).Return(testRuntimeWithLabels, nil).Once()
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

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, emptyRuntimeCtx, convertFormationAssignmentFromModel(faWithSourceRuntimeAndTargetApp), convertFormationAssignmentFromModel(faWithSourceRuntimeAndTargetAppReverse), model.ApplicationResourceType, testCustomerTenantContext, TestTenantID).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Success when getting application webhook return not found error",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testNotFoundErr).Once()
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, testAppTemplateID, model.ApplicationTemplateWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testNotFoundErr).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, TestSource).Return(testRuntimeWithLabels, nil).Once()
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

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, emptyRuntimeCtx, convertFormationAssignmentFromModel(faWithSourceRuntimeAndTargetApp), convertFormationAssignmentFromModel(faWithSourceRuntimeAndTargetAppReverse), model.ApplicationResourceType, testCustomerTenantContext, TestTenantID).Return(details, nil).Once()

				return notificationsBuilder
			},
		},
		{
			name:                "Error when getting application webhook by ID and type",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr).Once()
				return webhookRepo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, TestSource).Return(testRuntimeWithLabels, nil).Once()
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

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, emptyRuntimeCtx, convertFormationAssignmentFromModel(faWithSourceRuntimeAndTargetApp), convertFormationAssignmentFromModel(faWithSourceRuntimeAndTargetAppReverse), model.ApplicationResourceType, testCustomerTenantContext, TestTenantID).Return(details, nil).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when building notification request for source type runtime",
			formationAssignment: faWithSourceRuntimeAndTargetApp,
			formationOperation:  model.AssignFormation,
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
				webhookDataInputBuilder.On("PrepareRuntimeWithLabels", emptyCtx, TestTenantID, TestSource).Return(testRuntimeWithLabels, nil).Once()
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

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, emptyRuntimeCtx, convertFormationAssignmentFromModel(faWithSourceRuntimeAndTargetApp), convertFormationAssignmentFromModel(faWithSourceRuntimeAndTargetAppReverse), model.ApplicationResourceType, testCustomerTenantContext, TestTenantID).Return(details, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, details, testAppWebhook).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		// application formation assignment notifications with source runtime context
		{
			name:                "Successfully generate application notification when source type is runtime context",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			formationOperation:  model.AssignFormation,
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

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, convertFormationAssignmentFromModel(faWithSourceRuntimeCtxAndTargetApp), convertFormationAssignmentFromModel(faWithSourceRuntimeCtxAndTargetAppReverse), model.ApplicationResourceType, testCustomerTenantContext, TestTenantID).Return(details, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, details, testAppWebhook).Return(testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceRtmCtxAndTargetApp, nil).Once()

				return notificationsBuilder
			},
			expectedNotification: testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceRtmCtxAndTargetApp,
		},
		{
			name:                "Error when getting tenant fails",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			formationOperation:  model.AssignFormation,
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
			formationOperation:  model.AssignFormation,
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
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(nil, nil, testErr).Once()
				return webhookDataInputBuilder
			},
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing runtime context with labels for source type runtime context",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookDataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				webhookDataInputBuilder := &databuilderautomock.DataInputBuilder{}
				webhookDataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", emptyCtx, TestTenantID, TestTarget).Return(testAppWithLabels, testAppTemplateWithLabels, nil).Once()
				webhookDataInputBuilder.On("PrepareRuntimeContextWithLabels", emptyCtx, TestTenantID, TestSource).Return(nil, testErr).Once()
				return webhookDataInputBuilder
			},
			formationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing runtime with labels for source type runtime context",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
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
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting reverse formation assignment by source and target for source type runtime context",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
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
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing details for source type runtime context",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
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

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, convertFormationAssignmentFromModel(faWithSourceRuntimeCtxAndTargetApp), convertFormationAssignmentFromModel(faWithSourceRuntimeCtxAndTargetAppReverse), model.ApplicationResourceType, testCustomerTenantContext, TestTenantID).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Success when getting application webhook by ID and type return not found error",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testNotFoundErr).Once()
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, testAppTemplateID, model.ApplicationTemplateWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testNotFoundErr).Once()

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

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, convertFormationAssignmentFromModel(faWithSourceRuntimeCtxAndTargetApp), convertFormationAssignmentFromModel(faWithSourceRuntimeCtxAndTargetAppReverse), model.ApplicationResourceType, testCustomerTenantContext, TestTenantID).Return(details, nil).Once()

				return notificationsBuilder
			},
		},
		{
			name:                "Error when getting application webhook by ID and type",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			formationOperation:  model.AssignFormation,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", emptyCtx, faWithSourceRuntimeCtxAndTargetApp.TenantID).Return(TntParentID, nil)
				return repo
			},
			webhookRepo: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", emptyCtx, TestTenantID, TestTarget, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr).Once()
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

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, convertFormationAssignmentFromModel(faWithSourceRuntimeCtxAndTargetApp), convertFormationAssignmentFromModel(faWithSourceRuntimeCtxAndTargetAppReverse), model.ApplicationResourceType, testCustomerTenantContext, TestTenantID).Return(details, nil).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when building notification request for source type runtime context",
			formationAssignment: faWithSourceRuntimeCtxAndTargetApp,
			formationOperation:  model.AssignFormation,
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

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, convertFormationAssignmentFromModel(faWithSourceRuntimeCtxAndTargetApp), convertFormationAssignmentFromModel(faWithSourceRuntimeCtxAndTargetAppReverse), model.ApplicationResourceType, testCustomerTenantContext, TestTenantID).Return(details, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, details, testAppWebhook).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		// runtime formation assignment notifications with source application
		{
			name:                "Successfully generate runtime notification when source type is application",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			formationOperation:  model.AssignFormation,
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

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, emptyRuntimeCtx, convertFormationAssignmentFromModel(faWithSourceAppAndTargetRuntime), convertFormationAssignmentFromModel(faWithSourceAppAndTargetRuntimeReverse), model.RuntimeResourceType, testCustomerTenantContext, TestTenantID).Return(details, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, details, testRuntimeWebhook).Return(testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceAppAndTargetRuntime, nil).Once()

				return notificationsBuilder
			},
			expectedNotification: testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceAppAndTargetRuntime,
		},
		{
			name:                "Success when runtime webhook is not found",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			formationOperation:  model.AssignFormation,
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
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
		},
		{
			name:                "Error when getting tenant fails",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			formationOperation:  model.AssignFormation,
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
			formationOperation:  model.AssignFormation,
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
			formationOperation:  model.AssignFormation,
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
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: "while getting configuration changed webhook for runtime with ID:",
		},
		{
			name:                "Return nil when source type is different than application for runtime target",
			formationAssignment: faWithSourceInvalidAndTargetRuntime,
			formationOperation:  model.AssignFormation,
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
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
		},
		{
			name:                "Error when preparing app and app template with labels for source type application and runtime target",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			formationOperation:  model.AssignFormation,
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
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing runtime with labels for source type application",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			formationOperation:  model.AssignFormation,
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
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting reverse FA by application source and runtime target",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			formationOperation:  model.AssignFormation,
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
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing details for source type application and runtime target",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			formationOperation:  model.AssignFormation,
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

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, emptyRuntimeCtx, convertFormationAssignmentFromModel(faWithSourceAppAndTargetRuntime), convertFormationAssignmentFromModel(faWithSourceAppAndTargetRuntimeReverse), model.RuntimeResourceType, testCustomerTenantContext, TestTenantID).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when building notification request for source type application and runtime target",
			formationAssignment: faWithSourceAppAndTargetRuntime,
			formationOperation:  model.AssignFormation,
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

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, emptyRuntimeCtx, convertFormationAssignmentFromModel(faWithSourceAppAndTargetRuntime), convertFormationAssignmentFromModel(faWithSourceAppAndTargetRuntimeReverse), model.RuntimeResourceType, testCustomerTenantContext, TestTenantID).Return(details, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, details, testRuntimeWebhook).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		// runtime context formation assignment notifications with source application
		{
			name:                "Successfully generate runtime context notification when source type is application",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			formationOperation:  model.AssignFormation,
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

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, convertFormationAssignmentFromModel(faWithSourceAppCtxAndTargetRtmCtx), convertFormationAssignmentFromModel(faWithSourceAppCtxAndTargetRtmCtxReverse), model.RuntimeContextResourceType, testCustomerTenantContext, TestTenantID).Return(details, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, details, testRuntimeWebhook).Return(testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceAppAndTargetRtmCtx, nil).Once()

				return notificationsBuilder
			},
			expectedNotification: testAppNotificationReqWithFormationConfigurationChangeTypeWithSourceAppAndTargetRtmCtx,
		},
		{
			name:                "Error when getting tenant fails",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			formationOperation:  model.AssignFormation,
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
			formationOperation:  model.AssignFormation,
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
			formationOperation:  model.AssignFormation,
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
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Success when runtime webhook is not found for runtime context target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			formationOperation:  model.AssignFormation,
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
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
		},
		{
			name:                "Error when getting runtime webhook by ID and type for runtime context target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			formationOperation:  model.AssignFormation,
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
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: "while getting configuration changed webhook for runtime with ID:",
		},
		{
			name:                "Return nil when source type is different than application for runtime context target",
			formationAssignment: faWithSourceInvalidAndTargetRtmCtx,
			formationOperation:  model.AssignFormation,
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
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
		},
		{
			name:                "Error when preparing app and app template with labels for source type application and runtime ctx target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			formationOperation:  model.AssignFormation,
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
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when preparing runtime with labels for source type application and runtime ctx target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			formationOperation:  model.AssignFormation,
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
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when getting reverse FA by app source and runtime context target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			formationOperation:  model.AssignFormation,
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
				repo.On("Get", ctx, TestFormationID, TestTenantID).Return(formationOnlyWithFormationTemplateID, nil).Once()
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when building details source type application and runtime context target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			formationOperation:  model.AssignFormation,
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

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, convertFormationAssignmentFromModel(faWithSourceAppCtxAndTargetRtmCtx), convertFormationAssignmentFromModel(faWithSourceAppCtxAndTargetRtmCtxReverse), model.RuntimeContextResourceType, testCustomerTenantContext, TestTenantID).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when building notification request for source type application and runtime context target",
			formationAssignment: faWithSourceAppCtxAndTargetRtmCtx,
			formationOperation:  model.AssignFormation,
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

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, TestFormationTemplateID, formation, testAppTemplateWithLabels, testAppWithLabels, testRuntimeWithLabels, testRuntimeCtxWithLabels, convertFormationAssignmentFromModel(faWithSourceAppCtxAndTargetRtmCtx), convertFormationAssignmentFromModel(faWithSourceAppCtxAndTargetRtmCtxReverse), model.RuntimeContextResourceType, testCustomerTenantContext, TestTenantID).Return(details, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, details, testRuntimeWebhook).Return(nil, testErr).Once()

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

			faNotificationSvc := formationassignment.NewFormationAssignmentNotificationService(faRepo, nil, webhookRepo, tenantRepo, webhookDataInputBuilder, formationRepo, notificationBuilder, nil, nil, "", "")
			defer mock.AssertExpectationsForObjects(t, faRepo, webhookRepo, tenantRepo, webhookDataInputBuilder)

			// WHEN
			notificationReq, err := faNotificationSvc.GenerateFormationAssignmentNotification(emptyCtx, tCase.formationAssignment, tCase.formationOperation)

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

func Test_PrepareDetailsForNotificationStatusReturned(t *testing.T) {
	emptyCtx = context.TODO()

	faWithTargetTypeApplication := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), nil, nil)
	reverseFaWithTargetTypeApplication := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestTarget, TestSource, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), nil, nil)

	faWithTargetTypeRuntime := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeRuntime, string(model.ReadyAssignmentState), nil, nil)
	reverseFaWithTargetTypeRuntime := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestTarget, TestSource, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeRuntime, string(model.ReadyAssignmentState), nil, nil)

	faWithTargetTypeRuntimeCtx := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeRuntimeContext, model.FormationAssignmentTypeRuntimeContext, string(model.ReadyAssignmentState), nil, nil)
	reverseFaWithTargetTypeRuntimeCtx := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestTarget, TestSource, model.FormationAssignmentTypeRuntimeContext, model.FormationAssignmentTypeRuntimeContext, string(model.ReadyAssignmentState), nil, nil)

	runtimeLabelInput := &model.LabelInput{
		Key:        rtmTypeLabelKey,
		ObjectID:   TestTarget,
		ObjectType: model.RuntimeLabelableObject,
	}

	expectedDetailsForApp := fixNotificationStatusReturnedDetails(model.ApplicationResourceType, appSubtype, faWithTargetTypeApplication, reverseFaWithTargetTypeApplication, formationconstraint.JoinPointLocation{})

	testCases := []struct {
		name                    string
		formationAssignment     *model.FormationAssignment
		formationAssignmentRepo func() *automock.FormationAssignmentRepository
		formationRepo           func() *automock.FormationRepository
		labelSvc                func() *automock.LabelService
		runtimeCtxRepo          func() *automock.RuntimeContextRepository
		expectedDetails         *formationconstraint.NotificationStatusReturnedOperationDetails
		expectedErrMsg          string
	}{
		{
			name:                "Success for FA with target type application",
			formationAssignment: faWithTargetTypeApplication,
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(reverseFaWithTargetTypeApplication, nil).Once()
				return faRepo
			},
			formationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", emptyCtx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return formationRepo
			},
			labelSvc: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetLabel", emptyCtx, TestTenantID, applicationLabelInput).Return(applicationTypeLabel, nil).Once()
				return lblSvc
			},
			expectedDetails: expectedDetailsForApp,
		},
		{
			name:                "Success for FA with target type runtime",
			formationAssignment: faWithTargetTypeRuntime,
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(reverseFaWithTargetTypeRuntime, nil).Once()
				return faRepo
			},
			formationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", emptyCtx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return formationRepo
			},
			labelSvc: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetLabel", emptyCtx, TestTenantID, runtimeLabelInput).Return(runtimeTypeLabel, nil).Once()
				return lblSvc
			},
			expectedDetails: fixNotificationStatusReturnedDetails(model.RuntimeResourceType, rtmSubtype, faWithTargetTypeRuntime, reverseFaWithTargetTypeRuntime, formationconstraint.JoinPointLocation{}),
		},
		{
			name:                "Success for FA with target type runtime context",
			formationAssignment: faWithTargetTypeRuntimeCtx,
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(reverseFaWithTargetTypeRuntimeCtx, nil).Once()
				return faRepo
			},
			formationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", emptyCtx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return formationRepo
			},
			labelSvc: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetLabel", emptyCtx, TestTenantID, runtimeLabelInput).Return(runtimeTypeLabel, nil).Once()
				return lblSvc
			},
			runtimeCtxRepo: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetByID", emptyCtx, TestTenantID, TestTarget).Return(&model.RuntimeContext{RuntimeID: TestTarget}, nil).Once()
				return rtmCtxRepo
			},
			expectedDetails: fixNotificationStatusReturnedDetails(model.RuntimeContextResourceType, rtmSubtype, faWithTargetTypeRuntimeCtx, reverseFaWithTargetTypeRuntimeCtx, formationconstraint.JoinPointLocation{}),
		},
		{
			name:                "Success for application when there is no reverse fa",
			formationAssignment: faWithTargetTypeApplication,
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(nil, notFoundError).Once()
				return faRepo
			},
			formationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", emptyCtx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return formationRepo
			},
			labelSvc: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetLabel", emptyCtx, TestTenantID, applicationLabelInput).Return(applicationTypeLabel, nil).Once()
				return lblSvc
			},
			expectedDetails: fixNotificationStatusReturnedDetails(model.ApplicationResourceType, appSubtype, faWithTargetTypeApplication, nil, formationconstraint.JoinPointLocation{}),
		},
		{
			name:                "Error when can't get reverse fa",
			formationAssignment: faWithTargetTypeApplication,
			formationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				faRepo := &automock.FormationAssignmentRepository{}
				faRepo.On("GetReverseBySourceAndTarget", emptyCtx, TestTenantID, TestFormationID, TestSource, TestTarget).Return(nil, testErr).Once()
				return faRepo
			},
			formationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", emptyCtx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return formationRepo
			},
			labelSvc: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetLabel", emptyCtx, TestTenantID, applicationLabelInput).Return(applicationTypeLabel, nil).Once()
				return lblSvc
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when can't get formation",
			formationAssignment: faWithTargetTypeApplication,
			formationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", emptyCtx, TestFormationID, TestTenantID).Return(nil, testErr).Once()
				return formationRepo
			},
			labelSvc: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetLabel", emptyCtx, TestTenantID, applicationLabelInput).Return(applicationTypeLabel, nil).Once()
				return lblSvc
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when can't get application label",
			formationAssignment: faWithTargetTypeApplication,
			labelSvc: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetLabel", emptyCtx, TestTenantID, applicationLabelInput).Return(nil, testErr).Once()
				return lblSvc
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when can't get runtime label when target type is runtime",
			formationAssignment: faWithTargetTypeRuntime,
			labelSvc: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetLabel", emptyCtx, TestTenantID, runtimeLabelInput).Return(nil, testErr).Once()
				return lblSvc
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when can't get runtime label when target type is runtime context",
			formationAssignment: faWithTargetTypeRuntimeCtx,
			labelSvc: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetLabel", emptyCtx, TestTenantID, runtimeLabelInput).Return(nil, testErr).Once()
				return lblSvc
			},
			runtimeCtxRepo: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetByID", emptyCtx, TestTenantID, TestTarget).Return(&model.RuntimeContext{RuntimeID: TestTarget}, nil).Once()
				return rtmCtxRepo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when can't get runtime context",
			formationAssignment: faWithTargetTypeRuntimeCtx,
			runtimeCtxRepo: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetByID", emptyCtx, TestTenantID, TestTarget).Return(nil, testErr).Once()
				return rtmCtxRepo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                "Error when the object type is unknown",
			formationAssignment: &model.FormationAssignment{TargetType: "unknown"},
			expectedErrMsg:      fmt.Sprintf("unknown object type %q", "unknown"),
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			faRepo := unusedFormationAssignmentRepository()
			if tCase.formationAssignmentRepo != nil {
				faRepo = tCase.formationAssignmentRepo()
			}
			formationRepo := unusedFormationRepo()
			if tCase.formationRepo != nil {
				formationRepo = tCase.formationRepo()
			}
			labelSvc := &automock.LabelService{}
			if tCase.labelSvc != nil {
				labelSvc = tCase.labelSvc()
			}
			rtmCtxSvc := &automock.RuntimeContextRepository{}
			if tCase.runtimeCtxRepo != nil {
				rtmCtxSvc = tCase.runtimeCtxRepo()
			}

			defer mock.AssertExpectationsForObjects(t, faRepo, formationRepo, labelSvc, rtmCtxSvc)

			faNotificationSvc := formationassignment.NewFormationAssignmentNotificationService(faRepo, nil, nil, nil, nil, formationRepo, nil, rtmCtxSvc, labelSvc, rtmTypeLabelKey, appTypeLabelKey)

			// WHEN
			notificationReq, err := faNotificationSvc.PrepareDetailsForNotificationStatusReturned(emptyCtx, TestTenantID, tCase.formationAssignment, model.AssignFormation)

			// THEN
			if tCase.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tCase.expectedErrMsg)
				require.Nil(t, tCase.expectedDetails)
			} else {
				require.NoError(t, err)
				require.Equal(t, tCase.expectedDetails, notificationReq)
			}
		})
	}
}

func Test_GenerateFormationAssignmentNotificationExt(t *testing.T) {
	faRequestMapping := &formationassignment.FormationAssignmentRequestMapping{
		Request:             testAppNotificationReqWithTenantMappingType,
		FormationAssignment: faWithSourceAppAndTargetApp,
	}
	reverseFaRequestMapping := &formationassignment.FormationAssignmentRequestMapping{
		Request:             testAppNotificationReqWithTenantMappingType,
		FormationAssignment: faWithSourceAppAndTargetAppReverse,
	}

	expectedExtNotificationReq := &webhookclient.FormationAssignmentNotificationRequestExt{
		FormationAssignmentNotificationRequest: faRequestMapping.Request,
		Operation:                              model.AssignFormation,
		FormationAssignment:                    faWithSourceAppAndTargetApp,
		ReverseFormationAssignment:             faWithSourceAppAndTargetAppReverse,
		Formation:                              formation,
		TargetSubtype:                          appSubtype,
	}

	var testCases = []struct {
		name                       string
		faRequestMapping           *formationassignment.FormationAssignmentRequestMapping
		reverseFaRequestMapping    *formationassignment.FormationAssignmentRequestMapping
		formationRepo              func() *automock.FormationRepository
		labelSvc                   func() *automock.LabelService
		expectedExtNotificationReq *webhookclient.FormationAssignmentNotificationRequestExt
		expectedErrMsg             string
	}{
		{
			name:                    "Success",
			faRequestMapping:        faRequestMapping,
			reverseFaRequestMapping: reverseFaRequestMapping,
			formationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", emptyCtx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return formationRepo
			},
			labelSvc: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetLabel", emptyCtx, TestTenantID, applicationLabelInput).Return(applicationTypeLabel, nil).Once()
				return lblSvc
			},
			expectedExtNotificationReq: expectedExtNotificationReq,
		},
		{
			name:             "Success when there is no reverse fa",
			faRequestMapping: faRequestMapping,
			formationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", emptyCtx, TestFormationID, TestTenantID).Return(formation, nil).Once()
				return formationRepo
			},
			labelSvc: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetLabel", emptyCtx, TestTenantID, applicationLabelInput).Return(applicationTypeLabel, nil).Once()
				return lblSvc
			},
			expectedExtNotificationReq: &webhookclient.FormationAssignmentNotificationRequestExt{
				FormationAssignmentNotificationRequest: faRequestMapping.Request,
				Operation:                              model.AssignFormation,
				FormationAssignment:                    faWithSourceAppAndTargetApp,
				ReverseFormationAssignment:             nil,
				Formation:                              formation,
				TargetSubtype:                          appSubtype,
			},
		},
		{
			name:                    "Returns error when can't get formation",
			faRequestMapping:        faRequestMapping,
			reverseFaRequestMapping: reverseFaRequestMapping,
			formationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", emptyCtx, TestFormationID, TestTenantID).Return(nil, testErr).Once()
				return formationRepo
			},
			labelSvc: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetLabel", emptyCtx, TestTenantID, applicationLabelInput).Return(applicationTypeLabel, nil).Once()
				return lblSvc
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name:                    "Returns error when can't get subtype",
			faRequestMapping:        faRequestMapping,
			reverseFaRequestMapping: reverseFaRequestMapping,
			labelSvc: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetLabel", emptyCtx, TestTenantID, applicationLabelInput).Return(nil, testErr).Once()
				return lblSvc
			},
			expectedErrMsg: testErr.Error(),
		},
	}
	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			formationRepo := unusedFormationRepo()
			if tCase.formationRepo != nil {
				formationRepo = tCase.formationRepo()
			}
			labelSvc := &automock.LabelService{}
			if tCase.labelSvc != nil {
				labelSvc = tCase.labelSvc()
			}

			defer mock.AssertExpectationsForObjects(t, formationRepo, labelSvc)

			faNotificationSvc := formationassignment.NewFormationAssignmentNotificationService(nil, nil, nil, nil, nil, formationRepo, nil, nil, labelSvc, rtmTypeLabelKey, appTypeLabelKey)

			// WHEN
			extNotificationRequest, err := faNotificationSvc.GenerateFormationAssignmentNotificationExt(emptyCtx, tCase.faRequestMapping, tCase.reverseFaRequestMapping, model.AssignFormation)

			// THEN
			if tCase.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tCase.expectedErrMsg)
				require.Nil(t, tCase.expectedExtNotificationReq)
			} else {
				require.NoError(t, err)
				require.Equal(t, tCase.expectedExtNotificationReq, extNotificationRequest)
			}
		})
	}
}
