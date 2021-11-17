package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (i AutomaticScenarioAssignmentSetInput) Validate() error {
	return validation.Validate(&i.Selector)
}

// Validate missing godoc
func (i LabelSelectorInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Key, validation.In("global_subaccount_id")),
	)
}
