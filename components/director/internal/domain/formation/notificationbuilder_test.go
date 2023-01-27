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
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNotificationBuilderBuildNotificationRequest(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, Tnt, ExternalTnt)

	testErr := errors.New("test error")

	ApplicationID := "04f3568d-3e0c-4f6b-b646-e6979e9d060c"

	webhook := fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)
	gqlWebhook := &graphql.Webhook{ID: WebhookID, ApplicationID: &ApplicationID, Type: graphql.WebhookTypeConfigurationChanged}

	applicationTenantMappingWebhook := fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)
	gqlApplicationTenantMappingWebhook := &graphql.Webhook{ID: AppTenantMappingWebhookIDForApp1, ApplicationID: &ApplicationID, Type: graphql.WebhookTypeApplicationTenantMapping}

	testCases := []struct {
		Name                 string
		WebhookConverter     func() *automock.WebhookConverter
		ConstraintEngineFn   func() *automock.ConstraintEngine
		Details              *formationconstraint.GenerateNotificationOperationDetails
		Webhook              *model.Webhook
		ExpectedNotification *webhookclient.NotificationRequest
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
				engine.On("EnforceConstraints", ctx, preGenerateNotificationLocation, generateConfigurationChangeNotificationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postGenerateNotificationLocation, generateConfigurationChangeNotificationDetails, FormationTemplateID).Return(nil).Once()
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
				engine.On("EnforceConstraints", ctx, preGenerateNotificationLocation, generateAppToAppNotificationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postGenerateNotificationLocation, generateAppToAppNotificationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			Details:              generateAppToAppNotificationDetails,
			Webhook:              applicationTenantMappingWebhook,
			ExpectedNotification: appToAppNotificationWithSourceTemplate,
			ExpectedErrMessage:   "",
		},
		{
			Name: "error while enforcing constraints pre operation",
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preGenerateNotificationLocation, generateConfigurationChangeNotificationDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			Details:            generateConfigurationChangeNotificationDetails,
			Webhook:            webhook,
			ExpectedErrMessage: "While enforcing constraints for target operation \"GENERATE_NOTIFICATION\" and constraint type \"PRE\": test error",
		},
		{
			Name: "error when webhook type is not supported",
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preGenerateNotificationLocation, generateConfigurationChangeNotificationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			Details:            generateConfigurationChangeNotificationDetails,
			Webhook:            &model.Webhook{Type: model.WebhookTypeRegisterApplication},
			ExpectedErrMessage: "Unsupported Webhook Type",
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
				engine.On("EnforceConstraints", ctx, preGenerateNotificationLocation, generateConfigurationChangeNotificationDetails, FormationTemplateID).Return(nil).Once()
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
				engine.On("EnforceConstraints", ctx, preGenerateNotificationLocation, generateConfigurationChangeNotificationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postGenerateNotificationLocation, generateConfigurationChangeNotificationDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			Details:            generateConfigurationChangeNotificationDetails,
			Webhook:            webhook,
			ExpectedErrMessage: "While enforcing constraints for target operation \"GENERATE_NOTIFICATION\" and constraint type \"POST\": test error",
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
			actual, err := builder.BuildNotificationRequest(ctx, FormationTemplateID, testCase.Details, testCase.Webhook)

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

func TestNotificationBuilder_PrepareDetailsForConfigurationChangeNotificationGeneration(t *testing.T) {
	applicationTemplate := &webhookdir.ApplicationTemplateWithLabels{
		ApplicationTemplate: fixApplicationTemplateModel(),
		Labels:              fixApplicationTemplateLabelsMap(),
	}

	application := &webhookdir.ApplicationWithLabels{
		Application: fixApplicationModel(ApplicationID),
		Labels: map[string]interface{}{
			applicationType: applicationType,
		},
	}

	runtime := &webhookdir.RuntimeWithLabels{
		Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
		Labels: map[string]interface{}{
			runtimeType: runtimeType,
		},
	}

	applicationWithoutType := &webhookdir.ApplicationWithLabels{
		Application: fixApplicationModel(ApplicationID),
		Labels:      map[string]interface{}{},
	}

	runtimeWithoutType := &webhookdir.RuntimeWithLabels{
		Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
		Labels:  map[string]interface{}{},
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
		ExpectedNotificationDetails *formationconstraint.GenerateNotificationOperationDetails
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
			ExpectedNotificationDetails: &formationconstraint.GenerateNotificationOperationDetails{
				Operation:           model.AssignFormation,
				FormationID:         fixUUID(),
				ResourceType:        model.ApplicationResourceType,
				ResourceSubtype:     applicationType,
				ResourceID:          ApplicationID,
				ApplicationTemplate: applicationTemplate,
				Application:         application,
				Runtime:             runtime,
				RuntimeContext:      runtimeContext,
				Assignment:          emptyFormationAssignment,
				ReverseAssignment:   emptyFormationAssignment,
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
			ExpectedNotificationDetails: &formationconstraint.GenerateNotificationOperationDetails{
				Operation:           model.AssignFormation,
				FormationID:         fixUUID(),
				ResourceType:        model.RuntimeResourceType,
				ResourceSubtype:     runtimeType,
				ResourceID:          RuntimeContextRuntimeID,
				ApplicationTemplate: applicationTemplate,
				Application:         application,
				Runtime:             runtime,
				RuntimeContext:      runtimeContext,
				Assignment:          emptyFormationAssignment,
				ReverseAssignment:   emptyFormationAssignment,
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
			ExpectedNotificationDetails: &formationconstraint.GenerateNotificationOperationDetails{
				Operation:           model.AssignFormation,
				FormationID:         fixUUID(),
				ResourceType:        model.RuntimeContextResourceType,
				ResourceSubtype:     runtimeType,
				ResourceID:          RuntimeContextID,
				ApplicationTemplate: applicationTemplate,
				Application:         application,
				Runtime:             runtime,
				RuntimeContext:      runtimeContext,
				Assignment:          emptyFormationAssignment,
				ReverseAssignment:   emptyFormationAssignment,
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
			actual, err := builder.PrepareDetailsForConfigurationChangeNotificationGeneration(testCase.Operation, testCase.FormationID, testCase.ApplicationTemplate, testCase.Application, testCase.Runtime, testCase.RuntimeContext, testCase.Assignment, testCase.ReverseAssignment, testCase.TargetType, nil)

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
		Labels: map[string]interface{}{
			applicationType: applicationType,
		},
	}

	applicationWithoutType := &webhookdir.ApplicationWithLabels{
		Application: fixApplicationModel(ApplicationID),
		Labels:      map[string]interface{}{},
	}

	testCases := []struct {
		Name                        string
		Operation                   model.FormationOperation
		FormationID                 string
		SourceApplicationTemplate   *webhookdir.ApplicationTemplateWithLabels
		SourceApplication           *webhookdir.ApplicationWithLabels
		TargetApplicationTemplate   *webhookdir.ApplicationTemplateWithLabels
		TargetApplication           *webhookdir.ApplicationWithLabels
		Assignment                  *webhookdir.FormationAssignment
		ReverseAssignment           *webhookdir.FormationAssignment
		TargetType                  model.ResourceType
		ExpectedNotificationDetails *formationconstraint.GenerateNotificationOperationDetails
		ExpectedErrMessage          string
	}{
		{
			Name:                      "success resource type application",
			Operation:                 model.AssignFormation,
			FormationID:               fixUUID(),
			SourceApplicationTemplate: applicationTemplate,
			SourceApplication:         application,
			TargetApplicationTemplate: applicationTemplate,
			TargetApplication:         application,
			Assignment:                emptyFormationAssignment,
			ReverseAssignment:         emptyFormationAssignment,
			TargetType:                model.ApplicationResourceType,
			ExpectedNotificationDetails: &formationconstraint.GenerateNotificationOperationDetails{
				Operation:                 model.AssignFormation,
				FormationID:               fixUUID(),
				ResourceType:              model.ApplicationResourceType,
				ResourceSubtype:           applicationType,
				ResourceID:                ApplicationID,
				SourceApplicationTemplate: applicationTemplate,
				SourceApplication:         application,
				TargetApplicationTemplate: applicationTemplate,
				TargetApplication:         application,
				Assignment:                emptyFormationAssignment,
				ReverseAssignment:         emptyFormationAssignment,
			},
			ExpectedErrMessage: "",
		},
		{
			Name:                      "fail to determine target application type",
			Operation:                 model.AssignFormation,
			FormationID:               fixUUID(),
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
			actual, err := builder.PrepareDetailsForApplicationTenantMappingNotificationGeneration(testCase.Operation, testCase.FormationID, testCase.SourceApplicationTemplate, testCase.SourceApplication, testCase.TargetApplicationTemplate, testCase.TargetApplication, testCase.Assignment, testCase.ReverseAssignment, nil)

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
