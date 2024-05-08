package graphql

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

// Validate missing godoc
func (i FormationTemplateRegisterInput) Validate() error {
	subtypes := make([]interface{}, 0, len(i.RuntimeTypes)+len(i.ApplicationTypes))
	for _, subtype := range i.RuntimeTypes {
		subtypes = append(subtypes, subtype)
	}
	for _, subtype := range i.ApplicationTypes {
		subtypes = append(subtypes, subtype)
	}

	fieldRules := []*validation.FieldRules{
		validation.Field(&i.Name, validation.Required, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.ApplicationTypes, validation.Required, inputvalidation.Each(validation.Required, validation.RuneLength(0, longStringLengthLimit))),
		validation.Field(&i.RuntimeTypes, inputvalidation.Each(validation.Required, validation.RuneLength(0, longStringLengthLimit))),
		validation.Field(&i.DiscoveryConsumers, validation.Each(validation.Required, validation.In(subtypes...))),
		validation.Field(&i.Webhooks, validation.By(webhooksRuleFunc)),
		validation.Field(&i.Labels, inputvalidation.EachKey(validation.Required, validation.Match(alphanumericUnderscoreRegexp))),
	}

	if i.RuntimeTypeDisplayName != nil && (len(*i.RuntimeTypeDisplayName) == 0 || len(*i.RuntimeTypeDisplayName) > longLongStringLengthLimit) {
		return apperrors.NewInvalidDataError(fmt.Sprintf("Invalid %q: length should be between %q and %q", "RuntimeTypeDisplayName", 0, longLongStringLengthLimit))
	}

	if !validateRuntimeArtifactKind(i.RuntimeArtifactKind) {
		return apperrors.NewInvalidDataError(fmt.Sprintf("Invalid %q: should be one of %s", "RuntimeArtifactKind", AllArtifactType))
	}

	if !allRuntimeFieldsArePresent(i.RuntimeArtifactKind, i.RuntimeTypeDisplayName, i.RuntimeTypes) && !allRuntimeFieldsAreMissing(i.RuntimeArtifactKind, i.RuntimeTypeDisplayName, i.RuntimeTypes) {
		return apperrors.NewInvalidDataError("Either all RuntimeTypes, RuntimeArtifactKind and RuntimeArtifactKind fields should be present or all of them should be missing")
	}

	return validation.ValidateStruct(&i, fieldRules...)
}

func (i FormationTemplateUpdateInput) Validate() error {
	subtypes := make([]interface{}, 0, len(i.RuntimeTypes)+len(i.ApplicationTypes))
	for _, subtype := range i.RuntimeTypes {
		subtypes = append(subtypes, subtype)
	}
	for _, subtype := range i.ApplicationTypes {
		subtypes = append(subtypes, subtype)
	}

	fieldRules := []*validation.FieldRules{
		validation.Field(&i.Name, validation.Required, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.ApplicationTypes, validation.Required, inputvalidation.Each(validation.Required, validation.RuneLength(0, longStringLengthLimit))),
		validation.Field(&i.RuntimeTypes, inputvalidation.Each(validation.Required, validation.RuneLength(0, longStringLengthLimit))),
		validation.Field(&i.DiscoveryConsumers, validation.Each(validation.Required, validation.In(subtypes...))),
	}

	if i.RuntimeTypeDisplayName != nil && (len(*i.RuntimeTypeDisplayName) == 0 || len(*i.RuntimeTypeDisplayName) > longLongStringLengthLimit) {
		return apperrors.NewInvalidDataError(fmt.Sprintf("Invalid %q: length should be between %q and %q", "RuntimeTypeDisplayName", 0, longLongStringLengthLimit))
	}

	if !validateRuntimeArtifactKind(i.RuntimeArtifactKind) {
		return apperrors.NewInvalidDataError(fmt.Sprintf("Invalid %q: should be one of %s", "RuntimeArtifactKind", AllArtifactType))
	}

	if !allRuntimeFieldsArePresent(i.RuntimeArtifactKind, i.RuntimeTypeDisplayName, i.RuntimeTypes) && !allRuntimeFieldsAreMissing(i.RuntimeArtifactKind, i.RuntimeTypeDisplayName, i.RuntimeTypes) {
		return apperrors.NewInvalidDataError("Either all RuntimeTypes, RuntimeArtifactKind and RuntimeArtifactKind fields should be present or all of them should be missing")
	}

	return validation.ValidateStruct(&i, fieldRules...)
}

func validateRuntimeArtifactKind(runtimeArtifactKind *ArtifactType) bool {
	if runtimeArtifactKind != nil {
		isArtifactTypeValid := false
		for _, artifactType := range AllArtifactType {
			if *runtimeArtifactKind == artifactType {
				isArtifactTypeValid = true
				break
			}
		}
		return isArtifactTypeValid
	}

	return true
}

func allRuntimeFieldsAreMissing(runtimeArtifactKind *ArtifactType, runtimeTypeDisplayName *string, runtimeTypes []string) bool {
	return runtimeArtifactKind == nil && runtimeTypeDisplayName == nil && len(runtimeTypes) == 0
}

func allRuntimeFieldsArePresent(runtimeArtifactKind *ArtifactType, runtimeTypeDisplayName *string, runtimeTypes []string) bool {
	return runtimeArtifactKind != nil && runtimeTypeDisplayName != nil && len(runtimeTypes) > 0
}
