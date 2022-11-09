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
	formationAssignmentRepo FormationAssignmentRepository
	webhookConverter        webhookConverter
	webhookRepository       webhookRepository
	webhookDataInputBuilder webhook.DataInputBuilder
}

// NewFormationAssignmentNotificationService creates formation assignment notifications service
func NewFormationAssignmentNotificationService(formationAssignmentRepo FormationAssignmentRepository, webhookConverter webhookConverter, webhookRepository webhookRepository, webhookDataInputBuilder webhook.DataInputBuilder) *formationAssignmentNotificationService {
	return &formationAssignmentNotificationService{
		formationAssignmentRepo: formationAssignmentRepo,
		webhookConverter:        webhookConverter,
		webhookRepository:       webhookRepository,
		webhookDataInputBuilder: webhookDataInputBuilder,
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

		appWithLabels, appTemplateWithLabels, err := fan.webhookDataInputBuilder.PrepareApplicationAndAppTemplateWithLabels(ctx, tenant, appID)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		reverseAppWithLabels, reverseAppTemplateWithLabels, err := fan.webhookDataInputBuilder.PrepareApplicationAndAppTemplateWithLabels(ctx, tenant, reverseAppID)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		reverseFA, err := fan.formationAssignmentRepo.GetReverseBySourceAndTarget(ctx, tenant, fa.FormationID, fa.Source, fa.Target)
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
			ReverseAssignment:         convertFormationAssignmentFromModel(reverseFA),
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

		applicationWithLabels, appTemplateWithLabels, err := fan.webhookDataInputBuilder.PrepareApplicationAndAppTemplateWithLabels(ctx, tenant, appID)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		runtimeWithLabels, runtimeContextWithLabels, err := fan.webhookDataInputBuilder.PrepareRuntimeAndRuntimeContextWithLabels(ctx, tenant, runtimeID)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		reverseFA, err := fan.formationAssignmentRepo.GetReverseBySourceAndTarget(ctx, tenant, fa.FormationID, fa.Source, fa.Target)
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
			ReverseAssignment:   convertFormationAssignmentFromModel(reverseFA),
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

		applicationWithLabels, appTemplateWithLabels, err := fan.webhookDataInputBuilder.PrepareApplicationAndAppTemplateWithLabels(ctx, tenant, appID)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		runtimeContextWithLabels, err := fan.webhookDataInputBuilder.PrepareRuntimeContextWithLabels(ctx, tenant, runtimeCtxID)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		runtimeID := runtimeContextWithLabels.RuntimeContext.RuntimeID
		runtimeWithLabels, err := fan.webhookDataInputBuilder.PrepareRuntimeWithLabels(ctx, tenant, runtimeID)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		reverseFA, err := fan.formationAssignmentRepo.GetReverseBySourceAndTarget(ctx, tenant, fa.FormationID, fa.Source, fa.Target)
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
			ReverseAssignment:   convertFormationAssignmentFromModel(reverseFA),
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

	applicationWithLabels, appTemplateWithLabels, err := fan.webhookDataInputBuilder.PrepareApplicationAndAppTemplateWithLabels(ctx, tenant, appID)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}

	runtimeWithLabels, err := fan.webhookDataInputBuilder.PrepareRuntimeWithLabels(ctx, tenant, runtimeID)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}

	reverseFA, err := fan.formationAssignmentRepo.GetReverseBySourceAndTarget(ctx, tenant, fa.FormationID, fa.Source, fa.Target)
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
		ReverseAssignment:   convertFormationAssignmentFromModel(reverseFA),
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

	runtimeContextWithLabels, err := fan.webhookDataInputBuilder.PrepareRuntimeContextWithLabels(ctx, tenant, runtimeCtxID)
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

	applicationWithLabels, appTemplateWithLabels, err := fan.webhookDataInputBuilder.PrepareApplicationAndAppTemplateWithLabels(ctx, tenant, appID)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}

	runtimeWithLabels, err := fan.webhookDataInputBuilder.PrepareRuntimeWithLabels(ctx, tenant, runtimeID)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}

	reverseFA, err := fan.formationAssignmentRepo.GetReverseBySourceAndTarget(ctx, tenant, fa.FormationID, fa.Source, fa.Target)
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
		ReverseAssignment:   convertFormationAssignmentFromModel(reverseFA),
	}

	notificationReq, err := fan.createWebhookRequest(ctx, runtimeWebhook, whInput)
	if err != nil {
		return nil, err
	}

	return notificationReq, nil
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
