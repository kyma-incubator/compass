package systemfetcher

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

var (
	// ApplicationTemplates contains available Application Templates, should only be used for the unmarshalling of system data
	// It represents a model.ApplicationTemplate with its labels in the form of map[string]*model.Label
	ApplicationTemplates []TemplateMapping
	// ApplicationTemplateLabelFilter represent a label for the Application Templates which has a value that
	// should match to the SystemSourceKey's value of the fetched systems
	ApplicationTemplateLabelFilter string
	// SelectFilter represents the select filter that determines which properties of a system will be fetched
	SelectFilter []string
	// SystemSourceKey represents a key for filtering systems
	SystemSourceKey string
)

//// TemplateMappingKey is a mapping for regional Application Templates
//type TemplateMappingKey struct {
//	Label  string
//	Region string
//}

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

type CldFilterOperationType string

const (
	IncludeOperationType CldFilterOperationType = "include"
	ExcludeOperationType CldFilterOperationType = "exclude"
)

type CldFilter struct {
	Key       string
	Value     []string
	Operation CldFilterOperationType
}

type ProductIDFilterMapping struct {
	ProductID string      `json:"productID"`
	Filter    []CldFilter `json:"filter,omitempty"`
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
	for _, tm := range ApplicationTemplates {
		//if !s.isMatchedBySystemRole(tmKey) {
		//	continue
		//}
		cldFilter, cldFilterExists := tm.Labels["cldFilter"]
		if !cldFilterExists {
			return *s, errors.New("missing cld filter")
		}

		cldFilterValue, ok := cldFilter.Value.([]ProductIDFilterMapping)
		if !ok {
			return *s, errors.New("cld filter value cannot be cast to ProductIDFilterMapping")
		}

		systemSource, systemSourceValueExists := s.SystemPayload[SystemSourceKey]
		if !systemSourceValueExists {

		}

		for _, l := range cldFilterValue {
			if l.ProductID == systemSource {
				systemMatches, err := systemMatchesCldFilters(s.SystemPayload, l.Filter)
				if err != nil {
					return *s, err
				}

				if systemMatches {
					s.TemplateID = tm.AppTemplate.ID
					return *s, nil
				}
			}
		}

		//
		//// Global Application Template
		//if tmKey.Region == "" {
		//	s.TemplateID = tm.AppTemplate.ID
		//	break
		//}
		//
		//// Regional Application Template
		//// Use GenerateAppRegisterInput to resolve the application Input. This way we would know what is the actual
		//// region from the system payload
		//appInput, err := tm.Renderer.GenerateAppRegisterInput(context.Background(), *s, tm.AppTemplate, false)
		//if err != nil {
		//	return *s, err
		//}
		//
		//regionLabel, err := getLabelFromInput(appInput)
		//if err != nil {
		//	return *s, err
		//}
		//
		//foundTemplateMapping := getTemplateMappingBySystemRoleAndRegion(s.SystemPayload, regionLabel)
		//if foundTemplateMapping.AppTemplate == nil {
		//	return *s, errors.Errorf("cannot find an app template mapping for a system with payload: %+v", s.SystemPayload)
		//}
		//
		//s.TemplateID = foundTemplateMapping.AppTemplate.ID
		//break
	}

	return *s, nil
}

func systemMatchesCldFilters(systemPayload map[string]interface{}, cldFilters []CldFilter) (bool, error) {
	prefix := "$."

	payload, _ := json.Marshal(systemPayload)

	for _, f := range cldFilters {
		path := strings.TrimPrefix(f.Key, prefix)
		valueFromSystemPayload := gjson.Get(string(payload), path)

		switch f.Operation {
		case IncludeOperationType:
			if !valueMatchesFilter(valueFromSystemPayload.String(), f.Value, true) {
				return false, nil
			}
		case ExcludeOperationType:
			if !valueMatchesFilter(valueFromSystemPayload.String(), f.Value, false) {
				return false, nil
			}
		default:
			return false, errors.New("not from op type error")
		}
	}

	return true, nil
}

func valueMatchesFilter(value string, filterValues []string, expectedResultIfEqualValue bool) bool {
	for _, v := range filterValues {
		if value == v {
			return expectedResultIfEqualValue
		}
	}
	return !expectedResultIfEqualValue
}

//func (s *System) isMatchedBySystemRole(tmKey TemplateMappingKey) bool {
//	systemSource, systemSourceKeyExists := s.SystemPayload[SystemSourceKey]
//	if !systemSourceKeyExists {
//		return false
//	}
//
//	systemSourceValue, ok := systemSource.(string)
//	if !ok {
//		return false
//	}
//
//	return tmKey.Label == systemSourceValue
//}

//func getTemplateMappingBySystemRoleAndRegion(systemPayload map[string]interface{}, region string) TemplateMapping {
//	systemSource, systemSourceKeyExists := systemPayload[SystemSourceKey]
//	if !systemSourceKeyExists {
//		return TemplateMapping{}
//	}
//
//	systemSourceValue, ok := systemSource.(string)
//	if !ok {
//		return TemplateMapping{}
//	}
//
//	for key, mapping := range ApplicationTemplates {
//		if key.Label == systemSourceValue && key.Region == region {
//			return mapping
//		}
//	}
//
//	return TemplateMapping{}
//}

//func getLabelFromInput(appInput *model.ApplicationRegisterInput) (string, error) {
//	regionLabel, ok := appInput.Labels[selfregmanager.RegionLabel]
//	if !ok {
//		return "", errors.Errorf("%q label should be present for regional app templates", selfregmanager.RegionLabel)
//	}
//
//	regionLabelStr, ok := regionLabel.(string)
//	if !ok {
//		return "", errors.Errorf("%q label cannot be parsed to string", selfregmanager.RegionLabel)
//	}
//
//	return regionLabelStr, nil
//}
