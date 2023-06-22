package formationassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

	databuilder "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/pkg/errors"
)

// formationRepository represents the Formations repository layer
//
//go:generate mockery --exported --name=formationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationRepository interface {
	Get(ctx context.Context, id, tenantID string) (*model.Formation, error)
}

//go:generate mockery --exported --name=notificationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type notificationService interface {
	SendNotification(ctx context.Context, webhookNotificationReq webhookclient.WebhookExtRequest) (*webhook.Response, error)
}

//go:generate mockery --exported --name=notificationBuilder --output=automock --outpkg=automock --case=underscore --disable-version-string
type notificationBuilder interface {
	BuildFormationAssignmentNotificationRequest(ctx context.Context, formationTemplateID string, joinPointDetails *formationconstraint.GenerateFormationAssignmentNotificationOperationDetails, webhook *model.Webhook) (*webhookclient.FormationAssignmentNotificationRequest, error)
	PrepareDetailsForConfigurationChangeNotificationGeneration(operation model.FormationOperation, formationID string, formationTemplateID string, applicationTemplate *webhook.ApplicationTemplateWithLabels, application *webhook.ApplicationWithLabels, runtime *webhook.RuntimeWithLabels, runtimeContext *webhook.RuntimeContextWithLabels, assignment *webhook.FormationAssignment, reverseAssignment *webhook.FormationAssignment, targetType model.ResourceType, tenantContext *webhook.CustomerTenantContext, tenantID string) (*formationconstraint.GenerateFormationAssignmentNotificationOperationDetails, error)
	PrepareDetailsForApplicationTenantMappingNotificationGeneration(operation model.FormationOperation, formationID string, formationTemplateID string, sourceApplicationTemplate *webhook.ApplicationTemplateWithLabels, sourceApplication *webhook.ApplicationWithLabels, targetApplicationTemplate *webhook.ApplicationTemplateWithLabels, targetApplication *webhook.ApplicationWithLabels, assignment *webhook.FormationAssignment, reverseAssignment *webhook.FormationAssignment, tenantContext *webhook.CustomerTenantContext, tenantID string) (*formationconstraint.GenerateFormationAssignmentNotificationOperationDetails, error)
}

type formationAssignmentNotificationService struct {
	formationAssignmentRepo FormationAssignmentRepository
	webhookConverter        webhookConverter
	webhookRepository       webhookRepository
	tenantRepository        tenantRepository
	webhookDataInputBuilder databuilder.DataInputBuilder
	formationRepository     formationRepository
	notificationBuilder     notificationBuilder
	runtimeContextRepo      runtimeContextRepository
	labelService            labelService
	runtimeTypeLabelKey     string
	applicationTypeLabelKey string
}

// NewFormationAssignmentNotificationService creates formation assignment notifications service
func NewFormationAssignmentNotificationService(formationAssignmentRepo FormationAssignmentRepository, webhookConverter webhookConverter, webhookRepository webhookRepository, tenantRepository tenantRepository, webhookDataInputBuilder databuilder.DataInputBuilder, formationRepository formationRepository, notificationBuilder notificationBuilder, runtimeContextRepo runtimeContextRepository, labelService labelService, runtimeTypeLabelKey, applicationTypeLabelKey string) *formationAssignmentNotificationService {
	return &formationAssignmentNotificationService{
		formationAssignmentRepo: formationAssignmentRepo,
		webhookConverter:        webhookConverter,
		webhookRepository:       webhookRepository,
		tenantRepository:        tenantRepository,
		webhookDataInputBuilder: webhookDataInputBuilder,
		formationRepository:     formationRepository,
		notificationBuilder:     notificationBuilder,
		runtimeContextRepo:      runtimeContextRepo,
		labelService:            labelService,
		runtimeTypeLabelKey:     runtimeTypeLabelKey,
		applicationTypeLabelKey: applicationTypeLabelKey,
	}
}

