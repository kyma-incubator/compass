package formation_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	databuilderautomock "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder/automock"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_GenerateNotificationsAboutRuntimeAndRuntimeContextForTheApplicationThatIsAssigned(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	expectedFormation := model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            TntInternalID,
	}

	appTemplateWithLabels := &webhook.ApplicationTemplateWithLabels{
		ApplicationTemplate: fixApplicationTemplateModel(),
		Labels:              fixApplicationTemplateLabelsMap(),
	}

	appWithLabels := &webhook.ApplicationWithLabels{
		Application: fixApplicationModel(ApplicationID),
		Labels:      fixApplicationLabelsMap(),
	}

	appWithLabelsWithoutTemplate := &webhook.ApplicationWithLabels{
		Application: fixApplicationModelWithoutTemplate(ApplicationID),
		Labels:      fixApplicationLabelsMap(),
	}

	testCases := []struct {
		Name                  string
		ApplicationRepoFN     func() *automock.ApplicationRepository
		WebhookRepoFN         func() *automock.WebhookRepository
		WebhookClientFN       func() *automock.WebhookClient
		DataInputBuilder      func() *databuilderautomock.DataInputBuilder
		NotificationsBuilder  func() *automock.NotificationBuilder
		CustomerTenantContext *webhook.CustomerTenantContext
		ObjectID              string
		OperationType         model.FormationOperation
		InputFormation        model.Formation
		ExpectedRequests      []*webhookclient.FormationAssignmentNotificationRequest
		ExpectedErrMessage    string
	}{
		{
			Name: "success when generating notifications for application about runtimes and runtime contexts in formation",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimesAndRuntimeContextsMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(runtimesMapping, runtimeContextsMapping, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithLabels, runtimeCtx2WithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(appNotificationWithRtmCtxRtmIDAndTemplate, nil).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithRtmCtxWithLabels, runtimeCtxWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(appNotificationWithRtmCtxAndTemplate, nil).Once()

				return notificationsBuilder
			},
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				appNotificationWithRtmCtxRtmIDAndTemplate,
				appNotificationWithRtmCtxAndTemplate,
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "",
		},
		{
			Name: "success when using webhook from application template",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Application, ApplicationID)).Once()
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, ApplicationTemplateID, model.ApplicationTemplateWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationTemplateID, model.ApplicationTemplateWebhookReference), nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimesAndRuntimeContextsMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(runtimesMapping, runtimeContextsMapping, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithLabels, runtimeCtx2WithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, ApplicationTemplateID, model.ApplicationTemplateWebhookReference)).Return(appNotificationWithRtmCtxRtmIDAndTemplate, nil).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithRtmCtxWithLabels, runtimeCtxWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, ApplicationTemplateID, model.ApplicationTemplateWebhookReference)).Return(appNotificationWithRtmCtxAndTemplate, nil).Once()

				return notificationsBuilder
			},
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				appNotificationWithRtmCtxRtmIDAndTemplate,
				appNotificationWithRtmCtxAndTemplate,
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "",
		},
		{
			Name: "success when application and application template don`t have webhook",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Application, ApplicationID)).Once()
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, ApplicationTemplateID, model.ApplicationTemplateWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.ApplicationTemplate, ApplicationTemplateID)).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				return dataInputBuilder
			},
			ExpectedRequests:      []*webhookclient.FormationAssignmentNotificationRequest{},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "",
		},
		{
			Name: "success when application don`t have webhook and application template",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Application, ApplicationID)).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabelsWithoutTemplate, nil, nil).Once()
				return dataInputBuilder
			},
			ExpectedRequests:      []*webhookclient.FormationAssignmentNotificationRequest{},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "",
		},
		{
			Name: "success - error when generating notification results in notification not generated",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimesAndRuntimeContextsMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(runtimesMapping, runtimeContextsMapping, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithLabels, runtimeCtx2WithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(nil, testErr).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithRtmCtxWithLabels, runtimeCtxWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(appNotificationWithRtmCtxAndTemplate, nil).Once()

				return notificationsBuilder
			},
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				appNotificationWithRtmCtxAndTemplate,
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "",
		},
		{
			Name: "error when preparing details for notification",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimesAndRuntimeContextsMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(runtimesMapping, runtimeContextsMapping, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithLabels, runtimeCtx2WithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(nil, testErr).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithRtmCtxWithLabels, runtimeCtxWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Maybe()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(appNotificationWithRtmCtxAndTemplate, nil).Maybe()

				return notificationsBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    testErr.Error(),
		},
		{
			Name: "error when preparing runtime mappings",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimesAndRuntimeContextsMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(nil, nil, testErr).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while preparing runtime and runtime contexts mappings",
		},
		{
			Name: "error when getting application template webhook",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Application, ApplicationID)).Once()
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, ApplicationTemplateID, model.ApplicationTemplateWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				return dataInputBuilder
			},
			ExpectedRequests:      []*webhookclient.FormationAssignmentNotificationRequest{},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while listing CONFIGURATION_CHANGED webhooks for application template 58963c6f-24f6-4128-a05c-51d5356e7e09 on behalve of application",
		},
		{
			Name: "error when getting application webhook",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				return dataInputBuilder
			},
			ExpectedRequests:      []*webhookclient.FormationAssignmentNotificationRequest{},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while listing CONFIGURATION_CHANGED webhooks for application",
		},
		{
			Name: "error when getting application mappings",
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(nil, nil, testErr).Once()
				return dataInputBuilder
			},
			ExpectedRequests:      []*webhookclient.FormationAssignmentNotificationRequest{},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while preparing application and application template with labels",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			applicationRepo := unusedApplicationRepo()
			if testCase.ApplicationRepoFN != nil {
				applicationRepo = testCase.ApplicationRepoFN()
			}
			webhookRepo := unusedWebhookRepository()
			if testCase.WebhookRepoFN != nil {
				webhookRepo = testCase.WebhookRepoFN()
			}
			dataInputBuilder := unusedDataInputBuilder()
			if testCase.DataInputBuilder != nil {
				dataInputBuilder = testCase.DataInputBuilder()
			}
			notificationsBuilder := unusedNotificationsBuilder()
			if testCase.NotificationsBuilder != nil {
				notificationsBuilder = testCase.NotificationsBuilder()
			}

			notificationSvc := formation.NewNotificationsGenerator(applicationRepo, nil, nil, nil, nil, webhookRepo, dataInputBuilder, notificationsBuilder)

			// WHEN
			actual, err := notificationSvc.GenerateNotificationsAboutRuntimeAndRuntimeContextForTheApplicationThatIsAssigned(ctx, TntInternalID, testCase.ObjectID, &testCase.InputFormation, testCase.OperationType, testCase.CustomerTenantContext)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.ElementsMatch(t, testCase.ExpectedRequests, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, applicationRepo, webhookRepo, dataInputBuilder, notificationsBuilder)
		})
	}
}

