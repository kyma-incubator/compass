package oathkeeper

import "net/http"

const (
	ConnectorTokenHeader string = "Connector-Token"

	// TODO: we should agree on some meaningful name of those headers
	// TODO: we need to make sure that those headers will be always removed by Oathkeeper (so that the user cannot just specify them)
	ClientIdFromTokenHeader = "Client-Id-From-Token"
	TokenTypeHeader         = "Token-Type"

	ClientIdFromCertificateHeader = "Client-Id-From-Certificate"
	ClientCertificateHashHeader   = "Client-Certificate-Hash"
)

type AuthenticationSession struct {
	Subject string                 `json:"subject"`
	Extra   map[string]interface{} `json:"extra"`
	Header  http.Header            `json:"header"`
}

func (as *AuthenticationSession) TrimHeaders() {
	as.Header.Del(ClientIdFromTokenHeader)
	as.Header.Del(TokenTypeHeader)
	as.Header.Del(ClientIdFromCertificateHeader)
	as.Header.Del(ClientCertificateHashHeader)
}
