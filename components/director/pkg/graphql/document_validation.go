package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

func (i DocumentInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Title, validation.Required, validation.Length(1, shortStringLengthLimit)),
		validation.Field(&i.DisplayName, validation.Required, validation.Length(1, shortStringLengthLimit)),
		validation.Field(&i.Description, validation.Required, validation.Length(1, shortStringLengthLimit)),
		validation.Field(&i.Format, validation.Required, validation.In(DocumentFormatMarkdown)),
		validation.Field(&i.Kind, validation.Length(0, longStringLengthLimit)),
		validation.Field(&i.Data, validation.NilOrNotEmpty),
		validation.Field(&i.FetchRequest),
	)
}
