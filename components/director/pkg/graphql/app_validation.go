package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

// Validate missing godoc
func (i ApplicationRegisterInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, validation.RuneLength(1, appNameLengthLimit)),
		validation.Field(&i.ProviderName, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		validation.Field(&i.Labels, inputvalidation.EachKey(validation.Required, validation.Match(alphanumericUnderscoreRegexp))),
		validation.Field(&i.HealthCheckURL, inputvalidation.IsURL, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.Webhooks, validation.Each(validation.Required)))
}

// Validate missing godoc
func (i ApplicationJSONInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, validation.RuneLength(1, appNameLengthLimit)),
		validation.Field(&i.ProviderName, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		validation.Field(&i.Labels, inputvalidation.EachKey(validation.Required, validation.Match(alphanumericUnderscoreRegexp))),
		validation.Field(&i.HealthCheckURL, inputvalidation.IsURL, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.Webhooks, validation.Each(validation.Required)))
}

// Validate missing godoc
func (i ApplicationUpdateInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.ProviderName, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		validation.Field(&i.HealthCheckURL, inputvalidation.IsURL, validation.RuneLength(0, longStringLengthLimit)),
	)
}