func Test_GenerateNotificationsForRuntimeAboutTheApplicationThatIsAssigned(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	expectedFormation := model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            TntInternalID,
	}

	appTemplateWithLabels := &webhook.ApplicationTemplateWithLabels{
		ApplicationTemplate: fixApplicationTemplateModel(),
		Labels:              fixApplicationTemplateLabelsMap(),
	}

	appWithLabels := &webhook.ApplicationWithLabels{
		Application: fixApplicationModel(ApplicationID),
		Labels:      fixApplicationLabelsMap(),
	}

	testCases := []struct {
		Name                  string
		ApplicationRepoFN     func() *automock.ApplicationRepository
		WebhookRepoFN         func() *automock.WebhookRepository
		WebhookClientFN       func() *automock.WebhookClient
		DataInputBuilder      func() *databuilderautomock.DataInputBuilder
		NotificationsBuilder  func() *automock.NotificationBuilder
		CustomerTenantContext *webhook.CustomerTenantContext
		ObjectID              string
		OperationType         model.FormationOperation
		InputFormation        model.Formation
		ExpectedRequests      []*webhookclient.FormationAssignmentNotificationRequest
		ExpectedErrMessage    string
	}{
		{
			Name: "success when generating notifications for runtime about assigned application",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), fixConfigurationChangedWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)}, nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimesAndRuntimeContextsMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(runtimesMapping, runtimeContextsMapping2, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.RuntimeResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference)).Return(runtimeNotificationWithAppTemplate, nil).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithRtmCtxWithLabels, runtimeCtxWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.RuntimeResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)).Return(runtimeNotificationWithRtmCtxAndAppTemplate, nil).Once()

				return notificationsBuilder
			},
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				runtimeNotificationWithAppTemplate,
				runtimeNotificationWithRtmCtxAndAppTemplate,
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "",
		},
		{
			Name: "success when there are no runtime webhooks",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{}, nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "",
		},
		{
			Name: "success - error when generating notification results in not generated notification",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), fixConfigurationChangedWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)}, nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimesAndRuntimeContextsMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(runtimesMapping, runtimeContextsMapping2, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.RuntimeResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference)).Return(nil, testErr).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithRtmCtxWithLabels, runtimeCtxWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.RuntimeResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)).Return(runtimeNotificationWithRtmCtxAndAppTemplate, nil).Once()

				return notificationsBuilder
			},
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				runtimeNotificationWithRtmCtxAndAppTemplate,
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "",
		},
		{
			Name: "error when preparing details for notification",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), fixConfigurationChangedWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)}, nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimesAndRuntimeContextsMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(runtimesMapping, runtimeContextsMapping2, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.RuntimeResourceType, CustomerTenantContextAccount).Return(nil, testErr).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithRtmCtxWithLabels, runtimeCtxWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.RuntimeResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Maybe()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)).Return(runtimeNotificationWithRtmCtxAndAppTemplate, nil).Maybe()

				return notificationsBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    testErr.Error(),
		},
		{
			Name: "error when preparing runtime mappings",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), fixConfigurationChangedWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)}, nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimesAndRuntimeContextsMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(nil, nil, testErr).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while preparing runtime and runtime contexts mappings",
		},
		{
			Name: "error when listing webhooks",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, testErr).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "when listing configuration changed webhooks for runtimes",
		},
		{
			Name: "error while preparing application with labels and application template with labels",
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(nil, nil, testErr).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while preparing application and application template with labels",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			applicationRepo := unusedApplicationRepo()
			if testCase.ApplicationRepoFN != nil {
				applicationRepo = testCase.ApplicationRepoFN()
			}
			webhookRepo := unusedWebhookRepository()
			if testCase.WebhookRepoFN != nil {
				webhookRepo = testCase.WebhookRepoFN()
			}
			dataInputBuilder := unusedDataInputBuilder()
			if testCase.DataInputBuilder != nil {
				dataInputBuilder = testCase.DataInputBuilder()
			}
			notificationsBuilder := unusedNotificationsBuilder()
			if testCase.NotificationsBuilder != nil {
				notificationsBuilder = testCase.NotificationsBuilder()
			}

			notificationSvc := formation.NewNotificationsGenerator(applicationRepo, nil, nil, nil, nil, webhookRepo, dataInputBuilder, notificationsBuilder)

			// WHEN
			actual, err := notificationSvc.GenerateNotificationsForRuntimeAboutTheApplicationThatIsAssigned(ctx, TntInternalID, testCase.ObjectID, &testCase.InputFormation, testCase.OperationType, testCase.CustomerTenantContext)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.ElementsMatch(t, testCase.ExpectedRequests, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, applicationRepo, webhookRepo, dataInputBuilder, notificationsBuilder)
		})
	}
}

