package adapter

import (
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"time"
)

type Configuration struct {
	ServerTimeout time.Duration `envconfig:"default=3s"`
	Port          string        `envconfig:"default=8080"`
	Address       string        `envconfig:"default=127.0.0.1:8080"`
	Log           *log.Config
}
