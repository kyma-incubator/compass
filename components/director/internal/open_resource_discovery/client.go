package ord

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/accessstrategy"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// Client represents ORD documents client
//go:generate mockery --name=Client --output=automock --outpkg=automock --case=underscore
type Client interface {
	FetchOpenResourceDiscoveryDocuments(ctx context.Context, url string) (Documents, error)
}

type client struct {
	*http.Client
}

// NewClient creates new ORD Client via a provided http.Client
func NewClient(httpClient *http.Client) *client {
	return &client{
		Client: httpClient,
	}
}

// FetchOpenResourceDiscoveryDocuments fetches all the documents for a single ORD .well-known endpoint
func (c *client) FetchOpenResourceDiscoveryDocuments(ctx context.Context, url string) (Documents, error) {
	config, err := c.fetchConfig(ctx, url)
	if err != nil {
		return nil, err
	}

	baseURL, err := stripRelativePathFromURL(url)
	if err != nil {
		return nil, err
	}

	docs := make([]*Document, 0)
	for _, docDetails := range config.OpenResourceDiscoveryV1.Documents {
		strategy, ok := docDetails.AccessStrategies.GetSupported()
		if !ok {
			log.C(ctx).Warnf("Unsupported access strategies for ORD Document %q", baseURL+docDetails.URL)
			continue
		}
		doc, err := c.fetchOpenDiscoveryDocumentWithAccessStrategy(ctx, baseURL+docDetails.URL, strategy)
		if err != nil {
			return nil, errors.Wrapf(err, "error fetching ORD document from: %s", baseURL+docDetails.URL)
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

func (c *client) fetchOpenDiscoveryDocumentWithAccessStrategy(ctx context.Context, documentURL string, accessStrategy accessstrategy.AccessStrategyType) (*Document, error) {
	log.C(ctx).Infof("Fetching ORD Document %q with Access Strategy %q", documentURL, accessStrategy)
	resp, err := accessStrategy.Execute(ctx, c.Client, documentURL)
	if err != nil {
		return nil, err
	}

	defer closeBody(ctx, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("error while fetching open resource discovery document %q: status code %d", documentURL, resp.StatusCode)
	}

	resp.Body = http.MaxBytesReader(nil, resp.Body, 2097152)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading document body")
	}
	result := &Document{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling document")
	}
	return result, nil
}

func closeBody(ctx context.Context, body io.ReadCloser) {
	if err := body.Close(); err != nil {
		log.C(ctx).WithError(err).Warnf("Got error on closing response body")
	}
}

func (c *client) fetchConfig(ctx context.Context, url string) (*WellKnownConfig, error) {
	configURL, err := buildWellKnownEndpoint(url)
	if err != nil {
		return nil, err
	}

	resp, err := c.Get(configURL)
	if err != nil {
		return nil, errors.Wrap(err, "error while fetching open resource discovery well-known configuration")
	}

	defer closeBody(ctx, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("error while fetching open resource discovery well-known configuration: status code %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}

	config := WellKnownConfig{}
	if err := json.Unmarshal(bodyBytes, &config); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling json body")
	}

	return &config, nil
}

func buildWellKnownEndpoint(u string) (string, error) {
	parsedURL, err := url.ParseRequestURI(u)
	if err != nil {
		return "", errors.New("error while parsing input webhook url")
	}

	if parsedURL.Path != "" {
		return parsedURL.String(), nil
	}
	return parsedURL.String() + WellKnownEndpoint, nil
}

func stripRelativePathFromURL(u string) (string, error) {
	parsedURL, err := url.ParseRequestURI(u)
	if err != nil {
		return "", errors.New("error while parsing input webhook url")
	}

	parsedURL.Path = ""
	return parsedURL.String(), nil
}
