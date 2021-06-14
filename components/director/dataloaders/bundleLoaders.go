//go:generate go run github.com/vektah/dataloaden BundleLoader Param *github.com/kyma-incubator/compass/components/director/pkg/graphql.BundlePage

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKey = "dataloaders"

type Loaders struct {
	BundleById BundleLoader
}

type Param struct {
	ID    string
	First *int
	After *graphql.PageCursor
	Ctx   context.Context
}

func Handler(fetchFunc func(keys []Param) ([]*graphql.BundlePage, []error)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKey, &Loaders{
				BundleById: BundleLoader{
					maxBatch: 100,
					wait:     1 * time.Millisecond,
					fetch:    fetchFunc,
				},
			})
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}