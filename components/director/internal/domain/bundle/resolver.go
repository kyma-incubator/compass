package mp_bundle

import (
	"context"
	"fmt"
	dataloader "github.com/kyma-incubator/compass/components/director/dataloaders"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery --name=BundleService --output=automock --outpkg=automock --case=underscore
type BundleService interface {
	Create(ctx context.Context, applicationID string, in model.BundleCreateInput) (string, error)
	Update(ctx context.Context, id string, in model.BundleUpdateInput) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.Bundle, error)
}

//go:generate mockery --name=BundleConverter --output=automock --outpkg=automock --case=underscore
type BundleConverter interface {
	ToGraphQL(in *model.Bundle) (*graphql.Bundle, error)
	CreateInputFromGraphQL(in graphql.BundleCreateInput) (model.BundleCreateInput, error)
	UpdateInputFromGraphQL(in graphql.BundleUpdateInput) (*model.BundleUpdateInput, error)
}

//go:generate mockery --name=BundleInstanceAuthService --output=automock --outpkg=automock --case=underscore
type BundleInstanceAuthService interface {
	GetForBundle(ctx context.Context, id string, bundleID string) (*model.BundleInstanceAuth, error)
	List(ctx context.Context, id string) ([]*model.BundleInstanceAuth, error)
}

//go:generate mockery --name=BundleInstanceAuthConverter --output=automock --outpkg=automock --case=underscore
type BundleInstanceAuthConverter interface {
	ToGraphQL(in *model.BundleInstanceAuth) (*graphql.BundleInstanceAuth, error)
	MultipleToGraphQL(in []*model.BundleInstanceAuth) ([]*graphql.BundleInstanceAuth, error)
}

//go:generate mockery --name=APIService --output=automock --outpkg=automock --case=underscore
type APIService interface {
	ListAllByBundleIDs(ctx context.Context, bundleIDs []string, pageSize int, cursor string) ([]*model.APIDefinitionPage, error)
	GetForBundle(ctx context.Context, id string, bundleID string) (*model.APIDefinition, error)
	CreateInBundle(ctx context.Context, appID, bundleID string, in model.APIDefinitionInput, spec *model.SpecInput) (string, error)
	DeleteAllByBundleID(ctx context.Context, bundleID string) error
}

//go:generate mockery --name=APIConverter --output=automock --outpkg=automock --case=underscore
type APIConverter interface {
	ToGraphQL(in *model.APIDefinition, spec *model.Spec, bundleRef *model.BundleReference) (*graphql.APIDefinition, error)
	MultipleToGraphQL(in []*model.APIDefinition, specs []*model.Spec, bundleRefs []*model.BundleReference) ([]*graphql.APIDefinition, error)
	MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) ([]*model.APIDefinitionInput, []*model.SpecInput, error)
}

//go:generate mockery --name=EventService --output=automock --outpkg=automock --case=underscore
type EventService interface {
	ListAllByBundleIDs(ctx context.Context, bundleIDs []string, pageSize int, cursor string) ([]*model.EventDefinitionPage, error)
	GetForBundle(ctx context.Context, id string, bundleID string) (*model.EventDefinition, error)
	CreateInBundle(ctx context.Context, appID, bundleID string, in model.EventDefinitionInput, spec *model.SpecInput) (string, error)
	DeleteAllByBundleID(ctx context.Context, bundleID string) error
}

//go:generate mockery --name=EventConverter --output=automock --outpkg=automock --case=underscore
type EventConverter interface {
	ToGraphQL(in *model.EventDefinition, spec *model.Spec, bundleReference *model.BundleReference) (*graphql.EventDefinition, error)
	MultipleToGraphQL(in []*model.EventDefinition, specs []*model.Spec, bundleRefs []*model.BundleReference) ([]*graphql.EventDefinition, error)
	MultipleInputFromGraphQL(in []*graphql.EventDefinitionInput) ([]*model.EventDefinitionInput, []*model.SpecInput, error)
}

//go:generate mockery --name=DocumentService --output=automock --outpkg=automock --case=underscore
type DocumentService interface {
	GetForBundle(ctx context.Context, id string, bundleID string) (*model.Document, error)
	CreateInBundle(ctx context.Context, bundleID string, in model.DocumentInput) (string, error)
	ListAllByBundleIDs(ctx context.Context, bundleIDs []string, pageSize int, cursor string) ([]*model.DocumentPage, error)
}

