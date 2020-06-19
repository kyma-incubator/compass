package tenantfetcher

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	pkgErrors "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"
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
}

//go:generate mockery -name=MetricsPusher -output=automock -outpkg=automock -case=underscore
type MetricsPusher interface {
	RecordEventingRequest(method string, statusCode int, desc string)
}

type Client struct {
	httpClient    *http.Client
	metricsPusher MetricsPusher

	apiConfig APIConfig
}

const (
	pageSize            = 1000
	maxErrMessageLength = 50
)

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
	}
}

func (c *Client) SetMetricsPusher(metricsPusher MetricsPusher) {
	c.metricsPusher = metricsPusher
}

func (c *Client) FetchTenantEventsPage(eventsType EventsType, pageNumber int) (*TenantEventsResponse, error) {
	endpoint, err := c.getEndpointForEventsType(eventsType)
	if err != nil {
		return nil, err
	}

	reqURL, err := c.buildRequestURL(endpoint, pageNumber)
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

	var tenantEvents TenantEventsResponse
	err = json.NewDecoder(res.Body).Decode(&tenantEvents)
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, pkgErrors.Wrap(err, "while decoding response body")
	}

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

func (c *Client) buildRequestURL(endpoint string, pageNumber int) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", err
	}

	q.Add("ts", "1")
	q.Add("resultsPerPage", strconv.Itoa(pageSize))
	q.Add("page", strconv.Itoa(pageNumber))

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
