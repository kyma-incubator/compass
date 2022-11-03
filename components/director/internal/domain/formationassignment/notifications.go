package formationassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/pkg/errors"
)

type formationAssignmentNotificationService struct {
	applicationRepository         applicationRepository
	applicationTemplateRepository applicationTemplateRepository
	runtimeRepo                   runtimeRepository
	runtimeContextRepo            runtimeContextRepository
	labelRepository               labelRepository
	webhookRepository             webhookRepository
	webhookConverter              webhookConverter
}

// NewFormationAssignmentNotificationService creates formation assignment notifications service
func NewFormationAssignmentNotificationService(applicationRepo applicationRepository, applicationTemplateRepository applicationTemplateRepository, runtimeRepo runtimeRepository, runtimeContextRepo runtimeContextRepository, labelRepository labelRepository, webhookRepository webhookRepository, webhookConverter webhookConverter) *formationAssignmentNotificationService {
	return &formationAssignmentNotificationService{
		applicationRepository:         applicationRepo,
		applicationTemplateRepository: applicationTemplateRepository,
		runtimeRepo:                   runtimeRepo,
		runtimeContextRepo:            runtimeContextRepo,
		labelRepository:               labelRepository,
		webhookRepository:             webhookRepository,
		webhookConverter:              webhookConverter,
	}
}

// GenerateNotification generates notifications by provided model.FormationAssignment
func (fan *formationAssignmentNotificationService) GenerateNotification(ctx context.Context, fa *model.FormationAssignment) (*webhookclient.NotificationRequest, error) {
	log.C(ctx).Infof("Generating notification for formation assignment with ID: %q and target type: %q and target ID: %q", fa.ID, fa.TargetType, fa.Target)
	switch fa.TargetType {
	case model.FormationAssignmentTypeApplication:
		return fan.generateApplicationFANotification(ctx, fa)
	case model.FormationAssignmentTypeRuntime:
		return fan.generateRuntimeFANotification(ctx, fa)
	case model.FormationAssignmentTypeRuntimeContext:
		return fan.generateRuntimeContextFANotification(ctx, fa)
	default:
		return nil, errors.Errorf("Unknown formation assignment type: %q", fa.TargetType)
	}
}

