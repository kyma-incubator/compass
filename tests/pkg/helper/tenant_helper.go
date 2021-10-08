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

package helper

import (
    "bytes"
    "context"
    "fmt"
    directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
    "github.com/kyma-incubator/compass/tests/pkg/fixtures"
    "github.com/kyma-incubator/compass/tests/pkg/testctx"
    "github.com/kyma-incubator/compass/tests/pkg/token"
    "github.com/machinebox/graphql"
    "github.com/stretchr/testify/require"
    "github.com/tidwall/sjson"
    "net/http"
    "strings"
    "testing"
    "time"
)

const (
    TenantPathParamValue       = "tenant"
    RegionPathParamValue       = "eu-1"
    RegionKey				   = "region"
    DefaultSubdomain           = "default-subdomain"
    DefaultSubaccountSubdomain = "default-subaccount-subdomain"
)

type TenantIDs struct {
    TenantID               string
    SubaccountID           string
    CustomerID             string
    Subdomain              string
    SubscriptionProviderID string
}

type TenantIDProperties struct {
    TenantIDProperty               string
    SubaccountTenantIDProperty     string
    CustomerIDProperty             string
    SubdomainProperty              string
    SubscriptionProviderIDProperty string
}

// CreateTenantRequest returns a prepared tenant request with token in the header with the necessary tenant-fetcher claims
func CreateTenantRequest(t *testing.T, tenantIDs TenantIDs, tenantProperties TenantIDProperties, httpMethod, tenantFetcherUrl, externalServicesMockURL string) *http.Request {
    var (
        body = "{}"
        err  error
    )

    if len(tenantIDs.TenantID) > 0 {
        body, err = sjson.Set(body, tenantProperties.TenantIDProperty, tenantIDs.TenantID)
        require.NoError(t, err)
    }
    if len(tenantIDs.SubaccountID) > 0 {
        body, err = sjson.Set(body, tenantProperties.SubaccountTenantIDProperty, tenantIDs.SubaccountID)
        require.NoError(t, err)
    }
    if len(tenantIDs.CustomerID) > 0 {
        body, err = sjson.Set(body, tenantProperties.CustomerIDProperty, tenantIDs.CustomerID)
        require.NoError(t, err)
    }
    if len(tenantIDs.Subdomain) > 0 {
        body, err = sjson.Set(body, tenantProperties.SubdomainProperty, tenantIDs.Subdomain)
        require.NoError(t, err)
    }
    if len(tenantIDs.SubscriptionProviderID) > 0 {
        body, err = sjson.Set(body, tenantProperties.SubscriptionProviderIDProperty, tenantIDs.SubscriptionProviderID)
        require.NoError(t, err)
    }

    request, err := http.NewRequest(httpMethod, tenantFetcherUrl, bytes.NewBuffer([]byte(body)))
    require.NoError(t, err)
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
    request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.FetchTokenFromExternalServicesMock(t, externalServicesMockURL, claims)))

    return request
}

func ActualTenantID(tenantIDs TenantIDs) string {
    if len(tenantIDs.SubaccountID) > 0 {
        return tenantIDs.SubaccountID
    }

    return tenantIDs.TenantID
}

func AssertRuntimeSubscription(t *testing.T, ctx context.Context, runtimeID string, providedTenantIDs TenantIDs, dexGraphQLClient *graphql.Client, consumerSubaccountIDsLabelKey string) {
    actualTenant, err := fixtures.GetTenantByExternalID(dexGraphQLClient, providedTenantIDs.SubaccountID)
    require.NoError(t, err)
    assertTenant(t, actualTenant, providedTenantIDs.SubaccountID, providedTenantIDs.Subdomain)
    require.Equal(t, RegionPathParamValue, actualTenant.Labels[RegionKey])

    subscribedRuntime := directorSchema.RuntimeExt{}
    getSubscribedReq := fixtures.FixGetRuntimeRequest(runtimeID)
    err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, getSubscribedReq, &subscribedRuntime)
    require.NoError(t, err)
    consumerSubaccountIDsLabel, ok := subscribedRuntime.Labels[consumerSubaccountIDsLabelKey].([]interface{})
    require.Equal(t, true, ok)
    require.Len(t, consumerSubaccountIDsLabel, 1)
    labelValue, ok := consumerSubaccountIDsLabel[0].(string)
    require.Equal(t, true, ok)
    require.Equal(t, providedTenantIDs.SubaccountID, labelValue)
}

func assertTenant(t *testing.T, tenant *directorSchema.Tenant, tenantID, subdomain string) {
    require.Equal(t, tenantID, tenant.ID)
    if len(subdomain) > 0 {
        require.Equal(t, subdomain, tenant.Labels["subdomain"])
    }
}

func BuildTenantFetcherRegionalURL(regionalHandlerEndpoint, tenantPathParam, regionPathParam, tenantFetcherURL, rootAPI string) string {
    regionalEndpoint := strings.Replace(regionalHandlerEndpoint, fmt.Sprintf("{%s}", tenantPathParam), TenantPathParamValue, 1)
    regionalEndpoint = strings.Replace(regionalEndpoint, fmt.Sprintf("{%s}", regionPathParam), RegionPathParamValue, 1)
    tenantFetcherFullRegionalURL := tenantFetcherURL + rootAPI + regionalEndpoint
    return tenantFetcherFullRegionalURL
}