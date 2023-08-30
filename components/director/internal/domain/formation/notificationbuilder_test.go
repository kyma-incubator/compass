package formation_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	webhookdir "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBuildFormationAssignmentNotificationRequest(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	ApplicationID := "04f3568d-3e0c-4f6b-b646-e6979e9d060c"
	webhook := fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)
	gqlWebhook := &graphql.Webhook{ID: WebhookID, ApplicationID: &ApplicationID, Type: graphql.WebhookTypeConfigurationChanged}

	applicationTenantMappingWebhook := fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)
	gqlApplicationTenantMappingWebhook := &graphql.Webhook{ID: AppTenantMappingWebhookIDForApp1, ApplicationID: &ApplicationID, Type: graphql.WebhookTypeApplicationTenantMapping}

	testCases := []struct {
		Name                 string
		WebhookConverter     func() *automock.WebhookConverter
		ConstraintEngineFn   func() *automock.ConstraintEngine
		Details              *formationconstraint.GenerateFormationAssignmentNotificationOperationDetails
		Webhook              *model.Webhook
		ExpectedNotification *webhookclient.FormationAssignmentNotificationRequest
		ExpectedErrMessage   string
	}{
		{
			Name: "success for configuration changed webhook",
			WebhookConverter: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", webhook).Return(gqlWebhook, nil).Once()
				return conv
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preGenerateFormationAssignmentNotificationLocation, generateConfigurationChangeNotificationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postGenerateFormationAssignmentNotificationLocation, generateConfigurationChangeNotificationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			Details:              generateConfigurationChangeNotificationDetails,
			Webhook:              webhook,
			ExpectedNotification: applicationNotificationWithAppTemplate,
			ExpectedErrMessage:   "",
		},
		{
			Name: "success for application tenant mapping webhook",
			WebhookConverter: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", applicationTenantMappingWebhook).Return(gqlApplicationTenantMappingWebhook, nil).Once()
				return conv
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preGenerateFormationAssignmentNotificationLocation, generateAppToAppNotificationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postGenerateFormationAssignmentNotificationLocation, generateAppToAppNotificationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			Details:              generateAppToAppNotificationDetails,
			Webhook:              applicationTenantMappingWebhook,
			ExpectedNotification: appToAppNotificationWithoutSourceTemplateWithTargetTemplate,
			ExpectedErrMessage:   "",
		},
		{
			Name: "error while enforcing constraints pre operation",
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preGenerateFormationAssignmentNotificationLocation, generateConfigurationChangeNotificationDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			Details: generateConfigurationChangeNotificationDetails,
			Webhook: webhook,
		},
		{
			Name: "error when webhook type is not supported",
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preGenerateFormationAssignmentNotificationLocation, generateConfigurationChangeNotificationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			Details:            generateConfigurationChangeNotificationDetails,
			Webhook:            &model.Webhook{Type: model.WebhookTypeRegisterApplication},
			ExpectedErrMessage: "Unsupported webhook type",
		},
		{
			Name: "error while converting webhook",
			WebhookConverter: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", webhook).Return(nil, testErr).Once()
				return conv
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preGenerateFormationAssignmentNotificationLocation, generateConfigurationChangeNotificationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			Details:            generateConfigurationChangeNotificationDetails,
			Webhook:            webhook,
			ExpectedErrMessage: "while converting webhook with ID",
		},
		{
			Name: "error while enforcing constraints post operation",
			WebhookConverter: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", webhook).Return(gqlWebhook, nil).Once()
				return conv
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preGenerateFormationAssignmentNotificationLocation, generateConfigurationChangeNotificationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postGenerateFormationAssignmentNotificationLocation, generateConfigurationChangeNotificationDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			Details: generateConfigurationChangeNotificationDetails,
			Webhook: webhook,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			webhookConverter := unusedWebhookConverter()
			if testCase.WebhookConverter != nil {
				webhookConverter = testCase.WebhookConverter()
			}
			constraintEngine := unusedConstraintEngine()
			if testCase.ConstraintEngineFn != nil {
				constraintEngine = testCase.ConstraintEngineFn()
			}

			builder := formation.NewNotificationsBuilder(webhookConverter, constraintEngine, runtimeType, applicationType)

			// WHEN
			actual, err := builder.BuildFormationAssignmentNotificationRequest(ctx, testCase.Details, testCase.Webhook)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedNotification, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, webhookConverter, constraintEngine)
		})
	}
}

