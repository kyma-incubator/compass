package tenantfetcher_test

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
)

func fixTenantEventsResponse(events []byte, total, pages int) tenantfetcher.TenantEventsResponse {
	return tenantfetcher.TenantEventsResponse(fmt.Sprintf(`{
		"events":       %s,
		"total": %d,
		"pages":   %d,
	}`, string(events), total, pages))
}

func fixEvent(fields map[string]string) []byte {
	eventEntries := ""
	fieldsCount := len(fields)
	counter := 0

	for fieldName, fieldValue := range fields {
		counter++
		sign := ","
		if counter == fieldsCount {
			sign = ""
		}
		eventEntries = fmt.Sprintf(`%s "%s":"%s"%s`, eventEntries, fieldName, fieldValue, sign)
	}
	eventData := fmt.Sprintf("{%s}", eventEntries)

	return wrapIntoEventPageJSON(eventData)
}

func fixMovedRuntimeByLabelEvent(id, source, target string, fieldMapping tenantfetcher.MovedRuntimeByLabelFieldMapping) []byte {
	eventData := fmt.Sprintf(`{"%s":"%s","%s":"%s","%s":"%s"}`, fieldMapping.LabelValue, id, fieldMapping.SourceTenant, source, fieldMapping.TargetTenant, target)

	return wrapIntoEventPageJSON(eventData)
}

func fixEventWithDiscriminator(id, name, discriminator string, fieldMapping tenantfetcher.TenantFieldMapping) []byte {
	discriminatorData := ""
	if fieldMapping.DiscriminatorField != "" {
		discriminatorData = fmt.Sprintf(`"%s": "%s",`, fieldMapping.DiscriminatorField, discriminator)
	}

	eventData := fmt.Sprintf(`{"%s":"%s",%s"%s":"%s"}`, fieldMapping.IDField, id, discriminatorData, fieldMapping.NameField, name)

	return wrapIntoEventPageJSON(eventData)
}

func wrapIntoEventPageJSON(eventData string) []byte {
	return []byte(fmt.Sprintf(`{
		"id":        %s,
		"eventData": %s,
	}`, fixID(), eventData))
}

func fixBusinessTenantMappingInput(name, externalTenant, provider string, parent string, subdomain string, tenantType tenant.Type) model.BusinessTenantMappingInput {
	return model.BusinessTenantMappingInput{
		Name:           name,
		ExternalTenant: externalTenant,
		Provider:       provider,
		Parent:         parent,
		Subdomain:      subdomain,
		Type:           tenant.TypeToStr(tenantType),
	}
}

func fixID() string {
	return uuid.New().String()
}
