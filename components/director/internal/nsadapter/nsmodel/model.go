package nsmodel

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/tidwall/gjson"
)

var Mappings []systemfetcher.TemplateMapping

type SystemBase struct {
	Protocol     string `json:"protocol"`
	Host         string `json:"host"`
	SystemType   string `json:"type"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	SystemNumber string `json:"systemNumber"`
}

type System struct {
	SystemBase
	TemplateID string `json:"-"`
}

func (s System) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Protocol, validation.Required),
		validation.Field(&s.Host, validation.Required),
		validation.Field(&s.SystemType, validation.Required),
		validation.Field(&s.Description, validation.NotNil),
		validation.Field(&s.Status, validation.Required),
	)
}

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

func matchProps(data []byte, tm systemfetcher.TemplateMapping) bool {
	for i, sk := range tm.SourceKey {
		v := gjson.GetBytes(data, sk).String()
		if v != tm.SourceValue[i] {
			return false
		}
	}
	return true
}

type SCC struct {
	Subaccount     string   `json:"subaccount"`
	LocationID     string   `json:"locationID"`
	ExposedSystems []System `json:"exposedSystems"`
}

func (s SCC) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Subaccount, validation.Required),
		validation.Field(&s.LocationID, validation.NotNil),
		validation.Field(&s.ExposedSystems, validation.NotNil, validation.By(validateSystems)), //TODO test if exposed systems field is nil, will validateSystems method be executed?
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

type Report struct {
	ReportType string `json:"type"`
	Value      []SCC  `json:"value"`
}

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

func ToAppRegisterInput(system System, subaccount, locationID string) model.ApplicationRegisterInput {
	sccLabel := struct {
		Subaccount string `json:"subaccount"`
		Host       string `json:"host"`
		LocationId string `json:"locationId"`
	}{
		subaccount,
		system.Host,
		locationID,
	}
	return model.ApplicationRegisterInput{
		Name:         "on-premise-system", //TODO doublecheck the name
		ProviderName: str.Ptr("SAP"),
		Labels:       map[string]interface{}{"scc": sccLabel, "applicationType": system.SystemType, "systemProtocol": system.Protocol},
		SystemNumber: str.Ptr(system.SystemNumber),
		SystemStatus: str.Ptr(system.Status),
	}
}

func ToAppUpdateInput(system System) model.ApplicationUpdateInput {
	return model.ApplicationUpdateInput{
		Description:  str.Ptr(system.Description),
		SystemStatus: str.Ptr(system.Status),
	}
}
