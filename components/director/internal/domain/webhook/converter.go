package webhook

import (
	"database/sql"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

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

	var webhookMode *graphql.WebhookMode
	if in.Mode != nil {
		mode := graphql.WebhookMode(*in.Mode)
		webhookMode = &mode
	}

	return &graphql.Webhook{
		ID:                  in.ID,
		ApplicationID:       in.ApplicationID,
		RuntimeID:           in.RuntimeID,
		IntegrationSystemID: in.IntegrationSystemID,
		Type:                graphql.WebhookType(in.Type),
		Mode:                webhookMode,
		URL:                 in.URL,
		Auth:                auth,
		CorrelationIDKey:    in.CorrelationIDKey,
		RetryInterval:       in.RetryInterval,
		Timeout:             in.Timeout,
		URLTemplate:         in.URLTemplate,
		InputTemplate:       in.InputTemplate,
		HeaderTemplate:      in.HeaderTemplate,
		OutputTemplate:      in.OutputTemplate,
		StatusTemplate:      in.StatusTemplate,
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

	var webhookMode *model.WebhookMode
	if in.Mode != nil {
		mode := model.WebhookMode(*in.Mode)
		webhookMode = &mode
	}

	return &model.WebhookInput{
		Type:             model.WebhookType(in.Type),
		URL:              in.URL,
		Auth:             auth,
		Mode:             webhookMode,
		CorrelationIDKey: in.CorrelationIDKey,
		RetryInterval:    in.RetryInterval,
		Timeout:          in.Timeout,
		URLTemplate:      in.URLTemplate,
		InputTemplate:    in.InputTemplate,
		HeaderTemplate:   in.HeaderTemplate,
		OutputTemplate:   in.OutputTemplate,
		StatusTemplate:   in.StatusTemplate,
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

	var webhookMode sql.NullString
	if in.Mode != nil {
		webhookMode.String = string(*in.Mode)
		webhookMode.Valid = true
	}

	return Entity{
		ID:                  in.ID,
		TenantID:            in.TenantID,
		ApplicationID:       repo.NewNullableString(in.ApplicationID),
		RuntimeID:           repo.NewNullableString(in.RuntimeID),
		IntegrationSystemID: repo.NewNullableString(in.IntegrationSystemID),
		CollectionIDKey:     repo.NewNullableString(in.CorrelationIDKey),
		Type:                string(in.Type),
		URL:                 repo.NewNullableString(in.URL),
		Auth:                optionalAuth,
		Mode:                webhookMode,
		RetryInterval:       repo.NewNullableInt(in.RetryInterval),
		Timeout:             repo.NewNullableInt(in.Timeout),
		URLTemplate:         repo.NewNullableString(in.URLTemplate),
		InputTemplate:       repo.NewNullableString(in.InputTemplate),
		HeaderTemplate:      repo.NewNullableString(in.HeaderTemplate),
		OutputTemplate:      repo.NewNullableString(in.OutputTemplate),
		StatusTemplate:      repo.NewNullableString(in.StatusTemplate),
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

	var webhookMode *model.WebhookMode
	if in.Mode.Valid {
		webhookModeStr := model.WebhookMode(in.Mode.String)
		webhookMode = &webhookModeStr
	}

	return model.Webhook{
		ID:                  in.ID,
		TenantID:            in.TenantID,
		ApplicationID:       repo.StringPtrFromNullableString(in.ApplicationID),
		RuntimeID:           repo.StringPtrFromNullableString(in.RuntimeID),
		IntegrationSystemID: repo.StringPtrFromNullableString(in.IntegrationSystemID),
		CorrelationIDKey:    repo.StringPtrFromNullableString(in.CollectionIDKey),
		Type:                model.WebhookType(in.Type),
		URL:                 repo.StringPtrFromNullableString(in.URL),
		Auth:                auth,
		Mode:                webhookMode,
		RetryInterval:       repo.IntPtrFromNullableInt(in.RetryInterval),
		Timeout:             repo.IntPtrFromNullableInt(in.Timeout),
		URLTemplate:         repo.StringPtrFromNullableString(in.URLTemplate),
		InputTemplate:       repo.StringPtrFromNullableString(in.InputTemplate),
		HeaderTemplate:      repo.StringPtrFromNullableString(in.HeaderTemplate),
		OutputTemplate:      repo.StringPtrFromNullableString(in.OutputTemplate),
		StatusTemplate:      repo.StringPtrFromNullableString(in.StatusTemplate),
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

func nullableInt(n *int) int {
	if n != nil {
		return *n
	}
	return 0
}
