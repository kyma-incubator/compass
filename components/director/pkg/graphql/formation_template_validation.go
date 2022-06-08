package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

// Validate missing godoc
func (i FormationTemplateInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.MissingArtifactInfoMessage, validation.Required, validation.RuneLength(0, longLongStringLengthLimit)),
		validation.Field(&i.MissingArtifactWarningMessage, validation.Required, validation.RuneLength(0, longLongStringLengthLimit)),
		validation.Field(&i.RuntimeTypes, inputvalidation.Each(validation.Required, validation.RuneLength(0, longLongStringLengthLimit))),
		validation.Field(&i.ApplicationTypes, inputvalidation.Each(validation.Required, validation.RuneLength(0, longLongStringLengthLimit))))
}
