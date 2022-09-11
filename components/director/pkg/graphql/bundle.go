package graphql

import "github.com/kyma-incubator/compass/components/director/pkg/resource"

// Bundle missing godoc
type Bundle struct {
	Name                           string      `json:"name"`
	Description                    *string     `json:"description"`
	InstanceAuthRequestInputSchema *JSONSchema `json:"InstanceAuthRequestInputSchema"`
	CorrelationIDs                 []string    `json:"correlationIDs"`
	// When defined, all Auth requests fallback to defaultAuth.
	DefaultInstanceAuth *Auth `json:"defaultInstanceAuth"`
	*BaseEntity
}

// GetType missing godoc
func (e *Bundle) GetType() resource.Type {
	return resource.Bundle
}

// BundleExt missing godoc
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

// BundlePageExt missing godoc
type BundlePageExt struct {
	BundlePage
	Data []*BundleExt `json:"data"`
}
