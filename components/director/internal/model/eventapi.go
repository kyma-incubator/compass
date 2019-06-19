package model

type EventAPIDefinition struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	// group allows you to find the same API but in different version
	Group   *string       `json:"group"`
	// TODO: Replace with actual model
}

type EventAPIDefinitionInput struct {
	Name        string             `json:"name"`
	Description *string            `json:"description"`
	// TODO: Replace with actual model
}

func (e *EventAPIDefinitionInput) ToEventAPIDefinition() *EventAPIDefinition {
	// TODO: Replace with actual model
	return &EventAPIDefinition{

	}
}