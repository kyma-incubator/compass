package formationtemplateconstraintreferences

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

//go:generate mockery --exported --name=constraintReferenceConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type constraintReferenceConverter interface {
	ToModel(in *graphql.ConstraintReference) *model.FormationTemplateConstraintReference
	ToGraphql(in *model.FormationTemplateConstraintReference) *graphql.ConstraintReference
	ToEntity(in *model.FormationTemplateConstraintReference) *Entity
	FromEntity(e *Entity) *model.FormationTemplateConstraintReference
}

//go:generate mockery --exported --name=constraintReferenceService --output=automock --outpkg=automock --case=underscore --disable-version-string
type constraintReferenceService interface {
	Create(ctx context.Context, in *model.FormationTemplateConstraintReference) error
	Delete(ctx context.Context, constraintID, formationTemplateID string) error
}

// Resolver is the FormationConstraint resolver
type Resolver struct {
	transact persistence.Transactioner

	svc       constraintReferenceService
	converter constraintReferenceConverter
}

// NewResolver creates FormationConstraint resolver
func NewResolver(transact persistence.Transactioner, converter constraintReferenceConverter, svc constraintReferenceService) *Resolver {
	return &Resolver{
		transact:  transact,
		converter: converter,
		svc:       svc,
	}
}

// AttachConstraintToFormationTemplate creates a FormationTemplateConstraintReference using `in`
func (r *Resolver) AttachConstraintToFormationTemplate(ctx context.Context, constraintID, formationTemplateID string) (*graphql.ConstraintReference, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	in := &graphql.ConstraintReference{
		ConstraintID:        constraintID,
		FormationTemplateID: formationTemplateID,
	}
	err = r.svc.Create(ctx, r.converter.ToModel(in))
	if err != nil {
		return nil, err
	}
	log.C(ctx).Infof("Successfully created an Formation Template Constraint Reference for Constraint with ID %q and Formation Template with ID %q", in.ConstraintID, in.FormationTemplateID)

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return in, nil
}

// DetachConstraintFromFormationTemplate deletes the FormationTemplateConstraintReference matching ID `id`
func (r *Resolver) DetachConstraintFromFormationTemplate(ctx context.Context, constraintID, formationTemplateID string) (*graphql.ConstraintReference, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	err = r.svc.Delete(ctx, constraintID, formationTemplateID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &graphql.ConstraintReference{ConstraintID: constraintID, FormationTemplateID: formationTemplateColumn}, nil
}
