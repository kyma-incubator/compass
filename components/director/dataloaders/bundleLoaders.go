//go:generate go run github.com/vektah/dataloaden BundleLoader ParamBundle *github.com/kyma-incubator/compass/components/director/pkg/graphql.BundlePage

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeyBundle = "dataloadersBundle"

type BundleLoaders struct {
	BundleById BundleLoader
}

type ParamBundle struct {
	ID    string
	First *int
	After *graphql.PageCursor
	Ctx   context.Context
}

func HandlerBundle(fetchFunc func(keys []ParamBundle) ([]*graphql.BundlePage, []error)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeyBundle, &BundleLoaders{
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

func BundleFor(ctx context.Context) *BundleLoaders {
	return ctx.Value(loadersKeyBundle).(*BundleLoaders)
}
