package runtimectx

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// RuntimeContextService missing godoc
//go:generate mockery --name=RuntimeContextService --output=automock --outpkg=automock --case=underscore
type RuntimeContextService interface {
	Create(ctx context.Context, in model.RuntimeContextInput) (string, error)
	Update(ctx context.Context, id string, in model.RuntimeContextInput) error
	Get(ctx context.Context, id string) (*model.RuntimeContext, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, runtimeID string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimeContextPage, error)
	ListLabels(ctx context.Context, runtimeID string) (map[string]*model.Label, error)
}

// RuntimeContextConverter missing godoc
//go:generate mockery --name=RuntimeContextConverter --output=automock --outpkg=automock --case=underscore
type RuntimeContextConverter interface {
	ToGraphQL(in *model.RuntimeContext) *graphql.RuntimeContext
	MultipleToGraphQL(in []*model.RuntimeContext) []*graphql.RuntimeContext
	InputFromGraphQL(in graphql.RuntimeContextInput, runtimeID string) model.RuntimeContextInput
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

// RuntimeContexts missing godoc
func (r *Resolver) RuntimeContexts(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.RuntimeContextPage, error) {
	runtimeID, err := r.getRuntimeID(ctx)
	if err != nil {
		return nil, err
	}

	labelFilter := labelfilter.MultipleFromGraphQL(filter)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter 'first'")
	}

	runtimeContextsPage, err := r.runtimeContextService.List(ctx, runtimeID, labelFilter, *first, cursor)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlRuntimeContexts := r.converter.MultipleToGraphQL(runtimeContextsPage.Data)

	return &graphql.RuntimeContextPage{
		Data:       gqlRuntimeContexts,
		TotalCount: runtimeContextsPage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(runtimeContextsPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(runtimeContextsPage.PageInfo.EndCursor),
			HasNextPage: runtimeContextsPage.PageInfo.HasNextPage,
		},
	}, nil
}

// RuntimeContext missing godoc
func (r *Resolver) RuntimeContext(ctx context.Context, id string) (*graphql.RuntimeContext, error) {
	runtimeID, err := r.getRuntimeID(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	runtimeContext, err := r.runtimeContextService.Get(ctx, id)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	if runtimeID != runtimeContext.RuntimeID {
		log.C(ctx).Errorf("Runtime context owner mismatch: runtime context is owned by runtime with id %s which is different from calling runtime id %s", runtimeContext.RuntimeID, runtimeID)
		return nil, apperrors.NewUnauthorizedError("runtime context not accessible")
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.converter.ToGraphQL(runtimeContext), nil
}

// RegisterRuntimeContext missing godoc
func (r *Resolver) RegisterRuntimeContext(ctx context.Context, in graphql.RuntimeContextInput) (*graphql.RuntimeContext, error) {
	runtimeID, err := r.getRuntimeID(ctx)
	if err != nil {
		return nil, err
	}

	convertedIn := r.converter.InputFromGraphQL(in, runtimeID)

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

	runtimeContext, err := r.runtimeContextService.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlRuntimeContext := r.converter.ToGraphQL(runtimeContext)

	return gqlRuntimeContext, nil
}

// UpdateRuntimeContext missing godoc
func (r *Resolver) UpdateRuntimeContext(ctx context.Context, id string, in graphql.RuntimeContextInput) (*graphql.RuntimeContext, error) {
	runtimeID, err := r.getRuntimeID(ctx)
	if err != nil {
		return nil, err
	}

	convertedIn := r.converter.InputFromGraphQL(in, runtimeID)

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	err = r.runtimeContextService.Update(ctx, id, convertedIn)
	if err != nil {
		return nil, err
	}

	runtimeContext, err := r.runtimeContextService.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if runtimeID != runtimeContext.RuntimeID {
		log.C(ctx).Errorf("Runtime context owner mismatch: runtime context is owned by runtime with id %s which is different from calling runtime id %s", runtimeContext.RuntimeID, runtimeID)
		return nil, apperrors.NewUnauthorizedError("runtime context not accessible")
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlRuntimeContext := r.converter.ToGraphQL(runtimeContext)

	return gqlRuntimeContext, nil
}

// DeleteRuntimeContext missing godoc
func (r *Resolver) DeleteRuntimeContext(ctx context.Context, id string) (*graphql.RuntimeContext, error) {
	runtimeID, err := r.getRuntimeID(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	runtimeContext, err := r.runtimeContextService.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// TODO: Revisit when Runtime Context credentials and ASA are introduced
	/*auths, err := r.sysAuthSvc.ListForObject(ctx, model.RuntimeReference, runtime.ID)
	if err != nil {
		return nil, err
	}

	err = r.oAuth20Svc.DeleteMultipleClientCredentials(ctx, auths)
	if err != nil {
		return nil, err
	}

	err = r.deleteAssociatedScenarioAssignments(ctx, runtime.ID)
	if err != nil {
		return nil, err
	}*/

	if runtimeID != runtimeContext.RuntimeID {
		log.C(ctx).Errorf("Runtime context owner mismatch: runtime context is owned by runtime with id %s which is different from calling runtime id %s", runtimeContext.RuntimeID, runtimeID)
		return nil, apperrors.NewUnauthorizedError("runtime context not accessible")
	}

	deletedRuntimeContext := r.converter.ToGraphQL(runtimeContext)

	err = r.runtimeContextService.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return deletedRuntimeContext, nil
}

// Labels missing godoc
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

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	resultLabels := make(map[string]interface{})

	for _, label := range itemMap {
		if key == nil || label.Key == *key {
			resultLabels[label.Key] = label.Value
		}
	}

	var gqlLabels graphql.Labels = resultLabels
	return gqlLabels, nil
}

func (r *Resolver) getRuntimeID(ctx context.Context) (string, error) {
	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	if consumerInfo.ConsumerType != consumer.Runtime {
		log.C(ctx).Errorf("Consumer type is of type %v. Runtime Contexts can be consumed only by runtimes...", consumerInfo.ConsumerType)
		return "", apperrors.NewUnauthorizedError("runtime context access is restricted to runtimes only")
	}

	return consumerInfo.ConsumerID, nil
}
