package eventdef

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// EventDefService is responsible for the service-layer EventDefinition operations.
//
//go:generate mockery --name=EventDefService --output=automock --outpkg=automock --case=underscore --disable-version-string
type EventDefService interface {
	CreateInBundle(ctx context.Context, resourceType resource.Type, resourceID string, bundleID string, in model.EventDefinitionInput, spec *model.SpecInput) (string, error)
	CreateInApplication(ctx context.Context, appID string, in model.EventDefinitionInput, spec *model.SpecInput) (string, error)
	Update(ctx context.Context, resourceType resource.Type, id string, in model.EventDefinitionInput, spec *model.SpecInput) error
	UpdateForApplication(ctx context.Context, id string, in model.EventDefinitionInput, specIn *model.SpecInput) error
	Get(ctx context.Context, id string) (*model.EventDefinition, error)
	Delete(ctx context.Context, resourceType resource.Type, id string) error
	ListByApplicationIDPage(ctx context.Context, appID string, pageSize int, cursor string) (*model.EventDefinitionPage, error)
	ListFetchRequests(ctx context.Context, eventDefIDs []string) ([]*model.FetchRequest, error)
}

// EventDefConverter converts EventDefinitions between the model.EventDefinition service-layer representation and the graphql-layer representation.
//
//go:generate mockery --name=EventDefConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EventDefConverter interface {
	ToGraphQL(in *model.EventDefinition, spec *model.Spec, bundleReference *model.BundleReference) (*graphql.EventDefinition, error)
	MultipleToGraphQL(in []*model.EventDefinition, specs []*model.Spec, bundleRefs []*model.BundleReference) ([]*graphql.EventDefinition, error)
	MultipleInputFromGraphQL(in []*graphql.EventDefinitionInput) ([]*model.EventDefinitionInput, []*model.SpecInput, error)
	InputFromGraphQL(in *graphql.EventDefinitionInput) (*model.EventDefinitionInput, *model.SpecInput, error)
}

// FetchRequestConverter converts FetchRequest from the model.FetchRequest service-layer representation to the graphql-layer one.
//
//go:generate mockery --name=FetchRequestConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type FetchRequestConverter interface {
	ToGraphQL(in *model.FetchRequest) (*graphql.FetchRequest, error)
}

// BundleService is responsible for the service-layer Bundle operations.
//
//go:generate mockery --name=BundleService --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleService interface {
	Get(ctx context.Context, id string) (*model.Bundle, error)
}

// Resolver is an object responsible for resolver-layer EventDefinition operations
type Resolver struct {
	transact      persistence.Transactioner
	svc           EventDefService
	bndlSvc       BundleService
	bndlRefSvc    BundleReferenceService
	converter     EventDefConverter
	frConverter   FetchRequestConverter
	specConverter SpecConverter
	specService   SpecService
}

// NewResolver returns a new object responsible for resolver-layer EventDefinition operations.
func NewResolver(transact persistence.Transactioner, svc EventDefService, bndlSvc BundleService, bndlRefSvc BundleReferenceService, converter EventDefConverter, frConverter FetchRequestConverter, specService SpecService, specConverter SpecConverter) *Resolver {
	return &Resolver{
		transact:      transact,
		svc:           svc,
		bndlSvc:       bndlSvc,
		bndlRefSvc:    bndlRefSvc,
		converter:     converter,
		frConverter:   frConverter,
		specConverter: specConverter,
		specService:   specService,
	}
}

// EventDefinitionsForApplication lists all EventDefinitions for a given application ID with paging.
func (r *Resolver) EventDefinitionsForApplication(ctx context.Context, appID string, first *int, after *graphql.PageCursor) (*graphql.EventDefinitionPage, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Listing EventDefinition for Application with ID %s", appID)

	ctx = persistence.SaveToContext(ctx, tx)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}
	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter 'first'")
	}

	eventPage, err := r.svc.ListByApplicationIDPage(ctx, appID, *first, cursor)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing EventDefinition for Application with ID %s", appID)
	}

	gqlEvents := make([]*graphql.EventDefinition, 0, len(eventPage.Data))
	for _, event := range eventPage.Data {
		spec, err := r.specService.GetByReferenceObjectID(ctx, resource.Application, model.EventSpecReference, event.ID)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting spec for EventDefinition with id %q", event.ID)
		}

		gqlEvent, err := r.converter.ToGraphQL(event, spec, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting EventDefinition with id %q to graphQL", event.ID)
		}

		gqlEvents = append(gqlEvents, gqlEvent)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &graphql.EventDefinitionPage{
		Data:       gqlEvents,
		TotalCount: eventPage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(eventPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(eventPage.PageInfo.EndCursor),
			HasNextPage: eventPage.PageInfo.HasNextPage,
		},
	}, nil
}

