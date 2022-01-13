package cis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
	httpClient http.Client
	k8sClient  KubeClient
}

// NewCisService missing godoc
func NewCisService(client http.Client, kubeClient KubeClient) *service {
	return &service{
		httpClient: client,
		k8sClient:  kubeClient,
	}
}

// GetGlobalAccount missing godoc
func (s *service) GetGlobalAccount(ctx context.Context, region string, subaccountID string) (string, error) {
	url := s.k8sClient.GetRegionURL(region)
	endpoint := fmt.Sprintf("%s/%s", url, subaccountID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}

	token := s.k8sClient.GetRegionToken(region)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "while getting details for tenant")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.C(ctx).Error(err, "Failed to close HTTP response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			newToken, err := FetchToken(ctx, s.k8sClient.GetClientIDForRegion(region), s.k8sClient.GetClientSecretForRegion(region), s.k8sClient.GetTokenURLForRegion(region))
			if err != nil {
				log.C(ctx).Error(err)
			}
			s.k8sClient.SetRegionToken(region, newToken)
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := s.httpClient.Do(req)
			if err != nil {
				return "", errors.Wrap(err, "while getting details for tenant")
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					log.C(ctx).Error(err, "Failed to close HTTP response body")
				}
			}()
			if resp.StatusCode != http.StatusOK {
				return "", errors.Errorf("Received status code %d from CIS after retrying with new token", resp.StatusCode)
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
		} else {
			return "", errors.Errorf("Received status code %d from CIS", resp.StatusCode)
		}
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
