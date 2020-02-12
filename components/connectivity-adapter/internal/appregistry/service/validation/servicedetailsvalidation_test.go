/**
Copied from https://github.com/kyma-project/kyma/tree/master/components/application-registry
*/
package validation

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"

	"github.com/stretchr/testify/assert"
)

var (
	eventsRawSpec = compact([]byte("{\"name\":\"events\"}"))
)

func compact(src []byte) []byte {
	buffer := new(bytes.Buffer)
	err := json.Compact(buffer, src)
	if err != nil {
		return src
	}
	return buffer.Bytes()
}

func TestServiceDetailsValidator(t *testing.T) {
	t.Run("should accept service details with API", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.NoError(t, err)
	})

	t.Run("should accept service details with events", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Events: &model.Events{
				Spec: eventsRawSpec,
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.NoError(t, err)
	})

	t.Run("should accept service details with API and events", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
			},
			Events: &model.Events{
				Spec: eventsRawSpec,
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.NoError(t, err)
	})

	t.Run("should not accept service details without API and Events", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should not accept service details without name", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should not accept service details without provider", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should not accept service details without description", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:     "name",
			Provider: "provider",
			Api: &model.API{
				TargetUrl: "http://target.com",
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})
}

func TestServiceDetailsValidator_API(t *testing.T) {
	t.Run("should not accept API without targetUrl", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api:         &model.API{},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should not accept API spec with more than 1 type of auth", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
				Credentials: &model.CredentialsWithCSRF{
					BasicWithCSRF: &model.BasicAuthWithCSRF{
						BasicAuth: model.BasicAuth{
							Username: "username",
							Password: "password",
						},
					},
					OauthWithCSRF: &model.OauthWithCSRF{
						Oauth: model.Oauth{
							URL:          "http://test.com/token",
							ClientID:     "client",
							ClientSecret: "secret",
						},
					},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})
}

func TestServiceDetailsValidator_API_OAuth(t *testing.T) {
	t.Run("should accept OAuth credentials", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
				Credentials: &model.CredentialsWithCSRF{
					OauthWithCSRF: &model.OauthWithCSRF{
						Oauth: model.Oauth{
							URL:          "http://test.com/token",
							ClientID:     "client",
							ClientSecret: "secret",
						},
					},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.NoError(t, err)
	})

	t.Run("should not accept OAuth credentials with empty oauth", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
				Credentials: &model.CredentialsWithCSRF{
					OauthWithCSRF: &model.OauthWithCSRF{},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should not accept OAuth credentials with incomplete oauth", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
				Credentials: &model.CredentialsWithCSRF{
					OauthWithCSRF: &model.OauthWithCSRF{
						Oauth: model.Oauth{
							URL:      "http://test.com/token",
							ClientID: "client",
						},
					},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should not accept OAuth credentials with wrong oauth url", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
				Credentials: &model.CredentialsWithCSRF{
					OauthWithCSRF: &model.OauthWithCSRF{
						Oauth: model.Oauth{
							URL:          "test_com/token",
							ClientID:     "client",
							ClientSecret: "secret",
						},
					},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})
}

func TestServiceDetailsValidator_API_Basic(t *testing.T) {
	t.Run("should accept Basic Auth credentials", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
				Credentials: &model.CredentialsWithCSRF{
					BasicWithCSRF: &model.BasicAuthWithCSRF{
						BasicAuth: model.BasicAuth{
							Username: "username",
							Password: "password",
						},
					},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.NoError(t, err)
	})

	t.Run("should not accept Basic Auth credentials with empty basic", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
				Credentials: &model.CredentialsWithCSRF{
					BasicWithCSRF: &model.BasicAuthWithCSRF{},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should not accept Basic Auth credentials with incomplete basic", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
				Credentials: &model.CredentialsWithCSRF{
					BasicWithCSRF: &model.BasicAuthWithCSRF{
						BasicAuth: model.BasicAuth{
							Username: "username",
						},
					},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})
}

func TestServiceDetailsValidator_API_Certificate(t *testing.T) {
	t.Run("should accept Certificate credentials", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
				Credentials: &model.CredentialsWithCSRF{
					CertificateGenWithCSRF: &model.CertificateGenWithCSRF{},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.NoError(t, err)
	})
}

func TestServiceDetailsValidator_Specification_OAuth(t *testing.T) {
	t.Run("should accept OAuth specification credentials", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
				SpecificationCredentials: &model.Credentials{
					Oauth: &model.Oauth{
						URL:          "http://test.com/token",
						ClientID:     "client",
						ClientSecret: "secret",
					},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.NoError(t, err)
	})

	t.Run("should not accept OAuth specification credentials with empty oauth", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
				SpecificationCredentials: &model.Credentials{
					Oauth: &model.Oauth{},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should not accept OAuth specification credentials with incomplete oauth", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
				SpecificationCredentials: &model.Credentials{
					Oauth: &model.Oauth{
						URL:      "http://test.com/token",
						ClientID: "client",
					},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should not accept OAuth specification credentials with wrong oauth url", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
				SpecificationCredentials: &model.Credentials{
					Oauth: &model.Oauth{
						URL:          "test_com/token",
						ClientID:     "client",
						ClientSecret: "secret",
					},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})
}

func TestServiceDetailsValidator_Specification_Basic(t *testing.T) {
	t.Run("should accept Basic Auth specification credentials", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
				SpecificationCredentials: &model.Credentials{
					Basic: &model.BasicAuth{
						Username: "username",
						Password: "password",
					},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.NoError(t, err)
	})

	t.Run("should not accept Basic Auth specification credentials with empty basic", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
				SpecificationCredentials: &model.Credentials{
					Basic: &model.BasicAuth{},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should not accept Basic Auth specification credentials with incomplete basic", func(t *testing.T) {
		// given
		serviceDetails := model.ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
				SpecificationCredentials: &model.Credentials{
					Basic: &model.BasicAuth{
						Username: "username",
					},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})
}
