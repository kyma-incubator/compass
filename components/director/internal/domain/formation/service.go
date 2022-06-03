package formation

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=labelDefRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelDefRepository interface {
	Create(ctx context.Context, def model.LabelDefinition) error
	Exists(ctx context.Context, tenant string, key string) (bool, error)
	GetByKey(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error)
	UpdateWithVersion(ctx context.Context, def model.LabelDefinition) error
}

//go:generate mockery --exported --name=labelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelRepository interface {
	Delete(context.Context, string, model.LabelableObject, string, string) error
}

//go:generate mockery --exported --name=runtimeRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeRepository interface {
	ListAll(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) ([]*model.Runtime, error)
	Exists(ctx context.Context, tenant, id string) (bool, error)
}

//go:generate mockery --exported --name=runtimeContextRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeContextRepository interface {
	ListAll(ctx context.Context, tenant string) ([]*model.RuntimeContext, error)
	Exists(ctx context.Context, tenant, id string) (bool, error)
}

//go:generate mockery --exported --name=labelDefService --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelDefService interface {
	CreateWithFormations(ctx context.Context, tnt string, formations []string) error
	ValidateExistingLabelsAgainstSchema(ctx context.Context, schema interface{}, tenant, key string) error
	ValidateAutomaticScenarioAssignmentAgainstSchema(ctx context.Context, schema interface{}, tenantID, key string) error
	EnsureScenariosLabelDefinitionExists(ctx context.Context, tenantID string) error
	GetAvailableScenarios(ctx context.Context, tenantID string) ([]string, error)
}

//go:generate mockery --exported --name=labelService --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelService interface {
	CreateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error
	UpdateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error
	GetLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) (*model.Label, error)
}

//go:generate mockery --exported --name=uidService --output=automock --outpkg=automock --case=underscore --disable-version-string
type uidService interface {
	Generate() string
}

//go:generate mockery --exported --name=automaticFormationAssignmentService --output=automock --outpkg=automock --case=underscore --disable-version-string
type automaticFormationAssignmentService interface {
	GetForScenarioName(ctx context.Context, scenarioName string) (model.AutomaticScenarioAssignment, error)
}

//go:generate mockery --exported --name=automaticFormationAssignmentRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type automaticFormationAssignmentRepository interface {
	Create(ctx context.Context, model model.AutomaticScenarioAssignment) error
	DeleteForTargetTenant(ctx context.Context, tenantID string, targetTenantID string) error
	DeleteForScenarioName(ctx context.Context, tenantID string, scenarioName string) error
	ListAll(ctx context.Context, tenantID string) ([]*model.AutomaticScenarioAssignment, error)
}

//go:generate mockery --exported --name=tenantService --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantService interface {
	CreateManyIfNotExists(ctx context.Context, tenantInputs ...model.BusinessTenantMappingInput) error
	GetInternalTenant(ctx context.Context, externalTenant string) (string, error)
}

type service struct {
	labelDefRepository labelDefRepository
	labelRepository    labelRepository
	labelService       labelService
	labelDefService    labelDefService
	asaService         automaticFormationAssignmentService
	uuidService        uidService
	tenantSvc          tenantService
	repo               automaticFormationAssignmentRepository
	runtimeRepo        runtimeRepository
	runtimeContextRepo runtimeContextRepository
}

// NewService creates formation service
func NewService(labelDefRepository labelDefRepository, labelRepository labelRepository, labelService labelService, uuidService uidService, labelDefService labelDefService, asaRepo automaticFormationAssignmentRepository, asaService automaticFormationAssignmentService, tenantSvc tenantService, runtimeRepo runtimeRepository, runtimeContextRepo runtimeContextRepository) *service {
	return &service{
		labelDefRepository: labelDefRepository,
		labelRepository:    labelRepository,
		labelService:       labelService,
		labelDefService:    labelDefService,
		asaService:         asaService,
		uuidService:        uuidService,
		tenantSvc:          tenantSvc,
		repo:               asaRepo,
		runtimeRepo:        runtimeRepo,
		runtimeContextRepo: runtimeContextRepo,
	}
}

