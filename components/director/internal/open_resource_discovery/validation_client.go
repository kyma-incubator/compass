package ord

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
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
	Validate(ruleset, requestBody string) ([]ValidationResult, error)
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
	url    string
	client *http.Client
}

// NewValidationClient returns new validation client
func NewValidationClient(url string, client *http.Client) *ValidationClient {
	return &ValidationClient{
		url:    url,
		client: client,
	}
}

// Validate sends request to API Metadata Validator to validate one ORD document
func (vc *ValidationClient) Validate(ruleset string, requestBody string) ([]ValidationResult, error) {
	req, err := vc.createRequest(ruleset, requestBody)
	if err != nil {
		return nil, err
	}

	resp, err := vc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected response status code: %d. expected: %d", resp.StatusCode, http.StatusOK)
	}

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
