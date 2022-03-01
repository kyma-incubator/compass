package clientset

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/connector/config"
	"github.com/kyma-incubator/compass/components/connector/internal/api"
	"github.com/kyma-incubator/compass/components/connector/internal/authentication"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	gcliMocks "github.com/kyma-incubator/compass/components/connector/internal/tokens/automock"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/vrischmann/envconfig"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	caCertFile = "testdata/ca_crt.pem"
	caKeyFile  = "testdata/ca_key.pem"

	testSecretName    = "test-secret"
	testConfigMapName = "test-secret"
	oneTimeTokenURL   = "http://director.com"
	clientID          = "abcd-efgh"
)

var (
	externalAPIUrl string
	k8sClientSet   kubernetes.Interface
)

func TestMain(m *testing.M) {
	err := os.Setenv("APP_CA_SECRET_NAME", testSecretName)
	exitOnError(err, "Error setting APP_CA_SECRET_NAME env")
	err = os.Setenv("APP_REVOCATION_CONFIG_MAP_NAME", testConfigMapName)
	exitOnError(err, "Error setting APP_CA_SECRET_NAME env")
	err = os.Setenv("APP_ONE_TIME_TOKEN_URL", oneTimeTokenURL)
	exitOnError(err, "Error setting APP_ONE_TIME_TOKEN_URL env")

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

	directorGCLI := &gcliMocks.GraphQLClient{}
	directorGCLI.On("Run", mock.Anything, mock.Anything, mock.Anything).
		Run(GenerateTestToken(tokens.NewTokenResponse("abcd"))).Return(nil).Twice()
	internalComponents, certsLoader := config.InitInternalComponents(cfg, k8sClientSet, directorGCLI)

	go certsLoader.Run(context.TODO())

	externalAPIUrl = fmt.Sprintf("https://%s%s", cfg.ExternalAddress, cfg.APIEndpoint)

	certificateResolver := api.NewCertificateResolver(
		internalComponents.Authenticator,
		internalComponents.TokenService,
		internalComponents.CertificateService,
		internalComponents.CSRSubjectConsts,
		cfg.DirectorURL,
		cfg.CertificateSecuredConnectorURL,
		internalComponents.RevokedCertsRepository)

	authContextTestMiddleware := func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(authentication.PutIntoContext(r.Context(), authentication.ConsumerType, "Application"))
			r = r.WithContext(authentication.PutIntoContext(r.Context(), authentication.TenantKey, "tenant"))

			connectorToken := r.Header.Get(oathkeeper.ConnectorTokenHeader)
			if connectorToken != "" {

				r = r.WithContext(authentication.PutIntoContext(r.Context(), authentication.ClientIdFromTokenKey, clientID))

				handler.ServeHTTP(w, r)
				return
			}

			if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
				httputils.RespondWithError(r.Context(), w, http.StatusForbidden, errors.New("Token and Certificate not provided"))
				return
			}

			clientId := r.TLS.PeerCertificates[0].Subject.CommonName
			r = r.WithContext(authentication.PutIntoContext(r.Context(), authentication.ClientIdFromCertificateKey, clientId))
			r = r.WithContext(authentication.PutIntoContext(r.Context(), authentication.ClientCertificateHashKey, "hash"))

			handler.ServeHTTP(w, r)
		})
	}

	externalGqlServer, err := config.PrepareExternalGraphQLServer(cfg, certificateResolver, authContextTestMiddleware)
	exitOnError(err, "Error configuring external graphQL handler")

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

func GenerateTestToken(generated tokens.TokenResponse) func(args mock.Arguments) {
	return func(args mock.Arguments) {
		arg, ok := args.Get(2).(*tokens.TokenResponse)
		if !ok {
			log.Fatal("could not cast CSRTokenResponse")
		}
		*arg = generated
	}
}