// GetFormationsForObject returns slice of formations for entity with ID objID and type objType
func (s *service) GetFormationsForObject(ctx context.Context, tnt string, objType model.LabelableObject, objID string) ([]string, error) {
	labelInput := &model.LabelInput{
		Key:        model.ScenariosKey,
		ObjectID:   objID,
		ObjectType: objType,
	}
	existingLabel, err := s.labelService.GetLabel(ctx, tnt, labelInput)
	if err != nil {
		return nil, errors.Wrapf(err, "while fetching scenario label for %q with id %q", objType, objID)
	}

	return label.ValueToStringsSlice(existingLabel.Value)
}

// CreateFormation adds the provided formation to the scenario label definitions of the given tenant.
// If the scenario label definition does not exist it will be created
func (s *service) CreateFormation(ctx context.Context, tnt string, formation model.Formation) (*model.Formation, error) {
	f, err := s.modifyFormations(ctx, tnt, formation.Name, addFormation)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			if err = s.labelDefService.CreateWithFormations(ctx, tnt, []string{formation.Name}); err != nil {
				return nil, err
			}
			return &model.Formation{Name: formation.Name}, nil
		}
		return nil, err
	}
	return f, nil
}

// DeleteFormation removes the provided formation from the scenario label definitions of the given tenant.
func (s *service) DeleteFormation(ctx context.Context, tnt string, formation model.Formation) (*model.Formation, error) {
	return s.modifyFormations(ctx, tnt, formation.Name, deleteFormation)
}

// AssignFormation assigns object based on graphql.FormationObjectType.
// For objectTypes graphql.FormationObjectType is graphql.FormationObjectTypeApplication, graphql.FormationObjectTypeRuntime and
// graphql.FormationObjectTypeRuntimeContext it adds the provided formation to the scenario label of the entity if such exists,
// otherwise new scenario label is created for the entity with the provided formation.
// If the graphql.FormationObjectType is graphql.FormationObjectTypeTenant it will
// create automatic scenario assignment with the caller and target tenant.
func (s *service) AssignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error) {
	switch objectType {
	case graphql.FormationObjectTypeApplication, graphql.FormationObjectTypeRuntime, graphql.FormationObjectTypeRuntimeContext:
		f, err := s.modifyAssignedFormations(ctx, tnt, objectID, formation, objectTypeToLabelableObject(objectType), addFormation)
		if err != nil {
			if apperrors.IsNotFoundError(err) {
				labelInput := newLabelInput(formation.Name, objectID, objectTypeToLabelableObject(objectType))
				if err = s.labelService.CreateLabel(ctx, tnt, s.uuidService.Generate(), labelInput); err != nil {
					return nil, err
				}
				return &formation, nil
			}
			return nil, err
		}
		return f, nil
	case graphql.FormationObjectTypeTenant:
		tenantID, err := s.tenantSvc.GetInternalTenant(ctx, objectID)
		if err != nil {
			return nil, err
		}

		if _, err = s.CreateAutomaticScenarioAssignment(ctx, newAutomaticScenarioAssignmentModel(formation.Name, tnt, tenantID)); err != nil {
			return nil, err
		}
		return &formation, err
	default:
		return nil, fmt.Errorf("unknown formation type %s", objectType)
	}
}

