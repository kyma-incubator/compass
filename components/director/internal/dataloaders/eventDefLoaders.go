//go:generate go run github.com/vektah/dataloaden EventDefLoader ParamEventDef *github.com/kyma-incubator/compass/components/director/pkg/graphql.EventDefinitionPage

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeyEventDef contextKey = "dataloadersEventDef"

type EventDefLoaders struct {
	EventDefByID EventDefLoader
}

type ParamEventDef struct {
	ID    string
	First *int
	After *graphql.PageCursor
	Ctx   context.Context
}

func HandlerEventDef(fetchFunc func(keys []ParamEventDef) ([]*graphql.EventDefinitionPage, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeyEventDef, &EventDefLoaders{
				EventDefByID: EventDefLoader{
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

func EventDefFor(ctx context.Context) *EventDefLoaders {
	return ctx.Value(loadersKeyEventDef).(*EventDefLoaders)
}
