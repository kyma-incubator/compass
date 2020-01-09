package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

func (i FetchRequestInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.URL, validation.Required, is.URL, validation.RuneLength(1, longStringLengthLimit)),
		validation.Field(&i.Auth, validation.NilOrNotEmpty),
		validation.Field(&i.Mode, validation.NilOrNotEmpty, validation.In(FetchModeSingle, FetchModePackage, FetchModeIndex)),
		validation.Field(&i.Filter, validation.NilOrNotEmpty, validation.RuneLength(1, longStringLengthLimit)),
	)
}
