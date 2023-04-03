package formationconstraint

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=formationConstraintRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationConstraintRepository interface {
	Create(ctx context.Context, item *model.FormationConstraint) error
	Get(ctx context.Context, id string) (*model.FormationConstraint, error)
	ListAll(ctx context.Context) ([]*model.FormationConstraint, error)
	ListByIDs(ctx context.Context, formationConstraintIDs []string) ([]*model.FormationConstraint, error)
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, model *model.FormationConstraint) error
	ListMatchingFormationConstraints(ctx context.Context, formationConstraintIDs []string, location formationconstraint.JoinPointLocation, details formationconstraint.MatchingDetails) ([]*model.FormationConstraint, error)
	ListByIDsAndGlobal(ctx context.Context, formationConstraintIDs []string) ([]*model.FormationConstraint, error)
}

//go:generate mockery --exported --name=formationTemplateConstraintReferenceRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationTemplateConstraintReferenceRepository interface {
	ListByFormationTemplateID(ctx context.Context, formationTemplateID string) ([]*model.FormationTemplateConstraintReference, error)
	ListByFormationTemplateIDs(ctx context.Context, formationTemplateIDs []string) ([]*model.FormationTemplateConstraintReference, error)
}

//go:generate mockery --exported --name=uidService --output=automock --outpkg=automock --case=underscore --disable-version-string
type uidService interface {
	Generate() string
}

type service struct {
	repo                                     formationConstraintRepository
	formationTemplateConstraintReferenceRepo formationTemplateConstraintReferenceRepository
	converter                                formationConstraintConverter
	uidSvc                                   uidService
}

// NewService creates a FormationConstraint service
func NewService(repo formationConstraintRepository, formationTemplateConstraintReferenceRepo formationTemplateConstraintReferenceRepository, uidSvc uidService, converter formationConstraintConverter) *service {
	return &service{
		repo:                                     repo,
		formationTemplateConstraintReferenceRepo: formationTemplateConstraintReferenceRepo,
		uidSvc:                                   uidSvc,
		converter:                                converter,
	}
}

// Create creates formation constraint using the provided input
func (s *service) Create(ctx context.Context, in *model.FormationConstraintInput) (string, error) {
	formationConstraintID := s.uidSvc.Generate()

	log.C(ctx).Debugf("ID %s generated for Formation Constraint with name %s", formationConstraintID, in.Name)

	err := s.repo.Create(ctx, s.converter.FromModelInputToModel(in, formationConstraintID))
	if err != nil {
		return "", errors.Wrapf(err, "while creating Formation Constraint with name %s", in.Name)
	}

	return formationConstraintID, nil
}

// Get fetches formation constraint by id
func (s *service) Get(ctx context.Context, id string) (*model.FormationConstraint, error) {
	formationConstraint, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Formation Constraint with id %s", id)
	}

	return formationConstraint, nil
}

// List lists all formation constraints
func (s *service) List(ctx context.Context) ([]*model.FormationConstraint, error) {
	formationConstraints, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing all Formation Constraints")
	}
	return formationConstraints, nil
}

// ListByFormationTemplateID lists all formation constraints associated with the formation template
func (s *service) ListByFormationTemplateID(ctx context.Context, formationTemplateID string) ([]*model.FormationConstraint, error) {
	constraintReferences, err := s.formationTemplateConstraintReferenceRepo.ListByFormationTemplateID(ctx, formationTemplateID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing Formation Constraint References for FormationTemplate with ID: %s", formationTemplateID)
	}

	formationConstraintIDs := make([]string, 0, len(constraintReferences))
	for _, cr := range constraintReferences {
		formationConstraintIDs = append(formationConstraintIDs, cr.ConstraintID)
	}

	formationConstraints, err := s.repo.ListByIDs(ctx, formationConstraintIDs)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing Formation Constraints for FormationTemplate with ID: %s", formationTemplateID)
	}
	return formationConstraints, nil
}

