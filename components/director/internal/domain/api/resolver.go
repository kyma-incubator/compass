package api

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=APIService -output=automock -outpkg=automock -case=underscore
type APIService interface {
	Create(ctx context.Context, applicationID string, in model.APIDefinitionInput) (string, error)
	Update(ctx context.Context, id string, in model.APIDefinitionInput) error
	Get(ctx context.Context, id string) (*model.APIDefinition, error)
	Delete(ctx context.Context, id string) error
	SetAPIAuth(ctx context.Context, apiID string, runtimeID string, in model.AuthInput) (*model.RuntimeAuth, error)
	DeleteAPIAuth(ctx context.Context, apiID string, runtimeID string) (*model.RuntimeAuth, error)
	RefetchAPISpec(ctx context.Context, id string) (*model.APISpec, error)
	GetFetchRequest(ctx context.Context, apiDefID string) (*model.FetchRequest, error)
}

//go:generate mockery -name=APIConverter -output=automock -outpkg=automock -case=underscore
type APIConverter interface {
	ToGraphQL(in *model.APIDefinition) *graphql.APIDefinition
	MultipleToGraphQL(in []*model.APIDefinition) []*graphql.APIDefinition
	MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) []*model.APIDefinitionInput
	InputFromGraphQL(in *graphql.APIDefinitionInput) *model.APIDefinitionInput
}

//go:generate mockery -name=FetchRequestConverter -output=automock -outpkg=automock -case=underscore
type FetchRequestConverter interface {
	ToGraphQL(in *model.FetchRequest) *graphql.FetchRequest
	InputFromGraphQL(in *graphql.FetchRequestInput) *model.FetchRequestInput
}

//go:generate mockery -name=ApplicationService -output=automock -outpkg=automock -case=underscore
type ApplicationService interface {
	Exist(ctx context.Context, id string) (bool, error)
}

type Resolver struct {
	transact persistence.Transactioner
	svc           APIService
	appSvc        ApplicationService
	converter     APIConverter
	authConverter AuthConverter
	frConverter FetchRequestConverter
}

func NewResolver(transact persistence.Transactioner, svc APIService, appSvc ApplicationService, converter APIConverter, authConverter AuthConverter, frConverter FetchRequestConverter) *Resolver {
	return &Resolver{
		transact: transact,
		svc:           svc,
		appSvc:        appSvc,
		converter:     converter,
		authConverter: authConverter,
		frConverter:frConverter,
	}
}

func (r *Resolver) AddAPI(ctx context.Context, applicationID string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	convertedIn := r.converter.InputFromGraphQL(&in)

	found, err := r.appSvc.Exist(ctx, applicationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while checking existence of Application")
	}

	if !found {
		return nil, errors.New("Cannot add API to not existing Application")
	}

	id, err := r.svc.Create(ctx, applicationID, *convertedIn)
	if err != nil {
		return nil, err
	}

	api, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	gqlAPI := r.converter.ToGraphQL(api)

	return gqlAPI, nil
}
func (r *Resolver) UpdateAPI(ctx context.Context, id string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	convertedIn := r.converter.InputFromGraphQL(&in)

	err := r.svc.Update(ctx, id, *convertedIn)
	if err != nil {
		return nil, err
	}

	api, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	gqlAPI := r.converter.ToGraphQL(api)

	return gqlAPI, nil
}
func (r *Resolver) DeleteAPI(ctx context.Context, id string) (*graphql.APIDefinition, error) {
	api, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	deletedAPI := r.converter.ToGraphQL(api)

	err = r.svc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	return deletedAPI, nil
}
func (r *Resolver) RefetchAPISpec(ctx context.Context, apiID string) (*graphql.APISpec, error) {
	spec, err := r.svc.RefetchAPISpec(ctx, apiID)
	if err != nil {
		return nil, err
	}

	convertedOut := r.converter.ToGraphQL(&model.APIDefinition{Spec: spec})

	return convertedOut.Spec, nil
}

func (r *Resolver) SetAPIAuth(ctx context.Context, apiID string, runtimeID string, in graphql.AuthInput) (*graphql.RuntimeAuth, error) {
	convertedIn := r.authConverter.InputFromGraphQL(&in)

	runtimeAuth, err := r.svc.SetAPIAuth(ctx, apiID, runtimeID, *convertedIn)
	if err != nil {
		return nil, err
	}

	convertedOut := &graphql.RuntimeAuth{
		RuntimeID: runtimeAuth.RuntimeID,
		Auth:      r.authConverter.ToGraphQL(runtimeAuth.Auth),
	}

	return convertedOut, nil
}
func (r *Resolver) DeleteAPIAuth(ctx context.Context, apiID string, runtimeID string) (*graphql.RuntimeAuth, error) {
	runtimeAuth, err := r.svc.DeleteAPIAuth(ctx, apiID, runtimeID)
	if err != nil {
		return nil, err
	}

	convertedOut := &graphql.RuntimeAuth{
		RuntimeID: runtimeAuth.RuntimeID,
		Auth:      r.authConverter.ToGraphQL(runtimeAuth.Auth),
	}

	return convertedOut, nil
}

func (r *Resolver) FetchRequest(ctx context.Context, obj *graphql.APISpec) (*graphql.FetchRequest, error) {
	if obj == nil {
		return nil, errors.New("Document cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	if obj.DefinitionID == "" {
		return nil, errors.New("Internet Server Error: Cannot fetch FetchRequest. APIDefinition ID is empty")
	}

	fr, err := r.svc.GetFetchRequest(ctx, obj.DefinitionID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	frGQL := r.frConverter.ToGraphQL(fr)
	return frGQL, nil
}
