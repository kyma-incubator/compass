package graphql

type EventDefinition struct {
	ID               string  `json:"id"`
	OpenDiscoveryID  *string `json:"openDiscoveryID"`
	BundleID         string  `json:"bundleID"`
	Title            string  `json:"title"`
	ShortDescription string  `json:"shortDescription"`
	Description      *string `json:"description"`
	// group allows you to find the same API but in different version
	Group            *string      `json:"group"`
	Specs            []*EventSpec `json:"specs"`
	Version          *Version     `json:"version"`
	EventDefinitions JSON         `json:"eventDefinitions"`
	Tags             *JSON        `json:"tags"`
	Documentation    *string      `json:"documentation"`
	ChangelogEntries *JSON        `json:"changelogEntries"`
	Logo             *string      `json:"logo"`
	Image            *string      `json:"image"`
	URL              *string      `json:"url"`
	ReleaseStatus    string       `json:"releaseStatus"`
	LastUpdated      Timestamp    `json:"lastUpdated"`
	Extensions       *JSON        `json:"extensions"`
}

type EventSpec struct {
	ID     string        // Needed to resolve FetchRequest for given APISpec
	Data   *CLOB         `json:"data"`
	Type   EventSpecType `json:"type"`
	Format SpecFormat    `json:"format"`
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
