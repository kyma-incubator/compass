package api

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/graphql"
)

//go:generate mockery -name=APIService -output=automock -outpkg=automock -case=underscore
type APIService interface{
	Create(ctx context.Context, in model.APIDefinitionInput) (string, error)
	Update(ctx context.Context, id string, in model.APIDefinitionInput) error
	Get(ctx context.Context, id string) (*model.APIDefinition, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.APIDefinitionPage, error)
	SetApiAuth(ctx context.Context, apiID string, runtimeID string, in model.AuthInput) error
	DeleteAPIAuth(ctx context.Context, apiID string, runtimeID string) error
	RefetchAPISpec(ctx context.Context, id string) (*model.APISpec,error)
}

//go:generate mockery -name=APIConverter -output=automock -outpkg=automock -case=underscore
type APIConverter interface{
	ToGraphQL(in *model.APIDefinition) *graphql.APIDefinition
	MultipleToGraphQL(in []*model.APIDefinition) []*graphql.APIDefinition
	MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) []*model.APIDefinitionInput
	InputFromGraphQL(in *graphql.APIDefinitionInput) *model.APIDefinitionInput
}

type Resolver struct {
	svc       APIService
	converter APIConverter
}


func NewResolver(svc APIService, converter APIConverter) *Resolver {
	return &Resolver{
		svc:       svc,
		converter: converter,
	}
}

func (r *Resolver) ApiDefinitions(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.APIDefinitionPage,error){
	labelFilter := labelfilter.MultipleFromGraphQL(filter)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	apisPage, err := r.svc.List(ctx, labelFilter, first, &cursor)
	if err != nil {
		return nil, err
	}

	gqlApis := r.converter.MultipleToGraphQL(apisPage.Data)
	totalCount := len(gqlApis)

	return &graphql.APIDefinitionPage{
		Data:       gqlApis,
		TotalCount: totalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(apisPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(apisPage.PageInfo.EndCursor),
			HasNextPage: apisPage.PageInfo.HasNextPage,
		},
	}, nil
}

func (r *Resolver) ApiDefinition(ctx context.Context, id string) (*graphql.APIDefinition,error){
	api, err := r.svc.Get(ctx,id)
	if err != nil {
		return nil,err
	}

	return r.converter.ToGraphQL(api),nil
}

func (r *Resolver) AddAPI(ctx context.Context, applicationID string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	convertedIn := r.converter.InputFromGraphQL(&in)

	id, err := r.svc.Create(ctx,*convertedIn)
	if err != nil {
		return nil,err
	}

	api, err := r.svc.Get(ctx,id)
	if err != nil {
		return nil, err
	}

	gqlApi := r.converter.ToGraphQL(api)

	return gqlApi,nil

}
func (r *Resolver) UpdateAPI(ctx context.Context, id string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	convertedIn := r.converter.InputFromGraphQL(&in)

	err := r.svc.Update(ctx,id,*convertedIn)
	if err != nil {
		return nil,err
	}

	api, err := r.svc.Get(ctx,id)
	if err != nil {
		return nil, err
	}

	gqlApi := r.converter.ToGraphQL(api)

	return gqlApi,nil
}
func (r *Resolver) DeleteAPI(ctx context.Context, id string) (*graphql.APIDefinition, error) {
	api, err := r.svc.Get(ctx,id)
	if err != nil {
		return nil, err
	}

	deletedApi := r.converter.ToGraphQL(api)

	err = r.svc.Delete(ctx,id)
	if err != nil {
		return nil,err
	}

	return deletedApi,nil
}
func (r *Resolver) RefetchAPISpec(ctx context.Context, apiID string) (*graphql.APISpec, error) {
	spec, err := r.svc.RefetchAPISpec(ctx,apiID)
	if err != nil {
		return nil, err
	}

	convertedOut := r.converter.ToGraphQL(&model.APIDefinition{Spec: spec})

	return convertedOut.Spec, nil
}

func (r *Resolver) SetAPIAuth(ctx context.Context, apiID string, runtimeID string, in graphql.AuthInput) (*graphql.RuntimeAuth, error) {

	conv := auth.NewConverter()
	convertedIn := conv.InputFromGraphQL(&in)

	err := r.svc.SetApiAuth(ctx,apiID,runtimeID,*convertedIn)
	if err != nil {
		return  nil, err
	}

	api, err := r.svc.Get(ctx,apiID)
	if err != nil {
		return nil, err
	}

	convertedOut := &graphql.RuntimeAuth{
		RuntimeID: api.Auth.RuntimeID,
		Auth:      conv.ToGraphQL(api.Auth.Auth),
	}

	return convertedOut, nil
}
func (r *Resolver) DeleteAPIAuth(ctx context.Context, apiID string, runtimeID string) (*graphql.RuntimeAuth, error) {
	conv := auth.NewConverter()

	err := r.svc.DeleteAPIAuth(ctx,apiID,runtimeID)
	if err != nil {
		return nil, err
	}

	api, err := r.svc.Get(ctx,apiID)
	if err != nil {
		return nil, err
	}

	convertedOut := &graphql.RuntimeAuth{
		RuntimeID: api.Auth.RuntimeID,
		Auth:      conv.ToGraphQL(api.Auth.Auth),
	}
	return convertedOut, nil
}
