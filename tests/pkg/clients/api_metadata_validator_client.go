package clients

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/model"
	"github.com/kyma-incubator/compass/tests/pkg/util"
	"github.com/stretchr/testify/require"
)

// APIMetadataValidatorConfig holds the configuration
type APIMetadataValidatorConfig struct {
	ConfigureEndpoint string        `envconfig:"APP_API_METADATA_VALIDATOR_CONFIGURE_ENDPOINT"`
	Timeout           time.Duration `envconfig:"APP_DESTINATIONS_TIMEOUT,default=30s"`
}

type APIMetadataValidatorClient struct {
	httpClient *http.Client
	apiConfig  APIMetadataValidatorConfig
}

// NewAPIMetadataValidatorClient creates a new APIMetadataValidatorClient
func NewAPIMetadataValidatorClient(apiConfig APIMetadataValidatorConfig) *APIMetadataValidatorClient {
	httpClient := http.DefaultClient
	httpClient.Timeout = apiConfig.Timeout

	return &APIMetadataValidatorClient{
		httpClient: httpClient,
		apiConfig:  apiConfig,
	}
}

// ConfigureValidationErrors makes a request to external services mock in order to mock the list of returned ValidationResult.
func (c *APIMetadataValidatorClient) ConfigureValidationErrors(t *testing.T, validationErrors []model.ValidationResult) {
	validationErrorBytes, err := json.Marshal(validationErrors)
	require.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, c.apiConfig.ConfigureEndpoint, bytes.NewReader(validationErrorBytes))
	require.NoError(t, err)

	request.Header.Set(util.ContentTypeHeader, "application/json;charset=UTF-8")

	resp, err := c.httpClient.Do(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	t.Logf("Configured the mocked server to return %d validation errors", len(validationErrors))
}

// ClearValidationErrors makes a request to external services mock to clean up the mocked ValidationResult.
func (c *APIMetadataValidatorClient) ClearValidationErrors(t *testing.T) {
	url := c.apiConfig.ConfigureEndpoint
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	require.NoError(t, err)

	resp, err := c.httpClient.Do(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	t.Logf("Deleted the mocked validation errors")
}
