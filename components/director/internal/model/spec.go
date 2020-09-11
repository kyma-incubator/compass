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
}

func (s *Spec) ToAPISpec() *APISpec {
	return &APISpec{
		Data:   s.Data,
		Format: s.Format,
		Type:   APISpecType(s.Type), // TODO: Check
	}
}

type SpecType string

const (
	SpecTypeOdata    SpecType = "ODATA"
	SpecTypeOpenAPI  SpecType = "OPEN_API"
	SpecTypeAsyncAPI SpecType = "ASYNC_API"
)

type SpecInput struct {
	ID     string
	Tenant string
	Data   *string
	Format SpecFormat
	Type   SpecType
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
	}
}