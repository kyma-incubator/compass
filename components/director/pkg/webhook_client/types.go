package webhookclient

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
)

// WebhookStatusGoneErr represents an error type which represents a gone status code
// returned in response to calling delete webhook.
type WebhookStatusGoneErr struct {
	error
}

// NewWebhookStatusGoneErr constructs a new WebhookStatusGoneErr with the given error message
func NewWebhookStatusGoneErr(goneStatusCode int) WebhookStatusGoneErr {
	return WebhookStatusGoneErr{error: fmt.Errorf("gone response status %d was met while calling webhook", goneStatusCode)}
}

// FormationNotificationRequest represents a formation webhook request to be executed with added Operation, Formation and FormationType
type FormationNotificationRequest struct {
	*Request
	Operation     model.FormationOperation
	Formation     *model.Formation
	FormationType string
}

// GetObjectType returns FormationNotificationRequest object type
func (fnr *FormationNotificationRequest) GetObjectType() model.ResourceType {
	return model.FormationResourceType
}

// GetObjectSubtype returns FormationNotificationRequest object subtype
func (fnr *FormationNotificationRequest) GetObjectSubtype() string {
	return fnr.FormationType
}

// GetOperation returns FormationNotificationRequest operation
func (fnr *FormationNotificationRequest) GetOperation() model.FormationOperation {
	return fnr.Operation
}

// GetFormationAssignment returns FormationNotificationRequest formation assignment
func (fnr *FormationNotificationRequest) GetFormationAssignment() *webhook.FormationAssignment {
	return nil
}

// GetReverseFormationAssignment returns FormationNotificationRequest reverse formation assignment
func (fnr *FormationNotificationRequest) GetReverseFormationAssignment() *webhook.FormationAssignment {
	return nil
}

// GetFormation returns FormationNotificationRequest formation
func (fnr *FormationNotificationRequest) GetFormation() *model.Formation {
	return fnr.Formation
}

// FormationAssignmentNotificationRequest represents a formation assignment webhook request to be executed
type FormationAssignmentNotificationRequest struct {
	Webhook       graphql.Webhook
	Object        webhook.FormationAssignmentTemplateInput
	CorrelationID string
}

// FormationAssignmentNotificationRequestExt is extended FormationAssignmentRequest with Operation, FA, ReverseFA, Formation and Target subtype.
type FormationAssignmentNotificationRequestExt struct {
	*FormationAssignmentNotificationRequest
	Operation                  model.FormationOperation
	FormationAssignment        *webhook.FormationAssignment
	ReverseFormationAssignment *webhook.FormationAssignment
	Formation                  *model.Formation
	TargetSubtype              string
}

// GetObjectType returns FormationAssignmentNotificationRequestExt object type
func (f *FormationAssignmentNotificationRequestExt) GetObjectType() model.ResourceType {
	switch f.FormationAssignment.TargetType {
	case model.FormationAssignmentTypeApplication:
		return model.ApplicationResourceType

	case model.FormationAssignmentTypeRuntime:
		return model.RuntimeResourceType

	case model.FormationAssignmentTypeRuntimeContext:
		return model.RuntimeContextResourceType
	}
	return ""
}

// GetObjectSubtype returns FormationAssignmentNotificationRequestExt object subtype
func (f *FormationAssignmentNotificationRequestExt) GetObjectSubtype() string {
	return f.TargetSubtype
}

// GetOperation returns FormationAssignmentNotificationRequestExt operation
func (f *FormationAssignmentNotificationRequestExt) GetOperation() model.FormationOperation {
	return f.Operation
}

// GetFormationAssignment returns FormationAssignmentNotificationRequestExt formation assignment
func (f *FormationAssignmentNotificationRequestExt) GetFormationAssignment() *webhook.FormationAssignment {
	return f.FormationAssignment
}

// GetReverseFormationAssignment returns FormationAssignmentNotificationRequestExt reverse formation assignment
func (f *FormationAssignmentNotificationRequestExt) GetReverseFormationAssignment() *webhook.FormationAssignment {
	return f.ReverseFormationAssignment
}

// GetFormation returns FormationAssignmentNotificationRequestExt formation
func (f *FormationAssignmentNotificationRequestExt) GetFormation() *model.Formation {
	return f.Formation
}

// GetWebhook returns the Webhook associated with the FormationAssignmentNotificationRequest
func (nr *FormationAssignmentNotificationRequest) GetWebhook() graphql.Webhook {
	return nr.Webhook
}

// GetObject returns the Object associated with the FormationAssignmentNotificationRequest
func (nr *FormationAssignmentNotificationRequest) GetObject() webhook.TemplateInput {
	return nr.Object
}

// GetCorrelationID returns the CorrelationID assigned to the FormationAssignmentNotificationRequest
func (nr *FormationAssignmentNotificationRequest) GetCorrelationID() string {
	return nr.CorrelationID
}

// Clone returns a copy of the FormationAssignmentNotificationRequest
func (nr *FormationAssignmentNotificationRequest) Clone() *FormationAssignmentNotificationRequest {
	return &FormationAssignmentNotificationRequest{
		Webhook:       nr.Webhook,
		Object:        nr.Object.Clone(),
		CorrelationID: nr.CorrelationID,
	}
}

// WebhookRequest represent a request associated with registered webhook
type WebhookRequest interface {
	GetWebhook() graphql.Webhook
	GetObject() webhook.TemplateInput
	GetCorrelationID() string
}

// Request represents a webhook request to be executed
type Request struct {
	Webhook       graphql.Webhook
	Object        webhook.TemplateInput
	CorrelationID string
}

// GetWebhook return the Webhook associated with the Request
func (r *Request) GetWebhook() graphql.Webhook {
	return r.Webhook
}

// GetObject returns the Object associated with the Request
func (r *Request) GetObject() webhook.TemplateInput {
	return r.Object
}

// GetCorrelationID returns the CorrelationID assigned to the Request
func (r *Request) GetCorrelationID() string {
	return r.CorrelationID
}

// PollRequest represents a webhook poll request to be executed
type PollRequest struct {
	*Request
	PollURL string
}

// NewRequest constructs a webhook Request
func NewRequest(webhook graphql.Webhook, requestObject webhook.TemplateInput, correlationID string) *Request {
	return &Request{
		Webhook:       webhook,
		Object:        requestObject,
		CorrelationID: correlationID,
	}
}

// NewPollRequest constructs a webhook Request
func NewPollRequest(webhook graphql.Webhook, requestObject webhook.TemplateInput, correlationID string, pollURL string) *PollRequest {
	return &PollRequest{
		Request: NewRequest(webhook, requestObject, correlationID),
		PollURL: pollURL,
	}
}

// WebhookExtRequest represent an extended request associated with registered webhook
type WebhookExtRequest interface {
	WebhookRequest
	GetObjectType() model.ResourceType
	GetObjectSubtype() string
	GetOperation() model.FormationOperation
	GetFormationAssignment() *webhook.FormationAssignment
	GetReverseFormationAssignment() *webhook.FormationAssignment
	GetFormation() *model.Formation
}
