package tenantfetcher

import (
	"context"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/tidwall/gjson"

	pkgErrors "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	maxErrMessageLength = 50
)

type OAuth2Config struct {
	ClientID           string `envconfig:"APP_CLIENT_ID"`
	ClientSecret       string `envconfig:"APP_CLIENT_SECRET"`
	OAuthTokenEndpoint string `envconfig:"APP_OAUTH_TOKEN_ENDPOINT"`
}

type APIConfig struct {
	EndpointTenantCreated string `envconfig:"APP_ENDPOINT_TENANT_CREATED"`
	EndpointTenantDeleted string `envconfig:"APP_ENDPOINT_TENANT_DELETED"`
	EndpointTenantUpdated string `envconfig:"APP_ENDPOINT_TENANT_UPDATED"`

	TotalPagesField   string `envconfig:"APP_TENANT_TOTAL_PAGES_FIELD"`
	TotalResultsField string `envconfig:"APP_TENANT_TOTAL_RESULTS_FIELD"`
	EventsField       string `envconfig:"APP_TENANT_EVENTS_FIELD"`
}

//go:generate mockery -name=MetricsPusher -output=automock -outpkg=automock -case=underscore
type MetricsPusher interface {
	RecordEventingRequest(method string, statusCode int, desc string)
}

// QueryParams describes the key and the corresponding value for query parameters when requesting the service
type QueryParams map[string]string

// tenantResponseMapper describes which fields correspond to what value from the service's response
type tenantResponseMapper struct {
	TotalPagesField   string
	TotalResultsField string
	EventsField       string
}

// Remap returns the actual tenant event response mapped by the provided fields in the struct
func (trd *tenantResponseMapper) Remap(b []byte) TenantEventsResponse {
	events := gjson.GetBytes(b, trd.EventsField)
	if !events.Exists() {
		return TenantEventsResponse{
			Events:       nil,
			TotalResults: 0,
			TotalPages:   1,
		}
	}

	totalPages := gjson.GetBytes(b, trd.TotalPagesField).Int()
	totalResults := gjson.GetBytes(b, trd.TotalResultsField).Int()

	return TenantEventsResponse{
		Events:       []byte(events.Raw),
		TotalPages:   int(totalPages),
		TotalResults: int(totalResults),
	}
}

// Client implements the communication with the service
type Client struct {
	httpClient    *http.Client
	metricsPusher MetricsPusher

	responseMapper tenantResponseMapper
	apiConfig      APIConfig
}

func NewClient(oAuth2Config OAuth2Config, apiConfig APIConfig) *Client {
	cfg := clientcredentials.Config{
		ClientID:     oAuth2Config.ClientID,
		ClientSecret: oAuth2Config.ClientSecret,
		TokenURL:     oAuth2Config.OAuthTokenEndpoint,
	}

	httpClient := cfg.Client(context.Background())

	return &Client{
		httpClient: httpClient,
		apiConfig:  apiConfig,
		responseMapper: tenantResponseMapper{
			EventsField:       apiConfig.EventsField,
			TotalPagesField:   apiConfig.TotalPagesField,
			TotalResultsField: apiConfig.TotalResultsField,
		},
	}
}

func (c *Client) SetMetricsPusher(metricsPusher MetricsPusher) {
	c.metricsPusher = metricsPusher
}

func (c *Client) FetchTenantEventsPage(eventsType EventsType, additionalQueryParams QueryParams) (*TenantEventsResponse, error) {
	endpoint, err := c.getEndpointForEventsType(eventsType)
	if err != nil {
		return nil, err
	}

	reqURL, err := c.buildRequestURL(endpoint, additionalQueryParams)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Get(reqURL)
	if err != nil {
		if c.metricsPusher != nil {
			desc := c.failedRequestDesc(err)
			c.metricsPusher.RecordEventingRequest(http.MethodGet, 0, desc)
		}
		return nil, pkgErrors.Wrap(err, "while sending get request")
	}
	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Warnf("Unable to close response body. Cause: %v", err)
		}
	}()

	if c.metricsPusher != nil {
		c.metricsPusher.RecordEventingRequest(http.MethodGet, res.StatusCode, res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, pkgErrors.Wrap(err, "while reading response body")
	}
	tenantEvents := c.responseMapper.Remap(bytes)

	return &tenantEvents, nil
}

func (c *Client) getEndpointForEventsType(eventsType EventsType) (string, error) {
	switch eventsType {
	case CreatedEventsType:
		return c.apiConfig.EndpointTenantCreated, nil
	case DeletedEventsType:
		return c.apiConfig.EndpointTenantDeleted, nil
	case UpdatedEventsType:
		return c.apiConfig.EndpointTenantUpdated, nil
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

func (c *Client) failedRequestDesc(err error) string {
	var e *net.OpError
	if errors.As(err, &e) && e.Err != nil {
		return e.Err.Error()
	}

	if len(err.Error()) > maxErrMessageLength {
		// not all errors are actually wrapped, sometimes the error message is just concatenated with ":"
		errParts := strings.Split(err.Error(), ":")
		return errParts[len(errParts)-1]
	}

	return err.Error()
}
