package packageinstanceauth

import (
	"database/sql"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
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
	return &converter{
		authConverter: authConverter,
	}
}

func (c *converter) ToGraphQL(in *model.PackageInstanceAuth) *graphql.PackageInstanceAuth {
	if in == nil {
		return nil
	}

	return &graphql.PackageInstanceAuth{
		ID:          in.ID,
		Context:     c.strPtrToJSONPtr(in.Context),
		InputParams: c.strPtrToJSONPtr(in.InputParams),
		Auth:        c.authConverter.ToGraphQL(in.Auth),
		Status:      c.statusToGraphQL(in.Status),
	}
}

func (c *converter) MultipleToGraphQL(in []*model.PackageInstanceAuth) []*graphql.PackageInstanceAuth {
	var packageInstanceAuths []*graphql.PackageInstanceAuth
	for _, r := range in {
		if r == nil {
			continue
		}

		packageInstanceAuths = append(packageInstanceAuths, c.ToGraphQL(r))
	}

	return packageInstanceAuths
}

func (c *converter) RequestInputFromGraphQL(in graphql.PackageInstanceAuthRequestInput) model.PackageInstanceAuthRequestInput {
	return model.PackageInstanceAuthRequestInput{
		Context:     c.jsonPtrToStrPtr(in.Context),
		InputParams: c.jsonPtrToStrPtr(in.InputParams),
	}
}

func (c *converter) SetInputFromGraphQL(in graphql.PackageInstanceAuthSetInput) model.PackageInstanceAuthSetInput {
	out := model.PackageInstanceAuthSetInput{
		Auth: c.authConverter.InputFromGraphQL(in.Auth),
	}

	if in.Status != nil {
		out.Status = &model.PackageInstanceAuthStatusInput{
			Condition: model.PackageInstanceAuthSetStatusConditionInput(in.Status.Condition),
			Message:   in.Status.Message,
			Reason:    in.Status.Reason,
		}
	}

	return out
}

func (c *converter) ToEntity(in model.PackageInstanceAuth) (Entity, error) {
	out := Entity{
		ID:          in.ID,
		PackageID:   in.PackageID,
		TenantID:    in.Tenant,
		Context:     repo.NewNullableString(in.Context),
		InputParams: repo.NewNullableString(in.InputParams),
	}
	authValue, err := c.nullStringFromAuthPtr(in.Auth)
	if err != nil {
		return Entity{}, err
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

func (c *converter) FromEntity(in Entity) (model.PackageInstanceAuth, error) {
	auth, err := c.authPtrFromNullString(in.AuthValue)
	if err != nil {
		return model.PackageInstanceAuth{}, err
	}

	return model.PackageInstanceAuth{
		ID:          in.ID,
		PackageID:   in.PackageID,
		Tenant:      in.TenantID,
		Context:     repo.StringPtrFromNullableString(in.Context),
		InputParams: repo.StringPtrFromNullableString(in.InputParams),
		Auth:        auth,
		Status: &model.PackageInstanceAuthStatus{
			Condition: model.PackageInstanceAuthStatusCondition(in.StatusCondition),
			Timestamp: in.StatusTimestamp,
			Message:   in.StatusMessage,
			Reason:    in.StatusReason,
		},
	}, nil
}

func (c *converter) statusToGraphQL(in *model.PackageInstanceAuthStatus) *graphql.PackageInstanceAuthStatus {
	if in == nil {
		return nil
	}

	return &graphql.PackageInstanceAuthStatus{
		Condition: graphql.PackageInstanceAuthStatusCondition(in.Condition),
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
