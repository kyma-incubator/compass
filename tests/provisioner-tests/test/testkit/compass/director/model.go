package director

import directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type Runtime struct {
	ID          string                        `json:"id"`
	Name        string                        `json:"name"`
	Description *string                       `json:"description"`
	Status      *directorSchema.RuntimeStatus `json:"status"`
	Labels      directorSchema.Labels         `json:"labels"`
}
