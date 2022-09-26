package adapter

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/healthz"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

// Configuration contains ns-adapter specific configuration values
type Configuration struct {
	ReadHeadersTimeout time.Duration `envconfig:"APP_READ_REQUEST_HEADERS_TIMEOUT,default=30s"`
	ServerTimeout      time.Duration `envconfig:"APP_SERVER_TIMEOUT,default=30s"` // TODO What is the proper timeout value?
	ClientTimeout      time.Duration `envconfig:"APP_CLIENT_TIMEOUT,default=30s"`
	Address            string        `envconfig:"default=127.0.0.1:8080"`
	Log                *log.Config

	CertLoaderConfig certloader.Config

	Database persistence.DatabaseConfig

	SystemToTemplateMappings string `envconfig:"APP_SYSTEM_TO_TEMPLATE_MAPPINGS,default='{}'"`
	AllowJWTSigningNone      bool   `envconfig:"APP_ALLOW_JWT_SIGNING_NONE,default=false"`
	JwksEndpoint             string `envconfig:"APP_JWKS_ENDPOINT"`

	ReadyConfig healthz.ReadyConfig

	SelfRegisterDistinguishLabelKey string `envconfig:"APP_SELF_REGISTER_DISTINGUISH_LABEL_KEY"`
	RuntimeTypeLabelKey             string `envconfig:"APP_RUNTIME_TYPE_LABEL_KEY,default=runtimeType"`
	ApplicationTypeLabelKey         string `envconfig:"APP_APPLICATION_TYPE_LABEL_KEY,default=applicationType"`

	ORDWebhookMappings string `envconfig:"APP_ORD_WEBHOOK_MAPPINGS"`

	ExternalClientCertSecretName string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET_NAME"`
	ExtSvcClientCertSecretName   string `envconfig:"APP_EXT_SVC_CLIENT_CERT_SECRET_NAME"`
}
