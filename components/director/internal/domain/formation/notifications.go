package formation

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
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
	GenerateFormationLifecycleNotifications(ctx context.Context, formationTemplateWebhooks []*model.Webhook, tenantID string, formation *model.Formation, formationTemplateName, formationTemplateID string, formationOperation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.FormationNotificationRequest, error)
}

var emptyFormationAssignment = &webhookdir.FormationAssignment{Value: "\"\""}

type notificationsService struct {
	tenantRepository            tenantRepository
	webhookClient               webhookClient
	notificationsGenerator      notificationsGenerator
	constraintEngine            constraintEngine
	webhookConverter            webhookConverter
	formationTemplateRepository FormationTemplateRepository
}

// NewNotificationService creates notifications service for formation assignment and unassignment
func NewNotificationService(
	tenantRepository tenantRepository,
	webhookClient webhookClient,
	notificationsGenerator notificationsGenerator,
	constraintEngine constraintEngine,
	webhookConverter webhookConverter,
	formationTemplateRepository FormationTemplateRepository,
) *notificationsService {
	return &notificationsService{
		tenantRepository:            tenantRepository,
		webhookClient:               webhookClient,
		notificationsGenerator:      notificationsGenerator,
		constraintEngine:            constraintEngine,
		webhookConverter:            webhookConverter,
		formationTemplateRepository: formationTemplateRepository,
	}
}

// GenerateFormationAssignmentNotifications generates formation assignment notifications for all listening resources about the execution of `operation` for formation `formation` and object `objectID` of type `objectType`
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

// GenerateFormationNotifications generates formation notifications for the provided webhooks
func (ns *notificationsService) GenerateFormationNotifications(ctx context.Context, formationTemplateWebhooks []*model.Webhook, tenantID string, formation *model.Formation, formationTemplateName, formationTemplateID string, formationOperation model.FormationOperation) ([]*webhookclient.FormationNotificationRequest, error) {
	customerTenantContext, err := ns.extractCustomerTenantContext(ctx, formation.TenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while extracting customer tenant context for tenant with internal ID %s", formation.TenantID)
	}

	return ns.notificationsGenerator.GenerateFormationLifecycleNotifications(ctx, formationTemplateWebhooks, tenantID, formation, formationTemplateName, formationTemplateID, formationOperation, customerTenantContext)
}

func (ns *notificationsService) SendNotification(ctx context.Context, webhookNotificationReq webhookclient.WebhookExtRequest) (*webhookdir.Response, error) {
	joinPointDetails, err := ns.prepareDetailsForSendNotification(webhookNotificationReq)
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing details for send notification")
	}
	joinPointDetails.Location = formationconstraint.PreSendNotification
	if err = ns.constraintEngine.EnforceConstraints(ctx, formationconstraint.PreSendNotification, joinPointDetails, webhookNotificationReq.GetFormation().FormationTemplateID); err != nil {
		return nil, errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.SendNotificationOperation, model.PreOperation)
	}

	resp, err := ns.webhookClient.Do(ctx, webhookNotificationReq)
	if err != nil && resp != nil && resp.Error != nil && *resp.Error != "" {
		return resp, nil
	} else if err != nil {
		return resp, err
	}

	joinPointDetails.Location = formationconstraint.PostSendNotification
	if err = ns.constraintEngine.EnforceConstraints(ctx, formationconstraint.PostSendNotification, joinPointDetails, webhookNotificationReq.GetFormation().FormationTemplateID); err != nil {
		return nil, errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.SendNotificationOperation, model.PostOperation)
	}

	return resp, err
}

func (ns *notificationsService) PrepareDetailsForNotificationStatusReturned(ctx context.Context, formation *model.Formation, operation model.FormationOperation) (*formationconstraint.NotificationStatusReturnedOperationDetails, error) {
	template, err := ns.formationTemplateRepository.Get(ctx, formation.FormationTemplateID)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting formation template by ID: %q: %v", formation.FormationTemplateID, err)
		return nil, errors.Wrapf(err, "An error occurred while getting formation template by ID: %q", formation.FormationTemplateID)
	}

	return &formationconstraint.NotificationStatusReturnedOperationDetails{
		ResourceType:      model.FormationResourceType,
		ResourceSubtype:   template.Name,
		Operation:         operation,
		Formation:         formation,
		FormationTemplate: template,
	}, nil
}

func (ns *notificationsService) prepareDetailsForSendNotification(webhookNotificationReq webhookclient.WebhookExtRequest) (*formationconstraint.SendNotificationOperationDetails, error) {
	webhookGql := webhookNotificationReq.GetWebhook()
	webhookModel, err := ns.webhookConverter.ToModel(&webhookGql)
	if err != nil {
		return nil, errors.Wrap(err, "while converting webhook to model")
	}

	joinPointDetails := &formationconstraint.SendNotificationOperationDetails{
		ResourceType:               webhookNotificationReq.GetObjectType(),
		ResourceSubtype:            webhookNotificationReq.GetObjectSubtype(),
		Operation:                  webhookNotificationReq.GetOperation(),
		Webhook:                    webhookModel,
		CorrelationID:              webhookNotificationReq.GetCorrelationID(),
		TemplateInput:              webhookNotificationReq.GetObject(),
		FormationAssignment:        webhookNotificationReq.GetFormationAssignment(),
		ReverseFormationAssignment: webhookNotificationReq.GetReverseFormationAssignment(),
		Formation:                  webhookNotificationReq.GetFormation(),
	}

	return joinPointDetails, nil
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
