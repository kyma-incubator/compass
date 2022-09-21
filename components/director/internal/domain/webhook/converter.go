package webhook

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

// AuthConverter missing godoc
//go:generate mockery --name=AuthConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type AuthConverter interface {
	ToGraphQL(in *model.Auth) (*graphql.Auth, error)
	InputFromGraphQL(in *graphql.AuthInput) (*model.AuthInput, error)
}

type converter struct {
	authConverter AuthConverter
}

// NewConverter missing godoc
func NewConverter(authConverter AuthConverter) *converter {
	return &converter{authConverter: authConverter}
}

// ToGraphQL missing godoc
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

	var appID *string
	var runtimeID *string
	var appTemplateID *string
	var intSystemID *string
	switch in.ObjectType {
	case model.ApplicationWebhookReference:
		appID = &in.ObjectID
	case model.RuntimeWebhookReference:
		runtimeID = &in.ObjectID
	case model.ApplicationTemplateWebhookReference:
		appTemplateID = &in.ObjectID
	case model.IntegrationSystemWebhookReference:
		intSystemID = &in.ObjectID
	}

	return &graphql.Webhook{
		ID:                    in.ID,
		ApplicationID:         appID,
		ApplicationTemplateID: appTemplateID,
		RuntimeID:             runtimeID,
		IntegrationSystemID:   intSystemID,
		Type:                  graphql.WebhookType(in.Type),
		Mode:                  webhookMode,
		URL:                   in.URL,
		Auth:                  auth,
		CorrelationIDKey:      in.CorrelationIDKey,
		RetryInterval:         in.RetryInterval,
		Timeout:               in.Timeout,
		URLTemplate:           in.URLTemplate,
		InputTemplate:         in.InputTemplate,
		HeaderTemplate:        in.HeaderTemplate,
		OutputTemplate:        in.OutputTemplate,
		StatusTemplate:        in.StatusTemplate,
		CreatedAt:             timePtrToTimestampPtr(in.CreatedAt),
	}, nil
}

// MultipleToGraphQL missing godoc
func (c *converter) MultipleToGraphQL(in []*model.Webhook) ([]*graphql.Webhook, error) {
	webhooks := make([]*graphql.Webhook, 0, len(in))
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

// InputFromGraphQL missing godoc
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

// MultipleInputFromGraphQL missing godoc
func (c *converter) MultipleInputFromGraphQL(in []*graphql.WebhookInput) ([]*model.WebhookInput, error) {
	inputs := make([]*model.WebhookInput, 0, len(in))
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

// ToEntity missing godoc
func (c *converter) ToEntity(in *model.Webhook) (*Entity, error) {
	optionalAuth, err := c.toAuthEntity(*in)
	if err != nil {
		return nil, err
	}

	var webhookMode sql.NullString
	if in.Mode != nil {
		webhookMode.String = string(*in.Mode)
		webhookMode.Valid = true
	}

	var appID sql.NullString
	var runtimeID sql.NullString
	var appTemplateID sql.NullString
	var intSystemID sql.NullString
	switch in.ObjectType {
	case model.ApplicationWebhookReference:
		appID = repo.NewValidNullableString(in.ObjectID)
	case model.RuntimeWebhookReference:
		runtimeID = repo.NewValidNullableString(in.ObjectID)
	case model.ApplicationTemplateWebhookReference:
		appTemplateID = repo.NewValidNullableString(in.ObjectID)
	case model.IntegrationSystemWebhookReference:
		intSystemID = repo.NewValidNullableString(in.ObjectID)
	}

	return &Entity{
		ID:                    in.ID,
		ApplicationID:         appID,
		ApplicationTemplateID: appTemplateID,
		RuntimeID:             runtimeID,
		IntegrationSystemID:   intSystemID,
		CollectionIDKey:       repo.NewNullableString(in.CorrelationIDKey),
		Type:                  string(in.Type),
		URL:                   repo.NewNullableString(in.URL),
		Auth:                  optionalAuth,
		Mode:                  webhookMode,
		RetryInterval:         repo.NewNullableInt(in.RetryInterval),
		Timeout:               repo.NewNullableInt(in.Timeout),
		URLTemplate:           repo.NewNullableString(in.URLTemplate),
		InputTemplate:         repo.NewNullableString(in.InputTemplate),
		HeaderTemplate:        repo.NewNullableString(in.HeaderTemplate),
		OutputTemplate:        repo.NewNullableString(in.OutputTemplate),
		StatusTemplate:        repo.NewNullableString(in.StatusTemplate),
		CreatedAt:             in.CreatedAt,
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

// FromEntity missing godoc
func (c *converter) FromEntity(in *Entity) (*model.Webhook, error) {
	auth, err := c.fromEntityAuth(*in)
	if err != nil {
		return nil, err
	}

	var webhookMode *model.WebhookMode
	if in.Mode.Valid {
		webhookModeStr := model.WebhookMode(in.Mode.String)
		webhookMode = &webhookModeStr
	}

	objID, objType, err := c.objectReferenceFromEntity(*in)
	if err != nil {
		return nil, err
	}

	return &model.Webhook{
		ID:               in.ID,
		ObjectID:         objID,
		ObjectType:       objType,
		CorrelationIDKey: repo.StringPtrFromNullableString(in.CollectionIDKey),
		Type:             model.WebhookType(in.Type),
		URL:              repo.StringPtrFromNullableString(in.URL),
		Auth:             auth,
		Mode:             webhookMode,
		RetryInterval:    repo.IntPtrFromNullableInt(in.RetryInterval),
		Timeout:          repo.IntPtrFromNullableInt(in.Timeout),
		URLTemplate:      repo.StringPtrFromNullableString(in.URLTemplate),
		InputTemplate:    repo.StringPtrFromNullableString(in.InputTemplate),
		HeaderTemplate:   repo.StringPtrFromNullableString(in.HeaderTemplate),
		OutputTemplate:   repo.StringPtrFromNullableString(in.OutputTemplate),
		StatusTemplate:   repo.StringPtrFromNullableString(in.StatusTemplate),
		CreatedAt:        in.CreatedAt,
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

func (c *converter) objectReferenceFromEntity(in Entity) (string, model.WebhookReferenceObjectType, error) {
	if in.ApplicationID.Valid {
		return in.ApplicationID.String, model.ApplicationWebhookReference, nil
	}

	if in.RuntimeID.Valid {
		return in.RuntimeID.String, model.RuntimeWebhookReference, nil
	}

	if in.ApplicationTemplateID.Valid {
		return in.ApplicationTemplateID.String, model.ApplicationTemplateWebhookReference, nil
	}

	if in.IntegrationSystemID.Valid {
		return in.IntegrationSystemID.String, model.IntegrationSystemWebhookReference, nil
	}

	return "", "", fmt.Errorf("incorrect Object Reference ID and its type for Entity with ID '%s'", in.ID)
}

func timePtrToTimestampPtr(time *time.Time) *graphql.Timestamp {
	if time == nil {
		return nil
	}

	t := graphql.Timestamp(*time)
	return &t
}
