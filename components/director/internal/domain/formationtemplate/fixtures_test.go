package formationtemplate_test

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
)

const (
	testID                     = "d1fddec6-5456-4a1e-9ae0-74447f5d6ae9"
	formationTemplateName      = "formation-template-name"
	applicationTypesAsString   = "[\"some-application-type\"]"
	runtimeTypesAsString       = "[\"some-runtime-type\"]"
	testTenantID               = "d9fddec6-5456-4a1e-9ae0-74447f5d6ae9"
	testParentTenantID         = "d8fddec6-5456-4a1e-9ae0-74447f5d6ae9"
	testWebhookID              = "test-wh-id"
	leadingProductIDsAsString  = "[\"leading-product-id\",\"leading-product-id-2\"]"
	discoveryConsumersAsString = "[\"some-application-type\",\"some-runtime-type\"]"
)

var (
	runtimeTypeDisplayName      = "display-name-for-runtime"
	artifactKindAsString        = "SUBSCRIPTION"
	runtimeArtifactKind         = model.RuntimeArtifactKindSubscription
	artifactKind                = graphql.ArtifactTypeSubscription
	nilModelEntity              *model.FormationTemplate
	emptyTemplate               = `{}`
	url                         = "http://foo.com"
	modelWebhookMode            = model.WebhookModeSync
	graphqlWebhookMode          = graphql.WebhookModeSync
	applicationTypes            = []string{"some-application-type"}
	runtimeTypes                = []string{"some-runtime-type"}
	leadingProductIDs           = []string{"leading-product-id", "leading-product-id-2"}
	discoveryConsumers          = []string{"some-application-type", "some-runtime-type"}
	shouldReset                 = true
	formationTemplateModelInput = model.FormationTemplateInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: &runtimeTypeDisplayName,
		RuntimeArtifactKind:    &runtimeArtifactKind,
		LeadingProductIDs:      leadingProductIDs,
		DiscoveryConsumers:     discoveryConsumers,
		Webhooks:               fixModelWebhookInput(),
	}
	formationTemplateModelWithResetInput = model.FormationTemplateInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: &runtimeTypeDisplayName,
		RuntimeArtifactKind:    &runtimeArtifactKind,
		LeadingProductIDs:      leadingProductIDs,
		DiscoveryConsumers:     discoveryConsumers,
		Webhooks:               fixModelWebhookInput(),
		SupportsReset:          shouldReset,
	}
	formationTemplateGraphQLInput = graphql.FormationTemplateInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: str.Ptr(runtimeTypeDisplayName),
		RuntimeArtifactKind:    &artifactKind,
		LeadingProductIDs:      leadingProductIDs,
		DiscoveryConsumers:     discoveryConsumers,
		Webhooks:               fixGQLWebhookInput(),
	}
	formationTemplateWithResetGraphQLInput = graphql.FormationTemplateInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: str.Ptr(runtimeTypeDisplayName),
		RuntimeArtifactKind:    &artifactKind,
		LeadingProductIDs:      leadingProductIDs,
		DiscoveryConsumers:     discoveryConsumers,
		Webhooks:               fixGQLWebhookInput(),
		SupportsReset:          &shouldReset,
	}

	formationTemplateModelInputAppOnly = model.FormationTemplateInput{
		Name:               formationTemplateName,
		ApplicationTypes:   applicationTypes,
		LeadingProductIDs:  leadingProductIDs,
		DiscoveryConsumers: discoveryConsumers,
		Webhooks:           fixModelWebhookInput(),
	}
	formationTemplateGraphQLInputAppOnly = graphql.FormationTemplateInput{
		Name:               formationTemplateName,
		ApplicationTypes:   applicationTypes,
		LeadingProductIDs:  leadingProductIDs,
		DiscoveryConsumers: discoveryConsumers,
		Webhooks:           fixGQLWebhookInput(),
	}
	formationTemplateModelAppOnly = model.FormationTemplate{
		ID:                 testID,
		Name:               formationTemplateName,
		ApplicationTypes:   applicationTypes,
		LeadingProductIDs:  leadingProductIDs,
		DiscoveryConsumers: discoveryConsumers,
		TenantID:           str.Ptr(testTenantID),
		Webhooks:           []*model.Webhook{fixFormationTemplateModelWebhook()},
	}

	formationTemplateModel = model.FormationTemplate{
		ID:                     testID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: &runtimeTypeDisplayName,
		RuntimeArtifactKind:    &runtimeArtifactKind,
		LeadingProductIDs:      leadingProductIDs,
		DiscoveryConsumers:     discoveryConsumers,
		TenantID:               str.Ptr(testTenantID),
		Webhooks:               []*model.Webhook{fixFormationTemplateModelWebhook()},
	}
	formationTemplateModelNullTenant = model.FormationTemplate{
		ID:                     testID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: &runtimeTypeDisplayName,
		RuntimeArtifactKind:    &runtimeArtifactKind,
		LeadingProductIDs:      leadingProductIDs,
		DiscoveryConsumers:     discoveryConsumers,
		TenantID:               nil,
	}
	formationTemplateEntity = formationtemplate.Entity{
		ID:                     testID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypesAsString,
		RuntimeTypes:           repo.NewValidNullableString(runtimeTypesAsString),
		RuntimeTypeDisplayName: repo.NewValidNullableString(runtimeTypeDisplayName),
		RuntimeArtifactKind:    repo.NewValidNullableString(artifactKindAsString),
		LeadingProductIDs:      repo.NewNullableStringFromJSONRawMessage(json.RawMessage(leadingProductIDsAsString)),
		TenantID:               repo.NewValidNullableString(testTenantID),
		DiscoveryConsumers:     repo.NewNullableStringFromJSONRawMessage(json.RawMessage(discoveryConsumersAsString)),
	}
	formationTemplateEntityNullTenant = formationtemplate.Entity{
		ID:                     testID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypesAsString,
		RuntimeTypes:           repo.NewValidNullableString(runtimeTypesAsString),
		RuntimeTypeDisplayName: repo.NewValidNullableString(runtimeTypeDisplayName),
		RuntimeArtifactKind:    repo.NewValidNullableString(artifactKindAsString),
		LeadingProductIDs:      repo.NewNullableStringFromJSONRawMessage(json.RawMessage(leadingProductIDsAsString)),
		TenantID:               repo.NewValidNullableString(""),
		DiscoveryConsumers:     repo.NewNullableStringFromJSONRawMessage(json.RawMessage(discoveryConsumersAsString)),
	}
	graphQLFormationTemplate = graphql.FormationTemplate{
		ID:                     testID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: &runtimeTypeDisplayName,
		RuntimeArtifactKind:    &artifactKind,
		LeadingProductIDs:      leadingProductIDs,
		DiscoveryConsumers:     discoveryConsumers,
		Webhooks:               []*graphql.Webhook{fixFormationTemplateGQLWebhook()},
	}
	formationTemplateModelPage = model.FormationTemplatePage{
		Data: []*model.FormationTemplate{&formationTemplateModel},
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: 1,
	}
	formationTemplateModelNullTenantPage = model.FormationTemplatePage{
		Data: []*model.FormationTemplate{&formationTemplateModelNullTenant},
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: 1,
	}
	graphQLFormationTemplatePage = graphql.FormationTemplatePage{
		Data: []*graphql.FormationTemplate{&graphQLFormationTemplate},
		PageInfo: &graphql.PageInfo{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: 1,
	}

	constraintID1           = "constraintID1"
	constraintID2           = "constraintID2"
	operatorName            = operators.IsNotAssignedToAnyFormationOfTypeOperator
	formationConstraintName = "constraint-name"
	resourceSubtype         = "test subtype"
	inputTemplate           = `{"formation_template_id": "{{.FormationTemplateID}}","resource_type": "{{.ResourceType}}","resource_subtype": "{{.ResourceSubtype}}","resource_id": "{{.ResourceID}}","tenant": "{{.TenantID}}"}`

	formationConstraint1 = &model.FormationConstraint{
		ID:              constraintID1,
		Name:            formationConstraintName,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
	}
	formationConstraint2 = &model.FormationConstraint{
		ID:              constraintID2,
		Name:            formationConstraintName,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
	}

	formationConstraintGql1 = &graphql.FormationConstraint{
		ID:              constraintID1,
		Name:            formationConstraintName,
		ConstraintType:  graphql.ConstraintTypePre.String(),
		TargetOperation: graphql.TargetOperationAssignFormation.String(),
		Operator:        operatorName,
		ResourceType:    graphql.ResourceTypeApplication.String(),
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: graphql.ConstraintScopeFormationType.String(),
	}
	formationConstraintGql2 = &graphql.FormationConstraint{
		ID:              constraintID2,
		Name:            formationConstraintName,
		ConstraintType:  graphql.ConstraintTypePre.String(),
		TargetOperation: graphql.TargetOperationAssignFormation.String(),
		Operator:        operatorName,
		ResourceType:    graphql.ResourceTypeApplication.String(),
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: graphql.ConstraintScopeFormationType.String(),
	}
)

