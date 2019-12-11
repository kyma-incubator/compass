package director

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type CreateRuntimeResponse struct {
	Result *graphql.Runtime `json:"result"`
}

type DeleteRuntimeResponse struct {
	Result *graphql.Runtime `json:"result"`
}

type UpdateRuntimeResponse struct {
	Result *graphql.Runtime `json:"result"`
}

type RuntimeInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Labels      *Labels `json:"labels"`
}

// such API will be supported

//func (r *mutationResolver) CreateRuntime(ctx context.Context, in graphql.RuntimeInput) (*graphql.Runtime, error) {
//	return r.runtime.CreateRuntime(ctx, in)
//}
//func (r *mutationResolver) UpdateRuntime(ctx context.Context, id string, in graphql.RuntimeInput) (*graphql.Runtime, error) {
//	return r.runtime.UpdateRuntime(ctx, id, in)
//}
//func (r *mutationResolver) DeleteRuntime(ctx context.Context, id string) (*graphql.Runtime, error) {
//	return r.runtime.DeleteRuntime(ctx, id)
//}



type Labels map[string][]string
