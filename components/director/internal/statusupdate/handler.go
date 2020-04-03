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
	table    Table
	transact persistence.Transactioner
	repo     StatusUpdateRepository
}

type Table string

//go:generate mockery -name=StatusUpdateRepository -output=automock -outpkg=automock -case=underscore
type StatusUpdateRepository interface {
	UpdateStatus(ctx context.Context, id, table string) error
	IsConnected(ctx context.Context, id, table string) (bool, error)
}

const (
	applicationsTable Table = "applications"
	runtimesTable     Table = "runtimes"
)

func New(transact persistence.Transactioner, repo StatusUpdateRepository) *update {
	return &update{
		table:    "",
		transact: transact,
		repo:     repo,
	}
}

func (u *update) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			consumerInfo, err := consumer.LoadFromContext(ctx)
			if err != nil {
				u.writeError(w, errors.Wrap(err, "while fetching consumer info from from context").Error(), http.StatusBadRequest)
				return
			}

			switch consumerInfo.ConsumerType {
			case consumer.Application:
				u.table = applicationsTable
			case consumer.Runtime:
				u.table = runtimesTable
			default:
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			tx, err := u.transact.Begin()
			if err != nil {
				u.writeError(w, errors.Wrap(err, "while opening transaction").Error(), http.StatusInternalServerError)
				return
			}
			defer u.transact.RollbackUnlessCommited(tx)

			ctxWithDB := persistence.SaveToContext(ctx, tx)

			isConnected, err := u.repo.IsConnected(ctxWithDB, consumerInfo.ConsumerID, string(u.table))
			if err != nil {
				u.writeError(w, errors.Wrap(err, "while checking status").Error(), http.StatusInternalServerError)
				return
			}

			if isConnected {
				if err := tx.Commit(); err != nil {
					u.writeError(w, errors.Wrap(err, "while committing").Error(), http.StatusInternalServerError)
					return
				}
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			err = u.repo.UpdateStatus(ctxWithDB, consumerInfo.ConsumerID, string(u.table))
			if err != nil {

				u.writeError(w, errors.Wrap(err, "while updating status").Error(), http.StatusInternalServerError)
				return
			}

			if err := tx.Commit(); err != nil {
				u.writeError(w, errors.Wrap(err, "while committing").Error(), http.StatusInternalServerError)
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