func newModelBusinessTenantMappingWithType(tenantType tenant.Type) *model.BusinessTenantMapping {
	return &model.BusinessTenantMapping{
		ID:             testTenantID,
		Name:           "name",
		ExternalTenant: "external",
		Parent:         testParentTenantID,
		Type:           tenantType,
		Provider:       "test-provider",
		Status:         tenant.Active,
	}
}

func fixModelWebhookInput() []*model.WebhookInput {
	return []*model.WebhookInput{
		{
			Type:           model.WebhookTypeFormationLifecycle,
			URL:            &url,
			Auth:           &model.AuthInput{},
			Mode:           &modelWebhookMode,
			URLTemplate:    &emptyTemplate,
			InputTemplate:  &emptyTemplate,
			HeaderTemplate: &emptyTemplate,
			OutputTemplate: &emptyTemplate,
		},
	}
}

func fixGQLWebhookInput() []*graphql.WebhookInput {
	return []*graphql.WebhookInput{
		{
			Type:           graphql.WebhookTypeFormationLifecycle,
			URL:            &url,
			Auth:           &graphql.AuthInput{},
			Mode:           &graphqlWebhookMode,
			URLTemplate:    &emptyTemplate,
			InputTemplate:  &emptyTemplate,
			HeaderTemplate: &emptyTemplate,
			OutputTemplate: &emptyTemplate,
		},
	}
}

