package persistence

import (
	"fmt"
	"time"
)

const connStringf string = "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s"

// DatabaseConfig missing godoc
type DatabaseConfig struct {
	User               string        `envconfig:"default=postgres,APP_DB_USER"`
	Password           string        `envconfig:"default=pgsql@12345,APP_DB_PASSWORD"`
	Host               string        `envconfig:"default=localhost,APP_DB_HOST"`
	Port               string        `envconfig:"default=5432,APP_DB_PORT"`
	Name               string        `envconfig:"default=compass,APP_DB_NAME"`
	SSLMode            string        `envconfig:"default=disable,APP_DB_SSL"`
	MaxOpenConnections int           `envconfig:"default=5,APP_DB_MAX_OPEN_CONNECTIONS"`
	MaxIdleConnections int           `envconfig:"default=5,APP_DB_MAX_IDLE_CONNECTIONS"`
	ConnMaxLifetime    time.Duration `envconfig:"default=30m,APP_DB_CONNECTION_MAX_LIFETIME"`
}

// GetConnString missing godoc
func (cfg DatabaseConfig) GetConnString() string {
	return fmt.Sprintf(connStringf, cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)
}
