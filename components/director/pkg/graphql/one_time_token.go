package graphql

// TokenWithURL missing godoc
type TokenWithURL struct {
	Token        string           `json:"token"`
	ConnectorURL string           `json:"connectorURL"`
	Used         bool             `json:"used"`
	ExpiresAt    *Timestamp       `json:"expiresAt"`
	CreatedAt    *Timestamp       `json:"createdAt"`
	UsedAt       *Timestamp       `json:"usedAt"`
	Type         OneTimeTokenType `json:"type"`
}

// OneTimeTokenForApplication missing godoc
type OneTimeTokenForApplication struct {
	TokenWithURL
	LegacyConnectorURL string `json:"legacyConnectorURL"`
}

// IsOneTimeToken missing godoc
func (t *OneTimeTokenForApplication) IsOneTimeToken() {}

// OneTimeTokenForRuntime missing godoc
type OneTimeTokenForRuntime struct {
	TokenWithURL
}

// IsOneTimeToken missing godoc
func (t *OneTimeTokenForRuntime) IsOneTimeToken() {}

// OneTimeTokenForRuntimeExt missing godoc
type OneTimeTokenForRuntimeExt struct {
	OneTimeTokenForRuntime
	Raw        string `json:"raw"`
	RawEncoded string `json:"rawEncoded"`
}

// OneTimeTokenForApplicationExt missing godoc
type OneTimeTokenForApplicationExt struct {
	OneTimeTokenForApplication
	Raw        string `json:"raw"`
	RawEncoded string `json:"rawEncoded"`
}
