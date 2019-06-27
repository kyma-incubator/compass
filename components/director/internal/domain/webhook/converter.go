package webhook

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=AuthConverter -output=automock -outpkg=automock -case=underscore
type AuthConverter interface {
	ToGraphQL(in *model.Auth) *graphql.Auth
	InputFromGraphQL(in *graphql.AuthInput) *model.AuthInput
}

type converter struct {
	authConverter AuthConverter
}

func NewConverter(authConverter AuthConverter) *converter {
	return &converter{authConverter: authConverter}
}

func (c *converter) ToGraphQL(in *model.ApplicationWebhook) *graphql.ApplicationWebhook {
	if in == nil {
		return nil
	}

	return &graphql.ApplicationWebhook{
		ID:   in.ID,
		Type: graphql.ApplicationWebhookType(in.Type),
		URL:  in.URL,
		Auth: c.authConverter.ToGraphQL(in.Auth),
	}
}

func (c *converter) MultipleToGraphQL(in []*model.ApplicationWebhook) []*graphql.ApplicationWebhook {
	var webhooks []*graphql.ApplicationWebhook
	for _, r := range in {
		if r == nil {
			continue
		}

		webhooks = append(webhooks, c.ToGraphQL(r))
	}

	return webhooks
}

func (c *converter) InputFromGraphQL(in *graphql.ApplicationWebhookInput) *model.ApplicationWebhookInput {
	if in == nil {
		return nil
	}

	return &model.ApplicationWebhookInput{
		Type: model.ApplicationWebhookType(in.Type),
		URL:  in.URL,
		Auth: c.authConverter.InputFromGraphQL(in.Auth),
	}
}

func (c *converter) MultipleInputFromGraphQL(in []*graphql.ApplicationWebhookInput) []*model.ApplicationWebhookInput {
	var inputs []*model.ApplicationWebhookInput
	for _, r := range in {
		if r == nil {
			continue
		}

		inputs = append(inputs, c.InputFromGraphQL(r))
	}

	return inputs
}
