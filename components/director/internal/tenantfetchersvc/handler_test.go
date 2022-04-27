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
	SubaccountID           string `json:"subaccountTenantId"`
	TenantID               string `json:"tenantId"`
	Subdomain              string `json:"subdomain"`
	SubscriptionProviderID string `json:"subscriptionProviderId"`
	ProviderSubaccountID   string `json:"providerSubaccountId"`
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
		SubaccountID:           subaccountTenantExtID,
		TenantID:               tenantExtID,
		Subdomain:              regionalTenantSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
		ProviderSubaccountID:   providerSubaccountID,
	})
	assert.NoError(t, err)

	bodyWithMissingParent, err := json.Marshal(regionalTenantCreationRequest{
		SubaccountID:           subaccountTenantExtID,
		Subdomain:              regionalTenantSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
		ProviderSubaccountID:   providerSubaccountID,
	})
	assert.NoError(t, err)

	bodyWithMissingTenantSubdomain, err := json.Marshal(regionalTenantCreationRequest{
		SubaccountID:           subaccountTenantExtID,
		TenantID:               tenantExtID,
		SubscriptionProviderID: subscriptionProviderID,
		ProviderSubaccountID:   providerSubaccountID,
	})
	assert.NoError(t, err)

	bodyWithMissingSubscriptionConsumerID, err := json.Marshal(regionalTenantCreationRequest{
		SubaccountID:         subaccountTenantExtID,
		TenantID:             tenantExtID,
		Subdomain:            regionalTenantSubdomain,
		ProviderSubaccountID: providerSubaccountID,
	})
	assert.NoError(t, err)

	bodyWithMatchingAccountAndSubaccountIDs, err := json.Marshal(regionalTenantCreationRequest{
		TenantID:               tenantExtID,
		SubaccountID:           tenantExtID,
		Subdomain:              regionalTenantSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
		ProviderSubaccountID:   providerSubaccountID,
	})

	bodyWithMissingProviderSubaccountID, err := json.Marshal(regionalTenantCreationRequest{
		TenantID:               tenantExtID,
		SubaccountID:           tenantExtID,
		Subdomain:              regionalTenantSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
	})

	assert.NoError(t, err)

	validHandlerConfig := tenantfetchersvc.HandlerConfig{
		RegionPathParam: "region",
		TenantProviderConfig: tenantfetchersvc.TenantProviderConfig{
			TenantProvider:                 testProviderName,
			TenantIDProperty:               tenantProviderTenantIDProperty,
			SubaccountTenantIDProperty:     tenantProviderSubaccountTenantIDProperty,
			CustomerIDProperty:             tenantProviderCustomerIDProperty,
			SubdomainProperty:              tenantProviderSubdomainProperty,
			SubscriptionProviderIDProperty: subscriptionProviderIDProperty,
			ProviderSubaccountIDProperty:   providerSubaccountIDProperty,
		},
	}
	regionalTenant := tenantfetchersvc.TenantSubscriptionRequest{
		SubaccountTenantID:     subaccountTenantExtID,
		AccountTenantID:        tenantExtID,
		Subdomain:              regionalTenantSubdomain,
		Region:                 region,
		SubscriptionProviderID: subscriptionProviderID,
		ProviderSubaccountID:   providerSubaccountID,
	}
	regionalTenantWithMatchingParentID := tenantfetchersvc.TenantSubscriptionRequest{
		SubaccountTenantID:     "",
		AccountTenantID:        tenantExtID,
		Subdomain:              regionalTenantSubdomain,
		Region:                 region,
		SubscriptionProviderID: subscriptionProviderID,
		ProviderSubaccountID:   providerSubaccountID,
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
