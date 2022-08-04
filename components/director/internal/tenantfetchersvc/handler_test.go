package tenantfetchersvc_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type regionalTenantCreationRequest struct {
	SubaccountID                string `json:"subaccountTenantId"`
	TenantID                    string `json:"tenantId"`
	Subdomain                   string `json:"subdomain"`
	SubscriptionProviderID      string `json:"subscriptionProviderId"`
	ProviderSubaccountID        string `json:"providerSubaccountId"`
	ConsumerTenantID            string `json:"consumerTenantID"`
	SubscriptionProviderAppName string `json:"subscriptionProviderAppName"`
}

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

func TestService_SubscriptionFlows(t *testing.T) {
	// GIVEN
	region := "eu-1"

	target := "http://example.com/foo/:region"
	txtest.CtxWithDBMatcher()

	validRequestBody, err := json.Marshal(regionalTenantCreationRequest{
		SubaccountID:                subaccountTenantExtID,
		TenantID:                    tenantExtID,
		Subdomain:                   regionalTenantSubdomain,
		SubscriptionProviderID:      subscriptionProviderID,
		ProviderSubaccountID:        providerSubaccountID,
		ConsumerTenantID:            consumerTenantID,
		SubscriptionProviderAppName: subscriptionProviderAppName,
	})
	assert.NoError(t, err)

	bodyWithMissingParent, err := json.Marshal(regionalTenantCreationRequest{
		SubaccountID:                subaccountTenantExtID,
		Subdomain:                   regionalTenantSubdomain,
		SubscriptionProviderID:      subscriptionProviderID,
		ProviderSubaccountID:        providerSubaccountID,
		ConsumerTenantID:            consumerTenantID,
		SubscriptionProviderAppName: subscriptionProviderAppName,
	})
	assert.NoError(t, err)

	bodyWithMissingTenantSubdomain, err := json.Marshal(regionalTenantCreationRequest{
		SubaccountID:                subaccountTenantExtID,
		TenantID:                    tenantExtID,
		SubscriptionProviderID:      subscriptionProviderID,
		ProviderSubaccountID:        providerSubaccountID,
		ConsumerTenantID:            consumerTenantID,
		SubscriptionProviderAppName: subscriptionProviderAppName,
	})
	assert.NoError(t, err)

	bodyWithMissingSubscriptionConsumerID, err := json.Marshal(regionalTenantCreationRequest{
		SubaccountID:                subaccountTenantExtID,
		TenantID:                    tenantExtID,
		Subdomain:                   regionalTenantSubdomain,
		ProviderSubaccountID:        providerSubaccountID,
		ConsumerTenantID:            consumerTenantID,
		SubscriptionProviderAppName: subscriptionProviderAppName,
	})
	assert.NoError(t, err)

	bodyWithMatchingAccountAndSubaccountIDs, err := json.Marshal(regionalTenantCreationRequest{
		TenantID:                    tenantExtID,
		SubaccountID:                tenantExtID,
		Subdomain:                   regionalTenantSubdomain,
		SubscriptionProviderID:      subscriptionProviderID,
		ProviderSubaccountID:        providerSubaccountID,
		ConsumerTenantID:            consumerTenantID,
		SubscriptionProviderAppName: subscriptionProviderAppName,
	})
	assert.NoError(t, err)

	bodyWithMissingProviderSubaccountID, err := json.Marshal(regionalTenantCreationRequest{
		SubaccountID:                subaccountTenantExtID,
		TenantID:                    tenantExtID,
		Subdomain:                   regionalTenantSubdomain,
		SubscriptionProviderID:      subscriptionProviderID,
		ConsumerTenantID:            consumerTenantID,
		SubscriptionProviderAppName: subscriptionProviderAppName,
	})
	assert.NoError(t, err)

	bodyWithMissingConsumerTenantID, err := json.Marshal(regionalTenantCreationRequest{
		SubaccountID:                subaccountTenantExtID,
		TenantID:                    tenantExtID,
		Subdomain:                   regionalTenantSubdomain,
		SubscriptionProviderID:      subscriptionProviderID,
		ProviderSubaccountID:        providerSubaccountID,
		SubscriptionProviderAppName: subscriptionProviderAppName,
	})
	assert.NoError(t, err)

	bodyWithMissingSubscriptionProviderAppName, err := json.Marshal(regionalTenantCreationRequest{
		SubaccountID:           subaccountTenantExtID,
		TenantID:               tenantExtID,
		Subdomain:              regionalTenantSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
		ProviderSubaccountID:   providerSubaccountID,
		ConsumerTenantID:       consumerTenantID,
	})
	assert.NoError(t, err)

	validHandlerConfig := tenantfetchersvc.HandlerConfig{
		RegionPathParam: "region",
		TenantProviderConfig: tenantfetchersvc.TenantProviderConfig{
			TenantProvider:                      testProviderName,
			TenantIDProperty:                    tenantProviderTenantIDProperty,
			SubaccountTenantIDProperty:          tenantProviderSubaccountTenantIDProperty,
			CustomerIDProperty:                  tenantProviderCustomerIDProperty,
			SubdomainProperty:                   tenantProviderSubdomainProperty,
			SubscriptionProviderIDProperty:      subscriptionProviderIDProperty,
			ProviderSubaccountIDProperty:        providerSubaccountIDProperty,
			ConsumerTenantIDProperty:            consumerTenantIDProperty,
			SubscriptionProviderAppNameProperty: subscriptionProviderAppNameProperty,
		},
	}
	regionalTenant := tenantfetchersvc.TenantSubscriptionRequest{
		SubaccountTenantID:          subaccountTenantExtID,
		AccountTenantID:             tenantExtID,
		Subdomain:                   regionalTenantSubdomain,
		Region:                      region,
		SubscriptionProviderID:      subscriptionProviderID,
		ProviderSubaccountID:        providerSubaccountID,
		ConsumerTenantID:            consumerTenantID,
		SubscriptionProviderAppName: subscriptionProviderAppName,
	}
	regionalTenantWithMatchingParentID := tenantfetchersvc.TenantSubscriptionRequest{
		SubaccountTenantID:          "",
		AccountTenantID:             tenantExtID,
		Subdomain:                   regionalTenantSubdomain,
		Region:                      region,
		SubscriptionProviderID:      subscriptionProviderID,
		ProviderSubaccountID:        providerSubaccountID,
		ConsumerTenantID:            consumerTenantID,
		SubscriptionProviderAppName: subscriptionProviderAppName,
	}

	// Subscribe flow
	testCases := []struct {
		Name                  string
		TenantSubscriberFn    func() *automock.TenantSubscriber
		Request               *http.Request
		Region                string
		ExpectedErrorOutput   string
		ExpectedSuccessOutput string
		ExpectedStatusCode    int
	}{
		{
			Name: "Succeeds",
			TenantSubscriberFn: func() *automock.TenantSubscriber {
				subscriber := &automock.TenantSubscriber{}
				subscriber.On("Subscribe", mock.Anything, &regionalTenant).Return(nil).Once()
				return subscriber
			},
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:                region,
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name: "Succeeds to create account tenant when account ID and subaccount IDs are matching",
			TenantSubscriberFn: func() *automock.TenantSubscriber {
				subscriber := &automock.TenantSubscriber{}
				subscriber.On("Subscribe", mock.Anything, &regionalTenantWithMatchingParentID).Return(nil).Once()
				return subscriber
			},
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMatchingAccountAndSubaccountIDs)),
			Region:                region,
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name:                "Returns error when region path parameter is missing",
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: "Region path parameter is missing from request",
		},
		{
			Name:                "Returns error when parent tenant is not found in body",
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingParent)),
			Region:              region,
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", tenantProviderTenantIDProperty),
		},
		{
			Name:                "Returns error when SubscriptionProviderID is not found in body",
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingSubscriptionConsumerID)),
			Region:              region,
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", subscriptionProviderIDProperty),
		},
		{
			Name:                "Returns error when providerSubaccountIDProperty is not found in body",
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingProviderSubaccountID)),
			Region:              region,
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", providerSubaccountIDProperty),
		},
		{
			Name:                "Returns error when consumerTenantIDProperty is not found in body",
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingConsumerTenantID)),
			Region:              region,
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", consumerTenantIDProperty),
		},
		{
			Name:                "Returns error when subscriptionProviderAppNameProperty is not found in body",
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingSubscriptionProviderAppName)),
			Region:              region,
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", subscriptionProviderAppNameProperty),
		},
		{
			Name:                "Returns error when request body doesn't contain tenant subdomain",
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingTenantSubdomain)),
			Region:              region,
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", tenantProviderSubdomainProperty),
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			Name:                "Returns error when reading request body fails",
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			Request:             httptest.NewRequest(http.MethodPut, target, errReader(0)),
			Region:              region,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name: "Returns error when tenant subscription fails",
			TenantSubscriberFn: func() *automock.TenantSubscriber {
				subscriber := &automock.TenantSubscriber{}
				subscriber.On("Subscribe", mock.Anything, &regionalTenant).Return(testError).Once()
				return subscriber
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			subscriber := testCase.TenantSubscriberFn()
			defer mock.AssertExpectationsForObjects(t, subscriber)

			handler := tenantfetchersvc.NewTenantsHTTPHandler(subscriber, validHandlerConfig)
			req := testCase.Request

			if len(testCase.Region) > 0 {
				vars := map[string]string{
					"region": testCase.Region,
				}
				req = mux.SetURLVars(req, vars)
			}

			w := httptest.NewRecorder()

			// WHEN
			handler.SubscribeTenant(w, req)

			// THEN
			resp := w.Result()
			body, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err)

			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Contains(t, string(body), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}

			if testCase.ExpectedSuccessOutput != "" {
				assert.Equal(t, testCase.ExpectedSuccessOutput, string(body))
			}

			assert.Equal(t, testCase.ExpectedStatusCode, resp.StatusCode)
		})
	}

	// Unsubscribe flow
	t.Run("Unsubscribe", func(t *testing.T) {
		subscriber := &automock.TenantSubscriber{}
		subscriber.On("Unsubscribe", mock.Anything, &regionalTenant).Return(testError).Once()
		defer mock.AssertExpectationsForObjects(t, subscriber)

		handler := tenantfetchersvc.NewTenantsHTTPHandler(subscriber, validHandlerConfig)
		req := httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody))

		vars := map[string]string{
			validHandlerConfig.RegionPathParam: region,
		}
		req = mux.SetURLVars(req, vars)

		w := httptest.NewRecorder()

		// WHEN
		handler.UnSubscribeTenant(w, req)

		// THEN
		resp := w.Result()
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)

		assert.Contains(t, string(body), tenantfetchersvc.InternalServerError)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestService_Dependencies(t *testing.T) {
	const (
		regionPathVar  = "region"
		missingRegion  = "eu-2"
		existingRegion = "eu-1"
		xsappname      = "xsappname"
	)
	target := fmt.Sprintf("/v1/regional/:%s/dependencies", regionPathVar)

	subscriberSvc := &automock.TenantSubscriber{}

	validHandlerConfig := tenantfetchersvc.HandlerConfig{
		RegionPathParam: "region",
		RegionToDependenciesConfig: map[string][]tenantfetchersvc.Dependency{
			existingRegion: []tenantfetchersvc.Dependency{
				tenantfetchersvc.Dependency{Xsappname: xsappname},
			},
		},
	}

	validResponse := fmt.Sprintf("[{\"xsappname\":\"%s\"}]", xsappname)

	testCases := []struct {
		Name                  string
		Request               *http.Request
		PathParams            map[string]string
		ExpectedErrorOutput   string
		ExpectedStatusCode    int
		ExpectedSuccessOutput string
	}{
		{
			Name:                "Failure when region path param is missing",
			Request:             httptest.NewRequest(http.MethodGet, target, nil),
			PathParams:          map[string]string{},
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: "Region path parameter is missing from request",
		},
		{
			Name:    "Failure when region is invalid",
			Request: httptest.NewRequest(http.MethodGet, target, nil),
			PathParams: map[string]string{
				regionPathVar: missingRegion,
			},
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: fmt.Sprintf("Invalid region provided: %s", missingRegion),
		},
		{
			Name:    "Success when existing region is provided",
			Request: httptest.NewRequest(http.MethodGet, target, nil),
			PathParams: map[string]string{
				regionPathVar: existingRegion,
			},
			ExpectedStatusCode:    http.StatusOK,
			ExpectedSuccessOutput: validResponse,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			defer mock.AssertExpectationsForObjects(t, subscriberSvc)

			handler := tenantfetchersvc.NewTenantsHTTPHandler(subscriberSvc, validHandlerConfig)
			req := testCase.Request
			req = mux.SetURLVars(req, testCase.PathParams)

			w := httptest.NewRecorder()

			// WHEN
			handler.Dependencies(w, req)

			// THEN
			resp := w.Result()
			body, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err)

			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Contains(t, string(body), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}

			if testCase.ExpectedSuccessOutput != "" {
				assert.Equal(t, testCase.ExpectedSuccessOutput, string(body))
			}

			assert.Equal(t, testCase.ExpectedStatusCode, resp.StatusCode)
		})
	}
}
func TestService_FetchTenantOnDemand(t *testing.T) {
	const (
		parentIDPathVar = "tenantId"
		tenantIDPathVar = "parentTenantId"
		parentID        = "fd116270-b71d-4c49-a4d7-4a03785a5e6a"
		tenantID        = "f09ba084-0e82-49ab-ab2e-b7ecc988312d"
	)

	target := fmt.Sprintf("/v1/fetch/:%s/:%s", parentIDPathVar, tenantIDPathVar)

	validHandlerConfig := tenantfetchersvc.HandlerConfig{
		TenantPathParam:       "tenantId",
		ParentTenantPathParam: "parentTenantId",
	}

	testCases := []struct {
		Name                string
		Request             *http.Request
		PathParams          map[string]string
		TenantFetcherSvc    func() *automock.TenantFetcher
		ExpectedErrorOutput string
		ExpectedStatusCode  int
	}{
		{
			Name:    "Successful fetch on-demand",
			Request: httptest.NewRequest(http.MethodGet, target, nil),
			PathParams: map[string]string{
				validHandlerConfig.ParentTenantPathParam: parentID,
				validHandlerConfig.TenantPathParam:       tenantID,
			},
			TenantFetcherSvc: func() *automock.TenantFetcher {
				svc := &automock.TenantFetcher{}
				svc.On("FetchTenantOnDemand", mock.Anything, tenantID, parentID).Return(nil)
				return svc
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:    "Failure when parent ID is missing",
			Request: httptest.NewRequest(http.MethodGet, target, nil),
			PathParams: map[string]string{
				validHandlerConfig.ParentTenantPathParam: "",
				validHandlerConfig.TenantPathParam:       tenantID,
			},
			TenantFetcherSvc:    func() *automock.TenantFetcher { return &automock.TenantFetcher{} },
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: "Parent tenant ID path parameter is missing from request",
		},
		{
			Name:    "Failure when tenant ID is missing",
			Request: httptest.NewRequest(http.MethodGet, target, nil),
			PathParams: map[string]string{
				validHandlerConfig.ParentTenantPathParam: parentIDPathVar,
				validHandlerConfig.TenantPathParam:       "",
			},
			TenantFetcherSvc:    func() *automock.TenantFetcher { return &automock.TenantFetcher{} },
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: "Tenant path parameter is missing from request",
		},
		{
			Name:    "Failure when fetch on-demand returns an error",
			Request: httptest.NewRequest(http.MethodGet, target, nil),
			PathParams: map[string]string{
				validHandlerConfig.ParentTenantPathParam: parentID,
				validHandlerConfig.TenantPathParam:       tenantID,
			},
			TenantFetcherSvc: func() *automock.TenantFetcher {
				svc := &automock.TenantFetcher{}
				svc.On("FetchTenantOnDemand", mock.Anything, tenantID, parentID).Return(errors.New("error"))
				return svc
			},
			ExpectedStatusCode: http.StatusInternalServerError,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tf := testCase.TenantFetcherSvc()
			defer mock.AssertExpectationsForObjects(t, tf)

			handler := tenantfetchersvc.NewTenantFetcherHTTPHandler(tf, validHandlerConfig)
			req := testCase.Request
			req = mux.SetURLVars(req, testCase.PathParams)

			w := httptest.NewRecorder()

			// WHEN
			handler.FetchTenantOnDemand(w, req)

			// THEN
			resp := w.Result()
			body, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err)

			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Contains(t, string(body), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.ExpectedStatusCode, resp.StatusCode)
		})
	}
}