// generateApplicationFANotification generates application formation assignment notification based on the reverse(source) type of the formation assignment
func (fan *formationAssignmentNotificationService) generateApplicationFANotification(ctx context.Context, fa *model.FormationAssignment) (*webhookclient.NotificationRequest, error) {
	tenant := fa.TenantID
	appID := fa.Target

	appWebhook, err := fan.webhookRepository.GetByIDAndWebhookType(ctx, tenant, appID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			log.C(ctx).Infof("There is no configuration changed webhook for runtime with ID: %q. There are no notifications to be generated.", appID)
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while getting configuration changed webhook for runtime with ID: %q", appID)
	}

	if fa.SourceType == model.FormationAssignmentTypeApplication {
		reverseAppID := fa.Source
		log.C(ctx).Infof("The formation assignment reverse object type is %q and has ID: %q", model.FormationAssignmentTypeApplication, reverseAppID)

		appWithLabels, appTemplateWithLabels, err := fan.prepareApplicationAndAppTemplateWithLabels(ctx, tenant, appID)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		reverseAppWithLabels, reverseAppTemplateWithLabels, err := fan.prepareApplicationAndAppTemplateWithLabels(ctx, tenant, reverseAppID)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		whInput := &webhook.ApplicationTenantMappingInput{
			Operation:                 model.AssignFormation,
			FormationID:               fa.FormationID,
			SourceApplicationTemplate: reverseAppTemplateWithLabels,
			SourceApplication:         reverseAppWithLabels,
			TargetApplicationTemplate: appTemplateWithLabels,
			TargetApplication:         appWithLabels,
			Assignment:                convertFormationAssignmentFromModel(fa),
			ReverseAssignment:         convertFormationAssignmentFromModel(buildReverseFormationAssignment(fa)),
		}

		notificationReq, err := fan.createWebhookRequest(ctx, appWebhook, whInput)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		return notificationReq, nil
	} else if fa.SourceType == model.FormationAssignmentTypeRuntime {
		runtimeID := fa.Source
		log.C(ctx).Infof("The formation assignment reverse object type is %q and has ID: %q", model.FormationAssignmentTypeRuntime, runtimeID)

		applicationWithLabels, appTemplateWithLabels, err := fan.prepareApplicationAndAppTemplateWithLabels(ctx, tenant, appID)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		runtimeWithLabels, runtimeContextWithLabels, err := fan.prepareRuntimeAndRuntimeContextWithLabels(ctx, tenant, runtimeID)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		whInput := &webhook.FormationConfigurationChangeInput{
			Operation:           model.AssignFormation,
			FormationID:         fa.FormationID,
			ApplicationTemplate: appTemplateWithLabels,
			Application:         applicationWithLabels,
			Runtime:             runtimeWithLabels,
			RuntimeContext:      runtimeContextWithLabels,
			Assignment:          convertFormationAssignmentFromModel(fa),
			ReverseAssignment:   convertFormationAssignmentFromModel(buildReverseFormationAssignment(fa)),
		}

		notificationReq, err := fan.createWebhookRequest(ctx, appWebhook, whInput)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		return notificationReq, nil
	} else {
		runtimeCtxID := fa.Source
		log.C(ctx).Infof("The formation assignment reverse object type is %q and has ID: %q", model.FormationAssignmentTypeRuntimeContext, runtimeCtxID)

		applicationWithLabels, appTemplateWithLabels, err := fan.prepareApplicationAndAppTemplateWithLabels(ctx, tenant, appID)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		runtimeContextWithLabels, err := fan.prepareRuntimeContextWithLabels(ctx, tenant, runtimeCtxID)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		runtimeID := runtimeContextWithLabels.RuntimeContext.RuntimeID
		runtimeWithLabels, err := fan.prepareRuntimeWithLabels(ctx, tenant, runtimeID)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		whInput := &webhook.FormationConfigurationChangeInput{
			Operation:           model.AssignFormation,
			FormationID:         fa.FormationID,
			ApplicationTemplate: appTemplateWithLabels,
			Application:         applicationWithLabels,
			Runtime:             runtimeWithLabels,
			RuntimeContext:      runtimeContextWithLabels,
			Assignment:          convertFormationAssignmentFromModel(fa),
			ReverseAssignment:   convertFormationAssignmentFromModel(buildReverseFormationAssignment(fa)),
		}

		notificationReq, err := fan.createWebhookRequest(ctx, appWebhook, whInput)
		if err != nil {
			return nil, err
		}

		return notificationReq, nil
	}
}

// generateRuntimeFANotification generates runtime formation assignment notification based on the reverse(source) type of the formation assignment
func (fan *formationAssignmentNotificationService) generateRuntimeFANotification(ctx context.Context, fa *model.FormationAssignment) (*webhookclient.NotificationRequest, error) {
	tenant := fa.TenantID
	runtimeID := fa.Target

	runtimeWebhook, err := fan.webhookRepository.GetByIDAndWebhookType(ctx, tenant, runtimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			log.C(ctx).Infof("There is no configuration changed webhook for runtime with ID: %q. There are no notifications to be generated", runtimeID)
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while getting configuration changed webhook for runtime with ID: %q", runtimeID)
	}

	if fa.SourceType != model.FormationAssignmentTypeApplication {
		log.C(ctx).Errorf("The formation assignmet with ID: %q and target type: %q has unsupported reverse(source) type: %q", fa.ID, fa.TargetType, fa.SourceType)
		return nil, errors.Errorf("The formation assignmet with ID: %q and target type: %q has unsupported reverse(source) type: %q", fa.ID, fa.TargetType, fa.SourceType)
	}

	appID := fa.Source
	log.C(ctx).Infof("The formation assignment reverse object type is %q and has ID: %q", model.FormationAssignmentTypeApplication, appID)

	applicationWithLabels, appTemplateWithLabels, err := fan.prepareApplicationAndAppTemplateWithLabels(ctx, tenant, appID)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}

	runtimeWithLabels, err := fan.prepareRuntimeWithLabels(ctx, tenant, runtimeID)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}

	whInput := &webhook.FormationConfigurationChangeInput{
		Operation:           model.AssignFormation,
		FormationID:         fa.FormationID,
		ApplicationTemplate: appTemplateWithLabels,
		Application:         applicationWithLabels,
		Runtime:             runtimeWithLabels,
		RuntimeContext:      nil,
		Assignment:          convertFormationAssignmentFromModel(fa),
		ReverseAssignment:   convertFormationAssignmentFromModel(buildReverseFormationAssignment(fa)),
	}

	notificationReq, err := fan.createWebhookRequest(ctx, runtimeWebhook, whInput)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}

	return notificationReq, nil
}