func Test_GenerateNotificationsForApplicationsAboutTheApplicationThatIsAssigned(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	expectedFormation := model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            TntInternalID,
	}

	appTemplateWithLabels := &webhook.ApplicationTemplateWithLabels{
		ApplicationTemplate: fixApplicationTemplateModel(),
		Labels:              fixApplicationTemplateLabelsMap(),
	}

	appTemplateWithLabels2 := &webhook.ApplicationTemplateWithLabels{
		ApplicationTemplate: fixApplicationTemplateModelWithID(ApplicationTemplate2ID),
		Labels:              fixApplicationTemplateLabelsMap(),
	}

	appWithLabels := &webhook.ApplicationWithLabels{
		Application: fixApplicationModel(ApplicationID),
		Labels:      fixApplicationLabelsMap(),
	}

	testCases := []struct {
		Name                  string
		ApplicationRepoFN     func() *automock.ApplicationRepository
		WebhookRepoFN         func() *automock.WebhookRepository
		WebhookClientFN       func() *automock.WebhookClient
		DataInputBuilder      func() *databuilderautomock.DataInputBuilder
		NotificationsBuilder  func() *automock.NotificationBuilder
		CustomerTenantContext *webhook.CustomerTenantContext
		ObjectID              string
		OperationType         model.FormationOperation
		InputFormation        model.Formation
		ExpectedRequests      []*webhookclient.FormationAssignmentNotificationRequest
		ExpectedErrMessage    string
	}{
		{
			Name: "success when generating notifications for application about applications in formation",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping).Return(listeningApplications, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(WebhookID, ApplicationID), fixApplicationTenantMappingWebhookModel(Webhook2ID, Application2ID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping2, emptyApplicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate2, appTemplateWithLabels, appWithLabels, emptyFormationAssignment, emptyFormationAssignment, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixApplicationTenantMappingWebhookModel(WebhookID, ApplicationID)).Return(appToAppNotificationWithoutSourceTemplateWithTargetTemplate, nil).Once()

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate2, emptyFormationAssignment, emptyFormationAssignment, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixApplicationTenantMappingWebhookModel(Webhook2ID, Application2ID)).Return(appToAppNotificationWithSourceTemplateWithoutTargetTemplate, nil).Once()

				return notificationsBuilder
			},
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				appToAppNotificationWithoutSourceTemplateWithTargetTemplate,
				appToAppNotificationWithSourceTemplateWithoutTargetTemplate,
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "",
		},
		{
			Name: "success when generating notifications for application with templates",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping).Return(listeningApplications, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(WebhookID, ApplicationID), fixApplicationTenantMappingWebhookModel(Webhook2ID, Application2ID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMappingWithApplicationTemplate, applicationTemplateMappings2, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels2, applicationsMappingWithApplicationTemplate[Application2ID], appTemplateWithLabels, appWithLabels, emptyFormationAssignment, emptyFormationAssignment, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixApplicationTenantMappingWebhookModel(WebhookID, ApplicationID)).Return(appToAppNotificationWithSourceAndTargetTemplatesSwaped, nil).Once()

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, appTemplateWithLabels2, applicationsMappingWithApplicationTemplate[Application2ID], emptyFormationAssignment, emptyFormationAssignment, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixApplicationTenantMappingWebhookModel(Webhook2ID, Application2ID)).Return(appToAppNotificationWithSourceAndTargetTemplates, nil).Once()

				return notificationsBuilder
			},
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				appToAppNotificationWithSourceAndTargetTemplates,
				appToAppNotificationWithSourceAndTargetTemplatesSwaped,
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "",
		},
		{
			Name: "success when using webhook from application template",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping).Return(listeningApplications, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(WebhookID, ApplicationTemplateID), fixApplicationTenantMappingWebhookModel(Webhook2ID, Application2ID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping2, emptyApplicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate2, appTemplateWithLabels, appWithLabels, emptyFormationAssignment, emptyFormationAssignment, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixApplicationTenantMappingWebhookModel(WebhookID, ApplicationTemplateID)).Return(appToAppNotificationWithoutSourceTemplateWithTargetTemplate, nil).Once()

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate2, emptyFormationAssignment, emptyFormationAssignment, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixApplicationTenantMappingWebhookModel(Webhook2ID, Application2ID)).Return(appToAppNotificationWithSourceTemplateWithoutTargetTemplate, nil).Once()

				return notificationsBuilder
			},
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				appToAppNotificationWithoutSourceTemplateWithTargetTemplate,
				appToAppNotificationWithSourceTemplateWithoutTargetTemplate,
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "",
		},
		{
			Name: "success - error while generating notification for currently assigned application results in skipping generating the notification",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping).Return(listeningApplications, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(WebhookID, ApplicationID), fixApplicationTenantMappingWebhookModel(Webhook2ID, Application2ID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping2, emptyApplicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate2, appTemplateWithLabels, appWithLabels, emptyFormationAssignment, emptyFormationAssignment, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixApplicationTenantMappingWebhookModel(WebhookID, ApplicationID)).Return(nil, testErr).Once()

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate2, emptyFormationAssignment, emptyFormationAssignment, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixApplicationTenantMappingWebhookModel(Webhook2ID, Application2ID)).Return(appToAppNotificationWithSourceTemplateWithoutTargetTemplate, nil).Once()

				return notificationsBuilder
			},
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				appToAppNotificationWithSourceTemplateWithoutTargetTemplate,
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "",
		},
		{
			Name: "success - error while generating notification for already assigned application results in skipping generating the notification",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping).Return(listeningApplications, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(WebhookID, ApplicationID), fixApplicationTenantMappingWebhookModel(Webhook2ID, Application2ID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping2, emptyApplicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate2, appTemplateWithLabels, appWithLabels, emptyFormationAssignment, emptyFormationAssignment, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixApplicationTenantMappingWebhookModel(WebhookID, ApplicationID)).Return(appToAppNotificationWithoutSourceTemplateWithTargetTemplate, nil).Once()

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate2, emptyFormationAssignment, emptyFormationAssignment, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixApplicationTenantMappingWebhookModel(Webhook2ID, Application2ID)).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				appToAppNotificationWithoutSourceTemplateWithTargetTemplate,
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "",
		},
		{
			Name: "success when there are no listening applications",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping).Return([]*model.Application{}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(WebhookID, ApplicationID), fixApplicationTenantMappingWebhookModel(Webhook2ID, Application2ID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "",
		},
		{
			Name: "error when preparing details for notification for already assigned application",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping).Return(listeningApplications, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(WebhookID, ApplicationID), fixApplicationTenantMappingWebhookModel(Webhook2ID, Application2ID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping2, emptyApplicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate2, appTemplateWithLabels, appWithLabels, emptyFormationAssignment, emptyFormationAssignment, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixApplicationTenantMappingWebhookModel(WebhookID, ApplicationID)).Return(appToAppNotificationWithoutSourceTemplateWithTargetTemplate, nil).Once()

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate2, emptyFormationAssignment, emptyFormationAssignment, CustomerTenantContextAccount).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    testErr.Error(),
		},
		{
			Name: "error when preparing details for notification for currently assigned application",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping).Return(listeningApplications, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(WebhookID, ApplicationID), fixApplicationTenantMappingWebhookModel(Webhook2ID, Application2ID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping2, emptyApplicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForApplicationTenantMappingNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate2, appTemplateWithLabels, appWithLabels, emptyFormationAssignment, emptyFormationAssignment, CustomerTenantContextAccount).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    testErr.Error(),
		},
		{
			Name: "error when preparing application mappings",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping).Return(listeningApplications, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(WebhookID, ApplicationID), fixApplicationTenantMappingWebhookModel(Webhook2ID, Application2ID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(nil, nil, testErr).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while preparing application and application template mappings",
		},
		{
			Name: "error when listing listening applications",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping).Return(nil, testErr).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(WebhookID, ApplicationID), fixApplicationTenantMappingWebhookModel(Webhook2ID, Application2ID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while listing listening applications",
		},
		{
			Name: "error when listing webhooks",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeApplicationTenantMapping, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return(nil, testErr).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "when listing application tenant mapping webhooks for applications and their application templates",
		},
		{
			Name: "error when preparing application and application template with labels",
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, TntInternalID, ApplicationID).Return(nil, nil, testErr).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              ApplicationID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while preparing application and application template with labels",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			applicationRepo := unusedApplicationRepo()
			if testCase.ApplicationRepoFN != nil {
				applicationRepo = testCase.ApplicationRepoFN()
			}
			webhookRepo := unusedWebhookRepository()
			if testCase.WebhookRepoFN != nil {
				webhookRepo = testCase.WebhookRepoFN()
			}
			dataInputBuilder := unusedDataInputBuilder()
			if testCase.DataInputBuilder != nil {
				dataInputBuilder = testCase.DataInputBuilder()
			}
			notificationsBuilder := unusedNotificationsBuilder()
			if testCase.NotificationsBuilder != nil {
				notificationsBuilder = testCase.NotificationsBuilder()
			}

			notificationSvc := formation.NewNotificationsGenerator(applicationRepo, nil, nil, nil, nil, webhookRepo, dataInputBuilder, notificationsBuilder)

			// WHEN
			actual, err := notificationSvc.GenerateNotificationsForApplicationsAboutTheApplicationThatIsAssigned(ctx, TntInternalID, testCase.ObjectID, &testCase.InputFormation, testCase.OperationType, testCase.CustomerTenantContext)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.ElementsMatch(t, testCase.ExpectedRequests, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, applicationRepo, webhookRepo, dataInputBuilder, notificationsBuilder)
		})
	}
}

