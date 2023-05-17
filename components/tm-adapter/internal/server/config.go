package server

import "time"

type Config struct {
	Port                     int           `envconfig:"APP_PORT"`
	RootAPIPath              string        `envconfig:"APP_ROOT_API_PATH"`
	TenantMappingAPIEndpoint string        `envconfig:"APP_TENANT_MAPPING_API_ENDPOINT"`
	ReadTimeout              time.Duration `envconfig:"APP_READ_TIMEOUT,default=30s"`
	ReadHeaderTimeout        time.Duration `envconfig:"APP_READ_HEADER_TIMEOUT,default=30s"`
	WriteTimeout             time.Duration `envconfig:"APP_WRITE_TIMEOUT,default=50s"`
	IdleTimeout              time.Duration `envconfig:"APP_IDLE_TIMEOUT,default=30s"`
	Timeout                  time.Duration `envconfig:"APP_SERVER_TIMEOUT,default=60s"`
	ShutdownTimeout          time.Duration `envconfig:"default=15s"`
}
