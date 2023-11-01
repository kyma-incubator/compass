package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// TenantWithLabels represents a tenant with its corresponding labels
type TenantWithLabels struct {
	*model.BusinessTenantMapping
	Labels map[string]string
}

// ApplicationWithLabels represents an application with its corresponding labels
type ApplicationWithLabels struct {
	*model.Application
	Labels map[string]string
	Tenant *TenantWithLabels
}

// ApplicationTemplateWithLabels represents an application template with its corresponding labels
type ApplicationTemplateWithLabels struct {
	*model.ApplicationTemplate
	Labels       map[string]string
	Tenant       *TenantWithLabels
	TrustDetails *TrustDetails
}

// RuntimeWithLabels represents a runtime with its corresponding labels
type RuntimeWithLabels struct {
	*model.Runtime
	Labels       map[string]string
	Tenant       *TenantWithLabels
	TrustDetails *TrustDetails
}

// RuntimeContextWithLabels represents runtime context with its corresponding labels
type RuntimeContextWithLabels struct {
	*model.RuntimeContext
	Labels map[string]string
	Tenant *TenantWithLabels
}

// CustomerTenantContext represents the tenant hierarchy of the customer creating the formation. Both IDs are the external ones
type CustomerTenantContext struct {
	CustomerID string
	AccountID  *string
	Path       *string
}

// FormationAssignment represents the FormationAssignment model, but with the value stored as a string
// Because otherwise the template later renders it as a stringified []byte rather than a string
type FormationAssignment struct {
	ID          string                        `json:"id"`
	FormationID string                        `json:"formation_id"`
	TenantID    string                        `json:"tenant_id"`
	Source      string                        `json:"source"`
	SourceType  model.FormationAssignmentType `json:"source_type"`
	Target      string                        `json:"target"`
	TargetType  model.FormationAssignmentType `json:"target_type"`
	State       string                        `json:"state"`
	Value       *string                       `json:"value"`
	Error       *string                       `json:"error"`
}

// TrustDetails represents the certificate details
type TrustDetails struct {
	Subjects []string
}

// FormationConfigurationChangeInput struct contains the input for a formation notification
type FormationConfigurationChangeInput struct {
	Operation             model.FormationOperation
	FormationID           string
	Formation             *model.Formation
	ApplicationTemplate   *ApplicationTemplateWithLabels
	Application           *ApplicationWithLabels
	Runtime               *RuntimeWithLabels
	RuntimeContext        *RuntimeContextWithLabels
	CustomerTenantContext *CustomerTenantContext
	Assignment            *FormationAssignment
	ReverseAssignment     *FormationAssignment
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
		Value:       str.StringifyJSONRawMessage(assignment.Value),
		Error:       str.StringifyJSONRawMessage(assignment.Error),
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
		Value:       str.StringifyJSONRawMessage(reverseAssignment.Value),
		Error:       str.StringifyJSONRawMessage(reverseAssignment.Error),
	}
}

// Clone returns a copy of the FormationConfigurationChangeInput
func (rd *FormationConfigurationChangeInput) Clone() FormationAssignmentTemplateInput {
	return &FormationConfigurationChangeInput{
		Operation:             rd.Operation,
		FormationID:           rd.Formation.ID,
		Formation:             rd.Formation,
		ApplicationTemplate:   rd.ApplicationTemplate,
		Application:           rd.Application,
		Runtime:               rd.Runtime,
		RuntimeContext:        rd.RuntimeContext,
		CustomerTenantContext: rd.CustomerTenantContext,
		Assignment:            rd.Assignment,
		ReverseAssignment:     rd.ReverseAssignment,
	}
}
