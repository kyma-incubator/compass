package fixtures

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

func FixFormationTemplate(formationName string) graphql.FormationTemplateInput {
	return FixFormationTemplateWithType(formationName, "runtime-type")
}

func FixFormationTemplateWithType(formationName string, runtimeType string) graphql.FormationTemplateInput {
	return graphql.FormationTemplateInput{
		Name:                   formationName,
		ApplicationTypes:       []string{"app-type-1", "app-type-2"},
		RuntimeType:            str.Ptr(runtimeType),
		RuntimeTypeDisplayName: "test-display-name",
		RuntimeArtifactKind:    graphql.ArtifactTypeSubscription,
	}
}

func FixFormationTemplateWithApplicationTypes(formationName string, applicationTypes []string) graphql.FormationTemplateInput {
	in := FixFormationTemplate(formationName)
	in.ApplicationTypes = applicationTypes
	return in
}

func FixFormationTemplateWithRuntimeType(formationName string, runtimeType string) graphql.FormationTemplateInput {
	in := FixFormationTemplate(formationName)
	in.RuntimeType = str.Ptr(runtimeType)
	return in
}

func FixFormationTypeWithTypes(formationName string, runtimeType string, applicationTypes []string) graphql.FormationTemplateInput {
	in := FixFormationTemplate(formationName)
	in.RuntimeType = str.Ptr(runtimeType)
	in.ApplicationTypes = applicationTypes
	return in
}