// AddEventDefinitionToBundle adds an EventDefinition to a Bundle with a given ID.
func (r *Resolver) AddEventDefinitionToBundle(ctx context.Context, bundleID string, in graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Adding EventDefinition to bundle with id %s", bundleID)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, convertedSpec, err := r.converter.InputFromGraphQL(&in)
	if err != nil {
		return nil, errors.Wrap(err, "while converting GraphQL input to EventDefinition")
	}

	bndl, err := r.bndlSvc.Get(ctx, bundleID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, apperrors.NewInvalidDataError("cannot add Event Definition to not existing Bundle")
		}
		return nil, errors.Wrapf(err, "while checking existence of Bundle with id %s when adding EventDefinition", bundleID)
	}

	id, err := r.svc.CreateInBundle(ctx, resource.Application, str.PtrStrToStr(bndl.ApplicationID), bundleID, *convertedIn, convertedSpec)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating EventDefinition in Bundle with id %s", bundleID)
	}

	event, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	spec, err := r.specService.GetByReferenceObjectID(ctx, resource.Application, model.EventSpecReference, event.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for EventDefinition with id %q", event.ID)
	}

	bndlRef, err := r.bndlRefSvc.GetForBundle(ctx, model.BundleEventReference, &event.ID, &bundleID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting bundle reference for EventDefinition with id %q", event.ID)
	}

	gqlEvent, err := r.converter.ToGraphQL(event, spec, bndlRef)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting EventDefinition with id %q to graphQL", event.ID)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("EventDefinition with id %s successfully added to bundle with id %s", id, bundleID)
	return gqlEvent, nil
}

// AddEventDefinitionToApplication adds an EventDefinition in the context of an Application without Bundle
func (r *Resolver) AddEventDefinitionToApplication(ctx context.Context, appID string, in graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Adding EventDefinition to Application with id %s", appID)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, convertedSpec, err := r.converter.InputFromGraphQL(&in)
	if err != nil {
		return nil, errors.Wrap(err, "while converting GraphQL input to EventDefinition")
	}

	id, err := r.svc.CreateInApplication(ctx, appID, *convertedIn, convertedSpec)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating EventDefinition in Application with id %s", appID)
	}

	event, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	spec, err := r.specService.GetByReferenceObjectID(ctx, resource.Application, model.EventSpecReference, event.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for EventDefinition with id %q", event.ID)
	}

	gqlEvent, err := r.converter.ToGraphQL(event, spec, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting EventDefinition with id %q to graphQL", event.ID)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	log.C(ctx).Infof("EventDefinition with id %s successfully added to Application with id %s", id, appID)
	return gqlEvent, nil
}

// UpdateEventDefinition updates an EventDefinition by its ID.
func (r *Resolver) UpdateEventDefinition(ctx context.Context, id string, in graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Updating EventDefinition with id %s", id)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, convertedSpec, err := r.converter.InputFromGraphQL(&in)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting GraphQL input to EventDefinition with id %s", id)
	}

	err = r.svc.Update(ctx, resource.Application, id, *convertedIn, convertedSpec)
	if err != nil {
		return nil, err
	}

	event, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	spec, err := r.specService.GetByReferenceObjectID(ctx, resource.Application, model.EventSpecReference, event.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for EventDefinition with id %q", event.ID)
	}

	bndlRef, err := r.bndlRefSvc.GetForBundle(ctx, model.BundleEventReference, &event.ID, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting bundle reference for EventDefinition with id %q", event.ID)
	}

	gqlEvent, err := r.converter.ToGraphQL(event, spec, bndlRef)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting EventDefinition with id %q to graphQL", event.ID)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("EventDefinition with id %s successfully updated.", id)
	return gqlEvent, nil
}

