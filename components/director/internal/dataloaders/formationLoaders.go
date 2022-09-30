//go:generate go run github.com/vektah/dataloaden FormationAssignmentLoader ParamFormationAssignment *github.com/kyma-incubator/compass/components/director/pkg/graphql.FormationAssignmentPage

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeyFormation contextKey = "dataloadersFormation"

// FormationAssignmentLoaders missing godoc
type FormationAssignmentLoaders struct {
	FormationAssignmentByID FormationAssignmentLoader
}

// ParamFormationAssignment missing godoc
type ParamFormationAssignment struct {
	ID    string
	First *int
	After *graphql.PageCursor
	Ctx   context.Context
}

// HandlerFormation missing godoc
func HandlerFormation(fetchFunc func(keys []ParamFormationAssignment) ([]*graphql.FormationAssignmentPage, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeyFormation, &FormationAssignmentLoaders{
				FormationAssignmentByID: FormationAssignmentLoader{
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
func FormationFor(ctx context.Context) *FormationAssignmentLoaders {
	return ctx.Value(loadersKeyFormation).(*FormationAssignmentLoaders)
}