// GenerateFormationAssignmentNotification generates formation assignment notification by provided model.FormationAssignment
func (fan *formationAssignmentNotificationService) GenerateFormationAssignmentNotification(ctx context.Context, fa *model.FormationAssignment, operation model.FormationOperation) (*webhookclient.FormationAssignmentNotificationRequest, error) {
	log.C(ctx).Infof("Generating notification for formation assignment with ID: %q and target type: %q and target ID: %q", fa.ID, fa.TargetType, fa.Target)

	customerTenantContext, err := fan.extractCustomerTenantContext(ctx, fa.TenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while extracting customer tenant context for tenant with internal ID %s", fa.TenantID)
	}

	referencedFormation, err := fan.formationRepository.Get(ctx, fa.FormationID, fa.TenantID)
	if err != nil {
		return nil, err
	}

	switch fa.TargetType {
	case model.FormationAssignmentTypeApplication:
		return fan.generateApplicationFANotification(ctx, fa, referencedFormation, customerTenantContext, operation)
	case model.FormationAssignmentTypeRuntime:
		return fan.generateRuntimeFANotification(ctx, fa, referencedFormation, customerTenantContext, operation)
	case model.FormationAssignmentTypeRuntimeContext:
		return fan.generateRuntimeContextFANotification(ctx, fa, referencedFormation, customerTenantContext, operation)
	default:
		return nil, errors.Errorf("Unknown formation assignment type: %q", fa.TargetType)
	}
}

// PrepareDetailsForNotificationStatusReturned creates NotificationStatusReturnedOperationDetails by given tenantID, formation assignment and formation operation
func (fan *formationAssignmentNotificationService) PrepareDetailsForNotificationStatusReturned(ctx context.Context, tenantID string, fa *model.FormationAssignment, operation model.FormationOperation) (*formationconstraint.NotificationStatusReturnedOperationDetails, error) {
	var targetType model.ResourceType
	switch fa.TargetType {
	case model.FormationAssignmentTypeApplication:
		targetType = model.ApplicationResourceType
	case model.FormationAssignmentTypeRuntime:
		targetType = model.RuntimeResourceType
	case model.FormationAssignmentTypeRuntimeContext:
		targetType = model.RuntimeContextResourceType
	}

	targetSubtype, err := fan.getObjectSubtype(ctx, fa.TenantID, fa.Target, fa.TargetType)
	if err != nil {
		return nil, err
	}

	formation, err := fan.formationRepository.Get(ctx, fa.FormationID, tenantID)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting formation with ID %q in tenant %q: %v", fa.FormationID, tenantID, err)
		return nil, errors.Wrapf(err, "An error occurred while getting formation with ID %q in tenant %q", fa.FormationID, tenantID)
	}

	reverseFa, err := fan.getReverseBySourceAndTarget(ctx, tenantID, formation.ID, fa.Source, fa.Target)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			log.C(ctx).Errorf("An error occurred while getting reverse formation assignment: %v", err)
			return nil, errors.Wrap(err, "An error occurred while getting reverse formation assignment")
		}
		log.C(ctx).Debugf("Reverse assignment with source %q and target %q in formation with ID %q is not found.", fa.Target, fa.Source, formation.ID)
	}

	return &formationconstraint.NotificationStatusReturnedOperationDetails{
		ResourceType:               targetType,
		ResourceSubtype:            targetSubtype,
		Operation:                  operation,
		FormationAssignment:        fa,
		ReverseFormationAssignment: reverseFa,
		Formation:                  formation,
	}, nil
}

