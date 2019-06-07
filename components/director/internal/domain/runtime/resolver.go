package runtime

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

func (r *Resolver) Runtimes(ctx context.Context, filter []*gqlschema.LabelFilter, first *int, after *string) (*gqlschema.RuntimePage, error) {
	panic("not implemented")
}

func (r *Resolver) Runtime(ctx context.Context, id string) (*gqlschema.Runtime, error) {
	panic("not implemented")
}

func (r *Resolver) CreateRuntime(ctx context.Context, in gqlschema.RuntimeInput) (*gqlschema.Runtime, error) {
	panic("not implemented")
}
func (r *Resolver) UpdateRuntime(ctx context.Context, id string, in gqlschema.RuntimeInput) (*gqlschema.Runtime, error) {
	panic("not implemented")
}
func (r *Resolver) DeleteRuntime(ctx context.Context, id string) (*gqlschema.Runtime, error) {
	panic("not implemented")
}
func (r *Resolver) AddRuntimeLabel(ctx context.Context, runtimeID string, key string, values []string) ([]string, error) {
	panic("not implemented")
}
func (r *Resolver) DeleteRuntimeLabel(ctx context.Context, id string, key string, values []string) ([]string, error) {
	panic("not implemented")
}
func (r *Resolver) AddRuntimeAnnotation(ctx context.Context, runtimeID string, key string, value string) (string, error) {
	panic("not implemented")
}
func (r *Resolver) DeleteRuntimeAnnotation(ctx context.Context, id string, key string) (*string, error) {
	panic("not implemented")
}
