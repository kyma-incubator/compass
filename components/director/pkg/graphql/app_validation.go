package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/kyma-incubator/compass/components/director/pkg/customerrors"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

func (i ApplicationRegisterInput) Validate() error {
	err := validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, inputvalidation.DNSName),
		validation.Field(&i.ProviderName, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		validation.Field(&i.Labels, inputvalidation.EachKey(validation.Required, validation.Match(alphanumericUnderscoreRegexp))),
		validation.Field(&i.HealthCheckURL, inputvalidation.IsURL, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.Webhooks, validation.Each(validation.Required)),
	)
	if err != nil {
		return customerrors.GraphqlError{
			StatusCode: customerrors.InvalidData,
			Message:    err.Error(),
		}
	}
	return nil
}

func (i ApplicationUpdateInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.ProviderName, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		validation.Field(&i.HealthCheckURL, inputvalidation.IsURL, validation.RuneLength(0, longStringLengthLimit)),
	)
}
