package resolver

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/gqlschema"
)

type eventAPIResolver struct {

}

func (r *eventAPIResolver) AddEventAPI(ctx context.Context, applicationID string, in gqlschema.EventAPIDefinitionInput) (*gqlschema.EventAPIDefinition, error) {
	panic("not implemented")
}
func (r *eventAPIResolver) UpdateEventAPI(ctx context.Context, id string, in gqlschema.EventAPIDefinitionInput) (*gqlschema.EventAPIDefinition, error) {
	panic("not implemented")
}
func (r *eventAPIResolver) DeleteEventAPI(ctx context.Context, id string) (*gqlschema.EventAPIDefinition, error) {
	panic("not implemented")
}
func (r *eventAPIResolver) RefetchEventAPISpec(ctx context.Context, eventID string) (*gqlschema.EventAPISpec, error) {
	panic("not implemented")
}