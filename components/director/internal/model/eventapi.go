package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type EventAPIDefinition struct {
	ID            string
	Tenant        string
	ApplicationID string
	Name          string
	Description   *string
	Group         *string
	Spec          *EventAPISpec
	Version       *Version
}

type EventAPISpecType string

const (
	EventAPISpecTypeAsyncAPI EventAPISpecType = "ASYNC_API"
)

type EventAPISpec struct {
	Data           *string
	Type           EventAPISpecType
	Format         SpecFormat
	FetchRequestID *string
}

type EventAPIDefinitionPage struct {
	Data       []*EventAPIDefinition
	PageInfo   *pagination.Page
	TotalCount int
}

func (EventAPIDefinitionPage) IsPageable() {}

type EventAPIDefinitionInput struct {
	Name        string
	Description *string
	Spec        *EventAPISpecInput
	Group       *string
	Version     *VersionInput
}

type EventAPISpecInput struct {
	Data          *string
	EventSpecType EventAPISpecType
	Format        SpecFormat
	FetchRequest  *FetchRequestInput
}

func (e *EventAPIDefinitionInput) ToEventAPIDefinition(id, appID string, fetchRequestID *string) *EventAPIDefinition {
	if e == nil {
		return nil
	}

	return &EventAPIDefinition{
		ID:            id,
		ApplicationID: appID,
		Name:          e.Name,
		Description:   e.Description,
		Group:         e.Group,
		Spec:          e.Spec.ToEventAPISpec(fetchRequestID),
		Version:       e.Version.ToVersion(),
	}
}

func (e *EventAPISpecInput) ToEventAPISpec(fetchRequestID *string) *EventAPISpec {
	if e == nil {
		return nil
	}

	return &EventAPISpec{
		Data:           e.Data,
		Type:           e.EventSpecType,
		Format:         e.Format,
		FetchRequestID: fetchRequestID,
	}
}
