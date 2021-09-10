//go:generate go run github.com/vektah/dataloaden DocumentLoader ParamDocument *github.com/kyma-incubator/compass/components/director/pkg/graphql.DocumentPage

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeyDocument contextKey = "dataloadersDocument"

// DocumentLoaders missing godoc
type DocumentLoaders struct {
	DocumentByID DocumentLoader
}

// ParamDocument missing godoc
type ParamDocument struct {
	ID    string
	First *int
	After *graphql.PageCursor
	Ctx   context.Context
}

// HandlerDocument missing godoc
func HandlerDocument(fetchFunc func(keys []ParamDocument) ([]*graphql.DocumentPage, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeyDocument, &DocumentLoaders{
				DocumentByID: DocumentLoader{
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

// DocumentFor missing godoc
func DocumentFor(ctx context.Context) *DocumentLoaders {
	return ctx.Value(loadersKeyDocument).(*DocumentLoaders)
}
