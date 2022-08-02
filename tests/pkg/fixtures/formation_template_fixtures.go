package fixtures

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

func FixFormationTemplate(formationName string) graphql.FormationTemplateInput {
	return FixFormationTemplateWithType(formationName, "runtime-type")
}

func FixFormationTemplateWithType(formationName string, runtimeType string) graphql.FormationTemplateInput {
	return graphql.FormationTemplateInput{
		Name:                   formationName,
		ApplicationTypes:       []string{"app-type-1", "app-type-2"},
		RuntimeType:            runtimeType,
		RuntimeTypeDisplayName: "test-display-name",
		RuntimeArtifactKind:    graphql.ArtifactTypeSubscription,
	}
}
