package selfregmanager

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/securehttp"
	authpkg "github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
	"github.com/pkg/errors"
)

// CallerProvider is used to provide ExternalSvcCaller to call external services with given authentication
type CallerProvider struct{}

// GetCaller provides ExternalSvcCaller to call external services with given authentication
func (c *CallerProvider) GetCaller(config config.SelfRegConfig, region string) (ExternalSvcCaller, error) {
	instanceConfig, exists := config.RegionToInstanceConfig[region]
	if !exists {
		return nil, errors.Errorf("missing configuration for region: %s", region)
	}

	var credentials authpkg.Credentials
	if config.OAuthMode == oauth.Standard {
		credentials = &authpkg.OAuthCredentials{
			ClientID:     instanceConfig.ClientID,
			ClientSecret: instanceConfig.ClientSecret,
			TokenURL:     instanceConfig.URL + config.OauthTokenPath,
		}
	} else if config.OAuthMode == oauth.Mtls {
		mtlsCredentials, err := authpkg.NewOAuthMtlsCredentials(instanceConfig.ClientID, instanceConfig.Cert, instanceConfig.Key, instanceConfig.TokenURL, config.OauthTokenPath)
		if err != nil {
			return nil, errors.Wrap(err, "while creating OAuth Mtls credentials")
		}
		credentials = mtlsCredentials
	} else {
		return nil, errors.New(fmt.Sprintf("unsupported OAuth mode: %s", config.OAuthMode))
	}

	callerConfig := securehttp.CallerConfig{
		Credentials:       credentials,
		ClientTimeout:     config.ClientTimeout,
		SkipSSLValidation: config.SkipSSLValidation,
	}
	caller, err := securehttp.NewCaller(callerConfig)
	if err != nil {
		return nil, err
	}

	return caller, nil
}
