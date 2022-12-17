package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/pkg/errors"
)

// Validate missing godoc
func (i APIDefinitionInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, is.PrintableASCII, validation.Length(1, appNameLengthLimit)),
		validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		validation.Field(&i.TargetURL, validation.Required, inputvalidation.IsURL, validation.RuneLength(1, longStringLengthLimit)),
		validation.Field(&i.Group, validation.RuneLength(0, groupLengthLimit)),
		validation.Field(&i.Spec, validation.NilOrNotEmpty),
		validation.Field(&i.Version, validation.NilOrNotEmpty),
	)
}

// Validate missing godoc
func (i APISpecInput) Validate() error {
	return validation.Errors{
		"Rule.Type":                  validation.Validate(&i.Type, validation.Required, validation.In(APISpecTypeOdata, APISpecTypeOpenAPI)),
		"Rule.Format":                validation.Validate(&i.Format, validation.Required, validation.In(SpecFormatYaml, SpecFormatJSON, SpecFormatXML)),
		"Rule.MatchingTypeAndFormat": i.validateTypeWithMatchingSpecFormat(),
		"Rule.FetchRequest":          validation.Validate(&i.FetchRequest),
		"Rule.DataOrFetchRequest":    inputvalidation.ValidateExactlyOneNotNil("Only one of Data or Fetch Request must be passed", i.Data, i.FetchRequest),
	}.Filter()
}

func (i APISpecInput) validateTypeWithMatchingSpecFormat() error {
	switch i.Type {
	case APISpecTypeOdata:
		if !i.Format.isOneOf([]SpecFormat{SpecFormatXML, SpecFormatJSON}) {
			return errors.Errorf("%s is not a valid spec format for spec type %s", i.Format, i.Type)
		}
	case APISpecTypeOpenAPI:
		if !i.Format.isOneOf([]SpecFormat{SpecFormatJSON, SpecFormatYaml}) {
			return errors.Errorf("%s is not a valid spec format for spec type %s", i.Format, i.Type)
		}
	default:
		return errors.Errorf("%s is not a valid spec type", i.Type)
	}
	return nil
}
