package formationconstraint

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

// FormationConstraintRepository represents the Formation Constraint repository layer
//go:generate mockery --name=FormationConstraintRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationConstraintRepository interface {
	ListMatchingFormationConstraints(ctx context.Context, formationConstraintIDs []string, location JoinPointLocation, details MatchingDetails) ([]*model.FormationConstraint, error)
}

// FormationTemplateRepository represents the FormationTemplate repository layer
//go:generate mockery --name=FormationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationTemplateRepository interface {
	Get(ctx context.Context, id string) (*model.FormationTemplate, error)
}

// FormationTemplateConstraintReferenceRepository converts between the internal model and entity
//go:generate mockery --name=EntityConveFormationTemplateConstraintReferenceRepositoryrter --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationTemplateConstraintReferenceRepository interface {
	ListByFormationTemplateID(ctx context.Context, formationTemplateID string) ([]*model.FormationTemplateConstraintReference, error)
}

// UIDService generates UUIDs for new entities
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	repo                                     FormationConstraintRepository
	formationTemplateRepo                    FormationTemplateRepository
	formationTemplateConstraintReferenceRepo FormationTemplateConstraintReferenceRepository
	uidSvc                                   UIDService
}

// NewService creates a FormationConstraint service
func NewService(repo FormationConstraintRepository, formationTemplateRepo FormationTemplateRepository, formationTemplateConstraintReferenceRepo FormationTemplateConstraintReferenceRepository, uidSvc UIDService) *service {
	return &service{
		repo:                                     repo,
		formationTemplateRepo:                    formationTemplateRepo,
		formationTemplateConstraintReferenceRepo: formationTemplateConstraintReferenceRepo,
		uidSvc:                                   uidSvc,
	}
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
