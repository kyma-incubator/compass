package formation

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	webhookdir "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
)

//go:generate mockery --exported --name=notificationBuilder --output=automock --outpkg=automock --case=underscore --disable-version-string
type notificationBuilder interface {
	BuildFormationNotificationRequests(ctx context.Context, joinPointDetails *formationconstraint.GenerateFormationNotificationOperationDetails, formation *model.Formation, formationTemplateWebhooks []*model.Webhook) ([]*webhookclient.FormationNotificationRequest, error)
}

// NotificationsGenerator is responsible for generation of notification requests
type NotificationsGenerator struct {
	notificationBuilder notificationBuilder
}

// NewNotificationsGenerator returns an instance of NotificationsGenerator
func NewNotificationsGenerator(notificationBuilder notificationBuilder) *NotificationsGenerator {
	return &NotificationsGenerator{
		notificationBuilder: notificationBuilder,
	}
}

// GenerateFormationLifecycleNotifications generates formation notifications for the provided webhooks
func (ns *NotificationsGenerator) GenerateFormationLifecycleNotifications(ctx context.Context, formationTemplateWebhooks []*model.Webhook, tenantID string, formation *model.Formation, formationTemplateName, formationTemplateID string, formationOperation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.FormationNotificationRequest, error) {
	details := &formationconstraint.GenerateFormationNotificationOperationDetails{
		Operation:             formationOperation,
		FormationID:           formation.ID,
		FormationName:         formation.Name,
		FormationType:         formationTemplateName,
		FormationTemplateID:   formationTemplateID,
		TenantID:              tenantID,
		CustomerTenantContext: customerTenantContext,
	}

	reqs, err := ns.notificationBuilder.BuildFormationNotificationRequests(ctx, details, formation, formationTemplateWebhooks)
	if err != nil {
		log.C(ctx).Errorf("Failed to build formation notification requests due to: %v", err)
	}

	return reqs, nil
}
