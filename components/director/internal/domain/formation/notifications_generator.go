package formation

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	databuilder "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	webhookdir "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=applicationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.Application, error)
	ListAllByIDs(ctx context.Context, tenantID string, ids []string) ([]*model.Application, error)
	ListByScenariosNoPaging(ctx context.Context, tenant string, scenarios []string) ([]*model.Application, error)
	ListByScenariosAndIDs(ctx context.Context, tenant string, scenarios []string, ids []string) ([]*model.Application, error)
	ListListeningApplications(ctx context.Context, tenant string, whType model.WebhookType) ([]*model.Application, error)
}

//go:generate mockery --exported --name=applicationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationTemplateRepository interface {
	Get(ctx context.Context, id string) (*model.ApplicationTemplate, error)
	ListByIDs(ctx context.Context, ids []string) ([]*model.ApplicationTemplate, error)
}

//go:generate mockery --exported --name=webhookRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type webhookRepository interface {
	ListByReferenceObjectTypeAndWebhookType(ctx context.Context, tenant string, whType model.WebhookType, objType model.WebhookReferenceObjectType) ([]*model.Webhook, error)
	ListByReferenceObjectTypesAndWebhookType(ctx context.Context, tenant string, whType model.WebhookType, objTypes []model.WebhookReferenceObjectType) ([]*model.Webhook, error)
	ListByReferenceObjectIDGlobal(ctx context.Context, objID string, objType model.WebhookReferenceObjectType) ([]*model.Webhook, error)
	GetByIDAndWebhookType(ctx context.Context, tenant, objectID string, objectType model.WebhookReferenceObjectType, webhookType model.WebhookType) (*model.Webhook, error)
}

//go:generate mockery --exported --name=notificationBuilder --output=automock --outpkg=automock --case=underscore --disable-version-string
type notificationBuilder interface {
	BuildFormationAssignmentNotificationRequest(ctx context.Context, formationTemplateID string, joinPointDetails *formationconstraint.GenerateFormationAssignmentNotificationOperationDetails, webhook *model.Webhook) (*webhookclient.FormationAssignmentNotificationRequest, error)
	BuildFormationNotificationRequests(ctx context.Context, joinPointDetails *formationconstraint.GenerateFormationNotificationOperationDetails, formation *model.Formation, formationTemplateWebhooks []*model.Webhook) ([]*webhookclient.FormationNotificationRequest, error)
	PrepareDetailsForConfigurationChangeNotificationGeneration(operation model.FormationOperation, formationID string, formationTemplateID string, applicationTemplate *webhookdir.ApplicationTemplateWithLabels, application *webhookdir.ApplicationWithLabels, runtime *webhookdir.RuntimeWithLabels, runtimeContext *webhookdir.RuntimeContextWithLabels, assignment *webhookdir.FormationAssignment, reverseAssignment *webhookdir.FormationAssignment, targetType model.ResourceType, tenantContext *webhookdir.CustomerTenantContext, tenantID string) (*formationconstraint.GenerateFormationAssignmentNotificationOperationDetails, error)
	PrepareDetailsForApplicationTenantMappingNotificationGeneration(operation model.FormationOperation, formationID string, formationTemplateID string, sourceApplicationTemplate *webhookdir.ApplicationTemplateWithLabels, sourceApplication *webhookdir.ApplicationWithLabels, targetApplicationTemplate *webhookdir.ApplicationTemplateWithLabels, targetApplication *webhookdir.ApplicationWithLabels, assignment *webhookdir.FormationAssignment, reverseAssignment *webhookdir.FormationAssignment, tenantContext *webhookdir.CustomerTenantContext, tenantID string) (*formationconstraint.GenerateFormationAssignmentNotificationOperationDetails, error)
}

// NotificationsGenerator is responsible for generation of notification requests
type NotificationsGenerator struct {
	applicationRepository         applicationRepository
	applicationTemplateRepository applicationTemplateRepository
	runtimeRepo                   runtimeRepository
	runtimeContextRepo            runtimeContextRepository
	labelRepository               labelRepository
	webhookRepository             webhookRepository
	webhookDataInputBuilder       databuilder.DataInputBuilder
	notificationBuilder           notificationBuilder
}

