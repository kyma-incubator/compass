package formation

import (
	"context"
	"fmt"

	databuilder "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	webhookdir "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=applicationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.Application, error)
	ListByScenariosNoPaging(ctx context.Context, tenant string, scenarios []string) ([]*model.Application, error)
	ListByScenariosAndIDs(ctx context.Context, tenant string, scenarios []string, ids []string) ([]*model.Application, error)
}

//go:generate mockery --exported --name=applicationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationTemplateRepository interface {
	Get(ctx context.Context, id string) (*model.ApplicationTemplate, error)
	ListByIDs(ctx context.Context, ids []string) ([]*model.ApplicationTemplate, error)
}

//go:generate mockery --exported --name=webhookRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type webhookRepository interface {
	ListByReferenceObjectTypeAndWebhookType(ctx context.Context, tenant string, whType model.WebhookType, objType model.WebhookReferenceObjectType) ([]*model.Webhook, error)
	GetByIDAndWebhookType(ctx context.Context, tenant, objectID string, objectType model.WebhookReferenceObjectType, webhookType model.WebhookType) (*model.Webhook, error)
}

//go:generate mockery --exported --name=tenantRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantRepository interface {
	Get(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	GetCustomerIDParentRecursively(ctx context.Context, tenant string) (string, error)
}

//go:generate mockery --exported --name=webhookConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type webhookConverter interface {
	ToGraphQL(in *model.Webhook) (*graphql.Webhook, error)
}

//go:generate mockery --exported --name=webhookClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type webhookClient interface {
	Do(ctx context.Context, request webhookclient.WebhookRequest) (*webhookdir.Response, error)
}

var emptyFormationAssignment = &webhookdir.FormationAssignment{Value: "\"\""}

type notificationsService struct {
	applicationRepository         applicationRepository
	applicationTemplateRepository applicationTemplateRepository
	runtimeRepo                   runtimeRepository
	runtimeContextRepo            runtimeContextRepository
	labelRepository               labelRepository
	webhookRepository             webhookRepository
	tenantRepository              tenantRepository
	webhookConverter              webhookConverter
	webhookClient                 webhookClient
	webhookDataInputBuilder       databuilder.DataInputBuilder
}

// NewNotificationService creates notifications service for formation assignment and unassignment
func NewNotificationService(
	applicationRepository applicationRepository,
	applicationTemplateRepository applicationTemplateRepository,
	runtimeRepo runtimeRepository,
	runtimeContextRepo runtimeContextRepository,
	labelRepository labelRepository,
	webhookRepository webhookRepository,
	tenantRepository tenantRepository,
	webhookConverter webhookConverter,
	webhookClient webhookClient,
	webhookDataInputBuilder databuilder.DataInputBuilder,
) *notificationsService {
	return &notificationsService{
		applicationRepository:         applicationRepository,
		applicationTemplateRepository: applicationTemplateRepository,
		runtimeRepo:                   runtimeRepo,
		runtimeContextRepo:            runtimeContextRepo,
		labelRepository:               labelRepository,
		webhookRepository:             webhookRepository,
		tenantRepository:              tenantRepository,
		webhookClient:                 webhookClient,
		webhookConverter:              webhookConverter,
		webhookDataInputBuilder:       webhookDataInputBuilder,
	}
}

func (ns *notificationsService) GenerateNotifications(ctx context.Context, tenant, objectID string, formation *model.Formation, operation model.FormationOperation, objectType graphql.FormationObjectType) ([]*webhookclient.NotificationRequest, error) {
	customerTenantContext, err := ns.extractCustomerTenantContext(ctx, formation.TenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while extracting customer tenant context for tenant with internal ID %s", formation.TenantID)
	}
	switch objectType {
	case graphql.FormationObjectTypeApplication:
		appNotifications, err := ns.generateNotificationsAboutRuntimeAndRuntimeContextForTheApplicationThatIsAssigned(ctx, tenant, objectID, formation, operation, customerTenantContext)
		if err != nil {
			return nil, err
		}

		rtNotifications, err := ns.generateNotificationsForRuntimeAboutTheApplicationThatIsAssigned(ctx, tenant, objectID, formation, operation, customerTenantContext)
		if err != nil {
			return nil, err
		}

		appToAppNotifications, err := ns.generateNotificationsForApplicationsAboutTheApplicationThatIsAssigned(ctx, tenant, objectID, formation, operation, customerTenantContext)
		if err != nil {
			return nil, err
		}
		appNotifications = append(appNotifications, rtNotifications...)
		return append(appNotifications, appToAppNotifications...), nil
	case graphql.FormationObjectTypeRuntime:
		appNotifications, err := ns.generateNotificationsForApplicationsAboutTheRuntimeThatIsAssigned(ctx, tenant, objectID, formation, operation, customerTenantContext)
		if err != nil {
			return nil, err
		}
		rtNotifications, err := ns.generateNotificationsAboutApplicationsForTheRuntimeThatIsAssigned(ctx, tenant, objectID, formation, operation, customerTenantContext)
		if err != nil {
			return nil, err
		}
		return append(appNotifications, rtNotifications...), nil
	case graphql.FormationObjectTypeRuntimeContext:
		appNotifications, err := ns.generateNotificationsForApplicationsAboutTheRuntimeContextThatIsAssigned(ctx, tenant, objectID, formation, operation, customerTenantContext)
		if err != nil {
			return nil, err
		}
		rtCtxNotifications, err := ns.generateNotificationsAboutApplicationsForTheRuntimeContextThatIsAssigned(ctx, tenant, objectID, formation, operation, customerTenantContext)
		if err != nil {
			return nil, err
		}
		return append(appNotifications, rtCtxNotifications...), nil
	default:
		return nil, fmt.Errorf("unknown formation type %s", objectType)
	}
}

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

