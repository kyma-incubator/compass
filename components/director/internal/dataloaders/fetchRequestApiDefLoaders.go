//go:generate go run github.com/vektah/dataloaden FetchRequestApiDefLoader ParamFetchRequestApiDef *github.com/kyma-incubator/compass/components/director/pkg/graphql.FetchRequest

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeyFetchRequestApiDef contextKey = "dataloadersFetchRequestApiDef"

type LoadersFetchRequestApiDef struct {
	FetchRequestApiDefByID FetchRequestApiDefLoader
}

type ParamFetchRequestApiDef struct {
	ID  string
	Ctx context.Context
}

func HandlerFetchRequestApiDef(fetchFunc func(keys []ParamFetchRequestApiDef) ([]*graphql.FetchRequest, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeyFetchRequestApiDef, &LoadersFetchRequestApiDef{
				FetchRequestApiDefByID: FetchRequestApiDefLoader{
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

func ForFetchRequestApiDef(ctx context.Context) *LoadersFetchRequestApiDef {
	return ctx.Value(loadersKeyFetchRequestApiDef).(*LoadersFetchRequestApiDef)
}
