package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

func (i RuntimeInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, inputvalidation.RuntimeName),
		validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		validation.Field(&i.Labels, inputvalidation.EachKey(validation.Required, validation.Match(alphanumericUnderscoreRegexp))),
	)
}
