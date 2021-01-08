package statusupdate

import (
	"context"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
)

type update struct {
	transact persistence.Transactioner
	repo     StatusUpdateRepository
}

//go:generate mockery -name=StatusUpdateRepository -output=automock -outpkg=automock -case=underscore
type StatusUpdateRepository interface {
	UpdateStatus(ctx context.Context, id string, object WithStatusObject) error
	IsConnected(ctx context.Context, id string, object WithStatusObject) (bool, error)
}

type WithStatusObject string

const (
	Applications WithStatusObject = "applications"
	Runtimes     WithStatusObject = "runtimes"
)

func New(transact persistence.Transactioner, repo StatusUpdateRepository) *update {
	return &update{
		transact: transact,
		repo:     repo,
	}
}

func (u *update) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			consumerInfo, err := consumer.LoadFromContext(r.Context())
			logger := log.C(r.Context())

			if err != nil {
				logger.WithError(err).Error("An error has occurred while fetching consumer info from context.")
				next.ServeHTTP(w, r)
				return
			}
			var object WithStatusObject
			switch consumerInfo.ConsumerType {
			case consumer.Application:
				object = Applications
			case consumer.Runtime:
				object = Runtimes
			default:
				next.ServeHTTP(w, r)
				return
			}

			tx, err := u.transact.Begin()
			if err != nil {
				logger.WithError(err).Error("An error has occurred while opening transaction.")
				next.ServeHTTP(w, r)
				return
			}
			defer u.transact.RollbackUnlessCommitted(r.Context(), tx)

			ctxWithDB := persistence.SaveToContext(r.Context(), tx)

			isConnected, err := u.repo.IsConnected(ctxWithDB, consumerInfo.ConsumerID, object)
			if err != nil {
				logger.WithError(err).Error("An error has occurred while checking repository status.")
				next.ServeHTTP(w, r)
				return
			}

			if !isConnected {
				err = u.repo.UpdateStatus(ctxWithDB, consumerInfo.ConsumerID, object)
				if err != nil {
					logger.WithError(err).Error("An error has occurred while updating repository status.")
					next.ServeHTTP(w, r)
					return
				}
			}

			if err := tx.Commit(); err != nil {
				logger.WithError(err).Error("An error has occurred while committing transaction.")
				next.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
