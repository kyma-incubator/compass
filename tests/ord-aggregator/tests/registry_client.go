package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type client struct {
	*http.Client
	accessStrategyExecutorProvider accessstrategy.ExecutorProvider
}

func NewGlobalRegistryClient(httpClient *http.Client, accessStrategyExecutorProvider accessstrategy.ExecutorProvider) *client {
	return &client{
		Client:                         httpClient,
		accessStrategyExecutorProvider: accessStrategyExecutorProvider,
	}
}

func (c *client) GetGlobalProductsAndVendorsNumber(ctx context.Context, globalRegistryURL string) (int, int, error) {
	config, err := c.fetchConfig(ctx, globalRegistryURL)
	if err != nil {
		return 0, 0, err
	}

	baseURL, err := calculateBaseURL(globalRegistryURL, *config)
	if err != nil {
		return 0, 0, errors.Wrap(err, "while calculating baseURL")
	}

	globalProducts := 0
	globalVendors := 0
	for _, docDetails := range config.OpenResourceDiscoveryV1.Documents {
		documentURL, err := buildDocumentURL(docDetails.URL, baseURL)
		if err != nil {
			return globalProducts, globalVendors, errors.Wrap(err, "error building document URL")
		}
		strategy, ok := docDetails.AccessStrategies.GetSupported()
		if !ok {
			return globalProducts, globalVendors, fmt.Errorf("unsupported access strategies for ORD Document %q", documentURL)
		}
		doc, err := c.fetchOpenDiscoveryDocumentWithAccessStrategy(ctx, documentURL, strategy)
		if err != nil {
			return globalProducts, globalVendors, errors.Wrapf(err, "error fetching ORD document from: %s", documentURL)
		}

		numberOfProducts := len(gjson.Get(doc, "products").Array())
		globalProducts += numberOfProducts

		numberOfVendors := len(gjson.Get(doc, "vendors").Array())
		globalVendors += numberOfVendors
	}

	return globalProducts, globalVendors, nil
}

func (c *client) fetchOpenDiscoveryDocumentWithAccessStrategy(ctx context.Context, documentURL string, accessStrategy accessstrategy.Type) (string, error) {
	log.C(ctx).Infof("Fetching ORD Document %q with Access Strategy %q", documentURL, accessStrategy)
	executor, err := c.accessStrategyExecutorProvider.Provide(accessStrategy)
	if err != nil {
		return "", err
	}

	resp, err := executor.Execute(ctx, c.Client, documentURL, "")
	if err != nil {
		return "", err
	}

	defer closeBody(ctx, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("error while fetching open resource discovery document %q: status code %d", documentURL, resp.StatusCode)
	}

	resp.Body = http.MaxBytesReader(nil, resp.Body, 2097152)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "error reading document body")
	}
	return string(bodyBytes), nil
}

func closeBody(ctx context.Context, body io.ReadCloser) {
	if err := body.Close(); err != nil {
		log.C(ctx).WithError(err).Warnf("Got error on closing response body")
	}
}

func (c *client) fetchConfig(ctx context.Context, globalRegistryURL string) (*WellKnownConfig, error) {
	var resp *http.Response
	var err error
	resp, err = httputil.GetRequestWithoutCredentials(c.Client, globalRegistryURL, "")
	if err != nil {
		return nil, errors.Wrap(err, "error while fetching open resource discovery well-known configuration")
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

func buildDocumentURL(docURL, baseURL string) (string, error) {
	docURLParsed, err := url.Parse(docURL)
	if err != nil {
		return "", err
	}
	if docURLParsed.IsAbs() {
		return docURL, nil
	}
	return baseURL + docURL, nil
}

func calculateBaseURL(webhookURL string, config WellKnownConfig) (string, error) {
	if config.BaseURL != "" {
		return config.BaseURL, nil
	}

	parsedWebhookURL, err := url.ParseRequestURI(webhookURL)
	if err != nil {
		return "", errors.New("error while parsing input webhook url")
	}

	if strings.HasSuffix(parsedWebhookURL.Path, WellKnownEndpoint) {
		strippedPath := strings.ReplaceAll(parsedWebhookURL.Path, WellKnownEndpoint, "")
		parsedWebhookURL.Path = strippedPath
		return parsedWebhookURL.String(), nil
	}
	return "", nil
}

//comm
