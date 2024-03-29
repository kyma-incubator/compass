package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Validate missing godoc
func (i DocumentInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Title, validation.Required, validation.RuneLength(1, shortStringLengthLimit)),
		validation.Field(&i.DisplayName, validation.Required, validation.RuneLength(1, shortStringLengthLimit)),
		validation.Field(&i.Description, validation.Required, validation.RuneLength(1, descriptionStringLengthLimit)),
		validation.Field(&i.Format, validation.Required, validation.In(DocumentFormatMarkdown)),
		validation.Field(&i.Kind, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.Data, validation.NilOrNotEmpty),
		validation.Field(&i.FetchRequest),
	)
}
