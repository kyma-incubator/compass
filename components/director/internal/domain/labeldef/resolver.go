package labeldef

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// Resolver missing godoc
type Resolver struct {
	conv          ModelConverter
	srv           Service
	transactioner persistence.Transactioner
}

// NewResolver missing godoc
func NewResolver(transactioner persistence.Transactioner, srv Service, conv ModelConverter) *Resolver {
	return &Resolver{
		conv:          conv,
		srv:           srv,
		transactioner: transactioner,
	}
}

// ModelConverter missing godoc
//go:generate mockery --name=ModelConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type ModelConverter interface {
	// TODO: Use model.LabelDefinitionInput
	FromGraphQL(input graphql.LabelDefinitionInput, tenant string) (model.LabelDefinition, error)
	ToGraphQL(definition model.LabelDefinition) (graphql.LabelDefinition, error)
}

// Service missing godoc
//go:generate mockery --name=Service --output=automock --outpkg=automock --case=underscore --disable-version-string
type Service interface {
	Create(ctx context.Context, ld model.LabelDefinition) (model.LabelDefinition, error)
	Get(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error)
	List(ctx context.Context, tenant string) ([]model.LabelDefinition, error)
	Delete(ctx context.Context, tenant string, key string, deleteRelatedLabels bool) error
	Update(ctx context.Context, ld model.LabelDefinition) error
}

// CreateLabelDefinition missing godoc
func (r *Resolver) CreateLabelDefinition(ctx context.Context, in graphql.LabelDefinitionInput) (*graphql.LabelDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transactioner.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transactioner.RollbackUnlessCommitted(ctx, tx)

	ld, err := r.conv.FromGraphQL(in, tnt)
	if err != nil {
		return nil, err
	}

	ctx = persistence.SaveToContext(ctx, tx)

	createdLd, err := r.srv.Create(ctx, ld)
	if err != nil {
		return nil, errors.Wrap(err, "while creating label definition")
	}

	out, err := r.conv.ToGraphQL(createdLd)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return &out, nil
}

// LabelDefinitions missing godoc
func (r *Resolver) LabelDefinitions(ctx context.Context) ([]*graphql.LabelDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transactioner.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transactioner.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	defs, err := r.srv.List(ctx, tnt)
	if err != nil {
		return nil, errors.Wrap(err, "while listing Label Definitions")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	out := make([]*graphql.LabelDefinition, 0, len(defs))
	for _, def := range defs {
		c, err := r.conv.ToGraphQL(def)
		if err != nil {
			return nil, err
		}

		out = append(out, &c)
	}
	return out, nil
}

// LabelDefinition missing godoc
func (r *Resolver) LabelDefinition(ctx context.Context, key string) (*graphql.LabelDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transactioner.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transactioner.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)
	def, err := r.srv.Get(ctx, tnt, key)

	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, errors.Wrap(err, "while getting Label Definition")
	}
	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}
	if def == nil {
		return nil, apperrors.NewNotFoundError(resource.LabelDefinition, key)
	}
	c, err := r.conv.ToGraphQL(*def)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateLabelDefinition missing godoc
func (r *Resolver) UpdateLabelDefinition(ctx context.Context, in graphql.LabelDefinitionInput) (*graphql.LabelDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transactioner.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transactioner.RollbackUnlessCommitted(ctx, tx)

	ld, err := r.conv.FromGraphQL(in, tnt)
	if err != nil {
		return nil, err
	}

	ctx = persistence.SaveToContext(ctx, tx)

	err = r.srv.Update(ctx, ld)
	if err != nil {
		return nil, errors.Wrap(err, "while updating label definition")
	}

	updatedLd, err := r.srv.Get(ctx, tnt, in.Key)
	if err != nil {
		return nil, errors.Wrap(err, "while receiving updated label definition")
	}

	out, err := r.conv.ToGraphQL(*updatedLd)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return &out, nil
}

// DeleteLabelDefinition missing godoc
func (r *Resolver) DeleteLabelDefinition(ctx context.Context, key string, deleteRelatedLabels *bool) (*graphql.LabelDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transactioner.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transactioner.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if deleteRelatedLabels == nil {
		return nil, apperrors.NewInternalError("deleteRelatedLabels can not be nil")
	}

	ld, err := r.srv.Get(ctx, tnt, key)
	if err != nil {
		return nil, err
	}
	if ld == nil {
		return nil, fmt.Errorf("labelDefinition with key %s not found", key)
	}

	err = r.srv.Delete(ctx, tnt, key, *deleteRelatedLabels)
	if err != nil {
		return nil, err
	}

	deletedLD, err := r.conv.ToGraphQL(*ld)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return &deletedLD, nil
}
