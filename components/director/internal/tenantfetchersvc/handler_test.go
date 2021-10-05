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
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	validHandlerConfig = tenantfetchersvc.HandlerConfig{
		RegionPathParam:               "region",
		SubscriptionProviderLabelKey:  "SubscriptionProviderLabelKey",
		ConsumerSubaccountIDsLabelKey: "ConsumerSubaccountIDsLabelKey",
		TenantProviderConfig: tenantfetchersvc.TenantProviderConfig{
			TenantProvider:                 testProviderName,
			TenantIDProperty:               tenantProviderTenantIDProperty,
			CustomerIDProperty:             tenantProviderCustomerIDProperty,
			SubdomainProperty:              tenantProviderSubdomainProperty,
			SubscriptionProviderIDProperty: subscriptionProviderIDProperty,
			SubaccountTenantIDProperty:     tenantProviderSubaccountTenantIDProperty,
		},
	}
)

type tenantCreationRequest struct {
	TenantID               string `json:"tenantId"`
	CustomerID             string `json:"customerId"`
	Subdomain              string `json:"subdomain"`
	SubscriptionProviderID string `json:"subscriptionProviderId"`
	SubaccountID           string `json:"subaccountTenantId"`
}

type regionalTenantCreationRequest struct {
	SubaccountID           string `json:"subaccountTenantId"`
	TenantID               string `json:"tenantId"`
	Subdomain              string `json:"subdomain"`
	SubscriptionProviderID string `json:"subscriptionProviderId"`
}

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

