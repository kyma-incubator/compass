package labeldef

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
)

// Repository missing godoc
//go:generate mockery --name=Repository --output=automock --outpkg=automock --case=underscore
type Repository interface {
	Create(ctx context.Context, def model.LabelDefinition) error
	Upsert(ctx context.Context, tenant string, def model.LabelDefinition) error
	GetByKey(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error)
	Update(ctx context.Context, def model.LabelDefinition) error
	Exists(ctx context.Context, tenant string, key string) (bool, error)
	List(ctx context.Context, tenant string) ([]model.LabelDefinition, error)
	DeleteByKey(ctx context.Context, tenant, key string) error
}

// ScenarioAssignmentLister missing godoc
//go:generate mockery --name=ScenarioAssignmentLister --output=automock --outpkg=automock --case=underscore
type ScenarioAssignmentLister interface {
	List(ctx context.Context, tenant string, pageSize int, cursor string) (*model.AutomaticScenarioAssignmentPage, error)
}

// LabelRepository missing godoc
//go:generate mockery --name=LabelRepository --output=automock --outpkg=automock --case=underscore
type LabelRepository interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
	ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
	ListByKey(ctx context.Context, tenant, key string) ([]*model.Label, error)
	Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error
	DeleteAll(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) error
	DeleteByKey(ctx context.Context, tenant string, key string) error
}

// TenantRepository missing godoc
//go:generate mockery --name=TenantRepository --output=automock --outpkg=automock --case=underscore
type TenantRepository interface {
	Get(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

// UIDService missing godoc
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	defaultScenarioEnabled   bool
	repo                     Repository
	labelRepo                LabelRepository
	scenarioAssignmentLister ScenarioAssignmentLister
	tenantRepo               TenantRepository
	uidService               UIDService
}

// NewService creates new label definition service
func NewService(repo Repository, labelRepo LabelRepository, scenarioAssignmentLister ScenarioAssignmentLister, tenantRepo TenantRepository, uidService UIDService, defaultScenarioEnabled bool) *service {
	return &service{
		defaultScenarioEnabled:   defaultScenarioEnabled,
		repo:                     repo,
		labelRepo:                labelRepo,
		scenarioAssignmentLister: scenarioAssignmentLister,
		tenantRepo:               tenantRepo,
		uidService:               uidService,
	}
}

// Create creates label definition
func (s *service) Create(ctx context.Context, def model.LabelDefinition) (model.LabelDefinition, error) {
	id := s.uidService.Generate()
	def.ID = id

	if err := s.repo.Create(ctx, def); err != nil {
		return model.LabelDefinition{}, errors.Wrap(err, "while storing Label Definition")
	}
	// TODO get from DB?
	return def, nil
}

// CreateWithFormations creates label definition with the provided formations
func (s *service) CreateWithFormations(ctx context.Context, tnt string, formations []string) error {
	formations = s.addDefaultScenarioIfEnabled(formations)

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
	// TODO: Once proper tenant initialization, with creating scenarios LD, is introduced this hack should be removed
	if key == model.ScenariosKey {
		err := s.EnsureScenariosLabelDefinitionExists(ctx, tenant)
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

// List returns all label definitions for the provided tenant
func (s *service) List(ctx context.Context, tenant string) ([]model.LabelDefinition, error) {
	defs, err := s.repo.List(ctx, tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching Label Definitions")
	}
	return defs, nil
}

// Update updates the provided label definition
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
		if err := s.ValidateExistingLabelsAgainstSchema(ctx, *def.Schema, def.Tenant, def.Key); err != nil {
			return err
		}
		if err := s.ValidateAutomaticScenarioAssignmentAgainstSchema(ctx, *def.Schema, def.Tenant, def.Key); err != nil {
			return errors.Wrap(err, "while validating Scenario Assignments against a new schema")
		}
	}

	if err := s.repo.Update(ctx, *ld); err != nil {
		return errors.Wrap(err, "while updating Label Definition")
	}

	return nil
}

// Upsert creates or updates label definitions depending on its existence
func (s service) Upsert(ctx context.Context, tenant string, def model.LabelDefinition) error {
	def.ID = s.uidService.Generate()

	if err := s.repo.Upsert(ctx, tenant, def); err != nil {
		return errors.Wrapf(err, "while upserting Label Definition with id %s and key %s", def.ID, def.Key)
	}
	log.C(ctx).Debugf("Successfully upserted Label Definition with id %s and key %s", def.ID, def.Key)

	return nil
}

// Delete removes label definitions
func (s *service) Delete(ctx context.Context, tenant, key string, deleteRelatedLabels bool) error {
	if key == model.ScenariosKey {
		return fmt.Errorf("labelDefinition with key %s can not be deleted", model.ScenariosKey)
	}

	ld, err := s.Get(ctx, tenant, key)
	if err != nil {
		return errors.Wrap(err, "while getting Label Definition")
	}
	if ld == nil {
		return fmt.Errorf("labelDefinition with key %s not found", key)
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

// EnsureScenariosLabelDefinitionExists creates scenario label definition if missing
func (s *service) EnsureScenariosLabelDefinitionExists(ctx context.Context, tenant string) error {
	ldExists, err := s.repo.Exists(ctx, tenant, model.ScenariosKey)
	if err != nil {
		return errors.Wrapf(err, "while checking if Label Definition with key %s exists", model.ScenariosKey)
	}
	if !ldExists {
		schema, err := NewSchemaForFormations([]string{"DEFAULT"})
		if err != nil {
			return errors.Wrapf(err, "while creaing new schema for key %s", model.ScenariosKey)
		}
		return s.repo.Create(ctx, model.LabelDefinition{
			ID:      s.uidService.Generate(),
			Tenant:  tenant,
			Key:     model.ScenariosKey,
			Schema:  &schema,
			Version: 0,
		})
	}
	return nil
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

// AddDefaultScenarioIfEnabled adds DEFAULT scenario if defaultScenarioEnabled
func (s *service) AddDefaultScenarioIfEnabled(ctx context.Context, tnt string, labels *map[string]interface{}) {
	if labels == nil || !s.defaultScenarioEnabled {
		return
	}

	// do not add scenario to subaccounts
	btm, err := s.tenantRepo.Get(ctx, tnt)
	if err != nil {
		log.C(ctx).Errorf("Could not get tenant from db: %v", err)
		return
	}
	if btm.Type == tenant.Subaccount {
		log.C(ctx).Infof("Will not add DEFAULT scenario for tenant: %s with type %s", btm.ID, btm.Type)
		return
	}

	if _, ok := (*labels)[model.ScenariosKey]; !ok {
		if *labels == nil {
			*labels = map[string]interface{}{}
		}
		(*labels)[model.ScenariosKey] = model.ScenariosDefaultValue
		log.C(ctx).Debug("Successfully added Default scenario")
	}
}

func (s *service) addDefaultScenarioIfEnabled(formations []string) []string {
	if !s.defaultScenarioEnabled {
		return formations
	}

	for _, f := range formations {
		if f == "DEFAULT" {
			return formations
		}
	}

	return append(formations, "DEFAULT")
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
			return apperrors.NewInvalidDataError(fmt.Sprintf(`label with key="%s" is not valid against new schema for %s with ID="%s": %s`, label.Key, label.ObjectType, label.ObjectID, result.Error))
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
	newSchema := model.NewScenariosSchema()
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
