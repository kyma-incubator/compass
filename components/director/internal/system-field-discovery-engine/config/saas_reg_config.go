package config

import (
	"strings"

	"github.com/pkg/errors"
)

// SaasRegConfig is a saas reg config
type SaasRegConfig struct {
	ClientID        string
	ClientSecret    string
	TokenURL        string
	SaasRegistryURL string
}

// validate checks if all required fields are populated.
// In the end, the error message is aggregated by joining all error messages.
func (s *SaasRegConfig) validate() error {
	errorMessages := make([]string, 0)

	if s.ClientID == "" {
		errorMessages = append(errorMessages, "Client ID is missing")
	}
	if s.ClientSecret == "" {
		errorMessages = append(errorMessages, "Client Secret is missing")
	}
	if s.TokenURL == "" {
		errorMessages = append(errorMessages, "Token URL is missing")
	}
	if s.SaasRegistryURL == "" {
		errorMessages = append(errorMessages, "Saas Registry URL is missing")
	}

	errorMsg := strings.Join(errorMessages, ", ")
	if errorMsg != "" {
		return errors.New(errorMsg)
	}

	return nil
}
