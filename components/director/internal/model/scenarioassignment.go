package model

type AutomaticScenarioAssignment struct {
	ScenarioName string
	Tenant       string
	Selector     LabelSelector
}

type LabelSelector struct {
	Key   string
	Value string
}
