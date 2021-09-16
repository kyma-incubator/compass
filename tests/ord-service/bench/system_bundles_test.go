package bench

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/kyma-incubator/compass/tests/pkg/request"
	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/stretchr/testify/require"
)

const (
	tenantHeader = "Tenant"
)

func BenchmarkSystemBundles(b *testing.B) {
	//GIVEN
	ctx := context.Background()
	defaultTestTenant := tenant.TestTenants.GetDefaultTenantID()

	appsCount := 5
	apps := make([]graphql.ApplicationRegisterInput, 0, appsCount)
	for i := 0; i < appsCount; i++ {
		apps = append(apps, fixtures.CreateApp(fmt.Sprintf("%d", i)))
	}

	for _, app := range apps {
		appResp, err := fixtures.RegisterApplicationFromInput(b, ctx, dexGraphQLClient, defaultTestTenant, app)
		require.NoError(b, err)
		defer fixtures.UnregisterApplication(b, ctx, dexGraphQLClient, defaultTestTenant, appResp.ID)
	}

	b.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(b, ctx, dexGraphQLClient, "", "test-int-system")
	defer fixtures.CleanupIntegrationSystem(b, ctx, dexGraphQLClient, "", intSys)
	require.NoError(b, err)
	require.NotEmpty(b, intSys.ID)

	intSystemCredentials := fixtures.RequestClientCredentialsForIntegrationSystem(b, ctx, dexGraphQLClient, "", intSys.ID)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(b, ctx, dexGraphQLClient, intSystemCredentials.ID)

	unsecuredHttpClient := http.DefaultClient
	unsecuredHttpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	intSystemHttpClient := integrationSystemClient(b, ctx, unsecuredHttpClient, intSystemCredentials)

	b.ResetTimer() // Reset timer after the initialization

	for i := 0; i < b.N; i++ {
		respBody := makeRequestWithHeaders(b, intSystemHttpClient, fmt.Sprintf("%s/systemInstances?$expand=consumptionBundles($expand=apis,events)&$format=json", testConfig.ORDServiceURL), map[string][]string{tenantHeader: {defaultTestTenant}})

		b.Log(respBody)
		//THEN
		require.Len(b, len((gjson.Get(respBody, "value")).Array()), appsCount)
	}

	b.StopTimer() // Stop timer in order to exclude defers from the time
}

func makeRequestWithHeaders(b *testing.B, httpClient *http.Client, url string, headers map[string][]string) string {
	return request.MakeRequestWithHeadersAndStatusExpect(b, httpClient, url, headers, http.StatusOK, testConfig.ORDServiceDefaultResponseType)
}

func integrationSystemClient(b *testing.B, ctx context.Context, base *http.Client, intSystemCredentials *directorSchema.IntSysSystemAuth) *http.Client {
	oauthCredentialData, ok := intSystemCredentials.Auth.Credential.(*directorSchema.OAuthCredentialData)
	require.True(b, ok)

	conf := &clientcredentials.Config{
		ClientID:     oauthCredentialData.ClientID,
		ClientSecret: oauthCredentialData.ClientSecret,
		TokenURL:     oauthCredentialData.URL,
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, base)
	httpClient := conf.Client(ctx)
	httpClient.Timeout = 10 * time.Second

	return httpClient
}
