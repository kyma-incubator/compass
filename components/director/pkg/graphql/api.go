package graphql

import "github.com/kyma-incubator/compass/components/director/pkg/resource"

// APIDefinition missing godoc
type APIDefinition struct {
	BundleID    string   `json:"bundleID"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	Spec        *APISpec `json:"spec"`
	TargetURL   string   `json:"targetURL"`
	//  group allows you to find the same API but in different version
	Group   *string  `json:"group"`
	Version *Version `json:"version"`
	*BaseEntity
}

// GetType missing godoc
func (e *APIDefinition) GetType() resource.Type {
	return resource.API
}

// APISpec missing godoc
type APISpec struct {
	// when fetch request specified, data will be automatically populated
	ID           string      `json:"id"`
	Data         *CLOB       `json:"data"`
	Format       SpecFormat  `json:"format"`
	Type         APISpecType `json:"type"`
	DefinitionID string      // Needed to resolve FetchRequest for given APISpec
}

// APIDefinitionPageExt is an extended type used by external API
type APIDefinitionPageExt struct {
	APIDefinitionPage
	Data []*APIDefinitionExt `json:"data"`
}

// APIDefinitionExt missing godoc
type APIDefinitionExt struct {
	APIDefinition
	Spec *APISpecExt `json:"spec"`
}

// APISpecExt missing godoc
type APISpecExt struct {
	APISpec
	FetchRequest *FetchRequest `json:"fetchRequest"`
}
