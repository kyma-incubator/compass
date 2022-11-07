package fixtures

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

func FixFormationTemplateInput (formationName string) graphql.FormationTemplateInput {
	return FixFormationTemplateInputWithType(formationName, "runtime-type")
}

func FixFormationTemplateInputWithType(formationName string, runtimeType string) graphql.FormationTemplateInput {
	return graphql.FormationTemplateInput{
		Name:                   formationName,
		ApplicationTypes:       []string{"app-type-1", "app-type-2"},
		RuntimeType:            &runtimeType,
		RuntimeTypeDisplayName: "test-display-name",
		RuntimeArtifactKind:    graphql.ArtifactTypeSubscription,
	}
}

func FixFormationTemplateInputWithApplicationTypes(formationName string, applicationTypes []string) graphql.FormationTemplateInput {
	in := FixFormationTemplateInput(formationName)
	in.ApplicationTypes = applicationTypes
	return in
}

func FixFormationTemplateInputWithRuntimeType(formationName string, runtimeType string) graphql.FormationTemplateInput {
	in := FixFormationTemplateInput(formationName)
	in.RuntimeType = &runtimeType
	return in
}

func FixFormationTemplateInputWithTypes(formationName string, runtimeType string, applicationTypes []string) graphql.FormationTemplateInput {
	in := FixFormationTemplateInput(formationName)
	in.RuntimeType = &runtimeType
	in.ApplicationTypes = applicationTypes
	return in
}
