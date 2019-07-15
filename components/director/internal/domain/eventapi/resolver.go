package eventapi

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=EventAPIService -output=automock -outpkg=automock -case=underscore
type EventAPIService interface {
	Create(ctx context.Context, applicationID string, in model.EventAPIDefinitionInput) (string, error)
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

//go:generate mockery -name=ApplicationService -output=automock -outpkg=automock -case=underscore
type ApplicationService interface {
	Exist(ctx context.Context, id string) (bool, error)
}

type Resolver struct {
	svc       EventAPIService
	appSvc    ApplicationService
	converter EventAPIConverter
}

func NewResolver(svc EventAPIService, appSvc ApplicationService, converter EventAPIConverter) *Resolver {
	return &Resolver{
		svc:       svc,
		appSvc:    appSvc,
		converter: converter,
	}
}

func (r *Resolver) AddEventAPI(ctx context.Context, applicationID string, in graphql.EventAPIDefinitionInput) (*graphql.EventAPIDefinition, error) {
	convertedIn := r.converter.InputFromGraphQL(&in)

	found, err := r.appSvc.Exist(ctx, applicationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while checking existence of Application")
	}

	if found == false {
		return nil, errors.New("Cannot add EventAPI to not existing Application")
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
