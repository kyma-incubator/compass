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

// FormationNotificationRequest represents a formation webhook request to be executed
type FormationNotificationRequest struct {
	*Request
}

type FormationNotificationRequestExt struct {
	*Request
	Operation     model.FormationOperation
	Formation     *model.Formation
	FormationType string
}

func (fnr *FormationNotificationRequestExt) GetObjectType() model.ResourceType {
	return model.FormationResourceType
}

func (fnr *FormationNotificationRequestExt) GetObjectSubtype() string {
	return fnr.FormationType
}

func (fnr *FormationNotificationRequestExt) GetOperation() model.FormationOperation {
	return fnr.Operation
}

func (fnr *FormationNotificationRequestExt) GetFormationAssignment() *model.FormationAssignment {
	return nil
}

func (fnr *FormationNotificationRequestExt) GetReverseFormationAssignment() *model.FormationAssignment {
	return nil
}

func (fnr *FormationNotificationRequestExt) GetFormation() *model.Formation {
	return fnr.Formation
}

// FormationAssignmentNotificationRequest represents a formation assignment webhook request to be executed
type FormationAssignmentNotificationRequest struct {
	Webhook       graphql.Webhook
	Object        webhook.FormationAssignmentTemplateInput
	CorrelationID string
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
	GetFormationAssignment() *model.FormationAssignment
	GetReverseFormationAssignment() *model.FormationAssignment
	GetFormation() *model.Formation
}
