package model

type APIDefinitionInput struct {

	Name        string
	Description *string
	TargetURL   string
	Group       *string
	// TODO: Replace with actual model
}

type APIDefinition struct {
	ID          string
	Name        string
	Description *string
	TargetURL   string
	// TODO: Replace with actual model
}
