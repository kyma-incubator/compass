package runtimectx

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// RuntimeContextService missing godoc
//go:generate mockery --name=RuntimeContextService --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeContextService interface {
	Create(ctx context.Context, in model.RuntimeContextInput) (string, error)
	Update(ctx context.Context, id string, in model.RuntimeContextInput) error
	GetByID(ctx context.Context, id string) (*model.RuntimeContext, error)
	Delete(ctx context.Context, id string) error
	ListLabels(ctx context.Context, runtimeID string) (map[string]*model.Label, error)
}

// RuntimeContextConverter missing godoc
//go:generate mockery --name=RuntimeContextConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeContextConverter interface {
	ToGraphQL(in *model.RuntimeContext) *graphql.RuntimeContext
	InputFromGraphQL(in graphql.RuntimeContextInput) model.RuntimeContextInput
	InputFromGraphQLWithRuntimeID(in graphql.RuntimeContextInput, runtimeID string) model.RuntimeContextInput
}

// Resolver missing godoc
type Resolver struct {
	transact              persistence.Transactioner
	runtimeContextService RuntimeContextService
	converter             RuntimeContextConverter
}

// NewResolver missing godoc
func NewResolver(transact persistence.Transactioner, runtimeContextService RuntimeContextService, conv RuntimeContextConverter) *Resolver {
	return &Resolver{
		transact:              transact,
		runtimeContextService: runtimeContextService,
		converter:             conv,
	}
}

// RegisterRuntimeContext registers RuntimeContext from `in` and associates it with Runtime with ID `runtimeID`
func (r *Resolver) RegisterRuntimeContext(ctx context.Context, runtimeID string, in graphql.RuntimeContextInput) (*graphql.RuntimeContext, error) {
	convertedIn := r.converter.InputFromGraphQLWithRuntimeID(in, runtimeID)

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	id, err := r.runtimeContextService.Create(ctx, convertedIn)
	if err != nil {
		return nil, err
	}

	runtimeContext, err := r.runtimeContextService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	gqlRuntimeContext := r.converter.ToGraphQL(runtimeContext)

	return gqlRuntimeContext, nil
}

// UpdateRuntimeContext updates RuntimeContext with ID `id` using `in`
func (r *Resolver) UpdateRuntimeContext(ctx context.Context, id string, in graphql.RuntimeContextInput) (*graphql.RuntimeContext, error) {
	convertedIn := r.converter.InputFromGraphQL(in)

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err = r.runtimeContextService.Update(ctx, id, convertedIn); err != nil {
		return nil, err
	}

	runtimeContext, err := r.runtimeContextService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	gqlRuntimeContext := r.converter.ToGraphQL(runtimeContext)

	return gqlRuntimeContext, nil
}

// DeleteRuntimeContext deletes RuntimeContext with ID `id`
func (r *Resolver) DeleteRuntimeContext(ctx context.Context, id string) (*graphql.RuntimeContext, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	runtimeContext, err := r.runtimeContextService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = r.runtimeContextService.Delete(ctx, id); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	deletedRuntimeContext := r.converter.ToGraphQL(runtimeContext)

	return deletedRuntimeContext, nil
}

// Labels lists Labels with key `key`for RuntimeContext `obj`
func (r *Resolver) Labels(ctx context.Context, obj *graphql.RuntimeContext, key *string) (graphql.Labels, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Runtime Context cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	itemMap, err := r.runtimeContextService.ListLabels(ctx, obj.ID)
	if err != nil {
		if strings.Contains(err.Error(), "doesn't exist") {
			return nil, tx.Commit()
		}
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	resultLabels := make(map[string]interface{})

	for _, label := range itemMap {
		if key == nil || label.Key == *key {
			resultLabels[label.Key] = label.Value
		}
	}

	return resultLabels, nil
}
