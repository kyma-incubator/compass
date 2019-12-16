package graphql

type OneTimeToken struct {
	Token        string `json:"token"`
	ConnectorURL string `json:"connectorURL"`
}

type OneTimeTokenExt struct {
	OneTimeToken
	Raw        string `json:"raw"`
	RawEncoded string `json:"rawEncoded"`
}
