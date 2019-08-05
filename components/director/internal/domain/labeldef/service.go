package labeldef

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

type service struct {
	repo       Repository
	labelRepo  LabelRepository
	uidService UIDService
}

func NewService(r Repository, lr LabelRepository, uidService UIDService) *service {
	return &service{
		repo:       r,
		labelRepo:  lr,
		uidService: uidService,
	}
}

//go:generate mockery -name=Repository -output=automock -outpkg=automock -case=underscore
type Repository interface {
	Create(ctx context.Context, def model.LabelDefinition) error
	GetByKey(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error)
	List(ctx context.Context, tenant string) ([]model.LabelDefinition, error)
	Update(ctx context.Context, def model.LabelDefinition) error
	Exists(ctx context.Context, tenant string, key string) (bool, error)
}

//go:generate mockery -name=LabelRepository -output=automock -outpkg=automock -case=underscore
type LabelRepository interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
	ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
	ListByKey(ctx context.Context, tenant, key string) ([]*model.Label, error)
	Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error
	DeleteAll(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

func (s *service) Create(ctx context.Context, def model.LabelDefinition) (model.LabelDefinition, error) {
	id := s.uidService.Generate()
	def.ID = id
	if err := def.Validate(); err != nil {
		return model.LabelDefinition{}, errors.Wrap(err, "while validation Label Definition")
	}

	if err := s.repo.Create(ctx, def); err != nil {
		return model.LabelDefinition{}, errors.Wrap(err, "while storing Label Definition")
	}
	// TODO get from DB?
	return def, nil
}
func (s *service) Get(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error) {
	def, err := s.repo.GetByKey(ctx, tenant, key)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching Label Definition")
	}
	return def, nil
}

func (s *service) List(ctx context.Context, tenant string) ([]model.LabelDefinition, error) {
	defs, err := s.repo.List(ctx, tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching Label Definitions")
	}
	return defs, nil
}

func (s *service) Update(ctx context.Context, def model.LabelDefinition) (model.LabelDefinition, error) {
	if err := def.ValidateForUpdate(); err != nil {
		return model.LabelDefinition{}, errors.Wrap(err, "while validating Label Definition")
	}

	ld, err := s.repo.GetByKey(ctx, def.Tenant, def.Key)
	if err != nil {
		return model.LabelDefinition{}, errors.Wrap(err, "while receiving Label Definition")
	}

	if ld == nil {
		return model.LabelDefinition{}, fmt.Errorf("definition with %s key doesn't exist", def.Key)
	}

	ld.Schema = def.Schema

	existingLabels, err := s.labelRepo.ListByKey(ctx, def.Tenant, def.Key)

	validator, err := jsonschema.NewValidatorFromRawSchema(*def.Schema)
	if err != nil {
		return model.LabelDefinition{}, errors.Wrap(err, "while creating validator for new schema")
	}

	for _, label := range existingLabels {
		ok, err := validator.ValidateRaw(label)
		if err != nil {
			return model.LabelDefinition{}, errors.Wrap(err, "while validating existing labels against new schema")
		}

		if ok == false {
			return model.LabelDefinition{}, fmt.Errorf("label with key %s is not valid against new schema", label.Key)
		}

	}

	if err := s.repo.Update(ctx, *ld); err != nil {
		return model.LabelDefinition{}, errors.Wrap(err, "while updating Label Definition")
	}

	return *ld, nil
}