func (ns *notificationsService) generateNotificationsAboutRuntimeAndRuntimeContextForTheApplicationThatIsAssigned(ctx context.Context, tenant string, appID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.NotificationRequest, error) {
	log.C(ctx).Infof("Generating %s notifications about runtimes and runtime contexts in the same formation for application %s", operation, appID)
	applicationWithLabels, appTemplateWithLabels, err := ns.webhookDataInputBuilder.PrepareApplicationAndAppTemplateWithLabels(ctx, tenant, appID)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing application and application template with labels")
	}

	webhook, err := ns.webhookRepository.GetByIDAndWebhookType(ctx, tenant, appID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			log.C(ctx).Infof("There is no configuration changed webhook for application %s. There are no notifications to be generated.", appID)
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while listing configuration changed webhooks for application %s", appID)
	}

	runtimesInFormation, err := ns.runtimeRepo.ListByScenarios(ctx, tenant, []string{formation.Name})
	if err != nil {
		return nil, errors.Wrapf(err, "while listing runtimes in scenario %s", formation.Name)
	}

	runtimeContextsInFormation, err := ns.runtimeContextRepo.ListByScenarios(ctx, tenant, []string{formation.Name})
	if err != nil {
		return nil, errors.Wrapf(err, "while listing runtime contexts in scenario %s", formation.Name)
	}

	runtimeContextsIDs := make([]string, 0, len(runtimeContextsInFormation))
	parentRuntimeIDs := make([]string, 0, len(runtimeContextsInFormation))
	for _, rtCtx := range runtimeContextsInFormation {
		runtimeContextsIDs = append(runtimeContextsIDs, rtCtx.ID)
		parentRuntimeIDs = append(parentRuntimeIDs, rtCtx.RuntimeID)
	}

	// the parent runtime of the runtime context may not be in the formation - that's why we list them separately
	parentRuntimesOfRuntimeContextsInFormation, err := ns.runtimeRepo.ListByIDs(ctx, tenant, parentRuntimeIDs)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing parent runtimes of runtime contexts in scenario %s", formation.Name)
	}

	runtimesIDs := make(map[string]bool, len(runtimesInFormation)+len(parentRuntimesOfRuntimeContextsInFormation))
	for _, rt := range runtimesInFormation {
		runtimesIDs[rt.ID] = true
	}

	for _, rt := range parentRuntimesOfRuntimeContextsInFormation {
		runtimesIDs[rt.ID] = true
	}

	runtimesLabels, err := ns.labelRepository.ListForObjectIDs(ctx, tenant, model.RuntimeLabelableObject, setToSlice(runtimesIDs))
	if err != nil {
		return nil, errors.Wrap(err, "while listing runtime labels")
	}

	runtimesMapping := make(map[string]*webhookdir.RuntimeWithLabels, len(runtimesLabels))
	for _, rt := range runtimesInFormation {
		runtimesMapping[rt.ID] = &webhookdir.RuntimeWithLabels{
			Runtime: rt,
			Labels:  runtimesLabels[rt.ID],
		}
	}

	for _, rt := range parentRuntimesOfRuntimeContextsInFormation {
		runtimesMapping[rt.ID] = &webhookdir.RuntimeWithLabels{
			Runtime: rt,
			Labels:  runtimesLabels[rt.ID],
		}
	}

	runtimeContextsLabels, err := ns.labelRepository.ListForObjectIDs(ctx, tenant, model.RuntimeContextLabelableObject, runtimeContextsIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing labels for runtime contexts")
	}

	runtimesToRuntimeContextsMapping := make(map[string]*webhookdir.RuntimeContextWithLabels, len(runtimeContextsInFormation))
	for _, rtCtx := range runtimeContextsInFormation {
		runtimesToRuntimeContextsMapping[rtCtx.RuntimeID] = &webhookdir.RuntimeContextWithLabels{
			RuntimeContext: rtCtx,
			Labels:         runtimeContextsLabels[rtCtx.ID],
		}
	}

	requests := make([]*webhookclient.NotificationRequest, 0, len(runtimesMapping))
	for rtID := range runtimesMapping {
		rtCtx := runtimesToRuntimeContextsMapping[rtID]
		if rtCtx == nil {
			log.C(ctx).Infof("There is no runtime context for runtime %s in scenario %s. Will proceed without runtime context in the input for webhook %s", rtID, formation.Name, webhook.ID)
		}
		runtime := runtimesMapping[rtID]
		if appTemplateWithLabels == nil {
			log.C(ctx).Infof("Application %s has no application template. Will proceed without application template in the input for webhook %s", appID, webhook.ID)
		}
		input := &webhookdir.FormationConfigurationChangeInput{
			Operation:             operation,
			FormationID:           formation.ID,
			ApplicationTemplate:   appTemplateWithLabels,
			Application:           applicationWithLabels,
			Runtime:               runtime,
			RuntimeContext:        rtCtx,
			CustomerTenantContext: customerTenantContext,
			Assignment:            emptyFormationAssignment,
			ReverseAssignment:     emptyFormationAssignment,
		}
		req, err := ns.createWebhookRequest(ctx, webhook, input)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

func (ns *notificationsService) generateNotificationsForRuntimeAboutTheApplicationThatIsAssigned(ctx context.Context, tenant string, appID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.NotificationRequest, error) {
	log.C(ctx).Infof("Generating %s notifications about application %s for all listening runtimes in the same formation", operation, appID)
	applicationWithLabels, appTemplateWithLabels, err := ns.webhookDataInputBuilder.PrepareApplicationAndAppTemplateWithLabels(ctx, tenant, appID)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing application and application template with labels")
	}

	webhooks, err := ns.webhookRepository.ListByReferenceObjectTypeAndWebhookType(ctx, tenant, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference)
	if err != nil {
		return nil, errors.Wrap(err, "when listing configuration changed webhooks for runtimes")
	}

	listeningRuntimeIDs := make([]string, 0, len(webhooks))
	for _, wh := range webhooks {
		listeningRuntimeIDs = append(listeningRuntimeIDs, wh.ObjectID)
	}

	if len(listeningRuntimeIDs) == 0 {
		log.C(ctx).Infof("There are no runtimes listening for formation notifications in tenant %s", tenant)
		return nil, nil
	}

	log.C(ctx).Infof("There are %d runtimes listening for formation notifications in tenant %s", len(listeningRuntimeIDs), tenant)

	listeningRuntimes, err := ns.runtimeRepo.ListByIDs(ctx, tenant, listeningRuntimeIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing runtimes")
	}

	listeningRuntimesLabels, err := ns.labelRepository.ListForObjectIDs(ctx, tenant, model.RuntimeLabelableObject, listeningRuntimeIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing runtime labels")
	}

	listeningRuntimesMapping := make(map[string]*webhookdir.RuntimeWithLabels, len(listeningRuntimes))
	for i, rt := range listeningRuntimes {
		listeningRuntimesMapping[rt.ID] = &webhookdir.RuntimeWithLabels{
			Runtime: listeningRuntimes[i],
			Labels:  listeningRuntimesLabels[rt.ID],
		}
	}

	listeningRuntimesInScenario, err := ns.runtimeRepo.ListByScenariosAndIDs(ctx, tenant, []string{formation.Name}, listeningRuntimeIDs)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing runtimes in scenario %s", formation.Name)
	}

	log.C(ctx).Infof("There are %d out of %d runtimes listening for formation notifications in tenant %s that are in scenario %s", len(listeningRuntimesInScenario), len(listeningRuntimeIDs), tenant, formation.Name)

	runtimeContextsInScenarioForListeningRuntimes, err := ns.runtimeContextRepo.ListByScenariosAndRuntimeIDs(ctx, tenant, []string{formation.Name}, listeningRuntimeIDs)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing runtime contexts in scenario %s", formation.Name)
	}

	log.C(ctx).Infof("There are %d runtime contexts in tenant %s that are in scenario %s and are for any of the listening runtimes", len(runtimeContextsInScenarioForListeningRuntimes), tenant, formation.Name)

	runtimeContextsInScenarioForListeningRuntimesIDs := make([]string, 0, len(runtimeContextsInScenarioForListeningRuntimes))
	for _, rtCtx := range runtimeContextsInScenarioForListeningRuntimes {
		runtimeContextsInScenarioForListeningRuntimesIDs = append(runtimeContextsInScenarioForListeningRuntimesIDs, rtCtx.ID)
	}

	runtimeContextsLabels, err := ns.labelRepository.ListForObjectIDs(ctx, tenant, model.RuntimeContextLabelableObject, runtimeContextsInScenarioForListeningRuntimesIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing labels for runtime contexts")
	}

	runtimeIDsToBeNotified := make(map[string]bool, len(listeningRuntimesInScenario)+len(runtimeContextsInScenarioForListeningRuntimes))
	runtimeContextsInScenarioForListeningRuntimesMapping := make(map[string]*webhookdir.RuntimeContextWithLabels, len(runtimeContextsInScenarioForListeningRuntimes))
	for _, rt := range listeningRuntimesInScenario {
		runtimeIDsToBeNotified[rt.ID] = true
	}
	for i, rtCtx := range runtimeContextsInScenarioForListeningRuntimes {
		runtimeIDsToBeNotified[rtCtx.RuntimeID] = true
		runtimeContextsInScenarioForListeningRuntimesMapping[rtCtx.RuntimeID] = &webhookdir.RuntimeContextWithLabels{
			RuntimeContext: runtimeContextsInScenarioForListeningRuntimes[i],
			Labels:         runtimeContextsLabels[rtCtx.ID],
		}
	}

	webhooksToCall := make(map[string]*model.Webhook, len(runtimeIDsToBeNotified))
	for i := range webhooks {
		if runtimeIDsToBeNotified[webhooks[i].ObjectID] {
			webhooksToCall[webhooks[i].ObjectID] = webhooks[i]
		}
	}

	requests := make([]*webhookclient.NotificationRequest, 0, len(runtimeIDsToBeNotified))
	for rtID := range runtimeIDsToBeNotified {
		rtCtx := runtimeContextsInScenarioForListeningRuntimesMapping[rtID]
		if rtCtx == nil {
			log.C(ctx).Infof("There is no runtime context for runtime %s in scenario %s. Will proceed without runtime context in the input for webhook %s", rtID, formation.Name, webhooksToCall[rtID].ID)
		}
		runtime := listeningRuntimesMapping[rtID]
		if appTemplateWithLabels == nil {
			log.C(ctx).Infof("Application %s has no application template. Will proceed without application template in the input for webhook %s", appID, webhooksToCall[rtID].ID)
		}
		input := &webhookdir.FormationConfigurationChangeInput{
			Operation:             operation,
			FormationID:           formation.ID,
			ApplicationTemplate:   appTemplateWithLabels,
			Application:           applicationWithLabels,
			Runtime:               runtime,
			RuntimeContext:        rtCtx,
			CustomerTenantContext: customerTenantContext,
			Assignment:            emptyFormationAssignment,
			ReverseAssignment:     emptyFormationAssignment,
		}
		req, err := ns.createWebhookRequest(ctx, webhooksToCall[runtime.ID], input)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, nil
}

