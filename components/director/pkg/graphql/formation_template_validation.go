package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

// Validate missing godoc
func (i FormationTemplateInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.ApplicationTypes, validation.Required, inputvalidation.Each(validation.Required, validation.RuneLength(0, longStringLengthLimit))),
		validation.Field(&i.RuntimeTypes, validation.Required, inputvalidation.Each(validation.Required, validation.RuneLength(0, longStringLengthLimit))),
		validation.Field(&i.RuntimeTypeDisplayName, validation.Required, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.RuntimeArtifactKind, validation.Required, validation.In(ArtifactTypeSubscription, ArtifactTypeServiceInstance, ArtifactTypeEnvironmentInstance)))
}
