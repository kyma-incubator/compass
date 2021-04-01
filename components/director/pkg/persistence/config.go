package persistence

import (
	"fmt"
	"time"
)

const connStringf string = "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s"

type DatabaseConfig struct {
	User               string        `mapstructure:"user"`
	Password           string        `mapstructure:"password"`
	Host               string        `mapstructure:"host"`
	Port               string        `mapstructure:"port"`
	Name               string        `mapstructure:"name"`
	SSLMode            string        `mapstructure:"ssl"`
	MaxOpenConnections int           `mapstructure:"max_open_connections"`
	MaxIdleConnections int           `mapstructure:"max_idle_connections"`
	ConnMaxLifetime    time.Duration `mapstructure:"connection_max_lifetime"`
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
