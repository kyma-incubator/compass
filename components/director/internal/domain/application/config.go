package application

import (
	"encoding/json"
)

// ORDWebhookMapping represents a struct for ORD Webhook Mappings
type ORDWebhookMapping struct {
	Type                string   `json:"Type"`
	PpmsProductVersions []string `json:"PpmsProductVersions"`
	OrdURLPath          string   `json:"OrdUrlPath"`
	SubdomainSuffix     string   `json:"SubdomainSuffix"`
}

// UnmarshalMappings unmarshalls a string into []ORDWebhookMapping. This is done because of limitation of the envconfig library
func UnmarshalMappings(mappingsConfig string) ([]ORDWebhookMapping, error) {
	var mappings []ORDWebhookMapping
	if err := json.Unmarshal([]byte(mappingsConfig), &mappings); err != nil {
		return nil, err
	}

	return mappings, nil
}
