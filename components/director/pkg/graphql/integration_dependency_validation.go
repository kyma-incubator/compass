package graphql

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

const (
	// MinTitleLength represents the minimal accepted length of the Title field
	MinTitleLength = 1
	// MaxTitleLength represents the maximal accepted length of the Title field
	MaxTitleLength = 255
	// MinDescriptionLength represents the minimal accepted length of the Description field
	MinDescriptionLength = 1
	// MinOrdIDLength represents the minimal accepted length of the OrdID field
	MinOrdIDLength = 1
	// MaxOrdIDLength represents the maximal accepted length of the OrdID field
	MaxOrdIDLength = 255

	// AspectAPIResourceRegex represents the valid structure of the apiResource items in Integration Dependency Aspect
	AspectAPIResourceRegex = "^([a-z0-9]+(?:[.][a-z0-9]+)*):(apiResource):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"
	// AspectEventResourceRegex represents the valid structure of the eventResource items in Integration Dependency Aspect
	AspectEventResourceRegex = "^([a-z0-9]+(?:[.][a-z0-9]+)*):(eventResource):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"
	// EventResourceEventTypeRegex represents the valid structure of the event type items in event resource subset
	EventResourceEventTypeRegex = "^([a-z0-9A-Z]+(?:[.][a-z0-9A-Z]+)(?:[.][a-z0-9A-Z]+)+)\\.(v0|v[1-9][0-9]*)$"
	// IntegrationDependencyOrdIDRegexGQL represents the valid structure of the ordID of the Integration Dependency for the GraphQL scenario - app namespace can be empty
	IntegrationDependencyOrdIDRegexGQL = "^(([a-z0-9-]+(?:[.][a-z0-9-]+)*))*:(integrationDependency):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"
)

const (
	// ReleaseStatusBeta is one of the available release status options
	ReleaseStatusBeta string = "beta"
	// ReleaseStatusActive is one of the available release status options
	ReleaseStatusActive string = "active"
	// ReleaseStatusDeprecated is one of the available release status options
	ReleaseStatusDeprecated string = "deprecated"
)

// Validate validates IntegrationDependencyInput object
func (i IntegrationDependencyInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, is.PrintableASCII, validation.Length(MinTitleLength, MaxTitleLength)),
		validation.Field(&i.Description, validation.NilOrNotEmpty, validation.Length(MinDescriptionLength, descriptionStringLengthLimit)),
		validation.Field(&i.OrdID, validation.Required, validation.Length(MinOrdIDLength, MaxOrdIDLength), validation.Match(regexp.MustCompile(IntegrationDependencyOrdIDRegexGQL))),
		validation.Field(&i.PartOfPackage, validation.NilOrNotEmpty),
		validation.Field(&i.Visibility, validation.NilOrNotEmpty, validation.In("public", "internal", "private")),
		validation.Field(&i.ReleaseStatus, validation.NilOrNotEmpty, validation.In(ReleaseStatusBeta, ReleaseStatusActive, ReleaseStatusDeprecated)),
		validation.Field(&i.Mandatory),
		validation.Field(&i.Aspects),
		validation.Field(&i.Version, validation.NilOrNotEmpty),
	)
}

// Validate validates AspectInput object
func (a AspectInput) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.Name, validation.Required, is.PrintableASCII, validation.Length(MinTitleLength, MaxTitleLength)),
		validation.Field(&a.Description, validation.NilOrNotEmpty, validation.Length(MinDescriptionLength, descriptionStringLengthLimit)),
		validation.Field(&a.Mandatory),
		validation.Field(&a.APIResources),
		validation.Field(&a.EventResources),
	)
}

// Validate validates AspectAPIDefinitionInput object
func (a AspectAPIDefinitionInput) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.OrdID, validation.Required, validation.Length(MinOrdIDLength, MaxOrdIDLength), validation.Match(regexp.MustCompile(AspectAPIResourceRegex))))
}

// Validate validates AspectEventDefinitionInput object
func (a AspectEventDefinitionInput) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.OrdID, validation.Required, validation.Length(MinOrdIDLength, MaxOrdIDLength), validation.Match(regexp.MustCompile(AspectEventResourceRegex))),
		validation.Field(&a.Subset))
}

// Validate validates AspectEventDefinitionSubsetInput object
func (a AspectEventDefinitionSubsetInput) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.EventType, validation.Required, validation.Match(regexp.MustCompile(EventResourceEventTypeRegex))))
}
