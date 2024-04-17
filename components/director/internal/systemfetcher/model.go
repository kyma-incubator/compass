package systemfetcher

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	// SlisFilterLabelKey is the name of the slis filter label for app template
	SlisFilterLabelKey = "slisFilter"
	// ProductIDKey is the name of the property relating to system roles in system payload
	ProductIDKey = "productId"
	// TrimPrefix is the prefix which should be trimmed from the jsonpath input when building select criteria for system fetcher
	TrimPrefix = "$."
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

// SlisFilterOperationType represents the type of operation which considers whether to match the values in the slis filter with values for systems from system payload
type SlisFilterOperationType string

const (
	// IncludeOperationType is the type of slis filter operation which requires the values from slis filter to be present in the values from system payload
	IncludeOperationType SlisFilterOperationType = "include"
	// ExcludeOperationType is the type of slis filter operation which requires the values from slis filter to not be present in the values from system payload
	ExcludeOperationType SlisFilterOperationType = "exclude"
)

// SlisFilter represents the additional properties by which fetched systems are matched to an application template
type SlisFilter struct {
	Key       string                  `json:"key"`
	Value     []string                `json:"value"`
	Operation SlisFilterOperationType `json:"operation"`
}

// ProductIDFilterMapping represents the structure of the slis filter per product id
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
		slisFilter, slisFilterExists := tm.Labels[SlisFilterLabelKey]
		if !slisFilterExists {
			return *s, errors.Errorf("missing slis filter for application template with ID %q", tm.AppTemplate.ID)
		}

		productIDFilterMappings := make([]ProductIDFilterMapping, 0)

		slisFilterLabelJSON, err := json.Marshal(slisFilter.Value)
		if err != nil {
			return *s, err
		}

		err = json.Unmarshal(slisFilterLabelJSON, &productIDFilterMappings)
		if err != nil {
			return *s, err
		}

		systemSource, systemSourceValueExists := s.SystemPayload[SystemSourceKey]

		if !systemSourceValueExists {
			return *s, nil
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
	payload, err := json.Marshal(systemPayload)
	if err != nil {
		return false, err
	}

	for _, filter := range slisFilters {
		path := strings.TrimPrefix(filter.Key, TrimPrefix)
		valueFromSystemPayload := gjson.Get(string(payload), path)

		switch filter.Operation {
		case IncludeOperationType:
			if !valueMatchesFilter(valueFromSystemPayload.String(), filter.Value, true) {
				return false, nil
			}
		case ExcludeOperationType:
			if !valueMatchesFilter(valueFromSystemPayload.String(), filter.Value, false) {
				return false, nil
			}
		default:
			return false, errors.New("slis filter operation doesn't match the defined operation types")
		}
	}

	return true, nil
}

func valueMatchesFilter(value string, filterValues []string, expectedResultIfEqualValue bool) bool {
	for _, filterValue := range filterValues {
		if value == filterValue {
			return expectedResultIfEqualValue
		}
	}
	return !expectedResultIfEqualValue
}
