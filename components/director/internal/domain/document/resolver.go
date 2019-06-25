package document

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

//go:generate mockery -name=DocumentService -output=automock -outpkg=automock -case=underscore
type DocumentService interface {
	Create(ctx context.Context, applicationID string, in model.DocumentInput) (string, error)
	Get(ctx context.Context, id string) (*model.Document, error)
	Delete(ctx context.Context, id string) error
}

//go:generate mockery -name=DocumentConverter -output=automock -outpkg=automock -case=underscore
type DocumentConverter interface {
	ToGraphQL(in *model.Document) *graphql.Document
	InputFromGraphQL(in *graphql.DocumentInput) *model.DocumentInput
}

type Resolver struct {
	svc       DocumentService
	converter DocumentConverter
}

func NewResolver(svc DocumentService, frConverter FetchRequestConverter) *Resolver {
	return &Resolver{
		svc:       svc,
		converter: &converter{frConverter: frConverter},
	}
}

func (r *Resolver) AddDocument(ctx context.Context, applicationID string, in graphql.DocumentInput) (*graphql.Document, error) {
	convertedIn := r.converter.InputFromGraphQL(&in)

	id, err := r.svc.Create(ctx, applicationID, *convertedIn)
	if err != nil {
		return nil, err
	}

	document, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	gqlDocument := r.converter.ToGraphQL(document)

	return gqlDocument, nil
}

func (r *Resolver) DeleteDocument(ctx context.Context, id string) (*graphql.Document, error) {
	document, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	deletedDocument := r.converter.ToGraphQL(document)

	err = r.svc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	return deletedDocument, nil
}
