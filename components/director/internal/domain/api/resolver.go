package api

import (
	"context"

	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// APIService is responsible for the service-layer APIDefinition operations.
//go:generate mockery --name=APIService --output=automock --outpkg=automock --case=underscore
type APIService interface {
	CreateInBundle(ctx context.Context, appID, bundleID string, in model.APIDefinitionInput, spec *model.SpecInput) (string, error)
	Update(ctx context.Context, id string, in model.APIDefinitionInput, spec *model.SpecInput) error
	Get(ctx context.Context, id string) (*model.APIDefinition, error)
	Delete(ctx context.Context, id string) error
	ListFetchRequests(ctx context.Context, specIDs []string) ([]*model.FetchRequest, error)
}

// RuntimeService is responsible for the service-layer Runtime operations.
//go:generate mockery --name=RuntimeService --output=automock --outpkg=automock --case=underscore
type RuntimeService interface {
	Get(ctx context.Context, id string) (*model.Runtime, error)
}

// APIConverter converts EventDefinitions between the model.APIDefinition service-layer representation and the graphql-layer representation.
//go:generate mockery --name=APIConverter --output=automock --outpkg=automock --case=underscore
type APIConverter interface {
	ToGraphQL(in *model.APIDefinition, spec *model.Spec, bundleRef *model.BundleReference) (*graphql.APIDefinition, error)
	MultipleToGraphQL(in []*model.APIDefinition, specs []*model.Spec, bundleRefs []*model.BundleReference) ([]*graphql.APIDefinition, error)
	MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) ([]*model.APIDefinitionInput, []*model.SpecInput, error)
	InputFromGraphQL(in *graphql.APIDefinitionInput) (*model.APIDefinitionInput, *model.SpecInput, error)
}

// FetchRequestConverter converts FetchRequest between the model.FetchRequest service-layer representation and the graphql-layer one.
//go:generate mockery --name=FetchRequestConverter --output=automock --outpkg=automock --case=underscore
type FetchRequestConverter interface {
	ToGraphQL(in *model.FetchRequest) (*graphql.FetchRequest, error)
	InputFromGraphQL(in *graphql.FetchRequestInput) (*model.FetchRequestInput, error)
}

// BundleService is responsible for the service-layer Bundle operations.
//go:generate mockery --name=BundleService --output=automock --outpkg=automock --case=underscore
type BundleService interface {
	Get(ctx context.Context, id string) (*model.Bundle, error)
}

// ApplicationService is responsible for the service-layer Application operations.
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore
type ApplicationService interface {
	UpdateBaseURL(ctx context.Context, appID, targetURL string) error
}

// Resolver is an object responsible for resolver-layer APIDefinition operations
type Resolver struct {
	transact      persistence.Transactioner
	svc           APIService
	bndlSvc       BundleService
	bndlRefSvc    BundleReferenceService
	rtmSvc        RuntimeService
	converter     APIConverter
	frConverter   FetchRequestConverter
	specService   SpecService
	specConverter SpecConverter
	appSvc        ApplicationService
}

// NewResolver returns a new object responsible for resolver-layer APIDefinition operations.
func NewResolver(transact persistence.Transactioner, svc APIService, rtmSvc RuntimeService, bndlSvc BundleService, bndlRefSvc BundleReferenceService, converter APIConverter, frConverter FetchRequestConverter, specService SpecService, specConverter SpecConverter, appSvc ApplicationService) *Resolver {
	return &Resolver{
		transact:      transact,
		svc:           svc,
		rtmSvc:        rtmSvc,
		bndlSvc:       bndlSvc,
		bndlRefSvc:    bndlRefSvc,
		converter:     converter,
		frConverter:   frConverter,
		specService:   specService,
		specConverter: specConverter,
		appSvc:        appSvc,
	}
}

