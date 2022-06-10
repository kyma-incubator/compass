package graphql

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

var runtimeNameRgx = regexp.MustCompile(`^[a-zA-Z0-9-._]+$`)

// Validate missing godoc
func (i RuntimeRegisterInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, validation.RuneLength(1, longStringLengthLimit), validation.Match(runtimeNameRgx)),
		validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		validation.Field(&i.Labels, inputvalidation.EachKey(validation.Required, validation.Match(alphanumericUnderscoreRegexp))),
		validation.Field(&i.Webhooks, validation.Each(validation.Required)),
	)
}

// Validate missing godoc
func (i RuntimeUpdateInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, validation.RuneLength(1, longStringLengthLimit), validation.Match(runtimeNameRgx)),
		validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		validation.Field(&i.Labels, inputvalidation.EachKey(validation.Required, validation.Match(alphanumericUnderscoreRegexp))),
	)
}
