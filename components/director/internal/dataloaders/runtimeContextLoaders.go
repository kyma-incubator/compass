//go:generate go run github.com/vektah/dataloaden RuntimeContextLoader ParamRuntimeContext *github.com/kyma-incubator/compass/components/director/pkg/graphql.RuntimeContextPage

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeyRuntimeContext contextKey = "dataloadersRuntimeContext"

// RuntimeContextLoaders missing godoc
type RuntimeContextLoaders struct {
	RuntimeContextByID RuntimeContextLoader
}

// ParamRuntimeContext missing godoc
type ParamRuntimeContext struct {
	ID    string
	First *int
	After *graphql.PageCursor
	Ctx   context.Context
}

// HandlerRuntimeContext missing godoc
func HandlerRuntimeContext(fetchFunc func(keys []ParamRuntimeContext) ([]*graphql.RuntimeContextPage, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeyRuntimeContext, &RuntimeContextLoaders{
				RuntimeContextByID: RuntimeContextLoader{
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

// RuntimeContextFor missing godoc
func RuntimeContextFor(ctx context.Context) *RuntimeContextLoaders {
	return ctx.Value(loadersKeyRuntimeContext).(*RuntimeContextLoaders)
}
