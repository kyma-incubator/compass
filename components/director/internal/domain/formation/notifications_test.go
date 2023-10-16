package formation_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"

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

func Test_NotificationsService_GenerateNotifications(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	inputFormation := &model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            TntInternalID,
	}

	testCases := []struct {
		Name                   string
		TenantRepoFn           func() *automock.TenantRepository
		NotificationsGenerator func() *automock.NotificationsGenerator
		ObjectID               string
		ObjectType             graphql.FormationObjectType
		OperationType          model.FormationOperation
		InputFormation         *model.Formation
		ExpectedRequests       []*webhookclient.FormationAssignmentNotificationRequest
		ExpectedErrMessage     string
	}{
		// start testing 'generateNotificationsAboutApplicationsForTheRuntimeThatIsAssigned' and 'generateNotificationsForApplicationsAboutTheRuntimeThatIsAssigned' funcs
		{
			Name: "success for runtime",
			TenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", ctx, inputFormation.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", ctx, inputFormation.TenantID).Return(TntParentID, nil)
				return repo
			},
			NotificationsGenerator: func() *automock.NotificationsGenerator {
				generator := &automock.NotificationsGenerator{}

				generator.On("GenerateNotificationsForApplicationsAboutTheRuntimeThatIsAssigned", ctx, TntInternalID, RuntimeID, inputFormation, model.AssignFormation, customerTenantContext).Return(
					[]*webhookclient.FormationAssignmentNotificationRequest{
						applicationNotificationWithAppTemplate,
						applicationNotificationWithoutAppTemplate,
					}, nil).Once()

				generator.On("GenerateNotificationsAboutApplicationsForTheRuntimeThatIsAssigned", ctx, TntInternalID, RuntimeID, inputFormation, model.AssignFormation, customerTenantContext).Return(
					[]*webhookclient.FormationAssignmentNotificationRequest{
						runtimeNotificationWithAppTemplate,
						runtimeNotificationWithoutAppTemplate,
					}, nil).Once()

				return generator
			},
			ObjectType:     graphql.FormationObjectTypeRuntime,
			OperationType:  model.AssignFormation,
			ObjectID:       RuntimeID,
			InputFormation: inputFormation,
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				runtimeNotificationWithAppTemplate,
				runtimeNotificationWithoutAppTemplate,
				applicationNotificationWithAppTemplate,
				applicationNotificationWithoutAppTemplate,
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime when customer tenant context is resource group",
			TenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", ctx, inputFormation.TenantID).Return(rgTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", ctx, inputFormation.TenantID).Return(TntParentID, nil)
				return repo
			},
			NotificationsGenerator: func() *automock.NotificationsGenerator {
				generator := &automock.NotificationsGenerator{}

				generator.On("GenerateNotificationsForApplicationsAboutTheRuntimeThatIsAssigned", ctx, TntInternalID, RuntimeID, inputFormation, model.AssignFormation, rgCustomerTenantContext).Return(
					[]*webhookclient.FormationAssignmentNotificationRequest{
						applicationNotificationWithAppTemplate,
						applicationNotificationWithoutAppTemplate,
					}, nil).Once()

				generator.On("GenerateNotificationsAboutApplicationsForTheRuntimeThatIsAssigned", ctx, TntInternalID, RuntimeID, inputFormation, model.AssignFormation, rgCustomerTenantContext).Return(
					[]*webhookclient.FormationAssignmentNotificationRequest{
						runtimeNotificationWithAppTemplate,
						runtimeNotificationWithoutAppTemplate,
					}, nil).Once()

				return generator
			},
			ObjectType:     graphql.FormationObjectTypeRuntime,
			OperationType:  model.AssignFormation,
			ObjectID:       RuntimeID,
			InputFormation: inputFormation,
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				runtimeNotificationWithAppTemplate,
				runtimeNotificationWithoutAppTemplate,
				applicationNotificationWithAppTemplate,
				applicationNotificationWithoutAppTemplate,
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "error for runtime - when generating notifications for runtime",
			TenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", ctx, inputFormation.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", ctx, inputFormation.TenantID).Return(TntParentID, nil)
				return repo
			},
			NotificationsGenerator: func() *automock.NotificationsGenerator {
				generator := &automock.NotificationsGenerator{}

				generator.On("GenerateNotificationsForApplicationsAboutTheRuntimeThatIsAssigned", ctx, TntInternalID, RuntimeID, inputFormation, model.AssignFormation, customerTenantContext).Return(
					[]*webhookclient.FormationAssignmentNotificationRequest{
						applicationNotificationWithAppTemplate,
						applicationNotificationWithoutAppTemplate,
					}, nil).Once()

				generator.On("GenerateNotificationsAboutApplicationsForTheRuntimeThatIsAssigned", ctx, TntInternalID, RuntimeID, inputFormation, model.AssignFormation, customerTenantContext).Return(nil, testErr).Once()

				return generator
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			OperationType:      model.AssignFormation,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime - when generating notifications for application",
			TenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", ctx, inputFormation.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", ctx, inputFormation.TenantID).Return(TntParentID, nil)
				return repo
			},
			NotificationsGenerator: func() *automock.NotificationsGenerator {
				generator := &automock.NotificationsGenerator{}

				generator.On("GenerateNotificationsForApplicationsAboutTheRuntimeThatIsAssigned", ctx, TntInternalID, RuntimeID, inputFormation, model.AssignFormation, customerTenantContext).Return(nil, testErr).Once()

				return generator
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			OperationType:      model.AssignFormation,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		// start testing 'generateNotificationsForApplicationsAboutTheRuntimeContextThatIsAssigned' and 'generateNotificationsAboutApplicationsForTheRuntimeContextThatIsAssigned' funcs
		{
			Name: "success for runtime context",
			TenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", ctx, inputFormation.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", ctx, inputFormation.TenantID).Return(TntParentID, nil)
				return repo
			},
			NotificationsGenerator: func() *automock.NotificationsGenerator {
				generator := &automock.NotificationsGenerator{}

				generator.On("GenerateNotificationsForApplicationsAboutTheRuntimeContextThatIsAssigned", ctx, TntInternalID, RuntimeContextID, inputFormation, model.AssignFormation, customerTenantContext).Return(
					[]*webhookclient.FormationAssignmentNotificationRequest{
						appNotificationWithRtmCtxAndTemplate,
						appNotificationWithRtmCtxWithoutTemplate,
					}, nil).Once()

				generator.On("GenerateNotificationsAboutApplicationsForTheRuntimeContextThatIsAssigned", ctx, TntInternalID, RuntimeContextID, inputFormation, model.AssignFormation, customerTenantContext).Return(
					[]*webhookclient.FormationAssignmentNotificationRequest{
						runtimeCtxNotificationWithAppTemplate,
						runtimeCtxNotificationWithoutAppTemplate,
					}, nil).Once()

				return generator
			},
			ObjectType:    graphql.FormationObjectTypeRuntimeContext,
			OperationType: model.AssignFormation,
			ObjectID:      RuntimeContextID,
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				runtimeCtxNotificationWithAppTemplate,
				runtimeCtxNotificationWithoutAppTemplate,
				appNotificationWithRtmCtxAndTemplate,
				appNotificationWithRtmCtxWithoutTemplate,
			},
			InputFormation:     inputFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for runtime context - when generating notifications for runtime context",
			TenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", ctx, inputFormation.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", ctx, inputFormation.TenantID).Return(TntParentID, nil)
				return repo
			},
			NotificationsGenerator: func() *automock.NotificationsGenerator {
				generator := &automock.NotificationsGenerator{}

				generator.On("GenerateNotificationsForApplicationsAboutTheRuntimeContextThatIsAssigned", ctx, TntInternalID, RuntimeContextID, inputFormation, model.AssignFormation, customerTenantContext).Return(
					[]*webhookclient.FormationAssignmentNotificationRequest{
						appNotificationWithRtmCtxAndTemplate,
						appNotificationWithRtmCtxWithoutTemplate,
					}, nil).Once()

				generator.On("GenerateNotificationsAboutApplicationsForTheRuntimeContextThatIsAssigned", ctx, TntInternalID, RuntimeContextID, inputFormation, model.AssignFormation, customerTenantContext).Return(nil, testErr).Once()

				return generator
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			OperationType:      model.AssignFormation,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context - when generating notifications for applications",
			TenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", ctx, inputFormation.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", ctx, inputFormation.TenantID).Return(TntParentID, nil)
				return repo
			},
			NotificationsGenerator: func() *automock.NotificationsGenerator {
				generator := &automock.NotificationsGenerator{}

				generator.On("GenerateNotificationsForApplicationsAboutTheRuntimeContextThatIsAssigned", ctx, TntInternalID, RuntimeContextID, inputFormation, model.AssignFormation, customerTenantContext).Return(nil, testErr).Once()

				return generator
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			OperationType:      model.AssignFormation,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		// start testing 'generateRuntimeNotificationsForApplicationAssignment' and 'generateApplicationNotificationsForApplicationAssignment' funcs
		{
			Name: "success for application",
			TenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", ctx, inputFormation.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", ctx, inputFormation.TenantID).Return(TntParentID, nil)
				return repo
			},
			NotificationsGenerator: func() *automock.NotificationsGenerator {
				generator := &automock.NotificationsGenerator{}

				generator.On("GenerateNotificationsAboutRuntimeAndRuntimeContextForTheApplicationThatIsAssigned", ctx, TntInternalID, ApplicationID, inputFormation, model.AssignFormation, customerTenantContext).Return(
					[]*webhookclient.FormationAssignmentNotificationRequest{
						appNotificationWithRtmCtxRtmIDAndTemplate,
						appNotificationWithRtmCtxAndTemplate,
					}, nil).Once()

				generator.On("GenerateNotificationsForRuntimeAboutTheApplicationThatIsAssigned", ctx, TntInternalID, ApplicationID, inputFormation, model.AssignFormation, customerTenantContext).Return(
					[]*webhookclient.FormationAssignmentNotificationRequest{
						runtimeNotificationWithAppTemplate,
						runtimeNotificationWithRtmCtxAndAppTemplate,
					}, nil).Once()

				generator.On("GenerateNotificationsForApplicationsAboutTheApplicationThatIsAssigned", ctx, TntInternalID, ApplicationID, inputFormation, model.AssignFormation, customerTenantContext).Return(
					[]*webhookclient.FormationAssignmentNotificationRequest{
						appToAppNotificationWithoutSourceTemplateWithTargetTemplate,
						appToAppNotificationWithSourceTemplateWithoutTargetTemplate,
					}, nil).Once()

				return generator
			},
			ExpectedRequests: []*webhookclient.FormationAssignmentNotificationRequest{
				runtimeNotificationWithAppTemplate,
				runtimeNotificationWithRtmCtxAndAppTemplate,
				appNotificationWithRtmCtxRtmIDAndTemplate,
				appNotificationWithRtmCtxAndTemplate,
				appToAppNotificationWithoutSourceTemplateWithTargetTemplate,
				appToAppNotificationWithSourceTemplateWithoutTargetTemplate,
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			OperationType:      model.AssignFormation,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for application - when generating app to app notifications",
			TenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", ctx, inputFormation.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", ctx, inputFormation.TenantID).Return(TntParentID, nil)
				return repo
			},
			NotificationsGenerator: func() *automock.NotificationsGenerator {
				generator := &automock.NotificationsGenerator{}

				generator.On("GenerateNotificationsAboutRuntimeAndRuntimeContextForTheApplicationThatIsAssigned", ctx, TntInternalID, ApplicationID, inputFormation, model.AssignFormation, customerTenantContext).Return(
					[]*webhookclient.FormationAssignmentNotificationRequest{
						appNotificationWithRtmCtxRtmIDAndTemplate,
						appNotificationWithRtmCtxAndTemplate,
					}, nil).Once()

				generator.On("GenerateNotificationsForRuntimeAboutTheApplicationThatIsAssigned", ctx, TntInternalID, ApplicationID, inputFormation, model.AssignFormation, customerTenantContext).Return(
					[]*webhookclient.FormationAssignmentNotificationRequest{
						runtimeNotificationWithAppTemplate,
						runtimeNotificationWithRtmCtxAndAppTemplate,
					}, nil).Once()

				generator.On("GenerateNotificationsForApplicationsAboutTheApplicationThatIsAssigned", ctx, TntInternalID, ApplicationID, inputFormation, model.AssignFormation, customerTenantContext).Return(nil, testErr).Once()

				return generator
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			OperationType:      model.AssignFormation,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application - when generating notifications for runtimes",
			TenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", ctx, inputFormation.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", ctx, inputFormation.TenantID).Return(TntParentID, nil)
				return repo
			},
			NotificationsGenerator: func() *automock.NotificationsGenerator {
				generator := &automock.NotificationsGenerator{}

				generator.On("GenerateNotificationsAboutRuntimeAndRuntimeContextForTheApplicationThatIsAssigned", ctx, TntInternalID, ApplicationID, inputFormation, model.AssignFormation, customerTenantContext).Return(
					[]*webhookclient.FormationAssignmentNotificationRequest{
						appNotificationWithRtmCtxRtmIDAndTemplate,
						appNotificationWithRtmCtxAndTemplate,
					}, nil).Once()

				generator.On("GenerateNotificationsForRuntimeAboutTheApplicationThatIsAssigned", ctx, TntInternalID, ApplicationID, inputFormation, model.AssignFormation, customerTenantContext).Return(nil, testErr).Once()

				return generator
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			OperationType:      model.AssignFormation,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application - when generating notifications for applications about runtimes and runtime contexts",
			TenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", ctx, inputFormation.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", ctx, inputFormation.TenantID).Return(TntParentID, nil)
				return repo
			},
			NotificationsGenerator: func() *automock.NotificationsGenerator {
				generator := &automock.NotificationsGenerator{}

				generator.On("GenerateNotificationsAboutRuntimeAndRuntimeContextForTheApplicationThatIsAssigned", ctx, TntInternalID, ApplicationID, inputFormation, model.AssignFormation, customerTenantContext).Return(nil, testErr).Once()

				return generator
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			OperationType:      model.AssignFormation,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when getting customer parent",
			TenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", ctx, inputFormation.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", ctx, inputFormation.TenantID).Return("", testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			OperationType:      model.AssignFormation,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when getting tenant",
			TenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", ctx, inputFormation.TenantID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			OperationType:      model.AssignFormation,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when unknown formation object type",
			TenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", ctx, inputFormation.TenantID).Return(gaTenantObject, nil)
				repo.On("GetCustomerIDParentRecursively", ctx, inputFormation.TenantID).Return(TntParentID, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			OperationType:      model.AssignFormation,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "unknown formation type",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			tenantRepo := unusedTenantRepo()
			if testCase.TenantRepoFn() != nil {
				tenantRepo = testCase.TenantRepoFn()
			}

			notificationsGenerator := unusedNotificationsGenerator()
			if testCase.NotificationsGenerator != nil {
				notificationsGenerator = testCase.NotificationsGenerator()
			}

			notificationSvc := formation.NewNotificationService(tenantRepo, nil, notificationsGenerator, nil, nil, nil)

			// WHEN
			actual, err := notificationSvc.GenerateFormationAssignmentNotifications(ctx, TntInternalID, testCase.ObjectID, testCase.InputFormation, testCase.OperationType, testCase.ObjectType)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.ElementsMatch(t, testCase.ExpectedRequests, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, tenantRepo, notificationsGenerator)
		})
	}
}

