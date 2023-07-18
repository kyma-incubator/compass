package subscription

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/form3tech-oss/jwt-go"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	oauth2 "github.com/kyma-incubator/compass/components/external-services-mock/internal/oauth"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	oauth2 "github.com/kyma-incubator/compass/components/external-services-mock/internal/oauth"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

var (
	testErr                   = errors.New("test error")
	targetURL                 = "https://target-url.com"
	token                     = "token-value"
	keysPath                  = "file://testdata/jwks-private.json"
	providerSubaccID          = "c062f54a-5626-4ad1-907a-3cca6fe3b80d"
	standardFlow              = "standard"
	directDependencyFlow      = "directDependency"
	indirectDependencyFlow    = "indirectDependency"
	subscriptionFlowHeaderKey = "subscriptionFlow"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	if resp := f(req); resp == nil {
		return nil, testErr
	}
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestHandler_SubscribeAndUnsubscribe(t *testing.T) {
	// GIVEN
	privateJWKS, err := FetchJWK(context.TODO(), keysPath)
	require.NoError(t, err)
	key, ok := privateJWKS.Get(0)
	assert.True(t, ok)

	tokenWithClaim := createTokenWithSigningMethod(t, key)

	appName := "94764028-8cf8-11ec-9ffc-acde48001122"
	apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", appName)
	reqBody := "{\"subscriptionParams\": {}}"
	emptyTenantConfig := Config{}
	emptyProviderConfig := ProviderConfig{}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	tenantCfg := Config{
		TenantFetcherURL:                   "https://tenant-fetcher.com",
		RootAPI:                            "/tenants",
		RegionalHandlerEndpoint:            "/v1/regional/{region}/callback/{tenantId}",
		TenantPathParam:                    "tenantId",
		RegionPathParam:                    "region",
		SubscriptionProviderID:             "id-value!t12345",
		TenantFetcherFullRegionalURL:       "",
		TestConsumerAccountID:              "consumerAccountID",
		TestConsumerSubaccountID:           "consumberSubaccountID",
		TestConsumerTenantID:               "consumerTenantID",
		PropagatedProviderSubaccountHeader: "X-Propagated-Provider",
		SubscriptionProviderAppNameValue:   "subscriptionProviderAppNameValue",
		StandardFlow:                       standardFlow,
		DirectDependencyFlow:               directDependencyFlow,
		IndirectDependencyFlow:             indirectDependencyFlow,
		SubscriptionFlowHeaderKey:          subscriptionFlowHeaderKey,
	}

	providerCfg := ProviderConfig{
		TenantIDProperty:                                          "tenantProperty",
		SubaccountTenantIDProperty:                                "subaccountProperty",
		SubdomainProperty:                                         "subdomainProperty",
		LicenseTypeProperty:                                       "LicenseTypeProperty",
		SubscriptionProviderIDProperty:                            "subscriptionProviderProperty",
		ProviderSubaccountIDProperty:                              "providerSubaccountIDProperty",
		ConsumerTenantIDProperty:                                  "consumerTenantIdProperty",
		SubscriptionProviderAppNameProperty:                       "subscriptionProviderAppNameProperty",
		SubscriptionIDProperty:                                    "subscriptionIDProperty",
		DependentServiceInstancesInfoProperty:                     "dependentServiceInstancesInfoProperty",
		DependentServiceInstancesInfoAppIDProperty:                "dependentServiceInstancesInfoAppIDProperty",
		DependentServiceInstancesInfoAppNameProperty:              "dependentServiceInstancesInfoAppNameProperty",
		DependentServiceInstancesInfoProviderSubaccountIDProperty: "dependentServiceInstancesInfoProviderSubaccountIDProperty",
	}

	t.Run("Error when missing authorization header", func(t *testing.T) {
		//GIVEN
		subscribeReq, err := http.NewRequest(http.MethodPost, targetURL+apiPath, bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)
		h := NewHandler(httpClient, emptyTenantConfig, emptyProviderConfig, "")
		r := httptest.NewRecorder()

		//WHEN
		h.Subscribe(r, subscribeReq)
		resp := r.Result()

		//THEN
		expectedBody := "{\"error\":\"while executing subscribe request: authorization header is required\"}\n"
		assertExpectedResponse(t, resp, expectedBody, http.StatusUnauthorized)
	})

	t.Run("Error when missing Bearer token", func(t *testing.T) {
		//GIVEN
		subscribeReq, err := http.NewRequest(http.MethodPost, targetURL+apiPath, bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)
		subscribeReq.Header.Add(httphelpers.AuthorizationHeaderKey, "Bearer ")
		h := NewHandler(httpClient, emptyTenantConfig, emptyProviderConfig, "")
		r := httptest.NewRecorder()

		//WHEN
		h.Subscribe(r, subscribeReq)
		resp := r.Result()

		//THEN
		expectedBody := "{\"error\":\"while executing subscribe request: token value is required\"}\n"
		assertExpectedResponse(t, resp, expectedBody, http.StatusUnauthorized)
	})

	t.Run("Error when missing tenant path param", func(t *testing.T) {
		//GIVEN
		subReq, err := http.NewRequest(http.MethodPost, targetURL+fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", ""), bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)
		subReq.Header.Add(httphelpers.AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))
		h := NewHandler(httpClient, emptyTenantConfig, emptyProviderConfig, "")
		r := httptest.NewRecorder()

		//WHEN
		h.Subscribe(r, subReq)
		resp := r.Result()

		//THEN
		expectedBody := "{\"error\":\"while executing subscribe request: parameter [app_name] not provided\"}\n"
		assertExpectedResponse(t, resp, expectedBody, http.StatusBadRequest)
	})

	t.Run("Error when extracting tenant claim from token", func(t *testing.T) {
		//GIVEN
		subscribeReq, err := http.NewRequest(http.MethodPost, targetURL+apiPath, bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)
		subscribeReq.Header.Add(httphelpers.AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))
		subscribeReq = mux.SetURLVars(subscribeReq, map[string]string{"app_name": appName})
		h := NewHandler(httpClient, emptyTenantConfig, emptyProviderConfig, "")
		r := httptest.NewRecorder()

		//WHEN
		h.Subscribe(r, subscribeReq)
		resp := r.Result()

		//THEN
		expectedBody := "{\"error\":\"while executing subscribe request: while creating subscription request: error occurred when extracting consumer subaccount from token claims\"}\n"
		assertExpectedResponse(t, resp, expectedBody, http.StatusInternalServerError)
	})

	t.Run("Error when missing propagated provider subaccount header", func(t *testing.T) {
		//GIVEN
		//privateJWKS, err := FetchJWK(context.TODO(), keysPath)
		//require.NoError(t, err)
		//key, ok := privateJWKS.Get(0)
		//assert.True(t, ok)
		//
		//tokenWithClaim := createTokenWithSigningMethod(t, key)
		subscribeReq, err := http.NewRequest(http.MethodPost, targetURL+apiPath, bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)
		subscribeReq.Header.Add(httphelpers.AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", tokenWithClaim))
		subscribeReq = mux.SetURLVars(subscribeReq, map[string]string{"app_name": appName})
		h := NewHandler(httpClient, emptyTenantConfig, emptyProviderConfig, "")
		r := httptest.NewRecorder()

		//WHEN
		h.Subscribe(r, subscribeReq)
		resp := r.Result()

		//THEN
		expectedBody := "{\"error\":\"while executing subscribe request: while creating subscription request: An error occured when setting json value: path cannot be empty\"}\n"
		assertExpectedResponse(t, resp, expectedBody, http.StatusInternalServerError)
	})

	t.Run("Error when subscription request to tenant fetcher fails", func(t *testing.T) {
		//GIVEN
		subscribeReq, err := http.NewRequest(http.MethodPost, targetURL+apiPath, bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)
		subscribeReq.Header.Add(httphelpers.AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))
		subscribeReq.Header.Add(tenantCfg.PropagatedProviderSubaccountHeader, providerSubaccID)
		subscribeReq = mux.SetURLVars(subscribeReq, map[string]string{"app_name": appName})

		testErr = errors.New("while executing subscribe request to tenant fetcher")
		testClient := NewTestClient(func(req *http.Request) *http.Response {
			return nil
		})

		h := NewHandler(testClient, tenantCfg, providerCfg, "")
		r := httptest.NewRecorder()

		//WHEN
		h.Subscribe(r, subscribeReq)
		resp := r.Result()

		//THEN
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)
		require.Contains(t, string(body), "while executing subscribe request")
	})

	t.Run("Error when tenant fetcher returns unexpected status code on subscribe request", func(t *testing.T) {
		//GIVEN
		subscribeReq, err := http.NewRequest(http.MethodPost, targetURL+apiPath, bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)
		subscribeReq.Header.Add(httphelpers.AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", tokenWithClaim))
		subscribeReq.Header.Add(tenantCfg.PropagatedProviderSubaccountHeader, providerSubaccID)
		subscribeReq.Header.Add(subscriptionFlowHeaderKey, standardFlow)
		subscribeReq = mux.SetURLVars(subscribeReq, map[string]string{"app_name": appName})

		testClient := NewTestClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusAccepted,
			}
		})

		h := NewHandler(testClient, tenantCfg, providerCfg, "")
		r := httptest.NewRecorder()

		//WHEN
		h.Subscribe(r, subscribeReq)
		resp := r.Result()

		//THEN
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)
		require.Contains(t, string(body), "while executing subscribe request: wrong status code while executing subscription request")
	})

	t.Run("Error when unknown subscription flow", func(t *testing.T) {
		//GIVEN
		subscribeReq, err := http.NewRequest(http.MethodPost, targetURL+apiPath, bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)
		subscribeReq.Header.Add(oauth2.AuthorizationHeader, fmt.Sprintf("Bearer %s", tokenWithClaim))
		subscribeReq.Header.Add(tenantCfg.PropagatedProviderSubaccountHeader, providerSubaccID)
		subscribeReq.Header.Add(subscriptionFlowHeaderKey, "unknown")
		subscribeReq = mux.SetURLVars(subscribeReq, map[string]string{"app_name": appName})

		h := NewHandler(nil, tenantCfg, providerCfg, "")
		r := httptest.NewRecorder()

		//WHEN
		h.Subscribe(r, subscribeReq)
		resp := r.Result()

		//THEN
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)
		require.Contains(t, string(body), "Unknown subscription flow:")
	})

	t.Run("Successful API calls to tenant fetcher", func(t *testing.T) {
		subscribeReq, err := http.NewRequest(http.MethodPost, targetURL+apiPath, bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)

		unsubscribeReq, err := http.NewRequest(http.MethodDelete, targetURL+apiPath, bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)

		testCases := []struct {
			Name             string
			Request          *http.Request
			IsSubscription   bool
			SubscriptionFlow string
		}{
			{
				Name:             "Successfully executed subscribe request",
				Request:          subscribeReq,
				IsSubscription:   true,
				SubscriptionFlow: standardFlow,
			},
			{
				Name:             "Successfully executed subscribe request when adding second subscription",
				Request:          subscribeReq,
				IsSubscription:   true,
				SubscriptionFlow: standardFlow,
			},
			{
				Name:             "Successfully executed unsubscribe request when there are more than one subscriptions",
				Request:          unsubscribeReq,
				IsSubscription:   false,
				SubscriptionFlow: standardFlow,
			},
			{
				Name:             "Successfully executed unsubscribe request",
				Request:          unsubscribeReq,
				IsSubscription:   false,
				SubscriptionFlow: standardFlow,
			},
			{
				Name:             "Do not make unsubscribe request to tenant fetcher when there are not subscriptions to delete",
				Request:          unsubscribeReq,
				IsSubscription:   false,
				SubscriptionFlow: standardFlow,
			},
			{
				Name:             "Successfully executed subscribe request when indirect dependency flow",
				Request:          subscribeReq,
				IsSubscription:   true,
				SubscriptionFlow: indirectDependencyFlow,
			},
			{
				Name:             "Successfully executed unsubscribe request when indirect dependency flow",
				Request:          unsubscribeReq,
				IsSubscription:   false,
				SubscriptionFlow: indirectDependencyFlow,
			},
			{
				Name:             "Successfully executed subscribe request when direct dependency flow",
				Request:          subscribeReq,
				IsSubscription:   true,
				SubscriptionFlow: directDependencyFlow,
			},
			{
				Name:             "Successfully executed unsubscribe request when direct dependency flow",
				Request:          unsubscribeReq,
				IsSubscription:   false,
				SubscriptionFlow: directDependencyFlow,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.Name, func(t *testing.T) {
				//GIVEN
				req := testCase.Request
				req.Header.Add(httphelpers.AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", tokenWithClaim))
				req.Header.Add(tenantCfg.PropagatedProviderSubaccountHeader, providerSubaccID)
				req.Header.Add(subscriptionFlowHeaderKey, testCase.SubscriptionFlow)
				req = mux.SetURLVars(req, map[string]string{"app_name": appName})

				testClient := NewTestClient(func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusOK,
					}
				})

				h := NewHandler(testClient, tenantCfg, providerCfg, "jobID")
				r := httptest.NewRecorder()

				//WHEN
				if testCase.IsSubscription {
					h.Subscribe(r, req)
				} else {
					h.Unsubscribe(r, req)
				}
				resp := r.Result()

				//THEN
				require.Equal(t, http.StatusAccepted, resp.StatusCode)
				require.Equal(t, "/api/v1/jobs/jobID", resp.Header.Get("Location"))
				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)
				require.Empty(t, body)
			})
		}
	})

	t.Run("Error when executing unsubscribe request", func(t *testing.T) {
		//GIVEN
		subscribeReq, err := http.NewRequest(http.MethodPost, targetURL+apiPath, bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)
		h := NewHandler(httpClient, emptyTenantConfig, emptyProviderConfig, "")
		r := httptest.NewRecorder()

		//WHEN
		h.Unsubscribe(r, subscribeReq)
		resp := r.Result()

		//THEN
		expectedBody := "{\"error\":\"while executing unsubscribe request: authorization header is required\"}\n"
		assertExpectedResponse(t, resp, expectedBody, http.StatusUnauthorized)
	})
}

