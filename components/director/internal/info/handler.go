package info

import (
	"context"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
)

type Config struct {
	APIEndpoint string `envconfig:"APP_INFO_API_ENDPOINT,default=/v1/info" json:"-"`
	Issuer      string `envconfig:"APP_INFO_CERT_ISSUER" json:"certIssuer"`
	Subject     string `envconfig:"APP_INFO_CERT_SUBJECT" json:"certSubject"`
}

// NewInfoHandler returns handler returns an endpoint which gives information about the CMP client certificate
func NewInfoHandler(ctx context.Context, c Config) func(writer http.ResponseWriter, request *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		httputils.RespondWithBody(ctx, w, http.StatusOK, c)
	}
}