func Test_GenerateNotificationsForApplicationsAboutTheRuntimeContextThatIsAssigned(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	expectedFormation := model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            TntInternalID,
	}

	appTemplateWithLabels := &webhook.ApplicationTemplateWithLabels{
		ApplicationTemplate: fixApplicationTemplateModel(),
		Labels:              fixApplicationTemplateLabelsMap(),
	}

	appWithLabels := &webhook.ApplicationWithLabels{
		Application: fixApplicationModel(ApplicationID),
		Labels:      fixApplicationLabelsMap(),
	}

	testCases := []struct {
		Name                  string
		ApplicationRepoFN     func() *automock.ApplicationRepository
		WebhookRepoFN         func() *automock.WebhookRepository
		WebhookClientFN       func() *automock.WebhookClient
		DataInputBuilder      func() *databuilderautomock.DataInputBuilder
		NotificationsBuilder  func() *automock.NotificationBuilder
		CustomerTenantContext *webhook.CustomerTenantContext
		ObjectID              string
		ObjectType            graphql.FormationObjectType
		OperationType         model.FormationOperation
		InputFormation        model.Formation
		ExpectedRequests      []*webhookclient.FormationAssignmentNotificationRequest
		ExpectedErrMessage    string
	}{
		{
			Name: "success when generating notifications for all listening applications about the assigned runtime context in that formation",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeConfigurationChanged).Return(listeningApplications, nil).Once()

				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping, applicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithRtmCtxWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(appNotificationWithRtmCtxAndTemplate, nil).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate2, runtimeWithRtmCtxWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(appNotificationWithRtmCtxWithoutTemplate, nil).Once()

				return notificationsBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeContextID,
			InputFormation:        expectedFormation,
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				appNotificationWithRtmCtxAndTemplate,
				appNotificationWithRtmCtxWithoutTemplate,
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "success when using webhook from application template",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeConfigurationChanged).Return(listeningApplications, nil).Once()

				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationTemplateID, model.ApplicationTemplateWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping, applicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithRtmCtxWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, ApplicationTemplateID, model.ApplicationTemplateWebhookReference)).Return(appNotificationWithRtmCtxAndTemplate, nil).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate2, runtimeWithRtmCtxWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(appNotificationWithRtmCtxWithoutTemplate, nil).Once()

				return notificationsBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeContextID,
			InputFormation:        expectedFormation,
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				appNotificationWithRtmCtxAndTemplate,
				appNotificationWithRtmCtxWithoutTemplate,
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "success - using webhook from application when bot application and application template have webhooks",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeConfigurationChanged).Return(listeningApplications, nil).Once()

				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook3ID, ApplicationTemplateID, model.ApplicationTemplateWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping, applicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithRtmCtxWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(appNotificationWithRtmCtxAndTemplate, nil).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate2, runtimeWithRtmCtxWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(appNotificationWithRtmCtxWithoutTemplate, nil).Once()

				return notificationsBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeContextID,
			InputFormation:        expectedFormation,
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				appNotificationWithRtmCtxAndTemplate,
				appNotificationWithRtmCtxWithoutTemplate,
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Error when building NotificationRequest results in not generating notification",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeConfigurationChanged).Return(listeningApplications, nil).Once()

				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping, applicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithRtmCtxWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(nil, testErr).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate2, runtimeWithRtmCtxWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(appNotificationWithRtmCtxWithoutTemplate, nil).Once()

				return notificationsBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeContextID,
			InputFormation:        expectedFormation,
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				appNotificationWithRtmCtxWithoutTemplate,
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Error when generating details for application notification about runtime",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeConfigurationChanged).Return(listeningApplications, nil).Once()

				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping, applicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithRtmCtxWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(nil, testErr).Once()

				// the method we are testing iterates over map, so it is not certain whether this will be invoked or not
				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate2, runtimeWithRtmCtxWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Maybe()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(appNotificationWithRtmCtxWithoutTemplate, nil).Maybe()

				return notificationsBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeContextID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    testErr.Error(),
		},
		{
			Name: "error when preparing applications mapping and application template mapping",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeConfigurationChanged).Return(listeningApplications, nil).Once()

				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(nil, nil, testErr).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeContextID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    testErr.Error(),
		},
		{
			Name: "success when there are no listening applications",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeConfigurationChanged).Return([]*model.Application{}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeContextRuntimeID).Return(runtimeWithLabels, nil).Once()
				return dataInputBuilder
			},
			OperationType:      model.AssignFormation,
			ObjectID:           RuntimeContextID,
			InputFormation:     expectedFormation,
			ExpectedRequests:   []*webhookclient.FormationAssignmentNotificationRequest{},
			ExpectedErrMessage: "",
		},
		{
			Name: "error when listing listening applications",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeConfigurationChanged).Return(nil, testErr).Once()

				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeContextID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while listing listening applications",
		},
		{
			Name: "error when listing webhooks",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return(nil, testErr).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeContextID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "when listing configuration changed webhooks for applications and their application templates",
		},
		{
			Name: "error while preparing runtime with labels",
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeContextRuntimeID).Return(nil, testErr).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeContextID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while preparing runtime with labels",
		},
		{
			Name: "error while preparing runtime context with labels",
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(nil, testErr).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeContextID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while preparing runtime context with labels",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			applicationRepo := unusedApplicationRepo()
			if testCase.ApplicationRepoFN != nil {
				applicationRepo = testCase.ApplicationRepoFN()
			}
			webhookRepo := unusedWebhookRepository()
			if testCase.WebhookRepoFN != nil {
				webhookRepo = testCase.WebhookRepoFN()
			}
			dataInputBuilder := unusedDataInputBuilder()
			if testCase.DataInputBuilder != nil {
				dataInputBuilder = testCase.DataInputBuilder()
			}
			notificationsBuilder := unusedNotificationsBuilder()
			if testCase.NotificationsBuilder != nil {
				notificationsBuilder = testCase.NotificationsBuilder()
			}

			notificationSvc := formation.NewNotificationsGenerator(applicationRepo, nil, nil, nil, nil, webhookRepo, dataInputBuilder, notificationsBuilder)

			// WHEN
			actual, err := notificationSvc.GenerateNotificationsForApplicationsAboutTheRuntimeContextThatIsAssigned(ctx, TntInternalID, testCase.ObjectID, &testCase.InputFormation, testCase.OperationType, testCase.CustomerTenantContext)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.ElementsMatch(t, testCase.ExpectedRequests, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, applicationRepo, webhookRepo, dataInputBuilder, notificationsBuilder)
		})
	}
}

