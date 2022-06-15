package fixtures

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

func FixFormationTemplate(formationName string) graphql.FormationTemplateInput {
	return graphql.FormationTemplateInput{
		Name:                   formationName,
		ApplicationTypes:       []string{"app-type-1", "app-type-2"},
		RuntimeType:            "runtime-type",
		RuntimeTypeDisplayName: "test-display-name",
		RuntimeArtifactKind:    graphql.ArtifactTypeSubscription,
	}
}
