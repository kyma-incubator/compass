package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/pkg/errors"
)

func (i EventAPIDefinitionInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, inputvalidation.Name),
		validation.Field(&i.Description, validation.RuneLength(0, shortStringLengthLimit)),
		validation.Field(&i.Spec, validation.Required),
		validation.Field(&i.Group, validation.RuneLength(0, groupLengthLimit)),
		validation.Field(&i.Version, validation.NilOrNotEmpty),
	)
}

func (i EventAPISpecInput) Validate() error {
	return validation.Errors{
		"Rule.Type":                  validation.Validate(&i.EventSpecType, validation.Required, validation.In(EventAPISpecTypeAsyncAPI)),
		"Rule.Format":                validation.Validate(&i.Format, validation.Required, validation.In(SpecFormatYaml, SpecFormatJSON)),
		"Rule.MatchingTypeAndFormat": i.validateTypeWithMatchingSpecFormat(),
		"Rule.FetchRequest":          validation.Validate(&i.FetchRequest),
		"Rule.DataOrFetchRequest":    inputvalidation.ValidateExactlyOneNotNil("Only one of Data or Fetch Request must be passed", i.Data, i.FetchRequest),
	}.Filter()
}

func (i EventAPISpecInput) validateTypeWithMatchingSpecFormat() error {
	switch i.EventSpecType {
	case EventAPISpecTypeAsyncAPI:
		if !i.Format.isOneOf([]SpecFormat{SpecFormatYaml, SpecFormatJSON}) {
			return errors.Errorf("format %s is not a valid spec format for spec type %s", i.Format, i.EventSpecType)
		}
	default:
		return errors.Errorf("%s is an invalid spec type", i.EventSpecType)
	}
	return nil
}
