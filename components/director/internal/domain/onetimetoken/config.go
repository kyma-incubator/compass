package onetimetoken

type Config struct {
	//One time token URL
	OneTimeTokenURL string `envconfig:"APP_ONE_TIME_TOKEN_URL"`
	//Connector URL
	ConnectorURL string `envconfig:"APP_CONNECTOR_URL"`
}
