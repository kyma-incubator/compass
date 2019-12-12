package graphql

type OneTimeToken struct {
	Token        string `json:"token"`
	ConnectorURL string `json:"connectorURL"`
}
