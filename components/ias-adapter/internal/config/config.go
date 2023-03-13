package config

import (
	"time"

	"github.com/vrischmann/envconfig"
)

const envPrefix = "APP"

type Config struct {
	Address           string        `envconfig:"APP_ADDRESS,default=localhost:8080"`
	ReadTimeout       time.Duration `envconfig:"APP_READ_TIMEOUT,default=30s"`
	ReadHeaderTimeout time.Duration `envconfig:"APP_READ_HEADER_TIMEOUT,default=30s"`
	WriteTimeout      time.Duration `envconfig:"APP_WRITE_TIMEOUT,default=30s"`
	IdleTimeout       time.Duration `envconfig:"APP_IDLE_TIMEOUT,default=30s"`
	IASConfig         IAS
	Postgres          Postgres
}

type IAS struct {
	RequestTimeout        time.Duration `envconfig:"APP_IAS_REQUEST_TIMEOUT,default=30s"`
	CockpitClientCertPath string        `envconfig:"APP_IAS_COCKPIT_CLIENT_CERT_PATH,default=cockpit-client.cert"`
	CockpitClientKeyPath  string        `envconfig:"APP_IAS_COCKPIT_CLIENT_KEY_PATH,default=cockpit-client.key"`
	CockpitCAPath         string        `envconfig:"APP_IAS_COCKPIT_CA_PATH,default=cockpit-ca.cert"`
}

type Postgres struct {
	URI            string        `envconfig:"APP_POSTGRES_URI,default=30s"`
	ConnectTimeout time.Duration `envconfig:"APP_POSTGRES_CONNECT_TIMEOUT,default=30s"`
	RequestTimeout time.Duration `envconfig:"APP_POSTGRES_REQUEST_TIMEOUT,default=30s"`
}

func New() (Config, error) {
	cfg := Config{}
	return cfg, envconfig.InitWithPrefix(&cfg, envPrefix)
}
