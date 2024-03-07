package graphql

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// SemVerRegex represents the valid structure of the field
const SemVerRegex = "^(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$"

// Validate missing godoc
func (i VersionInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Value, validation.Required, validation.RuneLength(1, longStringLengthLimit), validation.Match(regexp.MustCompile(SemVerRegex))),
		validation.Field(&i.Deprecated, validation.NotNil),
		validation.Field(&i.DeprecatedSince, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.ForRemoval, validation.NotNil),
	)
}
