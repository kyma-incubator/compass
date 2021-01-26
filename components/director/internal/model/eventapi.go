package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type EventDefinition struct {
	ID          string
	Tenant      string
	BundleID    string
	Name        string
	Description *string
	Group       *string
	Spec        *EventSpec
	Version     *Version
}

type EventSpecType string

const (
	EventSpecTypeAsyncAPI EventSpecType = "ASYNC_API"
)

type EventSpec struct {
	Data   *string
	Type   EventSpecType
	Format SpecFormat
}

type EventDefinitionPage struct {
	Data       []*EventDefinition
	PageInfo   *pagination.Page
	TotalCount int
}

func (EventDefinitionPage) IsPageable() {}

type EventDefinitionInput struct {
	Name        string
	Description *string
	Spec        *EventSpecInput
	Group       *string
	Version     *VersionInput
}

type EventSpecInput struct {
	Data          *string
	EventSpecType EventSpecType
	Format        SpecFormat
	FetchRequest  *FetchRequestInput
}

func (e *EventDefinitionInput) ToEventDefinitionWithinBundle(id string, bndlID string, tenant string) *EventDefinition {
	if e == nil {
		return nil
	}

	return &EventDefinition{
		ID:          id,
		BundleID:    bndlID,
		Tenant:      tenant,
		Name:        e.Name,
		Description: e.Description,
		Group:       e.Group,
		Spec:        e.Spec.ToEventSpec(),
		Version:     e.Version.ToVersion(),
	}
}

func (e *EventSpecInput) ToEventSpec() *EventSpec {
	if e == nil {
		return nil
	}

	return &EventSpec{
		Data:   e.Data,
		Type:   e.EventSpecType,
		Format: e.Format,
	}
}
