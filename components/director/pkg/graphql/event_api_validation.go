package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/pkg/errors"
)

func (i EventDefinitionInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, is.PrintableASCII, validation.Length(1, 100)),
		validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		validation.Field(&i.Spec, validation.NilOrNotEmpty),
		validation.Field(&i.Group, validation.RuneLength(0, groupLengthLimit)),
		validation.Field(&i.Version, validation.NilOrNotEmpty),
	)
}

func (i EventSpecInput) Validate() error {
	return validation.Errors{
		"Rule.Type":                  validation.Validate(&i.Type, validation.Required, validation.In(EventSpecTypeAsyncAPI)),
		"Rule.Format":                validation.Validate(&i.Format, validation.Required, validation.In(SpecFormatYaml, SpecFormatJSON)),
		"Rule.MatchingTypeAndFormat": i.validateTypeWithMatchingSpecFormat(),
		"Rule.FetchRequest":          validation.Validate(&i.FetchRequest),
		"Rule.DataOrFetchRequest":    inputvalidation.ValidateExactlyOneNotNil("Only one of Data or Fetch Request must be passed", i.Data, i.FetchRequest),
	}.Filter()
}

func (i EventSpecInput) validateTypeWithMatchingSpecFormat() error {
	switch i.Type {
	case EventSpecTypeAsyncAPI:
		if !i.Format.isOneOf([]SpecFormat{SpecFormatYaml, SpecFormatJSON}) {
			return errors.Errorf("format %s is not a valid spec format for spec type %s", i.Format, i.Type)
		}
	default:
		return errors.Errorf("%s is an invalid spec type", i.Type)
	}
	return nil
}
