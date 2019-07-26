package runtime

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=ContextValueSetter -output=automock -outpkg=automock -case=underscore
type ContextValueSetter interface {
	WithValue(parent context.Context, key interface{}, val interface{}) context.Context
}

//go:generate mockery -name=RuntimeService -output=automock -outpkg=automock -case=underscore
type RuntimeService interface {
	Create(ctx context.Context, in model.RuntimeInput) (string, error)
	Update(ctx context.Context, id string, in model.RuntimeInput) error
	Get(ctx context.Context, id string) (*model.Runtime, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.RuntimePage, error)
	SetLabel(ctx context.Context, runtimeID string, key string, value interface{}) error
	DeleteLabel(ctx context.Context, runtimeID string, key string) error
}

//go:generate mockery -name=RuntimeConverter -output=automock -outpkg=automock -case=underscore
type RuntimeConverter interface {
	ToGraphQL(in *model.Runtime) *graphql.Runtime
	MultipleToGraphQL(in []*model.Runtime) []*graphql.Runtime
	InputFromGraphQL(in graphql.RuntimeInput) model.RuntimeInput
}

type Resolver struct {
	transact  persistence.Transactioner
	ctxvs     ContextValueSetter
	svc       RuntimeService
	converter RuntimeConverter
}

func NewResolver(transact persistence.Transactioner, ctxvs ContextValueSetter, svc RuntimeService, conv RuntimeConverter) *Resolver {
	return &Resolver{
		transact:  transact,
		ctxvs:     ctxvs,
		svc:       svc,
		converter: conv,
	}
}

// TODO: Proper error handling
// TODO: Pagination

func (r *Resolver) Runtimes(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.RuntimePage, error) {
	labelFilter := labelfilter.MultipleFromGraphQL(filter)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = r.ctxvs.WithValue(ctx, persistence.PersistenceCtxKey, tx)

	runtimesPage, err := r.svc.List(ctx, labelFilter, first, &cursor)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlRuntimes := r.converter.MultipleToGraphQL(runtimesPage.Data)
	totalCount := len(gqlRuntimes)

	return &graphql.RuntimePage{
		Data:       gqlRuntimes,
		TotalCount: totalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(runtimesPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(runtimesPage.PageInfo.EndCursor),
			HasNextPage: runtimesPage.PageInfo.HasNextPage,
		},
	}, nil
}

func (r *Resolver) Runtime(ctx context.Context, id string) (*graphql.Runtime, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = r.ctxvs.WithValue(ctx, persistence.PersistenceCtxKey, tx)

	runtime, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.converter.ToGraphQL(runtime), nil
}

func (r *Resolver) CreateRuntime(ctx context.Context, in graphql.RuntimeInput) (*graphql.Runtime, error) {
	convertedIn := r.converter.InputFromGraphQL(in)

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = r.ctxvs.WithValue(ctx, persistence.PersistenceCtxKey, tx)

	id, err := r.svc.Create(ctx, convertedIn)
	if err != nil {
		return nil, err
	}

	runtime, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlRuntime := r.converter.ToGraphQL(runtime)

	return gqlRuntime, nil
}
func (r *Resolver) UpdateRuntime(ctx context.Context, id string, in graphql.RuntimeInput) (*graphql.Runtime, error) {
	convertedIn := r.converter.InputFromGraphQL(in)

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = r.ctxvs.WithValue(ctx, persistence.PersistenceCtxKey, tx)

	err = r.svc.Update(ctx, id, convertedIn)
	if err != nil {
		return nil, err
	}

	runtime, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlRuntime := r.converter.ToGraphQL(runtime)

	return gqlRuntime, nil
}

func (r *Resolver) DeleteRuntime(ctx context.Context, id string) (*graphql.Runtime, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = r.ctxvs.WithValue(ctx, persistence.PersistenceCtxKey, tx)

	runtime, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	deletedRuntime := r.converter.ToGraphQL(runtime)

	err = r.svc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return deletedRuntime, nil
}

func (r *Resolver) SetRuntimeLabel(ctx context.Context, runtimeID string, key string, value interface{}) (*graphql.Label, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = r.ctxvs.WithValue(ctx, persistence.PersistenceCtxKey, tx)

	err = r.svc.SetLabel(ctx, runtimeID, key, value)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &graphql.Label{
		Key:   key,
		Value: value,
	}, nil
}
func (r *Resolver) DeleteRuntimeLabel(ctx context.Context, runtimeID string, key string) (*graphql.Label, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = r.ctxvs.WithValue(ctx, persistence.PersistenceCtxKey, tx)

	runtime, err := r.svc.Get(ctx, runtimeID)
	if err != nil {
		return nil, err
	}

	value := runtime.Labels[key]

	err = r.svc.DeleteLabel(ctx, runtimeID, key)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &graphql.Label{
		Key:   key,
		Value: value,
	}, nil
}
