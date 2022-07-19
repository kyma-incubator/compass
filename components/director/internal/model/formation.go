package model

import "github.com/kyma-incubator/compass/components/director/pkg/pagination"

// DefaultTemplateName will be used as default formation templane name if no other options are provided
const DefaultTemplateName = "Side-by-side extensibility with Kyma"

// FormationOperation defines the kind of operation done on a given formation
type FormationOperation string

const (
	// AssignFormation represents the assign operation done on a given formation
	AssignFormation FormationOperation = "assign"
	// UnassignFormation represents the unassign operation done on a given formation
	UnassignFormation FormationOperation = "unassign"
)

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
