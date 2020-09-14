package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type Spec struct {
	ID                string
	Tenant            string
	APIDefinitionID   *string
	EventDefinitionID *string
	Data              *string
	Format            SpecFormat
	Type              SpecType
	CustomType        *string
}

func (s *Spec) ToAPISpec() *APISpec {
	return &APISpec{
		ID:         s.ID,
		Data:       s.Data,
		Format:     s.Format,
		Type:       APISpecType(s.Type), // TODO: Check
		CustomType: s.CustomType,
	}
}

func (s *Spec) ToEventSpec() *EventSpec {
	return &EventSpec{
		ID:         s.ID,
		Data:       s.Data,
		Format:     s.Format,
		Type:       EventSpecType(s.Type), // TODO: Check
		CustomType: s.CustomType,
	}
}

type SpecType string

const (
	SpecTypeOdata    SpecType = "ODATA"
	SpecTypeOpenAPI  SpecType = "OPEN_API"
	SpecTypeAsyncAPI SpecType = "ASYNC_API"
	SpecTypeCustom   SpecType = "CUSTOM"
)

type SpecInput struct {
	ID           string
	Tenant       string
	Data         *string
	Format       SpecFormat
	Type         SpecType
	CustomType   *string
	FetchRequest *FetchRequestInput
}

type SpecPage struct {
	Data       []*Spec
	PageInfo   *pagination.Page
	TotalCount int
}

func (SpecPage) IsPageable() {}

func (a *SpecInput) ToSpecWithinAPI(apiID string, tenant string) *Spec {
	if a == nil {
		return nil
	}

	return &Spec{
		ID:              a.ID,
		Tenant:          tenant,
		APIDefinitionID: &apiID,
		Data:            a.Data,
		Format:          a.Format,
		Type:            a.Type,
		CustomType:      a.CustomType,
	}
}

func (a *SpecInput) ToSpecWithinEvent(eventID string, tenant string) *Spec {
	if a == nil {
		return nil
	}

	return &Spec{
		ID:                a.ID,
		Tenant:            tenant,
		EventDefinitionID: &eventID,
		Data:              a.Data,
		Format:            a.Format,
		Type:              a.Type,
		CustomType:        a.CustomType,
	}
}
