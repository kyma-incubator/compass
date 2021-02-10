package document

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=DocumentService -output=automock -outpkg=automock -case=underscore
type DocumentService interface {
	CreateInBundle(ctx context.Context, bundleID string, in model.DocumentInput) (string, error)
	Get(ctx context.Context, id string) (*model.Document, error)
	Delete(ctx context.Context, id string) error
	GetFetchRequest(ctx context.Context, documentID string) (*model.FetchRequest, error)
}

//go:generate mockery -name=DocumentConverter -output=automock -outpkg=automock -case=underscore
type DocumentConverter interface {
	ToGraphQL(in *model.Document) *graphql.Document
	InputFromGraphQL(in *graphql.DocumentInput) (*model.DocumentInput, error)
	ToEntity(in model.Document) (*Entity, error)
	FromEntity(in Entity) (model.Document, error)
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
	transact    persistence.Transactioner
	svc         DocumentService
	appSvc      ApplicationService
	bndlSvc     BundleService
	converter   DocumentConverter
	frConverter FetchRequestConverter
}

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

	found, err := r.bndlSvc.Exist(ctx, bundleID)
	if err != nil {
		return nil, errors.Wrapf(err, "while checking existence of Bundle")
	}

	if !found {
		return nil, apperrors.NewInvalidDataError("cannot add Document to not existing Bundle")
	}

	id, err := r.svc.CreateInBundle(ctx, bundleID, *convertedIn)
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

func (r *Resolver) FetchRequest(ctx context.Context, obj *graphql.Document) (*graphql.FetchRequest, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Document cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	fr, err := r.svc.GetFetchRequest(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	if fr == nil {
		return nil, tx.Commit()
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.frConverter.ToGraphQL(fr)
}
