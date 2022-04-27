package config

import "github.com/pkg/errors"

type HandlerConfig struct {
	RuntimeMappingEndpoint        string `envconfig:"default=/runtime-mapping"`
	AuthenticationMappingEndpoint string `envconfig:"default=/authn-mapping/{authenticator}"`
	TenantMappingEndpoint         string `envconfig:"default=/tenant-mapping"`
	CertResolverEndpoint          string `envconfig:"default=/v1/certificate/data/resolve"`
	TokenResolverEndpoint         string `envconfig:"default=/v1/tokens/resolve"`
}

// Validate ensures the constructed Config contains valid property values
func (c *HandlerConfig) Validate() error {
	if c.RuntimeMappingEndpoint == "" || c.AuthenticationMappingEndpoint == "" || c.TenantMappingEndpoint == "" || c.CertResolverEndpoint == "" || c.TokenResolverEndpoint == "" {
		return errors.New("Missing handler configuration")
	}

	return nil
}
