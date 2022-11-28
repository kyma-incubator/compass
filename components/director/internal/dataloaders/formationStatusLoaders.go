//go:generate go run github.com/vektah/dataloaden FormationStatusLoader ParamFormationStatus *github.com/kyma-incubator/compass/components/director/pkg/graphql.FormationStatus

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeyFormationStatus contextKey = "dataloadersFormationStatus"

// FormationStatusLoaders missing godoc
type FormationStatusLoaders struct {
	FormationStatusByID FormationStatusLoader
}

// ParamFormationStatus missing godoc
type ParamFormationStatus struct {
	ID  string
	Ctx context.Context
}

// HandlerFormationStatus missing godoc
func HandlerFormationStatus(fetchFunc func(keys []ParamFormationStatus) ([]*graphql.FormationStatus, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeyFormationStatus, &FormationStatusLoaders{
				FormationStatusByID: FormationStatusLoader{
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

// FormationStatusFor missing godoc
func FormationStatusFor(ctx context.Context) *FormationStatusLoaders {
	return ctx.Value(loadersKeyFormationStatus).(*FormationStatusLoaders)
}
