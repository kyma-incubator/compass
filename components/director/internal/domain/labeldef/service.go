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

func NewService(repo Repository, labelRepo LabelRepository, uidService UIDService) *service {
	return &service{
		repo:       repo,
		labelRepo:  labelRepo,
		uidService: uidService,
	}
}

//go:generate mockery -name=Repository -output=automock -outpkg=automock -case=underscore
type Repository interface {
	Create(ctx context.Context, def model.LabelDefinition) error
	GetByKey(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error)
	Update(ctx context.Context, def model.LabelDefinition) error
	Exists(ctx context.Context, tenant string, key string) (bool, error)
	List(ctx context.Context, tenant string) ([]model.LabelDefinition, error)
	DeleteByKey(ctx context.Context, tenant, key string) error
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

func (s *service) Update(ctx context.Context, def model.LabelDefinition) error {
	if err := def.ValidateForUpdate(); err != nil {
		return errors.Wrap(err, "while validating Label Definition")
	}

	ld, err := s.repo.GetByKey(ctx, def.Tenant, def.Key)
	if err != nil {
		return errors.Wrap(err, "while receiving Label Definition")
	}

	if ld == nil {
		return errors.Errorf("definition with %s key doesn't exist", def.Key)
	}

	ld.Schema = def.Schema

	if def.Schema != nil {
		if err := s.validateExistingLabelsAgainstSchema(ctx, *def.Schema, def.Tenant, def.Key); err != nil {
			return err
		}
	}

	if err := s.repo.Update(ctx, *ld); err != nil {
		return errors.Wrap(err, "while updating Label Definition")
	}

	return nil
}

// TODO: Add deleting related labels logic
func (s *service) Delete(ctx context.Context, tenant, key string, deleteRelatedLabels bool) error {
	if deleteRelatedLabels == true {
		return errors.New("deleting related labels is not yet supported")
	}

	if key == model.ScenariosKey {
		return fmt.Errorf("Label Definition with key %s can not be deleted", model.ScenariosKey)
	}

	ld, err := s.Get(ctx, tenant, key)
	if err != nil {
		return errors.Wrap(err, "while getting Label Definition")
	}
	if ld == nil {
		return fmt.Errorf("Label Definition with key %s not found", key)
	}

	existingLabels, err := s.labelRepo.ListByKey(ctx, tenant, key)
	if err != nil {
		return errors.Wrap(err, "while listing labels by key")
	}
	if len(existingLabels) > 0 {
		return errors.New("could not delete label definition, it is already used by at least one label")
	}

	return s.repo.DeleteByKey(ctx, tenant, ld.Key)
}

func (s *service) validateExistingLabelsAgainstSchema(ctx context.Context, schema interface{}, tenant, key string) error {
	existingLabels, err := s.labelRepo.ListByKey(ctx, tenant, key)
	if err != nil {
		return errors.Wrap(err, "while listing labels by key")
	}

	validator, err := jsonschema.NewValidatorFromRawSchema(schema)
	if err != nil {
		return errors.Wrap(err, "while creating validator for new schema")
	}

	for _, label := range existingLabels {
		result, err := validator.ValidateRaw(label.Value)
		if err != nil {
			return errors.Wrap(err, "while validating existing labels against new schema")
		}

		if !result.Valid {
			return errors.Wrapf(result.Error, `label with key "%s" is not valid against new schema for %s with ID "%s"`, label.Key, label.ObjectType, label.ObjectID)
		}
	}
	return nil
}
