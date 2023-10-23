package datainputbuilder

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/webhook"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=applicationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.Application, error)
	ListByScenariosNoPaging(ctx context.Context, tenant string, scenarios []string) ([]*model.Application, error)
}

//go:generate mockery --exported --name=applicationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationTemplateRepository interface {
	Get(ctx context.Context, id string) (*model.ApplicationTemplate, error)
	ListByIDs(ctx context.Context, ids []string) ([]*model.ApplicationTemplate, error)
}

//go:generate mockery --exported --name=runtimeRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.Runtime, error)
	ListByScenarios(ctx context.Context, tenant string, scenarios []string) ([]*model.Runtime, error)
	ListByIDs(ctx context.Context, tenant string, ids []string) ([]*model.Runtime, error)
}

//go:generate mockery --exported --name=runtimeContextRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeContextRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.RuntimeContext, error)
	GetByRuntimeID(ctx context.Context, tenant, runtimeID string) (*model.RuntimeContext, error)
	ListByScenarios(ctx context.Context, tenant string, scenarios []string) ([]*model.RuntimeContext, error)
}

//go:generate mockery --exported --name=labelInputBuilder --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelInputBuilder interface {
	GetLabelsForObject(ctx context.Context, tenant, objectID string, objectType model.LabelableObject) (map[string]string, error)
	GetLabelsForObjects(ctx context.Context, tenant string, objectIDs []string, objectType model.LabelableObject) (map[string]map[string]string, error)
}

//go:generate mockery --exported --name=tenantInputBuilder --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantInputBuilder interface {
	GetTenantsForApplicationTemplates(ctx context.Context, tenant string, labels map[string]map[string]string, objectIDs []string) (map[string]*webhook.TenantWithLabels, error)
	GetTenantForApplicationTemplate(ctx context.Context, tenant string, labels map[string]string) (*webhook.TenantWithLabels, error)
	GetTenantsForObjects(ctx context.Context, tenant string, objectIDs []string, resourceType resource.Type) (map[string]*webhook.TenantWithLabels, error)
	GetTenantForObject(ctx context.Context, objectID string, resourceType resource.Type) (*webhook.TenantWithLabels, error)
}

//go:generate mockery --exported --name=certSubjectInputBuilder --output=automock --outpkg=automock --case=underscore --disable-version-string
type certSubjectInputBuilder interface {
	GetTrustDetailsForObjects(ctx context.Context, objectIDs []string) (map[string]*webhook.TrustDetails, error)
	GetTrustDetailsForObject(ctx context.Context, objectID string) (*webhook.TrustDetails, error)
}

// DataInputBuilder is responsible to prepare and build different entity data needed for a webhook input
//
//go:generate mockery --exported --name=DataInputBuilder --output=automock --outpkg=automock --case=underscore --disable-version-string
type DataInputBuilder interface {
	PrepareApplicationAndAppTemplateWithLabels(ctx context.Context, tenant, appID string) (*webhook.ApplicationWithLabels, *webhook.ApplicationTemplateWithLabels, error)
	PrepareRuntimeWithLabels(ctx context.Context, tenant, runtimeID string) (*webhook.RuntimeWithLabels, error)
	PrepareRuntimeContextWithLabels(ctx context.Context, tenant, runtimeCtxID string) (*webhook.RuntimeContextWithLabels, error)
	PrepareRuntimeAndRuntimeContextWithLabels(ctx context.Context, tenant, runtimeID string) (*webhook.RuntimeWithLabels, *webhook.RuntimeContextWithLabels, error)
	PrepareRuntimesAndRuntimeContextsMappingsInFormation(ctx context.Context, tenant string, scenario string) (map[string]*webhook.RuntimeWithLabels, map[string]*webhook.RuntimeContextWithLabels, error)
	PrepareApplicationMappingsInFormation(ctx context.Context, tenant string, scenario string) (map[string]*webhook.ApplicationWithLabels, map[string]*webhook.ApplicationTemplateWithLabels, error)
}

