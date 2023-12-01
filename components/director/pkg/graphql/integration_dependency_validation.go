package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/kyma-incubator/compass/components/director/internal/common"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"regexp"
)

// Validate validates IntegrationDependencyInput object
func (i IntegrationDependencyInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, is.PrintableASCII, validation.Length(common.MinTitleLength, common.MaxTitleLength)),
		validation.Field(&i.Description, validation.NilOrNotEmpty, validation.Length(common.MinDescriptionLength, descriptionStringLengthLimit)),
		validation.Field(&i.OrdID, validation.NilOrNotEmpty, validation.Length(common.MinTitleLength, common.MaxTitleLength), validation.Match(regexp.MustCompile(common.IntegrationDependencyOrdIDRegex))),
		validation.Field(&i.PartOfPackage, validation.NilOrNotEmpty, validation.Length(common.MinTitleLength, common.MaxTitleLength), validation.Match(regexp.MustCompile(common.PackageOrdIDRegex))),
		validation.Field(&i.Visibility, validation.NilOrNotEmpty, validation.In("public", "internal", "private")),
		validation.Field(&i.ReleaseStatus, validation.NilOrNotEmpty, validation.In(common.ReleaseStatusBeta, common.ReleaseStatusActive, common.ReleaseStatusDeprecated)),
		validation.Field(&i.Mandatory),
		validation.Field(&i.Aspects),
		validation.Field(&i.Version, validation.NilOrNotEmpty),
	)
}

// Validate validates AspectInput object
func (a AspectInput) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.Name, validation.Required, is.PrintableASCII, validation.Length(common.MinTitleLength, common.MaxTitleLength)),
		validation.Field(&a.Description, validation.NilOrNotEmpty, validation.Length(common.MinDescriptionLength, descriptionStringLengthLimit)),
		validation.Field(&a.Mandatory),
		validation.Field(&a.APIResources),
		validation.Field(&a.EventResources),
	)
}

// Validate validates AspectAPIDefinitionInput object
func (a AspectAPIDefinitionInput) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.OrdID, validation.Required, validation.Length(common.MinOrdIDLength, common.MaxOrdIDLength), validation.Match(regexp.MustCompile(common.AspectAPIResourceRegex))))
}

// Validate validates AspectEventDefinitionInput object
func (a AspectEventDefinitionInput) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.OrdID, validation.Required, validation.Length(common.MinOrdIDLength, common.MaxOrdIDLength), validation.Match(regexp.MustCompile(common.AspectEventResourceRegex))),
		validation.Field(&a.Subset, validation.Required, inputvalidation.Each(validation.Match(regexp.MustCompile(common.EventResourceEventTypeRegex)))))
}
