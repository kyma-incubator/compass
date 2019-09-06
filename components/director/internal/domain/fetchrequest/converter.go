package fetchrequest

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
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

	auth, err := c.authToEntity(in.Auth)
	if err != nil {
		return Entity{}, errors.Wrap(err, "while converting Auth")
	}

	filter := repo.NewNullableString(in.Filter)

	refID := repo.NewValidNullableString(in.ObjectID)
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
	objectID, objectType, err := c.objectReferenceFromEntity(in)
	if err != nil {
		return model.FetchRequest{}, errors.Wrap(err, "while determining object reference")
	}

	auth, err := c.authToModel(in.Auth)
	if err != nil {
		return model.FetchRequest{}, errors.Wrap(err, "while converting Auth")
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
		Filter: repo.StringPtrFromNullableString(in.Filter),
		Auth:   auth,
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

func (c *converter) authToEntity(in *model.Auth) (sql.NullString, error) {
	var auth sql.NullString
	if in == nil {
		return sql.NullString{}, nil
	}

	authMarshalled, err := json.Marshal(in)
	if err != nil {
		return sql.NullString{}, errors.Wrap(err, "while marshalling Auth")
	}

	auth = repo.NewValidNullableString(string(authMarshalled))
	return auth, nil
}

func (c *converter) authToModel(in sql.NullString) (*model.Auth, error) {
	if !in.Valid {
		return nil, nil
	}

	var auth model.Auth
	err := json.Unmarshal([]byte(in.String), &auth)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshalling Auth")
	}

	return &auth, nil
}

func (c *converter) objectReferenceFromEntity(in Entity) (string, model.FetchRequestReferenceObjectType, error) {
	if in.APIDefID.Valid {
		return in.APIDefID.String, model.APIFetchRequestReference, nil
	}

	if in.EventAPIDefID.Valid {
		return in.EventAPIDefID.String, model.EventAPIFetchRequestReference, nil
	}

	if in.DocumentID.Valid {
		return in.DocumentID.String, model.DocumentFetchRequestReference, nil
	}

	return "", "", fmt.Errorf("Incorrect Object Reference ID and its type for Entity with ID '%s'", in.ID)
}
