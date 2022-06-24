package appmetadatavalidation_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

func fixBusinessTenantMappingModel(tenantID, externalTenantID string) *model.BusinessTenantMapping {
	return &model.BusinessTenantMapping{
		ID:             tenantID,
		Name:           "externalTnt",
		ExternalTenant: externalTenantID,
		Initialized:    nil,
	}
}

func fixTenantLabel(tenantID, regionLabelValue string) *model.Label {
	return &model.Label{
		ID:         "12344",
		Tenant:     &tenantID,
		Key:        tenant.RegionLabelKey,
		Value:      regionLabelValue,
		ObjectID:   tenantID,
		ObjectType: model.TenantLabelableObject,
	}
}
