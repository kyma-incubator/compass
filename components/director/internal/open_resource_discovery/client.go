package ord

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/pkg/errors"
)

const (
	s4HanaDomain = "s4hana.ondemand"
)

// ClientConfig contains configuration for the ORD aggregator client
type ClientConfig struct {
	maxParallelDocumentsPerApplication int
}

// NewClientConfig creates new ClientConfig from the supplied parameters
func NewClientConfig(maxParallelDocumentsPerApplication int) ClientConfig {
	return ClientConfig{
		maxParallelDocumentsPerApplication: maxParallelDocumentsPerApplication,
	}
}

// Client represents ORD documents client
//
//go:generate mockery --name=Client --output=automock --outpkg=automock --case=underscore --disable-version-string
type Client interface {
	FetchOpenResourceDiscoveryDocuments(ctx context.Context, app *model.Application, webhook *model.Webhook) (Documents, string, error)
}

type client struct {
	config ClientConfig
	*http.Client
	accessStrategyExecutorProvider accessstrategy.ExecutorProvider
	ORDProxyBaseURL                string
}

// NewClient creates new ORD Client via a provided http.Client
func NewClient(config ClientConfig, httpClient *http.Client, accessStrategyExecutorProvider accessstrategy.ExecutorProvider, ORDProxyBaseURL string) *client {
	return &client{
		config:                         config,
		Client:                         httpClient,
		accessStrategyExecutorProvider: accessStrategyExecutorProvider,
		ORDProxyBaseURL:                ORDProxyBaseURL,
	}
}

// FetchOpenResourceDiscoveryDocuments fetches all the documents for a single ORD .well-known endpoint
func (c *client) FetchOpenResourceDiscoveryDocuments(ctx context.Context, app *model.Application, webhook *model.Webhook) (Documents, string, error) {
	var tenantValue string

	if needsTenantHeader := webhook.ObjectType == model.ApplicationTemplateWebhookReference; needsTenantHeader {
		tntFromCtx, err := tenant.LoadTenantPairFromContext(ctx)
		if err != nil {
			return nil, "", errors.Wrapf(err, "while loading tenant from context for application template webhook flow")
		}

		tenantValue = tntFromCtx.ExternalID
	}

	baseURL := str.PtrStrToStr(app.BaseURL)
	systemTenantHeader := prepareS4SystemTenantHeader(baseURL)

	config, err := c.fetchConfig(ctx, app, webhook, tenantValue, systemTenantHeader)
	if err != nil {
		return nil, "", errors.Wrap(err, "while fetching config")
	}
	fmt.Printf("\n config: %+v\n", config)

	m, _ := json.MarshalIndent(webhook, "  ", "")
	fmt.Println(string(m))

	webhookURL, err := calculateBaseURL(webhook, *config)
	if err != nil {
		return nil, "", errors.Wrap(err, "while calculating baseURL")
	}

	err = config.Validate(webhookURL)
	if err != nil {
		return nil, "", errors.Wrap(err, "while validating ORD config")
	}

	docs := make([]*Document, 0)
	docMutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	workers := make(chan struct{}, c.config.maxParallelDocumentsPerApplication)
	fetchDocErrors := make([]error, 0)
	errMutex := sync.Mutex{}

	for _, docDetails := range config.OpenResourceDiscoveryV1.Documents {
		wg.Add(1)
		workers <- struct{}{}
		go func(docDetails DocumentDetails) {
			defer func() {
				wg.Done()
				<-workers
			}()

			documentURL, err := buildDocumentURL(docDetails.URL, baseURL, str.PtrStrToStr(webhook.ProxyURL), c.ORDProxyBaseURL)
			fmt.Println(fmt.Sprintf("\nBuilded documentURL %+v\n", documentURL))
			if err != nil {
				log.C(ctx).Warn(errors.Wrap(err, "error building document URL").Error())
				addError(&fetchDocErrors, err, &errMutex)
				return
			}
			strategy, ok := docDetails.AccessStrategies.GetSupported()
			if !ok {
				log.C(ctx).Warnf("Unsupported access strategies for ORD Document %q", documentURL)
			}
			doc, err := c.fetchOpenDiscoveryDocumentWithAccessStrategy(ctx, documentURL, strategy, tenantValue, systemTenantHeader)
			if err != nil {
				log.C(ctx).Warn(errors.Wrapf(err, "error fetching ORD document from: %s", documentURL).Error())
				addError(&fetchDocErrors, err, &errMutex)
				return
			}

			addDocument(&docs, doc, &docMutex)
		}(docDetails)
	}

	wg.Wait()

	var fetchDocErr error = nil
	if len(fetchDocErrors) > 0 {
		stringErrors := convertErrorsToStrings(fetchDocErrors)
		fetchDocErr = errors.Errorf(strings.Join(stringErrors, "\n"))
	}
	return docs, baseURL, fetchDocErr
}

func convertErrorsToStrings(errors []error) (result []string) {
	for _, err := range errors {
		result = append(result, err.Error())
	}
	return result
}

