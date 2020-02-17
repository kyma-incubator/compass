package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

func (i PackageCreateInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, is.PrintableASCII, validation.Length(1, 100)),
		validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		validation.Field(&i.DefaultInstanceAuth, validation.NilOrNotEmpty),
		validation.Field(&i.InstanceAuthRequestInputSchema, validation.NilOrNotEmpty),
		validation.Field(&i.APIDefinitions, inputvalidation.Each(validation.Required)),
		validation.Field(&i.EventDefinitions, inputvalidation.Each(validation.Required)),
		validation.Field(&i.Documents, inputvalidation.Each(validation.Required)),
	)
}

func (i PackageUpdateInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, is.PrintableASCII, validation.Length(1, 100)),
		validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		validation.Field(&i.DefaultInstanceAuth, validation.NilOrNotEmpty),
		validation.Field(&i.InstanceAuthRequestInputSchema, validation.NilOrNotEmpty),
	)
}

// TODO: Replace with real implementation
func (i PackageInstanceAuthRequestInput) Validate() error {
	return nil
}

