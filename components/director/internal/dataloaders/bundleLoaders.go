//go:generate go run github.com/vektah/dataloaden BundleLoader ParamBundle *github.com/kyma-incubator/compass/components/director/pkg/graphql.BundlePage

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeyBundle contextKey = "dataloadersBundle"

// BundleLoaders missing godoc
type BundleLoaders struct {
	BundleByID BundleLoader
}

// ParamBundle missing godoc
type ParamBundle struct {
	ID    string
	First *int
	After *graphql.PageCursor
	Ctx   context.Context
}

// HandlerBundle missing godoc
func HandlerBundle(fetchFunc func(keys []ParamBundle) ([]*graphql.BundlePage, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeyBundle, &BundleLoaders{
				BundleByID: BundleLoader{
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

// BundleFor missing godoc
func BundleFor(ctx context.Context) *BundleLoaders {
	return ctx.Value(loadersKeyBundle).(*BundleLoaders)
}
