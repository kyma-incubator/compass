package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/pkg/errors"
)

//go:generate mockery -name=APIService -output=automock -outpkg=automock -case=underscore
type APIService interface {
	CreateInPackage(ctx context.Context, packageID string, in model.APIDefinitionInput) (string, error)
	Update(ctx context.Context, id string, in model.APIDefinitionInput) error
	Get(ctx context.Context, id string) (*model.APIDefinition, error)
	Delete(ctx context.Context, id string) error
	RefetchAPISpec(ctx context.Context, id string) (*model.APISpec, error)
	GetFetchRequest(ctx context.Context, apiDefID string) (*model.FetchRequest, error)
}

//go:generate mockery -name=RuntimeService -output=automock -outpkg=automock -case=underscore
type RuntimeService interface {
	Get(ctx context.Context, id string) (*model.Runtime, error)
}

//go:generate mockery -name=APIConverter -output=automock -outpkg=automock -case=underscore
type APIConverter interface {
	ToGraphQL(in *model.APIDefinition) *graphql.APIDefinition
	MultipleToGraphQL(in []*model.APIDefinition) []*graphql.APIDefinition
	MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) ([]*model.APIDefinitionInput, error)
	InputFromGraphQL(in *graphql.APIDefinitionInput) (*model.APIDefinitionInput, error)
	SpecToGraphQL(definitionID string, in *model.APISpec) *graphql.APISpec
}

//go:generate mockery -name=FetchRequestConverter -output=automock -outpkg=automock -case=underscore
type FetchRequestConverter interface {
	ToGraphQL(in *model.FetchRequest) (*graphql.FetchRequest, error)
	InputFromGraphQL(in *graphql.FetchRequestInput) (*model.FetchRequestInput, error)
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
	svc         APIService
	appSvc      ApplicationService
	pkgSvc      PackageService
	rtmSvc      RuntimeService
	converter   APIConverter
	frConverter FetchRequestConverter
}

func NewResolver(transact persistence.Transactioner, svc APIService, appSvc ApplicationService, rtmSvc RuntimeService, pkgSvc PackageService, converter APIConverter, frConverter FetchRequestConverter) *Resolver {
	return &Resolver{
		transact:    transact,
		svc:         svc,
		appSvc:      appSvc,
		rtmSvc:      rtmSvc,
		pkgSvc:      pkgSvc,
		converter:   converter,
		frConverter: frConverter,
	}
}

func (r *Resolver) AddAPIDefinitionToPackage(ctx context.Context, packageID string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, err := r.converter.InputFromGraphQL(&in)
	if err != nil {
		return nil, errors.Wrap(err, "while converting APIDefinition input from GraphQL")
	}

	found, err := r.pkgSvc.Exist(ctx, packageID)
	if err != nil {
		return nil, errors.Wrapf(err, "while checking existence of package")
	}

	if !found {
		return nil, apperrors.NewInvalidDataError("cannot add API to not existing package")
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

func (r *Resolver) UpdateAPIDefinition(ctx context.Context, id string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, err := r.converter.InputFromGraphQL(&in)
	if err != nil {
		return nil, errors.Wrap(err, "while converting APIDefinition input from GraphQL")
	}

	err = r.svc.Update(ctx, id, *convertedIn)
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
func (r *Resolver) DeleteAPIDefinition(ctx context.Context, id string) (*graphql.APIDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	api, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = r.svc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.converter.ToGraphQL(api), nil
}
func (r *Resolver) RefetchAPISpec(ctx context.Context, apiID string) (*graphql.APISpec, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	spec, err := r.svc.RefetchAPISpec(ctx, apiID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	converted := r.converter.SpecToGraphQL(apiID, spec)
	return converted, nil
}

func (r *Resolver) FetchRequest(ctx context.Context, obj *graphql.APISpec) (*graphql.FetchRequest, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("API Spec cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if obj.DefinitionID == "" {
		return nil, apperrors.NewInternalError("Cannot fetch FetchRequest. APIDefinition ID is empty")
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

	return r.frConverter.ToGraphQL(fr)
}
