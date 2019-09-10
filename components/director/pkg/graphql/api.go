package graphql

type APIDefinition struct {
	ID            string   `json:"id"`
	ApplicationID string   `json:"applicationID"`
	Name          string   `json:"name"`
	Description   *string  `json:"description"`
	Spec          *APISpec `json:"spec"`
	TargetURL     string   `json:"targetURL"`
	//  group allows you to find the same API but in different version
	Group *string `json:"group"`
	// If defaultAuth is specified, it will be used for all Runtimes that does not specify Auth explicitly.
	DefaultAuth *Auth    `json:"defaultAuth"`
	Version     *Version `json:"version"`
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
	// "If runtime does not exist, an error is returned. If runtime exists but Auth for it is not set, defaultAuth is returned if specified.
	Auth *RuntimeAuth `json:"auth"`
	// Returns authentication details for all runtimes, even for a runtime, where Auth is not yet specified.
	Auths []*RuntimeAuth `json:"auths"`
}

type APISpecExt struct {
	APISpec
	FetchRequest *FetchRequest `json:"fetchRequest"`
}
