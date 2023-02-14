package formation

import (
	"context"
	"fmt"

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
	GenerateNotificationsAboutRuntimeAndRuntimeContextForTheApplicationThatIsAssigned(ctx context.Context, tenant string, appID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.NotificationRequest, error)
	GenerateNotificationsForRuntimeAboutTheApplicationThatIsAssigned(ctx context.Context, tenant string, appID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.NotificationRequest, error)
	GenerateNotificationsForApplicationsAboutTheApplicationThatIsAssigned(ctx context.Context, tenant string, appID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.NotificationRequest, error)
	GenerateNotificationsForApplicationsAboutTheRuntimeContextThatIsAssigned(ctx context.Context, tenant, runtimeCtxID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.NotificationRequest, error)
	GenerateNotificationsForApplicationsAboutTheRuntimeThatIsAssigned(ctx context.Context, tenant, runtimeID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.NotificationRequest, error)
	GenerateNotificationsAboutApplicationsForTheRuntimeContextThatIsAssigned(ctx context.Context, tenant, runtimeCtxID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.NotificationRequest, error)
	GenerateNotificationsAboutApplicationsForTheRuntimeThatIsAssigned(ctx context.Context, tenant, runtimeID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.NotificationRequest, error)
}

var emptyFormationAssignment = &webhookdir.FormationAssignment{Value: "\"\""}

type notificationsService struct {
	tenantRepository       tenantRepository
	webhookClient          webhookClient
	notificationsGenerator notificationsGenerator
}

// NewNotificationService creates notifications service for formation assignment and unassignment
func NewNotificationService(
	tenantRepository tenantRepository,
	webhookClient webhookClient,
	notificationsGenerator notificationsGenerator,
) *notificationsService {
	return &notificationsService{
		tenantRepository:       tenantRepository,
		webhookClient:          webhookClient,
		notificationsGenerator: notificationsGenerator,
	}
}

// GenerateNotifications generates notifications for all listening resources about the execution of `operation` for formation `formation` and object `objectID` of type `objectType`
func (ns *notificationsService) GenerateNotifications(ctx context.Context, tenant, objectID string, formation *model.Formation, operation model.FormationOperation, objectType graphql.FormationObjectType) ([]*webhookclient.NotificationRequest, error) {
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

// SendNotification sends the provided notification
func (ns *notificationsService) SendNotification(ctx context.Context, notification *webhookclient.NotificationRequest) (*webhookdir.Response, error) {
	if notification == nil {
		return nil, nil
	}
	resp, err := ns.webhookClient.Do(ctx, notification)
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
