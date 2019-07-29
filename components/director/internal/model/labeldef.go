package model

type LabelDefinition struct {
	ID     string
	Tenant string
	Key    string
	Schema *interface{}
}

func (def *LabelDefinition) Validate() error {
	// TODO later
	return nil
}
