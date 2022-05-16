package adapter

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/healthz"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

// Configuration contains ns-adapter specific configuration values
type Configuration struct {
	ServerTimeout time.Duration `envconfig:"default=30s"` // TODO What is the proper timeout value?
	ClientTimeout time.Duration `envconfig:"default=30s"`
	Address       string        `envconfig:"default=127.0.0.1:8080"`
	Log           *log.Config

	CertLoaderConfig certloader.Config

	DefaultScenarioEnabled bool `envconfig:"default=true"`
	Database               persistence.DatabaseConfig

	SystemToTemplateMappings string `envconfig:"APP_SYSTEM_TO_TEMPLATE_MAPPINGS,default='{}'"`
	AllowJWTSigningNone      bool   `envconfig:"APP_ALLOW_JWT_SIGNING_NONE,default=false"`
	JwksEndpoint             string `envconfig:"APP_JWKS_ENDPOINT"`

	ReadyConfig healthz.ReadyConfig
}
