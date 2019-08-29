package graphql

type EventAPISpec struct {
	Data   *CLOB            `json:"data"`
	Type   EventAPISpecType `json:"type"`
	Format SpecFormat       `json:"format"`
	DefinitionID string // Needed to resolve FetchRequest for given APISpec
}

// Extended types used by external API

type EventAPIDefinitionPageExt struct {
	Data       []*EventAPIDefinitionExt `json:"data"`
	PageInfo   *PageInfo                `json:"pageInfo"`
	TotalCount int                      `json:"totalCount"`
}

type EventAPIDefinitionExt struct {
	ID            string  `json:"id"`
	ApplicationID string  `json:"applicationID"`
	Name          string  `json:"name"`
	Description   *string `json:"description"`
	// group allows you to find the same API but in different version
	Group   *string       `json:"group"`
	Spec    *EventAPISpec `json:"spec"`
	Version *Version      `json:"version"`
}

type EventAPISpecExt struct {
	EventAPISpec
	FetchRequest *FetchRequest `json:"fetchRequest"`
}
