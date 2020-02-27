package graphql

type Package struct {
	ID                             string      `json:"id"`
	Name                           string      `json:"name"`
	Description                    *string     `json:"description"`
	InstanceAuthRequestInputSchema *JSONSchema `json:"InstanceAuthRequestInputSchema"`
	// When defined, all Auth requests fallback to defaultAuth.
	DefaultInstanceAuth *Auth `json:"defaultInstanceAuth"`
}

type PackageExt struct {
	Package
	APIDefinitions   APIDefinitionPageExt      `json:"apiDefinitions"`
	EventDefinitions EventAPIDefinitionPageExt `json:"eventDefinitions"`
	Documents        DocumentPageExt           `json:"documents"`
	APIDefinition    APIDefinitionExt          `json:"apiDefinition"`
	EventDefinition  EventDefinition           `json:"eventDefinition"`
	Document         Document                  `json:"document"`
}

type PackagePageExt struct {
	Data       []*PackageExt `json:"data"`
	PageInfo   *PageInfo     `json:"pageInfo"`
	TotalCount int           `json:"totalCount"`
}
