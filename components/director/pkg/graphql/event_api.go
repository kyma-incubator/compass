package graphql

type EventAPISpec struct {
	Data         *CLOB            `json:"data"`
	Type         EventAPISpecType `json:"type"`
	Format       SpecFormat       `json:"format"`
	DefinitionID string           // Needed to resolve FetchRequest for given APISpec
}

// Extended types used by external API

type EventAPIDefinitionPageExt struct {
	EventAPIDefinitionPage
	Data []*EventAPIDefinitionExt `json:"data"`
}

type EventAPIDefinitionExt struct {
	EventAPIDefinition
	Spec *EventAPISpecExt `json:"spec"`
}

type EventAPISpecExt struct {
	EventAPISpec
	FetchRequest *FetchRequest `json:"fetchRequest"`
}
