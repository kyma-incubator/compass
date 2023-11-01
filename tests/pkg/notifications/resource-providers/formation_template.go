package resource_providers

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	gcli "github.com/machinebox/graphql"
)

type Constraint interface {
	Attach(templateID string)
}

type Webhook interface {
	AddToFormationTemplate(templateID string)
}

type FormationTemplateProvider struct {
	formationTypeName  string
	supportsReset      bool
	leadingProductIDs  []string
	supportedResources []Resource
	constraints        []Constraint
	webhook            Webhook
	template           graphql.FormationTemplate
}

func NewFormationTemplateCreator(formationTypeName string) *FormationTemplateProvider {
	return &FormationTemplateProvider{formationTypeName: formationTypeName}
}

func (p *FormationTemplateProvider) WithWebhook(webhook Webhook) *FormationTemplateProvider {
	p.webhook = webhook
	return p
}
func (p *FormationTemplateProvider) WithConstraint(constraint Constraint) *FormationTemplateProvider {
	p.constraints = append(p.constraints, constraint)
	return p
}
func (p *FormationTemplateProvider) WithSupportedResources(resource ...Resource) *FormationTemplateProvider {
	p.supportedResources = append(p.supportedResources, resource...)
	return p
}

func (p *FormationTemplateProvider) WithLeadingProductIDs(leadingProductIDs []string) *FormationTemplateProvider {
	p.leadingProductIDs = leadingProductIDs
	return p
}

func (p *FormationTemplateProvider) WithSupportReset() *FormationTemplateProvider {
	p.supportsReset = true
	return p
}

func (p *FormationTemplateProvider) Provide(t *testing.T, ctx context.Context, gqlClient *gcli.Client) string {
	var applicationTypes []string
	var runtimeTypes []string
	var runtimeTypeDisplayName *string
	var runtimeArtifactKind *graphql.ArtifactType
	for _, resource := range p.supportedResources {
		if resource.GetType() == graphql.ResourceTypeApplication {
			applicationTypes = append(applicationTypes, resource.GetName())
		} else if resource.GetType() == graphql.ResourceTypeRuntime {
			runtimeTypes = append(runtimeTypes, resource.GetName())
			runtimeArtifactKind = resource.GetArtifactKind()
			runtimeTypeDisplayName = resource.GetDisplayName()
		}
	}

	in := graphql.FormationTemplateInput{
		Name:                   p.formationTypeName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: runtimeTypeDisplayName,
		RuntimeArtifactKind:    runtimeArtifactKind,
		LeadingProductIDs:      p.leadingProductIDs,
		SupportsReset:          &p.supportsReset,
	}
	formationTemplate := fixtures.CreateFormationTemplate(t, ctx, gqlClient, in)
	p.template = formationTemplate

	if p.webhook != nil {
		p.webhook.AddToFormationTemplate(formationTemplate.ID)
	}

	for _, constraint := range p.constraints {
		constraint.Attach(formationTemplate.ID)
	}

	return formationTemplate.ID
}

func (p *FormationTemplateProvider) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	fixtures.CleanupFormationTemplate(t, ctx, gqlClient, &p.template)
}
