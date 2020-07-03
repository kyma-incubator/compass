package webhook

import (
	"database/sql"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=AuthConverter -output=automock -outpkg=automock -case=underscore
type AuthConverter interface {
	ToGraphQL(in *model.Auth) (*graphql.Auth, error)
	InputFromGraphQL(in *graphql.AuthInput) (*model.AuthInput, error)
}

type converter struct {
	authConverter AuthConverter
}

func NewConverter(authConverter AuthConverter) *converter {
	return &converter{authConverter: authConverter}
}

func (c *converter) ToGraphQL(in *model.Webhook) (*graphql.Webhook, error) {
	if in == nil {
		return nil, nil
	}

	auth, err := c.authConverter.ToGraphQL(in.Auth)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Auth input")
	}

	return &graphql.Webhook{
		ID:            in.ID,
		ApplicationID: in.ApplicationID,
		Type:          graphql.ApplicationWebhookType(in.Type),
		URL:           in.URL,
		Auth:          auth,
	}, nil
}

func (c *converter) MultipleToGraphQL(in []*model.Webhook) ([]*graphql.Webhook, error) {
	var webhooks []*graphql.Webhook
	for _, r := range in {
		if r == nil {
			continue
		}

		webhook, err := c.ToGraphQL(r)
		if err != nil {
			return nil, err
		}

		webhooks = append(webhooks, webhook)
	}

	return webhooks, nil
}

func (c *converter) InputFromGraphQL(in *graphql.WebhookInput) (*model.WebhookInput, error) {
	if in == nil {
		return nil, nil
	}

	auth, err := c.authConverter.InputFromGraphQL(in.Auth)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Auth input")
	}

	return &model.WebhookInput{
		Type: model.WebhookType(in.Type),
		URL:  in.URL,
		Auth: auth,
	}, nil
}

func (c *converter) MultipleInputFromGraphQL(in []*graphql.WebhookInput) ([]*model.WebhookInput, error) {
	var inputs []*model.WebhookInput
	for _, r := range in {
		if r == nil {
			continue
		}
		webhookIn, err := c.InputFromGraphQL(r)
		if err != nil {
			return nil, err
		}

		inputs = append(inputs, webhookIn)
	}

	return inputs, nil
}

func (c *converter) ToEntity(in model.Webhook) (Entity, error) {
	optionalAuth, err := c.toAuthEntity(in)
	if err != nil {
		return Entity{}, err
	}

	return Entity{
		ID:       in.ID,
		Type:     string(in.Type),
		TenantID: in.Tenant,
		URL:      in.URL,
		AppID:    in.ApplicationID,
		Auth:     optionalAuth,
	}, nil
}

func (c *converter) toAuthEntity(in model.Webhook) (sql.NullString, error) {
	var optionalAuth sql.NullString
	if in.Auth == nil {
		return optionalAuth, nil
	}

	b, err := json.Marshal(in.Auth)
	if err != nil {
		return sql.NullString{}, errors.Wrap(err, "while marshalling Auth")
	}

	if err := optionalAuth.Scan(b); err != nil {
		return sql.NullString{}, errors.Wrap(err, "while scanning optional Auth")
	}
	return optionalAuth, nil
}

func (c *converter) FromEntity(in Entity) (model.Webhook, error) {
	auth, err := c.fromEntityAuth(in)
	if err != nil {
		return model.Webhook{}, err
	}
	return model.Webhook{
		ID:            in.ID,
		Type:          model.WebhookType(in.Type),
		Tenant:        in.TenantID,
		URL:           in.URL,
		ApplicationID: in.AppID,
		Auth:          auth,
	}, nil
}

func (c *converter) fromEntityAuth(in Entity) (*model.Auth, error) {
	if !in.Auth.Valid {
		return nil, nil
	}

	auth := &model.Auth{}
	val, err := in.Auth.Value()
	if err != nil {
		return nil, errors.Wrap(err, "while reading Auth from Entity")
	}

	b, ok := val.(string)
	if !ok {
		return nil, apperrors.NewInternalError("Auth should be slice of bytes")
	}
	if err := json.Unmarshal([]byte(b), auth); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling Auth")
	}

	return auth, nil
}