// generateRuntimeContextFANotification generates runtime context formation assignment notification based on the reverse(source) type of the formation assignment
func (fan *formationAssignmentNotificationService) generateRuntimeContextFANotification(ctx context.Context, fa *model.FormationAssignment) (*webhookclient.NotificationRequest, error) {
	tenant := fa.TenantID
	runtimeCtxID := fa.Target

	runtimeContextWithLabels, err := fan.prepareRuntimeContextWithLabels(ctx, tenant, runtimeCtxID)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}

	runtimeID := runtimeContextWithLabels.RuntimeContext.RuntimeID
	runtimeWebhook, err := fan.webhookRepository.GetByIDAndWebhookType(ctx, tenant, runtimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			log.C(ctx).Infof("There is no configuration changed webhook for runtime with ID: %q. There are no notifications to be generated", runtimeID)
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while getting configuration changed webhook for runtime with ID: %q", runtimeID)
	}

	if fa.SourceType != model.FormationAssignmentTypeApplication {
		log.C(ctx).Errorf("The formation assignmet with ID: %q and target type: %q has unsupported reverse(source) type: %q", fa.ID, fa.TargetType, fa.SourceType)
		return nil, errors.Errorf("The formation assignmet with ID: %q and target type: %q has unsupported reverse(source) type: %q", fa.ID, fa.TargetType, fa.SourceType)
	}

	appID := fa.Source
	log.C(ctx).Infof("The formation assignment reverse object type is %q and has ID: %q", model.FormationAssignmentTypeApplication, appID)

	applicationWithLabels, appTemplateWithLabels, err := fan.prepareApplicationAndAppTemplateWithLabels(ctx, tenant, appID)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}

	runtimeWithLabels, err := fan.prepareRuntimeWithLabels(ctx, tenant, runtimeID)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}

	whInput := &webhook.FormationConfigurationChangeInput{
		Operation:           model.AssignFormation,
		FormationID:         fa.FormationID,
		ApplicationTemplate: appTemplateWithLabels,
		Application:         applicationWithLabels,
		Runtime:             runtimeWithLabels,
		RuntimeContext:      runtimeContextWithLabels,
		Assignment:          convertFormationAssignmentFromModel(fa),
		ReverseAssignment:   convertFormationAssignmentFromModel(buildReverseFormationAssignment(fa)),
	}

	notificationReq, err := fan.createWebhookRequest(ctx, runtimeWebhook, whInput)
	if err != nil {
		return nil, err
	}

	return notificationReq, nil
}

func (fan *formationAssignmentNotificationService) prepareApplicationAndAppTemplateWithLabels(ctx context.Context, tenant, appID string) (*webhook.ApplicationWithLabels, *webhook.ApplicationTemplateWithLabels, error) {
	application, err := fan.applicationRepository.GetByID(ctx, tenant, appID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while getting application with ID: %q", appID)
	}
	applicationLabels, err := fan.getLabelsForObject(ctx, tenant, appID, model.ApplicationLabelableObject)
	if err != nil {
		return nil, nil, err
	}
	applicationWithLabels := &webhook.ApplicationWithLabels{
		Application: application,
		Labels:      applicationLabels,
	}

	var appTemplateWithLabels *webhook.ApplicationTemplateWithLabels
	if application.ApplicationTemplateID != nil {
		appTemplate, err := fan.applicationTemplateRepository.Get(ctx, *application.ApplicationTemplateID)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "while getting application template with ID: %q", *application.ApplicationTemplateID)
		}
		applicationTemplateLabels, err := fan.getLabelsForObject(ctx, tenant, appTemplate.ID, model.AppTemplateLabelableObject)
		if err != nil {
			return nil, nil, err
		}
		appTemplateWithLabels = &webhook.ApplicationTemplateWithLabels{
			ApplicationTemplate: appTemplate,
			Labels:              applicationTemplateLabels,
		}
	}
	return applicationWithLabels, appTemplateWithLabels, nil
}

