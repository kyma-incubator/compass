package graphql

type APIDefinition struct {
	ID               string     `json:"id"`
	OpenDiscoveryID  *string    `json:"openDiscoveryID"`
	BundleID         string     `json:"bundleID"`
	Title            string     `json:"title"`
	ShortDescription string     `json:"shortDescription"`
	Description      *string    `json:"description"`
	Specs            []*APISpec `json:"specs"`
	EntryPoint       string     `json:"entryPoint"`
	//  group allows you to find the same API but in different version
	Group            *string   `json:"group"`
	Version          *Version  `json:"version"`
	APIDefinitions   JSON      `json:"apiDefinitions"`
	Tags             *JSON     `json:"tags"`
	Documentation    *string   `json:"documentation"`
	ChangelogEntries *JSON     `json:"changelogEntries"`
	Logo             *string   `json:"logo"`
	Image            *string   `json:"image"`
	URL              *string   `json:"url"`
	ReleaseStatus    string    `json:"releaseStatus"`
	APIProtocol      string    `json:"apiProtocol"`
	Actions          JSON      `json:"actions"`
	LastUpdated      Timestamp `json:"lastUpdated"`
	Extensions       *JSON     `json:"extensions"`
}

type APISpec struct {
	// when fetch request specified, data will be automatically populated
	Data         *CLOB       `json:"data"`
	Format       SpecFormat  `json:"format"`
	Type         APISpecType `json:"type"`
	DefinitionID string      // Needed to resolve FetchRequest for given APISpec
}

// Extended types used by external API

type APIDefinitionPageExt struct {
	APIDefinitionPage
	Data []*APIDefinitionExt `json:"data"`
}

type APIDefinitionExt struct {
	APIDefinition
	Spec *APISpecExt `json:"spec"`
}

type APISpecExt struct {
	APISpec
	FetchRequest *FetchRequest `json:"fetchRequest"`
}
