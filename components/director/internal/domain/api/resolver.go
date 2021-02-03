package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

//go:generate mockery -name=APIService -output=automock -outpkg=automock -case=underscore
type APIService interface {
	CreateInBundle(ctx context.Context, bundleID string, in model.APIDefinitionInput, spec *model.SpecInput) (string, error)
	Update(ctx context.Context, id string, in model.APIDefinitionInput, spec *model.SpecInput) error
	Get(ctx context.Context, id string) (*model.APIDefinition, error)
	Delete(ctx context.Context, id string) error
	GetFetchRequest(ctx context.Context, apiDefID string) (*model.FetchRequest, error)
}

//go:generate mockery -name=RuntimeService -output=automock -outpkg=automock -case=underscore
type RuntimeService interface {
	Get(ctx context.Context, id string) (*model.Runtime, error)
}

//go:generate mockery -name=APIConverter -output=automock -outpkg=automock -case=underscore
type APIConverter interface {
	ToGraphQL(in *model.APIDefinition, spec *model.Spec) (*graphql.APIDefinition, error)
	MultipleToGraphQL(in []*model.APIDefinition, specs []*model.Spec) ([]*graphql.APIDefinition, error)
	MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) ([]*model.APIDefinitionInput, []*model.SpecInput, error)
	InputFromGraphQL(in *graphql.APIDefinitionInput) (*model.APIDefinitionInput, *model.SpecInput, error)
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

//go:generate mockery -name=BundleService -output=automock -outpkg=automock -case=underscore
type BundleService interface {
	Exist(ctx context.Context, id string) (bool, error)
}

type Resolver struct {
	transact      persistence.Transactioner
	svc           APIService
	appSvc        ApplicationService
	bndlSvc       BundleService
	rtmSvc        RuntimeService
	converter     APIConverter
	frConverter   FetchRequestConverter
	specService   SpecService
	specConverter SpecConverter
}

func NewResolver(transact persistence.Transactioner, svc APIService, appSvc ApplicationService, rtmSvc RuntimeService, bndlSvc BundleService, converter APIConverter, frConverter FetchRequestConverter, specService SpecService, specConverter SpecConverter) *Resolver {
	return &Resolver{
		transact:      transact,
		svc:           svc,
		appSvc:        appSvc,
		rtmSvc:        rtmSvc,
		bndlSvc:       bndlSvc,
		converter:     converter,
		frConverter:   frConverter,
		specService:   specService,
		specConverter: specConverter,
	}
}

func (r *Resolver) AddAPIDefinitionToBundle(ctx context.Context, bundleID string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Adding APIDefinition to bundle with id %s", bundleID)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, convertedSpec, err := r.converter.InputFromGraphQL(&in)
	if err != nil {
		return nil, errors.Wrap(err, "while converting GraphQL input to APIDefinition")
	}

	found, err := r.bndlSvc.Exist(ctx, bundleID)
	if err != nil {
		return nil, errors.Wrapf(err, "while checking existence of Bundle with id %s when adding APIDefinition", bundleID)
	}

	if !found {
		return nil, apperrors.NewInvalidDataError("cannot add API to not existing bundle")
	}

	id, err := r.svc.CreateInBundle(ctx, bundleID, *convertedIn, convertedSpec)
	if err != nil {
		return nil, errors.Wrapf(err, "Error occurred while creating APIDefinition in Bundle with id %s", bundleID)
	}

	api, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	spec, err := r.specService.GetByReferenceObjectID(ctx, model.APISpecReference, api.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for APIDefinition with id %q", api.ID)
	}

	gqlAPI, err := r.converter.ToGraphQL(api, spec)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting APIDefinition with id %q to graphQL", api.ID)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("APIDefinition with id %s successfully added to Bundle with id %s", id, bundleID)
	return gqlAPI, nil
}

func (r *Resolver) UpdateAPIDefinition(ctx context.Context, id string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Updating APIDefinition with id %s", id)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, convertedSpec, err := r.converter.InputFromGraphQL(&in)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting GraphQL input to APIDefinition with id %s", id)
	}

	err = r.svc.Update(ctx, id, *convertedIn, convertedSpec)
	if err != nil {
		return nil, err
	}

	api, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	spec, err := r.specService.GetByReferenceObjectID(ctx, model.APISpecReference, api.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for APIDefinition with id %q", api.ID)
	}

	gqlAPI, err := r.converter.ToGraphQL(api, spec)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting APIDefinition with id %q to graphQL", api.ID)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("APIDefinition with id %s successfully updated.", id)
	return gqlAPI, nil
}
func (r *Resolver) DeleteAPIDefinition(ctx context.Context, id string) (*graphql.APIDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Deleting APIDefinition with id %s", id)

	ctx = persistence.SaveToContext(ctx, tx)

	api, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	spec, err := r.specService.GetByReferenceObjectID(ctx, model.APISpecReference, api.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for APIDefinition with id %q", api.ID)
	}

	gqlAPI, err := r.converter.ToGraphQL(api, spec)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting APIDefinition with id %q to graphQL", api.ID)
	}

	err = r.svc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("APIDefinition with id %s successfully deleted.", id)
	return gqlAPI, nil
}

func (r *Resolver) RefetchAPISpec(ctx context.Context, apiID string) (*graphql.APISpec, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Refetching APISpec for API with id %s", apiID)

	ctx = persistence.SaveToContext(ctx, tx)

	dbSpec, err := r.specService.GetByReferenceObjectID(ctx, model.APISpecReference, apiID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for APIDefinition with id %q", apiID)
	}

	if dbSpec == nil {
		return nil, errors.Errorf("spec for API with id %q not found", apiID)
	}

	spec, err := r.specService.RefetchSpec(ctx, dbSpec.ID)
	if err != nil {
		return nil, err
	}

	converted, err := r.specConverter.ToGraphQLAPISpec(spec)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Successfully refetched APISpec for APIDefinition with id %s", apiID)
	return converted, nil
}

func (r *Resolver) FetchRequest(ctx context.Context, obj *graphql.APISpec) (*graphql.FetchRequest, error) {
	log.C(ctx).Infof("Fetching request for APIDefinition with id %s", obj.DefinitionID)

	if obj == nil {
		return nil, apperrors.NewInternalError("Error occurred when fetching request for APIDefinition. API Spec cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

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

	log.C(ctx).Infof("Successfully fetched request for APIDefinition %s", obj.DefinitionID)
	return r.frConverter.ToGraphQL(fr)
}