// GenerateFormationAssignmentNotificationExt generates extended formation assignment notification by given formation(and reverse formation) assignment request mapping and formation operation
func (fan *formationAssignmentNotificationService) GenerateFormationAssignmentNotificationExt(ctx context.Context, faRequestMapping, reverseFaRequestMapping *FormationAssignmentRequestMapping, operation model.FormationOperation) (*webhookclient.FormationAssignmentNotificationRequestExt, error) {
	targetSubtype, err := fan.getObjectSubtype(ctx, faRequestMapping.FormationAssignment.TenantID, faRequestMapping.FormationAssignment.Target, faRequestMapping.FormationAssignment.TargetType)
	if err != nil {
		return nil, err
	}

	formation, err := fan.formationRepository.Get(ctx, faRequestMapping.FormationAssignment.FormationID, faRequestMapping.FormationAssignment.TenantID)
	if err != nil {
		return nil, err
	}

	var reverseFa *model.FormationAssignment
	if reverseFaRequestMapping != nil {
		reverseFa = reverseFaRequestMapping.FormationAssignment
	}

	return &webhookclient.FormationAssignmentNotificationRequestExt{
		Operation:                              operation,
		FormationAssignmentNotificationRequest: faRequestMapping.Request,
		FormationAssignment:                    faRequestMapping.FormationAssignment,
		ReverseFormationAssignment:             reverseFa,
		Formation:                              formation,
		TargetSubtype:                          targetSubtype,
	}, nil
}

func (fan *formationAssignmentNotificationService) getObjectSubtype(ctx context.Context, tnt, objectID string, objectType model.FormationAssignmentType) (string, error) {
	switch objectType {
	case model.FormationAssignmentTypeApplication:
		applicationTypeLabel, err := fan.labelService.GetLabel(ctx, tnt, &model.LabelInput{
			Key:        fan.applicationTypeLabelKey,
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
		})
		if err != nil {
			if apperrors.IsNotFoundError(err) {
				return "", nil
			}
			return "", errors.Wrapf(err, "while getting label %q for application with ID %q", fan.applicationTypeLabelKey, objectID)
		}

		applicationType, ok := applicationTypeLabel.Value.(string)
		if !ok {
			return "", errors.Errorf("Missing application type for application %q", objectID)
		}
		return applicationType, nil

	case model.FormationAssignmentTypeRuntime:
		runtimeTypeLabel, err := fan.labelService.GetLabel(ctx, tnt, &model.LabelInput{
			Key:        fan.runtimeTypeLabelKey,
			ObjectID:   objectID,
			ObjectType: model.RuntimeLabelableObject,
		})
		if err != nil {
			if apperrors.IsNotFoundError(err) {
				return "", nil
			}
			return "", errors.Wrapf(err, "while getting label %q for runtime with ID %q", fan.runtimeTypeLabelKey, objectID)
		}

		runtimeType, ok := runtimeTypeLabel.Value.(string)
		if !ok {
			return "", errors.Errorf("Missing runtime type for runtime %q", objectID)
		}
		return runtimeType, nil

	case model.FormationAssignmentTypeRuntimeContext:
		rtmCtx, err := fan.runtimeContextRepo.GetByID(ctx, tnt, objectID)
		if err != nil {
			return "", errors.Wrapf(err, "while fetching runtime context with ID %q", objectID)
		}

		runtimeTypeLabel, err := fan.labelService.GetLabel(ctx, tnt, &model.LabelInput{
			Key:        fan.runtimeTypeLabelKey,
			ObjectID:   rtmCtx.RuntimeID,
			ObjectType: model.RuntimeLabelableObject,
		})
		if err != nil {
			return "", errors.Wrapf(err, "while getting label %q for runtime with ID %q", fan.runtimeTypeLabelKey, objectID)
		}

		runtimeType, ok := runtimeTypeLabel.Value.(string)
		if !ok {
			return "", errors.Errorf("Missing runtime type for runtime %q", rtmCtx.RuntimeID)
		}
		return runtimeType, nil

	default:
		return "", errors.Errorf("unknown formation type %s", objectType)
	}
}

