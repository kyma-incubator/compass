package eventdef

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=EventDefService -output=automock -outpkg=automock -case=underscore
type EventDefService interface {
	Create(ctx context.Context, applicationID string, in model.EventDefinitionInput) (string, error)
	CreateInPackage(ctx context.Context, packageID string, in model.EventDefinitionInput) (string, error)
	Update(ctx context.Context, id string, in model.EventDefinitionInput) error
	Get(ctx context.Context, id string) (*model.EventDefinition, error)
	Delete(ctx context.Context, id string) error
	RefetchAPISpec(ctx context.Context, id string) (*model.EventSpec, error)
	GetFetchRequest(ctx context.Context, eventAPIDefID string) (*model.FetchRequest, error)
}

//go:generate mockery -name=EventDefConverter -output=automock -outpkg=automock -case=underscore
type EventDefConverter interface {
	ToGraphQL(in *model.EventDefinition) *graphql.EventDefinition
	MultipleToGraphQL(in []*model.EventDefinition) []*graphql.EventDefinition
	MultipleInputFromGraphQL(in []*graphql.EventDefinitionInput) []*model.EventDefinitionInput
	InputFromGraphQL(in *graphql.EventDefinitionInput) *model.EventDefinitionInput
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

//go:generate mockery -name=PackageService -output=automock -outpkg=automock -case=underscore
type PackageService interface {
	Exist(ctx context.Context, id string) (bool, error)
}

type Resolver struct {
	transact    persistence.Transactioner
	svc         EventDefService
	appSvc      ApplicationService
	pkgSvc      PackageService
	converter   EventDefConverter
	frConverter FetchRequestConverter
}

func NewResolver(transact persistence.Transactioner, svc EventDefService, appSvc ApplicationService, pkgSvc PackageService, converter EventDefConverter, frConverter FetchRequestConverter) *Resolver {
	return &Resolver{
		transact:    transact,
		svc:         svc,
		appSvc:      appSvc,
		pkgSvc:      pkgSvc,
		converter:   converter,
		frConverter: frConverter,
	}
}

func (r *Resolver) AddEventDefinition(ctx context.Context, applicationID string, in graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn := r.converter.InputFromGraphQL(&in)

	found, err := r.appSvc.Exist(ctx, applicationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while checking existence of Application")
	}

	if !found {
		return nil, errors.New("Cannot add Event Definition to not existing Application")
	}

	id, err := r.svc.Create(ctx, applicationID, *convertedIn)
	if err != nil {
		return nil, err
	}

	api, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlAPI := r.converter.ToGraphQL(api)

	return gqlAPI, nil
}

func (r *Resolver) UpdateEventDefinition(ctx context.Context, id string, in graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn := r.converter.InputFromGraphQL(&in)

	err = r.svc.Update(ctx, id, *convertedIn)
	if err != nil {
		return nil, err
	}

	api, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	gqlAPI := r.converter.ToGraphQL(api)

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return gqlAPI, nil
}

func (r *Resolver) DeleteEventDefinition(ctx context.Context, id string) (*graphql.EventDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	api, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	deletedAPI := r.converter.ToGraphQL(api)

	err = r.svc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return deletedAPI, nil
}

func (r *Resolver) RefetchEventDefinitionSpec(ctx context.Context, eventID string) (*graphql.EventSpec, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	spec, err := r.svc.RefetchAPISpec(ctx, eventID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	convertedOut := r.converter.ToGraphQL(&model.EventDefinition{Spec: spec})

	return convertedOut.Spec, nil
}

func (r *Resolver) FetchRequest(ctx context.Context, obj *graphql.EventSpec) (*graphql.FetchRequest, error) {
	if obj == nil {
		return nil, errors.New("Event Spec cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if obj.DefinitionID == "" {
		return nil, errors.New("Internal Server Error: Cannot fetch FetchRequest. EventDefinition ID is empty")
	}

	fr, err := r.svc.GetFetchRequest(ctx, obj.DefinitionID)
	if err != nil {
		return nil, err
	}

	if fr == nil {
		return nil, nil
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	frGQL := r.frConverter.ToGraphQL(fr)
	return frGQL, nil
}

func (r *Resolver) AddEventDefinitionToPackage(ctx context.Context, packageID string, in graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn := r.converter.InputFromGraphQL(&in)

	found, err := r.pkgSvc.Exist(ctx, packageID)
	if err != nil {
		return nil, errors.Wrapf(err, "while checking existence of Package")
	}

	if !found {
		return nil, errors.New("Cannot add Event Definition to not existing Package")
	}

	id, err := r.svc.CreateInPackage(ctx, packageID, *convertedIn)
	if err != nil {
		return nil, err
	}

	api, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlAPI := r.converter.ToGraphQL(api)

	return gqlAPI, nil
}
