package client

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/securehttp"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/config"
	"github.com/pkg/errors"
)

// ExternalSvcCaller is used to call external services with given authentication
//
//go:generate mockery --name=ExternalSvcCaller --output=automock --outpkg=automock --case=underscore --disable-version-string
type ExternalSvcCaller interface {
	Call(*http.Request) (*http.Response, error)
}

// ExternalSvcCallerProvider provides ExternalSvcCaller based on the provided config and region
//
//go:generate mockery --name=ExternalSvcCallerProvider --output=automock --outpkg=automock --case=underscore --disable-version-string
type ExternalSvcCallerProvider interface {
	GetCaller(cfg config.Config, region string) (ExternalSvcCaller, error)
}

// CallerProvider is used to provide ExternalSvcCaller to call external services with given authentication
type CallerProvider struct{}

// NewCallerProvider creates new CallerProvider
func NewCallerProvider() *CallerProvider {
	return &CallerProvider{}
}

// GetCaller provides ExternalSvcCaller to call external services with given authentication
func (c *CallerProvider) GetCaller(config config.Config, region string) (ExternalSvcCaller, error) {
	instanceConfig, exists := config.RegionToInstanceConfig[region]
	if !exists {
		return nil, errors.Errorf("missing configuration for region: %s", region)
	}

	mtlsCredentials, err := auth.NewOAuthMtlsCredentials(instanceConfig.ClientID, instanceConfig.Certificate, instanceConfig.CertificateKey, instanceConfig.TokenURL, config.OAuthTokenPath, config.ExternalClientCertSecretName)
	if err != nil {
		return nil, errors.Wrap(err, "while creating OAuth Mtls credentials")
	}

	callerConfig := securehttp.CallerConfig{
		Credentials:                  mtlsCredentials,
		ClientTimeout:                config.ClientTimeout,
		SkipSSLValidation:            config.SkipSSLValidation,
		ExternalClientCertSecretName: config.ExternalClientCertSecretName,
	}

	caller, err := securehttp.NewCaller(callerConfig)
	if err != nil {
		return nil, err
	}

	return caller, nil
}
