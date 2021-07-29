package healthz

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/schema"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"net/http"
)

type ReadyConfig struct {
	SchemaMigrationVersion string `envconfig:"APP_SCHEMA_MIGRATION_VERSION"`
}

type Ready struct {
	ctx              context.Context
	transact         persistence.Transactioner
	schemaCompatible bool
	cfg              ReadyConfig
	repo             *schema.PgRepository
}

func NewReady(ctx context.Context, transactioner persistence.Transactioner, cfg ReadyConfig) (*Ready, error) {
	return &Ready{
		ctx:              ctx,
		transact:         transactioner,
		schemaCompatible: false,
		cfg:              cfg,
		repo:             schema.NewRepository(),
	}, nil
}

func (r *Ready) checkSchemaCompatibility() bool {
	if r.schemaCompatible {
		return true
	}

	schemaVersion, err := r.repo.GetVersion(r.ctx)
	if err != nil {
		return false
	}

	return r.cfg.SchemaMigrationVersion == schemaVersion
}

// NewReadinessHandler returns handler that returns OK if the db schema is compatible and successful
// db ping is performed or InternalServerError otherwise
func NewReadinessHandler(r *Ready) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		isCompatible := r.checkSchemaCompatibility()
		if isCompatible {
			r.schemaCompatible = true
		}

		if !r.schemaCompatible {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := r.transact.PingContext(r.ctx); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(http.StatusOK)
	}
}