// AddAPIDefinitionToBundle adds an APIDefinition to a Bundle with a given ID,
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

	bndl, err := r.bndlSvc.Get(ctx, bundleID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, apperrors.NewInvalidDataError("cannot add API to not existing bundle")
		}
		return nil, errors.Wrapf(err, "while getting Bundle with id %s when adding APIDefinition", bundleID)
	}

	id, err := r.svc.CreateInBundle(ctx, bndl.ApplicationID, bundleID, *convertedIn, convertedSpec)
	if err != nil {
		return nil, errors.Wrapf(err, "Error occurred while creating APIDefinition in Bundle with id %s", bundleID)
	}

	api, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = r.appSvc.UpdateBaseURL(ctx, api.ApplicationID, in.TargetURL); err != nil {
		return nil, errors.Wrapf(err, "while trying to update baseURL")
	}

	spec, err := r.specService.GetByReferenceObjectID(ctx, model.APISpecReference, api.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for APIDefinition with id %q", api.ID)
	}

	bndlRef, err := r.bndlRefSvc.GetForBundle(ctx, model.BundleAPIReference, &api.ID, &bundleID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting bundle reference for APIDefinition with id %q", api.ID)
	}

	gqlAPI, err := r.converter.ToGraphQL(api, spec, bndlRef)
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

// UpdateAPIDefinition updates an APIDefinition by its ID.
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

	bndlRef, err := r.bndlRefSvc.GetForBundle(ctx, model.BundleAPIReference, &api.ID, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting bundle reference for APIDefinition with id %q", api.ID)
	}

	gqlAPI, err := r.converter.ToGraphQL(api, spec, bndlRef)
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

// DeleteAPIDefinition deletes an APIDefinition by its ID.
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

	bndlRef, err := r.bndlRefSvc.GetForBundle(ctx, model.BundleAPIReference, &api.ID, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting bundle reference for APIDefinition with id %q", api.ID)
	}

	gqlAPI, err := r.converter.ToGraphQL(api, spec, bndlRef)
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

// RefetchAPISpec refetches an APISpec for APIDefinition with given ID.
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

	spec, err := r.specService.RefetchSpec(ctx, dbSpec.ID, model.APISpecReference)
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

// FetchRequest returns a FetchRequest by a given EventSpec via dataloaders.
func (r *Resolver) FetchRequest(ctx context.Context, obj *graphql.APISpec) (*graphql.FetchRequest, error) {
	params := dataloader.ParamFetchRequestAPIDef{ID: obj.ID, Ctx: ctx}
	return dataloader.ForFetchRequestAPIDef(ctx).FetchRequestAPIDefByID.Load(params)
}

// FetchRequestAPIDefDataLoader is the dataloader implementation.
func (r *Resolver) FetchRequestAPIDefDataLoader(keys []dataloader.ParamFetchRequestAPIDef) ([]*graphql.FetchRequest, []error) {
	if len(keys) == 0 {
		return nil, []error{apperrors.NewInternalError("No APIDef specs found")}
	}

	ctx := keys[0].Ctx

	specIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		if key.ID == "" {
			return nil, []error{apperrors.NewInternalError("Cannot fetch FetchRequest. APIDefinition Spec ID is empty")}
		}
		specIDs = append(specIDs, key.ID)
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, []error{err}
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	fetchRequests, err := r.svc.ListFetchRequests(ctx, specIDs)
	if err != nil {
		return nil, []error{err}
	}

	if fetchRequests == nil {
		return nil, nil
	}

	gqlFetchRequests := make([]*graphql.FetchRequest, 0, len(fetchRequests))
	for _, fr := range fetchRequests {
		fetchRequest, err := r.frConverter.ToGraphQL(fr)
		if err != nil {
			return nil, []error{err}
		}
		gqlFetchRequests = append(gqlFetchRequests, fetchRequest)
	}

	if err = tx.Commit(); err != nil {
		return nil, []error{err}
	}

	log.C(ctx).Infof("Successfully fetched requests for Specifications %v", specIDs)
	return gqlFetchRequests, nil
}
