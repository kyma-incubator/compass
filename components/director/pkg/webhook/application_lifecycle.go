package webhook

import (
	"encoding/json"
	"net/http"
)

// Resource is used to identify entities which can be part of a webhook's request data
type Resource interface {
	Sentinel()
}

// ApplicationLifecycleWebhookRequestObject struct contains parts of request that might be needed for later processing of a Webhook request
type ApplicationLifecycleWebhookRequestObject struct {
	Application Resource
	TenantID    string
	Headers     map[string]string
}

// ParseURLTemplate missing godoc
func (rd *ApplicationLifecycleWebhookRequestObject) ParseURLTemplate(tmpl *string) (*URL, error) {
	var url URL
	return &url, parseTemplate(tmpl, *rd, &url)
}

// ParseInputTemplate missing godoc
func (rd *ApplicationLifecycleWebhookRequestObject) ParseInputTemplate(tmpl *string) ([]byte, error) {
	res := json.RawMessage{}
	return res, parseTemplate(tmpl, *rd, &res)
}

// ParseHeadersTemplate missing godoc
func (rd *ApplicationLifecycleWebhookRequestObject) ParseHeadersTemplate(tmpl *string) (http.Header, error) {
	var headers http.Header
	return headers, parseTemplate(tmpl, *rd, &headers)
}

// GetParticipants missing godoc
func (rd *ApplicationLifecycleWebhookRequestObject) GetParticipants() []string {
	return []string{}
}
