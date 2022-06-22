package model

import "github.com/kyma-incubator/compass/components/director/pkg/pagination"

// DefaultTemplateName will be used as default formation templane name if no other options are provided
var DefaultTemplateName = "Side-by-side extensibility with Kyma"

// Formation missing godoc
type Formation struct {
	ID                  string
	TenantID            string
	FormationTemplateID string
	Name                string
}

type FormationPage struct {
	Data       []*Formation
	PageInfo   *pagination.Page
	TotalCount int
}
