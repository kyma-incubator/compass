package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Validate validates the ASA selector
func (i LabelSelectorInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Key, validation.In("global_subaccount_id")),
	)
}
