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
	Assignment        *model.FormationAssignment
	ReverseAssignment *model.FormationAssignment
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

// GetAssignments returns the assignments associated with the ApplicationTenantMappingInput
func (rd *ApplicationTenantMappingInput) GetAssignments() (*model.FormationAssignment, *model.FormationAssignment) {
	return rd.Assignment, rd.ReverseAssignment
}

// SetAssignments sets the assignments for the ApplicationTenantMappingInput to the provided ones
func (rd *ApplicationTenantMappingInput) SetAssignments(assignment, reverseAssignment *model.FormationAssignment) {
	rd.Assignment = assignment
	rd.ReverseAssignment = reverseAssignment
}

// GetAssignment returns the assignment associated with the ApplicationTenantMappingInput
func (rd *ApplicationTenantMappingInput) GetAssignment() *model.FormationAssignment {
	return rd.Assignment
}

// GetReverseAssignment returns the reverse assignment associated with the ApplicationTenantMappingInput
func (rd *ApplicationTenantMappingInput) GetReverseAssignment() *model.FormationAssignment {
	return rd.ReverseAssignment
}

// SetAssignment sets the assignment for the ApplicationTenantMappingInput to the provided one
func (rd *ApplicationTenantMappingInput) SetAssignment(assignment *model.FormationAssignment) {
	rd.Assignment = assignment
}

// SetReverseAssignment sets the reverseAssignment for the ApplicationTenantMappingInput to the provided one
func (rd *ApplicationTenantMappingInput) SetReverseAssignment(reverseAssignment *model.FormationAssignment) {
	rd.ReverseAssignment = reverseAssignment
}

// Clone return a copy of the ApplicationTenantMappingInput
func (rd *ApplicationTenantMappingInput) Clone() FormationAssignmentTemplateInput {
	return &ApplicationTenantMappingInput{
		Operation:                 rd.Operation,
		FormationID:               rd.FormationID,
		SourceApplicationTemplate: rd.SourceApplicationTemplate,
		SourceApplication:         rd.SourceApplication,
		TargetApplicationTemplate: rd.TargetApplicationTemplate,
		TargetApplication:         rd.TargetApplication,
		Assignment:                rd.Assignment,
		ReverseAssignment:         rd.ReverseAssignment,
	}
}
