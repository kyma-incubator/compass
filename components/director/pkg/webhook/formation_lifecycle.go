package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// FormationLifecycleInput struct contains the input for FORMATION_LIFECYCLE webhook
type FormationLifecycleInput struct {
	Formation             *model.Formation
	CustomerTenantContext *CustomerTenantContext
}

// ParseURLTemplate parses the URL template
func (fl *FormationLifecycleInput) ParseURLTemplate(tmpl *string) (*URL, error) {
	var url URL
	return &url, parseTemplate(tmpl, *fl, &url)
}

// ParseInputTemplate parses the input template
func (fl *FormationLifecycleInput) ParseInputTemplate(tmpl *string) ([]byte, error) {
	res := json.RawMessage{}
	if err := parseTemplate(tmpl, *fl, &res); err != nil {
		return nil, err
	}
	res = bytes.ReplaceAll(res, []byte("<nil>"), nil)
	return res, nil
}

// ParseHeadersTemplate parses the headers template
func (fl *FormationLifecycleInput) ParseHeadersTemplate(tmpl *string) (http.Header, error) {
	var headers http.Header
	return headers, parseTemplate(tmpl, *fl, &headers)
}
