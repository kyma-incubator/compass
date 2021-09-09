package model

import "github.com/kyma-incubator/compass/components/director/pkg/pagination"

// AutomaticScenarioAssignment missing godoc
type AutomaticScenarioAssignment struct {
	ScenarioName string
	Tenant       string
	Selector     LabelSelector
}

// LabelSelector missing godoc
type LabelSelector struct {
	Key   string
	Value string
}

// AutomaticScenarioAssignmentPage missing godoc
type AutomaticScenarioAssignmentPage struct {
	Data       []*AutomaticScenarioAssignment
	PageInfo   *pagination.Page
	TotalCount int
}
