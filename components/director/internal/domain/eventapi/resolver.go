package eventapi

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/uid"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//type EventAPIService interface{}

//go:generate mockery -name=EventAPIService -output=automock -outpkg=automock -case=underscore
type EventAPIService interface {
	Create(ctx context.Context, id string, applicationID string, in model.EventAPIDefinitionInput) (string, error)
	Update(ctx context.Context, id string, in model.EventAPIDefinitionInput) error
	Get(ctx context.Context, id string) (*model.EventAPIDefinition, error)
	Delete(ctx context.Context, id string) error
	RefetchAPISpec(ctx context.Context, id string) (*model.EventAPISpec, error)
}

//go:generate mockery -name=EventAPIConverter -output=automock -outpkg=automock -case=underscore
type EventAPIConverter interface {
	ToGraphQL(in *model.EventAPIDefinition) *graphql.EventAPIDefinition
	MultipleToGraphQL(in []*model.EventAPIDefinition) []*graphql.EventAPIDefinition
	MultipleInputFromGraphQL(in []*graphql.EventAPIDefinitionInput) []*model.EventAPIDefinitionInput
	InputFromGraphQL(in *graphql.EventAPIDefinitionInput) *model.EventAPIDefinitionInput
}

type Resolver struct {
	svc       EventAPIService
	converter EventAPIConverter
}

func NewResolver(svc EventAPIService, converter EventAPIConverter) *Resolver {
	return &Resolver{
		svc:       svc,
		converter: converter,
	}
}

func (r *Resolver) AddEventAPI(ctx context.Context, applicationID string, in graphql.EventAPIDefinitionInput) (*graphql.EventAPIDefinition, error) {
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

func (r *Resolver) UpdateEventAPI(ctx context.Context, id string, in graphql.EventAPIDefinitionInput) (*graphql.EventAPIDefinition, error) {
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

func (r *Resolver) DeleteEventAPI(ctx context.Context, id string) (*graphql.EventAPIDefinition, error) {
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

func (r *Resolver) RefetchEventAPISpec(ctx context.Context, eventID string) (*graphql.EventAPISpec, error) {
	spec, err := r.svc.RefetchAPISpec(ctx, eventID)
	if err != nil {
		return nil, err
	}

	convertedOut := r.converter.ToGraphQL(&model.EventAPIDefinition{Spec: spec})

	return convertedOut.Spec, nil
}
