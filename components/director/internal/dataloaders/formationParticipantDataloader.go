//go:generate go run github.com/vektah/dataloaden FormationParticipantDataloader ParamFormationParticipant github.com/kyma-incubator/compass/components/director/pkg/graphql.FormationParticipant

package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const loadersKeySourceFormationParticipant contextKey = "dataloadersSourceFormationParticipant"
const loadersKeyTargetFormationParticipant contextKey = "dataloadersTargetFormationParticipant"

// LoadersFormationParticipant is a dataloader for formation participants
type LoadersFormationParticipant struct {
	FormationParticipantDataloader FormationParticipantDataloader
}

// ParamFormationParticipant are parameters for the formation participant dataloader
type ParamFormationParticipant struct {
	ID              string
	ParticipantID   string
	ParticipantType string
	Ctx             context.Context
}

// HandlerFormationParticipant prepares the parameters for the dataloader
func HandlerFormationParticipant(fetchFunc func(keys []ParamFormationParticipant) ([]graphql.FormationParticipant, []error), maxBatch int, wait time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKeySourceFormationParticipant, &LoadersFormationParticipant{
				FormationParticipantDataloader: FormationParticipantDataloader{
					maxBatch: maxBatch,
					wait:     wait,
					fetch:    fetchFunc,
				},
			})
			ctx = context.WithValue(ctx, loadersKeyTargetFormationParticipant, &LoadersFormationParticipant{
				FormationParticipantDataloader: FormationParticipantDataloader{
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

// ForSourceFormationParticipant retrieves the dataloader for formation assignment sources from the context
func ForSourceFormationParticipant(ctx context.Context) *LoadersFormationParticipant {
	return ctx.Value(loadersKeySourceFormationParticipant).(*LoadersFormationParticipant)
}

// ForTargetFormationParticipant retrieves the dataloader for formation assignment targets from the context
func ForTargetFormationParticipant(ctx context.Context) *LoadersFormationParticipant {
	return ctx.Value(loadersKeyTargetFormationParticipant).(*LoadersFormationParticipant)
}
