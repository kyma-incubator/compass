package labeldef

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"
	"github.com/pkg/errors"
)

// Repository missing godoc
//go:generate mockery --name=Repository --output=automock --outpkg=automock --case=underscore --disable-version-string
type Repository interface {
	Create(ctx context.Context, def model.LabelDefinition) error
	Upsert(ctx context.Context, def model.LabelDefinition) error
	GetByKey(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error)
	Update(ctx context.Context, def model.LabelDefinition) error
	Exists(ctx context.Context, tenant string, key string) (bool, error)
	List(ctx context.Context, tenant string) ([]model.LabelDefinition, error)
}

// ScenarioAssignmentLister missing godoc
//go:generate mockery --name=ScenarioAssignmentLister --output=automock --outpkg=automock --case=underscore --disable-version-string
type ScenarioAssignmentLister interface {
	List(ctx context.Context, tenant string, pageSize int, cursor string) (*model.AutomaticScenarioAssignmentPage, error)
}

// LabelRepository missing godoc
//go:generate mockery --name=LabelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelRepository interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
	ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
	ListByKey(ctx context.Context, tenant, key string) ([]*model.Label, error)
	Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error
	DeleteAll(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) error
	DeleteByKey(ctx context.Context, tenant string, key string) error
}

// TenantRepository missing godoc
//go:generate mockery --name=TenantRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantRepository interface {
	Get(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

// UIDService missing godoc
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	repo                     Repository
	labelRepo                LabelRepository
	scenarioAssignmentLister ScenarioAssignmentLister
	tenantRepo               TenantRepository
	uidService               UIDService
}

// NewService creates new label definition service
func NewService(repo Repository, labelRepo LabelRepository, scenarioAssignmentLister ScenarioAssignmentLister, tenantRepo TenantRepository, uidService UIDService) *service {
	return &service{
		repo:                     repo,
		labelRepo:                labelRepo,
		scenarioAssignmentLister: scenarioAssignmentLister,
		tenantRepo:               tenantRepo,
		uidService:               uidService,
	}
}

// CreateWithFormations creates label definition with the provided formations
func (s *service) CreateWithFormations(ctx context.Context, tnt string, formations []string) error {
	schema, err := NewSchemaForFormations(formations)
	if err != nil {
		return errors.Wrapf(err, "while creaing new schema for key %s", model.ScenariosKey)
	}

	return s.repo.Create(ctx, model.LabelDefinition{
		ID:      s.uidService.Generate(),
		Tenant:  tnt,
		Key:     model.ScenariosKey,
		Schema:  &schema,
		Version: 0,
	})
}

// Get returns the tenant scoped label definition with the provided key
func (s *service) Get(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error) {
	def, err := s.repo.GetByKey(ctx, tenant, key)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching Label Definition")
	}
	return def, nil
}

// List returns all label definitions for the provided tenant
func (s *service) List(ctx context.Context, tenant string) ([]model.LabelDefinition, error) {
	defs, err := s.repo.List(ctx, tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching Label Definitions")
	}
	return defs, nil
}

// GetAvailableScenarios returns available scenarios based on scenario label definition
func (s *service) GetAvailableScenarios(ctx context.Context, tenantID string) ([]string, error) {
	def, err := s.repo.GetByKey(ctx, tenantID, model.ScenariosKey)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting `%s` label definition", model.ScenariosKey)
	}
	if def.Schema == nil {
		return nil, fmt.Errorf("missing schema for `%s` label definition", model.ScenariosKey)
	}

	return ParseFormationsFromSchema(def.Schema)
}

// ValidateExistingLabelsAgainstSchema validates the existing labels based on the provided schema
func (s *service) ValidateExistingLabelsAgainstSchema(ctx context.Context, schema interface{}, tenant, key string) error {
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
			return apperrors.NewInvalidDataError(fmt.Sprintf(`label with key="%s" and value="%s" is not valid against new schema for %s with ID="%s": %s`, label.Key, label.Value, label.ObjectType, label.ObjectID, result.Error))
		}
	}
	return nil
}

// ValidateAutomaticScenarioAssignmentAgainstSchema validates the existing scenario assignments based on the provided schema
func (s *service) ValidateAutomaticScenarioAssignmentAgainstSchema(ctx context.Context, schema interface{}, tenantID, key string) error {
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

// NewSchemaForFormations returns new scenario schema with the provided formations
func NewSchemaForFormations(formations []string) (interface{}, error) {
	newSchema := model.NewScenariosSchema([]string{})
	items, ok := newSchema["items"]
	if !ok {
		return nil, fmt.Errorf("mandatory property items is missing")
	}
	itemsMap, ok := items.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("items property could not be converted")
	}

	itemsMap["enum"] = formations
	return newSchema, nil
}

type formation struct {
	Items struct {
		Enum []string
	}
}

// ParseFormationsFromSchema returns available scenarios from the provided schema
func ParseFormationsFromSchema(schema *interface{}) ([]string, error) {
	b, err := json.Marshal(schema)
	if err != nil {
		return nil, errors.Wrapf(err, "while marshaling schema")
	}
	f := formation{}
	if err = json.Unmarshal(b, &f); err != nil {
		return nil, errors.Wrapf(err, "while unmarshaling schema to %T", f)
	}
	return f.Items.Enum, nil
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

	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out, nil
}
