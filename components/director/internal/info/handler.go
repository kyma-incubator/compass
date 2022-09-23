package info

import (
	"context"
	"crypto/x509"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"

	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"

	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
)

const (
	plusDelimiter  = "+"
	commaDelimiter = ","
)

// Config contains the data that should be exported on the info endpoint
type Config struct {
	APIEndpoint                  string `envconfig:"APP_INFO_API_ENDPOINT,default=/v1/info" json:"-"`
	RootCA                       string `envconfig:"APP_INFO_ROOT_CA"`
	ExternalClientCertSecretName string `envconfig:"EXTERNAL_CLIENT_CERT_SECRET_NAME"`
}

type responseData struct {
	Issuer     string `json:"certIssuer"`
	Subject    string `json:"certSubject"`
	RootCA     string `json:"rootCA"`
	OrdVersion string `json:"ordAggregatorVersion"`
}

// NewInfoHandler returns handler which gives information about the CMP client certificate.
// The issuer and the subject are dynamically loaded from the certificate itself (reading the cert from the cert cache every time the endpoint is hit) rather than using hardcoded env values.
func NewInfoHandler(ctx context.Context, c Config, certCache certloader.Cache) func(writer http.ResponseWriter, request *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		responseData, err := prepareResponseData(c, certCache)
		if err != nil {
			log.C(ctx).Errorf("Error while processing client certificate from cache: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		httputils.RespondWithBody(ctx, w, http.StatusOK, responseData)
	}
}

func prepareResponseData(c Config, certCache certloader.Cache) (responseData, error) {
	clientCert := certCache.Get()[c.ExternalClientCertSecretName]
	if clientCert == nil || len(clientCert.Certificate) == 0 {
		return responseData{}, errors.New("did not find client certificate in the cache")
	}

	parsedClientCert, err := x509.ParseCertificate(clientCert.Certificate[0])
	if err != nil {
		return responseData{}, errors.New("error while parsing client certificate")
	}

	certIssuer := replaceDelimiter(parsedClientCert.Issuer.String())
	certSubject := replaceDelimiter(parsedClientCert.Subject.String())

	return responseData{
		Issuer:     certIssuer,
		Subject:    certSubject,
		RootCA:     c.RootCA,
		OrdVersion: ord.SpecVersion,
	}, nil
}

func replaceDelimiter(input string) string {
	return strings.ReplaceAll(input, plusDelimiter, commaDelimiter)
}
