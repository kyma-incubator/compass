package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

// Validate missing godoc
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
		validation.Field(&i.Credential, validation.NilOrNotEmpty),
		validation.Field(&i.RequestAuth),
	)
}

// Validate missing godoc
func (i CredentialDataInput) Validate() error {
	return validation.Errors{
		"Rule.ExactlyOneNotNil": inputvalidation.ValidateExactlyOneNotNil(
			"exactly one credential input has to be specified",
			i.Basic, i.Oauth, i.CertificateOAuth,
		),
		"Basic":            validation.Validate(i.Basic),
		"Oauth":            validation.Validate(i.Oauth),
		"CertificateOAuth": validation.Validate(i.CertificateOAuth),
	}.Filter()
}

// Validate missing godoc
func (i BasicCredentialDataInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Username, validation.Required),
		validation.Field(&i.Password, validation.Required),
	)
}

// Validate missing godoc
func (i OAuthCredentialDataInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.ClientID, validation.Required),
		validation.Field(&i.ClientSecret, validation.Required),
		validation.Field(&i.URL, validation.Required, is.URL),
	)
}

// Validate missing godoc
func (i CertificateOAuthCredentialDataInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.ClientID, validation.Required),
		validation.Field(&i.Certificate, validation.Required),
		validation.Field(&i.URL, validation.Required, is.URL),
	)
}

// Validate missing godoc
func (i CredentialRequestAuthInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Csrf, validation.Required),
	)
}

// Validate missing godoc
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
		validation.Field(&i.Credential, validation.NilOrNotEmpty),
	)
}
