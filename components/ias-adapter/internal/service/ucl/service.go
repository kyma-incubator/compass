package ucl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
)

type Service struct {
	client *http.Client
}

func NewService(client *http.Client) Service {
	return Service{
		client: client,
	}
}

type StatusReport struct {
	State         types.State `json:"state"`
	Configuration any         `json:"configuration,omitempty"`
	Error         string      `json:"error,omitempty"`
}

func (s Service) ReportStatus(ctx context.Context, url string, statusReport StatusReport) error {
	body, err := json.Marshal(statusReport)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("unexpected response status %s: failed to read response body %w", resp.Status, err)
		}
		return fmt.Errorf("unexpected response status: %s, body: %s", resp.Status, string(responseBody))
	}

	return nil
}