// NewNotificationsGenerator returns an instance of NotificationsGenerator
func NewNotificationsGenerator(
	applicationRepository applicationRepository,
	applicationTemplateRepository applicationTemplateRepository,
	runtimeRepo runtimeRepository,
	runtimeContextRepo runtimeContextRepository,
	labelRepository labelRepository,
	webhookRepository webhookRepository,
	webhookDataInputBuilder databuilder.DataInputBuilder,
	notificationBuilder notificationBuilder) *NotificationsGenerator {
	return &NotificationsGenerator{
		applicationRepository:         applicationRepository,
		applicationTemplateRepository: applicationTemplateRepository,
		runtimeRepo:                   runtimeRepo,
		runtimeContextRepo:            runtimeContextRepo,
		labelRepository:               labelRepository,
		webhookRepository:             webhookRepository,
		webhookDataInputBuilder:       webhookDataInputBuilder,
		notificationBuilder:           notificationBuilder,
	}
}

// GenerateNotificationsAboutRuntimeAndRuntimeContextForTheApplicationThatIsAssigned generates notification with target the application that is assigned about for each runtime and each runtimeContext that is part of the formation
func (ns *NotificationsGenerator) GenerateNotificationsAboutRuntimeAndRuntimeContextForTheApplicationThatIsAssigned(ctx context.Context, tenant string, appID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.FormationAssignmentNotificationRequest, error) {
	log.C(ctx).Infof("Generating %q notifications during %q operation about runtimes and runtime contexts in the same formation for application with ID: %q", model.WebhookTypeConfigurationChanged, operation, appID)
	applicationWithLabels, appTemplateWithLabels, err := ns.webhookDataInputBuilder.PrepareApplicationAndAppTemplateWithLabels(ctx, tenant, appID)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing application and application template with labels")
	}

	appTemplateID := ""
	if appTemplateWithLabels != nil {
		appTemplateID = appTemplateWithLabels.ID
	}

	webhook, err := formationassignment.GetWebhookForApplication(ctx, ns.webhookRepository, tenant, appID, appTemplateID, model.WebhookTypeConfigurationChanged)
	if err != nil {
		return nil, err
	}
	if webhook == nil {
		return nil, nil
	}

	runtimesMapping, runtimesToRuntimeContextsMapping, err := ns.webhookDataInputBuilder.PrepareRuntimesAndRuntimeContextsMappingsInFormation(ctx, tenant, formation.Name)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing runtime and runtime contexts mappings")
	}

	requests := make([]*webhookclient.FormationAssignmentNotificationRequest, 0, len(runtimesMapping))
	for rtID := range runtimesMapping {
		rtCtx := runtimesToRuntimeContextsMapping[rtID]
		if rtCtx == nil {
			log.C(ctx).Infof("There is no runtime context for runtime with ID: %q in formation %q. Will proceed without runtime context in the input for webhook with ID: %q", rtID, formation.Name, webhook.ID)
		}
		runtime := runtimesMapping[rtID]
		if appTemplateWithLabels == nil {
			log.C(ctx).Infof("Application with ID: %q has no application template. Will proceed without application template in the input for webhook with ID: %q", appID, webhook.ID)
		}

		details, err := ns.notificationBuilder.PrepareDetailsForConfigurationChangeNotificationGeneration(
			operation,
			formation.ID,
			formation.FormationTemplateID,
			appTemplateWithLabels,
			applicationWithLabels,
			runtime,
			rtCtx,
			emptyFormationAssignment,
			emptyFormationAssignment,
			model.ApplicationResourceType,
			customerTenantContext,
			tenant)
		if err != nil {
			return nil, err
		}
		req, err := ns.notificationBuilder.BuildFormationAssignmentNotificationRequest(ctx, formation.FormationTemplateID, details, webhook)
		if err != nil {
			log.C(ctx).Errorf("Failed to build formation assignment notification request due to: %v", err)
		} else {
			requests = append(requests, req)
		}
	}
	return requests, nil
}

