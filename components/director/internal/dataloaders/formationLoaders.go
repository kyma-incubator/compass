//go:generate go run github.com/vektah/dataloaden FormationLoader ParamFormation *github.com/kyma-incubator/compass/components/director/pkg/graphql.FormationAssignmentPage

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeyFormation contextKey = "dataloadersFormation"

// FormationLoaders missing godoc
type FormationLoaders struct {
	FormationByID FormationLoader
}

// ParamFormation missing godoc
type ParamFormation struct {
	ID    string
	First *int
	After *graphql.PageCursor
	Ctx   context.Context
}

// HandlerFormation missing godoc
func HandlerFormation(fetchFunc func(keys []ParamFormation) ([]*graphql.FormationAssignmentPage, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeyFormation, &FormationLoaders{
				FormationByID: FormationLoader{
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

// FormationFor missing godoc
func FormationFor(ctx context.Context) *FormationLoaders {
	return ctx.Value(loadersKeyFormation).(*FormationLoaders)
}