func Test_GenerateNotificationsForApplicationsAboutTheRuntimeThatIsAssigned(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	expectedFormation := model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            TntInternalID,
	}

	appTemplateWithLabels := &webhook.ApplicationTemplateWithLabels{
		ApplicationTemplate: fixApplicationTemplateModel(),
		Labels:              fixApplicationTemplateLabelsMap(),
	}

	appWithLabels := &webhook.ApplicationWithLabels{
		Application: fixApplicationModel(ApplicationID),
		Labels:      fixApplicationLabelsMap(),
	}

	appWithLabelsWithoutTemplate := &webhook.ApplicationWithLabels{
		Application: fixApplicationModelWithoutTemplate(Application2ID),
		Labels:      fixApplicationLabelsMap(),
	}

	testCases := []struct {
		Name                  string
		ApplicationRepoFN     func() *automock.ApplicationRepository
		WebhookRepoFN         func() *automock.WebhookRepository
		WebhookClientFN       func() *automock.WebhookClient
		DataInputBuilder      func() *databuilderautomock.DataInputBuilder
		NotificationsBuilder  func() *automock.NotificationBuilder
		CustomerTenantContext *webhook.CustomerTenantContext
		ObjectID              string
		ObjectType            graphql.FormationObjectType
		OperationType         model.FormationOperation
		InputFormation        model.Formation
		ExpectedRequests      []*webhookclient.FormationAssignmentNotificationRequest
		ExpectedErrMessage    string
	}{
		{
			Name: "success when generating notifications for all listening applications about the assigned runtime in that formation",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeConfigurationChanged).Return(listeningApplications, nil).Once()

				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeID).Return(runtimeWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping, applicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(applicationNotificationWithAppTemplate, nil).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(applicationNotificationWithoutAppTemplate, nil)

				return notificationsBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeID,
			InputFormation:        expectedFormation,
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				applicationNotificationWithAppTemplate,
				applicationNotificationWithoutAppTemplate,
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "success when using webhook from application template",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeConfigurationChanged).Return(listeningApplications, nil).Once()

				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationTemplateID, model.ApplicationTemplateWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeID).Return(runtimeWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping, applicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, ApplicationTemplateID, model.ApplicationTemplateWebhookReference)).Return(applicationNotificationWithAppTemplate, nil).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(applicationNotificationWithoutAppTemplate, nil)

				return notificationsBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeID,
			InputFormation:        expectedFormation,
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				applicationNotificationWithAppTemplate,
				applicationNotificationWithoutAppTemplate,
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "success - using webhook from application when bot application and application template have webhooks",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeConfigurationChanged).Return(listeningApplications, nil).Once()

				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook3ID, ApplicationTemplateID, model.ApplicationTemplateWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeID).Return(runtimeWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping, applicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(applicationNotificationWithAppTemplate, nil).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(applicationNotificationWithoutAppTemplate, nil)

				return notificationsBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeID,
			InputFormation:        expectedFormation,
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				applicationNotificationWithAppTemplate,
				applicationNotificationWithoutAppTemplate,
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Error when building NotificationRequest results in not generating notification",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeConfigurationChanged).Return(listeningApplications, nil).Once()

				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeID).Return(runtimeWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping, applicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(applicationNotificationWithAppTemplate, nil).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(nil, testErr).Once()

				return notificationsBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeID,
			InputFormation:        expectedFormation,
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				applicationNotificationWithAppTemplate,
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Error when generating details for application notification about runtime",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeConfigurationChanged).Return(listeningApplications, nil).Once()

				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeID).Return(runtimeWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping, applicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(nil, testErr).Once()

				// the method we are testing iterates over map, so it is not certain whether this will be invoked or not
				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.ApplicationResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Maybe()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(applicationNotificationWithoutAppTemplate, nil).Maybe()

				return notificationsBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    testErr.Error(),
		},
		{
			Name: "error when preparing applications mapping and application template mapping",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeConfigurationChanged).Return(listeningApplications, nil).Once()

				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeID).Return(runtimeWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(nil, nil, testErr).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    testErr.Error(),
		},
		{
			Name: "success when there are no listening applications",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeConfigurationChanged).Return([]*model.Application{}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeID).Return(runtimeWithLabels, nil).Once()
				return dataInputBuilder
			},
			OperationType:      model.AssignFormation,
			ObjectID:           RuntimeID,
			InputFormation:     expectedFormation,
			ExpectedRequests:   []*webhookclient.FormationAssignmentNotificationRequest{},
			ExpectedErrMessage: "",
		},
		{
			Name: "error when listing listening applications",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListListeningApplications", ctx, TntInternalID, model.WebhookTypeConfigurationChanged).Return(nil, testErr).Once()

				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeID).Return(runtimeWithLabels, nil).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while listing listening applications",
		},
		{
			Name: "error when listing webhooks",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypesAndWebhookType", ctx, TntInternalID, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference}).Return(nil, testErr).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeID).Return(runtimeWithLabels, nil).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "when listing configuration changed webhooks for applications and their application templates",
		},
		{
			Name: "error while preparing runtime with labels",
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeID).Return(nil, testErr).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while preparing runtime with labels",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			applicationRepo := unusedApplicationRepo()
			if testCase.ApplicationRepoFN != nil {
				applicationRepo = testCase.ApplicationRepoFN()
			}
			webhookRepo := unusedWebhookRepository()
			if testCase.WebhookRepoFN != nil {
				webhookRepo = testCase.WebhookRepoFN()
			}
			dataInputBuilder := unusedDataInputBuilder()
			if testCase.DataInputBuilder != nil {
				dataInputBuilder = testCase.DataInputBuilder()
			}
			notificationsBuilder := unusedNotificationsBuilder()
			if testCase.NotificationsBuilder != nil {
				notificationsBuilder = testCase.NotificationsBuilder()
			}

			notificationSvc := formation.NewNotificationsGenerator(applicationRepo, nil, nil, nil, nil, webhookRepo, dataInputBuilder, notificationsBuilder)

			// WHEN
			actual, err := notificationSvc.GenerateNotificationsForApplicationsAboutTheRuntimeThatIsAssigned(ctx, TntInternalID, testCase.ObjectID, &testCase.InputFormation, testCase.OperationType, testCase.CustomerTenantContext)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.ElementsMatch(t, testCase.ExpectedRequests, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, applicationRepo, webhookRepo, dataInputBuilder, notificationsBuilder)
		})
	}
}

