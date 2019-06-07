package api

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/gqlschema"
)

type Resolver struct {
	svc       *Service
	converter *Converter
}

func NewResolver(svc *Service) *Resolver {
	return &Resolver{
		svc:       svc,
		converter: &Converter{},
	}
}

func (r *Resolver) AddAPI(ctx context.Context, applicationID string, in gqlschema.APIDefinitionInput) (*gqlschema.APIDefinition, error) {
	panic("not implemented")
}
func (r *Resolver) UpdateAPI(ctx context.Context, id string, in gqlschema.APIDefinitionInput) (*gqlschema.APIDefinition, error) {
	panic("not implemented")
}
func (r *Resolver) DeleteAPI(ctx context.Context, id string) (*gqlschema.APIDefinition, error) {
	panic("not implemented")
}
func (r *Resolver) RefetchAPISpec(ctx context.Context, apiID string) (*gqlschema.APISpec, error) {
	panic("not implemented")
}

func (r *Resolver) SetAPIAuth(ctx context.Context, apiID string, runtimeID string, in gqlschema.AuthInput) (*gqlschema.RuntimeAuth, error) {
	panic("not implemented")
}
func (r *Resolver) DeleteAPIAuth(ctx context.Context, apiID string, runtimeID string) (*gqlschema.RuntimeAuth, error) {
	panic("not implemented")
}
