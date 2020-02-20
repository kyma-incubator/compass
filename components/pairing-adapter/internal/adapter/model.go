package adapter

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type Configuration struct {
	Mapping Mapping
	OAuth   Oauth
	Port    string `envconfig:"default=8080"`
}

type Mapping struct {
	TemplateExternalURL       string
	TemplateHeaders           string
	TemplateJSONBody          string
	TemplateTokenFromResponse string
}

type Oauth struct {
	URL          string
	ClientID     string
	ClientSecret string
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
}

type ResponseData struct {
	Token string
}