func TestHandler_JobStatus(t *testing.T) {
	jobID := "d1a21d4a-be03-4da5-a0ce-a006fbc851a6"
	apiPath := fmt.Sprintf("/api/v1/jobs/%s", jobID)

	testCases := []struct {
		Name                 string
		RequestMethod        string
		RequestBody          string
		ExpectedResponseCode int
		ExpectedBody         string
		AuthHeader           string
		Token                string
	}{
		{
			Name:                 "Error when missing authorization header",
			RequestMethod:        http.MethodGet,
			ExpectedBody:         "{\"error\":\"authorization header is required\"}\n",
			ExpectedResponseCode: http.StatusUnauthorized,
			AuthHeader:           "",
			Token:                "",
		},
		{
			Name:                 "Error when missing token",
			RequestMethod:        http.MethodGet,
			ExpectedBody:         "{\"error\":\"token value is required\"}\n",
			ExpectedResponseCode: http.StatusUnauthorized,
			AuthHeader:           httphelpers.AuthorizationHeaderKey,
			Token:                "",
		},
		{
			Name:                 "Error when request method is not the expected one",
			RequestMethod:        http.MethodPost,
			ExpectedResponseCode: http.StatusMethodNotAllowed,
			AuthHeader:           httphelpers.AuthorizationHeaderKey,
			Token:                token,
		},
		{
			Name:                 "Successful job status response",
			RequestMethod:        http.MethodGet,
			ExpectedResponseCode: http.StatusOK,
			ExpectedBody:         fmt.Sprintf("{\"status\":\"COMPLETED\"}"),
			AuthHeader:           httphelpers.AuthorizationHeaderKey,
			Token:                token,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			getJobReq, err := http.NewRequest(testCase.RequestMethod, targetURL+apiPath, bytes.NewBuffer([]byte(testCase.RequestBody)))
			require.NoError(t, err)
			getJobReq.Header.Add(testCase.AuthHeader, fmt.Sprintf("Bearer %s", testCase.Token))
			if testCase.AuthHeader == "" {
				getJobReq.Header.Del(httphelpers.AuthorizationHeaderKey)
			}
			h := NewHandler(nil, Config{}, ProviderConfig{}, jobID)
			r := httptest.NewRecorder()

			//WHEN
			h.JobStatus(r, getJobReq)
			resp := r.Result()

			//THEN
			if len(testCase.ExpectedBody) > 0 {
				assertExpectedResponse(t, resp, testCase.ExpectedBody, testCase.ExpectedResponseCode)
			} else {
				require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode)
			}
		})
	}
}

