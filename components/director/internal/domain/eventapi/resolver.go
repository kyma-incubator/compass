package eventapi

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"

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
	GetFetchRequest(ctx context.Context, eventAPIDefID string) (*model.FetchRequest, error)
}

//go:generate mockery -name=EventAPIConverter -output=automock -outpkg=automock -case=underscore
type EventAPIConverter interface {
	ToGraphQL(in *model.EventAPIDefinition) *graphql.EventAPIDefinition
	MultipleToGraphQL(in []*model.EventAPIDefinition) []*graphql.EventAPIDefinition
	MultipleInputFromGraphQL(in []*graphql.EventAPIDefinitionInput) []*model.EventAPIDefinitionInput
	InputFromGraphQL(in *graphql.EventAPIDefinitionInput) *model.EventAPIDefinitionInput
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
	transact  persistence.Transactioner
	svc       EventAPIService
	appSvc    ApplicationService
	converter EventAPIConverter
	frConverter FetchRequestConverter
}

func NewResolver(transact persistence.Transactioner, svc EventAPIService, appSvc ApplicationService, converter EventAPIConverter, frConverter FetchRequestConverter) *Resolver {
	return &Resolver{
		transact: transact,
		svc:       svc,
		appSvc:    appSvc,
		converter: converter,
		frConverter: frConverter,
	}
}

func (r *Resolver) AddEventAPI(ctx context.Context, applicationID string, in graphql.EventAPIDefinitionInput) (*graphql.EventAPIDefinition, error) {
	convertedIn := r.converter.InputFromGraphQL(&in)

	found, err := r.appSvc.Exist(ctx, applicationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while checking existence of Application")
	}

	if !found {
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

func (r *Resolver) FetchRequest(ctx context.Context, obj *graphql.EventAPISpec) (*graphql.FetchRequest, error) {
	if obj == nil {
		return nil, errors.New("Document cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	if obj.DefinitionID == "" {
		return nil, errors.New("Internet Server Error: Cannot fetch FetchRequest. EventAPIDefinition ID is empty")
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