// GenerateNotificationsForRuntimeAboutTheApplicationThatIsAssigned generates notification per runtime that is part of the formation with target the runtime and source the application on which `operation` is performed
func (ns *NotificationsGenerator) GenerateNotificationsForRuntimeAboutTheApplicationThatIsAssigned(ctx context.Context, tenant string, appID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.FormationAssignmentNotificationRequest, error) {
	log.C(ctx).Infof("Generating %q notifications during %q operation about application with ID: %q for all listening runtimes in the same formation", model.WebhookTypeConfigurationChanged, operation, appID)
	applicationWithLabels, appTemplateWithLabels, err := ns.webhookDataInputBuilder.PrepareApplicationAndAppTemplateWithLabels(ctx, tenant, appID)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing application and application template with labels")
	}

	webhooks, err := ns.webhookRepository.ListByReferenceObjectTypeAndWebhookType(ctx, tenant, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference)
	if err != nil {
		return nil, errors.Wrap(err, "when listing configuration changed webhooks for runtimes")
	}

	if len(webhooks) == 0 {
		log.C(ctx).Infof("There are no runtimes listening for %q notifications in tenant %q for formation with name: %q", model.WebhookTypeConfigurationChanged, tenant, formation.Name)
		return nil, nil
	}

	listeningRuntimeIDs := make([]string, 0, len(webhooks))
	for _, wh := range webhooks {
		listeningRuntimeIDs = append(listeningRuntimeIDs, wh.ObjectID)
	}

	log.C(ctx).Infof("There is/are: %d runtimes listening for %q notifications in tenant %q for formation with name: %q", len(listeningRuntimeIDs), model.WebhookTypeConfigurationChanged, tenant, formation.Name)

	runtimesInFormationMappings, runtimeIDToRuntimeContextInFormationMappings, err := ns.webhookDataInputBuilder.PrepareRuntimesAndRuntimeContextsMappingsInFormation(ctx, tenant, formation.Name)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing runtime and runtime contexts mappings")
	}

	runtimeIDsToBeNotified := make(map[string]bool, len(listeningRuntimeIDs))
	for i := range listeningRuntimeIDs {
		if runtimesInFormationMappings[listeningRuntimeIDs[i]] != nil {
			runtimeIDsToBeNotified[listeningRuntimeIDs[i]] = true
		}
	}

	webhooksToCall := make(map[string]*model.Webhook, len(runtimeIDsToBeNotified))
	for i := range webhooks {
		if runtimeIDsToBeNotified[webhooks[i].ObjectID] {
			webhooksToCall[webhooks[i].ObjectID] = webhooks[i]
		}
	}

	requests := make([]*webhookclient.FormationAssignmentNotificationRequest, 0, len(runtimeIDsToBeNotified))
	for rtID := range runtimeIDsToBeNotified {
		rtCtx := runtimeIDToRuntimeContextInFormationMappings[rtID]
		if rtCtx == nil {
			log.C(ctx).Infof("There is no runtime context for runtime with ID: %q in formation %q. Will proceed without runtime context in the input for webhook with ID: %q", rtID, formation.Name, webhooksToCall[rtID].ID)
		}
		runtime := runtimesInFormationMappings[rtID]
		if appTemplateWithLabels == nil {
			log.C(ctx).Infof("Application with ID: %q has no application template. Will proceed without application template in the input for webhook with ID: %q", appID, webhooksToCall[rtID].ID)
		}

		details, err := ns.notificationBuilder.PrepareDetailsForConfigurationChangeNotificationGeneration(
			operation,
			formation.ID,
			formation.FormationTemplateID,
			appTemplateWithLabels,
			applicationWithLabels,
			runtime,
			rtCtx,
			emptyFormationAssignment,
			emptyFormationAssignment,
			model.RuntimeResourceType,
			customerTenantContext,
			tenant)
		if err != nil {
			return nil, err
		}

		req, err := ns.notificationBuilder.BuildFormationAssignmentNotificationRequest(ctx, formation.FormationTemplateID, details, webhooksToCall[runtime.ID])
		if err != nil {
			log.C(ctx).Errorf("Failed to build formation assignment notification request due to: %v", err)
		} else {
			requests = append(requests, req)
		}
	}

	return requests, nil
}