func (fan *formationAssignmentNotificationService) prepareRuntimeWithLabels(ctx context.Context, tenant, runtimeID string) (*webhook.RuntimeWithLabels, error) {
	runtime, err := fan.runtimeRepo.GetByID(ctx, tenant, runtimeID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtime by ID: %q", runtimeID)
	}

	runtimeLabels, err := fan.getLabelsForObject(ctx, tenant, runtimeID, model.RuntimeLabelableObject)
	if err != nil {
		return nil, err
	}

	runtimeWithLabels := &webhook.RuntimeWithLabels{
		Runtime: runtime,
		Labels:  runtimeLabels,
	}

	return runtimeWithLabels, nil
}

func (fan *formationAssignmentNotificationService) prepareRuntimeContextWithLabels(ctx context.Context, tenant, runtimeCtxID string) (*webhook.RuntimeContextWithLabels, error) {
	runtimeCtx, err := fan.runtimeContextRepo.GetByID(ctx, tenant, runtimeCtxID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtime context by ID: %q", runtimeCtxID)
	}

	runtimeCtxLabels, err := fan.getLabelsForObject(ctx, tenant, runtimeCtx.ID, model.RuntimeContextLabelableObject)
	if err != nil {
		return nil, err
	}

	runtimeContextWithLabels := &webhook.RuntimeContextWithLabels{
		RuntimeContext: runtimeCtx,
		Labels:         runtimeCtxLabels,
	}

	return runtimeContextWithLabels, nil
}

func (fan *formationAssignmentNotificationService) prepareRuntimeAndRuntimeContextWithLabels(ctx context.Context, tenant, runtimeID string) (*webhook.RuntimeWithLabels, *webhook.RuntimeContextWithLabels, error) {
	runtimeWithLabels, err := fan.prepareRuntimeWithLabels(ctx, tenant, runtimeID)
	if err != nil {
		return nil, nil, err
	}

	runtimeCtx, err := fan.runtimeContextRepo.GetByRuntimeID(ctx, tenant, runtimeID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while getting runtime context for runtime with ID: %q", runtimeID)
	}

	runtimeCtxLabels, err := fan.getLabelsForObject(ctx, tenant, runtimeCtx.ID, model.RuntimeContextLabelableObject)
	if err != nil {
		return nil, nil, err
	}

	runtimeContextWithLabels := &webhook.RuntimeContextWithLabels{
		RuntimeContext: runtimeCtx,
		Labels:         runtimeCtxLabels,
	}

	return runtimeWithLabels, runtimeContextWithLabels, nil
}

func (fan *formationAssignmentNotificationService) getLabelsForObject(ctx context.Context, tenant, objectID string, objectType model.LabelableObject) (map[string]interface{}, error) {
	labels, err := fan.labelRepository.ListForObject(ctx, tenant, objectType, objectID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing labels for %q with ID: %q", objectType, objectID)
	}
	labelsMap := make(map[string]interface{}, len(labels))
	for _, l := range labels {
		labelsMap[l.Key] = l.Value
	}
	return labelsMap, nil
}

func (fan *formationAssignmentNotificationService) createWebhookRequest(ctx context.Context, webhook *model.Webhook, input webhook.FormationAssignmentTemplateInput) (*webhookclient.NotificationRequest, error) {
	gqlWebhook, err := fan.webhookConverter.ToGraphQL(webhook)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting webhook with ID %s", webhook.ID)
	}
	return &webhookclient.NotificationRequest{
		Webhook:       *gqlWebhook,
		Object:        input,
		CorrelationID: correlation.CorrelationIDFromContext(ctx),
	}, nil
}

func convertFormationAssignmentFromModel(formationAssignment *model.FormationAssignment) *webhook.FormationAssignment {
	return &webhook.FormationAssignment{
		ID:          formationAssignment.ID,
		FormationID: formationAssignment.FormationID,
		TenantID:    formationAssignment.TenantID,
		Source:      formationAssignment.Source,
		SourceType:  formationAssignment.SourceType,
		Target:      formationAssignment.Target,
		TargetType:  formationAssignment.TargetType,
		State:       formationAssignment.State,
		Value:       string(formationAssignment.Value),
	}
}

func buildReverseFormationAssignment(formationAssignment *model.FormationAssignment) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          formationAssignment.ID,
		FormationID: formationAssignment.FormationID,
		TenantID:    formationAssignment.TenantID,
		Source:      formationAssignment.Source,
		SourceType:  formationAssignment.SourceType,
		Target:      formationAssignment.Target,
		TargetType:  formationAssignment.TargetType,
		State:       formationAssignment.State,
		Value:       formationAssignment.Value,
	}
}
