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

// LabelConverter missing godoc
//go:generate mockery --name=LabelConverter --output=automock --outpkg=automock --case=underscore
type LabelConverter interface {
	ToEntity(in model.Label) (label.Entity, error)
	FromEntity(in label.Entity) (model.Label, error)
}

// LabelDefRepository missing godoc
//go:generate mockery --name=LabelDefRepository --output=automock --outpkg=automock --case=underscore
type LabelDefRepository interface {
	Create(ctx context.Context, def model.LabelDefinition) error
	Exists(ctx context.Context, tenant string, key string) (bool, error)
	GetByKey(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error)
	UpdateWithVersion(ctx context.Context, def model.LabelDefinition) error
}

// LabelService missing godoc
//go:generate mockery --name=LabelService --output=automock --outpkg=automock --case=underscore
type LabelService interface {
	CreateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error
	UpdateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error
	GetLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) (*model.Label, error)
}

// UIDService missing godoc
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore
type UIDService interface {
	Generate() string
}

// LabelDefService missing godoc
//go:generate mockery --name=LabelDefService --output=automock --outpkg=automock --case=underscore
type LabelDefService interface {
	CreateWithFormations(ctx context.Context, tnt string, formations []string) error
	ValidateExistingLabelsAgainstSchema(ctx context.Context, schema interface{}, tenant, key string) error
	ValidateAutomaticScenarioAssignmentAgainstSchema(ctx context.Context, schema interface{}, tenantID, key string) error
}

type ModificationFunc func([]string, string) []string

type service struct {
	labelConverter     LabelConverter
	labelDefRepository LabelDefRepository
	labelService       LabelService
	labelDefService    LabelDefService
	uuidService        UIDService
}

func NewService(labelConverter LabelConverter, labelDefRepository LabelDefRepository, labelService LabelService, uuidService UIDService, labelDefService LabelDefService) *service {
	return &service{
		labelConverter:     labelConverter,
		labelDefRepository: labelDefRepository,
		labelService:       labelService,
		labelDefService:    labelDefService,
		uuidService:        uuidService,
	}
}

func (s *service) CreateFormation(ctx context.Context, tnt string, formation model.Formation) (*model.Formation, error) {
	return s.modifyFormations(ctx, tnt, formation.Name, addFormation)
}

func (s *service) DeleteFormation(ctx context.Context, tnt string, formation model.Formation) (*model.Formation, error) {
	return s.modifyFormations(ctx, tnt, formation.Name, deleteFormation)
}

func (s *service) AssignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error) {

	switch objectType {
	case graphql.FormationObjectTypeApplication:
		return s.assignApplication(ctx, tnt, objectID, formation, addFormation)
	case graphql.FormationObjectTypeTenant:

	default:
		return nil, fmt.Errorf("unknown formation type %s", objectType)
	}
	panic("")
}

func (s *service) modifyFormations(ctx context.Context, tnt, formationName string, modificationFunc ModificationFunc) (*model.Formation, error) {
	def, err := s.labelDefRepository.GetByKey(ctx, tnt, model.ScenariosKey)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			if err = s.labelDefService.CreateWithFormations(ctx, tnt, []string{formationName}); err != nil {
				return nil, err
			}
			return &model.Formation{Name: formationName}, nil
		}
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

func (s *service) assignApplication(ctx context.Context, tnt, objectID string, formation model.Formation, modificationFunc ModificationFunc) (*model.Formation, error) {
	labelInput := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{formation.Name},
		ObjectID:   objectID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}

	existingLabel, err := s.labelService.GetLabel(ctx, tnt, labelInput)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			if err = s.labelService.CreateLabel(ctx, tnt, s.uuidService.Generate(), labelInput); err != nil {
				return nil, err
			}
			return &formation, nil
		}
		return nil, err
	}

	existingFormations, err := label.ValueToStringsSlice(existingLabel.Value)
	if err != nil {
		return nil, err
	}

	labelInput.Value = modificationFunc(existingFormations, formation.Name)
	labelInput.Version = existingLabel.Version
	return &formation, s.labelService.UpdateLabel(ctx, tnt, existingLabel.ID, labelInput)
}

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
