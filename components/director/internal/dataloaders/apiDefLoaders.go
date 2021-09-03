//go:generate go run github.com/vektah/dataloaden ApiDefLoader ParamApiDef *github.com/kyma-incubator/compass/components/director/pkg/graphql.APIDefinitionPage

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type contextKey string

const loadersKeyApiDef contextKey = "dataloadersApiDef"

type ApiDefLoaders struct {
	ApiDefById ApiDefLoader
}

type ParamApiDef struct {
	ID    string
	Ctx   context.Context
	First *int
	After *graphql.PageCursor
}

func HandlerApiDef(fetchFunc func(keys []ParamApiDef) ([]*graphql.APIDefinitionPage, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeyApiDef, &ApiDefLoaders{
				ApiDefById: ApiDefLoader{
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

func ApiDefFor(ctx context.Context) *ApiDefLoaders {
	return ctx.Value(loadersKeyApiDef).(*ApiDefLoaders)
}
