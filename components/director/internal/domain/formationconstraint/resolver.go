package formationconstraint

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

//go:generate mockery --exported --name=formationTemplateConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationConstraintConverter interface {
	FromInputGraphQL(in *graphql.FormationConstraintInput) *model.FormationConstraintInput
	ToGraphQL(in *model.FormationConstraint) *graphql.FormationConstraint
	MultipleToGraphQL(in []*model.FormationConstraint) []*graphql.FormationConstraint
	FromModelInputToModel(in *model.FormationConstraintInput, id string) *model.FormationConstraint
}

//go:generate mockery --exported --name=formationTemplateService --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationConstraintService interface {
	Create(ctx context.Context, in *model.FormationConstraintInput) (string, error)
	Get(ctx context.Context, id string) (*model.FormationConstraint, error)
	List(ctx context.Context) ([]*model.FormationConstraint, error)
	ListByFormationTemplateID(ctx context.Context, formationTemplateID string) ([]*model.FormationConstraint, error)
	Delete(ctx context.Context, id string) error
}

// Resolver is the FormationConstraint resolver
type Resolver struct {
	transact persistence.Transactioner

	svc       formationConstraintService
	converter formationConstraintConverter
}

// NewResolver creates FormationConstraint resolver
func NewResolver(transact persistence.Transactioner, converter formationConstraintConverter, svc formationConstraintService) *Resolver {
	return &Resolver{
		transact:  transact,
		converter: converter,
		svc:       svc,
	}
}

// FormationConstraints lists all FormationConstraints
func (r *Resolver) FormationConstraints(ctx context.Context) ([]*graphql.FormationConstraint, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formationConstraints, err := r.svc.List(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.converter.MultipleToGraphQL(formationConstraints), nil
}

// FormationConstraintsByFormationType lists all FormationConstraints for the specified FormationTemplate
func (r *Resolver) FormationConstraintsByFormationType(ctx context.Context, formationTemplateID string) ([]*graphql.FormationConstraint, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formationConstraints, err := r.svc.ListByFormationTemplateID(ctx, formationTemplateID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.converter.MultipleToGraphQL(formationConstraints), nil
}

// FormationConstraint queries the FormationConstraint matching ID `id`
func (r *Resolver) FormationConstraint(ctx context.Context, id string) (*graphql.FormationConstraint, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formationConstraint, err := r.svc.Get(ctx, id)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.converter.ToGraphQL(formationConstraint), nil
}

// CreateFormationConstraint creates a FormationConstraint using `in`
func (r *Resolver) CreateFormationConstraint(ctx context.Context, in graphql.FormationConstraintInput) (*graphql.FormationConstraint, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err = in.Validate(); err != nil {
		return nil, err
	}

	id, err := r.svc.Create(ctx, r.converter.FromInputGraphQL(&in))
	if err != nil {
		return nil, err
	}
	log.C(ctx).Infof("Successfully created an Formation Constraint with name %s and id %s", in.Name, id)

	formationConstraint, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.converter.ToGraphQL(formationConstraint), nil
}

// DeleteFormationConstraint deletes the FormationConstraint matching ID `id`
func (r *Resolver) DeleteFormationConstraint(ctx context.Context, id string) (*graphql.FormationConstraint, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formationConstraint, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = r.svc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.converter.ToGraphQL(formationConstraint), nil
}
