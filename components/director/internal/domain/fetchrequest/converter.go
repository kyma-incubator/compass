package fetchrequest

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

func (c *converter) ToGraphQL(in *model.FetchRequest) *graphql.FetchRequest {
	if in == nil {
		return nil
	}

	return &graphql.FetchRequest{
		URL:    in.URL,
		Auth:   c.authConverter.ToGraphQL(in.Auth),
		Mode:   graphql.FetchMode(in.Mode),
		Filter: in.Filter,
		Status: c.statusToGraphQL(in.Status),
	}
}

func (c *converter) InputFromGraphQL(in *graphql.FetchRequestInput) *model.FetchRequestInput {
	if in == nil {
		return nil
	}

	var mode *model.FetchMode
	if in.Mode != nil {
		tmp := model.FetchMode(*in.Mode)
		mode = &tmp
	}

	return &model.FetchRequestInput{
		URL:    in.URL,
		Auth:   c.authConverter.InputFromGraphQL(in.Auth),
		Mode:   mode,
		Filter: in.Filter,
	}
}

func (c *converter) statusToGraphQL(in *model.FetchRequestStatus) *graphql.FetchRequestStatus {
	if in == nil {
		return &graphql.FetchRequestStatus{
			Condition: graphql.FetchRequestStatusConditionInitial,
		}
	}

	var condition graphql.FetchRequestStatusCondition

	switch in.Condition {
	case model.FetchRequestStatusConditionInitial:
		condition = graphql.FetchRequestStatusConditionInitial
	case model.FetchRequestStatusConditionFailed:
		condition = graphql.FetchRequestStatusConditionFailed
	case model.FetchRequestStatusConditionSucceeded:
		condition = graphql.FetchRequestStatusConditionSucceeded
	default:
		condition = graphql.FetchRequestStatusConditionInitial
	}

	return &graphql.FetchRequestStatus{
		Condition: condition,
		Timestamp: graphql.Timestamp(in.Timestamp),
	}
}
