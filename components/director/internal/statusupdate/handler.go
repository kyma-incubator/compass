package statusupdate

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/timestamp"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
)

type update struct {
	table        Table
	timestampGen timestamp.Generator
	transact     persistence.Transactioner
}

type Table string

const (
	applicationsTable Table = "applications"
	runtimesTable     Table = "runtimes"
)

func NewUpdate(transact persistence.Transactioner) *update {
	return &update{
		table:        "",
		timestampGen: timestamp.DefaultGenerator(),
		transact:     transact,
	}
}

func (u *update) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			consumerInfo, err := consumer.LoadFromContext(ctx)
			if err != nil {
				u.writeError(w, errors.Wrap(err, "while fetching consumer info from from context").Error(), http.StatusBadRequest)
				log.Error(err)
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

			isConnected, err := u.IsConnected(consumerInfo.ConsumerID)
			if err != nil {
				u.writeError(w, errors.Wrap(err, "while checking status").Error(), http.StatusInternalServerError)
				log.Error(err)
				return
			}

			if isConnected {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			err = u.UpdateStatus(consumerInfo.ConsumerID)
			if err != nil {

				u.writeError(w, errors.Wrap(err, "while updating status").Error(), http.StatusInternalServerError)
				log.Error(err)
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
