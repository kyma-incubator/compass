package graphql

type Bundle struct {
	ID                             string      `json:"id"`
	Title                          string      `json:"title"`
	ShortDescription               string      `json:"shortDescription"`
	Description                    *string     `json:"description"`
	Version                        *string     `json:"version"`
	InstanceAuthRequestInputSchema *JSONSchema `json:"InstanceAuthRequestInputSchema"`
	// When defined, all Auth requests fallback to defaultAuth.
	DefaultInstanceAuth *Auth     `json:"defaultInstanceAuth"`
	Tags                *JSON     `json:"tags"`
	LastUpdated         Timestamp `json:"lastUpdated"`
	Extensions          *JSON     `json:"extensions"`
}

type BundleExt struct {
	Bundle
	APIDefinitions   APIDefinitionPageExt      `json:"apiDefinitions"`
	EventDefinitions EventAPIDefinitionPageExt `json:"eventDefinitions"`
	Documents        DocumentPageExt           `json:"documents"`
	APIDefinition    APIDefinitionExt          `json:"apiDefinition"`
	EventDefinition  EventDefinition           `json:"eventDefinition"`
	Document         Document                  `json:"document"`
	InstanceAuth     *BundleInstanceAuth       `json:"instanceAuth"`
	InstanceAuths    []*BundleInstanceAuth     `json:"instanceAuths"`
}

type BundlePageExt struct {
	BundlePage
	Data []*BundleExt `json:"data"`
}
