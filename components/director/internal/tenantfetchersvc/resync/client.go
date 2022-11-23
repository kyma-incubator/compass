package resync

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
	bndlErrors "github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// OAuth2Config is the auth configuration used by Tenant Events API clients.
type OAuth2Config struct {
	X509Config
	ClientID           string
	ClientSecret       string
	OAuthTokenEndpoint string
	TokenPath          string
	SkipSSLValidation  bool
}

// AuthProviderConfig  is the configuration of the authentication secret used by tenants aggregator.
// The auth secret contains auth details for different regions. Each region reads its auth config from the secret file by given a specific key.
type AuthProviderConfig struct {
	AuthMappingConfig

	SecretFilePath    string `envconfig:"FILE_PATH" default:"/tmp/keyConfig"`
	TokenPath         string `envconfig:"TOKEN_PATH" required:"true"`
	SkipSSLValidation bool   `envconfig:"OAUTH_SKIP_SSL_VALIDATION" default:"false"`
}

// AuthMappingConfig is the mapping configuration between auth details and their paths in the auth secret file.
type AuthMappingConfig struct {
	ClientIDPath      string `envconfig:"CLIENT_ID_PATH" required:"true"`
	ClientSecretPath  string `envconfig:"CLIENT_SECRET_PATH"`
	TokenEndpointPath string `envconfig:"TOKEN_ENDPOINT_PATH" required:"true"`
	CertPath          string `envconfig:"CERT_PATH"`
	KeyPath           string `envconfig:"CERT_KEY_PATH"`
}

// Validate checks if the configuration is considered valid against the given auth mode.
// The configuration is considered valid if it contains all needed auth details for the given mode.
func (c OAuth2Config) Validate(oauthMode oauth.AuthMode) error {
	missingProperties := make([]string, 0)
	if len(c.ClientID) == 0 {
		missingProperties = append(missingProperties, "ClientID")
	}
	if len(c.OAuthTokenEndpoint) == 0 {
		missingProperties = append(missingProperties, "OAuthTokenEndpoint")
	}

	switch oauthMode {
	case oauth.Standard:
		if len(c.ClientSecret) == 0 {
			missingProperties = append(missingProperties, "ClientSecret")
		}
	case oauth.Mtls:
		if len(c.Cert) == 0 {
			missingProperties = append(missingProperties, "Certificate")
		}
		if len(c.Key) == 0 {
			missingProperties = append(missingProperties, "CertificateKey")
		}
	}

	if len(missingProperties) > 0 {
		return fmt.Errorf("missing API Client Auth config properties: %s", strings.Join(missingProperties, ","))
	}

	return nil
}

// X509Config is X509 configuration for getting an OAuth token via mtls
// same as struct in pkg/oauth but with different envconfig
type X509Config struct {
	Cert string
	Key  string
}

// ParseCertificate parses the TLS certificate contained in the X509Config
func (c *X509Config) ParseCertificate() (*tls.Certificate, error) {
	return cert.ParseCertificate(c.Cert, c.Key)
}

// APIEndpointsConfig missing godoc
type APIEndpointsConfig struct {
	EndpointTenantCreated     string `envconfig:"ENDPOINT_TENANT_CREATED"`
	EndpointTenantDeleted     string `envconfig:"ENDPOINT_TENANT_DELETED"`
	EndpointTenantUpdated     string `envconfig:"ENDPOINT_TENANT_UPDATED"`
	EndpointSubaccountCreated string `envconfig:"ENDPOINT_SUBACCOUNT_CREATED"`
	EndpointSubaccountDeleted string `envconfig:"ENDPOINT_SUBACCOUNT_DELETED"`
	EndpointSubaccountUpdated string `envconfig:"ENDPOINT_SUBACCOUNT_UPDATED"`
	EndpointSubaccountMoved   string `envconfig:"ENDPOINT_SUBACCOUNT_MOVED"`
}

func (c APIEndpointsConfig) isUnassignedOptionalProperty(eventsType EventsType) bool {
	return eventsType == MovedSubaccountType && len(c.EndpointSubaccountMoved) == 0
}

// QueryParams describes the key and the corresponding value for query parameters when requesting the service
type QueryParams map[string]string

// Client implements the communication with the service
type Client struct {
	config     ClientConfig
	httpClient *http.Client
}

