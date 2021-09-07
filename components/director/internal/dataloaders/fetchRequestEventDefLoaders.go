//go:generate go run github.com/vektah/dataloaden FetchRequestEventDefLoader ParamFetchRequestEventDef *github.com/kyma-incubator/compass/components/director/pkg/graphql.FetchRequest

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeyFetchRequestEventDef contextKey = "dataloadersFetchRequestEventDef"

// LoadersFetchRequestEventDef missing godoc
type LoadersFetchRequestEventDef struct {
	FetchRequestEventDefByID FetchRequestEventDefLoader
}

// ParamFetchRequestEventDef missing godoc
type ParamFetchRequestEventDef struct {
	ID  string
	Ctx context.Context
}

// HandlerFetchRequestEventDef missing godoc
func HandlerFetchRequestEventDef(fetchFunc func(keys []ParamFetchRequestEventDef) ([]*graphql.FetchRequest, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeyFetchRequestEventDef, &LoadersFetchRequestEventDef{
				FetchRequestEventDefByID: FetchRequestEventDefLoader{
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

// ForFetchRequestEventDef missing godoc
func ForFetchRequestEventDef(ctx context.Context) *LoadersFetchRequestEventDef {
	return ctx.Value(loadersKeyFetchRequestEventDef).(*LoadersFetchRequestEventDef)
}
