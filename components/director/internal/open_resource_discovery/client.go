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
	FetchOpenResourceDiscoveryDocuments(ctx context.Context, app *model.Application, webhook *model.Webhook) (Documents, string, error)
}

type client struct {
	*http.Client
	securedApplicationTypes        map[string]struct{}
	accessStrategyExecutorProvider accessstrategy.ExecutorProvider
}

// NewClient creates new ORD Client via a provided http.Client
func NewClient(httpClient *http.Client, securedApplicationTypes []string, accessStrategyExecutorProvider accessstrategy.ExecutorProvider) *client {
	return &client{
		Client:                         httpClient,
		securedApplicationTypes:        str.SliceToMap(securedApplicationTypes),
		accessStrategyExecutorProvider: accessStrategyExecutorProvider,
	}
}

// FetchOpenResourceDiscoveryDocuments fetches all the documents for a single ORD .well-known endpoint
func (c *client) FetchOpenResourceDiscoveryDocuments(ctx context.Context, app *model.Application, webhook *model.Webhook) (Documents, string, error) {
	config, err := c.fetchConfig(ctx, app, webhook)
	if err != nil {
		return nil, "", err
	}

	whBaseURL, err := stripRelativePathFromURL(*webhook.URL)
	if err != nil {
		return nil, "", err
	}

	baseURL := whBaseURL
	if len(config.BaseURL) > 0 {
		baseURL = config.BaseURL
	}

	docs := make([]*Document, 0)
	for _, docDetails := range config.OpenResourceDiscoveryV1.Documents {
		documentURL, err := buildDocumentURL(docDetails.URL, baseURL)
		if err != nil {
			return nil, "", errors.Wrap(err, "error building document URL")
		}
		strategy, ok := docDetails.AccessStrategies.GetSupported()
		if !ok {
			log.C(ctx).Warnf("Unsupported access strategies for ORD Document %q", documentURL)
			continue
		}
		doc, err := c.fetchOpenDiscoveryDocumentWithAccessStrategy(ctx, documentURL, strategy)
		if err != nil {
			return nil, "", errors.Wrapf(err, "error fetching ORD document from: %s", documentURL)
		}

		docs = append(docs, doc)
	}

	return docs, baseURL, nil
}

func (c *client) fetchOpenDiscoveryDocumentWithAccessStrategy(ctx context.Context, documentURL string, accessStrategy accessstrategy.Type) (*Document, error) {
	log.C(ctx).Infof("Fetching ORD Document %q with Access Strategy %q", documentURL, accessStrategy)
	executor, err := c.accessStrategyExecutorProvider.Provide(accessStrategy)
	if err != nil {
		return nil, err
	}

	resp, err := executor.Execute(ctx, c.Client, documentURL)
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
	if webhook.Auth != nil && webhook.Auth.AccessStrategy != nil && len(*webhook.Auth.AccessStrategy) > 0 {
		log.C(ctx).Infof("Application %q (id = %q, type = %q) ORD webhook is configured with %q access strategy.", app.Name, app.ID, app.Type, *webhook.Auth.AccessStrategy)
		executor, err := c.accessStrategyExecutorProvider.Provide(accessstrategy.Type(*webhook.Auth.AccessStrategy))
		if err != nil {
			return nil, errors.Wrapf(err, "cannot find executor for access strategy %q as part of webhook processing", *webhook.Auth.AccessStrategy)
		}
		resp, err = executor.Execute(ctx, c.Client, configURL)
		if err != nil {
			return nil, errors.Wrapf(err, "error while fetching open resource discovery well-known configuration with access strategy %q", *webhook.Auth.AccessStrategy)
		}
	} else if _, secured := c.securedApplicationTypes[app.Type]; secured {
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

func buildDocumentURL(docURL, baseURL string) (string, error) { // TODO: tests when the whole algorithm is implemented
	docURLParsed, err := url.Parse(docURL)
	if err != nil {
		return "", err
	}
	if docURLParsed.IsAbs() {
		return docURL, nil
	}
	return baseURL + docURL, nil
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
