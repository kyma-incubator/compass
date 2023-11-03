package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/templatehelper"
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
	return &url, templatehelper.ParseTemplate(tmpl, *rd, &url)
}

// ParseInputTemplate missing godoc
func (rd *ApplicationLifecycleWebhookRequestObject) ParseInputTemplate(tmpl *string) ([]byte, error) {
	res := json.RawMessage{}
	return res, templatehelper.ParseTemplate(tmpl, *rd, &res)
}

// ParseHeadersTemplate missing godoc
func (rd *ApplicationLifecycleWebhookRequestObject) ParseHeadersTemplate(tmpl *string) (http.Header, error) {
	var headers http.Header
	return headers, templatehelper.ParseTemplate(tmpl, *rd, &headers)
}
