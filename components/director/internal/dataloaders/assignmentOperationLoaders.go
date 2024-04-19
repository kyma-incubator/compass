//go:generate go run github.com/vektah/dataloaden AssignmentOperationLoader ParamAssignmentOperation *github.com/kyma-incubator/compass/components/director/pkg/graphql.AssignmentOperationPage

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeyForAssignmentOperation contextKey = "dataloadersAssignmentOperation"

type AssignmentOperationLoaders struct {
	AssignmentOperationByID AssignmentOperationLoader
}

type ParamAssignmentOperation struct {
	ID     string
	First  *int
	After  *graphql.PageCursor
	Ctx    context.Context
}

func HandlerAssignmentOperation(fetchFunc func(keys []ParamAssignmentOperation) ([]*graphql.AssignmentOperationPage, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeyFormationAssignment, &AssignmentOperationLoaders{
				AssignmentOperationByID: AssignmentOperationLoader{
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

func FormationAssignmentFor(ctx context.Context) *AssignmentOperationLoaders {
	return ctx.Value(loadersKeyForAssignmentOperation).(*AssignmentOperationLoaders)
}

