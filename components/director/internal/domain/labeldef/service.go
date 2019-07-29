package labeldef

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"
	"github.com/pkg/errors"
)

type service struct {
	repo       Repository
	uidService UIDService
}

func NewService(r Repository, uidService UIDService) *service {
	return &service{
		repo:       r,
		uidService: uidService,
	}
}

type Repository interface {
	Create(ctx context.Context, def model.LabelDefinition) error
}

type UIDService interface {
	Generate() string
}

func (s *service) Create(ctx context.Context, def model.LabelDefinition) (model.LabelDefinition, error) {
	id := s.uidService.Generate()
	def.ID = id
	// for given tenant, label def with given key does not exist
	if err := def.Validate(); err != nil {
		return model.LabelDefinition{}, errors.Wrapf(err, "while validation label definition [key=%s]", def.Key)
	}
	// if schema provided, they should be valid
	if def.Schema != nil {
		if _, err := jsonschema.NewValidatorFromRawSchema(def.Schema); err != nil {
			return model.LabelDefinition{}, errors.Wrapf(err, "while validating schema: [%v]", def.Schema)
		}
	}
	if err := s.repo.Create(ctx, def); err != nil {
		return model.LabelDefinition{}, errors.Wrap(err, "while storing Label Definition")
	}
	// TODO get from DB?
	return def, nil

}