// UnassignFormation unassigns object base on graphql.FormationObjectType.
// For objectType graphql.FormationObjectTypeApplication it removes the provided formation from the
// scenario label of the application.
// For objectTypes graphql.FormationObjectTypeRuntime and graphql.FormationObjectTypeRuntimeContext
// it removes the formation from the scenario label of the runtime if the provided formation is NOT assigned
// from ASA and does nothing if it is assigned from ASA.
// For objectType graphql.FormationObjectTypeTenant it will
// delete the automatic scenario assignment with the caller and target tenant.
func (s *service) UnassignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error) {
	switch objectType {
	case graphql.FormationObjectTypeApplication:
		return s.modifyAssignedFormations(ctx, tnt, objectID, formation, objectTypeToLabelableObject(objectType), deleteFormation)
	case graphql.FormationObjectTypeRuntime, graphql.FormationObjectTypeRuntimeContext:
		if isFormationComingFromASA, err := s.isFormationComingFromASA(ctx, objectID, formation.Name, objectType); err != nil {
			return nil, err
		} else if isFormationComingFromASA {
			return &formation, nil
		}

		return s.modifyAssignedFormations(ctx, tnt, objectID, formation, objectTypeToLabelableObject(objectType), deleteFormation)
	case graphql.FormationObjectTypeTenant:
		asa, err := s.asaService.GetForScenarioName(ctx, formation.Name)
		if err != nil {
			return nil, err
		}
		if err = s.DeleteAutomaticScenarioAssignment(ctx, asa); err != nil {
			return nil, err
		}
		return &formation, nil
	default:
		return nil, fmt.Errorf("unknown formation type %s", objectType)
	}
}

// CreateAutomaticScenarioAssignment creates a new AutomaticScenarioAssignment for a given ScenarioName, Tenant and TargetTenantID
// It also ensures that all runtimes with given scenarios are assigned for the TargetTenantID
func (s *service) CreateAutomaticScenarioAssignment(ctx context.Context, in model.AutomaticScenarioAssignment) (model.AutomaticScenarioAssignment, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return model.AutomaticScenarioAssignment{}, err
	}

	in.Tenant = tenantID
	if err := s.validateThatScenarioExists(ctx, in); err != nil {
		return model.AutomaticScenarioAssignment{}, err
	}

	if err = s.repo.Create(ctx, in); err != nil {
		if apperrors.IsNotUniqueError(err) {
			return model.AutomaticScenarioAssignment{}, apperrors.NewInvalidOperationError("a given scenario already has an assignment")
		}

		return model.AutomaticScenarioAssignment{}, errors.Wrap(err, "while persisting Assignment")
	}

	if err = s.EnsureScenarioAssigned(ctx, in); err != nil {
		return model.AutomaticScenarioAssignment{}, errors.Wrap(err, "while assigning scenario to runtimes matching selector")
	}

	return in, nil
}

// DeleteAutomaticScenarioAssignment deletes the assignment for a given scenario in a scope of a tenant
// It also removes corresponding assigned scenarios for the ASA
func (s *service) DeleteAutomaticScenarioAssignment(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "while loading tenant from context")
	}

	if err = s.repo.DeleteForScenarioName(ctx, tenantID, in.ScenarioName); err != nil {
		return errors.Wrap(err, "while deleting the Assignment")
	}

	if err = s.RemoveAssignedScenario(ctx, in); err != nil {
		return errors.Wrap(err, "while unassigning scenario from runtimes")
	}

	return nil
}

// EnsureScenarioAssigned ensures that the scenario is assigned to all the runtimes and runtimeContexts that are in the ASAs target_tenant_id
func (s *service) EnsureScenarioAssigned(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	return s.processScenario(ctx, in, s.AssignFormation, "assigning")
}

// RemoveAssignedScenario removes all the scenarios that are coming from the provided ASA
func (s *service) RemoveAssignedScenario(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	return s.processScenario(ctx, in, s.UnassignFormation, "unassigning")
}

type processingFunc func(context.Context, string, string, graphql.FormationObjectType, model.Formation) (*model.Formation, error)

