package resource_providers

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	gcli "github.com/machinebox/graphql"
	"testing"
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
	c := &FormationTemplateProvider{formationTypeName: formationTypeName}
	return c
}

func (c *FormationTemplateProvider) WithWebhook(webhook Webhook) *FormationTemplateProvider {
	c.webhook = webhook
	return c
}
func (c *FormationTemplateProvider) WithConstraint(constraint Constraint) *FormationTemplateProvider {
	c.constraints = append(c.constraints, constraint)
	return c
}
func (c *FormationTemplateProvider) WithSupportedResources(resource ...Resource) *FormationTemplateProvider {
	c.supportedResources = append(c.supportedResources, resource...)
	return c
}

func (c *FormationTemplateProvider) WithLeadingProductIDs(leadingProductIDs []string) *FormationTemplateProvider {
	c.leadingProductIDs = leadingProductIDs
	return c
}

func (c *FormationTemplateProvider) WithSupportReset() *FormationTemplateProvider {
	c.supportsReset = true
	return c
}

func (c *FormationTemplateProvider) Provide(t *testing.T, ctx context.Context, gqlClient *gcli.Client) string {
	var applicationTypes []string
	var runtimeTypes []string
	var runtimeTypeDisplayName *string
	var runtimeArtifactKind *graphql.ArtifactType
	for _, resource := range c.supportedResources {
		if resource.GetType() == "APPLICATION" {
			applicationTypes = append(applicationTypes, resource.GetName())
		} else if resource.GetType() == "RUNTIME" {
			runtimeTypes = append(runtimeTypes, resource.GetName())
			runtimeArtifactKind = resource.GetArtifactKind()
			runtimeTypeDisplayName = resource.GetDisplayName()
		}
	}

	in := graphql.FormationTemplateInput{
		Name:                   c.formationTypeName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: runtimeTypeDisplayName,
		RuntimeArtifactKind:    runtimeArtifactKind,
		LeadingProductIDs:      c.leadingProductIDs,
		SupportsReset:          &c.supportsReset,
	}
	formationTemplate := fixtures.CreateFormationTemplate(t, ctx, gqlClient, in)
	c.template = formationTemplate

	if c.webhook != nil {
		c.webhook.AddToFormationTemplate(formationTemplate.ID)
	}

	for _, constraint := range c.constraints {
		constraint.Attach(formationTemplate.ID)
	}

	return formationTemplate.ID
}

func (c *FormationTemplateProvider) TearDown(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	fixtures.CleanupFormationTemplate(t, ctx, gqlClient, &c.template)
}
