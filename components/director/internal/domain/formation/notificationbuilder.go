package formation

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	webhookdir "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/pkg/errors"
)

// NotificationBuilder is responsible for building notification requests
type NotificationBuilder struct {
	webhookConverter        webhookConverter
	constraintEngine        constraintEngine
	runtimeTypeLabelKey     string
	applicationTypeLabelKey string
}

// NewNotificationsBuilder creates new NotificationBuilder
func NewNotificationsBuilder(webhookConverter webhookConverter, constraintEngine constraintEngine, runtimeTypeLabelKey, applicationTypeLabelKey string) *NotificationBuilder {
	return &NotificationBuilder{
		webhookConverter:        webhookConverter,
		constraintEngine:        constraintEngine,
		runtimeTypeLabelKey:     runtimeTypeLabelKey,
		applicationTypeLabelKey: applicationTypeLabelKey,
	}
}

// BuildNotificationRequest builds new notification request
func (nb *NotificationBuilder) BuildNotificationRequest(
	ctx context.Context,
	formationTemplateID string,
	joinPointDetails *formationconstraintpkg.GenerateNotificationOperationDetails,
	webhook *model.Webhook,
) (*webhookclient.NotificationRequest, error) {
	log.C(ctx).Infof("Building notification request")
	if err := nb.constraintEngine.EnforceConstraints(ctx, formationconstraintpkg.JoinPointLocation{
		OperationName:  model.GenerateNotificationOperation,
		ConstraintType: model.PreOperation,
	}, joinPointDetails, formationTemplateID); err != nil {
		return nil, errors.Wrapf(err, "While enforcing constraints for target operation %q and constraint type %q", model.GenerateNotificationOperation, model.PreOperation)
	}

	inputBuilder, err := getInputBuilder(webhook.Type)
	if err != nil {
		return nil, err
	}

	req, err := nb.createWebhookRequest(ctx, webhook, inputBuilder(joinPointDetails))
	if err != nil {
		return nil, errors.Wrapf(err, "while creating webhook request")
	}

	if err := nb.constraintEngine.EnforceConstraints(ctx, formationconstraintpkg.JoinPointLocation{
		OperationName:  model.GenerateNotificationOperation,
		ConstraintType: model.PostOperation,
	}, joinPointDetails, formationTemplateID); err != nil {
		return nil, errors.Wrapf(err, "While enforcing constraints for target operation %q and constraint type %q", model.GenerateNotificationOperation, model.PostOperation)
	}

	return req, nil
}

func getInputBuilder(webhookType model.WebhookType) (InputBuilder, error) {
	switch webhookType {
	case model.WebhookTypeConfigurationChanged:
		return buildConfigurationChangeInputFromJoinpointDetails, nil
	case model.WebhookTypeApplicationTenantMapping:
		return buildApplicationTenantMappingInputFromJoinpointDetails, nil
	default:
		return nil, errors.Errorf("Unsupported Webhook Type %q", webhookType)
	}
}

// PrepareDetailsForConfigurationChangeNotificationGeneration returns GenerateNotificationOperationDetails for ConfigurationChanged webhooks
func (nb *NotificationBuilder) PrepareDetailsForConfigurationChangeNotificationGeneration(
	operation model.FormationOperation,
	formationID string,
	applicationTemplate *webhookdir.ApplicationTemplateWithLabels,
	application *webhookdir.ApplicationWithLabels,
	runtime *webhookdir.RuntimeWithLabels,
	runtimeContext *webhookdir.RuntimeContextWithLabels,
	assignment *webhookdir.FormationAssignment,
	reverseAssignment *webhookdir.FormationAssignment,
	targetType model.ResourceType,
	tenantContext *webhookdir.CustomerTenantContext,
) (*formationconstraintpkg.GenerateNotificationOperationDetails, error) {
	details := &formationconstraintpkg.GenerateNotificationOperationDetails{
		Operation:             operation,
		FormationID:           formationID,
		CustomerTenantContext: tenantContext,
		ApplicationTemplate:   applicationTemplate,
		Application:           application,
		Runtime:               runtime,
		RuntimeContext:        runtimeContext,
		Assignment:            assignment,
		ReverseAssignment:     reverseAssignment,
		ResourceType:          targetType,
	}
	switch targetType {
	case model.ApplicationResourceType:
		details.ResourceID = application.ID

		subtype, err := determineResourceSubtype(application.Labels, nb.applicationTypeLabelKey)
		if err != nil {
			return nil, errors.Wrapf(err, "While determining subtype for application with ID %q", application.ID)
		}

		details.ResourceSubtype = subtype
	case model.RuntimeResourceType:
		details.ResourceID = runtime.ID

		subtype, err := determineResourceSubtype(runtime.Labels, nb.runtimeTypeLabelKey)
		if err != nil {
			return nil, errors.Wrapf(err, "While determining subtype for runtime with ID %q", runtime.ID)
		}

		details.ResourceSubtype = subtype
	case model.RuntimeContextResourceType:
		details.ResourceID = runtimeContext.ID

		subtype, err := determineResourceSubtype(runtime.Labels, nb.runtimeTypeLabelKey)
		if err != nil {
			return nil, errors.Wrapf(err, "While determining subtype for runtime context with ID %q", runtimeContext.ID)
		}

		details.ResourceSubtype = subtype
	default:
		return nil, errors.Errorf("Unsuported target resource subtype %q", targetType)
	}

	return details, nil
}

