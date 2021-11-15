package adapter

import (
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"time"
)

type Configuration struct {
	ServerTimeout time.Duration `envconfig:"default=30s"` //TODO What is the proper timeout value?
	ClientTimeout time.Duration `envconfig:"default=30s"`
	Port          string        `envconfig:"default=8080"`
	Address       string        `envconfig:"default=127.0.0.1:8080"`
	Log           *log.Config

	DefaultScenarioEnabled bool `envconfig:"true"`
	Database               persistence.DatabaseConfig

	SystemToTemplateMappings string `envconfig:"APP_SYSTEM_TO_TEMPLATE_MAPPINGS"`
}