func fixFormationTemplateModelWebhook() *model.Webhook {
	return &model.Webhook{
		ID:             testWebhookID,
		ObjectID:       testID,
		ObjectType:     model.FormationTemplateWebhookReference,
		Type:           model.WebhookTypeFormationLifecycle,
		URL:            &url,
		Auth:           &model.Auth{},
		Mode:           &modelWebhookMode,
		URLTemplate:    &emptyTemplate,
		InputTemplate:  &emptyTemplate,
		HeaderTemplate: &emptyTemplate,
		OutputTemplate: &emptyTemplate,
	}
}

func fixFormationTemplateGQLWebhook() *graphql.Webhook {
	return &graphql.Webhook{
		ID:                  testWebhookID,
		FormationTemplateID: str.Ptr(testID),
		Type:                graphql.WebhookTypeFormationLifecycle,
		URL:                 &url,
		Auth:                &graphql.Auth{},
		Mode:                &graphqlWebhookMode,
		URLTemplate:         &emptyTemplate,
		InputTemplate:       &emptyTemplate,
		HeaderTemplate:      &emptyTemplate,
		OutputTemplate:      &emptyTemplate,
		CreatedAt:           &graphql.Timestamp{},
	}
}

func fixColumns() []string {
	return []string{"id", "name", "application_types", "runtime_types", "runtime_type_display_name", "runtime_artifact_kind", "leading_product_ids", "supports_reset", "discovery_consumers", "tenant_id"}
}

func UnusedFormationTemplateService() *automock.FormationTemplateService {
	return &automock.FormationTemplateService{}
}

func UnusedFormationTemplateRepository() *automock.FormationTemplateRepository {
	return &automock.FormationTemplateRepository{}
}

func UnusedFormationTemplateConverter() *automock.FormationTemplateConverter {
	return &automock.FormationTemplateConverter{}
}

func UnusedTenantService() *automock.TenantService {
	return &automock.TenantService{}
}

func UnusedWebhookService() *automock.WebhookService {
	return &automock.WebhookService{}
}

func UnusedFormationConstraintService() *automock.FormationConstraintService {
	return &automock.FormationConstraintService{}
}

func UnusedFormationConstraintConverter() *automock.FormationConstraintConverter {
	return &automock.FormationConstraintConverter{}
}
