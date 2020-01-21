package graphql

type TokenWithURL struct {
	Token        string `json:"token"`
	ConnectorURL string `json:"connectorURL"`
}

type OneTimeTokenForApplication struct {
	TokenWithURL
	LegacyConnectorURL string `json:"legacyConnectorURL"`
}

func (t *OneTimeTokenForApplication) IsOneTimeToken() {}

type OneTimeTokenForRuntime struct {
	TokenWithURL
}

func (t *OneTimeTokenForRuntime) IsOneTimeToken() {}

type OneTimeTokenForRuntimeExt struct {
	OneTimeTokenForRuntime
	Raw        string `json:"raw"`
	RawEncoded string `json:"rawEncoded"`
}

type OneTimeTokenForApplicationExt struct {
	OneTimeTokenForApplication
	Raw        string `json:"raw"`
	RawEncoded string `json:"rawEncoded"`
}