// GenerateNotificationsForApplicationsAboutTheApplicationThatIsAssigned generates notification per application that is part of the formation with target the application and source the application on which `operation` is performed
func (ns *NotificationsGenerator) GenerateNotificationsForApplicationsAboutTheApplicationThatIsAssigned(ctx context.Context, tenant string, appID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.FormationAssignmentNotificationRequest, error) {
	log.C(ctx).Infof("Generating %q notifications during %q operation for application with ID: %q", model.WebhookTypeApplicationTenantMapping, operation, appID)
	applicationWithLabels, appTemplateWithLabels, err := ns.webhookDataInputBuilder.PrepareApplicationAndAppTemplateWithLabels(ctx, tenant, appID)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing application and application template with labels")
	}

	webhooks, err := ns.webhookRepository.ListByReferenceObjectTypesAndWebhookType(ctx, tenant, model.WebhookTypeApplicationTenantMapping, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference})
	if err != nil {
		return nil, errors.Wrapf(err, "when listing %q webhooks for applications and their application templates", model.WebhookTypeApplicationTenantMapping)
	}

	resourceIDToWebhookMapping := make(map[string]*model.Webhook, len(webhooks))
	for _, webhook := range webhooks {
		resourceIDToWebhookMapping[webhook.ObjectID] = webhook
	}

	// list applications that either have WebhookTypeApplicationTenantMapping webhook or their applicationTemplate has WebhookTypeApplicationTenantMapping webhook
	listeningApps, err := ns.applicationRepository.ListListeningApplications(ctx, tenant, model.WebhookTypeApplicationTenantMapping)
	if err != nil {
		return nil, errors.Wrap(err, "while listing listening applications")
	}

	if len(listeningApps) == 0 {
		log.C(ctx).Infof("There are no applications listening for %q notifications in tenant %q for formation with name: %q", tenant, model.WebhookTypeApplicationTenantMapping, formation.Name)
		return nil, nil
	}

	listeningAppsByID := make(map[string]*model.Application, len(listeningApps))
	for i := range listeningApps {
		listeningAppsByID[listeningApps[i].ID] = listeningApps[i]
	}

	appIDToWebhookMapping := make(map[string]*model.Webhook)
	for _, app := range listeningApps {
		// if webhook for the application exists use it
		// if webhook for the application does not exist use the webhook for its application template
		if resourceIDToWebhookMapping[app.ID] != nil {
			appIDToWebhookMapping[app.ID] = resourceIDToWebhookMapping[app.ID]
		} else {
			appIDToWebhookMapping[app.ID] = resourceIDToWebhookMapping[str.PtrStrToStr(app.ApplicationTemplateID)]
		}
	}

	log.C(ctx).Infof("There are %d applications listening for %q notifications in tenant %q for formation with name: %q", len(listeningAppsByID), model.WebhookTypeApplicationTenantMapping, tenant, formation.Name)

	applicationsInFormationMapping, appTemplatesMapping, err := ns.webhookDataInputBuilder.PrepareApplicationMappingsInFormation(ctx, tenant, formation.Name)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing application and application template mappings")
	}

	requests := make([]*webhookclient.FormationAssignmentNotificationRequest, 0, len(listeningAppsByID))
	if listeningAppsByID[appID] != nil {
		log.C(ctx).Infof("The application with ID: %q that is being %q is also listening for %q notifications. Will create notifications about all other apps in the formation...", appID, operation, model.WebhookTypeApplicationTenantMapping)
		webhook := appIDToWebhookMapping[appID]

		appsInFormationCountExcludingAppCurrentlyAssigned := len(applicationsInFormationMapping)
		if operation == model.AssignFormation {
			appsInFormationCountExcludingAppCurrentlyAssigned -= 1
		}

		log.C(ctx).Infof("The number of other application(s) in formation %q is/are: %d. Notification(s) will be sent about them to application with ID: %q that is being %q.", formation.Name, appsInFormationCountExcludingAppCurrentlyAssigned, appID, operation)

		for _, sourceApp := range applicationsInFormationMapping {
			if sourceApp.ID == appID {
				continue // Do not notify about itself
			}
			var appTemplate *webhookdir.ApplicationTemplateWithLabels
			if sourceApp.ApplicationTemplateID != nil {
				appTemplate = appTemplatesMapping[*sourceApp.ApplicationTemplateID]
			} else {
				log.C(ctx).Infof("Application with ID: %q has no application template. Will proceed without application template for the source application in the input for webhook with ID: %q", sourceApp.ID, webhook.ID)
			}
			if appTemplateWithLabels == nil {
				log.C(ctx).Infof("Application with ID: %q has no application template. Will proceed without application template for the target application in the input for webhook with ID: %q", appID, webhook.ID)
			}

			details, err := ns.notificationBuilder.PrepareDetailsForApplicationTenantMappingNotificationGeneration(
				operation,
				formation.ID,
				formation.FormationTemplateID,
				appTemplate,
				sourceApp,
				appTemplateWithLabels,
				applicationWithLabels,
				emptyFormationAssignment,
				emptyFormationAssignment,
				customerTenantContext,
				tenant,
			)
			if err != nil {
				return nil, err
			}

			req, err := ns.notificationBuilder.BuildFormationAssignmentNotificationRequest(ctx, formation.FormationTemplateID, details, webhook)
			if err != nil {
				log.C(ctx).Errorf("Failed to build formation assignment notification request due to: %v", err)
			} else {
				requests = append(requests, req)
			}
		}

		delete(listeningAppsByID, appID)
	}

	listeningApplicationsInFormationIds := make([]string, 0, len(listeningAppsByID))
	for id := range listeningAppsByID {
		if applicationsInFormationMapping[id] != nil {
			listeningApplicationsInFormationIds = append(listeningApplicationsInFormationIds, id)
		}
	}

	for _, appID := range listeningApplicationsInFormationIds {
		targetApp := applicationsInFormationMapping[appID]
		var targetAppTemplate *webhookdir.ApplicationTemplateWithLabels
		if targetApp.ApplicationTemplateID != nil {
			targetAppTemplate = appTemplatesMapping[*targetApp.ApplicationTemplateID]
		} else {
			log.C(ctx).Infof("Application with ID: %q has no application template. Will proceed without application template for the target application in the input for webhook with ID: %q", appID, appIDToWebhookMapping[appID].ID)
		}
		if appTemplateWithLabels == nil {
			log.C(ctx).Infof("Application with ID: %q has no application template. Will proceed without application template for the source application in the input for webhook with ID: %q", appID, appIDToWebhookMapping[appID].ID)
		}

		details, err := ns.notificationBuilder.PrepareDetailsForApplicationTenantMappingNotificationGeneration(
			operation,
			formation.ID,
			formation.FormationTemplateID,
			appTemplateWithLabels,
			applicationWithLabels,
			targetAppTemplate,
			targetApp,
			emptyFormationAssignment,
			emptyFormationAssignment,
			customerTenantContext,
			tenant,
		)
		if err != nil {
			return nil, err
		}

		req, err := ns.notificationBuilder.BuildFormationAssignmentNotificationRequest(ctx, formation.FormationTemplateID, details, appIDToWebhookMapping[appID])
		if err != nil {
			log.C(ctx).Errorf("Failed to generate notification due to: %v", err)
		} else {
			requests = append(requests, req)
		}
	}

	log.C(ctx).Infof("Total number of %q notifications for application with ID: %q that is being %q is/are: %d", model.WebhookTypeApplicationTenantMapping, appID, operation, len(requests))

	return requests, nil
}