func Test_GenerateNotificationsAboutApplicationsForTheRuntimeContextThatIsAssigned(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	expectedFormation := model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            TntInternalID,
	}

	appTemplateWithLabels := &webhook.ApplicationTemplateWithLabels{
		ApplicationTemplate: fixApplicationTemplateModel(),
		Labels:              fixApplicationTemplateLabelsMap(),
	}

	appWithLabels := &webhook.ApplicationWithLabels{
		Application: fixApplicationModel(ApplicationID),
		Labels:      fixApplicationLabelsMap(),
	}

	appWithLabelsWithoutTemplate := &webhook.ApplicationWithLabels{
		Application: fixApplicationModelWithoutTemplate(Application2ID),
		Labels:      fixApplicationLabelsMap(),
	}

	testCases := []struct {
		Name                  string
		WebhookRepoFN         func() *automock.WebhookRepository
		DataInputBuilder      func() *databuilderautomock.DataInputBuilder
		NotificationsBuilder  func() *automock.NotificationBuilder
		CustomerTenantContext *webhook.CustomerTenantContext
		ObjectID              string
		OperationType         model.FormationOperation
		InputFormation        model.Formation
		ExpectedRequests      []*webhookclient.FormationAssignmentNotificationRequest
		ExpectedErrMessage    string
	}{
		{
			Name: "success when generating notifications for runtime contexts about all applications in that formation",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference), nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping, applicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithRtmCtxWithLabels, runtimeCtxWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.RuntimeContextResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)).Return(runtimeCtxNotificationWithAppTemplate, nil).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate, runtimeWithRtmCtxWithLabels, runtimeCtxWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.RuntimeContextResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)).Return(runtimeCtxNotificationWithoutAppTemplate, nil).Once()

				return notificationsBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeContextID,
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				runtimeCtxNotificationWithAppTemplate,
				runtimeCtxNotificationWithoutAppTemplate,
			},
			InputFormation:     expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success when the runtime does not have a webhook",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Runtime, RuntimeID)).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeContextID,
			ExpectedRequests:      []*webhookclient.FormationAssignmentNotificationRequest{},
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "",
		},
		{
			Name: "error when preparing details for notification",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference), nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping, applicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithRtmCtxWithLabels, runtimeCtxWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.RuntimeContextResourceType, CustomerTenantContextAccount).Return(nil, testErr).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate, runtimeWithRtmCtxWithLabels, runtimeCtxWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.RuntimeContextResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Maybe()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)).Return(runtimeCtxNotificationWithoutAppTemplate, nil).Maybe()

				return notificationsBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeContextID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    testErr.Error(),
		},
		{
			Name: "success - error when generating notifications results in not generating notification",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference), nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping, applicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithRtmCtxWithLabels, runtimeCtxWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.RuntimeContextResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)).Return(nil, testErr).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate, runtimeWithRtmCtxWithLabels, runtimeCtxWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.RuntimeContextResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)).Return(runtimeCtxNotificationWithoutAppTemplate, nil).Once()

				return notificationsBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeContextID,
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				runtimeCtxNotificationWithoutAppTemplate,
			},
			InputFormation:     expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error when preparing application mappings",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference), nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(nil, nil, testErr).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeContextID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    testErr.Error(),
		},
		{
			Name: "error when preparing runtime with labels",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference), nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeContextRuntimeID).Return(nil, testErr).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeContextID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while preparing runtime with labels",
		},
		{
			Name: "error when getting webhook",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeContextID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while listing configuration changed webhooks for runtime",
		},
		{
			Name: "error when preparing runtime context with labels",
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, TntInternalID, RuntimeContextID).Return(nil, testErr).Once()
				return dataInputBuilder
			},
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeContextID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while preparing runtime context with labels",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			webhookRepo := unusedWebhookRepository()
			if testCase.WebhookRepoFN != nil {
				webhookRepo = testCase.WebhookRepoFN()
			}
			dataInputBuilder := unusedDataInputBuilder()
			if testCase.DataInputBuilder != nil {
				dataInputBuilder = testCase.DataInputBuilder()
			}
			notificationsBuilder := unusedNotificationsBuilder()
			if testCase.NotificationsBuilder != nil {
				notificationsBuilder = testCase.NotificationsBuilder()
			}

			notificationSvc := formation.NewNotificationsGenerator(nil, nil, nil, nil, nil, webhookRepo, dataInputBuilder, notificationsBuilder)

			// WHEN
			actual, err := notificationSvc.GenerateNotificationsAboutApplicationsForTheRuntimeContextThatIsAssigned(ctx, TntInternalID, testCase.ObjectID, &testCase.InputFormation, testCase.OperationType, testCase.CustomerTenantContext)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.ElementsMatch(t, testCase.ExpectedRequests, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, webhookRepo, dataInputBuilder, notificationsBuilder)
		})
	}
}

