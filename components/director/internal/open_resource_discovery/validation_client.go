package ord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

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
	url string
}

// NewValidationClient returns new validation client
func NewValidationClient(url string) *ValidationClient {
	return &ValidationClient{
		url: url,
	}
}

// Validate sends request to API Metadata Validator to validate one ORD document
func (vc *ValidationClient) Validate(ruleset string, requestBody string) ([]ValidationResult, error) {
	url := fmt.Sprintf("%s/api/v1/document/validate?ruleset=%s", vc.url, ruleset)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("response status code is %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []ValidationResult
	if err = json.Unmarshal(body, &results); err != nil {
		return nil, err
	}

	return results, nil
}
