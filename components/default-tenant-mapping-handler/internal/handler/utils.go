package handler

import (
	"context"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// SuccessResponse structure used for JSON encoded success response
type SuccessResponse struct {
	State string `json:"state,omitempty"`
}

// ErrorResponse structure used for JSON encoded error response
type ErrorResponse struct {
	State   string `json:"state,omitempty"`
	Message string `json:"error"`
}

func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.C(ctx).Errorf("an error has occurred while closing response body: %v", err)
	}
}
