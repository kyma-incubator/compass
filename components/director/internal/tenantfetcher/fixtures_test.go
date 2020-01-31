package tenantfetcher_test

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
)

func fixEvent(id, name string, fieldMapping tenantfetcher.TenantFieldMapping) tenantfetcher.Event {
	eventData := fmt.Sprintf(`{"%s":"%s","%s":"%s"}`, fieldMapping.IDField, id, fieldMapping.NameField, name)

	return tenantfetcher.Event{
		"id":        fixID(),
		"eventData": eventData,
	}
}

func fixEventWithDiscriminator(id, name, discriminator string, fieldMapping tenantfetcher.TenantFieldMapping) tenantfetcher.Event {
	discriminatorData := ""
	if fieldMapping.DiscriminatorField != "" {
		discriminatorData = fmt.Sprintf(`"%s": "%s",`, fieldMapping.DiscriminatorField, discriminator)
	}

	eventData := fmt.Sprintf(`{"%s":"%s",%s"%s":"%s"}`, fieldMapping.IDField, id, discriminatorData, fieldMapping.NameField, name)

	return tenantfetcher.Event{
		"id":        fixID(),
		"eventData": eventData,
	}
}

func fixBusinessTenantMappingInput(name, externalTenant, provider string) model.BusinessTenantMappingInput {
	return model.BusinessTenantMappingInput{
		Name:           name,
		ExternalTenant: externalTenant,
		Provider:       provider,
	}
}

func fixTenantEventsResponse(events []tenantfetcher.Event, total, pages int) *tenantfetcher.TenantEventsResponse {
	return &tenantfetcher.TenantEventsResponse{
		Events:       events,
		TotalResults: total,
		TotalPages:   pages,
	}
}

func fixID() string {
	return uuid.New().String()
}
