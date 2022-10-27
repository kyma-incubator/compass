package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// ApplicationWithLabels represents an application with its corresponding labels
type ApplicationWithLabels struct {
	*model.Application
	Labels map[string]interface{}
}

// ApplicationTemplateWithLabels represents an application template with its corresponding labels
type ApplicationTemplateWithLabels struct {
	*model.ApplicationTemplate
	Labels map[string]interface{}
}

// RuntimeWithLabels represents a runtime with its corresponding labels
type RuntimeWithLabels struct {
	*model.Runtime
	Labels map[string]interface{}
}

// RuntimeContextWithLabels represents runtime context with its corresponding labels
type RuntimeContextWithLabels struct {
	*model.RuntimeContext
	Labels map[string]interface{}
}

// FormationAssignment represents the FormationAssignment model, but with the value stored as a string
// Because otherwise the template later renders it as a stringified []byte rather than a string
type FormationAssignment struct {
	ID          string `json:"id"`
	FormationID string `json:"formation_id"`
	TenantID    string `json:"tenant_id"`
	Source      string `json:"source"`
	SourceType  string `json:"source_type"`
	Target      string `json:"target"`
	TargetType  string `json:"target_type"`
	State       string `json:"state"`
	Value       string `json:"value"`
}

// FormationConfigurationChangeInput struct contains the input for a formation notification
type FormationConfigurationChangeInput struct {
	Operation           model.FormationOperation
	FormationID         string
	ApplicationTemplate *ApplicationTemplateWithLabels
	Application         *ApplicationWithLabels
	Runtime             *RuntimeWithLabels
	RuntimeContext      *RuntimeContextWithLabels
	Assignment          *FormationAssignment
	ReverseAssignment   *FormationAssignment
}

// ParseURLTemplate missing godoc
func (rd *FormationConfigurationChangeInput) ParseURLTemplate(tmpl *string) (*URL, error) {
	var url URL
	return &url, parseTemplate(tmpl, *rd, &url)
}

// ParseInputTemplate missing godoc
func (rd *FormationConfigurationChangeInput) ParseInputTemplate(tmpl *string) ([]byte, error) {
	res := json.RawMessage{}
	if err := parseTemplate(tmpl, *rd, &res); err != nil {
		return nil, err
	}
	res = bytes.ReplaceAll(res, []byte("<nil>"), nil)
	return res, nil
}

// ParseHeadersTemplate missing godoc
func (rd *FormationConfigurationChangeInput) ParseHeadersTemplate(tmpl *string) (http.Header, error) {
	var headers http.Header
	return headers, parseTemplate(tmpl, *rd, &headers)
}

// GetParticipantsIDs returns the list of IDs part of the FormationConfigurationChangeInput
func (rd *FormationConfigurationChangeInput) GetParticipantsIDs() []string {
	var participants []string
	if rd.Application != nil {
		participants = append(participants, rd.Application.ID)
	}
	if rd.Runtime != nil {
		participants = append(participants, rd.Runtime.ID)
	}
	if rd.RuntimeContext != nil {
		participants = append(participants, rd.RuntimeContext.ID)
	}

	return participants
}

// SetAssignment sets the assignment for the FormationConfigurationChangeInput to the provided one
func (rd *FormationConfigurationChangeInput) SetAssignment(assignment *model.FormationAssignment) {
	rd.Assignment = &FormationAssignment{
		ID:          assignment.ID,
		FormationID: assignment.FormationID,
		TenantID:    assignment.TenantID,
		Source:      assignment.Source,
		SourceType:  assignment.SourceType,
		Target:      assignment.Target,
		TargetType:  assignment.TargetType,
		State:       assignment.State,
		Value:       string(assignment.Value),
	}
}

// SetReverseAssignment sets the reverse assignment for the FormationConfigurationChangeInput to the provided one
func (rd *FormationConfigurationChangeInput) SetReverseAssignment(reverseAssignment *model.FormationAssignment) {
	rd.ReverseAssignment = &FormationAssignment{
		ID:          reverseAssignment.ID,
		FormationID: reverseAssignment.FormationID,
		TenantID:    reverseAssignment.TenantID,
		Source:      reverseAssignment.Source,
		SourceType:  reverseAssignment.SourceType,
		Target:      reverseAssignment.Target,
		TargetType:  reverseAssignment.TargetType,
		State:       reverseAssignment.State,
		Value:       string(reverseAssignment.Value),
	}
}

// Clone returns a copy of the FormationConfigurationChangeInput
func (rd *FormationConfigurationChangeInput) Clone() FormationAssignmentTemplateInput {
	return &FormationConfigurationChangeInput{
		Operation:           rd.Operation,
		FormationID:         rd.FormationID,
		ApplicationTemplate: rd.ApplicationTemplate,
		Application:         rd.Application,
		Runtime:             rd.Runtime,
		RuntimeContext:      rd.RuntimeContext,
		Assignment:          rd.Assignment,
		ReverseAssignment:   rd.ReverseAssignment,
	}
}
