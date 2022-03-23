package info

import (
	"context"
	"net/http"

	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"

	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
)

// Config contains the data that should be exported on the info endpoint
type Config struct {
	APIEndpoint string `envconfig:"APP_INFO_API_ENDPOINT,default=/v1/info" json:"-"`
	Issuer      string `envconfig:"APP_INFO_CERT_ISSUER"`
	Subject     string `envconfig:"APP_INFO_CERT_SUBJECT"`
	RootCA      string `envconfig:"APP_INFO_ROOT_CA"`
}

type responseData struct {
	Issuer     string `json:"certIssuer"`
	Subject    string `json:"certSubject"`
	RootCA     string `json:"rootCA"`
	OrdVersion string `json:"ordAggregatorVersion"`
}

func prepareResponseData(c Config) responseData {
	return responseData{
		Issuer:     c.Issuer,
		Subject:    c.Subject,
		RootCA:     c.RootCA,
		OrdVersion: ord.SpecVersion,
	}
}

// NewInfoHandler returns handler which gives information about the CMP client certificate
func NewInfoHandler(ctx context.Context, c Config) func(writer http.ResponseWriter, request *http.Request) {
	responseData := prepareResponseData(c)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		httputils.RespondWithBody(ctx, w, http.StatusOK, responseData)
	}
}
