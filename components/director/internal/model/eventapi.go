package model

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
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
	Data   *string
	Type   EventAPISpecType
	Format SpecFormat
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

func (i *EventAPIDefinitionInput) Validate() error {
	return validation.ValidateStruct(i,
		validation.Field(&i.Name, validation.Required, validation.By(inputvalidation.ValidateName)),
		validation.Field(&i.Description, validation.NilOrNotEmpty, validation.Length(1, 128), validation.By(inputvalidation.ValidatePrintableWithWhitespace)),
		validation.Field(&i.Spec, validation.Required),
		validation.Field(&i.Group, validation.NilOrNotEmpty, validation.Length(1, 36), validation.By(inputvalidation.ValidatePrintable)),
		validation.Field(&i.Version),
	)
}

type EventAPISpecInput struct {
	Data          *string
	EventSpecType EventAPISpecType
	Format        SpecFormat
	FetchRequest  *FetchRequestInput
}

func (i *EventAPISpecInput) Validate() error {
	return validation.Errors{
		"MatchingTypeAndFormat": i.validateMatchingSpecAndType(),
		"Data":                  validation.Validate(i.Data, validation.NilOrNotEmpty, validation.By(inputvalidation.ValidatePrintableWithWhitespace)),
		"EventSpecType":         validation.Validate(i.EventSpecType, validation.Required, validation.In(EventAPISpecTypeAsyncAPI), validation.By(inputvalidation.ValidatePrintable)),
		"Format":                validation.Validate(i.Format, validation.Required, validation.In(SpecFormatYaml, SpecFormatJSON), validation.By(inputvalidation.ValidatePrintable)),
		"FetchRequest":          validation.Validate(i.FetchRequest),
	}.Filter()
}

func (i *EventAPISpecInput) validateMatchingSpecAndType() error {
	switch i.EventSpecType {
	case EventAPISpecTypeAsyncAPI:
		if !i.formatIsOneOf(i.Format, []SpecFormat{SpecFormatYaml, SpecFormatJSON}) {
			return errors.Errorf("format %s is not a valid spec format for spec type %s", i.Format, i.EventSpecType)
		}
	default:
		return errors.New("invalid spec type")
	}
	return nil
}

func (i *EventAPISpecInput) formatIsOneOf(format SpecFormat, formats []SpecFormat) bool {
	var slice []string
	for _, value := range formats {
		slice = append(slice, (string)(value))
	}
	return str.IsInSlice((string)(i.Format), slice)
}

func (e *EventAPIDefinitionInput) ToEventAPIDefinition(id, appID, tenant string) *EventAPIDefinition {
	if e == nil {
		return nil
	}

	return &EventAPIDefinition{
		ID:            id,
		ApplicationID: appID,
		Tenant:        tenant,
		Name:          e.Name,
		Description:   e.Description,
		Group:         e.Group,
		Spec:          e.Spec.ToEventAPISpec(),
		Version:       e.Version.ToVersion(),
	}
}

func (e *EventAPISpecInput) ToEventAPISpec() *EventAPISpec {
	if e == nil {
		return nil
	}

	return &EventAPISpec{
		Data:   e.Data,
		Type:   e.EventSpecType,
		Format: e.Format,
	}
}
