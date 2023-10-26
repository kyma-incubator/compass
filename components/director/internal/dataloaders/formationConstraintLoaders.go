//go:generate go run github.com/vektah/dataloaden FormationConstraintLoader ParamFormationConstraint []*github.com/kyma-incubator/compass/components/director/pkg/graphql.FormationConstraint

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeyFormationConstraint contextKey = "dataloadersFormationConstraint"

// FormationConstraintLoaders missing godoc
type FormationConstraintLoaders struct {
	FormationConstraintByID FormationConstraintLoader
}

// ParamFormationConstraint missing godoc
type ParamFormationConstraint struct {
	ID  string
	Ctx context.Context
}

// HandlerFormationConstraint missing godoc
func HandlerFormationConstraint(fetchFunc func(keys []ParamFormationConstraint) ([][]*graphql.FormationConstraint, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeyFormationConstraint, &FormationConstraintLoaders{
				FormationConstraintByID: FormationConstraintLoader{
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

// FormationTemplateFor missing godoc
func FormationTemplateFor(ctx context.Context) *FormationConstraintLoaders {
	return ctx.Value(loadersKeyFormationConstraint).(*FormationConstraintLoaders)
}