func (ns *notificationsService) generateNotificationsForApplicationsAboutTheApplicationThatIsAssigned(ctx context.Context, tenant string, appID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.NotificationRequest, error) {
	log.C(ctx).Infof("Generating %s app-to-app formation notifications for application %s", operation, appID)
	applicationWithLabels, appTemplateWithLabels, err := ns.webhookDataInputBuilder.PrepareApplicationAndAppTemplateWithLabels(ctx, tenant, appID)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing application and application template with labels")
	}

	webhooks, err := ns.webhookRepository.ListByReferenceObjectTypeAndWebhookType(ctx, tenant, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference)
	if err != nil {
		return nil, errors.Wrap(err, "when listing application tenant mapping webhooks for applications")
	}

	listeningAppIDs := make(map[string]bool, len(webhooks))
	for _, wh := range webhooks {
		listeningAppIDs[wh.ObjectID] = true
	}

	if len(listeningAppIDs) == 0 {
		log.C(ctx).Infof("There are no applications listening for app-to-app formation notifications in tenant %s", tenant)
		return nil, nil
	}

	log.C(ctx).Infof("There are %d applications listening for app-to-app formation notifications in tenant %s", len(listeningAppIDs), tenant)

	requests := make([]*webhookclient.NotificationRequest, 0, len(listeningAppIDs))
	if listeningAppIDs[appID] {
		log.C(ctx).Infof("The application with ID %s that is being %s is also listening for app-to-app formation notifications. Will create notifications about all other apps in the formation...", appID, operation)
		var webhook *model.Webhook
		for i := range webhooks {
			if webhooks[i].ObjectID == appID {
				webhook = webhooks[i]
			}
		}

		applicationMappingsToBeNotifiedFor, applicationTemplatesMapping, err := ns.prepareApplicationMappingsInFormation(ctx, tenant, formation, appID)
		if err != nil {
			return nil, err
		}

		appsInFormationCountExcludingAppCurrentlyAssigned := len(applicationMappingsToBeNotifiedFor)
		if operation == model.AssignFormation {
			appsInFormationCountExcludingAppCurrentlyAssigned -= 1
		}

		log.C(ctx).Infof("There are %d applications in formation %s. Notification will be sent about them to application with id %s that is being %s.", appsInFormationCountExcludingAppCurrentlyAssigned, formation.Name, appID, operation)

		for _, sourceApp := range applicationMappingsToBeNotifiedFor {
			if sourceApp.ID == appID {
				continue // Do not notify about itself
			}
			var appTemplate *webhookdir.ApplicationTemplateWithLabels
			if sourceApp.ApplicationTemplateID != nil {
				appTemplate = applicationTemplatesMapping[*sourceApp.ApplicationTemplateID]
			} else {
				log.C(ctx).Infof("Application %s has no application template. Will proceed without application template for source application in the input for webhook %s", sourceApp.ID, webhook.ID)
			}
			if appTemplateWithLabels == nil {
				log.C(ctx).Infof("Application %s has no application template. Will proceed without application template for target application in the input for webhook %s", appID, webhook.ID)
			}
			input := &webhookdir.ApplicationTenantMappingInput{
				Operation:                 operation,
				FormationID:               formation.ID,
				SourceApplicationTemplate: appTemplate,
				SourceApplication:         sourceApp,
				TargetApplicationTemplate: appTemplateWithLabels,
				TargetApplication:         applicationWithLabels,
				CustomerTenantContext:     customerTenantContext,
				Assignment:                emptyFormationAssignment,
				ReverseAssignment:         emptyFormationAssignment,
			}
			req, err := ns.createWebhookRequest(ctx, webhook, input)
			if err != nil {
				return nil, err
			}
			requests = append(requests, req)
		}

		delete(listeningAppIDs, appID)
	}

	listeningAppsInScenario, err := ns.applicationRepository.ListByScenariosAndIDs(ctx, tenant, []string{formation.Name}, setToSlice(listeningAppIDs))
	if err != nil {
		return nil, errors.Wrapf(err, "while listing applications in scenario %s", formation.Name)
	}

	log.C(ctx).Infof("There are %d out of %d applications listening for app-to-app formation notifications in tenant %s that are in scenario %s", len(listeningAppsInScenario), len(listeningAppIDs), tenant, formation.Name)

	appIDsToBeNotified := make(map[string]bool, len(listeningAppsInScenario))
	applicationsTemplateIDs := make([]string, 0, len(listeningAppsInScenario))
	for _, app := range listeningAppsInScenario {
		appIDsToBeNotified[app.ID] = true
		if app.ApplicationTemplateID != nil {
			applicationsTemplateIDs = append(applicationsTemplateIDs, *app.ApplicationTemplateID)
		}
	}

	listeningAppsLabels, err := ns.labelRepository.ListForObjectIDs(ctx, tenant, model.ApplicationLabelableObject, setToSlice(appIDsToBeNotified))
	if err != nil {
		return nil, errors.Wrap(err, "while listing application labels")
	}

	listeningAppsMapping := make(map[string]*webhookdir.ApplicationWithLabels, len(listeningAppsInScenario))
	for i, app := range listeningAppsInScenario {
		listeningAppsMapping[app.ID] = &webhookdir.ApplicationWithLabels{
			Application: listeningAppsInScenario[i],
			Labels:      listeningAppsLabels[app.ID],
		}
	}

	applicationTemplates, err := ns.applicationTemplateRepository.ListByIDs(ctx, applicationsTemplateIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing application templates")
	}
	applicationTemplatesLabels, err := ns.labelRepository.ListForObjectIDs(ctx, tenant, model.AppTemplateLabelableObject, applicationsTemplateIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing labels for application templates")
	}
	applicationTemplatesMapping := make(map[string]*webhookdir.ApplicationTemplateWithLabels, len(applicationTemplates))
	for i, appTemplate := range applicationTemplates {
		applicationTemplatesMapping[appTemplate.ID] = &webhookdir.ApplicationTemplateWithLabels{
			ApplicationTemplate: applicationTemplates[i],
			Labels:              applicationTemplatesLabels[appTemplate.ID],
		}
	}

	webhooksToCall := make(map[string]*model.Webhook, len(appIDsToBeNotified))
	for i := range webhooks {
		if appIDsToBeNotified[webhooks[i].ObjectID] {
			webhooksToCall[webhooks[i].ObjectID] = webhooks[i]
		}
	}

	for _, targetApp := range listeningAppsMapping {
		var appTemplate *webhookdir.ApplicationTemplateWithLabels
		if targetApp.ApplicationTemplateID != nil {
			appTemplate = applicationTemplatesMapping[*targetApp.ApplicationTemplateID]
		} else {
			log.C(ctx).Infof("Application %s has no application template. Will proceed without application template for the target application in the input for webhook %s", targetApp.ID, webhooksToCall[targetApp.ID].ID)
		}
		if appTemplateWithLabels == nil {
			log.C(ctx).Infof("Application %s has no application template. Will proceed without application template for source application in the input for webhook %s", appID, webhooksToCall[targetApp.ID].ID)
		}
		input := &webhookdir.ApplicationTenantMappingInput{
			Operation:                 operation,
			FormationID:               formation.ID,
			SourceApplicationTemplate: appTemplateWithLabels,
			SourceApplication:         applicationWithLabels,
			TargetApplicationTemplate: appTemplate,
			TargetApplication:         targetApp,
			CustomerTenantContext:     customerTenantContext,
			Assignment:                emptyFormationAssignment,
			ReverseAssignment:         emptyFormationAssignment,
		}
		req, err := ns.createWebhookRequest(ctx, webhooksToCall[targetApp.ID], input)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	log.C(ctx).Infof("Total number of app-to-app notifications for application with ID %s that is being %s is %d", appID, operation, len(requests))

	return requests, nil
}

