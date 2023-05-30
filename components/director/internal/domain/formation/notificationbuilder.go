package formation

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	webhookdir "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=webhookConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type webhookConverter interface {
	ToGraphQL(in *model.Webhook) (*graphql.Webhook, error)
	ToModel(in *graphql.Webhook) (*model.Webhook, error)
}

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

// BuildFormationAssignmentNotificationRequest builds new formation assignment notification request
func (nb *NotificationBuilder) BuildFormationAssignmentNotificationRequest(
	ctx context.Context,
	formationTemplateID string,
	joinPointDetails *formationconstraintpkg.GenerateFormationAssignmentNotificationOperationDetails,
	webhook *model.Webhook,
) (*webhookclient.FormationAssignmentNotificationRequest, error) {
	log.C(ctx).Infof("Building formation assignment notification request...")
	if err := nb.constraintEngine.EnforceConstraints(ctx, formationconstraintpkg.PreGenerateFormationAssignmentNotifications, joinPointDetails, formationTemplateID); err != nil {
		return nil, errors.Wrapf(err, "While enforcing constraints for target operation %q and constraint type %q", model.GenerateFormationAssignmentNotificationOperation, model.PreOperation)
	}

	faInputBuilder, err := getFormationAssignmentInputBuilder(webhook.Type)
	if err != nil {
		return nil, err
	}

	req, err := nb.createWebhookRequest(ctx, webhook, faInputBuilder(joinPointDetails))
	if err != nil {
		return nil, errors.Wrapf(err, "while creating webhook request")
	}

	if err := nb.constraintEngine.EnforceConstraints(ctx, formationconstraintpkg.PostGenerateFormationAssignmentNotifications, joinPointDetails, formationTemplateID); err != nil {
		return nil, errors.Wrapf(err, "While enforcing constraints for target operation %q and constraint type %q", model.GenerateFormationAssignmentNotificationOperation, model.PostOperation)
	}

	return req, nil
}

// BuildFormationNotificationRequests builds new formation notification request
func (nb *NotificationBuilder) BuildFormationNotificationRequests(ctx context.Context, joinPointDetails *formationconstraintpkg.GenerateFormationNotificationOperationDetails, formation *model.Formation, formationTemplateWebhooks []*model.Webhook) ([]*webhookclient.FormationNotificationRequest, error) {
	log.C(ctx).Infof("Building formation notification request for formation with name: %q...", formation.Name)
	if err := nb.constraintEngine.EnforceConstraints(ctx, formationconstraintpkg.PreGenerateFormationNotifications, joinPointDetails, joinPointDetails.FormationTemplateID); err != nil {
		return nil, errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.GenerateFormationNotificationOperation, model.PreOperation)
	}

	if len(formationTemplateWebhooks) == 0 {
		log.C(ctx).Infof("Formation template with ID: %q does not have any webhooks. No notifications will be generated.", joinPointDetails.FormationTemplateID)
		return nil, nil
	}
	log.C(ctx).Infof("Formation template with ID: %q has/have %d webhook(s) of type: %q", joinPointDetails.FormationTemplateID, len(formationTemplateWebhooks), model.WebhookTypeFormationLifecycle)

	formationTemplateInput := buildFormationLifecycleInput(joinPointDetails, formation)

	requests := make([]*webhookclient.FormationNotificationRequest, 0, len(formationTemplateWebhooks))
	for _, webhook := range formationTemplateWebhooks {
		gqlWebhook, err := nb.webhookConverter.ToGraphQL(webhook)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting formation template webhook with ID: %s to graphql one", webhook.ID)
		}

		req := &webhookclient.FormationNotificationRequest{
			Request: webhookclient.NewRequest(
				*gqlWebhook,
				formationTemplateInput,
				correlation.CorrelationIDFromContext(ctx),
			),
		}
		requests = append(requests, req)
	}

	if err := nb.constraintEngine.EnforceConstraints(ctx, formationconstraintpkg.PostGenerateFormationNotifications, joinPointDetails, joinPointDetails.FormationTemplateID); err != nil {
		return nil, errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.GenerateFormationNotificationOperation, model.PostOperation)
	}

	return requests, nil
}

