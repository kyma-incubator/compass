package runtime

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

func (r *Resolver) Runtimes(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.RuntimePage, error) {
	panic("not implemented")
}

func (r *Resolver) Runtime(ctx context.Context, id string) (*graphql.Runtime, error) {
	panic("not implemented")
}

func (r *Resolver) CreateRuntime(ctx context.Context, in graphql.RuntimeInput) (*graphql.Runtime, error) {
	panic("not implemented")
}
func (r *Resolver) UpdateRuntime(ctx context.Context, id string, in graphql.RuntimeInput) (*graphql.Runtime, error) {
	panic("not implemented")
}
func (r *Resolver) DeleteRuntime(ctx context.Context, id string) (*graphql.Runtime, error) {
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
