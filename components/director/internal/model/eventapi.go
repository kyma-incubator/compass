package model

import "github.com/kyma-incubator/compass/components/director/pkg/pagination"

type EventAPIDefinition struct {
	ID            string
	ApplicationID string
	Name          string
	Description   *string
	Group   *string
	Spec    *EventAPISpec
	Version *Version
}

type EventAPISpecType string

const (
	EventAPISpecTypeAsyncAPI EventAPISpecType = "ASYNC_API"
)

type EventAPISpec struct {
	Data         *[]byte
	Type         EventAPISpecType
	Format       *SpecFormat
	FetchRequest *FetchRequest
}

type Version struct {
	// for example 4.6
	Value      string
	Deprecated *bool
	// for example 4.5
	DeprecatedSince *string
	// if true, will be removed in the next version
	ForRemoval *bool
}

type EventAPIDefinitionPage struct {
	Data       []*EventAPIDefinition
	PageInfo   *pagination.Page
	TotalCount int
}

func (EventAPIDefinitionPage) IsPageable() {}

type SpecFormat string

const (
	SpecFormatYaml SpecFormat = "YAML"
	SpecFormatJSON SpecFormat = "JSON"
)

type EventAPIDefinitionInput struct {
	ApplicationID string
	Name          string
	Description   *string
	Spec          *EventAPISpecInput
	Group         *string
	Version       *VersionInput
}

type EventAPISpecInput struct {
	Data          *[]byte
	EventSpecType EventAPISpecType
	FetchRequest  *FetchRequestInput
}

type VersionInput struct {
	Value           string
	Deprecated      *bool
	DeprecatedSince *string
	ForRemoval      *bool
}

func (e *EventAPIDefinitionInput) ToEventAPIDefinition() *EventAPIDefinition {
	// TODO: Replace with actual model
	return &EventAPIDefinition{}
}