// WebhookDataInputBuilder take cares to get and build different webhook input data such as application, runtime, runtime contexts
type WebhookDataInputBuilder struct {
	applicationRepository         applicationRepository
	applicationTemplateRepository applicationTemplateRepository
	runtimeRepo                   runtimeRepository
	runtimeContextRepo            runtimeContextRepository
	labelInputBuilder             labelInputBuilder
	tenantInputBuilder            tenantInputBuilder
	certSubjectInputBuilder       certSubjectInputBuilder
}

const globalSubaccountIDLabelKey = "global_subaccount_id"

// NewWebhookDataInputBuilder creates a WebhookDataInputBuilder
func NewWebhookDataInputBuilder(applicationRepository applicationRepository, applicationTemplateRepository applicationTemplateRepository, runtimeRepo runtimeRepository, runtimeContextRepo runtimeContextRepository, labelInputBuilder labelInputBuilder, tenantInputBuilder tenantInputBuilder, certSubjectInputBuilder certSubjectInputBuilder) *WebhookDataInputBuilder {
	return &WebhookDataInputBuilder{
		applicationRepository:         applicationRepository,
		applicationTemplateRepository: applicationTemplateRepository,
		runtimeRepo:                   runtimeRepo,
		runtimeContextRepo:            runtimeContextRepo,
		labelInputBuilder:             labelInputBuilder,
		tenantInputBuilder:            tenantInputBuilder,
		certSubjectInputBuilder:       certSubjectInputBuilder,
	}
}

// PrepareApplicationAndAppTemplateWithLabels construct ApplicationWithLabels and ApplicationTemplateWithLabels based on tenant and ID
func (b *WebhookDataInputBuilder) PrepareApplicationAndAppTemplateWithLabels(ctx context.Context, tenant, appID string) (*webhook.ApplicationWithLabels, *webhook.ApplicationTemplateWithLabels, error) {
	application, err := b.applicationRepository.GetByID(ctx, tenant, appID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while getting application by ID: %q", appID)
	}
	applicationLabels, err := b.labelInputBuilder.GetLabelsForObject(ctx, tenant, appID, model.ApplicationLabelableObject)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while building labels for application with ID %q", appID)
	}
	tenantWithLabelsForApplication, err := b.tenantInputBuilder.GetTenantForObject(ctx, appID, resource.Application)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while building tenant with labels for application with ID %q", appID)
	}

	applicationWithLabels := &webhook.ApplicationWithLabels{
		Application: application,
		Labels:      applicationLabels,
		Tenant:      tenantWithLabelsForApplication,
	}

	var appTemplateWithLabels *webhook.ApplicationTemplateWithLabels
	if application.ApplicationTemplateID != nil {
		appTemplate, err := b.applicationTemplateRepository.Get(ctx, *application.ApplicationTemplateID)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "while getting application template with ID: %q", *application.ApplicationTemplateID)
		}
		applicationTemplateLabels, err := b.labelInputBuilder.GetLabelsForObject(ctx, tenant, appTemplate.ID, model.AppTemplateLabelableObject)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "while building labels for application template with ID %q", appTemplate.ID)
		}

		tenantWithLabelsForApplicationTemplate, err := b.tenantInputBuilder.GetTenantForApplicationTemplate(ctx, tenant, applicationTemplateLabels)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "while building tenant with labels for application template with ID %q", appTemplate.ID)
		}

		trustDetailsForApplicationTemplate, err := b.certSubjectInputBuilder.GetTrustDetailsForObject(ctx, appTemplate.ID)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "while building trust details for application tempalate with ID %q", appTemplate.ID)
		}

		appTemplateWithLabels = &webhook.ApplicationTemplateWithLabels{
			ApplicationTemplate: appTemplate,
			Labels:              applicationTemplateLabels,
			Tenant:              tenantWithLabelsForApplicationTemplate,
			TrustDetails:        trustDetailsForApplicationTemplate,
		}
	}
	return applicationWithLabels, appTemplateWithLabels, nil
}

