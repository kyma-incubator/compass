package ord

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/pkg/errors"
)

const (
	rulesetQueryParam = "ruleset"
	validateEndpoint  = "/api/v1/document/validate"

	contentTypeHeaderKey       = "Content-Type"
	contentTypeApplicationJSON = "application/json;charset=UTF-8"
)

// ValidatorClient validates list of ORD documents with API Metadata Validator
//
//go:generate mockery --name=ValidatorClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type ValidatorClient interface {
	Validate(ctx context.Context, ruleset, requestBody string) ([]ValidationResult, error)
}

// ValidationResult represents the structure of the response from the successful requests to API Metadata Validator
type ValidationResult struct {
	Code             string   `json:"code"`
	Path             []string `json:"path"`
	Message          string   `json:"message"`
	Severity         string   `json:"severity"`
	ProductStandards []string `json:"productStandards"`
}

// ValidationClient represents the client for the API Metadata Validator
type ValidationClient struct {
	url     string
	client  *http.Client
	enabled bool
}

// NewValidationClient returns new validation client
func NewValidationClient(url string, client *http.Client, enabled bool) *ValidationClient {
	return &ValidationClient{
		url:     url,
		client:  client,
		enabled: enabled,
	}
}

// Validate sends request to API Metadata Validator to validate one ORD document
func (vc *ValidationClient) Validate(ctx context.Context, ruleset string, requestBody string) ([]ValidationResult, error) {
	log.C(ctx).Infof("Creating request to API Metadata Validator with base url %s and validation endpoint %s", vc.url, validateEndpoint)
	req, err := vc.createRequest(ruleset, requestBody)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Sending request to API Metadata Validator")
	resp, err := vc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected response status code: %d. expected: %d", resp.StatusCode, http.StatusOK)
	}

	log.C(ctx).Infof("Successful request to API Metadata Validator")

	var results []ValidationResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, errors.Wrap(err, "failed to decode response body")
	}

	return results, nil
}

func (vc *ValidationClient) createRequest(ruleset, requestBody string) (*http.Request, error) {
	parsedBaseURL, err := url.Parse(vc.url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse endpoint URL")
	}

	apiURL, err := url.Parse(validateEndpoint)
	if err != nil {
		return nil, err
	}

	requestURL := parsedBaseURL.ResolveReference(apiURL).String()

	req, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create HTTP request")
	}

	if ruleset != "" {
		queryParams := url.Values{}
		queryParams.Set(rulesetQueryParam, ruleset)
		req.URL.RawQuery = queryParams.Encode()
	}

	req.Header.Set(contentTypeHeaderKey, contentTypeApplicationJSON)

	return req, nil
}
