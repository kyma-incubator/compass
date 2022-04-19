package config

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

type Config struct {
	ExternalAddress       string `envconfig:"default=127.0.0.1:3000"`
	APIEndpoint           string `envconfig:"default=/graphql"`
	PlaygroundAPIEndpoint string `envconfig:"default=/graphql"`

	Log log.Config

	ServerTimeout time.Duration `envconfig:"default=100s"`

	CSRSubject struct {
		Country            string `envconfig:"default=PL"`
		Organization       string `envconfig:"default=Org"`
		OrganizationalUnit string `envconfig:"default=OrgUnit"`
		Locality           string `envconfig:"default=Locality"`
		Province           string `envconfig:"default=State"`
	}
	CertificateValidityTime time.Duration `envconfig:"default=2160h"`
	CASecret                struct {
		Name           string `envconfig:"default=kyma-integration/connector-service-app-ca"`
		CertificateKey string `envconfig:"default=ca.crt"`
		KeyKey         string `envconfig:"default=ca.key"`
	}
	RootCASecret struct {
		Name           string `envconfig:"optional"`
		CertificateKey string `envconfig:"optional"`
	}

	RevocationConfigMapName string `envconfig:"default=compass-system/revocations-Config"`

	DirectorURL                    string `envconfig:"default=127.0.0.1:3003"`
	CertificateSecuredConnectorURL string `envconfig:"default=https://compass-gateway-mtls.kyma.local"`
	KubernetesClient               struct {
		PollInteval time.Duration `envconfig:"default=2s"`
		PollTimeout time.Duration `envconfig:"default=1m"`
		Timeout     time.Duration `envconfig:"default=95s"`
	}

	OneTimeTokenURL             string
	HttpClientSkipSslValidation bool          `envconfig:"default=false"`
	HTTPClientTimeout           time.Duration `envconfig:"default=30s"`
}

func (c *Config) String() string {
	return fmt.Sprintf("ExternalAddress: %s, APIEndpoint: %s, "+
		"CSRSubjectCountry: %s, CSRSubjectOrganization: %s, CSRSubjectOrganizationalUnit: %s, "+
		"CSRSubjectLocality: %s, CSRSubjectProvince: %s, "+
		"CertificateValidityTime: %s, CASecretName: %s, CASecretCertificateKey: %s, CASecretKeyKey: %s, "+
		"RootCASecretName: %s, RootCASecretCertificateKey: %s, "+
		"CertificateSecuredConnectorURL: %s, "+
		"RevocationConfigMapName: %s, "+
		"DirectorURL: %s "+
		"KubernetesClientPollInteval: %s, KubernetesClientPollTimeout: %s"+
		"OneTimeTokenURL: %s, HTTPClienttimeout: %s",
		c.ExternalAddress, c.APIEndpoint,
		c.CSRSubject.Country, c.CSRSubject.Organization, c.CSRSubject.OrganizationalUnit,
		c.CSRSubject.Locality, c.CSRSubject.Province,
		c.CertificateValidityTime, c.CASecret.Name, c.CASecret.CertificateKey, c.CASecret.KeyKey,
		c.RootCASecret.Name, c.RootCASecret.CertificateKey,
		c.CertificateSecuredConnectorURL,
		c.RevocationConfigMapName,
		c.DirectorURL,
		c.KubernetesClient.PollInteval, c.KubernetesClient.PollTimeout,
		c.OneTimeTokenURL, c.HTTPClientTimeout)
}