// PrepareRuntimeWithLabels construct RuntimeWithLabels based on tenant and runtimeID
func (b *WebhookDataInputBuilder) PrepareRuntimeWithLabels(ctx context.Context, tenant, runtimeID string) (*webhook.RuntimeWithLabels, error) {
	runtime, err := b.runtimeRepo.GetByID(ctx, tenant, runtimeID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtime by ID: %q", runtimeID)
	}

	runtimeLabels, err := b.labelInputBuilder.GetLabelsForObject(ctx, tenant, runtimeID, model.RuntimeLabelableObject)
	if err != nil {
		return nil, errors.Wrapf(err, "while building labels for runtime with ID %q", runtimeID)
	}

	tenantWithLabelsForRuntime, err := b.tenantInputBuilder.GetTenantForObject(ctx, runtimeID, resource.Runtime)
	if err != nil {
		return nil, errors.Wrapf(err, "while building tenants with labels for runtime with ID %q", runtimeID)
	}

	trustDetailsForRuntime, err := b.certSubjectInputBuilder.GetTrustDetailsForObject(ctx, runtimeID)
	if err != nil {
		return nil, errors.Wrapf(err, "while building trust details for runtime with ID %q", runtimeID)
	}

	runtimeWithLabels := &webhook.RuntimeWithLabels{
		Runtime:      runtime,
		Labels:       runtimeLabels,
		Tenant:       tenantWithLabelsForRuntime,
		TrustDetails: trustDetailsForRuntime,
	}

	return runtimeWithLabels, nil
}

// PrepareRuntimeContextWithLabels construct RuntimeContextWithLabels based on tenant and runtimeCtxID
func (b *WebhookDataInputBuilder) PrepareRuntimeContextWithLabels(ctx context.Context, tenant, runtimeCtxID string) (*webhook.RuntimeContextWithLabels, error) {
	runtimeCtx, err := b.runtimeContextRepo.GetByID(ctx, tenant, runtimeCtxID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtime context by ID: %q", runtimeCtxID)
	}

	runtimeCtxLabels, err := b.labelInputBuilder.GetLabelsForObject(ctx, tenant, runtimeCtxID, model.RuntimeContextLabelableObject)
	if err != nil {
		return nil, errors.Wrapf(err, "while building labels for runtime context with ID %q", runtimeCtx.ID)
	}

	tenantWithLabelsForRuntimeContext, err := b.tenantInputBuilder.GetTenantForObject(ctx, runtimeCtxID, resource.RuntimeContext)
	if err != nil {
		return nil, errors.Wrapf(err, "while building tenant with labels for runtime context with ID %q", runtimeCtxID)
	}

	runtimeContextWithLabels := &webhook.RuntimeContextWithLabels{
		RuntimeContext: runtimeCtx,
		Labels:         runtimeCtxLabels,
		Tenant:         tenantWithLabelsForRuntimeContext,
	}

	return runtimeContextWithLabels, nil
}

// PrepareRuntimeAndRuntimeContextWithLabels construct RuntimeWithLabels and RuntimeContextWithLabels based on tenant and runtimeID
func (b *WebhookDataInputBuilder) PrepareRuntimeAndRuntimeContextWithLabels(ctx context.Context, tenant, runtimeID string) (*webhook.RuntimeWithLabels, *webhook.RuntimeContextWithLabels, error) {
	runtimeWithLabels, err := b.PrepareRuntimeWithLabels(ctx, tenant, runtimeID)
	if err != nil {
		return nil, nil, err
	}

	runtimeCtx, err := b.runtimeContextRepo.GetByRuntimeID(ctx, tenant, runtimeID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while getting runtime context for runtime with ID: %q", runtimeID)
	}

	runtimeCtxLabels, err := b.labelInputBuilder.GetLabelsForObject(ctx, tenant, runtimeCtx.ID, model.RuntimeContextLabelableObject)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while building labels for runtime context with ID %q", runtimeCtx.ID)
	}

	tenantWithLabelsForRuntimeContext, err := b.tenantInputBuilder.GetTenantForObject(ctx, runtimeCtx.ID, resource.RuntimeContext)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while building tenant with labels for runtime context with ID %q", runtimeCtx.ID)
	}

	runtimeContextWithLabels := &webhook.RuntimeContextWithLabels{
		RuntimeContext: runtimeCtx,
		Labels:         runtimeCtxLabels,
		Tenant:         tenantWithLabelsForRuntimeContext,
	}

	return runtimeWithLabels, runtimeContextWithLabels, nil
}

