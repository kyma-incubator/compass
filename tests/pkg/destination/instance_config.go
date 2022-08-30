package destination

import (
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
	"github.com/pkg/errors"
)

// InstanceConfig is a service instance config
type InstanceConfig struct {
	ClientID     string
	ClientSecret string
	URL          string
	TokenURL     string
	Cert         string
	Key          string
}

// validate checks if all required fields are populated based on Oauth Mode.
// In the end, the error message is aggregated by joining all error messages.
func (i *InstanceConfig) validate(oauthMode oauth.AuthMode) error {
	errorMessages := make([]string, 0)

	if i.ClientID == "" {
		errorMessages = append(errorMessages, "Client ID is missing")
	}
	if i.TokenURL == "" {
		errorMessages = append(errorMessages, "Token URL is missing")
	}
	if i.URL == "" {
		errorMessages = append(errorMessages, "URL is missing")
	}

	switch oauthMode {
	case oauth.Standard:
		if i.ClientSecret == "" {
			errorMessages = append(errorMessages, "Client Secret is missing")
		}
	case oauth.Mtls:
		if i.Cert == "" {
			errorMessages = append(errorMessages, "Certificate is missing")
		}
		if i.Key == "" {
			errorMessages = append(errorMessages, "Key is missing")
		}
	}

	errorMsg := strings.Join(errorMessages, ", ")
	if errorMsg != "" {
		return errors.New(errorMsg)
	}

	return nil
}
