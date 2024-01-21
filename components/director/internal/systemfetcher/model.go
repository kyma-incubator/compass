package systemfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/selfregmanager"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

var (
	// ApplicationTemplates contains available Application Templates, should only be used for the unmarshaling of system data
	// It represents a model.ApplicationTemplate with its labels in the form of map[string]*model.Label
	ApplicationTemplates map[TemplateMappingKey]TemplateMapping
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

func (s *System) EnhanceWithTemplateID() (System, error) {
	fmt.Println("EnhanceWithTemplateID", s)
	for tmKey, tm := range ApplicationTemplates {
		if !s.isMatchedBySystemRole(tmKey) {
			continue
		}
		fmt.Println("tmKey.Region", tmKey.Region)
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
			return *s, errors.New("label is missing - more info...")
		}

		regionLabelStr, ok := regionLabel.(string)
		if !ok {
			return *s, errors.New("label is not a string missing - more info...")
		}

		foundTemplateMapping := getTemplateMappingBySystemRoleAndRegion(s.SystemPayload, regionLabelStr)
		if foundTemplateMapping.AppTemplate == nil {
			return *s, errors.New("no template mapping found. more info...")
		}

		s.TemplateID = foundTemplateMapping.AppTemplate.ID
	}

	return *s, nil
}

func (s *System) isMatched(tmKey TemplateMappingKey, tm TemplateMapping) bool {
	lbl, ok := tm.Labels[ApplicationTemplateLabelFilter]
	if !ok {
		return false
	}

	templateMappingLabelValues, ok := lbl.Value.([]interface{})
	if !ok {
		return false
	}

	for _, labelValue := range templateMappingLabelValues {
		labelStr, ok := labelValue.(string)
		if !ok {
			continue
		}

		systemSource, systemSourceKeyExists := s.SystemPayload[SystemSourceKey]
		if !systemSourceKeyExists {
			continue
		}

		systemSourceValue, ok := systemSource.(string)
		if !ok {
			continue
		}

		if systemSourceValue == labelStr {
			return true
		}
	}

	return false
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
