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

type Labels map[string][]string
