package config

type IstioConfig struct {
	CompassGatewayURL     string `envconfig:"default=compass-gateway.kyma.local"`
	CompassMTLSGatewayURL string `envconfig:"default=compass-gateway-mtls.kyma.local"`
	RequestPayloadLimit   int    `envconfig:"default=2097152"` //2 MB
}
