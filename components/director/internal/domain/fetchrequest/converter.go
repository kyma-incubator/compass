package fetchrequest

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
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
func (c *converter) ToGraphQL(in *model.FetchRequest) (*graphql.FetchRequest, error) {
	if in == nil {
		return nil, nil
	}

	auth, err := c.authConverter.ToGraphQL(in.Auth)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Auth to GraphQL")
	}

	return &graphql.FetchRequest{
		URL:    in.URL,
		Auth:   auth,
		Mode:   graphql.FetchMode(in.Mode),
		Filter: in.Filter,
		Status: c.statusToGraphQL(in.Status),
	}, nil
}

// InputFromGraphQL missing godoc
func (c *converter) InputFromGraphQL(in *graphql.FetchRequestInput) (*model.FetchRequestInput, error) {
	if in == nil {
		return nil, nil
	}

	var mode *model.FetchMode
	if in.Mode != nil {
		tmp := model.FetchMode(*in.Mode)
		mode = &tmp
	}

	auth, err := c.authConverter.InputFromGraphQL(in.Auth)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Auth input from GraphQL")
	}

	return &model.FetchRequestInput{
		URL:    in.URL,
		Auth:   auth,
		Mode:   mode,
		Filter: in.Filter,
	}, nil
}

// ToEntity missing godoc
func (c *converter) ToEntity(in *model.FetchRequest) (*Entity, error) {
	if in.Status == nil {
		return nil, apperrors.NewInvalidDataError("Invalid input model")
	}

	auth, err := c.authToEntity(in.Auth)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Auth")
	}

	filter := repo.NewNullableString(in.Filter)
	message := repo.NewNullableString(in.Status.Message)
	refID := repo.NewValidNullableString(in.ObjectID)

	var specID sql.NullString
	var documentID sql.NullString
	switch in.ObjectType {
	case model.APISpecFetchRequestReference:
		fallthrough
	case model.EventSpecFetchRequestReference:
		specID = refID
	case model.DocumentFetchRequestReference:
		documentID = refID
	}

	return &Entity{
		ID:              in.ID,
		URL:             in.URL,
		Auth:            auth,
		SpecID:          specID,
		DocumentID:      documentID,
		Mode:            string(in.Mode),
		Filter:          filter,
		StatusCondition: string(in.Status.Condition),
		StatusMessage:   message,
		StatusTimestamp: in.Status.Timestamp,
	}, nil
}

// FromEntity missing godoc
func (c *converter) FromEntity(in *Entity, objectType model.FetchRequestReferenceObjectType) (*model.FetchRequest, error) {
	objectID, err := c.objectIDFromEntity(*in)
	if err != nil {
		return nil, errors.Wrap(err, "while determining object reference")
	}

	auth, err := c.authToModel(in.Auth)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Auth")
	}

	return &model.FetchRequest{
		ID:         in.ID,
		ObjectID:   objectID,
		ObjectType: objectType,
		Status: &model.FetchRequestStatus{
			Timestamp: in.StatusTimestamp,
			Message:   repo.StringPtrFromNullableString(in.StatusMessage),
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
		Message:   in.Message,
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

func (c *converter) objectIDFromEntity(in Entity) (string, error) {
	if in.SpecID.Valid {
		return in.SpecID.String, nil
	}

	if in.DocumentID.Valid {
		return in.DocumentID.String, nil
	}

	return "", fmt.Errorf("incorrect Object Reference ID and its type for Entity with ID %q", in.ID)
}
