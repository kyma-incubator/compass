package model

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type ApplicationTemplate struct {
	ID                   string
	Name                 string
	Description          *string
	ApplicationInputJSON string
	Placeholders         []ApplicationTemplatePlaceholder
	AccessLevel          ApplicationTemplateAccessLevel
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

func (a *ApplicationTemplateInput) ToApplicationTemplate(id string) ApplicationTemplate {
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
