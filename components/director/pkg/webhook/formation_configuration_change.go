package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type ApplicationWithLabels struct {
	model.Application
	Labels map[string]interface{}
}

type ApplicationTemplateWithLabels struct {
	model.ApplicationTemplate
	Labels map[string]interface{}
}

type RuntimeWithLabels struct {
	model.Runtime
	Labels map[string]interface{}
}

type RuntimeContextWithLabels struct {
	model.RuntimeContext
	Labels map[string]interface{}
}

// FormationConfigurationChangeInput struct contains the input for a formation notification
type FormationConfigurationChangeInput struct {
	Operation           model.FormationOperation
	FormationID         string
	ApplicationTemplate ApplicationTemplateWithLabels
	Application         ApplicationWithLabels
	Runtime             RuntimeWithLabels
	RuntimeContext      RuntimeContextWithLabels
}

// ParseURLTemplate missing godoc
func (rd *FormationConfigurationChangeInput) ParseURLTemplate(tmpl *string) (*URL, error) {
	var url URL
	return &url, parseTemplate(tmpl, *rd, &url)
}

// ParseInputTemplate missing godoc
func (rd *FormationConfigurationChangeInput) ParseInputTemplate(tmpl *string) ([]byte, error) {
	res := json.RawMessage{}
	return res, parseTemplate(tmpl, *rd, &res)
}

// ParseHeadersTemplate missing godoc
func (rd *FormationConfigurationChangeInput) ParseHeadersTemplate(tmpl *string) (http.Header, error) {
	var headers http.Header
	return headers, parseTemplate(tmpl, *rd, &headers)
}
