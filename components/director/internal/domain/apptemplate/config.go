package apptemplate

import (
	"encoding/json"
)

// UnmarshalTenantMappingConfig unmarshalls a string into map[string]interface{}
func UnmarshalTenantMappingConfig(tenantMappingWebhook string) (map[string]interface{}, error) {
	var tenantMappingConfig map[string]interface{}
	if err := json.Unmarshal([]byte(tenantMappingWebhook), &tenantMappingConfig); err != nil {
		return nil, err
	}

	return tenantMappingConfig, nil
}
