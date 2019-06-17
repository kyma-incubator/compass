package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/graphql"
)

type APIService interface{}

type APIConverter interface{}

type Resolver struct {
	svc       APIService
	converter APIConverter
}

func NewResolver(svc APIService) *Resolver {
	return &Resolver{
		svc:       svc,
		converter: &converter{},
	}
}

func (r *Resolver) AddAPI(ctx context.Context, applicationID string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	panic("not implemented")
}
func (r *Resolver) UpdateAPI(ctx context.Context, id string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	panic("not implemented")
}
func (r *Resolver) DeleteAPI(ctx context.Context, id string) (*graphql.APIDefinition, error) {
	panic("not implemented")
}
func (r *Resolver) RefetchAPISpec(ctx context.Context, apiID string) (*graphql.APISpec, error) {
	panic("not implemented")
}

func (r *Resolver) SetAPIAuth(ctx context.Context, apiID string, runtimeID string, in graphql.AuthInput) (*graphql.RuntimeAuth, error) {
	panic("not implemented")
}
func (r *Resolver) DeleteAPIAuth(ctx context.Context, apiID string, runtimeID string) (*graphql.RuntimeAuth, error) {
	panic("not implemented")
}
