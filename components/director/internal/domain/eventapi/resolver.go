package eventapi

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/graphql"
)

type EventAPIService interface{}

type EventAPIConverter interface{}

type Resolver struct {
	svc       EventAPIService
	converter EventAPIConverter
}

func NewResolver(svc EventAPIService) *Resolver {
	return &Resolver{
		svc:       svc,
		converter: &converter{},
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
