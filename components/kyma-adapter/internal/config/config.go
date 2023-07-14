package config

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// TenantInfo contains necessary configuration for determining the CMP tenant info
type TenantInfo struct {
	Endpoint           string        `envconfig:"APP_TENANT_INFO_ENDPOINT,default=localhost:8080/v1/info"`
	RequestTimeout     time.Duration `envconfig:"APP_TENANT_INFO_REQUEST_TIMEOUT,default=30s"`
	InsecureSkipVerify bool          `envconfig:"APP_TENANT_INFO_INSECURE_SKIP_VERIFY,default=false"`
}

// Config contains necessary configurations for the kyma-adapter to operate
type Config struct {
	APIRootPath               string        `envconfig:"APP_API_ROOT_PATH,default=/kyma-adapter"`
	APITenantMappingsEndpoint string        `envconfig:"API_TENANT_MAPPINGS_ENDPOINT,default=/v1/tenantMappings/{tenant-id}"`
	Address                   string        `envconfig:"APP_ADDRESS,default=localhost:8080"`
	ServerTimeout             time.Duration `envconfig:"APP_SERVER_TIMEOUT,default=110s"`
	ClientTimeout             time.Duration `envconfig:"APP_CLIENT_TIMEOUT,default=105s"`
	SkipSSLValidation         bool          `envconfig:"APP_HTTP_CLIENT_SKIP_SSL_VALIDATION,default=false"`
	DirectorURL               string        `envconfig:"APP_DIRECTOR_URL"`
	AuthorizationHeaderKey    string        `envconfig:"APP_AUTHORIZATION_HEADER_KEY,default=Authorization"`
	JWKSEndpoint              string        `envconfig:"APP_JWKS_ENDPOINT,default=file://hack/default-jwks.json"`
	AllowJWTSigningNone       bool          `envconfig:"APP_ALLOW_JWT_SIGNING_NONE,default=false"`
	Log                       log.Config
	TenantInfo                TenantInfo
}