// GenerateNotificationsForApplicationsAboutTheRuntimeContextThatIsAssigned generates notification per application that is part of the formation with target the application and source the runtime context on which `operation` is performed
func (ns *NotificationsGenerator) GenerateNotificationsForApplicationsAboutTheRuntimeContextThatIsAssigned(ctx context.Context, tenant, runtimeCtxID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.FormationAssignmentNotificationRequest, error) {
	log.C(ctx).Infof("Generating %q notifications about runtime context with ID: %q for all interested applications in the formation", operation, runtimeCtxID)
	runtimeCtxWithLabels, err := ns.webhookDataInputBuilder.PrepareRuntimeContextWithLabels(ctx, tenant, runtimeCtxID)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing runtime context with labels")
	}

	requests, err := ns.GenerateNotificationsForApplicationsAboutTheRuntimeThatIsAssigned(ctx, tenant, runtimeCtxWithLabels.RuntimeID, formation, operation, customerTenantContext)
	if err != nil {
		return nil, err
	}
	for _, request := range requests {
		request.Object.(*webhookdir.FormationConfigurationChangeInput).RuntimeContext = runtimeCtxWithLabels
	}
	return requests, nil
}

// GenerateNotificationsForApplicationsAboutTheRuntimeThatIsAssigned generates notification per application that is part of the formation with target the application and source the runtime on which `operation` is performed
func (ns *NotificationsGenerator) GenerateNotificationsForApplicationsAboutTheRuntimeThatIsAssigned(ctx context.Context, tenant, runtimeID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.FormationAssignmentNotificationRequest, error) {
	log.C(ctx).Infof("Generating %q notifications during %q operation about runtime with ID: %q for all interested applications in the formation", model.WebhookTypeConfigurationChanged, operation, runtimeID)
	runtimeWithLabels, err := ns.webhookDataInputBuilder.PrepareRuntimeWithLabels(ctx, tenant, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing runtime with labels")
	}

	webhooks, err := ns.webhookRepository.ListByReferenceObjectTypesAndWebhookType(ctx, tenant, model.WebhookTypeConfigurationChanged, []model.WebhookReferenceObjectType{model.ApplicationWebhookReference, model.ApplicationTemplateWebhookReference})
	if err != nil {
		return nil, errors.Wrapf(err, "when listing %q webhooks for applications and their application templates", model.WebhookTypeConfigurationChanged)
	}

	resourceIDToWebhookMapping := make(map[string]*model.Webhook, len(webhooks))
	for _, webhook := range webhooks {
		resourceIDToWebhookMapping[webhook.ObjectID] = webhook
	}

	listeningApps, err := ns.applicationRepository.ListListeningApplications(ctx, tenant, model.WebhookTypeConfigurationChanged)
	if err != nil {
		return nil, errors.Wrap(err, "while listing listening applications")
	}

	if len(listeningApps) == 0 {
		log.C(ctx).Infof("There are no applications listening for %q notifications in tenant %q for formation with name: %q", tenant, model.WebhookTypeConfigurationChanged, formation.Name)
		return nil, nil
	}

	listeningApplicationIDs := make([]string, 0, len(listeningApps))
	for _, app := range listeningApps {
		listeningApplicationIDs = append(listeningApplicationIDs, app.ID)
	}

	appIDToWebhookMapping := make(map[string]*model.Webhook)
	for _, app := range listeningApps {
		// if webhook for the application exists use it
		// if webhook for the application does not exist use the webhook for its application template
		if resourceIDToWebhookMapping[app.ID] != nil {
			appIDToWebhookMapping[app.ID] = resourceIDToWebhookMapping[app.ID]
		} else {
			appIDToWebhookMapping[app.ID] = resourceIDToWebhookMapping[str.PtrStrToStr(app.ApplicationTemplateID)]
		}
	}

	log.C(ctx).Infof("There are %d applications listening for %q notifications in tenant %q for formation with ID: %q", len(listeningApplicationIDs), model.WebhookTypeConfigurationChanged, tenant, formation.Name)

	applicationsInFormationMapping, appTemplatesMapping, err := ns.webhookDataInputBuilder.PrepareApplicationMappingsInFormation(ctx, tenant, formation.Name)
	if err != nil {
		return nil, err
	}

	listeningApplicationsInFormationIds := make([]string, 0, len(listeningApps))
	for i := range listeningApps {
		if applicationsInFormationMapping[listeningApps[i].ID] != nil {
			listeningApplicationsInFormationIds = append(listeningApplicationsInFormationIds, listeningApps[i].ID)
		}
	}

	requests := make([]*webhookclient.FormationAssignmentNotificationRequest, 0, len(applicationsInFormationMapping))
	for _, appID := range listeningApplicationsInFormationIds {
		app := applicationsInFormationMapping[appID]
		var appTemplate *webhookdir.ApplicationTemplateWithLabels
		if app.ApplicationTemplateID != nil {
			appTemplate = appTemplatesMapping[*app.ApplicationTemplateID]
		} else {
			log.C(ctx).Infof("Application with ID: %q has no application template. Will proceed without application template in the input for webhook with ID: %q", appID, appIDToWebhookMapping[appID].ID)
		}

		details, err := ns.notificationBuilder.PrepareDetailsForConfigurationChangeNotificationGeneration(
			operation,
			formation.ID,
			formation.FormationTemplateID,
			appTemplate,
			app,
			runtimeWithLabels,
			nil,
			emptyFormationAssignment,
			emptyFormationAssignment,
			model.ApplicationResourceType,
			customerTenantContext,
			tenant,
		)
		if err != nil {
			return nil, err
		}

		req, err := ns.notificationBuilder.BuildFormationAssignmentNotificationRequest(ctx, formation.FormationTemplateID, details, appIDToWebhookMapping[appID])
		if err != nil {
			log.C(ctx).Errorf("Failed to build formation assignment notification request due to: %v", err)
		} else {
			requests = append(requests, req)
		}
	}
	return requests, nil
}