func Test_NotificationService_GenerateFormationNotifications(t *testing.T) {
	ctx := context.Background()
	formationInput := fixFormationModelWithoutError()

	testCases := []struct {
		name                              string
		tenantRepoFn                      func() *automock.TenantRepository
		notificationsGeneratorFn          func() *automock.NotificationsGenerator
		expectedErrMsg                    string
		expectedFormationNotificationReqs []*webhookclient.FormationNotificationRequest
	}{
		{
			name: "Successfully generate formation notifications",
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", ctx, TntInternalID).Return(gaTenantObject, nil).Once()
				tenantRepo.On("GetCustomerIDParentRecursively", ctx, TntInternalID).Return(TntCustomerID, nil).Once()
				return tenantRepo
			},
			notificationsGeneratorFn: func() *automock.NotificationsGenerator {
				notificationGenerator := &automock.NotificationsGenerator{}
				notificationGenerator.On("GenerateFormationLifecycleNotifications", ctx, formationLifecycleSyncWebhooks, TntInternalID, formationInput, testFormationTemplateName, FormationTemplateID, model.CreateFormation, CustomerTenantContextAccount).Return(formationNotificationSyncCreateRequests, nil).Once()
				return notificationGenerator
			},
			expectedFormationNotificationReqs: formationNotificationSyncCreateRequests,
		},
		{
			name: "Error when extracting customer tenant context fails",
			tenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", ctx, TntInternalID).Return(nil, testErr)
				return repo
			},
			expectedErrMsg: testErr.Error(),
		},
		{
			name: "Error when generating formation lifecycle notifications fail",
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", ctx, TntInternalID).Return(gaTenantObject, nil).Once()
				tenantRepo.On("GetCustomerIDParentRecursively", ctx, TntInternalID).Return(TntCustomerID, nil).Once()
				return tenantRepo
			},
			notificationsGeneratorFn: func() *automock.NotificationsGenerator {
				notificationGenerator := &automock.NotificationsGenerator{}
				notificationGenerator.On("GenerateFormationLifecycleNotifications", ctx, formationLifecycleSyncWebhooks, TntInternalID, formationInput, testFormationTemplateName, FormationTemplateID, model.CreateFormation, CustomerTenantContextAccount).Return(nil, testErr).Once()
				return notificationGenerator
			},
			expectedErrMsg: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tenantRepo := unusedTenantRepo()
			if testCase.tenantRepoFn != nil {
				tenantRepo = testCase.tenantRepoFn()
			}

			notificationGenerator := unusedNotificationsGenerator()
			if testCase.notificationsGeneratorFn != nil {
				notificationGenerator = testCase.notificationsGeneratorFn()
			}

			defer mock.AssertExpectationsForObjects(t, tenantRepo, notificationGenerator)

			notificationSvc := formation.NewNotificationService(tenantRepo, nil, notificationGenerator, nil, nil, nil)

			formationNotificationReqs, err := notificationSvc.GenerateFormationNotifications(ctx, formationLifecycleSyncWebhooks, TntInternalID, formationInput, testFormationTemplateName, FormationTemplateID, model.CreateFormation)

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