func TestService_Create(t *testing.T) {
	//GIVEN
	txGen := txtest.NewTransactionContextGenerator(errors.New("err"))
	target := "http://example.com/foo"

	validRequestBody, err := json.Marshal(tenantCreationRequest{
		TenantID:               tenantExtID,
		CustomerID:             parentTenantExtID,
		Subdomain:              tenantSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
	})
	assert.NoError(t, err)

	bodyWithMissingTenant, err := json.Marshal(tenantCreationRequest{
		CustomerID:             parentTenantExtID,
		Subdomain:              tenantSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
	})
	assert.NoError(t, err)

	bodyWithMissingParent, err := json.Marshal(tenantCreationRequest{
		TenantID:               tenantExtID,
		Subdomain:              tenantSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
	})
	assert.NoError(t, err)
	bodyWithMathcingChild, err := json.Marshal(tenantCreationRequest{
		TenantID:               tenantExtID,
		SubaccountID:           tenantExtID,
		CustomerID:             parentTenantExtID,
		Subdomain:              tenantSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
	})
	assert.NoError(t, err)

	bodyWithMathcingParent, err := json.Marshal(tenantCreationRequest{
		TenantID:               tenantExtID,
		CustomerID:             tenantExtID,
		Subdomain:              tenantSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
	})
	assert.NoError(t, err)

	bodyWithMissingTenantSubdomain, err := json.Marshal(tenantCreationRequest{
		TenantID:               tenantExtID,
		CustomerID:             parentTenantExtID,
		SubscriptionProviderID: subscriptionProviderID,
	})
	assert.NoError(t, err)

	bodyWithMissingSubscriptionConsumerID, err := json.Marshal(tenantCreationRequest{
		TenantID:   tenantExtID,
		Subdomain:  tenantSubdomain,
		CustomerID: parentTenantExtID,
	})
	assert.NoError(t, err)

	accountProvisioningRequest := tenantfetchersvc.TenantSubscriptionRequest{
		AccountTenantID:        tenantExtID,
		CustomerTenantID:       parentTenantExtID,
		Subdomain:              tenantSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
	}
	accountWithoutParentProvisioningRequest := tenantfetchersvc.TenantSubscriptionRequest{
		AccountTenantID:        tenantExtID,
		Subdomain:              tenantSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
	}

	testCases := []struct {
		Name                  string
		TenantSubscriberFn    func() *automock.TenantSubscriber
		TxFn                  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		Request               *http.Request
		ExpectedErrorOutput   string
		ExpectedSuccessOutput string
		ExpectedStatusCode    int
	}{
		{
			Name: "Succeeds",
			TxFn: txGen.ThatSucceeds,
			TenantSubscriberFn: func() *automock.TenantSubscriber {
				subscriber := &automock.TenantSubscriber{}
				subscriber.On("Subscribe", txtest.CtxWithDBMatcher(), &accountProvisioningRequest, "").Return(nil).Once()
				return subscriber
			},
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name: "Succeeds when parent tenant is not found in body",
			TxFn: txGen.ThatSucceeds,
			TenantSubscriberFn: func() *automock.TenantSubscriber {
				subscriber := &automock.TenantSubscriber{}
				subscriber.On("Subscribe", txtest.CtxWithDBMatcher(), &accountWithoutParentProvisioningRequest, "").Return(nil).Once()
				return subscriber
			},
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingParent)),
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name: "Succeeds when matching subaccount and account tenant IDs are provided",
			TxFn: txGen.ThatSucceeds,
			TenantSubscriberFn: func() *automock.TenantSubscriber {
				subscriber := &automock.TenantSubscriber{}
				subscriber.On("Subscribe", txtest.CtxWithDBMatcher(), &accountProvisioningRequest, "").Return(nil).Once()
				return subscriber
			},
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMathcingChild)),
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name: "Succeeds when matching customer and account tenant IDs are provided",
			TxFn: txGen.ThatSucceeds,
			TenantSubscriberFn: func() *automock.TenantSubscriber {
				subscriber := &automock.TenantSubscriber{}
				subscriber.On("Subscribe", txtest.CtxWithDBMatcher(), &accountWithoutParentProvisioningRequest, "").Return(nil).Once()
				return subscriber
			},
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMathcingParent)),
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name:                "Returns error when reading request body fails",
			TxFn:                txGen.ThatDoesntStartTransaction,
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			Request:             httptest.NewRequest(http.MethodPut, target, errReader(0)),
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name:                "Returns error when request body doesn't contain tenantID",
			TxFn:                txGen.ThatDoesntStartTransaction,
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingTenant)),
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", tenantProviderTenantIDProperty),
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			Name:                "Returns error when request body doesn't contain tenant subdomain",
			TxFn:                txGen.ThatDoesntStartTransaction,
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingTenantSubdomain)),
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", tenantProviderSubdomainProperty),
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			Name:                "Returns error when request body doesn't contain SubscriptionProviderID",
			TxFn:                txGen.ThatDoesntStartTransaction,
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingSubscriptionConsumerID)),
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", subscriptionProviderIDProperty),
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			Name:                "Returns error when beginning transaction fails",
			TxFn:                txGen.ThatFailsOnBegin,
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name: "Returns error when tenant creation fails",
			TxFn: txGen.ThatSucceeds,
			TenantSubscriberFn: func() *automock.TenantSubscriber {
				subscriber := &automock.TenantSubscriber{}
				subscriber.On("Subscribe", txtest.CtxWithDBMatcher(), &accountProvisioningRequest, "").Return(testError).Once()
				return subscriber
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
		},
		{
			Name: "Returns error when transaction commit fails",
			TxFn: txGen.ThatFailsOnCommit,
			TenantSubscriberFn: func() *automock.TenantSubscriber {
				subscriber := &automock.TenantSubscriber{}
				subscriber.On("Subscribe", txtest.CtxWithDBMatcher(), &accountProvisioningRequest, "").Return(nil).Once()
				return subscriber
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, transact := testCase.TxFn()
			subscriber := testCase.TenantSubscriberFn()
			defer mock.AssertExpectationsForObjects(t, transact, subscriber)

			handler := tenantfetchersvc.NewTenantsHTTPHandler(subscriber, transact, validHandlerConfig)
			req := testCase.Request
			w := httptest.NewRecorder()

			//WHEN
			handler.Create(w, req)

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

func TestService_SubscriptionFlows(t *testing.T) {
	//GIVEN
	region := "eu-1"

	txGen := txtest.NewTransactionContextGenerator(errors.New("err"))
	target := "http://example.com/foo/:region"
	txtest.CtxWithDBMatcher()

	validRequestBody, err := json.Marshal(regionalTenantCreationRequest{
		SubaccountID:           subaccountTenantExtID,
		TenantID:               tenantExtID,
		Subdomain:              regionalTenantSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
	})
	assert.NoError(t, err)

	bodyWithMissingParent, err := json.Marshal(regionalTenantCreationRequest{
		SubaccountID:           subaccountTenantExtID,
		Subdomain:              tenantSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
	})
	assert.NoError(t, err)

	bodyWithMissingTenantSubdomain, err := json.Marshal(regionalTenantCreationRequest{
		SubaccountID:           subaccountTenantExtID,
		TenantID:               tenantExtID,
		SubscriptionProviderID: subscriptionProviderID,
	})
	assert.NoError(t, err)

	bodyWithMissingSubscriptionConsumerID, err := json.Marshal(regionalTenantCreationRequest{
		SubaccountID: subaccountTenantExtID,
		TenantID:     tenantExtID,
		Subdomain:    regionalTenantSubdomain,
	})
	assert.NoError(t, err)

	bodyWithMatchingAccountAndSubaccountIDs, err := json.Marshal(regionalTenantCreationRequest{
		TenantID:               tenantExtID,
		SubaccountID:           tenantExtID,
		Subdomain:              regionalTenantSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
	})
	assert.NoError(t, err)

	validHandlerConfig := tenantfetchersvc.HandlerConfig{
		RegionPathParam:               "region",
		SubscriptionProviderLabelKey:  "SubscriptionProviderLabelKey",
		ConsumerSubaccountIDsLabelKey: "ConsumerSubaccountIDsLabelKey",
		TenantProviderConfig: tenantfetchersvc.TenantProviderConfig{
			TenantProvider:                 testProviderName,
			TenantIDProperty:               tenantProviderTenantIDProperty,
			SubaccountTenantIDProperty:     tenantProviderSubaccountTenantIDProperty,
			CustomerIDProperty:             tenantProviderCustomerIDProperty,
			SubdomainProperty:              tenantProviderSubdomainProperty,
			SubscriptionProviderIDProperty: subscriptionProviderIDProperty,
		},
	}
	regionalTenant := tenantfetchersvc.TenantSubscriptionRequest{
		SubaccountTenantID:     subaccountTenantExtID,
		AccountTenantID:        tenantExtID,
		Subdomain:              regionalTenantSubdomain,
		Region:                 region,
		SubscriptionProviderID: subscriptionProviderID,
	}
	regionalTenantWithMatchingParentID := tenantfetchersvc.TenantSubscriptionRequest{
		SubaccountTenantID:     "",
		AccountTenantID:        tenantExtID,
		Subdomain:              regionalTenantSubdomain,
		Region:                 region,
		SubscriptionProviderID: subscriptionProviderID,
	}

	// Subscribe flow
	testCases := []struct {
		Name                  string
		TenantSubscriberFn    func() *automock.TenantSubscriber
		TxFn                  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
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
				subscriber.On("Subscribe", txtest.CtxWithDBMatcher(), &regionalTenant, region).Return(nil).Once()
				return subscriber
			},
			TxFn:                  txGen.ThatSucceeds,
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:                region,
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name: "Succeeds to create account tenant when account ID and subaccount IDs are matching",
			TenantSubscriberFn: func() *automock.TenantSubscriber {
				subscriber := &automock.TenantSubscriber{}
				subscriber.On("Subscribe", txtest.CtxWithDBMatcher(), &regionalTenantWithMatchingParentID, region).Return(nil).Once()
				return subscriber
			},
			TxFn:                  txGen.ThatSucceeds,
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMatchingAccountAndSubaccountIDs)),
			Region:                region,
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name:                "Returns error when region path parameter is missing",
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			TxFn:                txGen.ThatDoesntStartTransaction,
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: "Region path parameter is missing from request",
		},
		{
			Name:                "Returns error when parent tenant is not found in body",
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			TxFn:                txGen.ThatDoesntStartTransaction,
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingParent)),
			Region:              region,
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", tenantProviderTenantIDProperty),
		},
		{
			Name:                "Returns error when SubscriptionProviderID is not found in body",
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			TxFn:                txGen.ThatDoesntStartTransaction,
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingSubscriptionConsumerID)),
			Region:              region,
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", subscriptionProviderIDProperty),
		},
		{
			Name:                "Returns error when request body doesn't contain tenant subdomain",
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			TxFn:                txGen.ThatDoesntStartTransaction,
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingTenantSubdomain)),
			Region:              region,
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", tenantProviderSubdomainProperty),
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			Name:                "Returns error when reading request body fails",
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			TxFn:                txGen.ThatDoesntStartTransaction,
			Request:             httptest.NewRequest(http.MethodPut, target, errReader(0)),
			Region:              region,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name:                "Returns error when beginning transaction fails",
			TenantSubscriberFn:  func() *automock.TenantSubscriber { return &automock.TenantSubscriber{} },
			TxFn:                txGen.ThatFailsOnBegin,
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name: "Returns error when tenant subscription fails",
			TenantSubscriberFn: func() *automock.TenantSubscriber {
				subscriber := &automock.TenantSubscriber{}
				subscriber.On("Subscribe", txtest.CtxWithDBMatcher(), &regionalTenant, region).Return(testError).Once()
				return subscriber
			},
			TxFn:                txGen.ThatDoesntExpectCommit,
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
		},
		{
			Name: "Returns error when transaction commit fails",
			TenantSubscriberFn: func() *automock.TenantSubscriber {
				subscriber := &automock.TenantSubscriber{}
				subscriber.On("Subscribe", txtest.CtxWithDBMatcher(), &regionalTenant, region).Return(nil).Once()
				return subscriber
			},
			TxFn:                txGen.ThatFailsOnCommit,
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, transact := testCase.TxFn()
			subscriber := testCase.TenantSubscriberFn()
			defer mock.AssertExpectationsForObjects(t, transact, subscriber)

			handler := tenantfetchersvc.NewTenantsHTTPHandler(subscriber, transact, validHandlerConfig)
			req := testCase.Request

			if len(testCase.Region) > 0 {
				vars := map[string]string{
					"region": testCase.Region,
				}
				req = mux.SetURLVars(req, vars)
			}

			w := httptest.NewRecorder()

			//WHEN
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
		_, transact := txGen.ThatSucceeds()
		subscriber := &automock.TenantSubscriber{}
		subscriber.On("Unsubscribe", txtest.CtxWithDBMatcher(), &regionalTenant, region).Return(testError).Once()
		defer mock.AssertExpectationsForObjects(t, transact, subscriber)

		handler := tenantfetchersvc.NewTenantsHTTPHandler(subscriber, transact, validHandlerConfig)
		req := httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody))

		vars := map[string]string{
			"region": region,
		}
		req = mux.SetURLVars(req, vars)

		w := httptest.NewRecorder()

		//WHEN
		handler.UnSubscribeTenant(w, req)

		// THEN
		resp := w.Result()
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)

		assert.Contains(t, string(body), tenantfetchersvc.InternalServerError)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestService_Delete(t *testing.T) {
	//GIVEN
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)
	target := "http://example.com/foo"

	t.Run("Succeeds", func(t *testing.T) {
		requestBody, err := json.Marshal(tenantCreationRequest{
			TenantID:               tenantExtID,
			CustomerID:             parentTenantExtID,
			Subdomain:              tenantSubdomain,
			SubscriptionProviderID: subscriptionProviderID,
		})
		assert.NoError(t, err)

		_, transact := txGen.ThatDoesntStartTransaction()
		subscriber := &automock.TenantSubscriber{}
		defer mock.AssertExpectationsForObjects(t, transact, subscriber)

		handler := tenantfetchersvc.NewTenantsHTTPHandler(subscriber, transact, validHandlerConfig)
		req := httptest.NewRequest(http.MethodDelete, target, bytes.NewBuffer(requestBody))
		w := httptest.NewRecorder()

		//WHEN
		handler.DeleteByExternalID(w, req)
		// THEN
		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
	t.Run("Succeeds when tenant cannot be read from body", func(t *testing.T) {
		_, transact := txGen.ThatDoesntStartTransaction()
		subscriber := &automock.TenantSubscriber{}
		defer mock.AssertExpectationsForObjects(t, transact, subscriber)

		handler := tenantfetchersvc.NewTenantsHTTPHandler(subscriber, transact, validHandlerConfig)
		req := httptest.NewRequest(http.MethodDelete, target, errReader(0))
		w := httptest.NewRecorder()

		//WHEN
		handler.DeleteByExternalID(w, req)

		// THEN
		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
	t.Run("Succeeds when tenant is missing from body", func(t *testing.T) {
		requestBody, err := json.Marshal(tenantCreationRequest{
			CustomerID:             parentTenantExtID,
			Subdomain:              tenantSubdomain,
			SubscriptionProviderID: subscriptionProviderID,
		})
		assert.NoError(t, err)

		_, transact := txGen.ThatDoesntStartTransaction()
		subscriber := &automock.TenantSubscriber{}
		defer mock.AssertExpectationsForObjects(t, transact, subscriber)

		handler := tenantfetchersvc.NewTenantsHTTPHandler(subscriber, transact, validHandlerConfig)
		req := httptest.NewRequest(http.MethodDelete, target, bytes.NewBuffer(requestBody))
		w := httptest.NewRecorder()

		//WHEN
		handler.DeleteByExternalID(w, req)

		// THEN
		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
