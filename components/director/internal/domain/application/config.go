package application

import (
	"encoding/json"
)

// ORDWebhookMapping missing godoc
type ORDWebhookMapping struct {
	Type                string   `json:"Type"`
	PpmsProductVersions []string `json:"PpmsProductVersions"`
	OrdURLPath          string   `json:"OrdUrlPath"`
	SubdomainSuffix     string   `json:"SubdomainSuffix"`
}

func UnmarshalMappings(mappingsConfig string) ([]ORDWebhookMapping, error) {
	var mappings []ORDWebhookMapping
	if err := json.Unmarshal([]byte(mappingsConfig), &mappings); err != nil {
		return nil, err
	}

	return mappings, nil
}
