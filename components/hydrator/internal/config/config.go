package config

import "github.com/pkg/errors"

type HandlerConfig struct {
	TenantMappingEndpoint         string `envconfig:"default=/tenant-mapping"`
	AuthenticationMappingEndpoint string `envconfig:"default=/authn-mapping/{authenticator}"`
	RuntimeMappingEndpoint        string `envconfig:"default=/runtime-mapping"`
	ValidationIstioCertEndpoint   string `envconfig:"default=/v1/certificate/data/resolve"`
	TokenResolverEndpoint         string `envconfig:"default=/v1/tokens/resolve"`
}

// Validate ensures the constructed Config contains valid property values
func (c *HandlerConfig) Validate() error {
	if c.AuthenticationMappingEndpoint == "" || c.TenantMappingEndpoint == "" || c.ValidationIstioCertEndpoint == "" || c.RuntimeMappingEndpoint == "" || c.TokenResolverEndpoint == "" {
		return errors.New("Missing handler configuration")
	}

	return nil
}
