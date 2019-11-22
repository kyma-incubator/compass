package model

import "github.com/kyma-incubator/compass/components/director/pkg/pagination"

type ApplicationTemplate struct {
	ID               string
	Name             string
	Description      *string
	ApplicationInput *ApplicationCreateInput
	Placeholders     []ApplicationTemplatePlaceholder
	AccessLevel      ApplicationTemplateAccessLevel
}

type ApplicationTemplatePage struct {
	Data       []*ApplicationTemplate
	PageInfo   *pagination.Page
	TotalCount int
}

type ApplicationTemplateInput struct {
	Name             string
	Description      *string
	ApplicationInput *ApplicationCreateInput
	Placeholders     []ApplicationTemplatePlaceholder
	AccessLevel      ApplicationTemplateAccessLevel
}

type ApplicationTemplateAccessLevel string

const (
	GlobalApplicationTemplateAccessLevel ApplicationTemplateAccessLevel = "GLOBAL"
)

type ApplicationTemplatePlaceholder struct {
	Name        string
	Description *string
}

type ApplicationTemplateValueInput struct {
	Placeholder string
	Value       string
}
