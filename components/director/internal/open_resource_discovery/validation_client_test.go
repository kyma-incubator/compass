package ord_test

import (
	"bytes"
	"encoding/json"
	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var successfulRoundTripFuncCustom = func(t *testing.T) func(req *http.Request) *http.Response {
	return func(req *http.Request) *http.Response {
		mockResponse := []ord.ValidationResult{
			{
				Code:     "1001",
				Path:     []string{"path1", "path2"},
				Message:  "Validation message",
				Severity: "error",
			},
		}

		statusCode := http.StatusOK
		data, err := json.Marshal(mockResponse)
		require.NoError(t, err)
		return &http.Response{
			StatusCode: statusCode,
			Body:       io.NopCloser(bytes.NewBuffer(data)),
		}
	}
}

var errorRoundTripFunc = func(t *testing.T) func(req *http.Request) *http.Response {
	return func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
		}
	}
}

func TestValidationClient_Validate(t *testing.T) {
	// Mock response
	mockResponse := []ord.ValidationResult{
		{
			Code:     "1001",
			Path:     []string{"path1", "path2"},
			Message:  "Validation message",
			Severity: "error",
		},
	}

	f := successfulRoundTripFuncCustom(t)
	client := NewTestClient(f)

	vc := ord.NewValidationClient("http://example.com", client)

	// Test the Validate method
	results, err := vc.Validate("ruleset", "requestBody")
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, mockResponse, results)
}

func TestValidationClient_Validate_ReturnsError(t *testing.T) {
	f := errorRoundTripFunc(t)
	client := NewTestClient(f)

	vc := ord.NewValidationClient("http://example.com", client)

	// Test the Validate method
	results, err := vc.Validate("ruleset", "requestBody")
	assert.Error(t, err)
	assert.Nil(t, results)
}
