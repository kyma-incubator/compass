package systemfetcher

import (
	"encoding/json"

	"github.com/tidwall/gjson"
)

var (
	// Mappings global static configuration which is set after reading the configuration during startup, should only be used for the unmarshaling of system data
	// Template mappings describe what properties and their values should be in order to map to a certain application template ID
	// If there are multiple keys and values, all of them should match in order for the mapping to be successful
	Mappings []TemplateMapping
)

// TemplateMapping missing godoc
type TemplateMapping struct {
	Name        string
	Region      string
	ID          string
	SourceKey   []string
	SourceValue []string
	OrdReady    bool
}

// AdditionalURLs missing godoc
type AdditionalURLs map[string]string

// AdditionalAttributes missing godoc
type AdditionalAttributes map[string]string

// SystemBase missing godoc
type SystemBase struct {
	SystemNumber           string               `json:"systemNumber"`
	DisplayName            string               `json:"displayName"`
	ProductID              string               `json:"productId"`
	PpmsProductVersionID   string               `json:"ppmsProductVersionId"`
	ProductDescription     string               `json:"productDescription"`
	BaseURL                string               `json:"baseUrl"`
	InfrastructureProvider string               `json:"infrastructureProvider"`
	DataCenterID           string               `json:"regionId"`
	AdditionalURLs         AdditionalURLs       `json:"additionalUrls"`
	AdditionalAttributes   AdditionalAttributes `json:"additionalAttributes"`
}

// System missing godoc
type System struct {
	SystemBase
	TemplateID string `json:"-"`
}

// UnmarshalJSON missing godoc
func (s *System) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &s.SystemBase); err != nil {
		return err
	}

	for _, tm := range Mappings {
		if matchProps(data, tm) {
			s.TemplateID = tm.ID
			return nil
		}
	}

	return nil
}

func matchProps(data []byte, tm TemplateMapping) bool {
	for i, sk := range tm.SourceKey {
		v := gjson.GetBytes(data, sk).String()
		if v != tm.SourceValue[i] {
			return false
		}
	}
	return true
}
