package formation

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	webhookdir "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=tenantRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantRepository interface {
	Get(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	GetCustomerIDParentRecursively(ctx context.Context, tenant string) (string, error)
}

//go:generate mockery --exported --name=webhookClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type webhookClient interface {
	Do(ctx context.Context, request webhookclient.WebhookRequest) (*webhookdir.Response, error)
}

//go:generate mockery --exported --name=notificationsGenerator --output=automock --outpkg=automock --case=underscore --disable-version-string
type notificationsGenerator interface {
	GenerateNotificationsAboutRuntimeAndRuntimeContextForTheApplicationThatIsAssigned(ctx context.Context, tenant string, appID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.FormationAssignmentNotificationRequest, error)
	GenerateNotificationsForRuntimeAboutTheApplicationThatIsAssigned(ctx context.Context, tenant string, appID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.FormationAssignmentNotificationRequest, error)
	GenerateNotificationsForApplicationsAboutTheApplicationThatIsAssigned(ctx context.Context, tenant string, appID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.FormationAssignmentNotificationRequest, error)
	GenerateNotificationsForApplicationsAboutTheRuntimeContextThatIsAssigned(ctx context.Context, tenant, runtimeCtxID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.FormationAssignmentNotificationRequest, error)
	GenerateNotificationsForApplicationsAboutTheRuntimeThatIsAssigned(ctx context.Context, tenant, runtimeID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.FormationAssignmentNotificationRequest, error)
	GenerateNotificationsAboutApplicationsForTheRuntimeContextThatIsAssigned(ctx context.Context, tenant, runtimeCtxID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.FormationAssignmentNotificationRequest, error)
	GenerateNotificationsAboutApplicationsForTheRuntimeThatIsAssigned(ctx context.Context, tenant, runtimeID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.FormationAssignmentNotificationRequest, error)
}

var emptyFormationAssignment = &webhookdir.FormationAssignment{Value: "\"\""}

type notificationsService struct {
	tenantRepository       tenantRepository
	webhookClient          webhookClient
	webhookConverter       webhookConverter
	notificationsGenerator notificationsGenerator
}

// NewNotificationService creates notifications service for formation assignment and unassignment
func NewNotificationService(
	tenantRepository tenantRepository,
	webhookClient webhookClient,
	webhookConverter webhookConverter,
	notificationsGenerator notificationsGenerator,
) *notificationsService {
	return &notificationsService{
		tenantRepository:       tenantRepository,
		webhookClient:          webhookClient,
		webhookConverter:       webhookConverter,
		notificationsGenerator: notificationsGenerator,
	}
}

// GenerateFormationAssignmentNotifications generates notifications for all listening resources about the execution of `operation` for formation `formation` and object `objectID` of type `objectType`
func (ns *notificationsService) GenerateFormationAssignmentNotifications(ctx context.Context, tenant, objectID string, formation *model.Formation, operation model.FormationOperation, objectType graphql.FormationObjectType) ([]*webhookclient.FormationAssignmentNotificationRequest, error) {
	customerTenantContext, err := ns.extractCustomerTenantContext(ctx, formation.TenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while extracting customer tenant context for tenant with internal ID %s", formation.TenantID)
	}
	switch objectType {
	case graphql.FormationObjectTypeApplication:
		appNotifications, err := ns.notificationsGenerator.GenerateNotificationsAboutRuntimeAndRuntimeContextForTheApplicationThatIsAssigned(ctx, tenant, objectID, formation, operation, customerTenantContext)
		if err != nil {
			return nil, err
		}

		rtNotifications, err := ns.notificationsGenerator.GenerateNotificationsForRuntimeAboutTheApplicationThatIsAssigned(ctx, tenant, objectID, formation, operation, customerTenantContext)
		if err != nil {
			return nil, err
		}

		appToAppNotifications, err := ns.notificationsGenerator.GenerateNotificationsForApplicationsAboutTheApplicationThatIsAssigned(ctx, tenant, objectID, formation, operation, customerTenantContext)
		if err != nil {
			return nil, err
		}
		appNotifications = append(appNotifications, rtNotifications...)

		return append(appNotifications, appToAppNotifications...), nil
	case graphql.FormationObjectTypeRuntime:
		appNotifications, err := ns.notificationsGenerator.GenerateNotificationsForApplicationsAboutTheRuntimeThatIsAssigned(ctx, tenant, objectID, formation, operation, customerTenantContext)
		if err != nil {
			return nil, err
		}

		rtNotifications, err := ns.notificationsGenerator.GenerateNotificationsAboutApplicationsForTheRuntimeThatIsAssigned(ctx, tenant, objectID, formation, operation, customerTenantContext)
		if err != nil {
			return nil, err
		}

		return append(appNotifications, rtNotifications...), nil
	case graphql.FormationObjectTypeRuntimeContext:
		appNotifications, err := ns.notificationsGenerator.GenerateNotificationsForApplicationsAboutTheRuntimeContextThatIsAssigned(ctx, tenant, objectID, formation, operation, customerTenantContext)
		if err != nil {
			return nil, err
		}

		rtCtxNotifications, err := ns.notificationsGenerator.GenerateNotificationsAboutApplicationsForTheRuntimeContextThatIsAssigned(ctx, tenant, objectID, formation, operation, customerTenantContext)
		if err != nil {
			return nil, err
		}

		return append(appNotifications, rtCtxNotifications...), nil
	default:
		return nil, fmt.Errorf("unknown formation type %s", objectType)
	}
}

func (ns *notificationsService) GenerateFormationNotifications(ctx context.Context, formationTemplateWebhooks []*model.Webhook, tenantID string, formation *model.Formation, formationTemplateID string, formationOperation model.FormationOperation) ([]*webhookclient.FormationNotificationRequest, error) {
	if len(formationTemplateWebhooks) == 0 {
		log.C(ctx).Infof("Formation template with ID: %q does not have any webhooks", formationTemplateID)
		return nil, nil
	}

	log.C(ctx).Infof("There are %d formation template(s) listening for formation lifecycle notifications", len(formationTemplateWebhooks))

	customerTenantContext, err := ns.extractCustomerTenantContext(ctx, tenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while extracting customer tenant context for tenant with internal ID %s", tenantID)
	}

	formationTemplateInput := buildFormationLifecycleInput(formationOperation, formation, customerTenantContext)

	requests := make([]*webhookclient.FormationNotificationRequest, 0, len(formationTemplateWebhooks))
	for _, webhook := range formationTemplateWebhooks {
		gqlWebhook, err := ns.webhookConverter.ToGraphQL(webhook)
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

	return requests, nil
}

func (ns *notificationsService) SendNotification(ctx context.Context, webhookNotificationReq webhookclient.WebhookRequest) (*webhookdir.Response, error) {
	resp, err := ns.webhookClient.Do(ctx, webhookNotificationReq)
	if err != nil && resp != nil && resp.Error != nil && *resp.Error != "" {
		return resp, nil
	}

	return resp, err
}

func (ns *notificationsService) extractCustomerTenantContext(ctx context.Context, internalTenantID string) (*webhookdir.CustomerTenantContext, error) {
	tenantObject, err := ns.tenantRepository.Get(ctx, internalTenantID)
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

	customerID, err := ns.tenantRepository.GetCustomerIDParentRecursively(ctx, internalTenantID)
	if err != nil {
		return nil, err
	}

	return &webhookdir.CustomerTenantContext{
		CustomerID: customerID,
		AccountID:  accountID,
		Path:       path,
	}, nil
}