func (ns *notificationsService) generateNotificationsForApplicationsAboutTheRuntimeContextThatIsAssigned(ctx context.Context, tenant, runtimeCtxID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.NotificationRequest, error) {
	log.C(ctx).Infof("Generating %s notifications about runtime context %s for all interested applications in the formation", operation, runtimeCtxID)
	runtimeCtxWithLabels, err := ns.webhookDataInputBuilder.PrepareRuntimeContextWithLabels(ctx, tenant, runtimeCtxID)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing runtime context with labels")
	}

	requests, err := ns.generateNotificationsForApplicationsAboutTheRuntimeThatIsAssigned(ctx, tenant, runtimeCtxWithLabels.RuntimeID, formation, operation, customerTenantContext)
	if err != nil {
		return nil, err
	}
	for _, request := range requests {
		request.Object.(*webhookdir.FormationConfigurationChangeInput).RuntimeContext = runtimeCtxWithLabels
	}
	return requests, nil
}

func (ns *notificationsService) generateNotificationsForApplicationsAboutTheRuntimeThatIsAssigned(ctx context.Context, tenant, runtimeID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.NotificationRequest, error) {
	log.C(ctx).Infof("Generating %s notifications about runtime %s for all interested applications in the formation", operation, runtimeID)
	runtimeWithLabels, err := ns.webhookDataInputBuilder.PrepareRuntimeWithLabels(ctx, tenant, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing runtime with labels")
	}

	webhooks, err := ns.webhookRepository.ListByReferenceObjectTypeAndWebhookType(ctx, tenant, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference)
	if err != nil {
		return nil, errors.Wrap(err, "when listing configuration changed webhooks for applications")
	}

	listeningApplicationIDs := make([]string, 0, len(webhooks))
	for _, wh := range webhooks {
		listeningApplicationIDs = append(listeningApplicationIDs, wh.ObjectID)
	}

	if len(listeningApplicationIDs) == 0 {
		log.C(ctx).Infof("There are no applications listening for formation notifications in tenant %s", tenant)
		return nil, nil
	}

	log.C(ctx).Infof("There are %d applications listening for formation notifications in tenant %s", len(listeningApplicationIDs), tenant)

	listeningApplicationsInScenario, err := ns.applicationRepository.ListByScenariosAndIDs(ctx, tenant, []string{formation.Name}, listeningApplicationIDs)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing applications in scenario %s", formation.Name)
	}

	if len(listeningApplicationsInScenario) == 0 {
		log.C(ctx).Infof("There are no applications in scenario %s. No notifications will be generated about runtime with ID: %s", formation.Name, runtimeID)
		return nil, nil
	}

	log.C(ctx).Infof("There are %d out of %d applications listening for runtime-to-app formation notifications in tenant %s that are in scenario %s", len(listeningApplicationsInScenario), len(listeningApplicationIDs), tenant, formation.Name)

	applicationsToBeNotifiedIDs := make(map[string]bool, len(listeningApplicationsInScenario))
	applicationsTemplateIDs := make([]string, 0, len(listeningApplicationsInScenario))
	for _, app := range listeningApplicationsInScenario {
		applicationsToBeNotifiedIDs[app.ID] = true
		if app.ApplicationTemplateID != nil {
			applicationsTemplateIDs = append(applicationsTemplateIDs, *app.ApplicationTemplateID)
		}
	}

	applicationsToBeNotifiedForLabels, err := ns.labelRepository.ListForObjectIDs(ctx, tenant, model.ApplicationLabelableObject, setToSlice(applicationsToBeNotifiedIDs))
	if err != nil {
		return nil, errors.Wrap(err, "while listing labels for applications")
	}
	applicationMapping := make(map[string]*webhookdir.ApplicationWithLabels, len(applicationsToBeNotifiedIDs))
	for i, app := range listeningApplicationsInScenario {
		applicationMapping[app.ID] = &webhookdir.ApplicationWithLabels{
			Application: listeningApplicationsInScenario[i],
			Labels:      applicationsToBeNotifiedForLabels[app.ID],
		}
	}

	applicationTemplates, err := ns.applicationTemplateRepository.ListByIDs(ctx, applicationsTemplateIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing application templates")
	}
	applicationTemplatesLabels, err := ns.labelRepository.ListForObjectIDs(ctx, tenant, model.AppTemplateLabelableObject, applicationsTemplateIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing labels for application templates")
	}
	applicationTemplatesMapping := make(map[string]*webhookdir.ApplicationTemplateWithLabels, len(applicationTemplates))
	for i, appTemplate := range applicationTemplates {
		applicationTemplatesMapping[appTemplate.ID] = &webhookdir.ApplicationTemplateWithLabels{
			ApplicationTemplate: applicationTemplates[i],
			Labels:              applicationTemplatesLabels[appTemplate.ID],
		}
	}

	webhooksToCall := make(map[string]*model.Webhook, len(applicationsToBeNotifiedIDs))
	for _, wh := range webhooks {
		if applicationsToBeNotifiedIDs[wh.ObjectID] {
			webhooksToCall[wh.ObjectID] = wh
		}
	}

	requests := make([]*webhookclient.NotificationRequest, 0, len(applicationMapping))
	for _, app := range applicationMapping {
		var appTemplate *webhookdir.ApplicationTemplateWithLabels
		if app.ApplicationTemplateID != nil {
			appTemplate = applicationTemplatesMapping[*app.ApplicationTemplateID]
		} else {
			log.C(ctx).Infof("Application %s has no application template. Will proceed without application template in the input for webhook %s", app.ID, webhooksToCall[app.ID].ID)
		}
		input := &webhookdir.FormationConfigurationChangeInput{
			Operation:             operation,
			FormationID:           formation.ID,
			ApplicationTemplate:   appTemplate,
			Application:           app,
			Runtime:               runtimeWithLabels,
			RuntimeContext:        nil,
			CustomerTenantContext: customerTenantContext,
			Assignment:            emptyFormationAssignment,
			ReverseAssignment:     emptyFormationAssignment,
		}
		req, err := ns.createWebhookRequest(ctx, webhooksToCall[app.ID], input)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

func (ns *notificationsService) generateNotificationsAboutApplicationsForTheRuntimeContextThatIsAssigned(ctx context.Context, tenant, runtimeCtxID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.NotificationRequest, error) {
	log.C(ctx).Infof("Generating %s notifications for runtime context %s", operation, runtimeCtxID)
	runtimeCtxWithLabels, err := ns.webhookDataInputBuilder.PrepareRuntimeContextWithLabels(ctx, tenant, runtimeCtxID)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing runtime context with labels")
	}

	requests, err := ns.generateNotificationsAboutApplicationsForTheRuntimeThatIsAssigned(ctx, tenant, runtimeCtxWithLabels.RuntimeID, formation, operation, customerTenantContext)
	if err != nil {
		return nil, err
	}
	for _, request := range requests {
		request.Object.(*webhookdir.FormationConfigurationChangeInput).RuntimeContext = runtimeCtxWithLabels
	}
	return requests, nil
}