func Test_GenerateNotificationsAboutApplicationsForTheRuntimeThatIsAssigned(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	expectedFormation := model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            TntInternalID,
	}

	appTemplateWithLabels := &webhook.ApplicationTemplateWithLabels{
		ApplicationTemplate: fixApplicationTemplateModel(),
		Labels:              fixApplicationTemplateLabelsMap(),
	}

	appWithLabels := &webhook.ApplicationWithLabels{
		Application: fixApplicationModel(ApplicationID),
		Labels:      fixApplicationLabelsMap(),
	}

	appWithLabelsWithoutTemplate := &webhook.ApplicationWithLabels{
		Application: fixApplicationModelWithoutTemplate(Application2ID),
		Labels:      fixApplicationLabelsMap(),
	}

	testCases := []struct {
		Name                  string
		WebhookRepoFN         func() *automock.WebhookRepository
		WebhookClientFN       func() *automock.WebhookClient
		DataInputBuilder      func() *databuilderautomock.DataInputBuilder
		NotificationsBuilder  func() *automock.NotificationBuilder
		CustomerTenantContext *webhook.CustomerTenantContext
		ObjectID              string
		ObjectType            graphql.FormationObjectType
		OperationType         model.FormationOperation
		InputFormation        model.Formation
		ExpectedRequests      []*webhookclient.FormationAssignmentNotificationRequest
		ExpectedErrMessage    string
	}{
		{
			Name: "success when generating notifications for runtime about all applications in that formation",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeID).Return(runtimeWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping, applicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.RuntimeResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference)).Return(runtimeNotificationWithAppTemplate, nil).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.RuntimeResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference)).Return(runtimeNotificationWithoutAppTemplate, nil).Once()

				return notificationsBuilder
			},
			ObjectType:            graphql.FormationObjectTypeRuntime,
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeID,
			InputFormation:        expectedFormation,
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				runtimeNotificationWithAppTemplate,
				runtimeNotificationWithoutAppTemplate,
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "success - error when building notification request results in not generating notification",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeID).Return(runtimeWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping, applicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.RuntimeResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference)).Return(nil, testErr).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.RuntimeResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Once()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference)).Return(runtimeNotificationWithoutAppTemplate, nil).Once()

				return notificationsBuilder
			},
			ObjectType:            graphql.FormationObjectTypeRuntime,
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeID,
			InputFormation:        expectedFormation,
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				runtimeNotificationWithoutAppTemplate,
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "error when preparing notification details",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeID).Return(runtimeWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(applicationsMapping, applicationTemplateMappings, nil).Once()
				return dataInputBuilder
			},
			NotificationsBuilder: func() *automock.NotificationBuilder {
				notificationsBuilder := &automock.NotificationBuilder{}

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, appTemplateWithLabels, appWithLabels, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.RuntimeResourceType, CustomerTenantContextAccount).Return(nil, testErr).Once()

				notificationsBuilder.On("PrepareDetailsForConfigurationChangeNotificationGeneration", model.AssignFormation, FormationID, emptyAppTemplateWithLabels, appWithLabelsWithoutTemplate, runtimeWithLabels, emptyRuntimeContextWithLabels, emptyFormationAssignment, emptyFormationAssignment, model.RuntimeResourceType, CustomerTenantContextAccount).Return(notificationDetails, nil).Maybe()
				notificationsBuilder.On("BuildFormationAssignmentNotificationRequest", ctx, FormationTemplateID, notificationDetails, fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference)).Return(runtimeNotificationWithoutAppTemplate, nil).Maybe()

				return notificationsBuilder
			},
			ObjectType:            graphql.FormationObjectTypeRuntime,
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    testErr.Error(),
		},
		{
			Name: "error when preparing application mappings",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), nil).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeID).Return(runtimeWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationMappingsInFormation", ctx, TntInternalID, expectedFormation.Name).Return(nil, nil, testErr).Once()
				return dataInputBuilder
			},
			ObjectType:            graphql.FormationObjectTypeRuntime,
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    testErr.Error(),
		},
		{
			Name: "error when getting webhook",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeID).Return(runtimeWithLabels, nil).Once()
				return dataInputBuilder
			},
			ObjectType:            graphql.FormationObjectTypeRuntime,
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while listing configuration changed webhooks for runtime",
		},
		{
			Name: "success when webhook not found",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, TntInternalID, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.RuntimeWebhook, RuntimeID)).Once()
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeID).Return(runtimeWithLabels, nil).Once()
				return dataInputBuilder
			},
			ObjectType:            graphql.FormationObjectTypeRuntime,
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeID,
			InputFormation:        expectedFormation,
			ExpectedRequests:      []*webhookclient.FormationAssignmentNotificationRequest{},
			ExpectedErrMessage:    "",
		},
		{
			Name: "error preparing runtime with labels",
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, TntInternalID, RuntimeID).Return(nil, testErr).Once()
				return dataInputBuilder
			},
			ObjectType:            graphql.FormationObjectTypeRuntime,
			OperationType:         model.AssignFormation,
			CustomerTenantContext: customerTenantContext,
			ObjectID:              RuntimeID,
			InputFormation:        expectedFormation,
			ExpectedErrMessage:    "while preparing runtime with labels",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			webhookRepo := unusedWebhookRepository()
			if testCase.WebhookRepoFN != nil {
				webhookRepo = testCase.WebhookRepoFN()
			}
			dataInputBuilder := unusedDataInputBuilder()
			if testCase.DataInputBuilder != nil {
				dataInputBuilder = testCase.DataInputBuilder()
			}

			notificationsBuilder := unusedNotificationsBuilder()
			if testCase.NotificationsBuilder != nil {
				notificationsBuilder = testCase.NotificationsBuilder()
			}

			notificationSvc := formation.NewNotificationsGenerator(nil, nil, nil, nil, nil, webhookRepo, dataInputBuilder, notificationsBuilder)

			// WHEN
			actual, err := notificationSvc.GenerateNotificationsAboutApplicationsForTheRuntimeThatIsAssigned(ctx, TntInternalID, testCase.ObjectID, &testCase.InputFormation, testCase.OperationType, testCase.CustomerTenantContext)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.ElementsMatch(t, testCase.ExpectedRequests, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, webhookRepo, dataInputBuilder, notificationsBuilder)
		})
	}
}

