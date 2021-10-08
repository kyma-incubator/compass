package ord

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/accessstrategy"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/pkg/errors"
)

// Client represents ORD documents client
//go:generate mockery --name=Client --output=automock --outpkg=automock --case=underscore
type Client interface {
	FetchOpenResourceDiscoveryDocuments(ctx context.Context, app *model.Application, webhook *model.Webhook) (Documents, error)
}

type client struct {
	*http.Client
	securedApplicationTypes map[string]struct{}
}

// NewClient creates new ORD Client via a provided http.Client
func NewClient(httpClient *http.Client, securedApplicationTypes []string) *client {
	return &client{
		Client:                  httpClient,
		securedApplicationTypes: str.SliceToMap(securedApplicationTypes),
	}
}

// FetchOpenResourceDiscoveryDocuments fetches all the documents for a single ORD .well-known endpoint
func (c *client) FetchOpenResourceDiscoveryDocuments(ctx context.Context, app *model.Application, webhook *model.Webhook) (Documents, error) {
	config, err := c.fetchConfig(ctx, app, webhook)
	if err != nil {
		return nil, err
	}

	baseURL, err := stripRelativePathFromURL(*webhook.URL)
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

func (c *client) fetchConfig(ctx context.Context, app *model.Application, webhook *model.Webhook) (*WellKnownConfig, error) {
	configURL, err := buildWellKnownEndpoint(*webhook.URL)
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	if _, secured := c.securedApplicationTypes[app.Type]; secured {
		log.C(ctx).Infof("Application %q (id = %q, type = %q) configuration endpoint is secured and webhook credentials will be used", app.Name, app.ID, app.Type)
		resp, err = httputil.GetRequestWithCredentials(ctx, c.Client, configURL, webhook.Auth)
		if err != nil {
			return nil, errors.Wrap(err, "error while fetching open resource discovery well-known configuration with webhook credentials")
		}
	} else {
		log.C(ctx).Infof("Application %q (id = %q, type = %q) configuration endpoint is not secured", app.Name, app.ID, app.Type)
		resp, err = httputil.GetRequestWithoutCredentials(c.Client, configURL)
		if err != nil {
			return nil, errors.Wrap(err, "error while fetching open resource discovery well-known configuration")
		}
	}

	defer closeBody(ctx, resp.Body)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("error while fetching open resource discovery well-known configuration: status code %d Body: %s", resp.StatusCode, string(bodyBytes))
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
