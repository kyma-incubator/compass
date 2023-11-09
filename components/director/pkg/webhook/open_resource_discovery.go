package webhook

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/kyma-incubator/compass/components/director/pkg/templatehelper"
)

// Application represents the fields needed for determining the Template's values
type Application struct {
	BaseURL string `json:"BaseUrl"`
}

// OpenResourceDiscoveryWebhookRequestObject struct contains parts of request that might be needed for later processing of a Webhook request
type OpenResourceDiscoveryWebhookRequestObject struct {
	Application Application
	TenantID    string
	Headers     *sync.Map
}

// ParseURLTemplate missing godoc
func (rd *OpenResourceDiscoveryWebhookRequestObject) ParseURLTemplate(tmpl *string) (*URL, error) {
	var url URL
	return &url, templatehelper.ParseTemplate(tmpl, *rd, &url)
}

// ParseInputTemplate missing godoc
func (rd *OpenResourceDiscoveryWebhookRequestObject) ParseInputTemplate(tmpl *string) ([]byte, error) {
	res := json.RawMessage{}
	return res, templatehelper.ParseTemplate(tmpl, *rd, &res)
}

// ParseHeadersTemplate missing godoc
func (rd *OpenResourceDiscoveryWebhookRequestObject) ParseHeadersTemplate(tmpl *string) (http.Header, error) {
	var headers http.Header
	return headers, templatehelper.ParseTemplate(tmpl, rd, &headers)
}