// GenerateNotificationsAboutApplicationsForTheRuntimeContextThatIsAssigned generates notification per runtime context that is part of the formation with target the runtime context and source the application on which `operation` is performed
func (ns *NotificationsGenerator) GenerateNotificationsAboutApplicationsForTheRuntimeContextThatIsAssigned(ctx context.Context, tenant, runtimeCtxID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.FormationAssignmentNotificationRequest, error) {
	log.C(ctx).Infof("Generating %q notifications during %q operation for runtime context with ID: %q", model.WebhookTypeConfigurationChanged, operation, runtimeCtxID)
	runtimeCtxWithLabels, err := ns.webhookDataInputBuilder.PrepareRuntimeContextWithLabels(ctx, tenant, runtimeCtxID)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing runtime context with labels")
	}

	runtimeID := runtimeCtxWithLabels.RuntimeID

	webhook, err := ns.webhookRepository.GetByIDAndWebhookType(ctx, tenant, runtimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			log.C(ctx).Infof("There is no %q webhook for runtime with ID: %q. No notifications will be generated.", model.WebhookTypeConfigurationChanged, runtimeID)
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while listing configuration changed webhooks for runtime %s", runtimeID)
	}

	runtimeWithLabels, err := ns.webhookDataInputBuilder.PrepareRuntimeWithLabels(ctx, tenant, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing runtime with labels")
	}

	applicationMapping, applicationTemplatesMapping, err := ns.webhookDataInputBuilder.PrepareApplicationMappingsInFormation(ctx, tenant, formation.Name)
	if err != nil {
		return nil, err
	}

	requests := make([]*webhookclient.FormationAssignmentNotificationRequest, 0, len(applicationMapping))
	for _, app := range applicationMapping {
		var appTemplate *webhookdir.ApplicationTemplateWithLabels
		if app.ApplicationTemplateID != nil {
			appTemplate = applicationTemplatesMapping[*app.ApplicationTemplateID]
		} else {
			log.C(ctx).Infof("Application with ID: %q has no application template. Will proceed without application template in the input for webhook with ID: %q", app.ID, webhook.ID)
		}

		details, err := ns.notificationBuilder.PrepareDetailsForConfigurationChangeNotificationGeneration(
			operation,
			formation.ID,
			formation.FormationTemplateID,
			appTemplate,
			app,
			runtimeWithLabels,
			runtimeCtxWithLabels,
			emptyFormationAssignment,
			emptyFormationAssignment,
			model.RuntimeContextResourceType,
			customerTenantContext,
			tenant,
		)
		if err != nil {
			return nil, err
		}

		req, err := ns.notificationBuilder.BuildFormationAssignmentNotificationRequest(ctx, formation.FormationTemplateID, details, webhook)
		if err != nil {
			log.C(ctx).Errorf("Failed to build formation assignment notification request due to: %v", err)
		} else {
			requests = append(requests, req)
		}
	}

	return requests, nil
}

