package tests

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/util"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/machinebox/graphql"

	"github.com/pkg/errors"

	"github.com/vrischmann/envconfig"
)

type config struct {
	ExternalServicesMockURL        string        `envconfig:"EXTERNAL_SERVICES_MOCK_URL"`
	ClientID                       string        `envconfig:"CLIENT_ID"`
	ClientSecret                   string        `envconfig:"CLIENT_SECRET"`
	TokenURL                       string        `envconfig:"TOKEN_URL,optional"`
	InstanceURL                    string        `envconfig:"INSTANCE_URL,optional"`
	TokenPath                      string        `envconfig:"TOKEN_PATH,optional"`
	RegisterPath                   string        `envconfig:"REGISTER_PATH,optional"`
	Subaccount                     string        `envconfig:"NS_SUBACCOUNT,optional"`
	CreateClonePattern             string        `envconfig:"CREATE_CLONE_PATTERN,optional"`
	CreateBindingPattern           string        `envconfig:"CREATE_BINDING_PATTERN,optional"`
	DefaultTestTenant              string        `envconfig:"DEFAULT_TEST_TENANT"`
	AdapterURL                     string        `envconfig:"ADAPTER_URL"`
	Timeout                        time.Duration `envconfig:"default=60s"`
	X509Config                     oauth.X509Config
	CertLoaderConfig               certloader.Config
	DirectorExternalCertSecuredURL string
	SkipSSLValidation              bool   `envconfig:"default=false"`
	UseClone                       bool   `envconfig:"default=false,USE_CLONE"`
	ExternalClientCertSecretName   string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET_NAME"`
}

var (
	testConfig               config
	certSecuredGraphQLClient *graphql.Client
)

func TestMain(m *testing.M) {
	err := envconfig.Init(&testConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while initializing envconfig"))
	}

	tenant.TestTenants.Init()

	ctx := context.Background()

	cc, err := certloader.StartCertLoader(ctx, testConfig.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while starting cert cache"))
	}

	if err := util.WaitForCache(cc); err != nil {
		log.D().Fatal(err)
	}

	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(testConfig.DirectorExternalCertSecuredURL, cc.Get()[testConfig.ExternalClientCertSecretName].PrivateKey, cc.Get()[testConfig.ExternalClientCertSecretName].Certificate, testConfig.SkipSSLValidation)

	exitVal := m.Run()
	os.Exit(exitVal)
}

func getTokenFromExternalSVCMock(t *testing.T) string {
	return token.GetClientCredentialsToken(t, context.Background(), testConfig.ExternalServicesMockURL+"/secured/oauth/token", testConfig.ClientID, testConfig.ClientSecret, "nsAdapterClaims")
}

type bindingData struct {
	Certificate string `json:"certificate"`
	Key         string `json:"key"`
	ClientID    string `json:"clientid"`
	Certurl     string `json:"certurl"`
}

func getInstanceName(t *testing.T) string {
	uniqueSuffix, err := uuid.NewUUID()
	require.NoError(t, err)
	instanceName := "cmp-test" + uniqueSuffix.String()
	return instanceName
}

func getTokenFromClone(t *testing.T, instanceName string) string {
	token, err := getTokenFromMasterInstance()
	require.NoError(t, err)

	err = registerClone(token, instanceName)
	require.NoError(t, err)

	binding, err := getBinding(token, instanceName)
	require.NoError(t, err)

	tokenstr, err := getTokenFromCloneWithCertificate(binding)
	require.NoError(t, err)

	return tokenstr
}

func getTokenFromMasterInstance() (string, error) {
	return getTokenWithCertificate(&testConfig.X509Config, testConfig.ClientID, testConfig.TokenURL+testConfig.TokenPath)
}

func registerClone(token, instanceName string) error {
	url, err := urlpkg.Parse(testConfig.InstanceURL + testConfig.RegisterPath)
	if err != nil {
		return errors.Wrap(err, "while parsing url for clone registration")
	}

	q := url.Query()
	q.Add("serviceinstanceid", instanceName)
	q.Add("subaccountid", testConfig.Subaccount)
	url.RawQuery = q.Encode()

	body := strings.NewReader(fmt.Sprintf(testConfig.CreateClonePattern, instanceName))
	req, err := http.NewRequest(http.MethodPost, url.String(), body)
	if err != nil {
		return errors.Wrap(err, "while preparing request for clone registration")
	}

	req.Header.Set("Authorization", token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "while sending request for clone registration")
	}

	if resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrapf(err, "failed to register clone, status code: %d, failed to read response", resp.StatusCode)
		}
		return errors.New(fmt.Sprintf("failed to register clone, status code: %d, response: %s", resp.StatusCode, string(body)))
	}
	return nil
}

func getBinding(token, instanceName string) (bindingData, error) {
	url, err := urlpkg.Parse(testConfig.InstanceURL + testConfig.RegisterPath)
	if err != nil {
		return bindingData{}, errors.Wrap(err, "while parsing url for binding creation")
	}
	url.Path = path.Join(url.Path, instanceName, "binding", instanceName)

	req, err := http.NewRequest(http.MethodPut, url.String(), strings.NewReader(testConfig.CreateBindingPattern))
	if err != nil {
		return bindingData{}, errors.Wrap(err, "while preparing request for binding creation")
	}
	req.Header.Set("Authorization", token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return bindingData{}, errors.Wrap(err, "while sending request for binding creation")
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return bindingData{}, errors.Wrap(err, "while reading response from binding creation request")
	}

	binding := bindingData{}
	if err = json.Unmarshal(responseBody, &binding); err != nil {
		return bindingData{}, err
	}
	return binding, nil
}

func getTokenFromCloneWithCertificate(binding bindingData) (string, error) {
	certConfig := &oauth.X509Config{
		Cert: binding.Certificate,
		Key:  binding.Key,
	}

	return getTokenWithCertificate(certConfig, binding.ClientID, binding.Certurl+testConfig.TokenPath)
}

func getTokenWithCertificate(certConfig *oauth.X509Config, clientID, URL string) (string, error) {
	cert, err := certConfig.ParseCertificate()
	if err != nil {
		return "", errors.Wrap(err, "while parsing certificate from binding")
	}

	mtlClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{*cert},
			},
		},
		Timeout: testConfig.Timeout,
	}

	creds := &auth.OAuthMtlsCredentials{
		ClientID:          clientID,
		CertCache:         nil,
		TokenURL:          URL,
		Scopes:            "",
		AdditionalHeaders: nil,
	}

	ctx := auth.SaveToContext(context.Background(), creds)

	tokenProvider := auth.NewMtlsTokenAuthorizationProviderWithClient(mtlClient)
	authorization, err := tokenProvider.GetAuthorization(ctx)
	if err != nil {
		return "", errors.Wrap(err, "while requesting token with certificate")
	}

	return authorization, nil
}

func deleteClone(t *testing.T, instanceName string) {
	url, err := urlpkg.Parse(testConfig.InstanceURL)
	require.NoError(t, err)

	url.Path = path.Join(url.Path, testConfig.RegisterPath, instanceName)

	request, err := http.NewRequest(http.MethodDelete, url.String(), nil)
	require.NoError(t, err)

	request.Header.Set("Content-Type", "application/json")
}
