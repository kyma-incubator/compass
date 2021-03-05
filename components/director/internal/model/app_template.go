package model

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/uid"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type ApplicationTemplate struct {
	ID                   string
	Name                 string
	Description          *string
	ApplicationInputJSON string
	Placeholders         []ApplicationTemplatePlaceholder
	AccessLevel          ApplicationTemplateAccessLevel
	Webhooks             []Webhook
}

type ApplicationTemplatePage struct {
	Data       []*ApplicationTemplate
	PageInfo   *pagination.Page
	TotalCount int
}

type ApplicationTemplateInput struct {
	Name                 string
	Description          *string
	ApplicationInputJSON string
	Placeholders         []ApplicationTemplatePlaceholder
	AccessLevel          ApplicationTemplateAccessLevel
	Webhooks             []*WebhookInput
}

type ApplicationTemplateAccessLevel string

const (
	GlobalApplicationTemplateAccessLevel ApplicationTemplateAccessLevel = "GLOBAL"
)

type ApplicationFromTemplateInput struct {
	TemplateName string
	Values       ApplicationFromTemplateInputValues
}

type ApplicationFromTemplateInputValues []*ApplicationTemplateValueInput

func (in ApplicationFromTemplateInputValues) FindPlaceholderValue(name string) (string, error) {
	for _, value := range in {
		if value.Placeholder == name {
			return value.Value, nil
		}
	}
	return "", fmt.Errorf("value for placeholder name '%s' not found", name)
}

type ApplicationTemplatePlaceholder struct {
	Name        string
	Description *string
}

type ApplicationTemplateValueInput struct {
	Placeholder string
	Value       string
}

func (a *ApplicationTemplateInput) ToApplicationTemplate(appTemplateID string) ApplicationTemplate {
	if a == nil {
		return ApplicationTemplate{}
	}

	uidService := uid.NewService()
	webhooks := make([]Webhook, 0, 0)
	for _, webhookInput := range a.Webhooks {
		webhook := webhookInput.ToApplicationTemplateWebhook(uidService.Generate(), str.Ptr(""), appTemplateID)
		webhooks = append(webhooks, *webhook)
	}

	return ApplicationTemplate{
		ID:                   appTemplateID,
		Name:                 a.Name,
		Description:          a.Description,
		ApplicationInputJSON: a.ApplicationInputJSON,
		Placeholders:         a.Placeholders,
		AccessLevel:          a.AccessLevel,
		Webhooks:             webhooks,
	}
}