func (fan *formationAssignmentNotificationService) getReverseBySourceAndTarget(ctx context.Context, tnt, formationID, sourceID, targetID string) (*model.FormationAssignment, error) {
	log.C(ctx).Infof("Getting reverse formation assignment for formation ID: %q and source: %q and target: %q", formationID, sourceID, targetID)

	reverseFA, err := fan.formationAssignmentRepo.GetReverseBySourceAndTarget(ctx, tnt, formationID, sourceID, targetID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting reverse formation assignment for formation ID: %q and source: %q and target: %q", formationID, sourceID, targetID)
	}

	return reverseFA, nil
}

// generateApplicationFANotification generates application formation assignment notification based on the reverse(source) type of the formation assignment
func (fan *formationAssignmentNotificationService) generateApplicationFANotification(ctx context.Context, fa *model.FormationAssignment, referencedFormation *model.Formation, customerTenantContext *webhook.CustomerTenantContext, operation model.FormationOperation) (*webhookclient.FormationAssignmentNotificationRequest, error) {
	tenantID := fa.TenantID
	appID := fa.Target

	applicationWithLabels, appTemplateWithLabels, err := fan.webhookDataInputBuilder.PrepareApplicationAndAppTemplateWithLabels(ctx, tenantID, appID)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}

	appTemplateID := ""
	if appTemplateWithLabels != nil {
		appTemplateID = appTemplateWithLabels.ID
	}

	if fa.SourceType == model.FormationAssignmentTypeApplication {
		reverseAppID := fa.Source
		log.C(ctx).Infof("The formation assignment reverse object type is %q and has ID: %q", model.FormationAssignmentTypeApplication, reverseAppID)

		reverseAppWithLabels, reverseAppTemplateWithLabels, err := fan.webhookDataInputBuilder.PrepareApplicationAndAppTemplateWithLabels(ctx, tenantID, reverseAppID)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		reverseFA, err := fan.formationAssignmentRepo.GetReverseBySourceAndTarget(ctx, tenantID, fa.FormationID, fa.Source, fa.Target)
		if err != nil && !apperrors.IsNotFoundError(err) {
			log.C(ctx).Error(err)
			return nil, err
		}

		log.C(ctx).Infof("Preparing join point details for application tenant mapping notification generation")
		details, err := fan.notificationBuilder.PrepareDetailsForApplicationTenantMappingNotificationGeneration(
			operation,
			fa.FormationID,
			referencedFormation.FormationTemplateID,
			reverseAppTemplateWithLabels,
			reverseAppWithLabels,
			appTemplateWithLabels,
			applicationWithLabels,
			convertFormationAssignmentFromModel(fa),
			convertFormationAssignmentFromModel(reverseFA),
			customerTenantContext,
			tenantID,
		)
		if err != nil {
			log.C(ctx).Errorf("while preparing join point details for application tenant mapping notification generation: %v", err)
			return nil, err
		}

		appToAppWebhook, err := GetWebhookForApplication(ctx, fan.webhookRepository, tenantID, appID, appTemplateID, model.WebhookTypeApplicationTenantMapping)
		if err != nil {
			return nil, err
		}
		if appToAppWebhook == nil {
			return nil, nil
		}

		notificationReq, err := fan.notificationBuilder.BuildFormationAssignmentNotificationRequest(ctx, referencedFormation.FormationTemplateID, details, appToAppWebhook)
		if err != nil {
			log.C(ctx).Errorf("while building notification request: %v", err)
			return nil, err
		}

		return notificationReq, nil
	} else if fa.SourceType == model.FormationAssignmentTypeRuntime {
		runtimeID := fa.Source
		log.C(ctx).Infof("The formation assignment reverse object type is %q and has ID: %q", model.FormationAssignmentTypeRuntime, runtimeID)

		runtimeWithLabels, err := fan.webhookDataInputBuilder.PrepareRuntimeWithLabels(ctx, tenantID, runtimeID)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		reverseFA, err := fan.formationAssignmentRepo.GetReverseBySourceAndTarget(ctx, tenantID, fa.FormationID, fa.Source, fa.Target)
		if err != nil && !apperrors.IsNotFoundError(err) {
			log.C(ctx).Error(err)
			return nil, err
		}

		log.C(ctx).Infof("Preparing join point details for configuration change notification generation")
		details, err := fan.notificationBuilder.PrepareDetailsForConfigurationChangeNotificationGeneration(
			operation,
			fa.FormationID,
			referencedFormation.FormationTemplateID,
			appTemplateWithLabels,
			applicationWithLabels,
			runtimeWithLabels,
			nil,
			convertFormationAssignmentFromModel(fa),
			convertFormationAssignmentFromModel(reverseFA),
			model.ApplicationResourceType,
			customerTenantContext,
			tenantID,
		)
		if err != nil {
			log.C(ctx).Errorf("while preparing join point details for configuration change notification generation: %v", err)
			return nil, err
		}

		appWebhook, err := GetWebhookForApplication(ctx, fan.webhookRepository, tenantID, appID, appTemplateID, model.WebhookTypeConfigurationChanged)
		if err != nil {
			return nil, err
		}
		if appWebhook == nil {
			return nil, nil
		}

		notificationReq, err := fan.notificationBuilder.BuildFormationAssignmentNotificationRequest(ctx, referencedFormation.FormationTemplateID, details, appWebhook)
		if err != nil {
			log.C(ctx).Errorf("while building notification request: %v", err)
			return nil, err
		}

		return notificationReq, nil
	} else {
		runtimeCtxID := fa.Source
		log.C(ctx).Infof("The formation assignment reverse object type is %q and has ID: %q", model.FormationAssignmentTypeRuntimeContext, runtimeCtxID)

		runtimeContextWithLabels, err := fan.webhookDataInputBuilder.PrepareRuntimeContextWithLabels(ctx, tenantID, runtimeCtxID)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		runtimeID := runtimeContextWithLabels.RuntimeContext.RuntimeID
		runtimeWithLabels, err := fan.webhookDataInputBuilder.PrepareRuntimeWithLabels(ctx, tenantID, runtimeID)
		if err != nil {
			log.C(ctx).Error(err)
			return nil, err
		}

		reverseFA, err := fan.formationAssignmentRepo.GetReverseBySourceAndTarget(ctx, tenantID, fa.FormationID, fa.Source, fa.Target)
		if err != nil && !apperrors.IsNotFoundError(err) {
			log.C(ctx).Error(err)
			return nil, err
		}

		log.C(ctx).Infof("Preparing join point details for configuration change notification generation")
		details, err := fan.notificationBuilder.PrepareDetailsForConfigurationChangeNotificationGeneration(
			operation,
			fa.FormationID,
			referencedFormation.FormationTemplateID,
			appTemplateWithLabels,
			applicationWithLabels,
			runtimeWithLabels,
			runtimeContextWithLabels,
			convertFormationAssignmentFromModel(fa),
			convertFormationAssignmentFromModel(reverseFA),
			model.ApplicationResourceType,
			customerTenantContext,
			tenantID,
		)
		if err != nil {
			log.C(ctx).Errorf("while preparing join point details for configuration change notification generation: %v", err)
			return nil, err
		}

		appWebhook, err := GetWebhookForApplication(ctx, fan.webhookRepository, tenantID, appID, appTemplateID, model.WebhookTypeConfigurationChanged)
		if err != nil {
			return nil, err
		}
		if appWebhook == nil {
			return nil, nil
		}

		notificationReq, err := fan.notificationBuilder.BuildFormationAssignmentNotificationRequest(ctx, referencedFormation.FormationTemplateID, details, appWebhook)
		if err != nil {
			log.C(ctx).Errorf("while building notification request: %v", err)
			return nil, err
		}

		return notificationReq, nil
	}
}

