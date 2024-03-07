package validator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ValidationResult struct {
	Code             string   `json:"code"`
	Path             []string `json:"path"`
	Message          string   `json:"message"`
	Severity         string   `json:"severity"`
	Range            Range    `json:"range"`
	ProductStandards []string `json:"productStandards"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type ValidationClient struct {
	url string
}

func NewValidationClient(url string) *ValidationClient {
	return &ValidationClient{
		url: url,
	}
}

func (vc *ValidationClient) Validate(ruleset string, requestBody string) ([]ValidationResult, error) {
	url := fmt.Sprintf("%s/api/v1/document/validate?ruleset=%s", vc.url, ruleset)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
