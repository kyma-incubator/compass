package formationtemplate

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/uid"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

// NewConverter creates a new instance of gqlConverter
func NewConverter(webhook WebhookConverter) *converter {
	return &converter{webhook: webhook}
}

type converter struct {
	webhook WebhookConverter
}

// FromInputGraphQL converts from GraphQL input to internal model input
func (c *converter) FromInputGraphQL(in *graphql.FormationTemplateInput) (*model.FormationTemplateInput, error) {
	if in == nil {
		return nil, nil
	}

	webhooks, err := c.webhook.MultipleInputFromGraphQL(in.Webhooks)
	if err != nil {
		return nil, err
	}

	return &model.FormationTemplateInput{
		Name:                   in.Name,
		ApplicationTypes:       in.ApplicationTypes,
		RuntimeTypes:           in.RuntimeTypes,
		RuntimeTypeDisplayName: in.RuntimeTypeDisplayName,
		RuntimeArtifactKind:    model.RuntimeArtifactKind(in.RuntimeArtifactKind),
		Webhooks:               webhooks,
	}, nil
}

// FromModelInputToModel converts from internal model input and id to internal model
func (c *converter) FromModelInputToModel(in *model.FormationTemplateInput, id, tenantID string) *model.FormationTemplate {
	if in == nil {
		return nil
	}

	uidService := uid.NewService()
	webhooks := make([]*model.Webhook, 0)
	for _, webhookInput := range in.Webhooks {
		webhook := webhookInput.ToWebhook(uidService.Generate(), id, model.FormationTemplateWebhookReference)
		webhooks = append(webhooks, webhook)
	}

	var tntID *string
	if tenantID != "" {
		tntID = &tenantID
	}

	return &model.FormationTemplate{
		ID:                     id,
		Name:                   in.Name,
		ApplicationTypes:       in.ApplicationTypes,
		RuntimeTypes:           in.RuntimeTypes,
		RuntimeTypeDisplayName: in.RuntimeTypeDisplayName,
		RuntimeArtifactKind:    in.RuntimeArtifactKind,
		TenantID:               tntID,
		Webhooks:               webhooks,
	}
}

// ToGraphQL converts from internal model to GraphQL output
func (c *converter) ToGraphQL(in *model.FormationTemplate) (*graphql.FormationTemplate, error) {
	if in == nil {
		return nil, nil
	}

	webhooks, err := c.webhook.MultipleToGraphQL(in.Webhooks)
	if err != nil {
		return nil, err
	}

	return &graphql.FormationTemplate{
		ID:                     in.ID,
		Name:                   in.Name,
		ApplicationTypes:       in.ApplicationTypes,
		RuntimeTypes:           in.RuntimeTypes,
		RuntimeTypeDisplayName: in.RuntimeTypeDisplayName,
		RuntimeArtifactKind:    graphql.ArtifactType(in.RuntimeArtifactKind),
		Webhooks:               webhooks,
	}, nil
}

// MultipleToGraphQL converts multiple internal models to GraphQL models
func (c *converter) MultipleToGraphQL(in []*model.FormationTemplate) ([]*graphql.FormationTemplate, error) {
	if in == nil {
		return nil, nil
	}
	formationTemplates := make([]*graphql.FormationTemplate, 0, len(in))
	for _, r := range in {
		if r == nil {
			continue
		}

		converted, err := c.ToGraphQL(r)
		if err != nil {
			return nil, err
		}

		formationTemplates = append(formationTemplates, converted)
	}

	return formationTemplates, nil
}

// ToEntity converts from internal model to entity
func (c *converter) ToEntity(in *model.FormationTemplate) (*Entity, error) {
	if in == nil {
		return nil, nil
	}
	marshalledApplicationTypes, err := json.Marshal(in.ApplicationTypes)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling application types")
	}
	marshalledRuntimeTypes, err := json.Marshal(in.RuntimeTypes)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling application types")
	}

	return &Entity{
		ID:                     in.ID,
		Name:                   in.Name,
		ApplicationTypes:       string(marshalledApplicationTypes),
		RuntimeTypes:           string(marshalledRuntimeTypes),
		RuntimeTypeDisplayName: in.RuntimeTypeDisplayName,
		RuntimeArtifactKind:    string(in.RuntimeArtifactKind),
		TenantID:               repo.NewNullableString(in.TenantID),
	}, nil
}

// FromEntity converts from entity to internal model
func (c *converter) FromEntity(in *Entity) (*model.FormationTemplate, error) {
	if in == nil {
		return nil, nil
	}

	var unmarshalledApplicationTypes []string
	err := json.Unmarshal([]byte(in.ApplicationTypes), &unmarshalledApplicationTypes)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshalling application types")
	}

	var unmarshalledRuntimeTypes []string
	err = json.Unmarshal([]byte(in.RuntimeTypes), &unmarshalledRuntimeTypes)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshalling runtime types")
	}

	return &model.FormationTemplate{
		ID:                     in.ID,
		Name:                   in.Name,
		ApplicationTypes:       unmarshalledApplicationTypes,
		RuntimeTypes:           unmarshalledRuntimeTypes,
		RuntimeTypeDisplayName: in.RuntimeTypeDisplayName,
		RuntimeArtifactKind:    model.RuntimeArtifactKind(in.RuntimeArtifactKind),
		TenantID:               repo.StringPtrFromNullableString(in.TenantID),
	}, nil
}
