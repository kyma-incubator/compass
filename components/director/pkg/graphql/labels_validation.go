package graphql

import validation "github.com/go-ozzo/ozzo-validation"

func (i LabelInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Key, validation.Required, validation.Length(0, StringLengthLimitLong)),
		validation.Field(&i.Value, validation.Required),
	)
}