func (ns *notificationsService) generateNotificationsAboutApplicationsForTheRuntimeThatIsAssigned(ctx context.Context, tenant, runtimeID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.NotificationRequest, error) {
	log.C(ctx).Infof("Generating %s notifications about all applications in the formation for runtime %s", operation, runtimeID)
	runtimeWithLabels, err := ns.webhookDataInputBuilder.PrepareRuntimeWithLabels(ctx, tenant, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing runtime with labels")
	}

	webhook, err := ns.webhookRepository.GetByIDAndWebhookType(ctx, tenant, runtimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			log.C(ctx).Infof("There is no configuration changed webhook for runtime %s. There are no notifications to be generated.", runtimeID)
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while listing configuration changed webhooks for runtime %s", runtimeID)
	}

	applicationMapping, applicationTemplatesMapping, err := ns.prepareApplicationMappingsInFormation(ctx, tenant, formation, runtimeID)
	if err != nil {
		return nil, err
	}

	requests := make([]*webhookclient.NotificationRequest, 0, len(applicationMapping))
	for _, app := range applicationMapping {
		var appTemplate *webhookdir.ApplicationTemplateWithLabels
		if app.ApplicationTemplateID != nil {
			appTemplate = applicationTemplatesMapping[*app.ApplicationTemplateID]
		} else {
			log.C(ctx).Infof("Application %s has no application template. Will proceed without application template in the input for webhook %s", app.ID, webhook.ID)
		}
		input := &webhookdir.FormationConfigurationChangeInput{
			Operation:             operation,
			FormationID:           formation.ID,
			ApplicationTemplate:   appTemplate,
			Application:           app,
			Runtime:               runtimeWithLabels,
			RuntimeContext:        nil,
			CustomerTenantContext: customerTenantContext,
			Assignment:            emptyFormationAssignment,
			ReverseAssignment:     emptyFormationAssignment,
		}
		req, err := ns.createWebhookRequest(ctx, webhook, input)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, nil
}

func (ns *notificationsService) prepareApplicationMappingsInFormation(ctx context.Context, tenant string, formation *model.Formation, targetID string) (map[string]*webhookdir.ApplicationWithLabels, map[string]*webhookdir.ApplicationTemplateWithLabels, error) {
	applicationsToBeNotifiedFor, err := ns.applicationRepository.ListByScenariosNoPaging(ctx, tenant, []string{formation.Name})
	if err != nil {
		return nil, nil, errors.Wrap(err, "while listing scenario labels for applications")
	}
	if len(applicationsToBeNotifiedFor) == 0 {
		log.C(ctx).Infof("There are no applications in scenario %s. No notifications will be generated for %s", formation.Name, targetID)
		return nil, nil, nil
	}
	applicationsToBeNotifiedForIDs := make([]string, 0, len(applicationsToBeNotifiedFor))
	applicationsTemplateIDs := make([]string, 0, len(applicationsToBeNotifiedFor))
	for _, app := range applicationsToBeNotifiedFor {
		applicationsToBeNotifiedForIDs = append(applicationsToBeNotifiedForIDs, app.ID)
		if app.ApplicationTemplateID != nil {
			applicationsTemplateIDs = append(applicationsTemplateIDs, *app.ApplicationTemplateID)
		}
	}

	applicationsToBeNotifiedForLabels, err := ns.labelRepository.ListForObjectIDs(ctx, tenant, model.ApplicationLabelableObject, applicationsToBeNotifiedForIDs)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while listing labels for applications")
	}
	applicationMapping := make(map[string]*webhookdir.ApplicationWithLabels, len(applicationsToBeNotifiedForIDs))
	for i, app := range applicationsToBeNotifiedFor {
		applicationMapping[app.ID] = &webhookdir.ApplicationWithLabels{
			Application: applicationsToBeNotifiedFor[i],
			Labels:      applicationsToBeNotifiedForLabels[app.ID],
		}
	}

	applicationTemplates, err := ns.applicationTemplateRepository.ListByIDs(ctx, applicationsTemplateIDs)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while listing application templates")
	}
	applicationTemplatesLabels, err := ns.labelRepository.ListForObjectIDs(ctx, tenant, model.AppTemplateLabelableObject, applicationsTemplateIDs)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while listing labels for application templates")
	}
	applicationTemplatesMapping := make(map[string]*webhookdir.ApplicationTemplateWithLabels, len(applicationTemplates))
	for i, appTemplate := range applicationTemplates {
		applicationTemplatesMapping[appTemplate.ID] = &webhookdir.ApplicationTemplateWithLabels{
			ApplicationTemplate: applicationTemplates[i],
			Labels:              applicationTemplatesLabels[appTemplate.ID],
		}
	}

	return applicationMapping, applicationTemplatesMapping, nil
}

func (ns *notificationsService) createWebhookRequest(ctx context.Context, webhook *model.Webhook, input webhookdir.FormationAssignmentTemplateInput) (*webhookclient.NotificationRequest, error) {
	gqlWebhook, err := ns.webhookConverter.ToGraphQL(webhook)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting webhook with ID %s", webhook.ID)
	}
	return &webhookclient.NotificationRequest{
		Webhook:       *gqlWebhook,
		Object:        input,
		CorrelationID: correlation.CorrelationIDFromContext(ctx),
	}, nil
}

func (ns *notificationsService) extractCustomerTenantContext(ctx context.Context, internalTenantID string) (*webhookdir.CustomerTenantContext, error) {
	tenantObject, err := ns.tenantRepository.Get(ctx, internalTenantID)
	if err != nil {
		return nil, err
	}

	customerID, err := ns.tenantRepository.GetCustomerIDParentRecursively(ctx, internalTenantID)
	if err != nil {
		return nil, err
	}

	return &webhookdir.CustomerTenantContext{
		Tenant:     tenantObject.ExternalTenant,
		CustomerID: customerID,
	}, nil
}
