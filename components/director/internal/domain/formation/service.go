package formation

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=labelDefRepository --output=automock --outpkg=automock --case=underscore
type labelDefRepository interface {
	Create(ctx context.Context, def model.LabelDefinition) error
	Exists(ctx context.Context, tenant string, key string) (bool, error)
	GetByKey(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error)
	UpdateWithVersion(ctx context.Context, def model.LabelDefinition) error
}

//go:generate mockery --exported --name=labelRepository --output=automock --outpkg=automock --case=underscore
type labelRepository interface {
	Delete(context.Context, string, model.LabelableObject, string, string) error
}

//go:generate mockery --exported --name=labelDefService --output=automock --outpkg=automock --case=underscore
type labelDefService interface {
	CreateWithFormations(ctx context.Context, tnt string, formations []string) error
	ValidateExistingLabelsAgainstSchema(ctx context.Context, schema interface{}, tenant, key string) error
	ValidateAutomaticScenarioAssignmentAgainstSchema(ctx context.Context, schema interface{}, tenantID, key string) error
}

//go:generate mockery --exported --name=labelService --output=automock --outpkg=automock --case=underscore
type labelService interface {
	CreateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error
	UpdateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error
	GetLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) (*model.Label, error)
}

//go:generate mockery --exported --name=uidService --output=automock --outpkg=automock --case=underscore
type uidService interface {
	Generate() string
}

//go:generate mockery --exported --name=automaticFormationAssignmentService --output=automock --outpkg=automock --case=underscore
type automaticFormationAssignmentService interface {
	Create(ctx context.Context, in model.AutomaticScenarioAssignment) (model.AutomaticScenarioAssignment, error)
	GetForScenarioName(ctx context.Context, scenarioName string) (model.AutomaticScenarioAssignment, error)
	Delete(ctx context.Context, in model.AutomaticScenarioAssignment) error
}

//go:generate mockery --exported --name=tenantService --output=automock --outpkg=automock --case=underscore
type tenantService interface {
	CreateManyIfNotExists(ctx context.Context, tenantInputs ...model.BusinessTenantMappingInput) error
	GetInternalTenant(ctx context.Context, externalTenant string) (string, error)
	GetTenantByExternalID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

//go:generate mockery --exported --name=scenarioAssignmentEngine --output=automock --outpkg=automock --case=underscore
type scenarioAssignmentEngine interface {
	GetScenariosFromMatchingASAs(ctx context.Context, runtimeID string) ([]string, error)
}

type service struct {
	labelDefRepository labelDefRepository
	labelRepository    labelRepository
	labelService       labelService
	labelDefService    labelDefService
	asaService         automaticFormationAssignmentService
	uuidService        uidService
	tenantSvc          tenantService
	engine             scenarioAssignmentEngine
}

// NewService creates formation service
func NewService(labelDefRepository labelDefRepository, labelRepository labelRepository, labelService labelService, uuidService uidService, labelDefService labelDefService, asaService automaticFormationAssignmentService, tenantSvc tenantService, engine scenarioAssignmentEngine) *service {
	return &service{
		labelDefRepository: labelDefRepository,
		labelRepository:    labelRepository,
		labelService:       labelService,
		labelDefService:    labelDefService,
		asaService:         asaService,
		uuidService:        uuidService,
		tenantSvc:          tenantSvc,
		engine:             engine,
	}
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
// If the graphql.FormationObjectType is graphql.FormationObjectTypeApplication it adds the provided formation to the
// scenario label of the application.
// If the graphql.FormationObjectType is graphql.FormationObjectTypeRuntime it adds the provided formation to the
// scenario label of the runtime.
// If the graphql.FormationObjectType is graphql.FormationObjectTypeTenant it will
// create automatic scenario assignment with the caller and target tenant.
func (s *service) AssignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error) {
	switch objectType {
	case graphql.FormationObjectTypeApplication, graphql.FormationObjectTypeRuntime:
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
		if _, err := s.asaService.Create(ctx, newAutomaticScenarioAssignmentModel(formation.Name, tnt, objectID)); err != nil {
			return nil, err
		}
		return &formation, nil
	default:
		return nil, fmt.Errorf("unknown formation type %s", objectType)
	}
}

// UnassignFormation unassigns object base on graphql.FormationObjectType.
// If the graphql.FormationObjectType is graphql.FormationObjectTypeApplication it removes the provided formation from the
// scenario label of the application.
// If the graphql.FormationObjectType is graphql.FormationObjectTypeRuntime and the provided formation is not coming from ASA,
// it removes the formation from the scenario label of the runtime,
// but if the provided formation is assigned from ASA it does nothing.
// If the graphql.FormationObjectType is graphql.FormationObjectTypeTenant it will
// delete the automatic scenario assignment with the caller and target tenant.
func (s *service) UnassignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error) {
	switch objectType {
	case graphql.FormationObjectTypeApplication:
		return s.modifyAssignedFormations(ctx, tnt, objectID, formation, objectTypeToLabelableObject(objectType), deleteFormation)
	case graphql.FormationObjectTypeRuntime:
		if isFormationComingFromASA, err := s.isFormationComingFromASA(ctx, objectID, formation.Name); err != nil {
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
		if err = s.asaService.Delete(ctx, asa); err != nil {
			return nil, err
		}
		return &formation, nil
	default:
		return nil, fmt.Errorf("unknown formation type %s", objectType)
	}
}

func (s *service) isFormationComingFromASA(ctx context.Context, runtimeID, formation string) (bool, error) {
	formationsFromASA, err := s.engine.GetScenariosFromMatchingASAs(ctx, runtimeID)
	if err != nil {
		return false, errors.Wrapf(err, "while getting formations from ASAs for runtime with id: %q", runtimeID)
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
	}
	return labelableObj
}
