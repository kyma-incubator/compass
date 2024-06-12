package model

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/uid"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

// ApplicationTemplate missing godoc
type ApplicationTemplate struct {
	ID                   string
	Name                 string
	Description          *string
	ApplicationNamespace *string
	ApplicationInputJSON string
	Placeholders         []ApplicationTemplatePlaceholder
	AccessLevel          ApplicationTemplateAccessLevel
	Webhooks             []Webhook
	Labels               map[string]interface{}
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// SetCreatedAt is a setter for CreatedAt property
func (at *ApplicationTemplate) SetCreatedAt(t time.Time) {
	at.CreatedAt = t
}

// SetUpdatedAt is a setter for UpdatedAt property
func (at *ApplicationTemplate) SetUpdatedAt(t time.Time) {
	at.UpdatedAt = t
}

// ApplicationTemplatePage missing godoc
type ApplicationTemplatePage struct {
	Data       []*ApplicationTemplate
	PageInfo   *pagination.Page
	TotalCount int
}

// ApplicationTemplateInput missing godoc
type ApplicationTemplateInput struct {
	ID                   *string
	Name                 string
	Description          *string
	ApplicationNamespace *string
	ApplicationInputJSON string
	Placeholders         []ApplicationTemplatePlaceholder
	AccessLevel          ApplicationTemplateAccessLevel
	Labels               map[string]interface{}
	Webhooks             []*WebhookInput
}

// ApplicationTemplateAccessLevel missing godoc
type ApplicationTemplateAccessLevel string

// GlobalApplicationTemplateAccessLevel missing godoc
const (
	GlobalApplicationTemplateAccessLevel ApplicationTemplateAccessLevel = "GLOBAL"
)

// ApplicationFromTemplateInput missing godoc
type ApplicationFromTemplateInput struct {
	ID                  *string
	TemplateName        string
	Values              ApplicationFromTemplateInputValues
	PlaceholdersPayload *string
	Labels              map[string]interface{}
}

// ApplicationFromTemplateInputValues missing godoc
type ApplicationFromTemplateInputValues []*ApplicationTemplateValueInput

// FindPlaceholderValue missing godoc
func (in ApplicationFromTemplateInputValues) FindPlaceholderValue(name string) (string, error) {
	for _, value := range in {
		if value.Placeholder == name {
			return value.Value, nil
		}
	}
	return "", fmt.Errorf("value for placeholder name '%s' not found", name)
}

// ApplicationTemplatePlaceholder missing godoc
type ApplicationTemplatePlaceholder struct {
	Name        string
	Description *string
	JSONPath    *string
	Optional    *bool
}

// ApplicationTemplateValueInput missing godoc
type ApplicationTemplateValueInput struct {
	Placeholder string
	Value       string
}

// ToApplicationTemplate missing godoc
func (a *ApplicationTemplateInput) ToApplicationTemplate(id string) ApplicationTemplate {
	if a == nil {
		return ApplicationTemplate{}
	}

	uidService := uid.NewService()
	webhooks := make([]Webhook, 0)
	for _, webhookInput := range a.Webhooks {
		webhook := webhookInput.ToWebhook(uidService.Generate(), id, ApplicationTemplateWebhookReference)
		webhooks = append(webhooks, *webhook)
	}

	return ApplicationTemplate{
		ID:                   id,
		Name:                 a.Name,
		Description:          a.Description,
		ApplicationNamespace: a.ApplicationNamespace,
		ApplicationInputJSON: a.ApplicationInputJSON,
		Placeholders:         a.Placeholders,
		AccessLevel:          a.AccessLevel,
		Webhooks:             webhooks,
		Labels:               a.Labels,
	}
}

// ApplicationTemplateUpdateInput missing godoc
type ApplicationTemplateUpdateInput struct {
	Name                 string
	Description          *string
	ApplicationNamespace *string
	ApplicationInputJSON string
	Placeholders         []ApplicationTemplatePlaceholder
	AccessLevel          ApplicationTemplateAccessLevel
	Labels               map[string]interface{}
	Webhooks             []*WebhookInput
}

// ToApplicationTemplate missing godoc
func (a *ApplicationTemplateUpdateInput) ToApplicationTemplate(id string) ApplicationTemplate {
	if a == nil {
		return ApplicationTemplate{}
	}

	uidService := uid.NewService()
	webhooks := make([]Webhook, 0)
	for _, webhookInput := range a.Webhooks {
		webhook := webhookInput.ToWebhook(uidService.Generate(), id, ApplicationTemplateWebhookReference)
		webhooks = append(webhooks, *webhook)
	}

	return ApplicationTemplate{
		ID:                   id,
		Name:                 a.Name,
		Description:          a.Description,
		ApplicationNamespace: a.ApplicationNamespace,
		ApplicationInputJSON: a.ApplicationInputJSON,
		Placeholders:         a.Placeholders,
		AccessLevel:          a.AccessLevel,
		Labels:               a.Labels,
		Webhooks:             webhooks,
	}
}
