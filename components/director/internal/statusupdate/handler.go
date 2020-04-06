package statusupdate

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
)

type update struct {
	transact persistence.Transactioner
	repo     StatusUpdateRepository
	logger   *log.Logger
}

type Table string

//go:generate mockery -name=StatusUpdateRepository -output=automock -outpkg=automock -case=underscore
type StatusUpdateRepository interface {
	UpdateStatus(ctx context.Context, id string, table Table) error
	IsConnected(ctx context.Context, id string, table Table) (bool, error)
}

const (
	ApplicationsTable Table = "applications"
	RuntimesTable     Table = "runtimes"
)

func New(transact persistence.Transactioner, repo StatusUpdateRepository, logger *log.Logger) *update {
	return &update{
		transact: transact,
		repo:     repo,
		logger:   logger,
	}
}

func (u *update) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			consumerInfo, err := consumer.LoadFromContext(ctx)
			if err != nil {
				u.logger.Error(errors.Wrap(err, "while fetching consumer info from from context").Error())
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			var table Table
			switch consumerInfo.ConsumerType {
			case consumer.Application:
				table = ApplicationsTable
			case consumer.Runtime:
				table = RuntimesTable
			default:
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			tx, err := u.transact.Begin()
			if err != nil {
				u.logger.Error(errors.Wrap(err, "while opening transaction").Error())
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			defer u.transact.RollbackUnlessCommited(tx)

			ctxWithDB := persistence.SaveToContext(ctx, tx)

			isConnected, err := u.repo.IsConnected(ctxWithDB, consumerInfo.ConsumerID, table)
			if err != nil {
				u.logger.Error(errors.Wrap(err, "while checking status").Error())
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if !isConnected {
				err = u.repo.UpdateStatus(ctxWithDB, consumerInfo.ConsumerID, table)
				if err != nil {
					u.logger.Error(errors.Wrap(err, "while updating status").Error())
					next.ServeHTTP(w, r.WithContext(ctx))

					return
				}
			}

			if err := tx.Commit(); err != nil {
				u.logger.Error(errors.Wrap(err, "while committing").Error())
				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type errorResponse struct {
	Errors []gqlError `json:"errors"`
}

type gqlError struct {
	Message string `json:"message"`
}

func (u *update) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")

	resp := errorResponse{Errors: []gqlError{{Message: message}}}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error(err, "while encoding data")
	}
}
