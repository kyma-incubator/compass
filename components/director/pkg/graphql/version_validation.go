package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

func (i VersionInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Value, validation.Required, validation.RuneLength(1, longStringLengthLimit)),
		validation.Field(&i.Deprecated, validation.NotNil),
		validation.Field(&i.DeprecatedSince, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.ForRemoval, validation.NotNil),
	)
}
