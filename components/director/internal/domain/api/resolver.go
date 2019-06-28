package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/uid"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/graphql"
)

//go:generate mockery -name=APIService -output=automock -outpkg=automock -case=underscore
type APIService interface {
	Create(ctx context.Context, id string, applicationID string, in model.APIDefinitionInput) (string, error)
	Update(ctx context.Context, id string, in model.APIDefinitionInput) error
	Get(ctx context.Context, id string) (*model.APIDefinition, error)
	Delete(ctx context.Context, id string) error
	SetAPIAuth(ctx context.Context, apiID string, runtimeID string, in model.AuthInput) (*model.RuntimeAuth, error)
	DeleteAPIAuth(ctx context.Context, apiID string, runtimeID string) (*model.RuntimeAuth, error)
	RefetchAPISpec(ctx context.Context, id string) (*model.APISpec, error)
}

//go:generate mockery -name=APIConverter -output=automock -outpkg=automock -case=underscore
type APIConverter interface {
	ToGraphQL(in *model.APIDefinition) *graphql.APIDefinition
	MultipleToGraphQL(in []*model.APIDefinition) []*graphql.APIDefinition
	MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) []*model.APIDefinitionInput
	InputFromGraphQL(in *graphql.APIDefinitionInput) *model.APIDefinitionInput
}

type Resolver struct {
	svc           APIService
	converter     APIConverter
	authConverter AuthConverter
}

func NewResolver(svc APIService, converter APIConverter, authConverter AuthConverter) *Resolver {
	return &Resolver{
		svc:           svc,
		converter:     converter,
		authConverter: authConverter,
	}
}

func (r *Resolver) AddAPI(ctx context.Context, applicationID string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	convertedIn := r.converter.InputFromGraphQL(&in)

	id, err := r.svc.Create(ctx, uid.Generate(), applicationID, *convertedIn)
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
