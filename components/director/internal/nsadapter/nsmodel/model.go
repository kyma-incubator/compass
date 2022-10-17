package nsmodel

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/tidwall/gjson"
)

// Mappings stores mappings between system values and ApplicationTemplates
var Mappings []TemplateMapping

// TemplateMapping mapping for application templates
type TemplateMapping struct {
	Name        string
	Region      string
	ID          string
	SourceKey   []string
	SourceValue []string
}

// SystemBase represents on-premise system
type SystemBase struct {
	Protocol     string `json:"protocol"`
	Host         string `json:"host"`
	SystemType   string `json:"type"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	SystemNumber string `json:"systemNumber"`
}

// System represents on-premise system with ApplicationTemplate ID
type System struct {
	SystemBase
	TemplateID string `json:"-"`
}

// Validate validates System fields
func (s System) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Protocol, validation.Required),
		validation.Field(&s.Host, validation.Required),
		validation.Field(&s.SystemType, validation.Required),
		validation.Field(&s.Description, validation.NotNil),
		validation.Field(&s.Status, validation.Required),
		validation.Field(&s.SystemNumber, validation.NotNil),
	)
}

// UnmarshalJSON unmarshal the provided data into System
func (s *System) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &s.SystemBase); err != nil {
		return err
	}

	for _, mapping := range Mappings {
		if matchProps(data, mapping) {
			s.TemplateID = mapping.ID
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

// SCC represents SAP Cloud Connector
type SCC struct {
	ExternalSubaccountID string   `json:"subaccount"`
	InternalSubaccountID string   `json:"-"`
	LocationID           string   `json:"locationID"`
	ExposedSystems       []System `json:"exposedSystems"`
}

// Validate validates SCC fields
func (s SCC) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.ExternalSubaccountID, validation.Required),
		validation.Field(&s.LocationID, validation.NotNil),
		validation.Field(&s.ExposedSystems, validation.NotNil, validation.By(validateSystems)),
	)
}

func validateSystems(value interface{}) error {
	if systems, ok := value.([]System); ok {
		for _, s := range systems {
			if err := s.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Report represents Notification Service reports to CMP
type Report struct {
	ReportType string `json:"type"`
	Value      []SCC  `json:"value"`
}

// Validate validates Report fields
func (r Report) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ReportType, validation.Required),
		validation.Field(&r.Value, validation.NotNil, validation.By(validateSCCs)),
	)
}

func validateSCCs(value interface{}) error {
	if scc, ok := value.([]SCC); ok {
		for _, s := range scc {
			if err := s.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

// ToAppUpdateInput converts System to model.ApplicationUpdateInput
func ToAppUpdateInput(system System) model.ApplicationUpdateInput {
	return model.ApplicationUpdateInput{
		Description:  str.Ptr(system.Description),
		SystemStatus: str.Ptr(system.Status),
	}
}
