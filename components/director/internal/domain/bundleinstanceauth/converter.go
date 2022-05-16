package bundleinstanceauth

import (
	"database/sql"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

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
	return &converter{
		authConverter: authConverter,
	}
}

// ToGraphQL missing godoc
func (c *converter) ToGraphQL(in *model.BundleInstanceAuth) (*graphql.BundleInstanceAuth, error) {
	if in == nil {
		return nil, nil
	}

	auth, err := c.authConverter.ToGraphQL(in.Auth)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Auth to GraphQL")
	}

	return &graphql.BundleInstanceAuth{
		ID:               in.ID,
		Context:          c.strPtrToJSONPtr(in.Context),
		InputParams:      c.strPtrToJSONPtr(in.InputParams),
		Auth:             auth,
		Status:           c.statusToGraphQL(in.Status),
		RuntimeID:        in.RuntimeID,
		RuntimeContextID: in.RuntimeContextID,
	}, nil
}

// MultipleToGraphQL missing godoc
func (c *converter) MultipleToGraphQL(in []*model.BundleInstanceAuth) ([]*graphql.BundleInstanceAuth, error) {
	bundleInstanceAuths := make([]*graphql.BundleInstanceAuth, 0, len(in))
	for _, r := range in {
		if r == nil {
			continue
		}
		pia, err := c.ToGraphQL(r)
		if err != nil {
			return nil, err
		}
		bundleInstanceAuths = append(bundleInstanceAuths, pia)
	}

	return bundleInstanceAuths, nil
}

// RequestInputFromGraphQL missing godoc
func (c *converter) RequestInputFromGraphQL(in graphql.BundleInstanceAuthRequestInput) model.BundleInstanceAuthRequestInput {
	return model.BundleInstanceAuthRequestInput{
		ID:          in.ID,
		Context:     c.jsonPtrToStrPtr(in.Context),
		InputParams: c.jsonPtrToStrPtr(in.InputParams),
	}
}

// SetInputFromGraphQL missing godoc
func (c *converter) SetInputFromGraphQL(in graphql.BundleInstanceAuthSetInput) (model.BundleInstanceAuthSetInput, error) {
	auth, err := c.authConverter.InputFromGraphQL(in.Auth)
	if err != nil {
		return model.BundleInstanceAuthSetInput{}, errors.Wrap(err, "while converting Auth")
	}

	out := model.BundleInstanceAuthSetInput{
		Auth: auth,
	}

	if in.Status != nil {
		out.Status = &model.BundleInstanceAuthStatusInput{
			Condition: model.BundleInstanceAuthSetStatusConditionInput(in.Status.Condition),
			Message:   in.Status.Message,
			Reason:    in.Status.Reason,
		}
	}

	return out, nil
}

// ToEntity missing godoc
func (c *converter) ToEntity(in *model.BundleInstanceAuth) (*Entity, error) {
	out := &Entity{
		ID:               in.ID,
		BundleID:         in.BundleID,
		OwnerID:          in.Owner,
		RuntimeID:        repo.NewNullableString(in.RuntimeID),
		RuntimeContextID: repo.NewNullableString(in.RuntimeContextID),
		Context:          repo.NewNullableString(in.Context),
		InputParams:      repo.NewNullableString(in.InputParams),
	}
	authValue, err := c.nullStringFromAuthPtr(in.Auth)
	if err != nil {
		return nil, err
	}
	out.AuthValue = authValue

	if in.Status != nil {
		out.StatusCondition = string(in.Status.Condition)
		out.StatusTimestamp = in.Status.Timestamp
		out.StatusMessage = in.Status.Message
		out.StatusReason = in.Status.Reason
	}

	return out, nil
}

// FromEntity missing godoc
func (c *converter) FromEntity(in *Entity) (*model.BundleInstanceAuth, error) {
	auth, err := c.authPtrFromNullString(in.AuthValue)
	if err != nil {
		return nil, err
	}

	return &model.BundleInstanceAuth{
		ID:               in.ID,
		BundleID:         in.BundleID,
		Owner:            in.OwnerID,
		RuntimeID:        repo.StringPtrFromNullableString(in.RuntimeID),
		RuntimeContextID: repo.StringPtrFromNullableString(in.RuntimeContextID),
		Context:          repo.StringPtrFromNullableString(in.Context),
		InputParams:      repo.StringPtrFromNullableString(in.InputParams),
		Auth:             auth,
		Status: &model.BundleInstanceAuthStatus{
			Condition: model.BundleInstanceAuthStatusCondition(in.StatusCondition),
			Timestamp: in.StatusTimestamp,
			Message:   in.StatusMessage,
			Reason:    in.StatusReason,
		},
	}, nil
}

func (c *converter) statusToGraphQL(in *model.BundleInstanceAuthStatus) *graphql.BundleInstanceAuthStatus {
	if in == nil {
		return nil
	}

	return &graphql.BundleInstanceAuthStatus{
		Condition: graphql.BundleInstanceAuthStatusCondition(in.Condition),
		Timestamp: graphql.Timestamp(in.Timestamp),
		Message:   in.Message,
		Reason:    in.Reason,
	}
}

func (c *converter) strPtrToJSONPtr(in *string) *graphql.JSON {
	if in == nil {
		return nil
	}
	out := graphql.JSON(*in)
	return &out
}

func (c *converter) jsonPtrToStrPtr(in *graphql.JSON) *string {
	if in == nil {
		return nil
	}
	out := string(*in)
	return &out
}

func (c *converter) nullStringFromAuthPtr(in *model.Auth) (sql.NullString, error) {
	if in == nil {
		return sql.NullString{}, nil
	}
	valueMarshalled, err := json.Marshal(*in)
	if err != nil {
		return sql.NullString{}, errors.Wrap(err, "while marshalling Auth")
	}
	return sql.NullString{
		String: string(valueMarshalled),
		Valid:  true,
	}, nil
}

func (c *converter) authPtrFromNullString(in sql.NullString) (*model.Auth, error) {
	if !in.Valid {
		return nil, nil
	}
	var auth model.Auth
	err := json.Unmarshal([]byte(in.String), &auth)
	if err != nil {
		return nil, err
	}
	return &auth, nil
}
