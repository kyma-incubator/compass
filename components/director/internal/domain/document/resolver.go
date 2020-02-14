package document

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/package/mock"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=DocumentService -output=automock -outpkg=automock -case=underscore
type DocumentService interface {
	Create(ctx context.Context, applicationID string, in model.DocumentInput) (string, error)
	Get(ctx context.Context, id string) (*model.Document, error)
	Delete(ctx context.Context, id string) error
	GetFetchRequest(ctx context.Context, documentID string) (*model.FetchRequest, error)
}

//go:generate mockery -name=DocumentConverter -output=automock -outpkg=automock -case=underscore
type DocumentConverter interface {
	ToGraphQL(in *model.Document) *graphql.Document
	InputFromGraphQL(in *graphql.DocumentInput) *model.DocumentInput
	ToEntity(in model.Document) (Entity, error)
	FromEntity(in Entity) (model.Document, error)
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
	transact    persistence.Transactioner
	svc         DocumentService
	appSvc      ApplicationService
	converter   DocumentConverter
	frConverter FetchRequestConverter
}

func NewResolver(transact persistence.Transactioner, svc DocumentService, appSvc ApplicationService, frConverter FetchRequestConverter) *Resolver {
	return &Resolver{
		transact:    transact,
		svc:         svc,
		appSvc:      appSvc,
		frConverter: frConverter,
		converter:   &converter{frConverter: frConverter},
	}
}

func (r *Resolver) AddDocument(ctx context.Context, applicationID string, in graphql.DocumentInput) (*graphql.Document, error) {
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
		return nil, errors.New("Cannot add Document to not existing Application")
	}

	id, err := r.svc.Create(ctx, applicationID, *convertedIn)
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
	defer r.transact.RollbackUnlessCommited(tx)

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
		return nil, errors.New("Document cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	fr, err := r.svc.GetFetchRequest(ctx, obj.ID)
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

// TODO: Replace with real implementation
func (r *Resolver) AddDocumentToPackage(ctx context.Context, packageID string, in graphql.DocumentInput) (*graphql.Document, error) {
	return mock.FixDocument("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), nil
}
