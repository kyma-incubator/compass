package healthz

import (
	"context"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

//go:generate mockery --name=Repository --output=automock --outpkg=automock --case=underscore
type Repository interface {
	GetVersion(ctx context.Context) (string, error)
}

type ReadyConfig struct {
	SchemaMigrationVersion string `envconfig:"APP_SCHEMA_MIGRATION_VERSION"`
}

type Ready struct {
	ctx              context.Context
	transactioner    persistence.Transactioner
	schemaCompatible bool
	cfg              ReadyConfig
	repo             Repository
}

func NewReady(ctx context.Context, transactioner persistence.Transactioner, cfg ReadyConfig, repository Repository) *Ready {
	return &Ready{
		ctx:              ctx,
		transactioner:    transactioner,
		schemaCompatible: false,
		cfg:              cfg,
		repo:             repository,
	}
}

// NewReadinessHandler returns handler that returns OK if the db schema is compatible and successful
// db ping is performed or InternalServerError otherwise
func NewReadinessHandler(r *Ready) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		if r.schemaCompatible = r.checkSchemaCompatibility(); !r.schemaCompatible {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := r.transactioner.PingContext(r.ctx); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(http.StatusOK)
	}
}

func (r *Ready) checkSchemaCompatibility() bool {
	if r.schemaCompatible {
		return true
	}

	tx, err := r.transactioner.Begin()
	if err != nil {
		log.C(r.ctx).Errorf(errors.Wrap(err, "while starting transaction").Error())
		return false
	}
	defer r.transactioner.RollbackUnlessCommitted(r.ctx, tx)

	r.ctx = persistence.SaveToContext(r.ctx, tx)

	schemaVersion, err := r.repo.GetVersion(r.ctx)
	if err != nil {
		log.C(r.ctx).Error(err.Error())
		return false
	}

	if err := tx.Commit(); err != nil {
		log.C(r.ctx).Error(errors.Wrap(err, "while committing transaction").Error())
		return false
	}

	return r.cfg.SchemaMigrationVersion == schemaVersion
}
