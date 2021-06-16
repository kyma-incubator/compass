package graphql

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

var runtimeNameRgx = regexp.MustCompile(`^[a-zA-Z0-9-._]+$`)

func (i RuntimeInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, validation.RuneLength(1, appNameLengthLimit), validation.Match(runtimeNameRgx)),
		validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		validation.Field(&i.Labels, inputvalidation.EachKey(validation.Required, validation.Match(alphanumericUnderscoreRegexp))),
	)
}
