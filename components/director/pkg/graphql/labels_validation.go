package graphql

import validation "github.com/go-ozzo/ozzo-validation"

func (i LabelInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Key, validation.Required, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.Value, validation.Required),
	)
}
