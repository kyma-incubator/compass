package cis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// CISResponse response from the context enrichment api
type CISResponse struct {
	GlobalAccountID string `json:"globalAccountId"`
	SubaccountID    string `json:"subaccountId"`
	TenantID        string `json:"tenantId"`
	Subdomain       string `json:"subdomain"`
	Origin          string `json:"origin"`
}

type service struct {
	httpClient           http.Client
	contextEnrichmentURL string
	token                string
}

// NewCisService missing godoc
func NewCisService(client http.Client, contextEnrichmentURL string, token string) *service {
	return &service{
		httpClient:           client,
		contextEnrichmentURL: contextEnrichmentURL,
		token:                token,
	}
}

// GetGlobalAccount missing godoc
func (s *service) GetGlobalAccount(ctx context.Context, subaccountID string) (string, error) {
	endpoint := fmt.Sprintf("%s/%s", s.contextEnrichmentURL, subaccountID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSuffix(s.token, "\n"))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "while getting details for tenant")
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.C(ctx).Error(err, "Failed to close HTTP response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("Received status code %d from CIS", resp.StatusCode)
	}

	var response CISResponse
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "while reading response body")
	}
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to unmarshall HTTP response with body %s", string(bodyBytes)))
	}

	return response.GlobalAccountID, nil
}
