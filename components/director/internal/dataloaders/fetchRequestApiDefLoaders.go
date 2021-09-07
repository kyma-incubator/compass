//go:generate go run github.com/vektah/dataloaden FetchRequestAPIDefLoader ParamFetchRequestAPIDef *github.com/kyma-incubator/compass/components/director/pkg/graphql.FetchRequest

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeyFetchRequestAPIDef contextKey = "dataloadersFetchRequestAPIDef"

type LoadersFetchRequestAPIDef struct {
	FetchRequestAPIDefByID FetchRequestAPIDefLoader
}

type ParamFetchRequestAPIDef struct {
	ID  string
	Ctx context.Context
}

func HandlerFetchRequestAPIDef(fetchFunc func(keys []ParamFetchRequestAPIDef) ([]*graphql.FetchRequest, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeyFetchRequestAPIDef, &LoadersFetchRequestAPIDef{
				FetchRequestAPIDefByID: FetchRequestAPIDefLoader{
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

func ForFetchRequestAPIDef(ctx context.Context) *LoadersFetchRequestAPIDef {
	return ctx.Value(loadersKeyFetchRequestAPIDef).(*LoadersFetchRequestAPIDef)
}
