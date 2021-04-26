package labeldef

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery --name=Repository --output=automock --outpkg=automock --case=underscore
type Repository interface {
	Create(ctx context.Context, def model.LabelDefinition) error
	Upsert(ctx context.Context, def model.LabelDefinition) error
	GetByKey(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error)
	Update(ctx context.Context, def model.LabelDefinition) error
	Exists(ctx context.Context, tenant string, key string) (bool, error)
	List(ctx context.Context, tenant string) ([]model.LabelDefinition, error)
	DeleteByKey(ctx context.Context, tenant, key string) error
}

//go:generate mockery --name=ScenarioAssignmentLister --output=automock --outpkg=automock --case=underscore
type ScenarioAssignmentLister interface {
	List(ctx context.Context, tenant string, pageSize int, cursor string) (*model.AutomaticScenarioAssignmentPage, error)
}

//go:generate mockery --name=LabelRepository --output=automock --outpkg=automock --case=underscore
type LabelRepository interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
	ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
	ListByKey(ctx context.Context, tenant, key string) ([]*model.Label, error)
	Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error
	DeleteAll(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) error
	DeleteByKey(ctx context.Context, tenant string, key string) error
}

//go:generate mockery --name=ScenariosService --output=automock --outpkg=automock --case=underscore
type ScenariosService interface {
	EnsureScenariosLabelDefinitionExists(ctx context.Context, tenant string) error
}

//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	repo                     Repository
	labelRepo                LabelRepository
	scenarioAssignmentLister ScenarioAssignmentLister
	scenariosService         ScenariosService
	uidService               UIDService
}

func NewService(repo Repository, labelRepo LabelRepository, scenarioAssignmentLister ScenarioAssignmentLister, scenariosService ScenariosService, uidService UIDService) *service {
	return &service{
		repo:                     repo,
		labelRepo:                labelRepo,
		scenarioAssignmentLister: scenarioAssignmentLister,
		scenariosService:         scenariosService,
		uidService:               uidService,
	}
}

func (s *service) Create(ctx context.Context, def model.LabelDefinition) (model.LabelDefinition, error) {
	id := s.uidService.Generate()
	def.ID = id

	if err := s.repo.Create(ctx, def); err != nil {
		return model.LabelDefinition{}, errors.Wrap(err, "while storing Label Definition")
	}
	// TODO get from DB?
	return def, nil
}

func (s *service) Get(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error) {
	// TODO: Once proper tenant initialization, with creating scenarios LD, is introduced this hack should be removed
	if key == model.ScenariosKey {
		err := s.scenariosService.EnsureScenariosLabelDefinitionExists(ctx, tenant)
		if err != nil {
			return nil, err
		}
	}

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
		if err := s.validateAutomaticScenarioAssignmentAgainstSchema(ctx, *def.Schema, def.Tenant, def.Key); err != nil {
			return errors.Wrap(err, "while validating Scenario Assignments against a new schema")
		}
	}

	if err := s.repo.Update(ctx, *ld); err != nil {
		return errors.Wrap(err, "while updating Label Definition")
	}

	return nil
}

func (s service) Upsert(ctx context.Context, def model.LabelDefinition) error {
	def.ID = s.uidService.Generate()

	err := s.repo.Upsert(ctx, def)
	if err != nil {
		return errors.Wrapf(err, "while upserting Label Definition with id %s and key %s", def.ID, def.Key)
	}
	log.C(ctx).Debugf("Successfully upserted Label Definition with id %s and key %s", def.ID, def.Key)

	return nil
}

func (s *service) Delete(ctx context.Context, tenant, key string, deleteRelatedLabels bool) error {
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

	if deleteRelatedLabels {
		err := s.labelRepo.DeleteByKey(ctx, tenant, key)
		if err != nil {
			return errors.Wrapf(err, `while deleting labels with key "%s"`, key)
		}
	}

	existingLabels, err := s.labelRepo.ListByKey(ctx, tenant, key)
	if err != nil {
		return errors.Wrap(err, "while listing labels by key")
	}
	if len(existingLabels) > 0 {
		return apperrors.NewInvalidOperationError("could not delete label definition, it is already used by at least one label")
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
			return apperrors.NewInvalidDataError(fmt.Sprintf(`label with key="%s" is not valid against new schema for %s with ID="%s": %s`, label.Key, label.ObjectType, label.ObjectID, result.Error))
		}
	}
	return nil
}

func (s *service) validateAutomaticScenarioAssignmentAgainstSchema(ctx context.Context, schema interface{}, tenantID, key string) error {
	if key != model.ScenariosKey {
		return nil
	}

	validator, err := jsonschema.NewValidatorFromRawSchema(schema)
	if err != nil {
		return errors.Wrap(err, "while creating validator for a new schema")
	}
	inUse, err := s.fetchScenariosFromAssignments(ctx, tenantID)
	if err != nil {
		return err
	}
	for _, used := range inUse {
		res, err := validator.ValidateRaw([]interface{}{used})
		if err != nil {
			return errors.Wrapf(err, "while validating scenario assignment [scenario=%s] with a new schema", used)
		}
		if res.Error != nil {
			return errors.Wrapf(res.Error, "Scenario Assignment [scenario=%s] is not valid against a new schema", used)
		}
	}
	return nil
}

func (s *service) fetchScenariosFromAssignments(ctx context.Context, tenantID string) ([]string, error) {
	m := make(map[string]struct{})
	pageSize := 100
	cursor := ""
	for {
		page, err := s.scenarioAssignmentLister.List(ctx, tenantID, pageSize, cursor)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting page of Automatic Scenario Assignments")
		}
		for _, a := range page.Data {
			m[a.ScenarioName] = struct{}{}
		}
		if !page.PageInfo.HasNextPage {
			break
		}
		cursor = page.PageInfo.EndCursor
	}

	var out []string
	for k, _ := range m {
		out = append(out, k)
	}
	return out, nil
}
