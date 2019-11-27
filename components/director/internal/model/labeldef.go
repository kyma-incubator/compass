package model

type LabelDefinition struct {
	ID     string
	Tenant string
	Key    string
	Schema *interface{}
}
