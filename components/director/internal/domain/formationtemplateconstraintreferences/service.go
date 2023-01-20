package formationtemplateconstraintreferences

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=formationTemplateConstraintReferenceRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationTemplateConstraintReferenceRepository interface {
	Create(ctx context.Context, item *model.FormationTemplateConstraintReference) error
	Delete(ctx context.Context, formationTemplateID, constraintID string) error
}

type service struct {
	repo      formationTemplateConstraintReferenceRepository
	converter constraintReferenceConverter
}

// NewService creates a FormationTemplateConstraintReference service
func NewService(repo formationTemplateConstraintReferenceRepository, converter constraintReferenceConverter) *service {
	return &service{
		repo:      repo,
		converter: converter,
	}
}

// Create creates formation template constraint reference using the provided input
func (s *service) Create(ctx context.Context, in *model.FormationTemplateConstraintReference) error {
	log.C(ctx).Infof("Creating an Formation Template Constraint Reference for Constraint with ID %q and Formation Template with ID %q", in.ConstraintID, in.FormationTemplateID)

	err := s.repo.Create(ctx, in)
	if err != nil {
		return errors.Wrapf(err, "while creating Formation Template Constraint Reference for Constraint with ID %q and Formation Template with ID %q", in.ConstraintID, in.FormationTemplateID)
	}

	return nil
}

// Delete deletes formation template constraint reference by constraint ID and formation template ID
func (s *service) Delete(ctx context.Context, constraintID, formationTemplateID string) error {
	if err := s.repo.Delete(ctx, constraintID, formationTemplateID); err != nil {
		return errors.Wrapf(err, "while deleting Formation Template Constraint Reference for Constraint with ID %q and Formation Template with ID %q", constraintID, formationTemplateID)
	}

	return nil
}
