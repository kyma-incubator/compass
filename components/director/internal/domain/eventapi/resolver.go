package eventapi

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/graphql"
)

type svc interface{}

type Resolver struct {
	svc       svc
	converter *Converter
}

func NewResolver(svc svc) *Resolver {
	return &Resolver{
		svc:       svc,
		converter: &Converter{},
	}
}

func (r *Resolver) AddEventAPI(ctx context.Context, applicationID string, in graphql.EventAPIDefinitionInput) (*graphql.EventAPIDefinition, error) {
	panic("not implemented")
}
func (r *Resolver) UpdateEventAPI(ctx context.Context, id string, in graphql.EventAPIDefinitionInput) (*graphql.EventAPIDefinition, error) {
	panic("not implemented")
}
func (r *Resolver) DeleteEventAPI(ctx context.Context, id string) (*graphql.EventAPIDefinition, error) {
	panic("not implemented")
}
func (r *Resolver) RefetchEventAPISpec(ctx context.Context, eventID string) (*graphql.EventAPISpec, error) {
	panic("not implemented")
}
