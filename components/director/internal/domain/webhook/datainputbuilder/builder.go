package datainputbuilder

import (
	"context"

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
	GetByRuntimeID(ctx context.Context, tenant, runtimeID string) (*model.RuntimeContext, error)
}

//go:generate mockery --exported --name=labelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelRepository interface {
	ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
}

// DataInputBuilder is responsible to prepare and build different entity data needed for a webhook input
//go:generate mockery --exported --name=DataInputBuilder --output=automock --outpkg=automock --case=underscore --disable-version-string
type DataInputBuilder interface {
	PrepareApplicationAndAppTemplateWithLabels(ctx context.Context, tenant, appID string) (*webhook.ApplicationWithLabels, *webhook.ApplicationTemplateWithLabels, error)
	PrepareRuntimeWithLabels(ctx context.Context, tenant, runtimeID string) (*webhook.RuntimeWithLabels, error)
	PrepareRuntimeContextWithLabels(ctx context.Context, tenant, runtimeCtxID string) (*webhook.RuntimeContextWithLabels, error)
	PrepareRuntimeAndRuntimeContextWithLabels(ctx context.Context, tenant, runtimeID string) (*webhook.RuntimeWithLabels, *webhook.RuntimeContextWithLabels, error)
}

// WebhookDataInputBuilder take cares to get and build different webhook input data such as application, runtime, runtime contexts
type WebhookDataInputBuilder struct {
	applicationRepository         applicationRepository
	applicationTemplateRepository applicationTemplateRepository
	runtimeRepo                   runtimeRepository
	runtimeContextRepo            runtimeContextRepository
	labelRepository               labelRepository
}

// NewWebhookDataInputBuilder creates a WebhookDataInputBuilder
func NewWebhookDataInputBuilder(applicationRepository applicationRepository, applicationTemplateRepository applicationTemplateRepository, runtimeRepo runtimeRepository, runtimeContextRepo runtimeContextRepository, labelRepository labelRepository) *WebhookDataInputBuilder {
	return &WebhookDataInputBuilder{
		applicationRepository:         applicationRepository,
		applicationTemplateRepository: applicationTemplateRepository,
		runtimeRepo:                   runtimeRepo,
		runtimeContextRepo:            runtimeContextRepo,
		labelRepository:               labelRepository,
	}
}

// PrepareApplicationAndAppTemplateWithLabels construct ApplicationWithLabels and ApplicationTemplateWithLabels based on tenant and ID
func (b *WebhookDataInputBuilder) PrepareApplicationAndAppTemplateWithLabels(ctx context.Context, tenant, appID string) (*webhook.ApplicationWithLabels, *webhook.ApplicationTemplateWithLabels, error) {
	application, err := b.applicationRepository.GetByID(ctx, tenant, appID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while getting application by ID: %q", appID)
	}
	applicationLabels, err := b.getLabelsForObject(ctx, tenant, appID, model.ApplicationLabelableObject)
	if err != nil {
		return nil, nil, err
	}
	applicationWithLabels := &webhook.ApplicationWithLabels{
		Application: application,
		Labels:      applicationLabels,
	}

	var appTemplateWithLabels *webhook.ApplicationTemplateWithLabels
	if application.ApplicationTemplateID != nil {
		appTemplate, err := b.applicationTemplateRepository.Get(ctx, *application.ApplicationTemplateID)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "while getting application template with ID: %q", *application.ApplicationTemplateID)
		}
		applicationTemplateLabels, err := b.getLabelsForObject(ctx, tenant, appTemplate.ID, model.AppTemplateLabelableObject)
		if err != nil {
			return nil, nil, err
		}
		appTemplateWithLabels = &webhook.ApplicationTemplateWithLabels{
			ApplicationTemplate: appTemplate,
			Labels:              applicationTemplateLabels,
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

	runtimeLabels, err := b.getLabelsForObject(ctx, tenant, runtimeID, model.RuntimeLabelableObject)
	if err != nil {
		return nil, err
	}

	runtimeWithLabels := &webhook.RuntimeWithLabels{
		Runtime: runtime,
		Labels:  runtimeLabels,
	}

	return runtimeWithLabels, nil
}

// PrepareRuntimeContextWithLabels construct RuntimeContextWithLabels based on tenant and runtimeCtxID
func (b *WebhookDataInputBuilder) PrepareRuntimeContextWithLabels(ctx context.Context, tenant, runtimeCtxID string) (*webhook.RuntimeContextWithLabels, error) {
	runtimeCtx, err := b.runtimeContextRepo.GetByID(ctx, tenant, runtimeCtxID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtime context by ID: %q", runtimeCtxID)
	}

	runtimeCtxLabels, err := b.getLabelsForObject(ctx, tenant, runtimeCtx.ID, model.RuntimeContextLabelableObject)
	if err != nil {
		return nil, err
	}

	runtimeContextWithLabels := &webhook.RuntimeContextWithLabels{
		RuntimeContext: runtimeCtx,
		Labels:         runtimeCtxLabels,
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

	runtimeCtxLabels, err := b.getLabelsForObject(ctx, tenant, runtimeCtx.ID, model.RuntimeContextLabelableObject)
	if err != nil {
		return nil, nil, err
	}

	runtimeContextWithLabels := &webhook.RuntimeContextWithLabels{
		RuntimeContext: runtimeCtx,
		Labels:         runtimeCtxLabels,
	}

	return runtimeWithLabels, runtimeContextWithLabels, nil
}

func (b *WebhookDataInputBuilder) getLabelsForObject(ctx context.Context, tenant, objectID string, objectType model.LabelableObject) (map[string]interface{}, error) {
	labels, err := b.labelRepository.ListForObject(ctx, tenant, objectType, objectID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing labels for %q with ID: %q", objectType, objectID)
	}
	labelsMap := make(map[string]interface{}, len(labels))
	for _, l := range labels {
		labelsMap[l.Key] = l.Value
	}
	return labelsMap, nil
}
