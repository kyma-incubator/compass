package model

import "github.com/kyma-incubator/compass/components/director/pkg/pagination"

type AutomaticScenarioAssignment struct {
	ScenarioName string
	Tenant       string
	Selector     LabelSelector
}

type LabelSelector struct {
	Key   string
	Value string
}

type AutomaticScenarioAssignmentPage struct {
	Data       []*AutomaticScenarioAssignment
	PageInfo   *pagination.Page
	TotalCount int
}
