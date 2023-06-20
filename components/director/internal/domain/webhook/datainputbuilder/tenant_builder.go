package datainputbuilder

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=tenantRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantRepository interface {
	GetLowestOwnerForResource(ctx context.Context, resourceType resource.Type, objectID string) (string, error)
	GetByExternalTenant(ctx context.Context, externalTenant string) (*model.BusinessTenantMapping, error)
	Get(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

type WebhookTenantBuilder struct {
	labelInputBuilder labelInputBuilder
	tenantRepository  tenantRepository
}

// NewWebhookTenantBuilder creates a WebhookTenantBuilder
func NewWebhookTenantBuilder(labelInputBuilder labelInputBuilder, tenantRepository tenantRepository) *WebhookTenantBuilder {
	return &WebhookTenantBuilder{
		labelInputBuilder: labelInputBuilder,
		tenantRepository:  tenantRepository,
	}
}

func (b *WebhookTenantBuilder) GetTenantForApplicationTemplates(ctx context.Context, tenant string, labels map[string]map[string]string, objectIDs []string) (map[string]*webhook.TenantWithLabels, error) {
	tenantsForObjects := make(map[string]*webhook.TenantWithLabels)
	tenantIDs := make([]string, 0)
	objectTenantMapping := make(map[string]string)
	for _, objectId := range objectIDs {
		if subaccountId, ok := labels[objectId][globalSubaccountIDLabelKey]; ok {
			tenantModel, err := b.tenantRepository.GetByExternalTenant(ctx, subaccountId)
			if err != nil {
				return nil, errors.Wrapf(err, "while getting tenant by external id %q", subaccountId)
			}

			tenantsForObjects[objectId] = &webhook.TenantWithLabels{
				BusinessTenantMapping: tenantModel,
				Labels:                nil,
			}
			tenantIDs = append(tenantIDs, subaccountId)
			objectTenantMapping[objectId] = subaccountId
		}
	}

	tenantLabels, err := b.labelInputBuilder.GetLabelsForObjects(ctx, tenant, tenantIDs, model.TenantLabelableObject)
	if err != nil {
		return nil, errors.Wrap(err, "while listing tenant labels")
	}
	for objectID, tenantID := range objectTenantMapping {
		tenantsForObjects[objectID].Labels = tenantLabels[tenantID]
	}
	return tenantsForObjects, nil
}

func (b *WebhookTenantBuilder) GetTenantForApplicationTemplate(ctx context.Context, tenant string, labels map[string]string) (*webhook.TenantWithLabels, error) {
	if subaccountId, ok := labels[globalSubaccountIDLabelKey]; ok {
		tenantModel, err := b.tenantRepository.GetByExternalTenant(ctx, subaccountId)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting tenant by external id %q", subaccountId)
		}

		tenantLabels, err := b.labelInputBuilder.GetLabelsForObject(ctx, tenant, subaccountId, model.TenantLabelableObject)
		if err != nil {
			return nil, errors.Wrap(err, "while listing tenant labels")
		}

		return &webhook.TenantWithLabels{
			BusinessTenantMapping: tenantModel,
			Labels:                tenantLabels,
		}, nil
	}

	return nil, nil
}

func (b *WebhookTenantBuilder) GetTenantForObjects(ctx context.Context, tenant string, objectIDs []string, resourceType resource.Type) (map[string]*webhook.TenantWithLabels, error) {
	tenantsForObjects := make(map[string]*webhook.TenantWithLabels)
	tenantIDs := make([]string, 0)
	objectTenantMapping := make(map[string]string)
	for _, objectId := range objectIDs {
		tenantId, err := b.tenantRepository.GetLowestOwnerForResource(ctx, resourceType, objectId)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting tenant for object with id %q", tenantId)
		}
		tenantModel, err := b.tenantRepository.Get(ctx, tenantId)

		tenantsForObjects[objectId] = &webhook.TenantWithLabels{
			BusinessTenantMapping: tenantModel,
			Labels:                nil,
		}
		tenantIDs = append(tenantIDs, tenantId)
		objectTenantMapping[objectId] = tenantId
	}

	tenantLabels, err := b.labelInputBuilder.GetLabelsForObjects(ctx, tenant, tenantIDs, model.TenantLabelableObject)
	if err != nil {
		return nil, errors.Wrap(err, "while listing tenant labels")
	}
	for _, objectID := range objectIDs {
		tenantsForObjects[objectID].Labels = tenantLabels[objectTenantMapping[objectID]]
	}

	return tenantsForObjects, nil
}

func (b *WebhookTenantBuilder) GetTenantForObject(ctx context.Context, objectID string, resourceType resource.Type) (*webhook.TenantWithLabels, error) {
	tenantId, err := b.tenantRepository.GetLowestOwnerForResource(ctx, resourceType, objectID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting tenant for object with id %q", tenantId)
	}
	tenantModel, err := b.tenantRepository.Get(ctx, tenantId)

	tenantLabels, err := b.labelInputBuilder.GetLabelsForObject(ctx, tenantId, tenantId, model.TenantLabelableObject)
	if err != nil {
		return nil, errors.Wrap(err, "while listing tenant labels")
	}

	return &webhook.TenantWithLabels{
		BusinessTenantMapping: tenantModel,
		Labels:                tenantLabels,
	}, nil
}