// UpdateEventDefinitionForApplication updates an EventDefinition for Application without being in a Bundle
func (r *Resolver) UpdateEventDefinitionForApplication(ctx context.Context, id string, in graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Updating EventDefinition with id %s", id)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, convertedSpec, err := r.converter.InputFromGraphQL(&in)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting GraphQL input to EventDefinition with id %s", id)
	}

	err = r.svc.UpdateForApplication(ctx, id, *convertedIn, convertedSpec)
	if err != nil {
		return nil, err
	}

	event, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	spec, err := r.specService.GetByReferenceObjectID(ctx, resource.Application, model.EventSpecReference, event.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for EventDefinition with id %q", event.ID)
	}

	gqlEvent, err := r.converter.ToGraphQL(event, spec, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting EventDefinition with id %q to graphQL", event.ID)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	log.C(ctx).Infof("EventDefinition with id %s successfully updated.", id)
	return gqlEvent, nil
}

// DeleteEventDefinition deletes an EventDefinition by its ID.
func (r *Resolver) DeleteEventDefinition(ctx context.Context, id string) (*graphql.EventDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Deleting EventDefinition with id %s", id)

	ctx = persistence.SaveToContext(ctx, tx)

	event, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	spec, err := r.specService.GetByReferenceObjectID(ctx, resource.Application, model.EventSpecReference, event.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for EventDefinition with id %q", event.ID)
	}

	bndlRef, err := r.bndlRefSvc.GetForBundle(ctx, model.BundleEventReference, &event.ID, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting bundle reference for EventDefinition with id %q", event.ID)
	}

	gqlEvent, err := r.converter.ToGraphQL(event, spec, bndlRef)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting EventDefinition with id %q to graphQL", event.ID)
	}

	err = r.svc.Delete(ctx, resource.Application, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("EventDefinition with id %s successfully deleted.", id)
	return gqlEvent, nil
}

// RefetchEventDefinitionSpec refetches an EventDefinitionSpec for EventDefinition with given ID.
func (r *Resolver) RefetchEventDefinitionSpec(ctx context.Context, eventID string) (*graphql.EventSpec, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Refetching EventDefinitionSpec for EventDefinition with id %s", eventID)

	ctx = persistence.SaveToContext(ctx, tx)

	dbSpec, err := r.specService.GetByReferenceObjectID(ctx, resource.Application, model.EventSpecReference, eventID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for EventDefinition with id %q", eventID)
	}

	if dbSpec == nil {
		return nil, errors.Errorf("spec for Event with id %q not found", eventID)
	}

	spec, err := r.specService.RefetchSpec(ctx, dbSpec.ID, model.EventSpecReference)
	if err != nil {
		return nil, err
	}

	converted, err := r.specConverter.ToGraphQLEventSpec(spec)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Successfully refetched EventDefinitionSpec for EventDefinition with id %s", eventID)
	return converted, nil
}

// FetchRequest returns a FetchRequest by a given EventSpec via dataloaders.
func (r *Resolver) FetchRequest(ctx context.Context, obj *graphql.EventSpec) (*graphql.FetchRequest, error) {
	params := dataloader.ParamFetchRequestEventDef{ID: obj.ID, Ctx: ctx}
	return dataloader.ForFetchRequestEventDef(ctx).FetchRequestEventDefByID.Load(params)
}

// FetchRequestEventDefDataLoader is the dataloader implementation.
func (r *Resolver) FetchRequestEventDefDataLoader(keys []dataloader.ParamFetchRequestEventDef) ([]*graphql.FetchRequest, []error) {
	if len(keys) == 0 {
		return nil, []error{apperrors.NewInternalError("No EventDef specs found")}
	}

	ctx := keys[0].Ctx

	specIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		if key.ID == "" {
			return nil, []error{apperrors.NewInternalError("Cannot fetch FetchRequest. EventDefinition Spec ID is empty")}
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

	err = tx.Commit()
	if err != nil {
		return nil, []error{err}
	}

	log.C(ctx).Infof("Successfully fetched requests for Specifications %v", specIDs)
	return gqlFetchRequests, nil
}
