package fetchrequest

import (
	"database/sql"
	"encoding/json"

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

func (c *converter) ToEntity(in model.FetchRequest) (Entity, error) {
	if in.Status == nil {
		return Entity{}, errors.New("Invalid input model")
	}

	var authMarshalled []byte
	var err error

	var auth sql.NullString
	if in.Auth != nil {
		authMarshalled, err = json.Marshal(in.Auth)
		if err != nil {
			return Entity{}, errors.Wrap(err, "while marshalling Auth")
		}
		auth = sql.NullString{
			String: string(authMarshalled),
			Valid:  true,
		}
	}

	refID := sql.NullString{
		Valid:  true,
		String: in.ObjectID,
	}

	var filter sql.NullString
	if in.Filter != nil {
		filter.String = *in.Filter
		filter.Valid = true
	}

	var apiDefID sql.NullString
	var eventAPIDefID sql.NullString
	var documentID sql.NullString
	switch in.ObjectType {
	case model.EventAPIFetchRequestReference:
		eventAPIDefID = refID
	case model.APIFetchRequestReference:
		apiDefID = refID
	case model.DocumentFetchRequestReference:
		documentID = refID
	}

	return Entity{
		ID:              in.ID,
		TenantID:        in.Tenant,
		URL:             in.URL,
		Auth:            auth,
		APIDefID:        apiDefID,
		EventAPIDefID:   eventAPIDefID,
		DocumentID:      documentID,
		Mode:            string(in.Mode),
		Filter:          filter,
		StatusCondition: string(in.Status.Condition),
		StatusTimestamp: in.Status.Timestamp,
	}, nil
}

func (c *converter) FromEntity(in Entity) (model.FetchRequest, error) {
	var authPtr *model.Auth
	if in.Auth.Valid {
		var auth model.Auth
		err := json.Unmarshal([]byte(in.Auth.String), &auth)
		if err != nil {
			return model.FetchRequest{}, errors.Wrap(err, "while unmarshalling Auth")
		}

		authPtr = &auth
	}

	var objectType model.FetchRequestReferenceObjectType
	var objectID string

	if in.APIDefID.Valid {
		objectType = model.APIFetchRequestReference
		objectID = in.APIDefID.String
	} else if in.EventAPIDefID.Valid {
		objectType = model.EventAPIFetchRequestReference
		objectID = in.EventAPIDefID.String
	} else if in.DocumentID.Valid {
		objectType = model.DocumentFetchRequestReference
		objectID = in.DocumentID.String
	}

	var filter *string
	if in.Filter.Valid {
		filter = &in.Filter.String
	}

	return model.FetchRequest{
		ID:         in.ID,
		Tenant:     in.TenantID,
		ObjectID:   objectID,
		ObjectType: objectType,
		Status: &model.FetchRequestStatus{
			Timestamp: in.StatusTimestamp,
			Condition: model.FetchRequestStatusCondition(in.StatusCondition),
		},
		URL:    in.URL,
		Mode:   model.FetchMode(in.Mode),
		Filter: filter,
		Auth:   authPtr,
	}, nil
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