func (c *client) fetchOpenDiscoveryDocumentWithAccessStrategy(ctx context.Context, documentURL string, accessStrategy accessstrategy.Type, tenantValue, systemTenantHeader string) (*Document, error) {
	log.C(ctx).Infof("Fetching ORD Document %q with Access Strategy %q", documentURL, accessStrategy)
	executor, err := c.accessStrategyExecutorProvider.Provide(accessStrategy)
	if err != nil {
		return nil, err
	}

	resp, err := executor.Execute(ctx, c.Client, documentURL, tenantValue, systemTenantHeader)
	if err != nil {
		return nil, err
	}

	defer closeBody(ctx, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("error while fetching open resource discovery document %q: status code %d", documentURL, resp.StatusCode)
	}

	fmt.Printf("ALEX doc %s\n", documentURL)

	resp.Body = http.MaxBytesReader(nil, resp.Body, 2097152)
	bodyBytes, err := io.ReadAll(resp.Body)

	fmt.Println(string(bodyBytes))
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

func addDocument(docs *[]*Document, doc *Document, mutex *sync.Mutex) {
	mutex.Lock()
	defer mutex.Unlock()
	*docs = append(*docs, doc)
}

func addError(fetchDocErrors *[]error, err error, mutex *sync.Mutex) {
	mutex.Lock()
	defer mutex.Unlock()
	*fetchDocErrors = append(*fetchDocErrors, err)
}

func (c *client) fetchConfig(ctx context.Context, app *model.Application, webhook *model.Webhook, tenantValue, systemTenantHeader string) (*WellKnownConfig, error) {
	var resp *http.Response
	var err error
	webhookURL := *webhook.URL
	if webhook.ProxyURL != nil {
		webhookURL = *webhook.ProxyURL
	}
	if webhook.Auth != nil && webhook.Auth.AccessStrategy != nil && len(*webhook.Auth.AccessStrategy) > 0 {
		log.C(ctx).Infof("Application %q (id = %q, type = %q) ORD webhook is configured with %q access strategy.", app.Name, app.ID, app.Type, *webhook.Auth.AccessStrategy)
		executor, err := c.accessStrategyExecutorProvider.Provide(accessstrategy.Type(*webhook.Auth.AccessStrategy))
		if err != nil {
			return nil, errors.Wrapf(err, "cannot find executor for access strategy %q as part of webhook processing", *webhook.Auth.AccessStrategy)
		}
		resp, err = executor.Execute(ctx, c.Client, webhookURL, tenantValue, systemTenantHeader)
		if err != nil {
			return nil, errors.Wrapf(err, "error while fetching open resource discovery well-known configuration with access strategy %q", *webhook.Auth.AccessStrategy)
		}
	} else if webhook.Auth != nil {
		log.C(ctx).Infof("Application %q (id = %q, type = %q) configuration endpoint is secured and webhook credentials will be used", app.Name, app.ID, app.Type)
		resp, err = httputil.GetRequestWithCredentials(ctx, c.Client, webhookURL, tenantValue, webhook.Auth)
		if err != nil {
			return nil, errors.Wrap(err, "error while fetching open resource discovery well-known configuration with webhook credentials")
		}
	} else {
		log.C(ctx).Infof("Application %q (id = %q, type = %q) configuration endpoint is not secured", app.Name, app.ID, app.Type)
		resp, err = httputil.GetRequestWithoutCredentials(c.Client, webhookURL, tenantValue)
		if err != nil {
			return nil, errors.Wrap(err, "error while fetching open resource discovery well-known configuration")
		}
	}

	defer closeBody(ctx, resp.Body)

	bodyBytes, err := io.ReadAll(resp.Body)
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

func buildDocumentURL(docURL, appBaseURL, proxyURL, proxyBaseURL string) (string, error) {
	docURLParsed, err := url.Parse(docURL)
	if err != nil {
		return "", err
	}
	if docURLParsed.IsAbs() {
		return docURL, nil
	}

	fmt.Printf("proxyURL: %s, docURL: %s, appBaseURL: %s", proxyURL, docURL, appBaseURL)

	if proxyURL != "" {
		return proxyBaseURL + docURL, nil
	}

	return appBaseURL + docURL, nil
}

// if webhookURL is not /well-known, but there is a valid baseURL provided in the config - use it
// if webhookURL is /well-known, strip the suffix and use it as baseURL. In case both are provided - the config baseURL is used.
func calculateBaseURL(webhook *model.Webhook, config WellKnownConfig) (string, error) {
	if webhook.ProxyURL != nil {
		return str.PtrStrToStr(webhook.ProxyURL), nil
	}

	if config.BaseURL != "" {
		return config.BaseURL, nil
	}

	parsedWebhookURL, err := url.ParseRequestURI(str.PtrStrToStr(webhook.URL))
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

func prepareS4SystemTenantHeader(baseURL string) string {
	if strings.Contains(baseURL, s4HanaDomain) {
		baseURLRegexp := regexp.MustCompile(`^(http)(s)?[:/]+`)
		return baseURLRegexp.ReplaceAllLiteralString(baseURL, "")
	}

	return ""
}