func TestBuildFormationNotificationRequests(t *testing.T) {
	ctx := context.Background()
	formationLifecycleGQLWebhook := fixFormationLifecycleWebhookGQLModel(FormationLifecycleWebhookID, FormationTemplateID, graphql.WebhookModeSync)

	testCases := []struct {
		name                              string
		constraintEngineFn                func() *automock.ConstraintEngine
		webhookConverterFn                func() *automock.WebhookConverter
		formationTemplateWebhooks         []*model.Webhook
		expectedErrMsg                    string
		expectedFormationNotificationReqs []*webhookclient.FormationNotificationRequest
	}{
		{
			name: "Successfully build formation notification requests",
			constraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preGenerateFormationNotificationLocation, formationNotificationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postGenerateFormationNotificationLocation, formationNotificationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookConverterFn: func() *automock.WebhookConverter {
				webhookConv := &automock.WebhookConverter{}
				webhookConv.On("ToGraphQL", formationLifecycleSyncWebhook).Return(&formationLifecycleGQLWebhook, nil).Once()
				return webhookConv
			},
			formationTemplateWebhooks:         formationLifecycleSyncWebhooks,
			expectedFormationNotificationReqs: formationNotificationSyncCreateRequests,
		},
		{
			name: "Error when enforcing pre generate formation notification constraints",
			constraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preGenerateFormationNotificationLocation, formationNotificationDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			formationTemplateWebhooks: formationLifecycleSyncWebhooks,
			expectedErrMsg:            testErr.Error(),
		},
		{
			name: "Success when there are no formation template webhooks",
			constraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preGenerateFormationNotificationLocation, formationNotificationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			formationTemplateWebhooks: emptyFormationLifecycleWebhooks,
		},
		{
			name: "Error when converting formation template webhook to graphql one",
			constraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preGenerateFormationNotificationLocation, formationNotificationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookConverterFn: func() *automock.WebhookConverter {
				webhookConv := &automock.WebhookConverter{}
				webhookConv.On("ToGraphQL", formationLifecycleSyncWebhook).Return(nil, testErr).Once()
				return webhookConv
			},
			formationTemplateWebhooks: formationLifecycleSyncWebhooks,
			expectedErrMsg:            testErr.Error(),
		},
		{
			name: "Error when enforcing post generate formation notification constraints",
			constraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preGenerateFormationNotificationLocation, formationNotificationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postGenerateFormationNotificationLocation, formationNotificationDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			webhookConverterFn: func() *automock.WebhookConverter {
				webhookConv := &automock.WebhookConverter{}
				webhookConv.On("ToGraphQL", formationLifecycleSyncWebhook).Return(&formationLifecycleGQLWebhook, nil).Once()
				return webhookConv
			},
			formationTemplateWebhooks: formationLifecycleSyncWebhooks,
			expectedErrMsg:            testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			constraintEngine := unusedConstraintEngine()
			if testCase.constraintEngineFn != nil {
				constraintEngine = testCase.constraintEngineFn()
			}

			webhookConv := unusedWebhookConverter()
			if testCase.webhookConverterFn != nil {
				webhookConv = testCase.webhookConverterFn()
			}

			defer mock.AssertExpectationsForObjects(t, constraintEngine, webhookConv)

			builder := formation.NewNotificationsBuilder(webhookConv, constraintEngine, runtimeType, applicationType)
			formationNotificationReqs, err := builder.BuildFormationNotificationRequests(ctx, formationNotificationDetails, formationModelWithoutError, testCase.formationTemplateWebhooks)

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

func TestNotificationBuilder_PrepareDetailsForConfigurationChangeNotificationGeneration(t *testing.T) {
	applicationTemplate := &webhookdir.ApplicationTemplateWithLabels{
		ApplicationTemplate: fixApplicationTemplateModel(),
		Labels:              fixApplicationTemplateLabelsMap(),
	}

	application := &webhookdir.ApplicationWithLabels{
		Application: fixApplicationModel(ApplicationID),
		Labels: map[string]string{
			applicationType: applicationType,
		},
	}

	runtime := &webhookdir.RuntimeWithLabels{
		Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
		Labels: map[string]string{
			runtimeType: runtimeType,
		},
	}

	applicationWithoutType := &webhookdir.ApplicationWithLabels{
		Application: fixApplicationModel(ApplicationID),
		Labels:      map[string]string{},
	}

	runtimeWithoutType := &webhookdir.RuntimeWithLabels{
		Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
		Labels:  map[string]string{},
	}

	runtimeContext := &webhookdir.RuntimeContextWithLabels{
		RuntimeContext: fixRuntimeContextModel(),
		Labels:         fixRuntimeContextLabelsMap(),
	}

	testCases := []struct {
		Name                        string
		Operation                   model.FormationOperation
		FormationID                 string
		ApplicationTemplate         *webhookdir.ApplicationTemplateWithLabels
		Application                 *webhookdir.ApplicationWithLabels
		Runtime                     *webhookdir.RuntimeWithLabels
		RuntimeContext              *webhookdir.RuntimeContextWithLabels
		Assignment                  *webhookdir.FormationAssignment
		ReverseAssignment           *webhookdir.FormationAssignment
		TargetType                  model.ResourceType
		ExpectedNotificationDetails *formationconstraint.GenerateFormationAssignmentNotificationOperationDetails
		ExpectedErrMessage          string
	}{
		{
			Name:                "success resource type application",
			Operation:           model.AssignFormation,
			FormationID:         fixUUID(),
			ApplicationTemplate: applicationTemplate,
			Application:         application,
			Runtime:             runtime,
			RuntimeContext:      runtimeContext,
			Assignment:          emptyFormationAssignment,
			ReverseAssignment:   emptyFormationAssignment,
			TargetType:          model.ApplicationResourceType,
			ExpectedNotificationDetails: &formationconstraint.GenerateFormationAssignmentNotificationOperationDetails{
				Operation:           model.AssignFormation,
				FormationTemplateID: FormationTemplateID,
				ResourceType:        model.ApplicationResourceType,
				ResourceSubtype:     applicationType,
				ResourceID:          ApplicationID,
				Formation:           formationModelWithoutError,
				ApplicationTemplate: applicationTemplate,
				Application:         application,
				Runtime:             runtime,
				RuntimeContext:      runtimeContext,
				Assignment:          emptyFormationAssignment,
				ReverseAssignment:   emptyFormationAssignment,
				TenantID:            tenantID.String(),
			},
			ExpectedErrMessage: "",
		},
		{
			Name:                "success resource type runtime",
			Operation:           model.AssignFormation,
			FormationID:         fixUUID(),
			ApplicationTemplate: applicationTemplate,
			Application:         application,
			Runtime:             runtime,
			RuntimeContext:      runtimeContext,
			Assignment:          emptyFormationAssignment,
			ReverseAssignment:   emptyFormationAssignment,
			TargetType:          model.RuntimeResourceType,
			ExpectedNotificationDetails: &formationconstraint.GenerateFormationAssignmentNotificationOperationDetails{
				Operation:           model.AssignFormation,
				FormationTemplateID: FormationTemplateID,
				ResourceType:        model.RuntimeResourceType,
				ResourceSubtype:     runtimeType,
				ResourceID:          RuntimeContextRuntimeID,
				Formation:           formationModelWithoutError,
				ApplicationTemplate: applicationTemplate,
				Application:         application,
				Runtime:             runtime,
				RuntimeContext:      runtimeContext,
				Assignment:          emptyFormationAssignment,
				ReverseAssignment:   emptyFormationAssignment,
				TenantID:            tenantID.String(),
			},
			ExpectedErrMessage: "",
		},
		{
			Name:                "success resource type runtime context",
			Operation:           model.AssignFormation,
			FormationID:         fixUUID(),
			ApplicationTemplate: applicationTemplate,
			Application:         application,
			Runtime:             runtime,
			RuntimeContext:      runtimeContext,
			Assignment:          emptyFormationAssignment,
			ReverseAssignment:   emptyFormationAssignment,
			TargetType:          model.RuntimeContextResourceType,
			ExpectedNotificationDetails: &formationconstraint.GenerateFormationAssignmentNotificationOperationDetails{
				Operation:           model.AssignFormation,
				FormationTemplateID: FormationTemplateID,
				ResourceType:        model.RuntimeContextResourceType,
				ResourceSubtype:     runtimeType,
				ResourceID:          RuntimeContextID,
				Formation:           formationModelWithoutError,
				ApplicationTemplate: applicationTemplate,
				Application:         application,
				Runtime:             runtime,
				RuntimeContext:      runtimeContext,
				Assignment:          emptyFormationAssignment,
				ReverseAssignment:   emptyFormationAssignment,
				TenantID:            tenantID.String(),
			},
			ExpectedErrMessage: "",
		},
		{
			Name:                "fail to determine application type",
			Operation:           model.AssignFormation,
			FormationID:         fixUUID(),
			ApplicationTemplate: applicationTemplate,
			Application:         applicationWithoutType,
			Runtime:             runtime,
			RuntimeContext:      runtimeContext,
			Assignment:          emptyFormationAssignment,
			ReverseAssignment:   emptyFormationAssignment,
			TargetType:          model.ApplicationResourceType,
			ExpectedErrMessage:  "While determining subtype for application with ID",
		},
		{
			Name:                "fail to determine runtime type",
			Operation:           model.AssignFormation,
			FormationID:         fixUUID(),
			ApplicationTemplate: applicationTemplate,
			Application:         application,
			Runtime:             runtimeWithoutType,
			RuntimeContext:      runtimeContext,
			Assignment:          emptyFormationAssignment,
			ReverseAssignment:   emptyFormationAssignment,
			TargetType:          model.RuntimeResourceType,
			ExpectedErrMessage:  "While determining subtype for runtime with ID",
		},
		{
			Name:                "fail to determine runtime context type",
			Operation:           model.AssignFormation,
			FormationID:         fixUUID(),
			ApplicationTemplate: applicationTemplate,
			Application:         application,
			Runtime:             runtimeWithoutType,
			RuntimeContext:      runtimeContext,
			Assignment:          emptyFormationAssignment,
			ReverseAssignment:   emptyFormationAssignment,
			TargetType:          model.RuntimeContextResourceType,
			ExpectedErrMessage:  "While determining subtype for runtime context with ID",
		},
		{
			Name:                "unsupported resource type",
			Operation:           model.AssignFormation,
			FormationID:         fixUUID(),
			ApplicationTemplate: applicationTemplate,
			Application:         application,
			Runtime:             runtimeWithoutType,
			RuntimeContext:      runtimeContext,
			Assignment:          emptyFormationAssignment,
			ReverseAssignment:   emptyFormationAssignment,
			TargetType:          model.FormationResourceType,
			ExpectedErrMessage:  "Unsuported target resource subtype \"FORMATION\"",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			builder := formation.NewNotificationsBuilder(nil, nil, runtimeType, applicationType)

			// WHEN
			actual, err := builder.PrepareDetailsForConfigurationChangeNotificationGeneration(testCase.Operation, FormationTemplateID, formationModelWithoutError, testCase.ApplicationTemplate, testCase.Application, testCase.Runtime, testCase.RuntimeContext, testCase.Assignment, testCase.ReverseAssignment, testCase.TargetType, nil, tenantID.String())

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedNotificationDetails, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}
		})
	}
}

func TestNotificationBuilder_PrepareDetailsForApplicationTenantMappingNotificationGeneration(t *testing.T) {
	applicationTemplate := &webhookdir.ApplicationTemplateWithLabels{
		ApplicationTemplate: fixApplicationTemplateModel(),
		Labels:              fixApplicationTemplateLabelsMap(),
	}

	application := &webhookdir.ApplicationWithLabels{
		Application: fixApplicationModel(ApplicationID),
		Labels: map[string]string{
			applicationType: applicationType,
		},
	}

	applicationWithoutType := &webhookdir.ApplicationWithLabels{
		Application: fixApplicationModel(ApplicationID),
		Labels:      map[string]string{},
	}

	testCases := []struct {
		Name                        string
		Operation                   model.FormationOperation
		SourceApplicationTemplate   *webhookdir.ApplicationTemplateWithLabels
		SourceApplication           *webhookdir.ApplicationWithLabels
		TargetApplicationTemplate   *webhookdir.ApplicationTemplateWithLabels
		TargetApplication           *webhookdir.ApplicationWithLabels
		Assignment                  *webhookdir.FormationAssignment
		ReverseAssignment           *webhookdir.FormationAssignment
		TargetType                  model.ResourceType
		ExpectedNotificationDetails *formationconstraint.GenerateFormationAssignmentNotificationOperationDetails
		ExpectedErrMessage          string
	}{
		{
			Name:                      "success resource type application",
			Operation:                 model.AssignFormation,
			SourceApplicationTemplate: applicationTemplate,
			SourceApplication:         application,
			TargetApplicationTemplate: applicationTemplate,
			TargetApplication:         application,
			Assignment:                emptyFormationAssignment,
			ReverseAssignment:         emptyFormationAssignment,
			TargetType:                model.ApplicationResourceType,
			ExpectedNotificationDetails: &formationconstraint.GenerateFormationAssignmentNotificationOperationDetails{
				Operation:                 model.AssignFormation,
				FormationTemplateID:       FormationTemplateID,
				ResourceType:              model.ApplicationResourceType,
				ResourceSubtype:           applicationType,
				ResourceID:                ApplicationID,
				Formation:                 formationModelWithoutError,
				SourceApplicationTemplate: applicationTemplate,
				SourceApplication:         application,
				TargetApplicationTemplate: applicationTemplate,
				TargetApplication:         application,
				Assignment:                emptyFormationAssignment,
				ReverseAssignment:         emptyFormationAssignment,
				TenantID:                  tenantID.String(),
			},
			ExpectedErrMessage: "",
		},
		{
			Name:                      "fail to determine target application type",
			Operation:                 model.AssignFormation,
			SourceApplicationTemplate: applicationTemplate,
			SourceApplication:         application,
			TargetApplicationTemplate: applicationTemplate,
			TargetApplication:         applicationWithoutType,
			Assignment:                emptyFormationAssignment,
			ReverseAssignment:         emptyFormationAssignment,
			TargetType:                model.ApplicationResourceType,
			ExpectedErrMessage:        "While determining subtype for application with",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			builder := formation.NewNotificationsBuilder(nil, nil, runtimeType, applicationType)

			// WHEN
			actual, err := builder.PrepareDetailsForApplicationTenantMappingNotificationGeneration(testCase.Operation, FormationTemplateID, formationModelWithoutError, testCase.SourceApplicationTemplate, testCase.SourceApplication, testCase.TargetApplicationTemplate, testCase.TargetApplication, testCase.Assignment, testCase.ReverseAssignment, nil, tenantID.String())

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedNotificationDetails, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}
		})
	}
}
