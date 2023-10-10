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
	// SelectFilter represents the select filter that determines which properties of a system will be fetched
	SelectFilter []string
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

// System missing godoc
type System struct {
	SystemPayload   map[string]interface{}
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
	if err := json.Unmarshal(data, &s.SystemPayload); err != nil {
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

	templateMappingLabelValues, ok := lbl.Value.([]string)
	if !ok {
		return false
	}

	found := false
	for _, labelValue := range templateMappingLabelValues {
		if systemSourceKeyValue := gjson.GetBytes(data, SystemSourceKey).String(); systemSourceKeyValue == labelValue {
			found = true
			break
		}
	}

	return found
}
