package graphql

type APIDefinition struct {
	ID          string   `json:"id"`
	BundleID    string   `json:"bundleID"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	Spec        *APISpec `json:"spec"`
	TargetURL   string   `json:"targetURL"`
	//  group allows you to find the same API but in different version
	Group   *string  `json:"group"`
	Version *Version `json:"version"`
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
