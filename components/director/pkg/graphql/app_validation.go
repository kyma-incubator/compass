package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

func (i ApplicationCreateInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, inputvalidation.Name),
		validation.Field(&i.Description, validation.RuneLength(0, shortStringLengthLimit)),
		validation.Field(&i.Labels, inputvalidation.EachKey(validation.Required)),
		validation.Field(&i.HealthCheckURL, inputvalidation.IsURL, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.Webhooks, validation.Each(validation.Required)),
		validation.Field(&i.Apis, inputvalidation.Each(validation.Required)),
		validation.Field(&i.EventAPIs, inputvalidation.Each(validation.Required)),
		validation.Field(&i.Documents, inputvalidation.Each(validation.Required)),
	)
}

func (i ApplicationUpdateInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, inputvalidation.Name),
		validation.Field(&i.Description, validation.RuneLength(0, shortStringLengthLimit)),
		validation.Field(&i.HealthCheckURL, inputvalidation.IsURL, validation.RuneLength(0, longStringLengthLimit)),
	)
}
