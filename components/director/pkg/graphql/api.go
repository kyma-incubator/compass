package graphql

import "github.com/kyma-incubator/compass/components/director/pkg/resource"

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

func (e *APIDefinition) GetType() resource.Type {
	return resource.API
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
