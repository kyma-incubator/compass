package graphql

type EventDefinition struct {
	ID          string  `json:"id"`
	BundleID    string  `json:"bundleID"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	// group allows you to find the same API but in different version
	Group     *string    `json:"group"`
	Spec      *EventSpec `json:"spec"`
	Version   *Version   `json:"version"`
	Ready     bool       `json:"ready"`
	CreatedAt Timestamp  `json:"createdAt"`
	UpdatedAt Timestamp  `json:"updatedAt"`
	DeletedAt Timestamp  `json:"deletedAt"`
	Error     *string    `json:"error"`
}

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
