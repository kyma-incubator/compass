package fixtures

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func FixFormationTemplateInput(formationName string) graphql.FormationTemplateInput {
	return FixFormationTemplateInputWithType(formationName, "runtime-type")
}

func FixFormationTemplateInputWithType(formationName string, runtimeType string) graphql.FormationTemplateInput {
	return graphql.FormationTemplateInput{
		Name:                   formationName,
		ApplicationTypes:       []string{"app-type-1", "app-type-2"},
		RuntimeTypes:           []string{runtimeType},
		RuntimeTypeDisplayName: "runtime-type-display-name",
		RuntimeArtifactKind:    graphql.ArtifactTypeSubscription,
	}
}

func FixFormationTemplateInputWithTypes(formationName string, runtimeType string, applicationTypes []string) graphql.FormationTemplateInput {
	in := FixFormationTemplateInput(formationName)
	in.RuntimeTypes = []string{runtimeType}
	in.ApplicationTypes = applicationTypes
	return in
}

func FixFormationTemplateInputWithLeadingProductIDs(formationTemplateName, runtimeType string, applicationTypes []string, runtimeArtifactKind graphql.ArtifactType, leadingProductIDs []*string) graphql.FormationTemplateInput {
	return graphql.FormationTemplateInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           []string{runtimeType},
		RuntimeTypeDisplayName: "runtime-type-display-name",
		RuntimeArtifactKind:    runtimeArtifactKind,
		LeadingProductIDs:      leadingProductIDs,
	}
}
