package statusupdate

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/timestamp"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
)

type Repo interface {
}

type update struct {
	repo         Repo
	table        Table
	timestampGen timestamp.Generator
	transact     persistence.Transactioner
}

type Table string

const (
	applicationsTable Table = "applications"
	runtimesTable     Table = "runtimes"
)

func NewUpdate(repo Repo, transact persistence.Transactioner) *update {
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
				log.Error(err)
			}

			switch consumerInfo.ConsumerType {
			case consumer.Application:
				{
					u.table = applicationsTable
				}
			case consumer.Runtime:
				{
					u.table = runtimesTable
				}
			default:
				{
					next.ServeHTTP(w, r)
				}

			}
			err = u.UpdateStatus(ctx, consumerInfo.ConsumerID)
			if err != nil {
				log.Error(err)
			}

			next.ServeHTTP(w, r)
		})
	}
}
