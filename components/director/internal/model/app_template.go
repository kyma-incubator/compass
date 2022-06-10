package model

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/uid"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

// ApplicationTemplate missing godoc
type ApplicationTemplate struct {
	ID                   string
	Name                 string
	Description          *string
	ApplicationInputJSON string
	Placeholders         []ApplicationTemplatePlaceholder
	AccessLevel          ApplicationTemplateAccessLevel
	Webhooks             []Webhook
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
	TemplateName string
	Values       ApplicationFromTemplateInputValues
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
		ApplicationInputJSON: a.ApplicationInputJSON,
		Placeholders:         a.Placeholders,
		AccessLevel:          a.AccessLevel,
		Webhooks:             webhooks,
	}
}

// ApplicationTemplateUpdateInput missing godoc
type ApplicationTemplateUpdateInput struct {
	Name                 string
	Description          *string
	ApplicationInputJSON string
	Placeholders         []ApplicationTemplatePlaceholder
	AccessLevel          ApplicationTemplateAccessLevel
}

// ToApplicationTemplate missing godoc
func (a *ApplicationTemplateUpdateInput) ToApplicationTemplate(id string) ApplicationTemplate {
	if a == nil {
		return ApplicationTemplate{}
	}

	return ApplicationTemplate{
		ID:                   id,
		Name:                 a.Name,
		Description:          a.Description,
		ApplicationInputJSON: a.ApplicationInputJSON,
		Placeholders:         a.Placeholders,
		AccessLevel:          a.AccessLevel,
	}
}
