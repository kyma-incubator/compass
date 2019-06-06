package resolver

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/gqlschema"
)

type apiResolver struct {

}

func (r *apiResolver) AddAPI(ctx context.Context, applicationID string, in gqlschema.APIDefinitionInput) (*gqlschema.APIDefinition, error) {
	panic("not implemented")
}
func (r *apiResolver) UpdateAPI(ctx context.Context, id string, in gqlschema.APIDefinitionInput) (*gqlschema.APIDefinition, error) {
	panic("not implemented")
}
func (r *apiResolver) DeleteAPI(ctx context.Context, id string) (*gqlschema.APIDefinition, error) {
	panic("not implemented")
}
func (r *apiResolver) RefetchAPISpec(ctx context.Context, apiID string) (*gqlschema.APISpec, error) {
	panic("not implemented")
}

func (r *apiResolver) SetAPIAuth(ctx context.Context, apiID string, runtimeID string, in gqlschema.AuthInput) (*gqlschema.RuntimeAuth, error) {
	panic("not implemented")
}
func (r *apiResolver) DeleteAPIAuth(ctx context.Context, apiID string, runtimeID string) (*gqlschema.RuntimeAuth, error) {
	panic("not implemented")
}