package apptemplateversion

import (
	"database/sql"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type converter struct {
}

// NewConverter missing godoc
func NewConverter() *converter {
	return &converter{}
}

// ToEntity missing godoc
func (c *converter) ToEntity(in *model.ApplicationTemplateVersion) (*Entity, error) {
	if in == nil {
		return nil, nil
	}

	correlationIDs, err := c.correlationIDsToJSON(in.CorrelationIDs)
	if err != nil {
		return nil, apperrors.NewInternalError("")
	}

	output := &Entity{
		ID:                    in.ID,
		Version:               in.Version,
		Title:                 repo.NewNullableString(in.Title),
		ReleaseDate:           in.ReleaseDate,
		CorrelationIDs:        correlationIDs,
		CreatedAt:             in.CreatedAt,
		ApplicationTemplateID: in.ApplicationTemplateID,
	}

	return output, nil
}

// FromEntity missing godoc
func (c *converter) FromEntity(entity *Entity) (*model.ApplicationTemplateVersion, error) {
	if entity == nil {
		return nil, nil
	}

	correlationIDs, err := c.correlationIDsToModel(entity.CorrelationIDs)
	if err != nil {
		return nil, apperrors.NewInternalError("")
	}

	output := &model.ApplicationTemplateVersion{
		ID:                    entity.ID,
		Version:               entity.Version,
		Title:                 repo.StringPtrFromNullableString(entity.Title),
		ReleaseDate:           entity.ReleaseDate,
		CorrelationIDs:        correlationIDs,
		CreatedAt:             entity.CreatedAt,
		ApplicationTemplateID: entity.ApplicationTemplateID,
	}

	return output, nil
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
