package resolver

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/gqlschema"
)

type runtimeResolver struct {

}

func (r *runtimeResolver) Runtimes(ctx context.Context, filter []*gqlschema.LabelFilter, first *int, after *string) (*gqlschema.RuntimePage, error) {
	panic("not implemented")
}

func (r *runtimeResolver) Runtime(ctx context.Context, id string) (*gqlschema.Runtime, error) {
	panic("not implemented")
}

func (r *runtimeResolver) CreateRuntime(ctx context.Context, in gqlschema.RuntimeInput) (*gqlschema.Runtime, error) {

}
func (r *runtimeResolver) UpdateRuntime(ctx context.Context, id string, in gqlschema.RuntimeInput) (*gqlschema.Runtime, error) {

}
func (r *runtimeResolver) DeleteRuntime(ctx context.Context, id string) (*gqlschema. Runtime, error) {

}
func (r *runtimeResolver) AddRuntimeLabel(ctx context.Context, runtimeID string, key string, values []string) ([]string, error) {

}
func (r *runtimeResolver) DeleteRuntimeLabel(ctx context.Context, id string, key string, values []string) ([]string, error) {

}
func (r *runtimeResolver) AddRuntimeAnnotation(ctx context.Context, runtimeID string, key string, value string) (string, error) {

}
func (r *runtimeResolver) DeleteRuntimeAnnotation(ctx context.Context, id string, key string) (*string, error) {

}
