package fixtures

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

func FixFormationTemplateInput(formationName string) graphql.FormationTemplateInput {
	return FixFormationTemplateInputWithType(formationName, "runtime-type")
}

func FixFormationTemplateInputWithType(formationName string, runtimeType string) graphql.FormationTemplateInput {
	subscription := graphql.ArtifactTypeSubscription
	return graphql.FormationTemplateInput{
		Name:                   formationName,
		ApplicationTypes:       []string{"app-type-1", "app-type-2"},
		RuntimeTypes:           []string{runtimeType},
		RuntimeTypeDisplayName: str.Ptr("runtime-type-display-name"),
		RuntimeArtifactKind:    &subscription,
	}
}

func FixAppOnlyFormationTemplateInput(formationName string) graphql.FormationTemplateInput {
	return graphql.FormationTemplateInput{
		Name:             formationName,
		ApplicationTypes: []string{"app-type-1", "app-type-2"},
	}
}

func FixInvalidFormationTemplateInputWithRuntimeArtifactKind(formationName string) graphql.FormationTemplateInput {
	subscription := graphql.ArtifactTypeSubscription
	return graphql.FormationTemplateInput{
		Name:                formationName,
		ApplicationTypes:    []string{"app-type-1", "app-type-2"},
		RuntimeArtifactKind: &subscription,
	}
}

func FixInvalidFormationTemplateInputWithRuntimeTypeDisplayName(formationName string) graphql.FormationTemplateInput {
	return graphql.FormationTemplateInput{
		Name:                   formationName,
		ApplicationTypes:       []string{"app-type-1", "app-type-2"},
		RuntimeTypeDisplayName: str.Ptr("runtime-type-display-name"),
	}
}

func FixInvalidFormationTemplateInputWithRuntimeTypes(formationName string, runtimeType string) graphql.FormationTemplateInput {
	return graphql.FormationTemplateInput{
		Name:             formationName,
		ApplicationTypes: []string{"app-type-1", "app-type-2"},
		RuntimeTypes:     []string{runtimeType},
	}
}

func FixFormationTemplateInputWithTypes(formationName string, runtimeType string, applicationTypes []string) graphql.FormationTemplateInput {
	in := FixFormationTemplateInput(formationName)
	in.RuntimeTypes = []string{runtimeType}
	in.ApplicationTypes = applicationTypes
	return in
}

func FixFormationTemplateInputWithLeadingProductIDs(formationTemplateName, runtimeType string, applicationTypes []string, runtimeArtifactKind graphql.ArtifactType, leadingProductIDs []string) graphql.FormationTemplateInput {
	return graphql.FormationTemplateInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           []string{runtimeType},
		RuntimeTypeDisplayName: str.Ptr("runtime-type-display-name"),
		RuntimeArtifactKind:    &runtimeArtifactKind,
		LeadingProductIDs:      leadingProductIDs,
	}
}
