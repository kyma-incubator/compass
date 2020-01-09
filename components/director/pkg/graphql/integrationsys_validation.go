package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

func (i IntegrationSystemInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, inputvalidation.Name),
		validation.Field(&i.Description, validation.RuneLength(0, shortStringLengthLimit)),
	)
}