//go:generate mockery --name=DocumentConverter --output=automock --outpkg=automock --case=underscore
type DocumentConverter interface {
	ToGraphQL(in *model.Document) *graphql.Document
	MultipleToGraphQL(in []*model.Document) []*graphql.Document
	MultipleInputFromGraphQL(in []*graphql.DocumentInput) ([]*model.DocumentInput, error)
}

//go:generate mockery --name=SpecService --output=automock --outpkg=automock --case=underscore
type SpecService interface {
	CreateByReferenceObjectID(ctx context.Context, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) (string, error)
	UpdateByReferenceObjectID(ctx context.Context, id string, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) error
	GetByReferenceObjectID(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) (*model.Spec, error)
	ListByReferenceObjectIDs(ctx context.Context, objectType model.SpecReferenceObjectType, objectIDs []string) ([]*model.Spec, error)
	RefetchSpec(ctx context.Context, id string) (*model.Spec, error)
}

//go:generate mockery --name=BundleReferenceService --output=automock --outpkg=automock --case=underscore
type BundleReferenceService interface {
	GetForBundle(ctx context.Context, objectType model.BundleReferenceObjectType, objectID, bundleID *string) (*model.BundleReference, error)
	ListAllByBundleIDs(ctx context.Context, objectType model.BundleReferenceObjectType, bundleIDs []string, pageSize int, cursor string) ([]*model.BundleReference, map[string]int, error)
}

type Resolver struct {
	transact persistence.Transactioner

	bundleSvc             BundleService
	bundleInstanceAuthSvc BundleInstanceAuthService
	bundleReferenceSvc    BundleReferenceService
	apiSvc                APIService
	eventSvc              EventService
	documentSvc           DocumentService

	bundleConverter             BundleConverter
	bundleInstanceAuthConverter BundleInstanceAuthConverter
	apiConverter                APIConverter
	eventConverter              EventConverter
	documentConverter           DocumentConverter

	specService SpecService
}

func NewResolver(
	transact persistence.Transactioner,
	bundleSvc BundleService,
	bundleInstanceAuthSvc BundleInstanceAuthService,
	bundleReferenceSvc BundleReferenceService,
	apiSvc APIService,
	eventSvc EventService,
	documentSvc DocumentService,
	bundleConverter BundleConverter,
	bundleInstanceAuthConverter BundleInstanceAuthConverter,
	apiConv APIConverter,
	eventConv EventConverter,
	documentConv DocumentConverter,
	specSerice SpecService) *Resolver {
	return &Resolver{
		transact:                    transact,
		bundleConverter:             bundleConverter,
		bundleSvc:                   bundleSvc,
		bundleInstanceAuthSvc:       bundleInstanceAuthSvc,
		bundleReferenceSvc:          bundleReferenceSvc,
		apiSvc:                      apiSvc,
		eventSvc:                    eventSvc,
		documentSvc:                 documentSvc,
		bundleInstanceAuthConverter: bundleInstanceAuthConverter,
		apiConverter:                apiConv,
		eventConverter:              eventConv,
		documentConverter:           documentConv,
		specService:                 specSerice,
	}
}

func (r *Resolver) AddBundle(ctx context.Context, applicationID string, in graphql.BundleCreateInput) (*graphql.Bundle, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Adding bundle to Application with id %s", applicationID)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, err := r.bundleConverter.CreateInputFromGraphQL(in)
	if err != nil {
		return nil, errors.Wrap(err, "while converting input from GraphQL")
	}

	id, err := r.bundleSvc.Create(ctx, applicationID, convertedIn)
	if err != nil {
		return nil, err
	}

	bndl, err := r.bundleSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	gqlBundle, err := r.bundleConverter.ToGraphQL(bndl)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Bundle with id %s to GraphQL", id)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Bundle with id %s successfully added to Application with id %s", id, applicationID)
	return gqlBundle, nil
}