func (s *service) processScenario(ctx context.Context, in model.AutomaticScenarioAssignment, processingFunc processingFunc, processingType string) error {
	runtimes, err := s.runtimeRepo.ListAll(ctx, in.TargetTenantID, nil)
	if err != nil {
		return errors.Wrapf(err, "while fetching runtimes in target tenant: %s", in.TargetTenantID)
	}

	for _, r := range runtimes {
		if _, err = processingFunc(ctx, in.Tenant, r.ID, graphql.FormationObjectTypeRuntime, model.Formation{Name: in.ScenarioName}); err != nil {
			return errors.Wrapf(err, "while %s runtime with id %s from formation %s coming from ASA", processingType, r.ID, in.ScenarioName)
		}
	}

	runtimeContexts, err := s.runtimeContextRepo.ListAll(ctx, in.TargetTenantID)
	if err != nil {
		return errors.Wrapf(err, "while fetching runtime contexts in target tenant: %s", in.TargetTenantID)
	}

	for _, rc := range runtimeContexts {
		if _, err = processingFunc(ctx, in.Tenant, rc.ID, graphql.FormationObjectTypeRuntimeContext, model.Formation{Name: in.ScenarioName}); err != nil {
			return errors.Wrapf(err, "while %s runtime context with id %s from formation %s coming from ASA", processingType, rc.ID, in.ScenarioName)
		}
	}

	return nil
}

// RemoveAssignedScenarios removes all the scenarios that are coming from any of the provided ASAs
func (s *service) RemoveAssignedScenarios(ctx context.Context, in []*model.AutomaticScenarioAssignment) error {
	for _, asa := range in {
		if err := s.RemoveAssignedScenario(ctx, *asa); err != nil {
			return errors.Wrapf(err, "while deleting automatic scenario assigment: %s", asa.ScenarioName)
		}
	}
	return nil
}

// DeleteManyASAForSameTargetTenant deletes a list of ASAs for the same targetTenant
// It also removes corresponding scenario assignments coming from the ASAs
func (s *service) DeleteManyASAForSameTargetTenant(ctx context.Context, in []*model.AutomaticScenarioAssignment) error {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	targetTenant, err := s.ensureSameTargetTenant(in)
	if err != nil {
		return errors.Wrap(err, "while ensuring input is valid")
	}

	if err = s.repo.DeleteForTargetTenant(ctx, tenantID, targetTenant); err != nil {
		return errors.Wrap(err, "while deleting the Assignments")
	}

	if err = s.RemoveAssignedScenarios(ctx, in); err != nil {
		return errors.Wrap(err, "while unassigning scenario from runtimes")
	}

	return nil
}

// MergeScenariosFromInputLabelsAndAssignments merges all the scenarios that are part of the resource labels (already added + to be added with the current operation)
// with all the scenarios that should be assigned based on ASAs.
func (s *service) MergeScenariosFromInputLabelsAndAssignments(ctx context.Context, inputLabels map[string]interface{}, runtimeID string) ([]interface{}, error) {
	scenariosFromAssignments, err := s.GetScenariosFromMatchingASAs(ctx, runtimeID, graphql.FormationObjectTypeRuntime)
	scenariosSet := make(map[string]struct{}, len(scenariosFromAssignments))

	if err != nil {
		return nil, errors.Wrapf(err, "while getting scenarios for selector labels")
	}

	for _, scenario := range scenariosFromAssignments {
		scenariosSet[scenario] = struct{}{}
	}

	scenariosFromInput, isScenarioLabelInInput := inputLabels[model.ScenariosKey]

	if isScenarioLabelInInput {
		scenarioLabels, err := label.ValueToStringsSlice(scenariosFromInput)
		if err != nil {
			return nil, errors.Wrap(err, "while converting scenarios label to a string slice")
		}

		for _, scenario := range scenarioLabels {
			scenariosSet[scenario] = struct{}{}
		}
	}

	scenarios := make([]interface{}, 0, len(scenariosSet))
	for k := range scenariosSet {
		scenarios = append(scenarios, k)
	}
	return scenarios, nil
}

