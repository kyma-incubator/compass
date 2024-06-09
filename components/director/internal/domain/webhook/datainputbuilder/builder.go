package datainputbuilder

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/webhook"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=applicationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.Application, error)
}

//go:generate mockery --exported --name=applicationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationTemplateRepository interface {
	Get(ctx context.Context, id string) (*model.ApplicationTemplate, error)
}

//go:generate mockery --exported --name=runtimeRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.Runtime, error)
}

//go:generate mockery --exported --name=runtimeContextRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeContextRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.RuntimeContext, error)
}

//go:generate mockery --exported --name=labelInputBuilder --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelInputBuilder interface {
	GetLabelsForObject(ctx context.Context, tenant, objectID string, objectType model.LabelableObject) (map[string]string, error)
	GetLabelsForObjects(ctx context.Context, tenant string, objectIDs []string, objectType model.LabelableObject) (map[string]map[string]string, error)
}

//go:generate mockery --exported --name=tenantInputBuilder --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantInputBuilder interface {
	GetTenantForApplicationTemplate(ctx context.Context, tenant string, labels map[string]string) (*webhook.TenantWithLabels, error)
	GetTenantForObject(ctx context.Context, objectID string, resourceType resource.Type) (*webhook.TenantWithLabels, error)
}

//go:generate mockery --exported --name=certSubjectInputBuilder --output=automock --outpkg=automock --case=underscore --disable-version-string
type certSubjectInputBuilder interface {
	GetTrustDetailsForObject(ctx context.Context, objectID string) (*webhook.TrustDetails, error)
}

// DataInputBuilder is responsible to prepare and build different entity data needed for a webhook input
//
//go:generate mockery --exported --name=DataInputBuilder --output=automock --outpkg=automock --case=underscore --disable-version-string
type DataInputBuilder interface {
	PrepareApplicationAndAppTemplateWithLabels(ctx context.Context, tenant, appID string) (*webhook.ApplicationWithLabels, *webhook.ApplicationTemplateWithLabels, error)
	PrepareRuntimeWithLabels(ctx context.Context, tenant, runtimeID string) (*webhook.RuntimeWithLabels, error)
	PrepareRuntimeContextWithLabels(ctx context.Context, tenant, runtimeCtxID string) (*webhook.RuntimeContextWithLabels, error)
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
