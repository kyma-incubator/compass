package fixtures

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

const (
	FormationTemplateLabelKey   = "e2eTestFormationTemplateLabelKey"
	FormationTemplateLabelValue = "e2eTestFormationTemplateLabelValue"
)

func FixFormationTemplateRegisterInput(formationName string) graphql.FormationTemplateRegisterInput {
	return FixFormationTemplateRegisterInputWithRuntimeTypes(formationName, []string{"runtime-type"})
}

func FixFormationTemplateRegisterInputWithApplicationTypes(formationName string, applicationTypes []string) graphql.FormationTemplateRegisterInput {
	subscription := graphql.ArtifactTypeSubscription
	return graphql.FormationTemplateRegisterInput{
		Name:                   formationName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           []string{"runtime-type"},
		RuntimeTypeDisplayName: str.Ptr("runtime-type-display-name"),
		RuntimeArtifactKind:    &subscription,
	}
}

func FixFormationTemplateRegisterInputWithRuntimeTypes(formationName string, runtimeTypes []string) graphql.FormationTemplateRegisterInput {
	subscription := graphql.ArtifactTypeSubscription
	return graphql.FormationTemplateRegisterInput{
		Name:                   formationName,
		ApplicationTypes:       []string{"app-type-1", "app-type-2"},
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: str.Ptr("runtime-type-display-name"),
		RuntimeArtifactKind:    &subscription,
		Labels:                 map[string]interface{}{FormationTemplateLabelKey: FormationTemplateLabelValue},
	}
}

func FixFormationTemplateRegisterInputWithTypes(formationName string, runtimeTypes, applicationTypes []string) graphql.FormationTemplateRegisterInput {
	in := FixFormationTemplateRegisterInput(formationName)
	in.RuntimeTypes = runtimeTypes
	in.ApplicationTypes = applicationTypes
	return in
}

func FixAppOnlyFormationTemplateRegisterInput(formationName string) graphql.FormationTemplateRegisterInput {
	return graphql.FormationTemplateRegisterInput{
		Name:             formationName,
		ApplicationTypes: []string{"app-type-1", "app-type-2"},
	}
}

func FixInvalidFormationTemplateRegisterInputWithRuntimeArtifactKind(formationName string) graphql.FormationTemplateRegisterInput {
	subscription := graphql.ArtifactTypeSubscription
	return graphql.FormationTemplateRegisterInput{
		Name:                formationName,
		ApplicationTypes:    []string{"app-type-1", "app-type-2"},
		RuntimeArtifactKind: &subscription,
	}
}

func FixInvalidFormationTemplateUpdateInputWithRuntimeArtifactKind(formationName string) graphql.FormationTemplateUpdateInput {
	subscription := graphql.ArtifactTypeSubscription
	return graphql.FormationTemplateUpdateInput{
		Name:                formationName,
		ApplicationTypes:    []string{"app-type-1", "app-type-2"},
		RuntimeArtifactKind: &subscription,
	}
}

func FixInvalidFormationTemplateRegisterInputWithRuntimeTypeDisplayName(formationName string) graphql.FormationTemplateRegisterInput {
	return graphql.FormationTemplateRegisterInput{
		Name:                   formationName,
		ApplicationTypes:       []string{"app-type-1", "app-type-2"},
		RuntimeTypeDisplayName: str.Ptr("runtime-type-display-name"),
	}
}

func FixInvalidFormationTemplateUpdateInputWithRuntimeTypeDisplayName(formationName string) graphql.FormationTemplateUpdateInput {
	return graphql.FormationTemplateUpdateInput{
		Name:                   formationName,
		ApplicationTypes:       []string{"app-type-1", "app-type-2"},
		RuntimeTypeDisplayName: str.Ptr("runtime-type-display-name"),
	}
}

func FixInvalidFormationTemplateRegisterInputWithRuntimeTypes(formationName string, runtimeType string) graphql.FormationTemplateRegisterInput {
	return graphql.FormationTemplateRegisterInput{
		Name:             formationName,
		ApplicationTypes: []string{"app-type-1", "app-type-2"},
		RuntimeTypes:     []string{runtimeType},
	}
}

func FixInvalidFormationTemplateUpdateInputWithRuntimeTypes(formationName string, runtimeType string) graphql.FormationTemplateUpdateInput {
	return graphql.FormationTemplateUpdateInput{
		Name:             formationName,
		ApplicationTypes: []string{"app-type-1", "app-type-2"},
		RuntimeTypes:     []string{runtimeType},
	}
}

func FixInvalidFormationTemplateRegisterInputWithoutArtifactKind(formationName string, runtimeType string) graphql.FormationTemplateRegisterInput {
	return graphql.FormationTemplateRegisterInput{
		Name:                   formationName,
		ApplicationTypes:       []string{"app-type-1", "app-type-2"},
		RuntimeTypes:           []string{runtimeType},
		RuntimeTypeDisplayName: str.Ptr("runtime-type-display-name"),
	}
}

func FixInvalidFormationTemplateUpdateInputWithoutArtifactKind(formationName string, runtimeType string) graphql.FormationTemplateUpdateInput {
	return graphql.FormationTemplateUpdateInput{
		Name:                   formationName,
		ApplicationTypes:       []string{"app-type-1", "app-type-2"},
		RuntimeTypes:           []string{runtimeType},
		RuntimeTypeDisplayName: str.Ptr("runtime-type-display-name"),
	}
}

func FixInvalidFormationTemplateRegisterInputWithoutDisplayName(formationName string, runtimeType string) graphql.FormationTemplateRegisterInput {
	subscription := graphql.ArtifactTypeSubscription
	return graphql.FormationTemplateRegisterInput{
		Name:                formationName,
		ApplicationTypes:    []string{"app-type-1", "app-type-2"},
		RuntimeTypes:        []string{runtimeType},
		RuntimeArtifactKind: &subscription,
	}
}

func FixInvalidFormationTemplateUpdateInputWithoutDisplayName(formationName string, runtimeType string) graphql.FormationTemplateUpdateInput {
	subscription := graphql.ArtifactTypeSubscription
	return graphql.FormationTemplateUpdateInput{
		Name:                formationName,
		ApplicationTypes:    []string{"app-type-1", "app-type-2"},
		RuntimeTypes:        []string{runtimeType},
		RuntimeArtifactKind: &subscription,
	}
}

func FixFormationTemplateRegisterInputWithLeadingProductIDs(formationTemplateName string, applicationTypes, runtimeTypes []string, runtimeArtifactKind graphql.ArtifactType, leadingProductIDs []string) graphql.FormationTemplateRegisterInput {
	return graphql.FormationTemplateRegisterInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: &formationTemplateName,
		RuntimeArtifactKind:    &runtimeArtifactKind,
		LeadingProductIDs:      leadingProductIDs,
	}
}