func Test_NotificationsService_SendNotification(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	subtype := "subtype"

	fa := fixFormationAssignmentModelWithParameters("id1", FormationID, RuntimeID, ApplicationID, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeApplication, model.InitialFormationState)
	reverseFa := fixFormationAssignmentModelWithParameters("id2", FormationID, ApplicationID, RuntimeID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.InitialFormationState)

	templateInput := &webhook.FormationConfigurationChangeInput{
		Operation:   model.AssignFormation,
		FormationID: FormationID,
		ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
			ApplicationTemplate: fixApplicationTemplateModel(),
			Labels:              fixApplicationTemplateLabelsMap(),
		},
		Application: &webhook.ApplicationWithLabels{
			Application: fixApplicationModel(ApplicationID),
			Labels:      fixApplicationLabelsMap(),
		},
		Runtime:               fixRuntimeWithLabels(RuntimeID),
		RuntimeContext:        nil,
		CustomerTenantContext: fixCustomerTenantContext(TntParentID, TntExternalID),
		Assignment:            emptyFormationAssignment,
		ReverseAssignment:     emptyFormationAssignment,
	}

	preJoinPointDetails := &formationconstraint.SendNotificationOperationDetails{
		ResourceType:               model.ApplicationResourceType,
		ResourceSubtype:            subtype,
		Location:                   formationconstraint.PreSendNotification,
		Operation:                  model.AssignFormation,
		Webhook:                    fixRuntimeWebhookGQLModel(WebhookID, RuntimeID),
		CorrelationID:              "",
		TemplateInput:              templateInput,
		FormationAssignment:        fa,
		ReverseFormationAssignment: reverseFa,
		Formation:                  formationModel,
	}
	postJoinPointDetails := &formationconstraint.SendNotificationOperationDetails{
		ResourceType:               model.ApplicationResourceType,
		ResourceSubtype:            subtype,
		Location:                   formationconstraint.PostSendNotification,
		Operation:                  model.AssignFormation,
		Webhook:                    fixRuntimeWebhookGQLModel(WebhookID, RuntimeID),
		CorrelationID:              "",
		TemplateInput:              templateInput,
		FormationAssignment:        fa,
		ReverseFormationAssignment: reverseFa,
		Formation:                  formationModel,
	}

	faRequestExt := &webhookclient.FormationAssignmentNotificationRequestExt{
		FormationAssignmentNotificationRequest: &webhookclient.FormationAssignmentNotificationRequest{
			Webhook:       fixRuntimeWebhookGQLModel(WebhookID, RuntimeID),
			Object:        templateInput,
			CorrelationID: "",
		},
		Operation:                  model.AssignFormation,
		FormationAssignment:        fa,
		ReverseFormationAssignment: reverseFa,
		Formation:                  formationModel,
		TargetSubtype:              subtype,
	}

	testCases := []struct {
		Name               string
		WebhookClientFN    func() *automock.WebhookClient
		ConstraintEngine   func() *automock.ConstraintEngine
		WebhookConverter   func() *automock.WebhookConverter
		InputRequest       *webhookclient.FormationAssignmentNotificationRequestExt
		ExpectedErrMessage string
	}{
		{
			Name: "success when webhook client call doesn't return error",
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, faRequestExt).Return(nil, nil)
				return client
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, formationconstraint.PreSendNotification, preJoinPointDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, formationconstraint.PostSendNotification, postJoinPointDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			WebhookConverter: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToModel", fixRuntimeWebhookGQLModel(WebhookID, RuntimeID)).Return(fixRuntimeWebhookModel(WebhookID, RuntimeID), nil).Once()
				return conv
			},
			InputRequest: faRequestExt,
		},
		{
			Name: "fail when webhook client call fails",
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, faRequestExt).Return(nil, testErr)
				return client
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, formationconstraint.PreSendNotification, preJoinPointDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			WebhookConverter: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToModel", fixRuntimeWebhookGQLModel(WebhookID, RuntimeID)).Return(fixRuntimeWebhookModel(WebhookID, RuntimeID), nil).Once()
				return conv
			},
			InputRequest:       faRequestExt,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "fail when enforcing POST constraints returns error",
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, faRequestExt).Return(nil, nil)
				return client
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, formationconstraint.PreSendNotification, preJoinPointDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, formationconstraint.PostSendNotification, postJoinPointDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			WebhookConverter: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToModel", fixRuntimeWebhookGQLModel(WebhookID, RuntimeID)).Return(fixRuntimeWebhookModel(WebhookID, RuntimeID), nil).Once()
				return conv
			},
			InputRequest:       faRequestExt,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "fail when enforcing PRE constraints returns error",
			ConstraintEngine: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, formationconstraint.PreSendNotification, preJoinPointDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			WebhookConverter: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToModel", fixRuntimeWebhookGQLModel(WebhookID, RuntimeID)).Return(fixRuntimeWebhookModel(WebhookID, RuntimeID), nil).Once()
				return conv
			},
			InputRequest:       faRequestExt,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "fail when converting gql webhook to model returns error",
			WebhookConverter: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToModel", fixRuntimeWebhookGQLModel(WebhookID, RuntimeID)).Return(nil, testErr).Once()
				return conv
			},
			InputRequest:       faRequestExt,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			webhookClient := unusedWebhookClient()
			if testCase.WebhookClientFN != nil {
				webhookClient = testCase.WebhookClientFN()
			}
			constraintEngine := &automock.ConstraintEngine{}
			if testCase.ConstraintEngine != nil {
				constraintEngine = testCase.ConstraintEngine()
			}
			webhookConverter := &automock.WebhookConverter{}
			if testCase.WebhookConverter != nil {
				webhookConverter = testCase.WebhookConverter()
			}

			notificationSvc := formation.NewNotificationService(nil, webhookClient, nil, constraintEngine, webhookConverter, nil)

			// WHEN
			_, err := notificationSvc.SendNotification(ctx, testCase.InputRequest)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}
			mock.AssertExpectationsForObjects(t, webhookClient, constraintEngine, webhookConverter)
		})
	}
}
