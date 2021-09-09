package graphql

import "github.com/kyma-incubator/compass/components/director/pkg/resource"

// EventDefinition missing godoc
type EventDefinition struct {
	BundleID    string  `json:"bundleID"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	// group allows you to find the same API but in different version
	Group   *string    `json:"group"`
	Spec    *EventSpec `json:"spec"`
	Version *Version   `json:"version"`
	*BaseEntity
}

// GetType missing godoc
func (e *EventDefinition) GetType() resource.Type {
	return resource.EventDefinition
}

// EventSpec missing godoc
type EventSpec struct {
	ID           string        `json:"id"`
	Data         *CLOB         `json:"data"`
	Type         EventSpecType `json:"type"`
	Format       SpecFormat    `json:"format"`
	DefinitionID string        // Needed to resolve FetchRequest for given APISpec
}

// EventAPIDefinitionPageExt is an extended types used by external API
type EventAPIDefinitionPageExt struct {
	EventDefinitionPage
	Data []*EventAPIDefinitionExt `json:"data"`
}

// EventAPIDefinitionExt missing godoc
type EventAPIDefinitionExt struct {
	EventDefinition
	Spec *EventAPISpecExt `json:"spec"`
}

// EventAPISpecExt missing godoc
type EventAPISpecExt struct {
	EventSpec
	FetchRequest *FetchRequest `json:"fetchRequest"`
}
