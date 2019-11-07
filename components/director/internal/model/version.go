package model

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

type Version struct {
	// for example 4.6
	Value      string
	Deprecated *bool
	// for example 4.5
	DeprecatedSince *string
	// if true, will be removed in the next version
	ForRemoval *bool
}

type VersionInput struct {
	Value           string
	Deprecated      *bool
	DeprecatedSince *string
	ForRemoval      *bool
}

func (i *VersionInput) Validate() error {
	return validation.ValidateStruct(i,
		validation.Field(&i.Value, validation.Required, validation.Length(1, 256), validation.By(inputvalidation.ValidatePrintable)),
		validation.Field(&i.Deprecated, validation.Required),
		validation.Field(&i.DeprecatedSince, validation.NilOrNotEmpty, validation.Length(1, 256), validation.By(inputvalidation.ValidatePrintable)),
		validation.Field(&i.ForRemoval, validation.Required),
	)
}

func (v *VersionInput) ToVersion() *Version {
	if v == nil {
		return nil
	}

	return &Version{
		Value:           v.Value,
		Deprecated:      v.Deprecated,
		DeprecatedSince: v.DeprecatedSince,
		ForRemoval:      v.ForRemoval,
	}
}
