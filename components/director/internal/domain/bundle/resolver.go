package mp_bundle

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=BundleService -output=automock -outpkg=automock -case=underscore
type BundleService interface {
	Create(ctx context.Context, applicationID string, in model.BundleCreateInput) (string, error)
	Update(ctx context.Context, id string, in model.BundleUpdateInput) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.Bundle, error)
}

//go:generate mockery -name=BundleConverter -output=automock -outpkg=automock -case=underscore
type BundleConverter interface {
	ToGraphQL(in *model.Bundle) (*graphql.Bundle, error)
	CreateInputFromGraphQL(in graphql.BundleCreateInput) (model.BundleCreateInput, error)
	UpdateInputFromGraphQL(in graphql.BundleUpdateInput) (*model.BundleUpdateInput, error)
}

//go:generate mockery -name=BundleInstanceAuthService -output=automock -outpkg=automock -case=underscore
type BundleInstanceAuthService interface {
	GetForBundle(ctx context.Context, id string, bundleID string) (*model.BundleInstanceAuth, error)
	List(ctx context.Context, id string) ([]*model.BundleInstanceAuth, error)
}

//go:generate mockery -name=BundleInstanceAuthConverter -output=automock -outpkg=automock -case=underscore
type BundleInstanceAuthConverter interface {
	ToGraphQL(in *model.BundleInstanceAuth) (*graphql.BundleInstanceAuth, error)
	MultipleToGraphQL(in []*model.BundleInstanceAuth) ([]*graphql.BundleInstanceAuth, error)
}

//go:generate mockery -name=APIService -output=automock -outpkg=automock -case=underscore
type APIService interface {
	ListForBundle(ctx context.Context, bundleID string, pageSize int, cursor string) (*model.APIDefinitionPage, error)
	GetForBundle(ctx context.Context, id string, bundleID string) (*model.APIDefinition, error)
	CreateInBundle(ctx context.Context, bundleID string, in model.APIDefinitionInput, spec *model.SpecInput) (string, error)
}

//go:generate mockery -name=APIConverter -output=automock -outpkg=automock -case=underscore
type APIConverter interface {
	ToGraphQL(in *model.APIDefinition, spec *model.Spec) (*graphql.APIDefinition, error)
	MultipleToGraphQL(in []*model.APIDefinition, specs []*model.Spec) ([]*graphql.APIDefinition, error)
	MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) ([]*model.APIDefinitionInput, []*model.SpecInput, error)
}

//go:generate mockery -name=EventService -output=automock -outpkg=automock -case=underscore
type EventService interface {
	ListForBundle(ctx context.Context, bundleID string, pageSize int, cursor string) (*model.EventDefinitionPage, error)
	GetForBundle(ctx context.Context, id string, bundleID string) (*model.EventDefinition, error)
	CreateInBundle(ctx context.Context, bundleID string, in model.EventDefinitionInput, spec *model.SpecInput) (string, error)
}

//go:generate mockery -name=EventConverter -output=automock -outpkg=automock -case=underscore
type EventConverter interface {
	ToGraphQL(in *model.EventDefinition, spec *model.Spec) (*graphql.EventDefinition, error)
	MultipleToGraphQL(in []*model.EventDefinition, specs []*model.Spec) ([]*graphql.EventDefinition, error)
	MultipleInputFromGraphQL(in []*graphql.EventDefinitionInput) ([]*model.EventDefinitionInput, []*model.SpecInput, error)
}

//go:generate mockery -name=DocumentService -output=automock -outpkg=automock -case=underscore
type DocumentService interface {
	ListForBundle(ctx context.Context, bundleID string, pageSize int, cursor string) (*model.DocumentPage, error)
	GetForBundle(ctx context.Context, id string, bundleID string) (*model.Document, error)
	CreateInBundle(ctx context.Context, bundleID string, in model.DocumentInput) (string, error)
}

//go:generate mockery -name=DocumentConverter -output=automock -outpkg=automock -case=underscore
type DocumentConverter interface {
	ToGraphQL(in *model.Document) *graphql.Document
	MultipleToGraphQL(in []*model.Document) []*graphql.Document
	MultipleInputFromGraphQL(in []*graphql.DocumentInput) ([]*model.DocumentInput, error)
}

//go:generate mockery -name=SpecService -output=automock -outpkg=automock -case=underscore
type SpecService interface {
	CreateByReferenceObjectID(ctx context.Context, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) (string, error)
	UpdateByReferenceObjectID(ctx context.Context, id string, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) error
	GetByReferenceObjectID(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) (*model.Spec, error)
	RefetchSpec(ctx context.Context, id string) (*model.Spec, error)
}

type Resolver struct {
	transact persistence.Transactioner

	bundleSvc             BundleService
	bundleInstanceAuthSvc BundleInstanceAuthService
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

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.bundleInstanceAuthConverter.ToGraphQL(bndl)

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

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.bundleInstanceAuthConverter.MultipleToGraphQL(bndlInstanceAuths)
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

	gqlAPI, err := r.apiConverter.ToGraphQL(api, spec)
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
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter 'first'")
	}

	apisPage, err := r.apiSvc.ListForBundle(ctx, obj.ID, *first, cursor)
	if err != nil {
		return nil, err
	}

	apiSpecs := make([]*model.Spec, 0, len(apisPage.Data))
	for _, api := range apisPage.Data {
		spec, err := r.specService.GetByReferenceObjectID(ctx, model.APISpecReference, api.ID)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting spec for APIDefinition with id %q", api.ID)
		}

		apiSpecs = append(apiSpecs, spec)
	}

	gqlApis, err := r.apiConverter.MultipleToGraphQL(apisPage.Data, apiSpecs)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting apis")
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &graphql.APIDefinitionPage{
		Data:       gqlApis,
		TotalCount: apisPage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(apisPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(apisPage.PageInfo.EndCursor),
			HasNextPage: apisPage.PageInfo.HasNextPage,
		},
	}, nil
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

	gqlEvent, err := r.eventConverter.ToGraphQL(eventAPI, spec)
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
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter 'first'")
	}

	eventAPIPage, err := r.eventSvc.ListForBundle(ctx, obj.ID, *first, cursor)
	if err != nil {
		return nil, err
	}

	eventSpecs := make([]*model.Spec, 0, len(eventAPIPage.Data))
	for _, event := range eventAPIPage.Data {
		spec, err := r.specService.GetByReferenceObjectID(ctx, model.EventSpecReference, event.ID)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting spec for EventDefinition with id %q", event.ID)
		}

		eventSpecs = append(eventSpecs, spec)
	}

	gqlEvents, err := r.eventConverter.MultipleToGraphQL(eventAPIPage.Data, eventSpecs)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting events")
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &graphql.EventDefinitionPage{
		Data:       gqlEvents,
		TotalCount: eventAPIPage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(eventAPIPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(eventAPIPage.PageInfo.EndCursor),
			HasNextPage: eventAPIPage.PageInfo.HasNextPage,
		},
	}, nil
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
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter 'first'")
	}

	documentsPage, err := r.documentSvc.ListForBundle(ctx, obj.ID, *first, cursor)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlDocuments := r.documentConverter.MultipleToGraphQL(documentsPage.Data)

	return &graphql.DocumentPage{
		Data:       gqlDocuments,
		TotalCount: documentsPage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(documentsPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(documentsPage.PageInfo.EndCursor),
			HasNextPage: documentsPage.PageInfo.HasNextPage,
		},
	}, nil
}
