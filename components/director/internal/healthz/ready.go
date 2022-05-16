package healthz

import (
	"context"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// Repository missing godoc
//go:generate mockery --name=Repository --output=automock --outpkg=automock --case=underscore --disable-version-string
type Repository interface {
	GetVersion(ctx context.Context) (string, bool, error)
}

// ReadyConfig missing godoc
type ReadyConfig struct {
	SchemaMigrationVersion string `envconfig:"APP_SCHEMA_MIGRATION_VERSION"`
}

// Ready missing godoc
type Ready struct {
	transactioner    persistence.Transactioner
	schemaCompatible bool
	cfg              ReadyConfig
	repo             Repository
}

// NewReady missing godoc
func NewReady(transactioner persistence.Transactioner, cfg ReadyConfig, repository Repository) *Ready {
	return &Ready{
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
		if r.schemaCompatible = r.checkSchemaCompatibility(request.Context()); !r.schemaCompatible {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := r.transactioner.PingContext(request.Context()); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(http.StatusOK)
	}
}

func (r *Ready) checkSchemaCompatibility(ctx context.Context) bool {
	if r.schemaCompatible {
		return true
	}

	tx, err := r.transactioner.Begin()
	if err != nil {
		log.C(ctx).Errorf(errors.Wrap(err, "while starting transaction").Error())
		return false
	}
	defer r.transactioner.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	schemaVersion, dirty, err := r.repo.GetVersion(ctx)
	if err != nil {
		log.C(ctx).Error(err.Error())
		return false
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).Error(errors.Wrap(err, "while committing transaction").Error())
		return false
	}

	if r.cfg.SchemaMigrationVersion != schemaVersion || dirty {
		log.C(ctx).Errorf("Incompatible schema version. Expected: %s, Current (version, dirty): (%s, %v)", r.cfg.SchemaMigrationVersion, schemaVersion, dirty)
		return false
	}

	return true
}
