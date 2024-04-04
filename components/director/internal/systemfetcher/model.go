package systemfetcher

import (
	"encoding/json"
	"fmt"
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

type SlisFilterOperationType string

const (
	IncludeOperationType SlisFilterOperationType = "include"
	ExcludeOperationType SlisFilterOperationType = "exclude"
)

type SlisFilter struct {
	Key       string
	Value     []string
	Operation SlisFilterOperationType
}

type ProductIDFilterMapping struct {
	ProductID string       `json:"productID"`
	Filter    []SlisFilter `json:"filter,omitempty"`
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
		slisFilter, slisFilterExists := tm.Labels["slisFilter"]
		if !slisFilterExists {
			return *s, errors.New("missing slis filter")
		}

		productIDFilterMappings := make([]ProductIDFilterMapping, 0)

		jsonData, err := json.Marshal(slisFilter.Value)
		if err != nil {
			fmt.Println("Error marshalling JSON:", err)

		}

		err = json.Unmarshal(jsonData, &productIDFilterMappings)
		if err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
		}

		systemSource, systemSourceValueExists := s.SystemPayload[SystemSourceKey]

		if !systemSourceValueExists {
			//return *s, errors.New("system role does not exist")
		}

		for _, mapping := range productIDFilterMappings {
			if mapping.ProductID == systemSource {
				systemMatches, err := systemMatchesSlisFilters(s.SystemPayload, mapping.Filter)
				if err != nil {
					return *s, err
				}

				if systemMatches {
					s.TemplateID = tm.AppTemplate.ID
					return *s, nil
				}
			}
		}
	}

	return *s, nil
}

func systemMatchesSlisFilters(systemPayload map[string]interface{}, slisFilters []SlisFilter) (bool, error) {
	prefix := "$."

	payload, _ := json.Marshal(systemPayload)

	for _, f := range slisFilters {
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
