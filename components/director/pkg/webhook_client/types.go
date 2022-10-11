package webhookclient

import (
	"fmt"

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

// Request represents a webhook request to be executed
type Request struct {
	Webhook       graphql.Webhook
	Object        webhook.TemplateInput
	CorrelationID string
}

// NotificationRequest represents a webhook request to be executed
type NotificationRequest struct {
	Webhook       graphql.Webhook
	Object        webhook.FormationAssignmentTemplateInput
	CorrelationID string
}

type WebhookRequest interface {
	GetWebhook() graphql.Webhook
	GetObject() webhook.TemplateInput
	GetCorrelationID() string
}

func (r *Request) GetWebhook() graphql.Webhook {
	return r.Webhook
}

func (r *Request) GetObject() webhook.TemplateInput {
	return r.Object
}

func (r *Request) GetCorrelationID() string {
	return r.CorrelationID
}

func (nr *NotificationRequest) GetWebhook() graphql.Webhook {
	return nr.Webhook
}

func (nr *NotificationRequest) GetObject() webhook.TemplateInput {
	return nr.Object
}

func (nr *NotificationRequest) GetCorrelationID() string {
	return nr.CorrelationID
}

func (nr *NotificationRequest) Clone() *NotificationRequest {
	return &NotificationRequest{
		Webhook:       nr.Webhook,
		Object:        nr.Object.Clone(),
		CorrelationID: nr.CorrelationID,
	}
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
