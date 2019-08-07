package labeldef

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

type Resolver struct {
	conv          Converter
	srv           Service
	transactioner persistence.Transactioner
}

func NewResolver(srv Service, conv Converter, transactioner persistence.Transactioner) *Resolver {
	return &Resolver{
		conv:          conv,
		srv:           srv,
		transactioner: transactioner,
	}
}

// dependencies
//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	FromGraphQL(input graphql.LabelDefinitionInput, tenant string) model.LabelDefinition
	ToGraphQL(definition model.LabelDefinition) graphql.LabelDefinition
	ToEntity(in model.LabelDefinition) (Entity, error)
	FromEntity(in Entity) (model.LabelDefinition, error)
}

//go:generate mockery -name=Service -output=automock -outpkg=automock -case=underscore
type Service interface {
	Create(ctx context.Context, ld model.LabelDefinition) (model.LabelDefinition, error)
	Get(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error)
	List(ctx context.Context, tenant string) ([]model.LabelDefinition, error)
	Delete(ctx context.Context, tenant string, key string, deleteRelatedLabels bool) error
	Update(ctx context.Context, ld model.LabelDefinition) error
}

func (r *Resolver) CreateLabelDefinition(ctx context.Context, in graphql.LabelDefinitionInput) (*graphql.LabelDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transactioner.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transactioner.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	// TODO: Use LabelDefinitionInput
	ld := r.conv.FromGraphQL(in, tnt)
	createdLd, err := r.srv.Create(ctx, ld)
	if err != nil {
		return nil, errors.Wrap(err, "while creating label definition")
	}
	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}
	out := r.conv.ToGraphQL(createdLd)

	return &out, nil
}

func (r *Resolver) LabelDefinitions(ctx context.Context) ([]*graphql.LabelDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}
	tx, err := r.transactioner.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transactioner.RollbackUnlessCommited(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	defs, err := r.srv.List(ctx, tnt)
	if err != nil {
		return nil, errors.Wrap(err, "while listing Label Definitions")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	var out []*graphql.LabelDefinition
	for _, def := range defs {
		c := r.conv.ToGraphQL(def)
		out = append(out, &c)
	}
	return out, nil
}

func (r *Resolver) LabelDefinition(ctx context.Context, key string) (*graphql.LabelDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transactioner.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transactioner.RollbackUnlessCommited(tx)
	ctx = persistence.SaveToContext(ctx, tx)
	def, err := r.srv.Get(ctx, tnt, key)

	if err != nil {
		return nil, errors.Wrap(err, "while getting Label Definition")
	}
	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	if def == nil {
		return nil, fmt.Errorf("label definition with key '%s' does not exist", key)
	}
	c := r.conv.ToGraphQL(*def)
	return &c, nil
}

func (r *Resolver) UpdateLabelDefinition(ctx context.Context, in graphql.LabelDefinitionInput) (*graphql.LabelDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transactioner.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transactioner.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	// TODO: Use LabelDefinitionInput
	ld := r.conv.FromGraphQL(in, tnt)

	err = r.srv.Update(ctx, ld)
	if err != nil {
		return nil, errors.Wrap(err, "while updating label definition")
	}

	updatedLd, err := r.srv.Get(ctx, tnt, in.Key)
	if err != nil {
		return nil, errors.Wrap(err, "while receiving updated label definition")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	out := r.conv.ToGraphQL(*updatedLd)

	return &out, nil
}

func (r *Resolver) DeleteLabelDefinition(ctx context.Context, key string, deleteRelatedLabels *bool) (*graphql.LabelDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transactioner.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transactioner.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if deleteRelatedLabels == nil {
		return nil, errors.New("deleteRelatedLabels can not be nil, internal server error")
	}

	ld, err := r.srv.Get(ctx, tnt, key)
	if err != nil {
		return nil, err
	}
	if ld == nil {
		return nil, fmt.Errorf("Label Definition with key %s not found", key)
	}

	deletedLD := r.conv.ToGraphQL(*ld)

	err = r.srv.Delete(ctx, tnt, key, *deleteRelatedLabels)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return &deletedLD, nil
}
