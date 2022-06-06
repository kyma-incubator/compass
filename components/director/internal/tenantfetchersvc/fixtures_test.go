package tenantfetcher_test

import (
	"encoding/json"
	"fmt"

	"github.com/stretchr/testify/require"

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

func fixEvent(t require.TestingT, eventType, ga string, fields map[string]string) []byte {
	eventData, err := json.Marshal(fields)
	if err != nil {
		require.NoError(t, err)
	}
	return wrapIntoEventPageJSON(string(eventData), eventType, ga)
}

func wrapIntoEventPageJSON(eventData, eventType, ga string) []byte {
	return []byte(fmt.Sprintf(`{
		"id":        "%s",
		"type": "%s",
        "globalAccountGUID": "%s",
		"eventData": %s
	}`, fixID(), eventType, ga, eventData))
}

func fixBusinessTenantMappingInput(name, externalTenant, provider, subdomain, region, parent string, tenantType tenant.Type) model.BusinessTenantMappingInput {
	return model.BusinessTenantMappingInput{
		Name:           name,
		ExternalTenant: externalTenant,
		Provider:       provider,
		Subdomain:      subdomain,
		Region:         region,
		Parent:         parent,
		Type:           tenant.TypeToStr(tenantType),
	}
}

func fixID() string {
	return uuid.New().String()
}
