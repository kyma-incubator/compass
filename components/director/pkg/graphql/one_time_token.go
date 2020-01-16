package graphql

type OneTimeTokenForApplication struct {
	Token              string `json:"token"`
	ConnectorURL       string `json:"connectorURL"`
	LegacyConnectorURL string `json:"legacyConnectorURL"`
}

func (t *OneTimeTokenForApplication) IsOneTimeToken() {}

type OneTimeTokenForRuntime struct {
	Token        string `json:"token"`
	ConnectorURL string `json:"connectorURL"`
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