func assertExpectedResponse(t *testing.T, response *http.Response, expectedBody string, expectedStatusCode int) {
	require.Equal(t, expectedStatusCode, response.StatusCode)
	body, err := ioutil.ReadAll(response.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)
	require.Equal(t, expectedBody, string(body))
}

func createTokenWithSigningMethod(t *testing.T, key jwk.Key) string {
	tokenClaims := struct {
		Tenant string `json:"tenant"`
		jwt.StandardClaims
	}{
		Tenant: "test",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, tokenClaims)

	var rawKey interface{}
	err := key.Raw(&rawKey)
	require.NoError(t, err)

	signedToken, err := token.SignedString(rawKey)
	require.NoError(t, err)

	return signedToken
}

func FetchJWK(ctx context.Context, urlstring string, options ...jwk.FetchOption) (jwk.Set, error) {
	u, err := url.Parse(urlstring)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse url")
	}

	switch u.Scheme {
	case "http", "https":
		return jwk.Fetch(ctx, urlstring, options...)
	case "file":
		filePath := strings.TrimPrefix(urlstring, "file://")
		f, err := os.Open(filePath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to open jwk file")
		}
		defer func() {
			err := f.Close()
			if err != nil {
				log.C(ctx).WithError(err).Errorf("An error has occurred while closing file: %v", err)
			}
		}()

		buf, err := io.ReadAll(f)
		if err != nil {
			return nil, errors.Wrap(err, "failed read content from jwk file")
		}
		return jwk.Parse(buf)
	}
	return nil, errors.Errorf("invalid url scheme %s", u.Scheme)
}
