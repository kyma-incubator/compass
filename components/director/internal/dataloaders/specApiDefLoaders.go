//go:generate go run github.com/vektah/dataloaden SpecAPIDefLoader ParamSpecAPIDef *github.com/kyma-incubator/compass/components/director/pkg/graphql.APISpec

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeySpecAPIDef contextKey = "dataloadersSpecAPIDef"

// LoadersSpecAPIDef missing godoc
type LoadersSpecAPIDef struct {
	SpecAPIDefByID SpecAPIDefLoader
}

// ParamSpecAPIDef missing godoc
type ParamSpecAPIDef struct {
	ID  string
	Ctx context.Context
}

// HandlerSpecAPIDef missing godoc
func HandlerSpecAPIDef(fetchFunc func(keys []ParamSpecAPIDef) ([]*graphql.APISpec, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeySpecAPIDef, &LoadersSpecAPIDef{
				SpecAPIDefByID: SpecAPIDefLoader{
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

// ForSpecAPIDef missing godoc
func ForSpecAPIDef(ctx context.Context) *LoadersSpecAPIDef {
	return ctx.Value(loadersKeySpecAPIDef).(*LoadersSpecAPIDef)
}