func Test_GenerateFormationLifecycleNotifications(t *testing.T) {
	ctx := context.Background()
	formationInput := fixFormationModelWithoutError()

	testCases := []struct {
		name                              string
		notificationsBuilderFn            func() *automock.NotificationBuilder
		expectedErrMsg                    string
		expectedFormationNotificationReqs []*webhookclient.FormationNotificationRequest
	}{
		{
			name: "Successfully generate formation notifications",
			notificationsBuilderFn: func() *automock.NotificationBuilder {
				notificationBuilder := &automock.NotificationBuilder{}
				notificationBuilder.On("BuildFormationNotificationRequests", ctx, formationNotificationDetails, formationInput, formationLifecycleSyncWebhooks).Return(formationNotificationRequests, nil).Once()
				return notificationBuilder
			},
			expectedFormationNotificationReqs: formationNotificationRequests,
		},
		{
			name: "Success when generating formation lifecycle notifications results in not generated notification due to error",
			notificationsBuilderFn: func() *automock.NotificationBuilder {
				notificationBuilder := &automock.NotificationBuilder{}
				notificationBuilder.On("BuildFormationNotificationRequests", ctx, formationNotificationDetails, formationInput, formationLifecycleSyncWebhooks).Return(nil, testErr).Once()
				return notificationBuilder
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			notificationsBuilder := unusedNotificationsBuilder()
			if testCase.notificationsBuilderFn != nil {
				notificationsBuilder = testCase.notificationsBuilderFn()
			}

			defer mock.AssertExpectationsForObjects(t, notificationsBuilder)

			notificationGenerator := formation.NewNotificationsGenerator(nil, nil, nil, nil, nil, nil, nil, notificationsBuilder)
			formationNotificationReqs, err := notificationGenerator.GenerateFormationLifecycleNotifications(ctx, formationLifecycleSyncWebhooks, TntInternalID, formationInput, testFormationTemplateName, FormationTemplateID, model.CreateFormation, CustomerTenantContextAccount)

			if testCase.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErrMsg)
				require.Empty(t, formationNotificationReqs)
			} else {
				require.NoError(t, err)
				require.ElementsMatch(t, formationNotificationReqs, testCase.expectedFormationNotificationReqs)
			}
		})
	}
}