// PrepareDetailsForConfigurationChangeNotificationGeneration returns GenerateFormationAssignmentNotificationOperationDetails for ConfigurationChanged webhooks
func (nb *NotificationBuilder) PrepareDetailsForConfigurationChangeNotificationGeneration(
	operation model.FormationOperation,
	formationID string,
	formationTemplateID string,
	applicationTemplate *webhookdir.ApplicationTemplateWithLabels,
	application *webhookdir.ApplicationWithLabels,
	runtime *webhookdir.RuntimeWithLabels,
	runtimeContext *webhookdir.RuntimeContextWithLabels,
	assignment *webhookdir.FormationAssignment,
	reverseAssignment *webhookdir.FormationAssignment,
	targetType model.ResourceType,
	tenantContext *webhookdir.CustomerTenantContext,
	tenantID string,
) (*formationconstraintpkg.GenerateFormationAssignmentNotificationOperationDetails, error) {
	details := &formationconstraintpkg.GenerateFormationAssignmentNotificationOperationDetails{
		Operation:             operation,
		FormationID:           formationID,
		FormationTemplateID:   formationTemplateID,
		CustomerTenantContext: tenantContext,
		ApplicationTemplate:   applicationTemplate,
		Application:           application,
		Runtime:               runtime,
		RuntimeContext:        runtimeContext,
		Assignment:            assignment,
		ReverseAssignment:     reverseAssignment,
		ResourceType:          targetType,
		TenantID:              tenantID,
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

// PrepareDetailsForApplicationTenantMappingNotificationGeneration returns GenerateFormationAssignmentNotificationOperationDetails for applicationTenantMapping webhooks
func (nb *NotificationBuilder) PrepareDetailsForApplicationTenantMappingNotificationGeneration(
	operation model.FormationOperation,
	formationID string,
	formationTemplateID string,
	sourceApplicationTemplate *webhookdir.ApplicationTemplateWithLabels,
	sourceApplication *webhookdir.ApplicationWithLabels,
	targetApplicationTemplate *webhookdir.ApplicationTemplateWithLabels,
	targetApplication *webhookdir.ApplicationWithLabels,
	assignment *webhookdir.FormationAssignment,
	reverseAssignment *webhookdir.FormationAssignment,
	tenantContext *webhookdir.CustomerTenantContext,
	tenantID string,
) (*formationconstraintpkg.GenerateFormationAssignmentNotificationOperationDetails, error) {
	details := &formationconstraintpkg.GenerateFormationAssignmentNotificationOperationDetails{
		Operation:                 operation,
		FormationID:               formationID,
		FormationTemplateID:       formationTemplateID,
		CustomerTenantContext:     tenantContext,
		SourceApplicationTemplate: sourceApplicationTemplate,
		SourceApplication:         sourceApplication,
		TargetApplicationTemplate: targetApplicationTemplate,
		TargetApplication:         targetApplication,
		Assignment:                assignment,
		ReverseAssignment:         reverseAssignment,
		ResourceType:              model.ApplicationResourceType,
		ResourceID:                targetApplication.ID,
		TenantID:                  tenantID,
	}

	subtype, err := determineResourceSubtype(targetApplication.Labels, nb.applicationTypeLabelKey)
	if err != nil {
		return nil, errors.Wrapf(err, "While determining subtype for application with ID %q", targetApplication.ID)
	}

	details.ResourceSubtype = subtype

	return details, nil
}

func (nb *NotificationBuilder) createWebhookRequest(ctx context.Context, webhook *model.Webhook, formationAssignmentTemplateInput webhookdir.FormationAssignmentTemplateInput) (*webhookclient.FormationAssignmentNotificationRequest, error) {
	gqlWebhook, err := nb.webhookConverter.ToGraphQL(webhook)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting webhook with ID %s", webhook.ID)
	}
	return &webhookclient.FormationAssignmentNotificationRequest{
		Webhook:       *gqlWebhook,
		Object:        formationAssignmentTemplateInput,
		CorrelationID: correlation.CorrelationIDFromContext(ctx),
	}, nil
}

func getFormationAssignmentInputBuilder(webhookType model.WebhookType) (FormationAssignmentInputBuilder, error) {
	switch webhookType {
	case model.WebhookTypeConfigurationChanged:
		return buildConfigurationChangeInputFromJoinpointDetails, nil
	case model.WebhookTypeApplicationTenantMapping:
		return buildApplicationTenantMappingInputFromJoinpointDetails, nil
	default:
		return nil, errors.Errorf("Unsupported Webhook Type %q", webhookType)
	}
}

// FormationAssignmentInputBuilder represents expected signature for methods that create operator input from the provided details
type FormationAssignmentInputBuilder func(details *formationconstraintpkg.GenerateFormationAssignmentNotificationOperationDetails) webhookdir.FormationAssignmentTemplateInput

func buildConfigurationChangeInputFromJoinpointDetails(details *formationconstraintpkg.GenerateFormationAssignmentNotificationOperationDetails) webhookdir.FormationAssignmentTemplateInput {
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

func buildApplicationTenantMappingInputFromJoinpointDetails(details *formationconstraintpkg.GenerateFormationAssignmentNotificationOperationDetails) webhookdir.FormationAssignmentTemplateInput {
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

func buildFormationLifecycleInput(details *formationconstraintpkg.GenerateFormationNotificationOperationDetails, formation *model.Formation) *webhookdir.FormationLifecycleInput {
	return &webhookdir.FormationLifecycleInput{
		Operation:             details.Operation,
		Formation:             formation,
		CustomerTenantContext: details.CustomerTenantContext,
	}
}

func determineResourceSubtype(labels map[string]string, labelKey string) (string, error) {
	labelValue, ok := labels[labelKey]
	if !ok {
		return "", errors.Errorf("Missing %q label", labelKey)
	}

	return labelValue, nil
}
