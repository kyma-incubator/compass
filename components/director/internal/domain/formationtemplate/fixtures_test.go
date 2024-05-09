package formationtemplate_test

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

const (
	testFormationTemplateID    = "d1fddec6-5456-4a1e-9ae0-74447f5d6ae9"
	formationTemplateName      = "formation-template-name"
	applicationTypesAsString   = "[\"some-application-type\"]"
	runtimeTypesAsString       = "[\"some-runtime-type\"]"
	testTenantID               = "d9fddec6-5456-4a1e-9ae0-74447f5d6ae9"
	testWebhookID              = "test-wh-id"
	leadingProductIDsAsString  = "[\"leading-product-id\",\"leading-product-id-2\"]"
	discoveryConsumersAsString = "[\"some-application-type\",\"some-runtime-type\"]"
)

var (
	ctx                          = context.Background()
	testErr                      = errors.New("test error")
	formationTemplateNotFoundErr = apperrors.NewNotFoundError(resource.FormationTemplate, testFormationTemplateID)
	nilModelEntity               *model.FormationTemplate
	nilLabelFilters              []*labelfilter.LabelFilter
	emptyLabelFilters            = []*labelfilter.LabelFilter{}

	runtimeTypeDisplayName = "display-name-for-runtime"
	artifactKindAsString   = "SUBSCRIPTION"
	runtimeArtifactKind    = model.RuntimeArtifactKindSubscription
	artifactKind           = graphql.ArtifactTypeSubscription
	emptyTemplate          = `{}`
	url                    = "http://foo.com"
	modelWebhookMode       = model.WebhookModeSync
	graphqlWebhookMode     = graphql.WebhookModeSync
	applicationTypes       = []string{"some-application-type"}
	runtimeTypes           = []string{"some-runtime-type"}
	leadingProductIDs      = []string{"leading-product-id", "leading-product-id-2"}
	discoveryConsumers     = []string{"some-application-type", "some-runtime-type"}
	shouldReset            = true
	testTime               = time.Date(2024, 05, 07, 9, 9, 9, 9, time.Local)
	testLabelKey           = "testLblKey"
	testLabelValue         = "testLblValue"

	registerInputLabels = map[string]interface{}{testLabelKey: testLabelValue}
	lblInput            = fixGQLLabelInput(testLabelKey, testLabelValue)

	gqlLabel  = fixGQLLabel(testLabelKey, testLabelValue)
	gqlLabels = graphql.Labels{testLabelKey: testLabelValue}

	modelLabel  = fixModelLabel(testLabelKey, testLabelValue)
	modelLabels = map[string]*model.Label{testLabelKey: modelLabel}

	testLabelFilter = []*labelfilter.LabelFilter{
		{
			Key:   testLabelKey,
			Query: nil,
		},
	}

	formationTemplateLabelInput = fixModelLabelInput(testLabelKey, testLabelValue, testFormationTemplateID, model.FormationTemplateLabelableObject)

	formationTemplateRegisterInputModel = model.FormationTemplateRegisterInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: &runtimeTypeDisplayName,
		RuntimeArtifactKind:    &runtimeArtifactKind,
		LeadingProductIDs:      leadingProductIDs,
		DiscoveryConsumers:     discoveryConsumers,
		Webhooks:               fixModelWebhookInput(),
	}

	formationTemplateRegisterInputModelWithLabels = model.FormationTemplateRegisterInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: &runtimeTypeDisplayName,
		RuntimeArtifactKind:    &runtimeArtifactKind,
		LeadingProductIDs:      leadingProductIDs,
		DiscoveryConsumers:     discoveryConsumers,
		Webhooks:               fixModelWebhookInput(),
		Labels:                 registerInputLabels,
	}

	formationTemplateUpdateInputModel = model.FormationTemplateUpdateInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: &runtimeTypeDisplayName,
		RuntimeArtifactKind:    &runtimeArtifactKind,
		LeadingProductIDs:      leadingProductIDs,
		DiscoveryConsumers:     discoveryConsumers,
	}

	formationTemplateModelWithResetInput = model.FormationTemplateRegisterInput{
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

	formationTemplateGraphQLInput = graphql.FormationTemplateRegisterInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: str.Ptr(runtimeTypeDisplayName),
		RuntimeArtifactKind:    &artifactKind,
		LeadingProductIDs:      leadingProductIDs,
		DiscoveryConsumers:     discoveryConsumers,
		Webhooks:               fixGQLWebhookInput(),
	}

	formationTemplateUpdateInputGraphQL = graphql.FormationTemplateUpdateInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: str.Ptr(runtimeTypeDisplayName),
		RuntimeArtifactKind:    &artifactKind,
		LeadingProductIDs:      leadingProductIDs,
		DiscoveryConsumers:     discoveryConsumers,
	}

	formationTemplateWithResetGraphQLInput = graphql.FormationTemplateRegisterInput{
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

	formationTemplateModelInputAppOnly = model.FormationTemplateRegisterInput{
		Name:               formationTemplateName,
		ApplicationTypes:   applicationTypes,
		LeadingProductIDs:  leadingProductIDs,
		DiscoveryConsumers: discoveryConsumers,
		Webhooks:           fixModelWebhookInput(),
	}

	formationTemplateGraphQLInputAppOnly = graphql.FormationTemplateRegisterInput{
		Name:               formationTemplateName,
		ApplicationTypes:   applicationTypes,
		LeadingProductIDs:  leadingProductIDs,
		DiscoveryConsumers: discoveryConsumers,
		Webhooks:           fixGQLWebhookInput(),
	}

	formationTemplateModelAppOnly = model.FormationTemplate{
		ID:                 testFormationTemplateID,
		Name:               formationTemplateName,
		ApplicationTypes:   applicationTypes,
		LeadingProductIDs:  leadingProductIDs,
		DiscoveryConsumers: discoveryConsumers,
		TenantID:           str.Ptr(testTenantID),
		Webhooks:           []*model.Webhook{fixFormationTemplateModelWebhook()},
	}

	formationTemplateModel = model.FormationTemplate{
		ID:                     testFormationTemplateID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: &runtimeTypeDisplayName,
		RuntimeArtifactKind:    &runtimeArtifactKind,
		LeadingProductIDs:      leadingProductIDs,
		DiscoveryConsumers:     discoveryConsumers,
		TenantID:               str.Ptr(testTenantID),
		Webhooks:               []*model.Webhook{fixFormationTemplateModelWebhook()},
		CreatedAt:              testTime,
		UpdatedAt:              &testTime,
	}

	formationTemplateModelNullTenant = model.FormationTemplate{
		ID:                     testFormationTemplateID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: &runtimeTypeDisplayName,
		RuntimeArtifactKind:    &runtimeArtifactKind,
		LeadingProductIDs:      leadingProductIDs,
		DiscoveryConsumers:     discoveryConsumers,
		TenantID:               nil,
		CreatedAt:              testTime,
		UpdatedAt:              &testTime,
	}

	formationTemplateEntity = formationtemplate.Entity{
		ID:                     testFormationTemplateID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypesAsString,
		RuntimeTypes:           repo.NewValidNullableString(runtimeTypesAsString),
		RuntimeTypeDisplayName: repo.NewValidNullableString(runtimeTypeDisplayName),
		RuntimeArtifactKind:    repo.NewValidNullableString(artifactKindAsString),
		LeadingProductIDs:      repo.NewNullableStringFromJSONRawMessage(json.RawMessage(leadingProductIDsAsString)),
		TenantID:               repo.NewValidNullableString(testTenantID),
		DiscoveryConsumers:     repo.NewNullableStringFromJSONRawMessage(json.RawMessage(discoveryConsumersAsString)),
		CreatedAt:              testTime,
		UpdatedAt:              &testTime,
	}

	formationTemplateEntityNullTenant = formationtemplate.Entity{
		ID:                     testFormationTemplateID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypesAsString,
		RuntimeTypes:           repo.NewValidNullableString(runtimeTypesAsString),
		RuntimeTypeDisplayName: repo.NewValidNullableString(runtimeTypeDisplayName),
		RuntimeArtifactKind:    repo.NewValidNullableString(artifactKindAsString),
		LeadingProductIDs:      repo.NewNullableStringFromJSONRawMessage(json.RawMessage(leadingProductIDsAsString)),
		TenantID:               repo.NewValidNullableString(""),
		DiscoveryConsumers:     repo.NewNullableStringFromJSONRawMessage(json.RawMessage(discoveryConsumersAsString)),
		CreatedAt:              testTime,
		UpdatedAt:              &testTime,
	}

	graphQLFormationTemplate = graphql.FormationTemplate{
		ID:                     testFormationTemplateID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: &runtimeTypeDisplayName,
		RuntimeArtifactKind:    &artifactKind,
		LeadingProductIDs:      leadingProductIDs,
		DiscoveryConsumers:     discoveryConsumers,
		Webhooks:               []*graphql.Webhook{fixFormationTemplateGQLWebhook()},
		CreatedAt:              graphql.Timestamp(testTime),
		UpdatedAt:              graphql.TimePtrToGraphqlTimestampPtr(&testTime),
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

func fixModelLabelInput(key string, value interface{}, objectID string, objectType model.LabelableObject) *model.LabelInput {
	return &model.LabelInput{
		Key:        key,
		Value:      value,
		ObjectID:   objectID,
		ObjectType: objectType,
	}
}

func fixModelLabel(key string, value interface{}) *model.Label {
	return &model.Label{
		Key:   key,
		Value: value,
	}
}

func fixGQLLabel(key string, value interface{}) *graphql.Label {
	return &graphql.Label{
		Key:   key,
		Value: value,
	}
}

func fixGQLLabelInput(key string, value interface{}) graphql.LabelInput {
	return graphql.LabelInput{
		Key:   key,
		Value: value,
	}
}

func fixFormationTemplateEntity(createdAt time.Time, updatedAt *time.Time) *formationtemplate.Entity {
	return &formationtemplate.Entity{
		ID:                     testFormationTemplateID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypesAsString,
		RuntimeTypes:           repo.NewValidNullableString(runtimeTypesAsString),
		RuntimeTypeDisplayName: repo.NewValidNullableString(runtimeTypeDisplayName),
		RuntimeArtifactKind:    repo.NewValidNullableString(artifactKindAsString),
		LeadingProductIDs:      repo.NewNullableStringFromJSONRawMessage(json.RawMessage(leadingProductIDsAsString)),
		TenantID:               repo.NewValidNullableString(testTenantID),
		DiscoveryConsumers:     repo.NewNullableStringFromJSONRawMessage(json.RawMessage(discoveryConsumersAsString)),
		CreatedAt:              createdAt,
		UpdatedAt:              updatedAt,
	}
}

func fixFormationTemplateModel(createdAt time.Time, updatedAt *time.Time) *model.FormationTemplate {
	return &model.FormationTemplate{
		ID:                     testFormationTemplateID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: &runtimeTypeDisplayName,
		RuntimeArtifactKind:    &runtimeArtifactKind,
		LeadingProductIDs:      leadingProductIDs,
		DiscoveryConsumers:     discoveryConsumers,
		TenantID:               str.Ptr(testTenantID),
		Webhooks:               []*model.Webhook{fixFormationTemplateModelWebhook()},
		CreatedAt:              createdAt,
		UpdatedAt:              updatedAt,
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
		ObjectID:       testFormationTemplateID,
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
		FormationTemplateID: str.Ptr(testFormationTemplateID),
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
	return []string{"id", "name", "application_types", "runtime_types", "runtime_type_display_name", "runtime_artifact_kind", "leading_product_ids", "supports_reset", "discovery_consumers", "created_at", "updated_at", "tenant_id"}
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

func UnusedLabelService() *automock.LabelService {
	return &automock.LabelService{}
}

func UnusedWebhookService() *automock.WebhookService {
	return &automock.WebhookService{}
}

func UnusedWebhookRepo() *automock.WebhookRepository {
	return &automock.WebhookRepository{}
}

func UnusedFormationConstraintService() *automock.FormationConstraintService {
	return &automock.FormationConstraintService{}
}

func UnusedFormationConstraintConverter() *automock.FormationConstraintConverter {
	return &automock.FormationConstraintConverter{}
}
