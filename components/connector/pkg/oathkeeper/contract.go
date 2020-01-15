package oathkeeper

import "net/http"

const (
	ConnectorTokenHeader string = "Connector-Token"

	ConnectorTokenQueryParam string = "token"

	ClientIdFromTokenHeader       = "Client-Id-From-Token"
	ClientIdFromCertificateHeader = "Client-Id-From-Certificate"
	ClientCertificateHashHeader   = "Client-Certificate-Hash"
)

type AuthenticationSession struct {
	Subject string                 `json:"subject"`
	Extra   map[string]interface{} `json:"extra"`
	Header  http.Header            `json:"header"`
}
