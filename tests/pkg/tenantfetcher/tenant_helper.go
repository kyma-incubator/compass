/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tenantfetcher

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"
)

const (
	TenantPathParamValue       = "tenant"
	RegionPathParamValue       = "eu-1"
	RegionKey                  = "region"
	DefaultSubdomain           = "default-subdomain"
	DefaultSubaccountSubdomain = "default-subaccount-subdomain"
)

type Tenant struct {
	TenantID               string
	SubaccountID           string
	CustomerID             string
	Subdomain              string
	SubscriptionProviderID string
	ProviderSubaccountID   string
}

type TenantIDProperties struct {
	TenantIDProperty               string
	SubaccountTenantIDProperty     string
	CustomerIDProperty             string
	SubdomainProperty              string
	SubscriptionProviderIDProperty string
	ProviderSubaccountIdProperty   string
}

// CreateTenantRequest returns a prepared tenant request with token in the header with the necessary tenant-fetcher claims
func CreateTenantRequest(t *testing.T, tenants Tenant, tenantProperties TenantIDProperties, httpMethod, tenantFetcherUrl, externalServicesMockURL, clientID, clientSecret string) *http.Request {
	var (
		body = "{}"
		err  error
	)

	if len(tenants.TenantID) > 0 {
		body, err = sjson.Set(body, tenantProperties.TenantIDProperty, tenants.TenantID)
		require.NoError(t, err)
	}
	if len(tenants.SubaccountID) > 0 {
		body, err = sjson.Set(body, tenantProperties.SubaccountTenantIDProperty, tenants.SubaccountID)
		require.NoError(t, err)
	}
	if len(tenants.CustomerID) > 0 {
		body, err = sjson.Set(body, tenantProperties.CustomerIDProperty, tenants.CustomerID)
		require.NoError(t, err)
	}
	if len(tenants.Subdomain) > 0 {
		body, err = sjson.Set(body, tenantProperties.SubdomainProperty, tenants.Subdomain)
		require.NoError(t, err)
	}
	if len(tenants.SubscriptionProviderID) > 0 {
		body, err = sjson.Set(body, tenantProperties.SubscriptionProviderIDProperty, tenants.SubscriptionProviderID)
		require.NoError(t, err)
	}
	if len(tenants.ProviderSubaccountID) > 0 {
		body, err = sjson.Set(body, tenantProperties.ProviderSubaccountIdProperty, tenants.ProviderSubaccountID)
		require.NoError(t, err)
	}

	request, err := http.NewRequest(httpMethod, tenantFetcherUrl, bytes.NewBuffer([]byte(body)))
	require.NoError(t, err)

	tkn := token.GetClientCredentialsToken(t, context.Background(), externalServicesMockURL+"/secured/oauth/token", clientID, clientSecret, "tenantFetcherClaims")
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tkn))

	return request
}

func ActualTenantID(tenantIDs Tenant) string {
	if len(tenantIDs.SubaccountID) > 0 {
		return tenantIDs.SubaccountID
	}

	return tenantIDs.TenantID
}

func BuildTenantFetcherRegionalURL(regionalHandlerEndpoint, tenantPathParam, regionPathParam, tenantFetcherURL, rootAPI string) string {
	regionalEndpoint := strings.Replace(regionalHandlerEndpoint, fmt.Sprintf("{%s}", tenantPathParam), TenantPathParamValue, 1)
	regionalEndpoint = strings.Replace(regionalEndpoint, fmt.Sprintf("{%s}", regionPathParam), RegionPathParamValue, 1)
	tenantFetcherFullRegionalURL := tenantFetcherURL + rootAPI + regionalEndpoint
	return tenantFetcherFullRegionalURL
}

func DefaultClaims(externalServicesMockURL string) map[string]interface{} {
	claims := map[string]interface{}{
		"test": "tenant-fetcher",
		"scope": []string{
			"prefix.Callback",
		},
		"tenant":   "tenant",
		"identity": "tenant-fetcher-tests",
		"iss":      externalServicesMockURL,
		"exp":      time.Now().Unix() + int64(time.Minute.Seconds()),
	}
	return claims
}
