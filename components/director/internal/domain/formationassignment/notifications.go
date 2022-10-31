package formationassignment

import (
	"context"

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
	webhookClient                 webhookClient
}

// NewFormationAssignmentNotificationService creates formation assignment notifications service
func NewFormationAssignmentNotificationService(applicationRepo applicationRepository, applicationTemplateRepository applicationTemplateRepository, runtimeRepo runtimeRepository, runtimeContextRepo runtimeContextRepository, labelRepository labelRepository, webhookRepository webhookRepository, webhookConverter webhookConverter, webhookClient webhookClient) *formationAssignmentNotificationService {
	return &formationAssignmentNotificationService{
		applicationRepository:         applicationRepo,
		applicationTemplateRepository: applicationTemplateRepository,
		runtimeRepo:                   runtimeRepo,
		runtimeContextRepo:            runtimeContextRepo,
		labelRepository:               labelRepository,
		webhookRepository:             webhookRepository,
		webhookClient:                 webhookClient,
		webhookConverter:              webhookConverter,
	}
}

func (fan *formationAssignmentNotificationService) GenerateNotification(ctx context.Context, tenant, objectID, formationName string, faType, faReverseType model.FormationAssignmentType) (*webhookclient.NotificationRequest, error) {
	switch faType {
	case model.FormationAssignmentTypeApplication:
		return fan.generateReverseFormationAssignmentNotification(faReverseType)
	case model.FormationAssignmentTypeRuntime:
		return fan.generateReverseFormationAssignmentNotification(faReverseType)
	case model.FormationAssignmentTypeRuntimeContext:
		return fan.generateReverseFormationAssignmentNotification(faReverseType)
	default:
		return nil, errors.Errorf("Unknown formation assignment type: %q", faType)
	}
}

func (fan *formationAssignmentNotificationService) generateReverseFormationAssignmentNotification(faReverseType model.FormationAssignmentType) (*webhookclient.NotificationRequest, error) {
	switch faReverseType {
	case model.FormationAssignmentTypeApplication:
		return fan.generateApplicationFormationAssignmentNotification()
	case model.FormationAssignmentTypeRuntime:
		return fan.generateRuntimeFormationAssignmentNotification()
	case model.FormationAssignmentTypeRuntimeContext:
		return fan.generateRuntimeContextFormationAssignmentNotification()
	default:
		return nil, errors.Errorf("Unknown formation assignment type: %q", faReverseType)
	}
}

func (fan *formationAssignmentNotificationService) generateApplicationFormationAssignmentNotification() (*webhookclient.NotificationRequest, error) {
	panic("todo::: implement")
}

func (fan *formationAssignmentNotificationService) generateRuntimeFormationAssignmentNotification() (*webhookclient.NotificationRequest, error) {
	panic("todo::: implement")
}

func (fan *formationAssignmentNotificationService) generateRuntimeContextFormationAssignmentNotification() (*webhookclient.NotificationRequest, error) {
	panic("todo::: implement")
}
