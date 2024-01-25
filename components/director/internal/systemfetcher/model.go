package systemfetcher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/selfregmanager"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

var (
	// ApplicationTemplates contains available Application Templates, should only be used for the unmarshaling of system data
	// It represents a model.ApplicationTemplate with its labels in the form of map[string]*model.Label
	ApplicationTemplates map[TemplateMappingKey]TemplateMapping
	// SortedTemplateMappingKeys contains an array of TemplateMappingKey that have been sorted by the label proeprty
	SortedTemplateMappingKeys []TemplateMappingKey
	// ApplicationTemplateLabelFilter represent a label for the Application Templates which has a value that
	// should match to the SystemSourceKey's value of the fetched systems
	ApplicationTemplateLabelFilter string
	// SelectFilter represents the select filter that determines which properties of a system will be fetched
	SelectFilter []string
	// SystemSourceKey represents a key for filtering systems
	SystemSourceKey string
)

// TemplateMappingKey is a mapping for regional Application Templates
type TemplateMappingKey struct {
	Label  string
	Region string
}

// TemplateMapping holds data for Application Templates and their Labels
type TemplateMapping struct {
	AppTemplate *model.ApplicationTemplate
	Labels      map[string]*model.Label
	Renderer    TemplateRenderer
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

	return nil
}

// EnhanceWithTemplateID tries to find an Application Template ID for the system and attach it to the object.
func (s *System) EnhanceWithTemplateID() (System, error) {
	for tmKey, tm := range ApplicationTemplates {
		if !s.isMatchedBySystemRole(tmKey) {
			continue
		}
		// Global Application Template
		if tmKey.Region == "" {
			s.TemplateID = tm.AppTemplate.ID
			break
		}

		// Regional Application Template
		appInput, err := tm.Renderer.GenerateAppRegisterInput(context.Background(), *s, tm.AppTemplate, false)
		if err != nil {
			return *s, err
		}

		regionLabel, ok := appInput.Labels[selfregmanager.RegionLabel]
		if !ok {
			return *s, errors.Errorf("%q label should be present for regional app templates", selfregmanager.RegionLabel)
		}

		regionLabelStr, ok := regionLabel.(string)
		if !ok {
			return *s, errors.Errorf("%q label cannot be parsed to string", selfregmanager.RegionLabel)
		}

		foundTemplateMapping := getTemplateMappingBySystemRoleAndRegion(s.SystemPayload, regionLabelStr)
		if foundTemplateMapping.AppTemplate == nil {
			return *s, errors.Errorf("cannot find an app template mapping for a system with payload: %+v", s.SystemPayload)
		}

		s.TemplateID = foundTemplateMapping.AppTemplate.ID
	}

	return *s, nil
}

func (s *System) isMatchedBySystemRole(tmKey TemplateMappingKey) bool {
	systemSource, systemSourceKeyExists := s.SystemPayload[SystemSourceKey]
	if !systemSourceKeyExists {
		return false
	}

	systemSourceValue, ok := systemSource.(string)
	if !ok {
		return false
	}

	return tmKey.Label == systemSourceValue
}

func getTemplateMappingBySystemRoleAndRegion(systemPayload map[string]interface{}, region string) TemplateMapping {
	systemSource, systemSourceKeyExists := systemPayload[SystemSourceKey]
	if !systemSourceKeyExists {
		return TemplateMapping{}
	}

	systemSourceValue, ok := systemSource.(string)
	if !ok {
		return TemplateMapping{}
	}

	for key, mapping := range ApplicationTemplates {
		if key.Label == systemSourceValue && key.Region == region {
			return mapping
		}
	}

	return TemplateMapping{}
}