// GetScenariosFromMatchingASAs gets all the scenarios that should be added to the runtime based on the matching Automatic Scenario Assignments
// In order to do that, the ASAs should be searched in the caller tenant as this is the tenant that modifies the runtime and this is the tenant that the ASA
// produced labels should be added to.
func (s *service) GetScenariosFromMatchingASAs(ctx context.Context, objectID string, objType graphql.FormationObjectType) ([]string, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	match, err := s.getMatchingFuncByFormationObjectType(objType)
	if err != nil {
		return nil, err
	}

	scenarioAssignments, err := s.repo.ListAll(ctx, tenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listinng Automatic Scenario Assignments in tenant: %s", tenantID)
	}

	matchingASAs := make([]*model.AutomaticScenarioAssignment, 0, len(scenarioAssignments))

	for _, scenarioAssignment := range scenarioAssignments {
		matches, err := match(ctx, scenarioAssignment, objectID)
		if err != nil {
			return nil, errors.Wrapf(err, "while checkig if asa matches runtime with ID %s", objectID)
		}
		if matches {
			matchingASAs = append(matchingASAs, scenarioAssignment)
		}
	}

	scenarios := make([]string, 0)
	for _, sa := range matchingASAs {
		scenarios = append(scenarios, sa.ScenarioName)
	}
	return scenarios, nil
}

type matchingFunc func(ctx context.Context, asa *model.AutomaticScenarioAssignment, runtimeID string) (bool, error)

func (s *service) getMatchingFuncByFormationObjectType(objType graphql.FormationObjectType) (matchingFunc, error) {
	switch objType {
	case graphql.FormationObjectTypeRuntime:
		return s.isASAMatchingRuntime, nil
	case graphql.FormationObjectTypeRuntimeContext:
		return s.isASAMatchingRuntimeContext, nil
	}
	return nil, errors.New(fmt.Sprintf("unexpected formation object type %q", objType))
}

func (s *service) isASAMatchingRuntime(ctx context.Context, asa *model.AutomaticScenarioAssignment, runtimeID string) (bool, error) {
	return s.runtimeRepo.Exists(ctx, asa.TargetTenantID, runtimeID)
}

func (s *service) isASAMatchingRuntimeContext(ctx context.Context, asa *model.AutomaticScenarioAssignment, runtimeContextID string) (bool, error) {
	return s.runtimeContextRepo.Exists(ctx, asa.TargetTenantID, runtimeContextID)
}

func (s *service) isFormationComingFromASA(ctx context.Context, objectID, formation string, objectType graphql.FormationObjectType) (bool, error) {
	formationsFromASA, err := s.GetScenariosFromMatchingASAs(ctx, objectID, objectType)
	if err != nil {
		return false, errors.Wrapf(err, "while getting formations from ASAs for %s with id: %q", objectType, objectID)
	}

	for _, formationFromASA := range formationsFromASA {
		if formation == formationFromASA {
			return true, nil
		}
	}

	return false, nil
}

func (s *service) modifyFormations(ctx context.Context, tnt, formationName string, modificationFunc modificationFunc) (*model.Formation, error) {
	def, err := s.labelDefRepository.GetByKey(ctx, tnt, model.ScenariosKey)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting `%s` label definition", model.ScenariosKey)
	}
	if def.Schema == nil {
		return nil, fmt.Errorf("missing schema for `%s` label definition", model.ScenariosKey)
	}

	formations, err := labeldef.ParseFormationsFromSchema(def.Schema)
	if err != nil {
		return nil, err
	}

	formations = modificationFunc(formations, formationName)

	schema, err := labeldef.NewSchemaForFormations(formations)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing scenarios")
	}

	if err = s.labelDefService.ValidateExistingLabelsAgainstSchema(ctx, schema, tnt, model.ScenariosKey); err != nil {
		return nil, err
	}
	if err = s.labelDefService.ValidateAutomaticScenarioAssignmentAgainstSchema(ctx, schema, tnt, model.ScenariosKey); err != nil {
		return nil, errors.Wrap(err, "while validating Scenario Assignments against a new schema")
	}

	if err = s.labelDefRepository.UpdateWithVersion(ctx, model.LabelDefinition{
		ID:      def.ID,
		Tenant:  tnt,
		Key:     model.ScenariosKey,
		Schema:  &schema,
		Version: def.Version,
	}); err != nil {
		return nil, err
	}
	return &model.Formation{Name: formationName}, nil
}

