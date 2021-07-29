package healthz

import (
	"context"
	"net/http"
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
	pinger           Pinger
	schemaCompatible bool
	cfg              ReadyConfig
	repo             Repository
}

func NewReady(ctx context.Context, pinger Pinger, cfg ReadyConfig, repository Repository) *Ready {
	return &Ready{
		ctx:              ctx,
		pinger:           pinger,
		schemaCompatible: false,
		cfg:              cfg,
		repo:             repository,
	}
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

		if err := r.pinger.PingContext(r.ctx); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(http.StatusOK)
	}
}
