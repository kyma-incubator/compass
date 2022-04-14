package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Validate missing godoc
func (i RuntimeContextInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Key, validation.Required, validation.RuneLength(0, longLongStringLengthLimit), validation.Match(alphanumericUnderscoreRegexp)),
	)
}
