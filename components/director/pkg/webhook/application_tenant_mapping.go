package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// ApplicationTenantMappingInput struct contains the input for an app-to-app formation notification
type ApplicationTenantMappingInput struct {
	Operation                 model.FormationOperation
	FormationID               string
	SourceApplicationTemplate *ApplicationTemplateWithLabels
	// SourceApplication is the application that the notification is about
	SourceApplication         *ApplicationWithLabels
	TargetApplicationTemplate *ApplicationTemplateWithLabels
	// TargetApplication is the application that the notification is for (the one with the webhook / the one receiving the notification)
	TargetApplication *ApplicationWithLabels
}

// ParseURLTemplate missing godoc
func (rd *ApplicationTenantMappingInput) ParseURLTemplate(tmpl *string) (*URL, error) {
	var url URL
	return &url, parseTemplate(tmpl, *rd, &url)
}

// ParseInputTemplate missing godoc
func (rd *ApplicationTenantMappingInput) ParseInputTemplate(tmpl *string) ([]byte, error) {
	res := json.RawMessage{}
	if err := parseTemplate(tmpl, *rd, &res); err != nil {
		return nil, err
	}
	res = bytes.ReplaceAll(res, []byte("<nil>"), nil)
	return res, nil
}

// ParseHeadersTemplate missing godoc
func (rd *ApplicationTenantMappingInput) ParseHeadersTemplate(tmpl *string) (http.Header, error) {
	var headers http.Header
	return headers, parseTemplate(tmpl, *rd, &headers)
}

// GetParticipantsIDs returns the list of IDs part of the FormationConfigurationChangeInput
func (rd *ApplicationTenantMappingInput) GetParticipantsIDs() []string {
	return []string{rd.SourceApplication.ID, rd.TargetApplication.ID}
}
