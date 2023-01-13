package formationconstraint

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// FormationConstraintRepository represents the Formation Constraint repository layer
//go:generate mockery --name=FormationConstraintRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationConstraintRepository interface {
	Create(ctx context.Context, item *model.FormationConstraint) error
	Get(ctx context.Context, id string) (*model.FormationConstraint, error)
	ListAll(ctx context.Context) ([]*model.FormationConstraint, error)
	ListByIDs(ctx context.Context, formationConstraintIDs []string) ([]*model.FormationConstraint, error)
	Delete(ctx context.Context, id string) error
	ListMatchingFormationConstraints(ctx context.Context, formationConstraintIDs []string, location JoinPointLocation, details MatchingDetails) ([]*model.FormationConstraint, error)
}

// FormationTemplateConstraintReferenceRepository converts between the internal model and entity
//go:generate mockery --name=EntityConveFormationTemplateConstraintReferenceRepositoryrter --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationTemplateConstraintReferenceRepository interface {
	Create(ctx context.Context, item *model.FormationTemplateConstraintReference) error
	ListByFormationTemplateID(ctx context.Context, formationTemplateID string) ([]*model.FormationTemplateConstraintReference, error)
}

// UIDService generates UUIDs for new entities
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	repo                                     FormationConstraintRepository
	formationTemplateConstraintReferenceRepo FormationTemplateConstraintReferenceRepository
	converter                                FormationConstraintConverter
	uidSvc                                   UIDService
}

// NewService creates a FormationConstraint service
func NewService(repo FormationConstraintRepository, formationTemplateConstraintReferenceRepo FormationTemplateConstraintReferenceRepository, uidSvc UIDService) *service {
	return &service{
		repo:                                     repo,
		formationTemplateConstraintReferenceRepo: formationTemplateConstraintReferenceRepo,
		uidSvc:                                   uidSvc,
	}
}

func (s *service) Create(ctx context.Context, in *model.FormationConstraintInput) (string, error) {
	formationConstraintID := s.uidSvc.Generate()

	log.C(ctx).Debugf("ID %s generated for Formation Constraint with name %s", formationConstraintID, in.Name)

	err := s.repo.Create(ctx, s.converter.FromModelInputToModel(in, formationConstraintID))
	if err != nil {
		return "", errors.Wrapf(err, "while creating Formation Constraint with name %s", in.Name)
	}

	constraintReference := &model.FormationTemplateConstraintReference{
		Constraint:        formationConstraintID,
		FormationTemplate: in.FormationTemplateID,
	}
	if err = s.formationTemplateConstraintReferenceRepo.Create(ctx, constraintReference); err != nil {
		return "", errors.Wrapf(err, "while creting Reference for Formation Template and Formation Constraint")
	}

	return formationConstraintID, nil
}

func (s *service) Get(ctx context.Context, id string) (*model.FormationConstraint, error) {
	formationConstraint, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Formation Constraint with id %s", id)
	}

	return formationConstraint, nil
}

func (s *service) List(ctx context.Context) ([]*model.FormationConstraint, error) {
	formationConstraints, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing all Formation Constraints")
	}
	return formationConstraints, nil
}

func (s *service) ListByFormationTemplateID(ctx context.Context, formationTemplateID string) ([]*model.FormationConstraint, error) {
	constraintReferences, err := s.formationTemplateConstraintReferenceRepo.ListByFormationTemplateID(ctx, formationTemplateID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing Formation Constraint References for FormationTemplate with ID: %s", formationTemplateID)
	}

	formationConstraintIDs := make([]string, 0, len(constraintReferences))
	for _, cr := range constraintReferences {
		formationConstraintIDs = append(formationConstraintIDs, cr.Constraint)
	}

	formationConstraints, err := s.repo.ListByIDs(ctx, formationConstraintIDs)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing Formation Constraints for FormationTemplate with ID: %s", formationTemplateID)
	}
	return formationConstraints, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "while deleting Formation Constraint with ID %s", id)
	}

	return nil
}

func (s *service) ListMatchingConstraints(ctx context.Context, formationTemplateID string, location JoinPointLocation, details MatchingDetails) ([]*model.FormationConstraint, error) {
	formationTemplateConstraintReferences, err := s.formationTemplateConstraintReferenceRepo.ListByFormationTemplateID(ctx, formationTemplateID)
	if err != nil {
		return nil, err
	}

	constraintIDs := make([]string, 0, len(formationTemplateConstraintReferences))
	for _, reference := range formationTemplateConstraintReferences {
		constraintIDs = append(constraintIDs, reference.Constraint)
	}

	constraints, err := s.repo.ListMatchingFormationConstraints(ctx, constraintIDs, location, details)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing matching formation constraints for formation template with ID %q, target operation %q, constraint type %q, resource type %q and resource subtype %q", formationTemplateID, location.OperationName, location.ConstraintType, details.resourceType, details.resourceSubtype)
	}

	return constraints, nil
}
