package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/templatehelper"
)

// ApplicationTenantMappingInput struct contains the input for an app-to-app formation notification
type ApplicationTenantMappingInput struct {
	Operation                 model.FormationOperation
	FormationID               string
	Formation                 *model.Formation
	SourceApplicationTemplate *ApplicationTemplateWithLabels
	// SourceApplication is the application that the notification is about
	SourceApplication         *ApplicationWithLabels
	TargetApplicationTemplate *ApplicationTemplateWithLabels
	// TargetApplication is the application that the notification is for (the one with the webhook / the one receiving the notification)
	TargetApplication     *ApplicationWithLabels
	CustomerTenantContext *CustomerTenantContext
	Assignment            *FormationAssignment
	ReverseAssignment     *FormationAssignment
}

// ParseURLTemplate missing godoc
func (rd *ApplicationTenantMappingInput) ParseURLTemplate(tmpl *string) (*URL, error) {
	var url URL
	return &url, templatehelper.ParseTemplate(tmpl, *rd, &url)
}

// ParseInputTemplate missing godoc
func (rd *ApplicationTenantMappingInput) ParseInputTemplate(tmpl *string) ([]byte, error) {
	res := json.RawMessage{}
	if err := templatehelper.ParseTemplate(tmpl, *rd, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// ParseHeadersTemplate missing godoc
func (rd *ApplicationTenantMappingInput) ParseHeadersTemplate(tmpl *string) (http.Header, error) {
	var headers http.Header
	return headers, templatehelper.ParseTemplate(tmpl, *rd, &headers)
}

// GetParticipantsIDs returns the list of IDs part of the FormationConfigurationChangeInput
func (rd *ApplicationTenantMappingInput) GetParticipantsIDs() []string {
	return []string{rd.SourceApplication.ID, rd.TargetApplication.ID}
}

// SetAssignment sets the assignment for the ApplicationTenantMappingInput to the provided one
func (rd *ApplicationTenantMappingInput) SetAssignment(assignment *model.FormationAssignment) {
	rd.Assignment = &FormationAssignment{
		ID:          assignment.ID,
		FormationID: assignment.FormationID,
		TenantID:    assignment.TenantID,
		Source:      assignment.Source,
		SourceType:  assignment.SourceType,
		Target:      assignment.Target,
		TargetType:  assignment.TargetType,
		State:       assignment.State,
		Value:       str.StringifyJSONRawMessage(assignment.Value),
		Error:       str.StringifyJSONRawMessage(assignment.Error),
	}
}

// SetReverseAssignment sets the reverseAssignment for the ApplicationTenantMappingInput to the provided one
func (rd *ApplicationTenantMappingInput) SetReverseAssignment(reverseAssignment *model.FormationAssignment) {
	rd.ReverseAssignment = &FormationAssignment{
		ID:          reverseAssignment.ID,
		FormationID: reverseAssignment.FormationID,
		TenantID:    reverseAssignment.TenantID,
		Source:      reverseAssignment.Source,
		SourceType:  reverseAssignment.SourceType,
		Target:      reverseAssignment.Target,
		TargetType:  reverseAssignment.TargetType,
		State:       reverseAssignment.State,
		Value:       str.StringifyJSONRawMessage(reverseAssignment.Value),
		Error:       str.StringifyJSONRawMessage(reverseAssignment.Error),
	}
}

// Clone returns a copy of the ApplicationTenantMappingInput
func (rd *ApplicationTenantMappingInput) Clone() FormationAssignmentTemplateInput {
	return &ApplicationTenantMappingInput{
		Operation:                 rd.Operation,
		FormationID:               rd.FormationID,
		Formation:                 rd.Formation,
		SourceApplicationTemplate: rd.SourceApplicationTemplate,
		SourceApplication:         rd.SourceApplication,
		TargetApplicationTemplate: rd.TargetApplicationTemplate,
		TargetApplication:         rd.TargetApplication,
		CustomerTenantContext:     rd.CustomerTenantContext,
		Assignment:                rd.Assignment,
		ReverseAssignment:         rd.ReverseAssignment,
	}
}
