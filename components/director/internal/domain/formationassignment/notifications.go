package formationassignment

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"

	databuilder "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/pkg/errors"
)

// formationRepository represents the Formations repository layer
//go:generate mockery --exported --name=formationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationRepository interface {
	Get(ctx context.Context, id, tenantID string) (*model.Formation, error)
}

//go:generate mockery --exported --name=notificationBuilder --output=automock --outpkg=automock --case=underscore --disable-version-string
type notificationBuilder interface {
	BuildNotificationRequest(ctx context.Context, formationTemplateID string, joinPointDetails *formationconstraint.GenerateNotificationOperationDetails, webhook *model.Webhook) (*webhookclient.NotificationRequest, error)
	PrepareDetailsForConfigurationChangeNotificationGeneration(operation model.FormationOperation, formationID string, applicationTemplate *webhook.ApplicationTemplateWithLabels, application *webhook.ApplicationWithLabels, runtime *webhook.RuntimeWithLabels, runtimeContext *webhook.RuntimeContextWithLabels, assignment *webhook.FormationAssignment, reverseAssignment *webhook.FormationAssignment, targetType model.ResourceType) (*formationconstraint.GenerateNotificationOperationDetails, error)
	PrepareDetailsForApplicationTenantMappingNotificationGeneration(operation model.FormationOperation, formationID string, sourceApplicationTemplate *webhook.ApplicationTemplateWithLabels, sourceApplication *webhook.ApplicationWithLabels, targetApplicationTemplate *webhook.ApplicationTemplateWithLabels, targetApplication *webhook.ApplicationWithLabels, assignment *webhook.FormationAssignment, reverseAssignment *webhook.FormationAssignment) (*formationconstraint.GenerateNotificationOperationDetails, error)
}

type formationAssignmentNotificationService struct {
	formationAssignmentRepo FormationAssignmentRepository
	webhookConverter        webhookConverter
	webhookRepository       webhookRepository
	webhookDataInputBuilder databuilder.DataInputBuilder
	formationRepository     formationRepository
	notificationBuilder     notificationBuilder
}

// NewFormationAssignmentNotificationService creates formation assignment notifications service
func NewFormationAssignmentNotificationService(formationAssignmentRepo FormationAssignmentRepository, webhookConverter webhookConverter, webhookRepository webhookRepository, webhookDataInputBuilder databuilder.DataInputBuilder, formationRepository formationRepository, notificationBuilder notificationBuilder) *formationAssignmentNotificationService {
	return &formationAssignmentNotificationService{
		formationAssignmentRepo: formationAssignmentRepo,
		webhookConverter:        webhookConverter,
		webhookRepository:       webhookRepository,
		webhookDataInputBuilder: webhookDataInputBuilder,
		formationRepository:     formationRepository,
		notificationBuilder:     notificationBuilder,
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

	referencedFormation, err := fan.formationRepository.Get(ctx, fa.FormationID, fa.TenantID)
	if err != nil {
		return nil, err
	}

	appWebhook, err := fan.webhookRepository.GetByIDAndWebhookType(ctx, tenant, appID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			log.C(ctx).Infof("There is no configuration changed webhook for runtime with ID: %q. There are no notifications to be generated.", appID)
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while getting configuration changed webhook for runtime with ID: %q", appID)
	}

	applicationWithLabels, appTemplateWithLabels, err := fan.webhookDataInputBuilder.PrepareApplicationAndAppTemplateWithLabels(ctx, tenant, appID)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}

	if fa.SourceType == model.FormationAssignmentTypeApplication {
		reverseAppID := fa.Source
		log.C(ctx).Infof("The formation assignment reverse object type is %q and has ID: %q", model.FormationAssignmentTypeApplication, reverseAppID)

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

		details, err := fan.notificationBuilder.PrepareDetailsForApplicationTenantMappingNotificationGeneration(
			model.AssignFormation,
			fa.FormationID,
			reverseAppTemplateWithLabels,
			reverseAppWithLabels,
			appTemplateWithLabels,
			applicationWithLabels,
			convertFormationAssignmentFromModel(fa),
			convertFormationAssignmentFromModel(reverseFA),
		)
		if err != nil {
			return nil, err
		}

		notificationReq, err := fan.notificationBuilder.BuildNotificationRequest(ctx, referencedFormation.FormationTemplateID, details, appWebhook)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		return notificationReq, nil
	} else if fa.SourceType == model.FormationAssignmentTypeRuntime {
		runtimeID := fa.Source
		log.C(ctx).Infof("The formation assignment reverse object type is %q and has ID: %q", model.FormationAssignmentTypeRuntime, runtimeID)

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

		details, err := fan.notificationBuilder.PrepareDetailsForConfigurationChangeNotificationGeneration(
			model.AssignFormation,
			fa.FormationID,
			appTemplateWithLabels,
			applicationWithLabels,
			runtimeWithLabels,
			runtimeContextWithLabels,
			convertFormationAssignmentFromModel(fa),
			convertFormationAssignmentFromModel(reverseFA),
			model.ApplicationResourceType)
		if err != nil {
			return nil, err
		}

		notificationReq, err := fan.notificationBuilder.BuildNotificationRequest(ctx, referencedFormation.FormationTemplateID, details, appWebhook)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		return notificationReq, nil
	} else {
		runtimeCtxID := fa.Source
		log.C(ctx).Infof("The formation assignment reverse object type is %q and has ID: %q", model.FormationAssignmentTypeRuntimeContext, runtimeCtxID)

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

		details, err := fan.notificationBuilder.PrepareDetailsForConfigurationChangeNotificationGeneration(
			model.AssignFormation,
			fa.FormationID,
			appTemplateWithLabels,
			applicationWithLabels,
			runtimeWithLabels,
			runtimeContextWithLabels,
			convertFormationAssignmentFromModel(fa),
			convertFormationAssignmentFromModel(reverseFA),
			model.ApplicationResourceType)
		if err != nil {
			return nil, err
		}

		notificationReq, err := fan.notificationBuilder.BuildNotificationRequest(ctx, referencedFormation.FormationTemplateID, details, appWebhook)
		if err != nil {
			return nil, err
		}

		return notificationReq, nil
	}
}