func (r *Resolver) UpdateBundle(ctx context.Context, id string, in graphql.BundleUpdateInput) (*graphql.Bundle, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Updating Bundle with id %s", id)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, err := r.bundleConverter.UpdateInputFromGraphQL(in)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting converting GraphQL input to Bundle with id %s", id)
	}

	err = r.bundleSvc.Update(ctx, id, *convertedIn)
	if err != nil {
		return nil, err
	}

	bndl, err := r.bundleSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	gqlBndl, err := r.bundleConverter.ToGraphQL(bndl)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Bundle with id %s to GraphQL", id)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Bundle with id %s successfully updated.", id)
	return gqlBndl, nil
}

func (r *Resolver) DeleteBundle(ctx context.Context, id string) (*graphql.Bundle, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Deleting Bundle with id %s", id)

	ctx = persistence.SaveToContext(ctx, tx)

	bndl, err := r.bundleSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = r.apiSvc.DeleteAllByBundleID(ctx, id)
	if err != nil {
		return nil, err
	}

	err = r.eventSvc.DeleteAllByBundleID(ctx, id)
	if err != nil {
		return nil, err
	}

	err = r.bundleSvc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	deletedBndl, err := r.bundleConverter.ToGraphQL(bndl)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Bundle with id %s to GraphQL", id)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Bundle with id %s successfully deleted.", id)
	return deletedBndl, nil
}

