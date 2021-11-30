package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Validate validates the ASA selector of the ASA Input
func (i AutomaticScenarioAssignmentSetInput) Validate() error {
	return validation.Validate(&i.Selector)
}

// Validate validates the ASA selector
func (i LabelSelectorInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Key, validation.In("global_subaccount_id")),
	)
}