// generateRuntimeFANotification generates runtime formation assignment notification based on the reverse(source) type of the formation
func (fan *formationAssignmentNotificationService) generateRuntimeFANotification(ctx context.Context, fa *model.FormationAssignment, referencedFormation *model.Formation, customerTenantContext *webhook.CustomerTenantContext, operation model.FormationOperation) (*webhookclient.FormationAssignmentNotificationRequest, error) {
	tenantID := fa.TenantID
	runtimeID := fa.Target

	runtimeWebhook, err := fan.webhookRepository.GetByIDAndWebhookType(ctx, tenantID, runtimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged)
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

	applicationWithLabels, appTemplateWithLabels, err := fan.webhookDataInputBuilder.PrepareApplicationAndAppTemplateWithLabels(ctx, tenantID, appID)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}

	runtimeWithLabels, err := fan.webhookDataInputBuilder.PrepareRuntimeWithLabels(ctx, tenantID, runtimeID)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}

	reverseFA, err := fan.formationAssignmentRepo.GetReverseBySourceAndTarget(ctx, tenantID, fa.FormationID, fa.Source, fa.Target)
	if err != nil && !apperrors.IsNotFoundError(err) {
		log.C(ctx).Error(err)
		return nil, err
	}

	log.C(ctx).Infof("Preparing join point details for configuration change notification generation")
	details, err := fan.notificationBuilder.PrepareDetailsForConfigurationChangeNotificationGeneration(
		operation,
		fa.FormationID,
		referencedFormation.FormationTemplateID,
		appTemplateWithLabels,
		applicationWithLabels,
		runtimeWithLabels,
		nil,
		convertFormationAssignmentFromModel(fa),
		convertFormationAssignmentFromModel(reverseFA),
		model.RuntimeResourceType,
		customerTenantContext,
		tenantID,
	)
	if err != nil {
		log.C(ctx).Errorf("while preparing join point details for configuration change notification generation: %v", err)
		return nil, err
	}

	notificationReq, err := fan.notificationBuilder.BuildFormationAssignmentNotificationRequest(ctx, referencedFormation.FormationTemplateID, details, runtimeWebhook)
	if err != nil {
		log.C(ctx).Errorf("while building notification request: %v", err)
		return nil, err
	}

	return notificationReq, nil
}

