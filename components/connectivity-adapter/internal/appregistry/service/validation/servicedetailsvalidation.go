/**
Copied from https://github.com/kyma-project/kyma/tree/master/components/application-registry
*/
package validation

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"

	"github.com/asaskevich/govalidator"
)

type ServiceDetailsValidator interface {
	Validate(details model.ServiceDetails) apperrors.AppError
}

type ServiceDetailsValidatorFunc func(details model.ServiceDetails) apperrors.AppError

func (f ServiceDetailsValidatorFunc) Validate(details model.ServiceDetails) apperrors.AppError {
	return f(details)
}

func NewServiceDetailsValidator() ServiceDetailsValidator {
	return ServiceDetailsValidatorFunc(func(details model.ServiceDetails) apperrors.AppError {
		_, err := govalidator.ValidateStruct(details)
		if err != nil {
			return apperrors.WrongInput("Incorrect structure of service definition, %s", err.Error())
		}

		if details.Api == nil && details.Events == nil {
			return apperrors.WrongInput(
				"At least one of service definition attributes: 'api', 'events' have to be provided")
		}

		var apperr apperrors.AppError

		if details.Api != nil {
			apperr := validateApiSpec(details.Api.Spec)
			if apperr != nil {
				return apperr
			}

			apperr = validateApiCredentials(details.Api.Credentials)
			if apperr != nil {
				return apperr
			}

			apperr = validateSpecificationCredentials(details.Api.SpecificationCredentials)
			if apperr != nil {
				return apperr
			}
		}

		apperr = validateEventsSpec(details.Events)
		if apperr != nil {
			return apperr
		}

		return nil
	})
}

func validateApiSpec(spec json.RawMessage) apperrors.AppError {
	if spec != nil {
		err := validateSpec(spec)
		if err != nil {
			return apperrors.WrongInput("api.Spec is not a proper json object, %s", err.Error())
		}
	}

	return nil
}

func validateEventsSpec(events *model.Events) apperrors.AppError {
	if events != nil && events.Spec != nil {
		err := validateSpec(events.Spec)
		if err != nil {
			return apperrors.WrongInput("events.Spec is not a proper json object, %s", err.Error())
		}
	}

	return nil
}

func validateApiCredentials(credentials *model.CredentialsWithCSRF) apperrors.AppError {
	if credentials != nil {
		var basic *model.BasicAuth
		var oauth *model.Oauth
		var cert *model.CertificateGen

		if credentials.BasicWithCSRF != nil {
			basic = &credentials.BasicWithCSRF.BasicAuth
		}

		if credentials.OauthWithCSRF != nil {
			oauth = &credentials.OauthWithCSRF.Oauth
		}

		if credentials.CertificateGenWithCSRF != nil {
			cert = &credentials.CertificateGenWithCSRF.CertificateGen
		}

		if validateCredentials(basic, oauth, cert) {
			return apperrors.WrongInput("api.CredentialsWithCSRF is invalid: to many authentication methods provided")
		}
	}

	return nil
}

func validateSpecificationCredentials(credentials *model.Credentials) apperrors.AppError {
	if credentials != nil {
		basic := credentials.Basic
		oauth := credentials.Oauth

		if validateCredentials(basic, oauth, nil) {
			return apperrors.WrongInput("api.CredentialsWithCSRF is invalid: to many authentication methods provided")
		}
	}

	return nil
}

func validateSpec(rawMessage json.RawMessage) error {
	var m map[string]*json.RawMessage
	return json.Unmarshal(rawMessage, &m)
}

func validateCredentials(basic *model.BasicAuth, oauth *model.Oauth, cert *model.CertificateGen) bool {
	credentialsCount := 0

	if basic != nil {
		credentialsCount++
	}

	if oauth != nil {
		credentialsCount++
	}

	if cert != nil {
		credentialsCount++
	}

	return credentialsCount > 1
}