// GenerateNotificationsAboutApplicationsForTheRuntimeThatIsAssigned generates notification per runtime that is part of the formation with target the runtime and source the application on which `operation` is performed
func (ns *NotificationsGenerator) GenerateNotificationsAboutApplicationsForTheRuntimeThatIsAssigned(ctx context.Context, tenant, runtimeID string, formation *model.Formation, operation model.FormationOperation, customerTenantContext *webhookdir.CustomerTenantContext) ([]*webhookclient.FormationAssignmentNotificationRequest, error) {
	log.C(ctx).Infof("Generating %q notifications during %q operation about all applications in the formation for runtime with ID: %q", model.WebhookTypeConfigurationChanged, operation, runtimeID)
	runtimeWithLabels, err := ns.webhookDataInputBuilder.PrepareRuntimeWithLabels(ctx, tenant, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing runtime with labels")
	}

	webhook, err := ns.webhookRepository.GetByIDAndWebhookType(ctx, tenant, runtimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			log.C(ctx).Infof("There is no %q webhook for runtime with ID: %q. No notifications will be generated.", model.WebhookTypeConfigurationChanged, runtimeID)
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while listing configuration changed webhooks for runtime %s", runtimeID)
	}

	applicationMapping, applicationTemplatesMapping, err := ns.webhookDataInputBuilder.PrepareApplicationMappingsInFormation(ctx, tenant, formation.Name)
	if err != nil {
		return nil, err
	}

	requests := make([]*webhookclient.FormationAssignmentNotificationRequest, 0, len(applicationMapping))
	for _, app := range applicationMapping {
		var appTemplate *webhookdir.ApplicationTemplateWithLabels
		if app.ApplicationTemplateID != nil {
			appTemplate = applicationTemplatesMapping[*app.ApplicationTemplateID]
		} else {
			log.C(ctx).Infof("Application with ID: %q has no application template. Will proceed without application template in the input for webhook with ID: %q", app.ID, webhook.ID)
		}

		details, err := ns.notificationBuilder.PrepareDetailsForConfigurationChangeNotificationGeneration(
			operation,
			formation.ID,
			formation.FormationTemplateID,
			appTemplate,
			app,
			runtimeWithLabels,
			nil,
			emptyFormationAssignment,
			emptyFormationAssignment,
			model.RuntimeResourceType,
			customerTenantContext,
			tenant,
		)
		if err != nil {
			return nil, err
		}

		req, err := ns.notificationBuilder.BuildFormationAssignmentNotificationRequest(ctx, formation.FormationTemplateID, details, webhook)
		if err != nil {
			log.C(ctx).Errorf("Failed to build formation assignment notification request due to: %v", err)
		} else {
			requests = append(requests, req)
		}
	}

	return requests, nil
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