// ClientConfig is the client specific configuration of the Events API
type ClientConfig struct {
	TenantProvider      string
	APIConfig           APIEndpointsConfig
	FieldMapping        TenantFieldMapping
	MovedSAFieldMapping MovedSubaccountsFieldMapping
}

// NewClient missing godoc
func NewClient(oAuth2Config OAuth2Config, authMode oauth.AuthMode, clientConfig ClientConfig, timeout time.Duration) (*Client, error) {
	ctx := context.Background()
	cfg := clientcredentials.Config{
		ClientID:     oAuth2Config.ClientID,
		ClientSecret: oAuth2Config.ClientSecret,
		TokenURL:     oAuth2Config.OAuthTokenEndpoint + oAuth2Config.TokenPath,
	}

	switch authMode {
	case oauth.Standard:
		// do nothing
	case oauth.Mtls:
		cert, err := oAuth2Config.X509Config.ParseCertificate()
		if nil != err {
			return nil, err
		}

		// When the auth style is InParams, the TokenSource
		// will not add the clientSecret if it's empty
		cfg.AuthStyle = oauth2.AuthStyleInParams
		cfg.ClientSecret = ""

		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{*cert},
				InsecureSkipVerify: oAuth2Config.SkipSSLValidation,
			},
		}

		mtlClient := &http.Client{
			Transport: transport,
			Timeout:   timeout,
		}

		ctx = context.WithValue(ctx, oauth2.HTTPClient, mtlClient)
	default:
		return nil, errors.New("unsupported auth mode:" + string(authMode))
	}

	httpClient := cfg.Client(ctx)
	httpClient.Timeout = timeout

	return &Client{
		httpClient: httpClient,
		config:     clientConfig,
	}, nil
}

// FetchTenantEventsPage missing godoc
func (c *Client) FetchTenantEventsPage(ctx context.Context, eventsType EventsType, additionalQueryParams QueryParams) (*EventsPage, error) {
	if c.config.APIConfig.isUnassignedOptionalProperty(eventsType) {
		log.C(ctx).Warnf("Optional property for event type %s was not set", eventsType)
		return nil, nil
	}

	endpoint, err := c.getEndpointForEventsType(eventsType)
	if endpoint == "" && err == nil {
		log.C(ctx).Warnf("Endpoint for event %s is not set", eventsType)
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	reqURL, err := c.buildRequestURL(endpoint, additionalQueryParams)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, bndlErrors.Wrap(err, "while sending get request")
	}
	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.C(ctx).Warnf("Unable to close response body. Cause: %v", err)
		}
	}()

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, bndlErrors.Wrap(err, "while reading response body")
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("request to %q returned status code %d and body %q", reqURL, res.StatusCode, bytes)
	}

	if len(bytes) == 0 {
		return nil, nil
	}

	return &EventsPage{
		FieldMapping:                 c.config.FieldMapping,
		MovedSubaccountsFieldMapping: c.config.MovedSAFieldMapping,
		ProviderName:                 c.config.TenantProvider,
		Payload:                      bytes,
	}, nil
}

// SetHTTPClient sets the underlying HTTP client
func (c *Client) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

// GetHTTPClient returns the underlying HTTP client
func (c *Client) GetHTTPClient() *http.Client {
	return c.httpClient
}

func (c *Client) getEndpointForEventsType(eventsType EventsType) (string, error) {
	switch eventsType {
	case CreatedAccountType:
		return c.config.APIConfig.EndpointTenantCreated, nil
	case DeletedAccountType:
		return c.config.APIConfig.EndpointTenantDeleted, nil
	case UpdatedAccountType:
		return c.config.APIConfig.EndpointTenantUpdated, nil
	case CreatedSubaccountType:
		return c.config.APIConfig.EndpointSubaccountCreated, nil
	case DeletedSubaccountType:
		return c.config.APIConfig.EndpointSubaccountDeleted, nil
	case UpdatedSubaccountType:
		return c.config.APIConfig.EndpointSubaccountUpdated, nil
	case MovedSubaccountType:
		return c.config.APIConfig.EndpointSubaccountMoved, nil
	default:
		return "", apperrors.NewInternalError("unknown events type")
	}
}

func (c *Client) buildRequestURL(endpoint string, queryParams QueryParams) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", err
	}

	for qKey, qValue := range queryParams {
		q.Add(qKey, qValue)
	}

	u.RawQuery = q.Encode()

	return u.String(), nil
}