func (r *Resolver) InstanceAuth(ctx context.Context, obj *graphql.Bundle, id string) (*graphql.BundleInstanceAuth, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Bundle cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	bndl, err := r.bundleInstanceAuthSvc.GetForBundle(ctx, id, obj.ID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	gqlAuth, err := r.bundleInstanceAuthConverter.ToGraphQL(bndl)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return gqlAuth, nil
}

func (r *Resolver) InstanceAuths(ctx context.Context, obj *graphql.Bundle) ([]*graphql.BundleInstanceAuth, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Bundle cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	bndlInstanceAuths, err := r.bundleInstanceAuthSvc.List(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	gqlAuths, err := r.bundleInstanceAuthConverter.MultipleToGraphQL(bndlInstanceAuths)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return gqlAuths, nil
}

func (r *Resolver) APIDefinition(ctx context.Context, obj *graphql.Bundle, id string) (*graphql.APIDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	api, err := r.apiSvc.GetForBundle(ctx, id, obj.ID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	spec, err := r.specService.GetByReferenceObjectID(ctx, model.APISpecReference, api.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for APIDefinition with id %q", api.ID)
	}

	bndlRef, err := r.bundleReferenceSvc.GetForBundle(ctx, model.BundleAPIReference, &api.ID, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting bundle reference for APIDefinition with id %q", api.ID)
	}

	gqlAPI, err := r.apiConverter.ToGraphQL(api, spec, bndlRef)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting APIDefinition with id %q to graphQL", api.ID)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return gqlAPI, nil
}

func (r *Resolver) APIDefinitions(ctx context.Context, obj *graphql.Bundle, group *string, first *int, after *graphql.PageCursor) (*graphql.APIDefinitionPage, error) {
	param := dataloader.ParamApiDef{ID: obj.ID, Ctx: ctx, First: first, After: after}
	return dataloader.ApiDefFor(ctx).ApiDefById.Load(param)
}

func (r *Resolver) ApiDefinitionsDataLoader(keys []dataloader.ParamApiDef) ([]*graphql.APIDefinitionPage, []error) {
	if len(keys) == 0 {
		return nil, []error{apperrors.NewInternalError("No Bundles found")}
	}

	ctx := keys[0].Ctx
	first := keys[0].First
	after := keys[0].After

	bundleIDs := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		bundleIDs[i] = keys[i].ID
	}

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	if first == nil {
		return nil, []error{apperrors.NewInvalidDataError("missing required parameter 'first'")}
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, []error{err}
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	apiDefPages, err := r.apiSvc.ListAllByBundleIDs(ctx, bundleIDs, *first, cursor)
	if err != nil {
		return nil, []error{err}
	}

	var apiDefIDs []string
	for _, page := range apiDefPages {
		for _, apiDefinition := range page.Data {
			apiDefIDs = append(apiDefIDs, apiDefinition.ID)
		}
	}

	specs, err := r.specService.ListByReferenceObjectIDs(ctx, model.APISpecReference, apiDefIDs)
	if err != nil {
		return nil, []error{err}
	}

	references, _, err := r.bundleReferenceSvc.ListAllByBundleIDs(ctx, model.BundleAPIReference, bundleIDs, *first, cursor)
	if err != nil {
		return nil, []error{err}
	}

	refsByBundleId := map[string][]*model.BundleReference{}
	for _, ref := range references {
		refsByBundleId[*ref.BundleID] = append(refsByBundleId[*ref.BundleID], ref)
	}

	apiDefIDtoSpec := make(map[string]*model.Spec)
	for _, spec := range specs {
		apiDefIDtoSpec[spec.ObjectID] = spec
	}

	var gqlApiDefs []*graphql.APIDefinitionPage
	for i, apisPage := range apiDefPages {
		apiSpecs := make([]*model.Spec, 0, len(apisPage.Data))
		apiBundleRefs := make([]*model.BundleReference, 0, len(apisPage.Data))
		for _, api := range apisPage.Data {
			apiSpecs = append(apiSpecs, apiDefIDtoSpec[api.ID])
			br, err := getBundleReferenceForAPI(api.ID, refsByBundleId[bundleIDs[i]])
			if err != nil {
				return nil, []error{err}
			}
			apiBundleRefs = append(apiBundleRefs, br)
		}

		gqlApis, err := r.apiConverter.MultipleToGraphQL(apisPage.Data, apiSpecs, apiBundleRefs)
		if err != nil {
			return nil, []error{errors.Wrapf(err, "while converting api definitions")}
		}

		gqlApiDefs = append(gqlApiDefs, &graphql.APIDefinitionPage{Data: gqlApis, TotalCount: apisPage.TotalCount, PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(apisPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(apisPage.PageInfo.EndCursor),
			HasNextPage: apisPage.PageInfo.HasNextPage,
		}})
	}

	err = tx.Commit()
	if err != nil {
		return nil, []error{err}
	}

	log.C(ctx).Infof("Successfully fetched api definitions for bundles %v", bundleIDs)
	return gqlApiDefs, nil
}

func (r *Resolver) EventDefinition(ctx context.Context, obj *graphql.Bundle, id string) (*graphql.EventDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	eventAPI, err := r.eventSvc.GetForBundle(ctx, id, obj.ID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	spec, err := r.specService.GetByReferenceObjectID(ctx, model.EventSpecReference, eventAPI.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for EventDefinition with id %q", eventAPI.ID)
	}

	bndlRef, err := r.bundleReferenceSvc.GetForBundle(ctx, model.BundleEventReference, &eventAPI.ID, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting bundle reference for EventDefinition with id %q", eventAPI.ID)
	}

	gqlEvent, err := r.eventConverter.ToGraphQL(eventAPI, spec, bndlRef)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting EventDefinition with id %q to graphQL", eventAPI.ID)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return gqlEvent, nil
}

func (r *Resolver) EventDefinitions(ctx context.Context, obj *graphql.Bundle, group *string, first *int, after *graphql.PageCursor) (*graphql.EventDefinitionPage, error) {
	param := dataloader.ParamEventDef{ID: obj.ID, Ctx: ctx, First: first, After: after}
	return dataloader.EventDefFor(ctx).EventDefById.Load(param)
}

func (r *Resolver) EventDefinitionsDataLoader(keys []dataloader.ParamEventDef) ([]*graphql.EventDefinitionPage, []error) {
	if len(keys) == 0 {
		return nil, []error{apperrors.NewInternalError("No Bundles found")}
	}

	ctx := keys[0].Ctx
	first := keys[0].First
	after := keys[0].After

	bundleIDs := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		bundleIDs[i] = keys[i].ID
	}

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	if first == nil {
		return nil, []error{apperrors.NewInvalidDataError("missing required parameter 'first'")}
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, []error{err}
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	eventAPIDefPages, err := r.eventSvc.ListAllByBundleIDs(ctx, bundleIDs, *first, cursor)
	if err != nil {
		return nil, []error{err}
	}

	var eventAPIDefIDs []string
	for _, page := range eventAPIDefPages {
		for _, eventAPIDefinition := range page.Data {
			eventAPIDefIDs = append(eventAPIDefIDs, eventAPIDefinition.ID)
		}
	}

	specs, err := r.specService.ListByReferenceObjectIDs(ctx, model.EventSpecReference, eventAPIDefIDs)
	if err != nil {
		return nil, []error{err}
	}

	references, _, err := r.bundleReferenceSvc.ListAllByBundleIDs(ctx, model.BundleEventReference, bundleIDs, *first, cursor)
	if err != nil {
		return nil, []error{err}
	}

	eventAPIDefIDtoSpec := make(map[string]*model.Spec)
	for _, spec := range specs {
		eventAPIDefIDtoSpec[spec.ObjectID] = spec
	}

	eventAPIDefIDtoRef := make(map[string]*model.BundleReference)
	for _, reference := range references {
		eventAPIDefIDtoRef[*reference.ObjectID] = reference
	}

	var gqlEventDefs []*graphql.EventDefinitionPage
	for _, eventPage := range eventAPIDefPages {
		eventSpecs := make([]*model.Spec, 0, len(eventPage.Data))
		eventBundleRefs := make([]*model.BundleReference, 0, len(eventPage.Data))
		for _, event := range eventPage.Data {
			eventSpecs = append(eventSpecs, eventAPIDefIDtoSpec[event.ID])
			eventBundleRefs = append(eventBundleRefs, eventAPIDefIDtoRef[event.ID])

		}

		gqlEvents, err := r.eventConverter.MultipleToGraphQL(eventPage.Data, eventSpecs, eventBundleRefs)
		if err != nil {
			return nil, []error{errors.Wrapf(err, "while converting event definitions")}
		}

		gqlEventDefs = append(gqlEventDefs, &graphql.EventDefinitionPage{Data: gqlEvents, TotalCount: eventPage.TotalCount, PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(eventPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(eventPage.PageInfo.EndCursor),
			HasNextPage: eventPage.PageInfo.HasNextPage,
		}})
	}

	err = tx.Commit()
	if err != nil {
		return nil, []error{err}
	}

	log.C(ctx).Infof("Successfully fetched event definitions for bundles %v", bundleIDs)
	return gqlEventDefs, nil
}

func (r *Resolver) Document(ctx context.Context, obj *graphql.Bundle, id string) (*graphql.Document, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	eventAPI, err := r.documentSvc.GetForBundle(ctx, id, obj.ID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.documentConverter.ToGraphQL(eventAPI), nil
}

func (r *Resolver) Documents(ctx context.Context, obj *graphql.Bundle, first *int, after *graphql.PageCursor) (*graphql.DocumentPage, error) {
	param := dataloader.ParamDocument{ID: obj.ID, Ctx: ctx, First: first, After: after}
	return dataloader.DocumentFor(ctx).DocumentById.Load(param)
}

func (r *Resolver) DocumentsDataLoader(keys []dataloader.ParamDocument) ([]*graphql.DocumentPage, []error) {
	if len(keys) == 0 {
		return nil, []error{apperrors.NewInternalError("No Bundles found")}
	}

	ctx := keys[0].Ctx
	first := keys[0].First
	after := keys[0].After

	bundleIDs := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		bundleIDs[i] = keys[i].ID
	}

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	if first == nil {
		return nil, []error{apperrors.NewInvalidDataError("missing required parameter 'first'")}
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, []error{err}
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	documentPages, err := r.documentSvc.ListAllByBundleIDs(ctx, bundleIDs, *first, cursor)
	if err != nil {
		return nil, []error{err}
	}

	err = tx.Commit()
	if err != nil {
		return nil, []error{err}
	}

	var gqlDocumentPages []*graphql.DocumentPage
	for _, page := range documentPages {
		gqlDocuments := r.documentConverter.MultipleToGraphQL(page.Data)

		gqlDocumentPages = append(gqlDocumentPages, &graphql.DocumentPage{
			Data:       gqlDocuments,
			TotalCount: page.TotalCount,
			PageInfo: &graphql.PageInfo{
				StartCursor: graphql.PageCursor(page.PageInfo.StartCursor),
				EndCursor:   graphql.PageCursor(page.PageInfo.EndCursor),
				HasNextPage: page.PageInfo.HasNextPage,
			},
		})
	}

	return gqlDocumentPages, nil
}

func getBundleReferenceForAPI(apiID string, bundleReferences []*model.BundleReference) (*model.BundleReference, error) {
	for _, br := range bundleReferences {
		if str.PtrStrToStr(br.ObjectID) == apiID {
			return br, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("could not find BundleReference for API with id %s", apiID))
}