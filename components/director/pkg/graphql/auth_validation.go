package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

func (i AuthInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.AdditionalHeaders,
			inputvalidation.EachKey(validation.Required),                                         // key
			inputvalidation.Each(validation.Required, inputvalidation.Each(validation.Required)), // value
		),
		validation.Field(&i.AdditionalQueryParams,
			inputvalidation.EachKey(validation.Required),                                         // key
			inputvalidation.Each(validation.Required, inputvalidation.Each(validation.Required)), // value
		),
		validation.Field(&i.Credential, validation.Required),
		validation.Field(&i.RequestAuth),
	)
}

func (i CredentialDataInput) Validate() error {
	return validation.Errors{
		"rule.ExactlyOneNotNil": inputvalidation.ValidateExactlyOneNotNil(
			"exactly one credential input has to be specified",
			i.Basic, i.Oauth,
		),
		"Basic": validation.Validate(i.Basic),
		"Oauth": validation.Validate(i.Oauth),
	}.Filter()
}

func (i BasicCredentialDataInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Username, validation.Required),
		validation.Field(&i.Password, validation.Required),
	)
}

func (i OAuthCredentialDataInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.ClientID, validation.Required),
		validation.Field(&i.ClientSecret, validation.Required),
		validation.Field(&i.URL, validation.Required, is.URL),
	)
}

func (i CredentialRequestAuthInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Csrf, validation.Required),
	)
}

func (i CSRFTokenCredentialRequestAuthInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.AdditionalHeaders,
			inputvalidation.EachKey(validation.Required),                                         // key
			inputvalidation.Each(validation.Required, inputvalidation.Each(validation.Required)), // value
		),
		validation.Field(&i.AdditionalQueryParams,
			inputvalidation.EachKey(validation.Required),                                         // key
			inputvalidation.Each(validation.Required, inputvalidation.Each(validation.Required)), // value
		),
		validation.Field(&i.TokenEndpointURL, validation.Required, is.URL),
		validation.Field(&i.Credential, validation.Required),
	)
}
