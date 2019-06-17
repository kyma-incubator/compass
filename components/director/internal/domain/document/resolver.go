package document

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/graphql"
)

type DocumentService interface{}

type DocumentConverter interface{}

type Resolver struct {
	svc       DocumentService
	converter DocumentConverter
}

func NewResolver(svc DocumentService) *Resolver {
	return &Resolver{
		svc:       svc,
		converter: &converter{},
	}
}

func (r *Resolver) AddDocument(ctx context.Context, applicationID string, in graphql.DocumentInput) (*graphql.Document, error) {
	panic("not implemented")
}
func (r *Resolver) DeleteDocument(ctx context.Context, id string) (*graphql.Document, error) {
	panic("not implemented")
}
