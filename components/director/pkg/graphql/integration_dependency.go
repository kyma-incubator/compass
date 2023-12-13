package graphql

// IntegrationDependency represents the Integration Dependency object
type IntegrationDependency struct {
	Name          string    `json:"name"`
	Description   *string   `json:"description"`
	OrdID         *string   `json:"ordID"`
	PartOfPackage *string   `json:"partOfPackage"`
	Visibility    *string   `json:"visibility"`
	ReleaseStatus *string   `json:"releaseStatus"`
	Mandatory     *bool     `json:"mandatory"`
	Aspects       []*Aspect `json:"aspects"`
	Version       *Version  `json:"version"`
	*BaseEntity
}

// Aspect represents the Aspect object
type Aspect struct {
	Name           string                   `json:"name"`
	Description    *string                  `json:"description"`
	Mandatory      *bool                    `json:"mandatory"`
	APIResources   []*AspectAPIDefinition   `json:"apiResources"`
	EventResources []*AspectEventDefinition `json:"eventResources"`
	*BaseEntity
}

// AspectAPIDefinition represents the Aspect API Definition object
type AspectAPIDefinition struct {
	OrdID string `json:"ordID"`
}

// AspectEventDefinition represents the Aspect Event Definition object
type AspectEventDefinition struct {
	OrdID  string                         `json:"ordID"`
	Subset []*AspectEventDefinitionSubset `json:"subset"`
	*BaseEntity
}

// AspectEventDefinitionSubset represents the Aspect Event Definition Subset object
type AspectEventDefinitionSubset struct {
	EventType *string `json:"eventType"`
}
