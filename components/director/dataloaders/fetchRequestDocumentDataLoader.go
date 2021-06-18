//go:generate go run github.com/vektah/dataloaden FetchRequestDocumentLoader ParamFetchRequestDocument *github.com/kyma-incubator/compass/components/director/pkg/graphql.FetchRequest

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeyFetchRequestDocument = "dataloadersFetchRequestDocument"

type LoadersFetchRequestDocument struct {
	FetchRequestDocumentById FetchRequestDocumentLoader
}

type ParamFetchRequestDocument struct {
	ID  string
	Ctx context.Context
}

func HandlerFetchRequestDocument(fetchFunc func(keys []ParamFetchRequestDocument) ([]*graphql.FetchRequest, []error)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeyFetchRequestDocument, &LoadersFetchRequestDocument{
				FetchRequestDocumentById: FetchRequestDocumentLoader{
					maxBatch: 500,
					wait:     3 * time.Millisecond,
					fetch:    fetchFunc,
				},
			})
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func ForFetchRequestDocument(ctx context.Context) *LoadersFetchRequestDocument {
	return ctx.Value(loadersKeyFetchRequestDocument).(*LoadersFetchRequestDocument)
}