package config

import (
	"fmt"
	"time"

	"github.com/vrischmann/envconfig"
)

const envPrefix = "APP"

type Config struct {
	APIRootPath       string        `envconfig:"APP_API_ROOT_PATH,default=/ias-adapter"`
	Address           string        `envconfig:"APP_ADDRESS,default=localhost:8080"`
	ReadTimeout       time.Duration `envconfig:"APP_READ_TIMEOUT,default=30s"`
	ReadHeaderTimeout time.Duration `envconfig:"APP_READ_HEADER_TIMEOUT,default=30s"`
	WriteTimeout      time.Duration `envconfig:"APP_WRITE_TIMEOUT,default=30s"`
	IdleTimeout       time.Duration `envconfig:"APP_IDLE_TIMEOUT,default=30s"`
	TenantInfo        TenantInfo
	IASConfig         IAS
	Postgres          Postgres
}

type TenantInfo struct {
	Endpoint       string        `envconfig:"APP_TENANT_INFO_ENDPOINT,default=localhost:8080/v1/info"`
	RequestTimeout time.Duration `envconfig:"APP_TENANT_INFO_REQUEST_TIMEOUT,default=30s"`
}

type IAS struct {
	CockpitSecretPath string        `envconfig:"APP_IAS_COCKPIT_PATH"`
	RequestTimeout    time.Duration `envconfig:"APP_IAS_REQUEST_TIMEOUT,default=30s"`
}

type Postgres struct {
	User           string        `envconfig:"APP_POSTGRES_USER,default=user"`
	Password       string        `envconfig:"APP_POSTGRES_PASSWORD,default=password"`
	Host           string        `envconfig:"APP_POSTGRES_HOST,default=localhost"`
	Port           uint16        `envconfig:"APP_POSTGRES_PORT,default=5432"`
	DatabaseName   string        `envconfig:"APP_POSTGRES_DB_NAME,default=db_name"`
	SSLMode        string        `envconfig:"APP_POSTGRES_SSL_MODE,default=disable"`
	ConnectTimeout time.Duration `envconfig:"APP_POSTGRES_CONNECT_TIMEOUT,default=30s"`
	RequestTimeout time.Duration `envconfig:"APP_POSTGRES_REQUEST_TIMEOUT,default=30s"`
}

func (p Postgres) ConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.DatabaseName, p.SSLMode)
}

func New() (Config, error) {
	cfg := Config{}
	return cfg, envconfig.InitWithPrefix(&cfg, envPrefix)
}
