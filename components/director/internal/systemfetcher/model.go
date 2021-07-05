package systemfetcher

import (
	"encoding/json"

	"github.com/tidwall/gjson"
)

var (
	// Mappings global static configuration which is set after reading the configuration during startup, should only be used for the unmarshaling of system data
	// Template mappings describe what properties and their values should be in order to map to a certain application template ID
	// If there are multiple keys and values, all of them should match in order for the mapping to be succesful
	Mappings []TemplateMapping
)

type TemplateMapping struct {
	Name        string
	ID          string
	SourceKey   []string
	SourceValue []string
}

type AdditionalUrls map[string]string

type AdditionalAttributes map[string]string

type SystemBase struct {
	SystemNumber           string               `json:"systemNumber"`
	DisplayName            string               `json:"displayName"`
	ProductDescription     string               `json:"productDescription"`
	BaseURL                string               `json:"baseUrl"`
	InfrastructureProvider string               `json:"infrastructureProvider"`
	AdditionalUrls         AdditionalUrls       `json:"additionalUrls"`
	AdditionalAttributes   AdditionalAttributes `json:"additionalAttributes"`
}

type System struct {
	SystemBase
	TemplateID string `json:"-"`
}

func (s *System) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &s.SystemBase); err != nil {
		return err
	}

	for _, tm := range Mappings {
		mapped := true
		for i, sk := range tm.SourceKey {
			v := gjson.GetBytes(data, sk).String()
			if v != tm.SourceValue[i] {
				mapped = false
				break
			}
		}
		if mapped {
			s.TemplateID = tm.ID
		}
	}

	return nil
}
