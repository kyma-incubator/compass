package persistence

import "fmt"

const connStringf string = "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s"

type DatabaseConfig struct {
	User     string `envconfig:"default=postgres,APP_DB_USER"`
	Password string `envconfig:"default=pgsql@12345,APP_DB_PASSWORD"`
	Host     string `envconfig:"default=localhost,APP_DB_HOST"`
	Port     string `envconfig:"default=5432,APP_DB_PORT"`
	Name     string `envconfig:"default=postgres,APP_DB_NAME"`
	SSLMode  string `envconfig:"default=disable,APP_DB_SSL"`
}

func GetConnString(cfg DatabaseConfig) string {
	return fmt.Sprintf(connStringf, cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)
}
