package graphql

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/pkg/errors"
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

func (i PackageInstanceAuthRequestInput) Validate() error {
	// Validation of inputParams against JSON schema is done in Service
	return validation.ValidateStruct(&i,
		validation.Field(&i.Context, validation.NilOrNotEmpty),
		validation.Field(&i.InputParams, validation.NilOrNotEmpty),
	)
}

func (i PackageInstanceAuthSetInput) Validate() error {
	if i.Auth == nil && i.Status == nil {
		return errors.New("at least one field (Auth or Status) has to be provided")
	}

	if i.Auth != nil && i.Status != nil && i.Status.Condition != PackageInstanceAuthSetStatusConditionInputSucceeded {
		return fmt.Errorf("status condition has to be equal to %s when the auth is provided", PackageInstanceAuthSetStatusConditionInputSucceeded)
	}

	return validation.ValidateStruct(&i,
		validation.Field(&i.Status, validation.NilOrNotEmpty),
		validation.Field(&i.Auth, validation.NilOrNotEmpty),
	)
}
