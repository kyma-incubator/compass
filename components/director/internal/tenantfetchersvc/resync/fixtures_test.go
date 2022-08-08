package resync_test

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/resync"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

func fixTenantEventsResponse(events []byte, total, pages int) resync.EventsPage {
	response := fmt.Sprintf(`{
		"events":       %s,
		"total": %d,
		"pages":   %d,
	}`, string(events), total, pages)
	return resync.EventsPage{
		FieldMapping:                 resync.TenantFieldMapping{},
		MovedSubaccountsFieldMapping: resync.MovedSubaccountsFieldMapping{},
		ProviderName:                 "",
		Payload:                      []byte(response),
	}

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

func fixBusinessTenantMappingInput(externalTenant, provider, subdomain, region, parent string, tenantType tenant.Type) model.BusinessTenantMappingInput {
	return model.BusinessTenantMappingInput{
		Name:           externalTenant,
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

func eventsToJSONArray(events ...[]byte) []byte {
	return []byte(fmt.Sprintf(`[%s]`, bytes.Join(events, []byte(","))))
}