// PrepareRuntimesAndRuntimeContextsMappingsInFormation constructs:
// map from runtime ID to RuntimeWithLabels with entries for each runtime part of the formation and for each runtime whose child runtime context is part of the formation
// map from parent runtime ID to RuntimeContextWithLabels with entries for all runtime contexts part of the formation.
func (b *WebhookDataInputBuilder) PrepareRuntimesAndRuntimeContextsMappingsInFormation(ctx context.Context, tenant string, scenario string) (map[string]*webhook.RuntimeWithLabels, map[string]*webhook.RuntimeContextWithLabels, error) {
	runtimesInFormation, err := b.runtimeRepo.ListByScenarios(ctx, tenant, []string{scenario})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while listing runtimes in scenario %s", scenario)
	}

	runtimeContextsInFormation, err := b.runtimeContextRepo.ListByScenarios(ctx, tenant, []string{scenario})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while listing runtime contexts in scenario %s", scenario)
	}

	runtimeContextsIDs := make([]string, 0, len(runtimeContextsInFormation))
	parentRuntimeIDs := make([]string, 0, len(runtimeContextsInFormation))
	for _, rtCtx := range runtimeContextsInFormation {
		runtimeContextsIDs = append(runtimeContextsIDs, rtCtx.ID)
		parentRuntimeIDs = append(parentRuntimeIDs, rtCtx.RuntimeID)
	}

	// the parent runtime of the runtime context may not be in the formation - that's why we list them separately
	parentRuntimesOfRuntimeContextsInFormation, err := b.runtimeRepo.ListByIDs(ctx, tenant, parentRuntimeIDs)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while listing parent runtimes of runtime contexts in scenario %s", scenario)
	}

	runtimes := append(runtimesInFormation, parentRuntimesOfRuntimeContextsInFormation...)
	runtimesIDs := make([]string, 0, len(runtimes))
	for _, rt := range runtimes {
		runtimesIDs = append(runtimesIDs, rt.ID)
	}

	runtimesLabels, err := b.labelInputBuilder.GetLabelsForObjects(ctx, tenant, runtimesIDs, model.RuntimeLabelableObject)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while listing runtime labels")
	}

	tenantsWithLabelsForRuntimes, err := b.tenantInputBuilder.GetTenantsForObjects(ctx, tenant, runtimesIDs, resource.Runtime)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while building tenants with labels for runtimes")
	}

	trustDetailsForRuntimes, err := b.certSubjectInputBuilder.GetTrustDetailsForObjects(ctx, runtimesIDs)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while building trust details for runtimes with IDs %q", runtimesIDs)
	}

	runtimesMapping := make(map[string]*webhook.RuntimeWithLabels, len(runtimesLabels))
	for _, rt := range runtimes {
		runtimesMapping[rt.ID] = &webhook.RuntimeWithLabels{
			Runtime:      rt,
			Labels:       runtimesLabels[rt.ID],
			Tenant:       tenantsWithLabelsForRuntimes[rt.ID],
			TrustDetails: trustDetailsForRuntimes[rt.ID],
		}
	}

	runtimeContextsLabels, err := b.labelInputBuilder.GetLabelsForObjects(ctx, tenant, runtimeContextsIDs, model.RuntimeContextLabelableObject)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while listing labels for runtime contexts")
	}

	tenantsWithLabelsForRuntimeContexts, err := b.tenantInputBuilder.GetTenantsForObjects(ctx, tenant, runtimeContextsIDs, resource.RuntimeContext)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while building tenants with labels for runtime contexts")
	}

	runtimesToRuntimeContextsMapping := make(map[string]*webhook.RuntimeContextWithLabels, len(runtimeContextsInFormation))
	for _, rtCtx := range runtimeContextsInFormation {
		runtimesToRuntimeContextsMapping[rtCtx.RuntimeID] = &webhook.RuntimeContextWithLabels{
			RuntimeContext: rtCtx,
			Labels:         runtimeContextsLabels[rtCtx.ID],
			Tenant:         tenantsWithLabelsForRuntimeContexts[rtCtx.ID],
		}
	}

	return runtimesMapping, runtimesToRuntimeContextsMapping, nil
}

