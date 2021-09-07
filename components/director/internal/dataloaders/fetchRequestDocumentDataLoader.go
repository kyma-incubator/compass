//go:generate go run github.com/vektah/dataloaden FetchRequestDocumentLoader ParamFetchRequestDocument *github.com/kyma-incubator/compass/components/director/pkg/graphql.FetchRequest

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeyFetchRequestDocument contextKey = "dataloadersFetchRequestDocument"

// LoadersFetchRequestDocument missing godoc
type LoadersFetchRequestDocument struct {
	FetchRequestDocumentByID FetchRequestDocumentLoader
}

// ParamFetchRequestDocument missing godoc
type ParamFetchRequestDocument struct {
	ID  string
	Ctx context.Context
}

// HandlerFetchRequestDocument missing godoc
func HandlerFetchRequestDocument(fetchFunc func(keys []ParamFetchRequestDocument) ([]*graphql.FetchRequest, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeyFetchRequestDocument, &LoadersFetchRequestDocument{
				FetchRequestDocumentByID: FetchRequestDocumentLoader{
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

// ForFetchRequestDocument missing godoc
func ForFetchRequestDocument(ctx context.Context) *LoadersFetchRequestDocument {
	return ctx.Value(loadersKeyFetchRequestDocument).(*LoadersFetchRequestDocument)
}
