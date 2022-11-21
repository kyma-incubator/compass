//go:generate go run github.com/vektah/dataloaden SpecEventDefLoader ParamSpecEventDef *github.com/kyma-incubator/compass/components/director/pkg/graphql.EventSpec

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeySpecEventDef contextKey = "dataloadersSpecEventDef"

// LoadersSpecEventDef missing godoc
type LoadersSpecEventDef struct {
	SpecEventDefByID SpecEventDefLoader
}

// ParamSpecEventDef missing godoc
type ParamSpecEventDef struct {
	ID  string
	Ctx context.Context
}

// HandlerSpecEventDef missing godoc
func HandlerSpecEventDef(fetchFunc func(keys []ParamSpecEventDef) ([]*graphql.EventSpec, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeySpecEventDef, &LoadersSpecEventDef{
				SpecEventDefByID: SpecEventDefLoader{
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

// ForSpecEventDef missing godoc
func ForSpecEventDef(ctx context.Context) *LoadersSpecEventDef {
	return ctx.Value(loadersKeySpecEventDef).(*LoadersSpecEventDef)
}