//todo logs
// generateRuntimeFANotification generates runtime formation assignment notification based on the reverse(source) type of the formation assignment
func (fan *formationAssignmentNotificationService) generateRuntimeFANotification(ctx context.Context, fa *model.FormationAssignment) (*webhookclient.NotificationRequest, error) {
	tenant := fa.TenantID
	runtimeID := fa.Target

	referencedFormation, err := fan.formationRepository.Get(ctx, fa.FormationID, fa.TenantID)
	if err != nil {
		return nil, err
	}

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

	details, err := fan.notificationBuilder.PrepareDetailsForConfigurationChangeNotificationGeneration(
		model.AssignFormation,
		fa.FormationID,
		appTemplateWithLabels,
		applicationWithLabels,
		runtimeWithLabels,
		nil,
		convertFormationAssignmentFromModel(fa),
		convertFormationAssignmentFromModel(reverseFA),
		model.RuntimeResourceType)
	if err != nil {
		return nil, err
	}

	notificationReq, err := fan.notificationBuilder.BuildNotificationRequest(ctx, referencedFormation.FormationTemplateID, details, runtimeWebhook)
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

	referencedFormation, err := fan.formationRepository.Get(ctx, fa.FormationID, fa.TenantID)
	if err != nil {
		return nil, err
	}

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

	details, err := fan.notificationBuilder.PrepareDetailsForConfigurationChangeNotificationGeneration(
		model.AssignFormation,
		fa.FormationID,
		appTemplateWithLabels,
		applicationWithLabels,
		runtimeWithLabels,
		runtimeContextWithLabels,
		convertFormationAssignmentFromModel(fa),
		convertFormationAssignmentFromModel(reverseFA),
		model.RuntimeContextResourceType)
	if err != nil {
		return nil, err
	}

	notificationReq, err := fan.notificationBuilder.BuildNotificationRequest(ctx, referencedFormation.FormationTemplateID, details, runtimeWebhook)
	if err != nil {
		return nil, err
	}

	return notificationReq, nil
}

func convertFormationAssignmentFromModel(formationAssignment *model.FormationAssignment) *webhook.FormationAssignment {
	config := string(formationAssignment.Value)
	if config == "" {
		config = "\"\""
	}
	return &webhook.FormationAssignment{
		ID:          formationAssignment.ID,
		FormationID: formationAssignment.FormationID,
		TenantID:    formationAssignment.TenantID,
		Source:      formationAssignment.Source,
		SourceType:  formationAssignment.SourceType,
		Target:      formationAssignment.Target,
		TargetType:  formationAssignment.TargetType,
		State:       formationAssignment.State,
		Value:       config,
	}
}
