package graphql

type EventSpec struct {
	Data         *CLOB         `json:"data"`
	Type         EventSpecType `json:"type"`
	Format       SpecFormat    `json:"format"`
	DefinitionID string        // Needed to resolve FetchRequest for given APISpec
}

// Extended types used by external API

type EventAPIDefinitionPageExt struct {
	EventDefinitionPage
	Data []*EventAPIDefinitionExt `json:"data"`
}

type EventAPIDefinitionExt struct {
	EventDefinition
	Spec *EventAPISpecExt `json:"spec"`
}

type EventAPISpecExt struct {
	EventSpec
	FetchRequest *FetchRequest `json:"fetchRequest"`
}