func (s *service) modifyAssignedFormations(ctx context.Context, tnt, objectID string, formation model.Formation, objectType model.LabelableObject, modificationFunc modificationFunc) (*model.Formation, error) {
	labelInput := newLabelInput(formation.Name, objectID, objectType)

	existingLabel, err := s.labelService.GetLabel(ctx, tnt, labelInput)
	if err != nil {
		return nil, err
	}

	existingFormations, err := label.ValueToStringsSlice(existingLabel.Value)
	if err != nil {
		return nil, err
	}

	formations := modificationFunc(existingFormations, formation.Name)
	// can not set scenario label to empty value, violates the scenario label definition
	if len(formations) == 0 {
		if err := s.labelRepository.Delete(ctx, tnt, objectType, objectID, model.ScenariosKey); err != nil {
			return nil, err
		}
		return &formation, nil
	}

	labelInput.Value = formations
	labelInput.Version = existingLabel.Version
	if err := s.labelService.UpdateLabel(ctx, tnt, existingLabel.ID, labelInput); err != nil {
		return nil, err
	}
	return &formation, nil
}

type modificationFunc func([]string, string) []string

func addFormation(formations []string, formation string) []string {
	for _, f := range formations {
		if f == formation {
			return formations
		}
	}

	return append(formations, formation)
}

func deleteFormation(formations []string, formation string) []string {
	filteredFormations := make([]string, 0, len(formations))
	for _, f := range formations {
		if f != formation {
			filteredFormations = append(filteredFormations, f)
		}
	}

	return filteredFormations
}

func newLabelInput(formation, objectID string, objectType model.LabelableObject) *model.LabelInput {
	return &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{formation},
		ObjectID:   objectID,
		ObjectType: objectType,
		Version:    0,
	}
}

func newAutomaticScenarioAssignmentModel(formation, callerTenant, targetTenant string) model.AutomaticScenarioAssignment {
	return model.AutomaticScenarioAssignment{
		ScenarioName:   formation,
		Tenant:         callerTenant,
		TargetTenantID: targetTenant,
	}
}

func objectTypeToLabelableObject(objectType graphql.FormationObjectType) (labelableObj model.LabelableObject) {
	switch objectType {
	case graphql.FormationObjectTypeApplication:
		labelableObj = model.ApplicationLabelableObject
	case graphql.FormationObjectTypeRuntime:
		labelableObj = model.RuntimeLabelableObject
	case graphql.FormationObjectTypeTenant:
		labelableObj = model.TenantLabelableObject
	case graphql.FormationObjectTypeRuntimeContext:
		labelableObj = model.RuntimeContextLabelableObject
	}
	return labelableObj
}

func (s *service) ensureSameTargetTenant(in []*model.AutomaticScenarioAssignment) (string, error) {
	if len(in) == 0 || in[0] == nil {
		return "", apperrors.NewInternalError("expected at least one item in Assignments slice")
	}

	targetTenant := in[0].TargetTenantID

	for _, item := range in {
		if item != nil && item.TargetTenantID != targetTenant {
			return "", apperrors.NewInternalError("all input items have to have the same target tenant")
		}
	}

	return targetTenant, nil
}

func (s *service) validateThatScenarioExists(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	availableScenarios, err := s.getAvailableScenarios(ctx, in.Tenant)
	if err != nil {
		return err
	}

	for _, av := range availableScenarios {
		if av == in.ScenarioName {
			return nil
		}
	}

	return apperrors.NewNotFoundError(resource.AutomaticScenarioAssigment, in.ScenarioName)
}

func (s *service) getAvailableScenarios(ctx context.Context, tenantID string) ([]string, error) {
	if err := s.labelDefService.EnsureScenariosLabelDefinitionExists(ctx, tenantID); err != nil {
		return nil, errors.Wrap(err, "while ensuring that `scenarios` label definition exist")
	}

	out, err := s.labelDefService.GetAvailableScenarios(ctx, tenantID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting available scenarios")
	}
	return out, nil
}
