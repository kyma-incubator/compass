package graphql

type Bundle struct {
	ID                             string      `json:"id"`
	Name                           string      `json:"name"`
	Description                    *string     `json:"description"`
	InstanceAuthRequestInputSchema *JSONSchema `json:"InstanceAuthRequestInputSchema"`
	// When defined, all Auth requests fallback to defaultAuth.
	DefaultInstanceAuth *Auth `json:"defaultInstanceAuth"`
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
