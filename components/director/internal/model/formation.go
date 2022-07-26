package model

import "github.com/kyma-incubator/compass/components/director/pkg/pagination"

// DefaultTemplateName will be used as default formation template name if no other options are provided
const DefaultTemplateName = "Side-by-side extensibility with Kyma"

// Formation missing godoc
type Formation struct {
	ID                  string
	TenantID            string
	FormationTemplateID string
	Name                string
}

// FormationPage contains Formation data with page info
type FormationPage struct {
	Data       []*Formation
	PageInfo   *pagination.Page
	TotalCount int
}
