package graphql

import "github.com/kyma-incubator/compass/components/director/pkg/resource"

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

func (e *EventDefinition) GetType() resource.Type {
	return resource.EventDefinition
}

type EventSpec struct {
	Data         *CLOB         `json:"data"`
	Type         EventSpecType `json:"type"`
	Format       SpecFormat    `json:"format"`
	DefinitionID string        // Needed to resolve FetchRequest for given APISpec
}

// Extended types used by external API

type EventAPIDefinitionPageExt struct {
	EventDefinitionPage
	Data []*EventAPIDefinitionExt `json:"data"`
}

type EventAPIDefinitionExt struct {
	EventDefinition
	Spec *EventAPISpecExt `json:"spec"`
}

type EventAPISpecExt struct {
	EventSpec
	FetchRequest *FetchRequest `json:"fetchRequest"`
}
