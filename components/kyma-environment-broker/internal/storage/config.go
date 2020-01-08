package storage

import "fmt"

const ConnectionURLFormat = "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s"

type Config struct {
	User     string `envconfig:"default=postgres"`
	Password string `envconfig:"default=password"`
	Host     string `envconfig:"default=localhost"`
	Port     string `envconfig:"default=5432"`
	Name     string `envconfig:"default=broker"`
	SSLMode  string `envconfig:"default=disable"`
}

func (cfg *Config) ConnectionURL() string {
	return fmt.Sprintf(ConnectionURLFormat, cfg.Host, cfg.Port, cfg.User,
		cfg.Password, cfg.Name, cfg.SSLMode)
}
