package model

type APIDefinitionInput struct {
	Name        string        `json:"name"`
	Description *string       `json:"description"`
	TargetURL   string        `json:"targetURL"`
	Group       *string       `json:"group"`
	// TODO: Replace with actual model
}
