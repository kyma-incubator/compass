package clientset

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"

	"k8s.io/client-go/kubernetes"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/connector/internal/httputils"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"

	"github.com/kyma-incubator/compass/components/connector/config"

	"github.com/kyma-incubator/compass/components/connector/internal/api"
	"github.com/kyma-incubator/compass/components/connector/internal/authentication"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"

	"k8s.io/client-go/kubernetes/fake"
)

const (
	caCertFile = "testdata/ca_crt.pem"
	caKeyFile  = "testdata/ca_key.pem"

	testSecretName    = "test-secret"
	testConfigMapName = "test-secret"
)

var (
	externalAPIUrl string
	tokenService   tokens.Service
	k8sClientSet   kubernetes.Interface
)

func TestMain(m *testing.M) {
	err := os.Setenv("APP_CA_SECRET_NAME", testSecretName)
	exitOnError(err, "Error setting APP_CA_SECRET_NAME env")
	err = os.Setenv("APP_REVOCATION_CONFIG_MAP_NAME", testConfigMapName)
	exitOnError(err, "Error setting APP_CA_SECRET_NAME env")

	cfg := config.Config{}
	err = envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app Config")

	caCertificate, err := ioutil.ReadFile(caCertFile)
	exitOnError(err, "Error reading CA cert file")
	caKey, err := ioutil.ReadFile(caKeyFile)
	exitOnError(err, "Error reading CA key file")

	k8sClientSet = fake.NewSimpleClientset(
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "test-secret", Namespace: "default"},
			Data: map[string][]byte{
				"ca.crt": []byte(caCertificate),
				"ca.key": []byte(caKey),
			},
		},
		&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: testSecretName, Namespace: "default"},
			Data:       nil,
			BinaryData: nil,
		},
	)

	internalComponents := config.InitInternalComponents(cfg, k8sClientSet)

	tokenService = internalComponents.TokenService
	externalAPIUrl = fmt.Sprintf("https://%s%s", cfg.ExternalAddress, cfg.APIEndpoint)

	certificateResolver := api.NewCertificateResolver(
		internalComponents.Authenticator,
		internalComponents.TokenService,
		internalComponents.CertificateService,
		internalComponents.SubjectConsts,
		cfg.DirectorURL,
		cfg.CertificateSecuredConnectorURL,
		internalComponents.RevocationsRepository)

	authContextTestMiddleware := func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			connectorToken := r.Header.Get(oathkeeper.ConnectorTokenHeader)
			if connectorToken != "" {
				tokenData, err := tokenService.Resolve(connectorToken)
				if err != nil {
					httputils.RespondWithError(w, http.StatusForbidden, err)
					return
				}
				r = r.WithContext(authentication.PutIntoContext(r.Context(), authentication.ClientIdFromTokenKey, tokenData.ClientId))

				handler.ServeHTTP(w, r)
			}

			if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
				httputils.RespondWithError(w, http.StatusForbidden, errors.New("Token and Certificate not provided"))
				return
			}

			clientId := r.TLS.PeerCertificates[0].Subject.CommonName
			r = r.WithContext(authentication.PutIntoContext(r.Context(), authentication.ClientIdFromCertificateKey, clientId))
			r = r.WithContext(authentication.PutIntoContext(r.Context(), authentication.ClientCertificateHashKey, "hash"))

			handler.ServeHTTP(w, r)
		})
	}

	externalGqlServer := config.PrepareExternalGraphQLServer(cfg, certificateResolver, authContextTestMiddleware)
	externalGqlServer.TLSConfig = &tls.Config{ClientAuth: tls.RequestClientCert}

	go func() {
		if err := externalGqlServer.ListenAndServeTLS("testdata/ca_crt.pem", "testdata/ca_key.pem"); err != nil {
			panic(err)
		}
	}()

	time.Sleep(5 * time.Second)

	code := m.Run()
	os.Exit(code)
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}
