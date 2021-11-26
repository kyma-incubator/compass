package adapter

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const (
	AuthStyleAutoDetect AuthStyle = "AuthDetect"
	AuthStyleInParams   AuthStyle = "InParams"
	AuthStyleInHeader   AuthStyle = "InHeader"
	AuthTypeOauth                 = "oauth"
	AuthTypeMTLS                  = "mtls"
)

type AuthStyle string

type Configuration struct {
	Mapping       Mapping
	Auth          Auth
	Port          string        `envconfig:"default=8080"`
	ClientTimeout time.Duration `envconfig:"default=30s"`
	ServerTimeout time.Duration `envconfig:"default=30s"`
	Log           *log.Config
}

type Mapping struct {
	TemplateExternalURL       string
	TemplateHeaders           string
	TemplateJSONBody          string
	TemplateTokenFromResponse string
}

type Auth struct {
	Type         string
	ClientID     string    `envconfig:"optional"`
	ClientSecret string    `envconfig:"optional"`
	URL          string    `envconfig:"optional"`
	AuthStyle    AuthStyle `envconfig:"optional,default=AuthDetect"`
}

// swagger:response externalToken
type ExternalToken struct {
	Token string
}

// Request Data represents information about an Application for which token is going to be created.
//
// swagger:parameters adapter
type RequestData struct {
	// in: body
	Application graphql.Application
	// in: body
	Tenant string
	// in: body
	ClientUser string
}

type ResponseData struct {
	Token string
}
