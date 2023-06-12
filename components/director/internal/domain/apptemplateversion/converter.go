package apptemplateversion

import (
	"database/sql"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type converter struct {
}

// NewConverter returns a new Converter that can later be used to make the conversions between the service and repository layer representations of a Compass ApplicationTemplateVersion.
func NewConverter() *converter {
	return &converter{}
}

// ToEntity converts the provided service-layer representation of an ApplicationTemplateVersion to the repository-layer one.
func (c *converter) ToEntity(in *model.ApplicationTemplateVersion) *Entity {
	if in == nil {
		return nil
	}

	output := &Entity{
		ID:                    in.ID,
		Version:               in.Version,
		Title:                 repo.NewNullableString(in.Title),
		ReleaseDate:           repo.NewNullableString(in.ReleaseDate),
		CorrelationIDs:        repo.NewNullableStringFromJSONRawMessage(in.CorrelationIDs),
		CreatedAt:             in.CreatedAt,
		ApplicationTemplateID: in.ApplicationTemplateID,
	}

	return output
}

// FromEntity converts the provided Entity repo-layer representation of an ApplicationTemplateVersion to the service-layer representation model.ApplicationTemplateVersion.
func (c *converter) FromEntity(entity *Entity) *model.ApplicationTemplateVersion {
	if entity == nil {
		return nil
	}

	output := &model.ApplicationTemplateVersion{
		ID:                    entity.ID,
		Version:               entity.Version,
		Title:                 repo.StringPtrFromNullableString(entity.Title),
		ReleaseDate:           repo.StringPtrFromNullableString(entity.ReleaseDate),
		CorrelationIDs:        repo.JSONRawMessageFromNullableString(entity.CorrelationIDs),
		CreatedAt:             entity.CreatedAt,
		ApplicationTemplateID: entity.ApplicationTemplateID,
	}

	return output
}

func (c *converter) correlationIDsToModel(in sql.NullString) ([]string, error) {
	if !in.Valid || in.String == "" {
		return nil, nil
	}

	var correlationIDs []string
	err := json.Unmarshal([]byte(in.String), &correlationIDs)
	if err != nil {
		return nil, err
	}

	return correlationIDs, nil
}

func (c *converter) correlationIDsToJSON(in []string) (sql.NullString, error) {
	result := sql.NullString{}

	if in == nil {
		return result, nil
	}

	correlationIDsMarshalled, err := json.Marshal(in)
	if err != nil {
		return result, errors.Wrap(err, "while marshalling correlation IDs")
	}

	return repo.NewValidNullableString(string(correlationIDsMarshalled)), nil
}
