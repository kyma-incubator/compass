package eventapi

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

func (r *Resolver) AddEventAPI(ctx context.Context, applicationID string, in gqlschema.EventAPIDefinitionInput) (*gqlschema.EventAPIDefinition, error) {
	panic("not implemented")
}
func (r *Resolver) UpdateEventAPI(ctx context.Context, id string, in gqlschema.EventAPIDefinitionInput) (*gqlschema.EventAPIDefinition, error) {
	panic("not implemented")
}
func (r *Resolver) DeleteEventAPI(ctx context.Context, id string) (*gqlschema.EventAPIDefinition, error) {
	panic("not implemented")
}
func (r *Resolver) RefetchEventAPISpec(ctx context.Context, eventID string) (*gqlschema.EventAPISpec, error) {
	panic("not implemented")
}
