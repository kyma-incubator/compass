package config

import (
	"fmt"
	"time"
)

type Config struct {
	ExternalAddress       string `envconfig:"default=127.0.0.1:3000"`
	InternalAddress       string `envconfig:"default=127.0.0.1:3001"`
	APIEndpoint           string `envconfig:"default=/graphql"`
	PlaygroundAPIEndpoint string `envconfig:"default=/graphql"`

	HydratorAddress string `envconfig:"default=127.0.0.1:8080"`

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

	CertificateDataHeader   string `envconfig:"default=Certificate-Data"`
	RevocationConfigMapName string `envconfig:"default=compass-system/revocations-Config"`

	Token struct {
		Length                int           `envconfig:"default=64"`
		RuntimeExpiration     time.Duration `envconfig:"default=60m"`
		ApplicationExpiration time.Duration `envconfig:"default=5m"`
		CSRExpiration         time.Duration `envconfig:"default=5m"`
	}

	DirectorURL                    string `envconfig:"default=127.0.0.1:3003"`
	CertificateSecuredConnectorURL string `envconfig:"default=https://compass-gateway-mtls.kyma.local"`
	KubernetesClient               struct {
		PollInteval time.Duration `envconfig:"default=2s"`
		PollTimeout time.Duration `envconfig:"default=1m"`
	}
}

func (c *Config) String() string {
	return fmt.Sprintf("ExternalAddress: %s, InternalAddress: %s, APIEndpoint: %s, HydratorAddress: %s, "+
		"CSRSubjectCountry: %s, CSRSubjectOrganization: %s, CSRSubjectOrganizationalUnit: %s, "+
		"CSRSubjectLocality: %s, CSRSubjectProvince: %s, "+
		"CertificateValidityTime: %s, CASecretName: %s, CASecretCertificateKey: %s, CASecretKeyKey: %s, "+
		"RootCASecretName: %s, RootCASecretCertificateKey: %s, CertificateDataHeader: %s, "+
		"CertificateSecuredConnectorURL: %s, "+
		"RevocationConfigMapName: %s, "+
		"TokenLength: %d, TokenRuntimeExpiration: %s, TokenApplicationExpiration: %s, TokenCSRExpiration: %s, "+
		"DirectorURL: %s "+
		"KubernetesClientPollInteval: %s, KubernetesClientPollTimeout: %s",
		c.ExternalAddress, c.InternalAddress, c.APIEndpoint, c.HydratorAddress,
		c.CSRSubject.Country, c.CSRSubject.Organization, c.CSRSubject.OrganizationalUnit,
		c.CSRSubject.Locality, c.CSRSubject.Province,
		c.CertificateValidityTime, c.CASecret.Name, c.CASecret.CertificateKey, c.CASecret.KeyKey,
		c.RootCASecret.Name, c.RootCASecret.CertificateKey, c.CertificateDataHeader,
		c.CertificateSecuredConnectorURL,
		c.RevocationConfigMapName,
		c.Token.Length, c.Token.RuntimeExpiration.String(), c.Token.ApplicationExpiration.String(), c.Token.CSRExpiration.String(),
		c.DirectorURL,
		c.KubernetesClient.PollInteval, c.KubernetesClient.PollTimeout)
}
