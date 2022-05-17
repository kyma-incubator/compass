package document

import (
	"context"

	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// DocumentService missing godoc
//go:generate mockery --name=DocumentService --output=automock --outpkg=automock --case=underscore --disable-version-string
type DocumentService interface {
	CreateInBundle(ctx context.Context, appID, bundleID string, in model.DocumentInput) (string, error)
	Get(ctx context.Context, id string) (*model.Document, error)
	Delete(ctx context.Context, id string) error
	ListFetchRequests(ctx context.Context, documentIDs []string) ([]*model.FetchRequest, error)
}

// DocumentConverter missing godoc
//go:generate mockery --name=DocumentConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type DocumentConverter interface {
	ToGraphQL(in *model.Document) *graphql.Document
	InputFromGraphQL(in *graphql.DocumentInput) (*model.DocumentInput, error)
}

// FetchRequestConverter missing godoc
//go:generate mockery --name=FetchRequestConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type FetchRequestConverter interface {
	ToGraphQL(in *model.FetchRequest) (*graphql.FetchRequest, error)
	InputFromGraphQL(in *graphql.FetchRequestInput) (*model.FetchRequestInput, error)
}

// ApplicationService missing godoc
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationService interface {
	Exist(ctx context.Context, id string) (bool, error)
}

// BundleService missing godoc
//go:generate mockery --name=BundleService --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleService interface {
	Get(ctx context.Context, id string) (*model.Bundle, error)
}

// Resolver missing godoc
type Resolver struct {
	transact    persistence.Transactioner
	svc         DocumentService
	appSvc      ApplicationService
	bndlSvc     BundleService
	converter   DocumentConverter
	frConverter FetchRequestConverter
}

// NewResolver missing godoc
func NewResolver(transact persistence.Transactioner, svc DocumentService, appSvc ApplicationService, bndlSvc BundleService, frConverter FetchRequestConverter) *Resolver {
	return &Resolver{
		transact:    transact,
		svc:         svc,
		appSvc:      appSvc,
		bndlSvc:     bndlSvc,
		frConverter: frConverter,
		converter:   &converter{frConverter: frConverter},
	}
}

// AddDocumentToBundle missing godoc
func (r *Resolver) AddDocumentToBundle(ctx context.Context, bundleID string, in graphql.DocumentInput) (*graphql.Document, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, err := r.converter.InputFromGraphQL(&in)
	if err != nil {
		return nil, errors.Wrap(err, "while converting DocumentInput from GraphQL")
	}

	bndl, err := r.bndlSvc.Get(ctx, bundleID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, apperrors.NewInvalidDataError("cannot add Document to not existing Bundle")
		}
		return nil, errors.Wrapf(err, "while getting bundle %s", bundleID)
	}

	id, err := r.svc.CreateInBundle(ctx, bndl.ApplicationID, bundleID, *convertedIn)
	if err != nil {
		return nil, err
	}

	document, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlDocument := r.converter.ToGraphQL(document)

	return gqlDocument, nil
}

// DeleteDocument missing godoc
func (r *Resolver) DeleteDocument(ctx context.Context, id string) (*graphql.Document, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	document, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	deletedDocument := r.converter.ToGraphQL(document)

	err = r.svc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return deletedDocument, nil
}

// FetchRequest missing godoc
func (r *Resolver) FetchRequest(ctx context.Context, obj *graphql.Document) (*graphql.FetchRequest, error) {
	params := dataloader.ParamFetchRequestDocument{ID: obj.ID, Ctx: ctx}
	return dataloader.ForFetchRequestDocument(ctx).FetchRequestDocumentByID.Load(params)
}

// FetchRequestDocumentDataLoader missing godoc
func (r *Resolver) FetchRequestDocumentDataLoader(keys []dataloader.ParamFetchRequestDocument) ([]*graphql.FetchRequest, []error) {
	if len(keys) == 0 {
		return nil, []error{apperrors.NewInternalError("No Documents found")}
	}

	ctx := keys[0].Ctx
	documentIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		if key.ID == "" {
			return nil, []error{apperrors.NewInternalError("Cannot fetch FetchRequest. Document ID is empty")}
		}
		documentIDs = append(documentIDs, key.ID)
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, []error{err}
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	fetchRequests, err := r.svc.ListFetchRequests(ctx, documentIDs)
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

	log.C(ctx).Infof("Successfully fetched requests for Documents %v", documentIDs)
	return gqlFetchRequests, nil
}
