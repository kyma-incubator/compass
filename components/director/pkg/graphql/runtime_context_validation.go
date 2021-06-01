package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

func (i RuntimeContextInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Key, validation.Required, validation.RuneLength(0, longLongStringLengthLimit), validation.Match(alphanumericUnderscoreRegexp)),
		validation.Field(&i.Labels, inputvalidation.EachKey(validation.Required, validation.Match(alphanumericUnderscoreRegexp))),
	)
}
