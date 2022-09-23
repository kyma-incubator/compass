package config

import "github.com/kyma-incubator/compass/components/director/pkg/certloader"

type IstioConfig struct {
	CompassGatewayURL              string `envconfig:"default=compass-gateway.kyma.local"`
	CompassMTLSGatewayURL          string `envconfig:"default=compass-gateway-mtls.kyma.local"`
	DirectorExternalCertSecuredURL string `envconfig:"default=http://compass-director-external-mtls.compass-system.svc.cluster.local:3000/graphql"`
	RequestPayloadLimit            int    `envconfig:"default=2097152"` //2 MB
	SkipSSLValidation              bool   `envconfig:"default=false"`
	CertLoaderConfig               certloader.Config
	DefaultTenant                  string
	ExternalClientCertSecretName string `envconfig:"EXTERNAL_CLIENT_CERT_SECRET_NAME"`
}
