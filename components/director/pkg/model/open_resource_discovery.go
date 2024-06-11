package model

// ValidationResult represents the structure of the response from the successful requests to API Metadata Validator
type ValidationResult struct {
	Code             string   `json:"code"`
	Path             []string `json:"path"`
	Message          string   `json:"message"`
	Severity         string   `json:"severity"`
	ProductStandards []string `json:"productStandards"`
}