// generateRuntimeContextFANotification generates runtime context formation assignment notification based on the reverse(source) type of the formation assignment
func (fan *formationAssignmentNotificationService) generateRuntimeContextFANotification(ctx context.Context, fa *model.FormationAssignment, referencedFormation *model.Formation, customerTenantContext *webhook.CustomerTenantContext, operation model.FormationOperation) (*webhookclient.FormationAssignmentNotificationRequest, error) {
	tenantID := fa.TenantID
	runtimeCtxID := fa.Target

	runtimeContextWithLabels, err := fan.webhookDataInputBuilder.PrepareRuntimeContextWithLabels(ctx, tenantID, runtimeCtxID)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}

	runtimeID := runtimeContextWithLabels.RuntimeContext.RuntimeID
	runtimeWebhook, err := fan.webhookRepository.GetByIDAndWebhookType(ctx, tenantID, runtimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged)
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

	applicationWithLabels, appTemplateWithLabels, err := fan.webhookDataInputBuilder.PrepareApplicationAndAppTemplateWithLabels(ctx, tenantID, appID)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}

	runtimeWithLabels, err := fan.webhookDataInputBuilder.PrepareRuntimeWithLabels(ctx, tenantID, runtimeID)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}

	reverseFA, err := fan.formationAssignmentRepo.GetReverseBySourceAndTarget(ctx, tenantID, fa.FormationID, fa.Source, fa.Target)
	if err != nil && !apperrors.IsNotFoundError(err) {
		log.C(ctx).Error(err)
		return nil, err
	}

	log.C(ctx).Infof("Preparing join point details for configuration change notification generation")
	details, err := fan.notificationBuilder.PrepareDetailsForConfigurationChangeNotificationGeneration(
		operation,
		fa.FormationID,
		referencedFormation.FormationTemplateID,
		appTemplateWithLabels,
		applicationWithLabels,
		runtimeWithLabels,
		runtimeContextWithLabels,
		convertFormationAssignmentFromModel(fa),
		convertFormationAssignmentFromModel(reverseFA),
		model.RuntimeContextResourceType,
		customerTenantContext,
		tenantID,
	)
	if err != nil {
		log.C(ctx).Errorf("while preparing join point details for configuration change notification generation: %v", err)
		return nil, err
	}

	notificationReq, err := fan.notificationBuilder.BuildFormationAssignmentNotificationRequest(ctx, referencedFormation.FormationTemplateID, details, runtimeWebhook)
	if err != nil {
		log.C(ctx).Errorf("while building notification request: %v", err)
		return nil, err
	}

	return notificationReq, nil
}

