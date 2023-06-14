package graphql

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

// Validate missing godoc
func (i BundleCreateInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, is.PrintableASCII, validation.Length(1, 100)),
		validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		validation.Field(&i.DefaultInstanceAuth, validation.NilOrNotEmpty),
		validation.Field(&i.InstanceAuthRequestInputSchema, validation.NilOrNotEmpty),
		validation.Field(&i.APIDefinitions, inputvalidation.Each(validation.Required)),
		validation.Field(&i.EventDefinitions, inputvalidation.Each(validation.Required)),
		validation.Field(&i.Documents, inputvalidation.Each(validation.Required)),
	)
}

// Validate missing godoc
func (i BundleUpdateInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, is.PrintableASCII, validation.Length(1, 100)),
		validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		validation.Field(&i.DefaultInstanceAuth, validation.NilOrNotEmpty),
		validation.Field(&i.InstanceAuthRequestInputSchema, validation.NilOrNotEmpty),
	)
}

// Validate missing godoc
func (i BundleInstanceAuthRequestInput) Validate() error {
	// Validation of inputParams against JSON schema is done in Service
	return validation.ValidateStruct(&i,
		validation.Field(&i.ID, is.UUIDv4),
		validation.Field(&i.Context, validation.NilOrNotEmpty),
		validation.Field(&i.InputParams, validation.NilOrNotEmpty),
	)
}

// Validate missing godoc
func (i BundleInstanceAuthSetInput) Validate() error {
	if i.Auth == nil && i.Status == nil {
		return apperrors.NewInvalidDataError("at least one field (Auth or Status) has to be provided")
	}

	if i.Status != nil {
		if i.Auth != nil && i.Status.Condition != BundleInstanceAuthSetStatusConditionInputSucceeded {
			return fmt.Errorf("status condition has to be equal to %s when the auth is provided", BundleInstanceAuthSetStatusConditionInputSucceeded)
		}

		if i.Auth == nil && i.Status.Condition == BundleInstanceAuthSetStatusConditionInputSucceeded {
			return fmt.Errorf("status cannot be equal to %s when auth is not provided", BundleInstanceAuthSetStatusConditionInputSucceeded)
		}
	}

	return validation.ValidateStruct(&i,
		validation.Field(&i.Status, validation.NilOrNotEmpty),
		validation.Field(&i.Auth, validation.NilOrNotEmpty),
	)
}

// Validate missing godoc
func (i BundleInstanceAuthStatusInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Reason, validation.Required),
		validation.Field(&i.Message, validation.Required),
		validation.Field(&i.Condition, validation.Required),
	)
}

// Validate validates BundleInstanceAuthCreateInput
func (i BundleInstanceAuthCreateInput) Validate() error {
	if i.Auth == nil {
		return apperrors.NewInvalidDataError("Auth should be provided")
	}

	// Exactly one of RuntimeID and RuntimeContextID should be provided
	if i.RuntimeID == nil && i.RuntimeContextID == nil {
		return apperrors.NewInvalidDataError("At least one of RuntimeID or RuntimeContextID should be provided")
	} else if i.RuntimeID != nil && i.RuntimeContextID != nil {
		return apperrors.NewInvalidDataError("Only one of RuntimeID or RuntimeContextID should be provided")
	}

	return validation.ValidateStruct(&i,
		validation.Field(&i.Context, validation.NilOrNotEmpty),
		validation.Field(&i.InputParams, validation.NilOrNotEmpty),
	)
}
