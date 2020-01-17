package tenantfetcher

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
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

type Client struct {
	httpClient *http.Client

	apiConfig APIConfig
}

const (
	pageSize = 1000
)

func NewClient(oAuth2Config OAuth2Config, apiConfig APIConfig) Client {
	cfg := clientcredentials.Config{
		ClientID:     oAuth2Config.ClientID,
		ClientSecret: oAuth2Config.ClientSecret,
		TokenURL:     oAuth2Config.OAuthTokenEndpoint,
	}
	httpClient := cfg.Client(context.Background())

	return Client{
		httpClient: httpClient,
		apiConfig:  apiConfig,
	}
}

func (c Client) FetchTenantEventsPage(eventsType EventsType, pageNumber int) (*TenantEventsResponse, error) {
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
		return nil, errors.Wrap(err, "while sending get request")
	}
	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Warnf("WARNING: Unable to close response body. Cause: %v", err)
		}
	}()

	var tenantEvents TenantEventsResponse
	err = json.NewDecoder(res.Body).Decode(&tenantEvents)
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, errors.Wrap(err, "while decoding response body")
	}

	return &tenantEvents, nil
}

func (c Client) getEndpointForEventsType(eventsType EventsType) (string, error) {
	switch eventsType {
	case CreatedEventsType:
		return c.apiConfig.EndpointTenantCreated, nil
	case DeletedEventsType:
		return c.apiConfig.EndpointTenantDeleted, nil
	case UpdatedEventsType:
		return c.apiConfig.EndpointTenantUpdated, nil
	default:
		return "", errors.New("unknown events type")
	}
}

func (c Client) buildRequestURL(endpoint string, pageNumber int) (string, error) {
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
