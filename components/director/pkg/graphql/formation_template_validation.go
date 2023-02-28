package graphql

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

// Validate missing godoc
func (i FormationTemplateInput) Validate() error {
	fieldRules := []*validation.FieldRules{
		validation.Field(&i.Name, validation.Required, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.ApplicationTypes, validation.Required, inputvalidation.Each(validation.Required, validation.RuneLength(0, longStringLengthLimit))),
		validation.Field(&i.RuntimeTypes, inputvalidation.Each(validation.Required, validation.RuneLength(0, longStringLengthLimit))),
		validation.Field(&i.Webhooks, validation.By(webhooksRuleFunc)),
	}

	if i.RuntimeTypeDisplayName != nil && (len(*i.RuntimeTypeDisplayName) == 0 || len(*i.RuntimeTypeDisplayName) > longLongStringLengthLimit) {
		return apperrors.NewInvalidDataError(fmt.Sprintf("Invalid %q: length should be between %q and %q", "RuntimeTypeDisplayName", 0, longLongStringLengthLimit))
	}

	if !validateRuntimeArtifactKind(i) {
		return apperrors.NewInvalidDataError(fmt.Sprintf("Invalid %q: should be one of %s", "RuntimeArtifactKind", AllArtifactType))
	}

	if !allRuntimeFieldsArePresent(i) && !allRuntimeFieldsAreMissing(i) {
		return apperrors.NewInvalidDataError("Either all RuntimeTypes, RuntimeArtifactKind and RuntimeArtifactKind fields should be present or all of them should be missing")
	}

	return validation.ValidateStruct(&i, fieldRules...)
}

func validateRuntimeArtifactKind(i FormationTemplateInput) bool {
	if i.RuntimeArtifactKind != nil {
		artifactTypeIsValid := false
		for _, artifactType := range AllArtifactType {
			if *i.RuntimeArtifactKind == artifactType {
				artifactTypeIsValid = true
			}
		}
		return artifactTypeIsValid
	}

	return true
}

func allRuntimeFieldsAreMissing(i FormationTemplateInput) bool {
	return i.RuntimeArtifactKind == nil && i.RuntimeTypeDisplayName == nil && len(i.RuntimeTypes) == 0
}

func allRuntimeFieldsArePresent(i FormationTemplateInput) bool {
	return i.RuntimeArtifactKind != nil && i.RuntimeTypeDisplayName != nil && len(i.RuntimeTypes) > 0
}
