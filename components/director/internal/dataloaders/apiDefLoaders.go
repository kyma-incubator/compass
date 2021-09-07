//go:generate go run github.com/vektah/dataloaden APIDefLoader ParamAPIDef *github.com/kyma-incubator/compass/components/director/pkg/graphql.APIDefinitionPage

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type contextKey string

const loadersKeyAPIDef contextKey = "dataloadersAPIDef"

// APIDefLoaders missing godoc
type APIDefLoaders struct {
	APIDefByID APIDefLoader
}

// ParamAPIDef missing godoc
type ParamAPIDef struct {
	ID    string
	Ctx   context.Context
	First *int
	After *graphql.PageCursor
}

// HandlerAPIDef missing godoc
func HandlerAPIDef(fetchFunc func(keys []ParamAPIDef) ([]*graphql.APIDefinitionPage, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeyAPIDef, &APIDefLoaders{
				APIDefByID: APIDefLoader{
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

// APIDefFor missing godoc
func APIDefFor(ctx context.Context) *APIDefLoaders {
	return ctx.Value(loadersKeyAPIDef).(*APIDefLoaders)
}