// PrepareDetailsForApplicationTenantMappingNotificationGeneration returns GenerateNotificationOperationDetails for applicationTenantMapping webhooks
func (nb *NotificationBuilder) PrepareDetailsForApplicationTenantMappingNotificationGeneration(
	operation model.FormationOperation,
	formationID string,
	sourceApplicationTemplate *webhookdir.ApplicationTemplateWithLabels,
	sourceApplication *webhookdir.ApplicationWithLabels,
	targetApplicationTemplate *webhookdir.ApplicationTemplateWithLabels,
	targetApplication *webhookdir.ApplicationWithLabels,
	assignment *webhookdir.FormationAssignment,
	reverseAssignment *webhookdir.FormationAssignment,
	tenantContext *webhookdir.CustomerTenantContext,
) (*formationconstraintpkg.GenerateNotificationOperationDetails, error) {
	details := &formationconstraintpkg.GenerateNotificationOperationDetails{
		Operation:                 operation,
		FormationID:               formationID,
		CustomerTenantContext:     tenantContext,
		SourceApplicationTemplate: sourceApplicationTemplate,
		SourceApplication:         sourceApplication,
		TargetApplicationTemplate: targetApplicationTemplate,
		TargetApplication:         targetApplication,
		Assignment:                assignment,
		ReverseAssignment:         reverseAssignment,
		ResourceType:              model.ApplicationResourceType,
		ResourceID:                targetApplication.ID,
	}

	subtype, err := determineResourceSubtype(targetApplication.Labels, nb.applicationTypeLabelKey)
	if err != nil {
		return nil, errors.Wrapf(err, "While determining subtype for application with ID %q", targetApplication.ID)
	}

	details.ResourceSubtype = subtype

	return details, nil
}

func (nb *NotificationBuilder) createWebhookRequest(ctx context.Context, webhook *model.Webhook, input webhookdir.FormationAssignmentTemplateInput) (*webhookclient.NotificationRequest, error) {
	gqlWebhook, err := nb.webhookConverter.ToGraphQL(webhook)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting webhook with ID %s", webhook.ID)
	}
	return &webhookclient.NotificationRequest{
		Webhook:       *gqlWebhook,
		Object:        input,
		CorrelationID: correlation.CorrelationIDFromContext(ctx),
	}, nil
}

// InputBuilder represents expected signature for methods that create operator input from the provided details
type InputBuilder func(details *formationconstraintpkg.GenerateNotificationOperationDetails) webhookdir.FormationAssignmentTemplateInput

func buildConfigurationChangeInputFromJoinpointDetails(details *formationconstraintpkg.GenerateNotificationOperationDetails) webhookdir.FormationAssignmentTemplateInput {
	return &webhookdir.FormationConfigurationChangeInput{
		Operation:             details.Operation,
		FormationID:           details.FormationID,
		ApplicationTemplate:   details.ApplicationTemplate,
		Application:           details.Application,
		Runtime:               details.Runtime,
		RuntimeContext:        details.RuntimeContext,
		CustomerTenantContext: details.CustomerTenantContext,
		Assignment:            details.Assignment,
		ReverseAssignment:     details.ReverseAssignment,
	}
}

func buildApplicationTenantMappingInputFromJoinpointDetails(details *formationconstraintpkg.GenerateNotificationOperationDetails) webhookdir.FormationAssignmentTemplateInput {
	return &webhookdir.ApplicationTenantMappingInput{
		Operation:                 details.Operation,
		FormationID:               details.FormationID,
		SourceApplicationTemplate: details.SourceApplicationTemplate,
		SourceApplication:         details.SourceApplication,
		TargetApplicationTemplate: details.TargetApplicationTemplate,
		TargetApplication:         details.TargetApplication,
		CustomerTenantContext:     details.CustomerTenantContext,
		Assignment:                details.Assignment,
		ReverseAssignment:         details.ReverseAssignment,
	}
}

func determineResourceSubtype(labels map[string]interface{}, labelKey string) (string, error) {
	labelValue, ok := labels[labelKey]
	if !ok {
		return "", errors.Errorf("Missing %q label", labelKey)
	}

	subtype, ok := labelValue.(string)
	if !ok {
		return "", errors.Errorf("Failed to convert %q label to string", labelKey)
	}

	return subtype, nil
}
