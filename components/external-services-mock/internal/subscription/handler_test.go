package subscription

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	oauth2 "github.com/kyma-incubator/compass/components/external-services-mock/internal/oauth"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

var (
	testErr          = errors.New("test error")
	url              = "https://target-url.com"
	token            = "token-value"
	tokenWithClaim   = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0ZW5hbnQiOiJ0ZXN0In0.5Jg0ylN1CI1vH-tmHbqCoGvOj6j-j8iFg-fZlz1BdFc"
	providerSubaccID = "c062f54a-5626-4ad1-907a-3cca6fe3b80d"
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
		subscribeReq, err := http.NewRequest(http.MethodPost, url+apiPath, bytes.NewBuffer([]byte(reqBody)))
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
		subscribeReq, err := http.NewRequest(http.MethodPost, url+apiPath, bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)
		subscribeReq.Header.Add(oauth2.AuthorizationHeader, "Bearer ")
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
		subReq, err := http.NewRequest(http.MethodPost, url+fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", ""), bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)
		subReq.Header.Add(oauth2.AuthorizationHeader, fmt.Sprintf("Bearer %s", token))
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
		subscribeReq, err := http.NewRequest(http.MethodPost, url+apiPath, bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)
		subscribeReq.Header.Add(oauth2.AuthorizationHeader, fmt.Sprintf("Bearer %s", token))
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
		subscribeReq, err := http.NewRequest(http.MethodPost, url+apiPath, bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)
		subscribeReq.Header.Add(oauth2.AuthorizationHeader, fmt.Sprintf("Bearer %s", tokenWithClaim))
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
		subscribeReq, err := http.NewRequest(http.MethodPost, url+apiPath, bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)
		subscribeReq.Header.Add(oauth2.AuthorizationHeader, fmt.Sprintf("Bearer %s", token))
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
		subscribeReq, err := http.NewRequest(http.MethodPost, url+apiPath, bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)
		subscribeReq.Header.Add(oauth2.AuthorizationHeader, fmt.Sprintf("Bearer %s", tokenWithClaim))
		subscribeReq.Header.Add(tenantCfg.PropagatedProviderSubaccountHeader, providerSubaccID)
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

	t.Run("Successful API calls to tenant fetcher", func(t *testing.T) {
		subscribeReq, err := http.NewRequest(http.MethodPost, url+apiPath, bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)

		unsubscribeReq, err := http.NewRequest(http.MethodDelete, url+apiPath, bytes.NewBuffer([]byte(reqBody)))
		require.NoError(t, err)

		testCases := []struct {
			Name           string
			Request        *http.Request
			IsSubscription bool
			isIndirectDep  bool
		}{
			{
				Name:           "Successfully executed subscribe request",
				Request:        subscribeReq,
				IsSubscription: true,
			},
			{
				Name:           "Successfully executed subscribe request when adding second subscription",
				Request:        subscribeReq,
				IsSubscription: true,
			},
			{
				Name:           "Successfully executed unsubscribe request when there are more than one subscriptions",
				Request:        unsubscribeReq,
				IsSubscription: false,
			},
			{
				Name:           "Successfully executed unsubscribe request",
				Request:        unsubscribeReq,
				IsSubscription: false,
			},
			{
				Name:           "Do not make unsubscribe request to tenant fetcher when there are not subscriptions to delete",
				Request:        unsubscribeReq,
				IsSubscription: false,
			},
			{
				Name:           "Successfully executed subscribe request when CMP is indirect dependency",
				Request:        subscribeReq,
				IsSubscription: true,
				isIndirectDep:  true,
			},
			{
				Name:           "Successfully executed unsubscribe request when CMP is indirect dependency",
				Request:        unsubscribeReq,
				IsSubscription: false,
				isIndirectDep:  true,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.Name, func(t *testing.T) {
				//GIVEN
				req := testCase.Request
				req.Header.Add(oauth2.AuthorizationHeader, fmt.Sprintf("Bearer %s", tokenWithClaim))
				req.Header.Add(tenantCfg.PropagatedProviderSubaccountHeader, providerSubaccID)
				if testCase.isIndirectDep {
					req.Header.Add("isIndirectDependency", "true")
				}
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
				require.NotEmpty(t, body)
			})
		}
	})

	t.Run("Error when executing unsubscribe request", func(t *testing.T) {
		//GIVEN
		subscribeReq, err := http.NewRequest(http.MethodPost, url+apiPath, bytes.NewBuffer([]byte(reqBody)))
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
			AuthHeader:           oauth2.AuthorizationHeader,
			Token:                "",
		},
		{
			Name:                 "Error when request method is not the expected one",
			RequestMethod:        http.MethodPost,
			ExpectedResponseCode: http.StatusMethodNotAllowed,
			AuthHeader:           oauth2.AuthorizationHeader,
			Token:                token,
		},
		{
			Name:                 "Successful job status response",
			RequestMethod:        http.MethodGet,
			ExpectedResponseCode: http.StatusOK,
			ExpectedBody:         fmt.Sprintf("{\"status\":\"COMPLETED\"}"),
			AuthHeader:           oauth2.AuthorizationHeader,
			Token:                token,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			getJobReq, err := http.NewRequest(testCase.RequestMethod, url+apiPath, bytes.NewBuffer([]byte(testCase.RequestBody)))
			require.NoError(t, err)
			getJobReq.Header.Add(testCase.AuthHeader, fmt.Sprintf("Bearer %s", testCase.Token))
			if testCase.AuthHeader == "" {
				getJobReq.Header.Del(oauth2.AuthorizationHeader)
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
