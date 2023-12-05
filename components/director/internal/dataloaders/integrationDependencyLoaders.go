//go:generate go run github.com/vektah/dataloaden IntegrationDependencyLoader ParamIntegrationDependency *github.com/kyma-incubator/compass/components/director/pkg/graphql.IntegrationDependencyPage

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeyIntegrationDependency contextKey = "dataloadersIntegrationDependency"

// IntegrationDependencyLoaders missing godoc
type IntegrationDependencyLoaders struct {
	IntegrationDependencyByID IntegrationDependencyLoader
}

// ParamIntegrationDependency missing godoc
type ParamIntegrationDependency struct {
	ID    string
	First *int
	After *graphql.PageCursor
	Ctx   context.Context
}

// HandlerIntegrationDependency missing godoc
func HandlerIntegrationDependency(fetchFunc func(keys []ParamIntegrationDependency) ([]*graphql.IntegrationDependencyPage, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeyIntegrationDependency, &IntegrationDependencyLoaders{
				IntegrationDependencyByID: IntegrationDependencyLoader{
					maxBatch: maxBatch,
					wait:     wait,
					fetch:    fetchFunc,
				},
			})
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// IntegrationDependencyFor missing godoc
func IntegrationDependencyFor(ctx context.Context) *IntegrationDependencyLoaders {
	return ctx.Value(loadersKeyIntegrationDependency).(*IntegrationDependencyLoaders)
}
