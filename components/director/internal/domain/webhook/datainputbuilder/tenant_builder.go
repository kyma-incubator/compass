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

// WebhookTenantBuilder takes care to get and build tenants and their labels for objects
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

// GetTenantForApplicationTemplates builds tenants with labels for application templates
func (b *WebhookTenantBuilder) GetTenantForApplicationTemplates(ctx context.Context, tenant string, labels map[string]map[string]string, objectIDs []string) (map[string]*webhook.TenantWithLabels, error) {
	tenantsForObjects := make(map[string]*webhook.TenantWithLabels)
	tenantIDs := make([]string, 0)
	tenantIDsMap := make(map[string]*model.BusinessTenantMapping)
	objectTenantMapping := make(map[string]string)
	for _, objectID := range objectIDs {
		if subaccountID, ok := labels[objectID][globalSubaccountIDLabelKey]; ok {
			if tenantModel, ok := tenantIDsMap[subaccountID]; ok {
				tenantsForObjects[objectID] = &webhook.TenantWithLabels{
					BusinessTenantMapping: tenantModel,
					Labels:                nil,
				}
				objectTenantMapping[objectID] = tenantModel.ID
				continue
			}
			tenantModel, err := b.tenantRepository.GetByExternalTenant(ctx, subaccountID)
			if err != nil {
				return nil, errors.Wrapf(err, "while getting tenant by external ID %q", subaccountID)
			}

			tenantsForObjects[objectID] = &webhook.TenantWithLabels{
				BusinessTenantMapping: tenantModel,
				Labels:                nil,
			}
			tenantIDs = append(tenantIDs, tenantModel.ID)
			tenantIDsMap[tenantModel.ID] = tenantModel
			objectTenantMapping[objectID] = tenantModel.ID
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

// GetTenantForApplicationTemplate builds tenant with labels for application template
func (b *WebhookTenantBuilder) GetTenantForApplicationTemplate(ctx context.Context, tenant string, labels map[string]string) (*webhook.TenantWithLabels, error) {
	if subaccountID, ok := labels[globalSubaccountIDLabelKey]; ok {
		tenantModel, err := b.tenantRepository.GetByExternalTenant(ctx, subaccountID)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting tenant by external ID %q", subaccountID)
		}

		tenantLabels, err := b.labelInputBuilder.GetLabelsForObject(ctx, tenant, subaccountID, model.TenantLabelableObject)
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

// GetTenantForObjects builds tenants with labels for objects of type runtime, runtime context or application
func (b *WebhookTenantBuilder) GetTenantForObjects(ctx context.Context, tenant string, objectIDs []string, resourceType resource.Type) (map[string]*webhook.TenantWithLabels, error) {
	tenantsForObjects := make(map[string]*webhook.TenantWithLabels)
	tenantIDs := make([]string, 0)
	tenantIDsMap := make(map[string]*model.BusinessTenantMapping)
	objectTenantMapping := make(map[string]string)
	for _, objectID := range objectIDs {
		tenantID, err := b.tenantRepository.GetLowestOwnerForResource(ctx, resourceType, objectID)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting tenant for object with ID %q", objectID)
		}
		objectTenantMapping[objectID] = tenantID

		// Check if we have loaded the tenant already, if so, no need for extra queries
		if tenantModel, ok := tenantIDsMap[tenantID]; ok {
			tenantsForObjects[objectID] = &webhook.TenantWithLabels{
				BusinessTenantMapping: tenantModel,
				Labels:                nil,
			}
			continue
		}

		tenantModel, err := b.tenantRepository.Get(ctx, tenantID)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting tenant with ID %q", tenantID)
		}

		tenantsForObjects[objectID] = &webhook.TenantWithLabels{
			BusinessTenantMapping: tenantModel,
			Labels:                nil,
		}
		tenantIDs = append(tenantIDs, tenantID)
		tenantIDsMap[tenantID] = tenantModel
	}

	tenantLabels, err := b.labelInputBuilder.GetLabelsForObjects(ctx, tenant, tenantIDs, model.TenantLabelableObject)
	if err != nil {
		return nil, errors.Wrap(err, "while building tenant labels")
	}
	for _, objectID := range objectIDs {
		tenantsForObjects[objectID].Labels = tenantLabels[objectTenantMapping[objectID]]
	}

	return tenantsForObjects, nil
}

// GetTenantForObject builds tenant with labels for object of type runtime, runtime context or application
func (b *WebhookTenantBuilder) GetTenantForObject(ctx context.Context, objectID string, resourceType resource.Type) (*webhook.TenantWithLabels, error) {
	tenantID, err := b.tenantRepository.GetLowestOwnerForResource(ctx, resourceType, objectID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting tenant lowest owner for object with id %q", objectID)
	}
	tenantModel, err := b.tenantRepository.Get(ctx, tenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting tenant for object with id %q", tenantID)
	}

	tenantLabels, err := b.labelInputBuilder.GetLabelsForObject(ctx, tenantID, tenantID, model.TenantLabelableObject)
	if err != nil {
		return nil, errors.Wrap(err, "while listing tenant labels")
	}

	return &webhook.TenantWithLabels{
		BusinessTenantMapping: tenantModel,
		Labels:                tenantLabels,
	}, nil
}
