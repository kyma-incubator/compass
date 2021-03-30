package persistence

import (
	"fmt"
	"time"
)

const connStringf string = "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s"

type DatabaseConfig struct {
	User               string        `mapstructure:"DB_USER"`
	Password           string        `mapstructure:"DB_PASSWORD"`
	Host               string        `mapstructure:"DB_HOST"`
	Port               string        `mapstructure:"DB_PORT"`
	Name               string        `mapstructure:"DB_NAME"`
	SSLMode            string        `mapstructure:"DB_SSL"`
	MaxOpenConnections int           `mapstructure:"DB_MAX_OPEN_CONNECTIONS"`
	MaxIdleConnections int           `mapstructure:"DB_MAX_IDLE_CONNECTIONS"`
	ConnMaxLifetime    time.Duration `mapstructure:"DB_CONNECTION_MAX_LIFETIME"`
}

func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		User:               "postgres",
		Password:           "pgsql@12345",
		Host:               "localhost",
		Port:               "5432",
		Name:               "compass",
		SSLMode:            "disable",
		MaxOpenConnections: 2,
		MaxIdleConnections: 2,
		ConnMaxLifetime:    30 * time.Minute,
	}
}

func (cfg DatabaseConfig) GetConnString() string {
	return fmt.Sprintf(connStringf, cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)
}
