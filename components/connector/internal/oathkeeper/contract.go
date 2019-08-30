package oathkeeper

import "net/http"

const (
	// TODO: we should agree on some meaningful name of those headers
	// TODO: we need to make sure that those headers will be always removed by Oathkeeper (so that the user cannot just specify them)
	ClientIdFromTokenHeader = "Client-Id-From-Token"
	TokenTypeHeader         = "Token-Type"
)

type AuthenticationSession struct {
	Subject string                 `json:"subject"`
	Extra   map[string]interface{} `json:"extra"`
	Header  http.Header            `json:"header"`
}