// ListByFormationTemplateIDs lists all formation constraints associated with the formation templates
func (s *service) ListByFormationTemplateIDs(ctx context.Context, formationTemplateIDs []string) ([][]*model.FormationConstraint, error) {
	constraintRefs, err := s.formationTemplateConstraintReferenceRepo.ListByFormationTemplateIDs(ctx, formationTemplateIDs)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing Formation Constraint References for FormationTemplates with IDs: %q", formationTemplateIDs)
	}

	constraintIDs := make([]string, 0, len(constraintRefs))
	formationTemplateIDToConstraintIDsMap := make(map[string][]string, len(constraintRefs))

	for _, cr := range constraintRefs {
		constraintIDs = append(constraintIDs, cr.ConstraintID)

		formationTemplateIDToConstraintIDsMap[cr.FormationTemplateID] = append(formationTemplateIDToConstraintIDsMap[cr.FormationTemplateID], cr.ConstraintID)
	}

	formationConstraints, err := s.repo.ListByIDsAndGlobal(ctx, constraintIDs)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing Formation Constraints by IDs %q and the global ones", formationTemplateIDs)
	}

	globalConstraints := make([]*model.FormationConstraint, 0, len(formationConstraints))
	attachedConstraintsMap := make(map[string]*model.FormationConstraint, len(formationConstraints))

	for _, constraint := range formationConstraints {
		if constraint.ConstraintScope == model.GlobalFormationConstraintScope {
			globalConstraints = append(globalConstraints, constraint)
		} else {
			attachedConstraintsMap[constraint.ID] = constraint
		}
	}

	formationConstraintsPerFormationTemplate := make([][]*model.FormationConstraint, len(formationTemplateIDs))

	for i, ftID := range formationTemplateIDs {
		// Add attached constraints
		constraintsPerFormation := formationTemplateIDToConstraintIDsMap[ftID]
		for _, constraintID := range constraintsPerFormation {
			formationConstraintsPerFormationTemplate[i] = append(formationConstraintsPerFormationTemplate[i], attachedConstraintsMap[constraintID])
		}

		// Add global constraints
		formationConstraintsPerFormationTemplate[i] = append(formationConstraintsPerFormationTemplate[i], globalConstraints...)
	}

	return formationConstraintsPerFormationTemplate, nil
}

// Delete deletes formation constraint by id
func (s *service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "while deleting Formation Constraint with ID %s", id)
	}

	return nil
}

// ListMatchingConstraints lists formation constraints that math the provided JoinPointLocation and JoinPointDetails
func (s *service) ListMatchingConstraints(ctx context.Context, formationTemplateID string, location formationconstraint.JoinPointLocation, details formationconstraint.MatchingDetails) ([]*model.FormationConstraint, error) {
	formationTemplateConstraintReferences, err := s.formationTemplateConstraintReferenceRepo.ListByFormationTemplateID(ctx, formationTemplateID)
	if err != nil {
		return nil, err
	}

	constraintIDs := make([]string, 0, len(formationTemplateConstraintReferences))
	for _, reference := range formationTemplateConstraintReferences {
		constraintIDs = append(constraintIDs, reference.ConstraintID)
	}

	constraints, err := s.repo.ListMatchingFormationConstraints(ctx, constraintIDs, location, details)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing matching formation constraints for formation template with ID %q, target operation %q, constraint type %q, resource type %q and resource subtype %q", formationTemplateID, location.OperationName, location.ConstraintType, details.ResourceType, details.ResourceSubtype)
	}

	return constraints, nil
}

// Update updates a FormationConstraint matching ID `id` using `in`
func (s *service) Update(ctx context.Context, id string, in *model.FormationConstraintInput) error {
	err := s.repo.Update(ctx, s.converter.FromModelInputToModel(in, id))
	if err != nil {
		return errors.Wrapf(err, "while updating Formation Constraint with ID %s", id)
	}

	return nil
}
