package systemfetcher

import (
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/tidwall/gjson"
)

var (
	// ApplicationTemplates global static configuration which is set after reading the configuration during startup, should only be used for the unmarshaling of system data
	// It represents a model.ApplicationTemplate with its labels in the form of map[string]*model.Label
	ApplicationTemplates []TemplateMapping
	// ApplicationTemplateLabelFilter represent a label for the Application Templates which has a value that
	// should match to the SystemSourceKey's value of the fetched systems
	ApplicationTemplateLabelFilter string
	// SystemSourceKey represents a key for filtering systems
	SystemSourceKey string
	// SystemSynchronizationTimestamps represents the systems last synchronization timestamps for each tenant
	SystemSynchronizationTimestamps map[string]map[string]SystemSynchronizationTimestamp
)

// TemplateMapping holds data for Application Templates and their Labels
type TemplateMapping struct {
	AppTemplate *model.ApplicationTemplate
	Labels      map[string]*model.Label
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
	AdditionalURLs         AdditionalURLs       `json:"additionalUrls"`
	AdditionalAttributes   AdditionalAttributes `json:"additionalAttributes"`
}

// System missing godoc
type System struct {
	SystemBase
	TemplateID      string                           `json:"-"`
	StatusCondition model.ApplicationStatusCondition `json:"-"`
}

// SystemSynchronizationTimestamp represents the last synchronization time of a system
type SystemSynchronizationTimestamp struct {
	ID                string
	LastSyncTimestamp time.Time
}

// UnmarshalJSON missing godoc
func (s *System) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &s.SystemBase); err != nil {
		return err
	}

	for _, tm := range ApplicationTemplates {
		if matchProps(data, tm) {
			s.TemplateID = tm.AppTemplate.ID
			return nil
		}
	}

	return nil
}

func matchProps(data []byte, tm TemplateMapping) bool {
	lbl, ok := tm.Labels[ApplicationTemplateLabelFilter]
	if !ok {
		return false
	}

	templateMappingLabelValue, ok := lbl.Value.(string)
	if !ok {
		return false
	}

	if systemSourceKeyValue := gjson.GetBytes(data, SystemSourceKey).String(); systemSourceKeyValue != templateMappingLabelValue {
		return false
	}

	return true
}
