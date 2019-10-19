package runtime

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=RuntimeService -output=automock -outpkg=automock -case=underscore
type RuntimeService interface {
	Create(ctx context.Context, in model.RuntimeInput) (string, error)
	Update(ctx context.Context, id string, in model.RuntimeInput) error
	Get(ctx context.Context, id string) (*model.Runtime, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimePage, error)
	SetLabel(ctx context.Context, label *model.LabelInput) error
	GetLabel(ctx context.Context, runtimeID string, key string) (*model.Label, error)
	ListLabels(ctx context.Context, runtimeID string) (map[string]*model.Label, error)
	DeleteLabel(ctx context.Context, runtimeID string, key string) error
}

//go:generate mockery -name=SystemAuthService -output=automock -outpkg=automock -case=underscore
type SystemAuthService interface {
	ListForObject(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string) ([]model.SystemAuth, error)
}

//go:generate mockery -name=RuntimeConverter -output=automock -outpkg=automock -case=underscore
type RuntimeConverter interface {
	ToGraphQL(in *model.Runtime) *graphql.Runtime
	MultipleToGraphQL(in []*model.Runtime) []*graphql.Runtime
	InputFromGraphQL(in graphql.RuntimeInput) model.RuntimeInput
}

//go:generate mockery -name=SystemAuthConverter -output=automock -outpkg=automock -case=underscore
type SystemAuthConverter interface {
	ToGraphQL(in *model.SystemAuth) *graphql.SystemAuth
}

//go:generate mockery -name=OAuth20Service -output=automock -outpkg=automock -case=underscore
type OAuth20Service interface {
	DeleteClientCredentials(ctx context.Context, clientID string) error
}

type Resolver struct {
	transact    persistence.Transactioner
	svc         RuntimeService
	sysAuthSvc  SystemAuthService
	converter   RuntimeConverter
	sysAuthConv SystemAuthConverter
	oAuth20Svc  OAuth20Service
}

func NewResolver(transact persistence.Transactioner, svc RuntimeService, sysAuthSvc SystemAuthService, oAuthSvc OAuth20Service, conv RuntimeConverter, sysAuthConv SystemAuthConverter) *Resolver {
	return &Resolver{
		transact:    transact,
		svc:         svc,
		sysAuthSvc:  sysAuthSvc,
		oAuth20Svc:  oAuthSvc,
		converter:   conv,
		sysAuthConv: sysAuthConv,
	}
}

// TODO: Proper error handling
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

	ctx = persistence.SaveToContext(ctx, tx)

	if first == nil {
		return nil, errors.New("missing required parameter 'first'")
	}

	runtimesPage, err := r.svc.List(ctx, labelFilter, *first, cursor)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlRuntimes := r.converter.MultipleToGraphQL(runtimesPage.Data)

	return &graphql.RuntimePage{
		Data:       gqlRuntimes,
		TotalCount: runtimesPage.TotalCount,
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

	ctx = persistence.SaveToContext(ctx, tx)

	runtime, err := r.svc.Get(ctx, id)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, nil
		}
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

	ctx = persistence.SaveToContext(ctx, tx)

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

	ctx = persistence.SaveToContext(ctx, tx)

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

	ctx = persistence.SaveToContext(ctx, tx)

	runtime, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	auths, err := r.sysAuthSvc.ListForObject(ctx, model.RuntimeReference, runtime.ID)
	if err != nil {
		return nil, err
	}

	for _, auth := range auths {
		if auth.Value != nil {
			if auth.Value.Credential.Oauth != nil {
				err := r.oAuth20Svc.DeleteClientCredentials(ctx, auth.Value.Credential.Oauth.ClientID)
				if err != nil {
					return nil, errors.Wrap(err, "while deleting OAuth 2.0 client")
				}
			}
		}
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

	ctx = persistence.SaveToContext(ctx, tx)

	err = r.svc.SetLabel(ctx, &model.LabelInput{
		Key:        key,
		Value:      value,
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   runtimeID,
	})
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

	ctx = persistence.SaveToContext(ctx, tx)

	label, err := r.svc.GetLabel(ctx, runtimeID, key)
	if err != nil {
		return nil, err
	}

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
		Value: label.Value,
	}, nil
}

func (r *Resolver) Labels(ctx context.Context, obj *graphql.Runtime, key *string) (graphql.Labels, error) {
	if obj == nil {
		return nil, errors.New("Runtime cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	itemMap, err := r.svc.ListLabels(ctx, obj.ID)
	if err != nil {
		if strings.Contains(err.Error(), "doesn't exist") { // TODO: Use custom error and check its type
			return graphql.Labels{}, nil
		}

		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	resultLabels := make(map[string]interface{})

	for _, label := range itemMap {
		resultLabels[label.Key] = label.Value
	}

	return resultLabels, nil
}

func (r *Resolver) Auths(ctx context.Context, obj *graphql.Runtime) ([]*graphql.SystemAuth, error) {
	if obj == nil {
		return nil, errors.New("Runtime cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	sysAuths, err := r.sysAuthSvc.ListForObject(ctx, model.RuntimeReference, obj.ID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	var out []*graphql.SystemAuth
	for _, sa := range sysAuths {
		c := r.sysAuthConv.ToGraphQL(&sa)
		out = append(out, c)
	}

	return out, nil
}
