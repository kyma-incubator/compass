package eventapi

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
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
	// TODO panic("not implemented")
	return &graphql.EventAPIDefinition{}, nil
}
func (r *Resolver) UpdateEventAPI(ctx context.Context, id string, in graphql.EventAPIDefinitionInput) (*graphql.EventAPIDefinition, error) {
	//TODO panic("not implemented")
	return &graphql.EventAPIDefinition{}, nil
}
func (r *Resolver) DeleteEventAPI(ctx context.Context, id string) (*graphql.EventAPIDefinition, error) {
	//TODO panic("not implemented")
	return &graphql.EventAPIDefinition{}, nil
}
func (r *Resolver) RefetchEventAPISpec(ctx context.Context, eventID string) (*graphql.EventAPISpec, error) {
	//TODO panic("not implemented")
	return &graphql.EventAPISpec{}, nil
}
