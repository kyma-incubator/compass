package clients

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/model"
	"github.com/kyma-incubator/compass/tests/pkg/util"
	"github.com/stretchr/testify/require"
)

type APIMetadataValidatorConfig struct {
	ConfigureEndpoint string        `envconfig:"APP_API_METADATA_VALIDATOR_CONFIGURE_ENDPOINT"`
	Timeout           time.Duration `envconfig:"APP_DESTINATIONS_TIMEOUT,default=30s"`
}

type ORDMetadataValidatorClient struct {
	httpClient *http.Client
	apiConfig  APIMetadataValidatorConfig
}

func NewAPIMetadataValidatorClient(apiConfig APIMetadataValidatorConfig) *ORDMetadataValidatorClient {
	httpClient := http.DefaultClient
	httpClient.Timeout = apiConfig.Timeout

	return &ORDMetadataValidatorClient{
		httpClient: httpClient,
		apiConfig:  apiConfig,
	}
}

func (c *ORDMetadataValidatorClient) ConfigureValidationErrors(t *testing.T, validationErrors []model.ValidationResult) {
	validationErrorBytes, err := json.Marshal(validationErrors)
	require.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, c.apiConfig.ConfigureEndpoint, bytes.NewReader(validationErrorBytes))
	require.NoError(t, err)

	request.Header.Set(util.ContentTypeHeader, "application/json;charset=UTF-8")

	resp, err := c.httpClient.Do(request)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	t.Logf("Configure validation errors response: %s", string(body))

	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func (c *ORDMetadataValidatorClient) ClearValidationErrors(t *testing.T) {
	url := c.apiConfig.ConfigureEndpoint
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	require.NoError(t, err)

	resp, err := c.httpClient.Do(request)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	t.Logf("Delete validation errors response: %s", string(body))

	require.Equal(t, http.StatusOK, resp.StatusCode)
}