func convertFormationAssignmentFromModel(formationAssignment *model.FormationAssignment) *webhook.FormationAssignment {
	if formationAssignment == nil {
		return &webhook.FormationAssignment{Value: "\"\""}
	}
	config := string(formationAssignment.Value)
	if config == "" || formationAssignment.State == string(model.CreateErrorAssignmentState) || formationAssignment.State == string(model.DeleteErrorAssignmentState) {
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

func (fan *formationAssignmentNotificationService) extractCustomerTenantContext(ctx context.Context, internalTenantID string) (*webhook.CustomerTenantContext, error) {
	tenantObject, err := fan.tenantRepository.Get(ctx, internalTenantID)
	if err != nil {
		return nil, err
	}

	customerID, err := fan.tenantRepository.GetCustomerIDParentRecursively(ctx, internalTenantID)
	if err != nil {
		return nil, err
	}

	var accountID *string
	var path *string
	if tenantObject.Type == tenant.Account {
		accountID = &tenantObject.ExternalTenant
	} else if tenantObject.Type == tenant.ResourceGroup {
		path = &tenantObject.ExternalTenant
	}

	return &webhook.CustomerTenantContext{
		CustomerID: customerID,
		AccountID:  accountID,
		Path:       path,
	}, nil
}

// GetWebhookForApplication gets webhook of type webhookType for the application with ID appID
// If the application has webhook of type webhookType it is returned
// If the application does not have a webhook of type webhookType, but its application template has one it is returned
// If both application and application template does not have a webhook of type webhookType, no webhook is returned
func GetWebhookForApplication(ctx context.Context, webhookRepo webhookRepository, tenant, appID, appTemplateID string, webhookType model.WebhookType) (*model.Webhook, error) {
	webhook, err := webhookRepo.GetByIDAndWebhookType(ctx, tenant, appID, model.ApplicationWebhookReference, webhookType)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return nil, errors.Wrapf(err, "while listing %s webhooks for application %s", webhookType, appID)
		}

		log.C(ctx).Infof("There is no %q webhook attached to application with ID: %q. Looking for %q webhook attached to application template.", webhookType, appID, webhookType)

		if appTemplateID == "" {
			log.C(ctx).Infof("There is no application template for application with ID: %q. No notifications will be generated.", appID)
			return nil, nil
		}

		webhook, err = webhookRepo.GetByIDAndWebhookType(ctx, tenant, appTemplateID, model.ApplicationTemplateWebhookReference, webhookType)
		if err != nil {
			if !apperrors.IsNotFoundError(err) {
				return nil, errors.Wrapf(err, "while listing %q webhooks for application template with ID: %q on behalf of application with ID: %q", webhookType, appTemplateID, appID)
			}

			log.C(ctx).Infof("There is no %q webhook attached to application template with ID: %q. No notifications will be generated.", webhookType, appTemplateID)
			return nil, nil
		}
	}

	return webhook, nil
}
