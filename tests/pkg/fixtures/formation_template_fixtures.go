package fixtures

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

func FixFormationTemplateInput(formationName string) graphql.FormationTemplateInput {
	return FixFormationTemplateInputWithRuntimeTypes(formationName, []string{"runtime-type"})
}

func FixFormationTemplateInputWithRuntimeTypes(formationName string, runtimeTypes []string) graphql.FormationTemplateInput {
	subscription := graphql.ArtifactTypeSubscription
	return graphql.FormationTemplateInput{
		Name:                   formationName,
		ApplicationTypes:       []string{"app-type-1", "app-type-2"},
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: str.Ptr("runtime-type-display-name"),
		RuntimeArtifactKind:    &subscription,
	}
}

func FixFormationTemplateInputWithTypes(formationName string, runtimeTypes, applicationTypes []string) graphql.FormationTemplateInput {
	in := FixFormationTemplateInput(formationName)
	in.RuntimeTypes = runtimeTypes
	in.ApplicationTypes = applicationTypes
	return in
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

func FixInvalidFormationTemplateInputWithoutArtifactKind(formationName string, runtimeType string) graphql.FormationTemplateInput {
	return graphql.FormationTemplateInput{
		Name:                   formationName,
		ApplicationTypes:       []string{"app-type-1", "app-type-2"},
		RuntimeTypes:           []string{runtimeType},
		RuntimeTypeDisplayName: str.Ptr("runtime-type-display-name"),
	}
}

func FixInvalidFormationTemplateInputWithoutDisplayName(formationName string, runtimeType string) graphql.FormationTemplateInput {
	subscription := graphql.ArtifactTypeSubscription
	return graphql.FormationTemplateInput{
		Name:                formationName,
		ApplicationTypes:    []string{"app-type-1", "app-type-2"},
		RuntimeTypes:        []string{runtimeType},
		RuntimeArtifactKind: &subscription,
	}
}

func FixFormationTemplateInputWithLeadingProductIDs(formationTemplateName string, applicationTypes, runtimeTypes []string, runtimeArtifactKind graphql.ArtifactType, leadingProductIDs []string) graphql.FormationTemplateInput {
	return graphql.FormationTemplateInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: &formationTemplateName,
		RuntimeArtifactKind:    &runtimeArtifactKind,
		LeadingProductIDs:      leadingProductIDs,
	}
}

func FixFormationTemplateInputWithWebhook(formationTemplateName string, applicationTypes, runtimeTypes []string, runtimeArtifactKind graphql.ArtifactType, webhookType graphql.WebhookType, webhookMode graphql.WebhookMode, urlTemplate, inputTemplate, outputTemplate string) graphql.FormationTemplateInput {
	return graphql.FormationTemplateInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: &formationTemplateName,
		RuntimeArtifactKind:    &runtimeArtifactKind,
		Webhooks: []*graphql.WebhookInput{
			{
				Type: webhookType,
				Auth: &graphql.AuthInput{
					AccessStrategy: str.Ptr("sap:cmp-mtls:v1"),
				},
				Mode:           &webhookMode,
				URLTemplate:    &urlTemplate,
				InputTemplate:  &inputTemplate,
				OutputTemplate: &outputTemplate,
			},
		},
	}
}

func FixFormationTemplateInputWithWebhookAndLeadingProductIDs(formationTemplateName string, applicationTypes, runtimeTypes []string, runtimeArtifactKind graphql.ArtifactType, leadingProductIDs []string, webhookType graphql.WebhookType, webhookMode graphql.WebhookMode, urlTemplate, inputTemplate, outputTemplate string) graphql.FormationTemplateInput {
	ftInput := FixFormationTemplateInputWithLeadingProductIDs(formationTemplateName, applicationTypes, runtimeTypes, runtimeArtifactKind, leadingProductIDs)
	ftInput.Webhooks = []*graphql.WebhookInput{
		{
			Type: webhookType,
			Auth: &graphql.AuthInput{
				AccessStrategy: str.Ptr("sap:cmp-mtls:v1"),
			},
			Mode:           &webhookMode,
			URLTemplate:    &urlTemplate,
			InputTemplate:  &inputTemplate,
			OutputTemplate: &outputTemplate,
		},
	}
	return ftInput
}
