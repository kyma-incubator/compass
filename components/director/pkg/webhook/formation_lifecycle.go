package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/templatehelper"
)

// FormationLifecycleInput struct contains the input for FORMATION_LIFECYCLE webhook
type FormationLifecycleInput struct {
	Operation             model.FormationOperation
	Formation             *model.Formation
	CustomerTenantContext *CustomerTenantContext
}

// ParseURLTemplate parses the URL template
func (fl *FormationLifecycleInput) ParseURLTemplate(tmpl *string) (*URL, error) {
	var url URL
	return &url, templatehelper.ParseTemplate(tmpl, *fl, &url)
}

// ParseInputTemplate parses the input template
func (fl *FormationLifecycleInput) ParseInputTemplate(tmpl *string) ([]byte, error) {
	res := json.RawMessage{}
	if err := templatehelper.ParseTemplate(tmpl, *fl, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// ParseHeadersTemplate parses the headers template
func (fl *FormationLifecycleInput) ParseHeadersTemplate(tmpl *string) (http.Header, error) {
	var headers http.Header
	return headers, templatehelper.ParseTemplate(tmpl, *fl, &headers)
}