// PrepareApplicationMappingsInFormation constructs:
// map from application ID to ApplicationWithLabels with entries for each application part of the formation
// map from applicationTemplate ID to ApplicationTemplateWithLabels with entries for each application template whose child application is part of the formation
func (b *WebhookDataInputBuilder) PrepareApplicationMappingsInFormation(ctx context.Context, tenant string, scenario string) (map[string]*webhook.ApplicationWithLabels, map[string]*webhook.ApplicationTemplateWithLabels, error) {
	applicationsToBeNotifiedFor, err := b.applicationRepository.ListByScenariosNoPaging(ctx, tenant, []string{scenario})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while listing applications in formation %s", scenario)
	}

	if len(applicationsToBeNotifiedFor) == 0 {
		log.C(ctx).Infof("There are no applications in scenario %s.", scenario)
		return make(map[string]*webhook.ApplicationWithLabels, 0), make(map[string]*webhook.ApplicationTemplateWithLabels, 0), nil
	}

	applicationsToBeNotifiedForIDs := make([]string, 0, len(applicationsToBeNotifiedFor))
	applicationsTemplateIDs := make([]string, 0, len(applicationsToBeNotifiedFor))
	for _, app := range applicationsToBeNotifiedFor {
		applicationsToBeNotifiedForIDs = append(applicationsToBeNotifiedForIDs, app.ID)
		if app.ApplicationTemplateID != nil {
			applicationsTemplateIDs = append(applicationsTemplateIDs, *app.ApplicationTemplateID)
		}
	}

	applicationsToBeNotifiedForLabels, err := b.labelInputBuilder.GetLabelsForObjects(ctx, tenant, applicationsToBeNotifiedForIDs, model.ApplicationLabelableObject)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while listing labels for applications")
	}

	tenantsWithLabelsForApplication, err := b.tenantInputBuilder.GetTenantsForObjects(ctx, tenant, applicationsToBeNotifiedForIDs, resource.Application)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while building tenants with labels for applications")
	}

	applicationMapping := make(map[string]*webhook.ApplicationWithLabels, len(applicationsToBeNotifiedForIDs))
	for i, app := range applicationsToBeNotifiedFor {
		applicationMapping[app.ID] = &webhook.ApplicationWithLabels{
			Application: applicationsToBeNotifiedFor[i],
			Labels:      applicationsToBeNotifiedForLabels[app.ID],
			Tenant:      tenantsWithLabelsForApplication[app.ID],
		}
	}

	applicationTemplates, err := b.applicationTemplateRepository.ListByIDs(ctx, applicationsTemplateIDs)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while listing application templates")
	}

	applicationTemplatesLabels, err := b.labelInputBuilder.GetLabelsForObjects(ctx, tenant, applicationsTemplateIDs, model.AppTemplateLabelableObject)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while listing labels for application templates")
	}

	tenantsWithLabelsForApplicationTemplates, err := b.tenantInputBuilder.GetTenantsForApplicationTemplates(ctx, tenant, applicationTemplatesLabels, applicationsTemplateIDs)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while building tenants with labels for application templates")
	}

	trustDetailsForApplicationTemplates, err := b.certSubjectInputBuilder.GetTrustDetailsForObjects(ctx, applicationsTemplateIDs)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while building trust details for application tempalates with IDs %q", applicationsTemplateIDs)
	}

	applicationTemplatesMapping := make(map[string]*webhook.ApplicationTemplateWithLabels, len(applicationTemplates))
	for i, appTemplate := range applicationTemplates {
		applicationTemplatesMapping[appTemplate.ID] = &webhook.ApplicationTemplateWithLabels{
			ApplicationTemplate: applicationTemplates[i],
			Labels:              applicationTemplatesLabels[appTemplate.ID],
			Tenant:              tenantsWithLabelsForApplicationTemplates[appTemplate.ID],
			TrustDetails:        trustDetailsForApplicationTemplates[appTemplate.ID],
		}
	}

	return applicationMapping, applicationTemplatesMapping, nil
}
