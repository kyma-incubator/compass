package tenantfetchersvc_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	tenantExtID     = "tenant-external-id"
	tenantSubdomain = "mytenant"
	tenantRegion    = "myregion"

	regionalTenantSubdomain = "myregionaltenant"
	subaccountTenantExtID   = "subaccount-tenant-external-id"
	subscriptionConsumerID  = "123"

	parentTenantExtID = "parent-tenant-external-id"

	tenantProviderTenantIDProperty           = "tenantId"
	tenantProviderCustomerIDProperty         = "customerId"
	tenantProviderSubdomainProperty          = "subdomain"
	tenantProviderSubaccountTenantIDProperty = "subaccountTenantId"
	subscriptionConsumerIDProperty           = "subscriptionConsumerId"

	compassURL = "https://github.com/kyma-incubator/compass"
)

var (
	testError          = errors.New("test error")
	notFoundErr        = apperrors.NewNotFoundErrorWithType(resource.Runtime)
	validHandlerConfig = tenantfetchersvc.HandlerConfig{
		RegionPathParam:               "region",
		RegionLabelKey:                "regionKey",
		SubscriptionConsumerLabelKey:  "SubscriptionConsumerLabelKey",
		ConsumerSubaccountIDsLabelKey: "ConsumerSubaccountIDsLabelKey",
		TenantProviderConfig: tenantfetchersvc.TenantProviderConfig{
			TenantProvider:                 testProviderName,
			TenantIDProperty:               tenantProviderTenantIDProperty,
			CustomerIDProperty:             tenantProviderCustomerIDProperty,
			SubdomainProperty:              tenantProviderSubdomainProperty,
			SubscriptionConsumerIDProperty: subscriptionConsumerIDProperty,
			SubaccountTenantIDProperty:     tenantProviderSubaccountTenantIDProperty,
		},
	}
	testRuntime = model.Runtime{
		ID:                "321",
		Name:              "test",
		Description:       nil,
		Tenant:            "test-tenant",
		Status:            nil,
		CreationTimestamp: time.Time{},
	}
	invalidTestLabel = model.Label{
		ID:         "456",
		Tenant:     "test-tenant",
		Key:        validHandlerConfig.ConsumerSubaccountIDsLabelKey,
		Value:      "",
		ObjectID:   testRuntime.ID,
		ObjectType: model.RuntimeLabelableObject,
	}
	testLabel = model.Label{
		ID:         "456",
		Tenant:     "test-tenant",
		Key:        validHandlerConfig.ConsumerSubaccountIDsLabelKey,
		Value:      []interface{}{"789"},
		ObjectID:   testRuntime.ID,
		ObjectType: model.RuntimeLabelableObject,
	}
	updateLabelInput = model.LabelInput{
		Key:        validHandlerConfig.ConsumerSubaccountIDsLabelKey,
		Value:      []string{"789", subaccountTenantExtID},
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   testRuntime.ID,
	}
	createLabelInput = model.LabelInput{
		Key:        validHandlerConfig.ConsumerSubaccountIDsLabelKey,
		Value:      []string{subaccountTenantExtID},
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   testRuntime.ID,
	}
	removeLabelInput = model.LabelInput{
		Key:        validHandlerConfig.ConsumerSubaccountIDsLabelKey,
		Value:      []string{"789"},
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   testRuntime.ID,
	}
	emptyLabelInput = model.LabelInput{
		Key:        validHandlerConfig.ConsumerSubaccountIDsLabelKey,
		Value:      []string{},
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   testRuntime.ID,
	}
)

type tenantCreationRequest struct {
	TenantID               string `json:"tenantId"`
	CustomerID             string `json:"customerId"`
	Subdomain              string `json:"subdomain"`
	SubscriptionConsumerID string `json:"subscriptionConsumerId"`
	SubaccountID           string `json:"subaccountTenantId"`
}

type regionalTenantCreationRequest struct {
	SubaccountID           string `json:"subaccountTenantId"`
	TenantID               string `json:"tenantId"`
	Subdomain              string `json:"subdomain"`
	SubscriptionConsumerID string `json:"subscriptionConsumerId"`
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
		SubscriptionConsumerID: subscriptionConsumerID,
	})
	assert.NoError(t, err)

	bodyWithMissingTenant, err := json.Marshal(tenantCreationRequest{
		CustomerID:             parentTenantExtID,
		Subdomain:              tenantSubdomain,
		SubscriptionConsumerID: subscriptionConsumerID,
	})
	assert.NoError(t, err)

	bodyWithMissingParent, err := json.Marshal(tenantCreationRequest{
		TenantID:               tenantExtID,
		Subdomain:              tenantSubdomain,
		SubscriptionConsumerID: subscriptionConsumerID,
	})
	assert.NoError(t, err)
	bodyWithMathcingChild, err := json.Marshal(tenantCreationRequest{
		TenantID:               tenantExtID,
		SubaccountID:           tenantExtID,
		CustomerID:             parentTenantExtID,
		Subdomain:              tenantSubdomain,
		SubscriptionConsumerID: subscriptionConsumerID,
	})
	assert.NoError(t, err)

	bodyWithMissingTenantSubdomain, err := json.Marshal(tenantCreationRequest{
		TenantID:               tenantExtID,
		CustomerID:             parentTenantExtID,
		SubscriptionConsumerID: subscriptionConsumerID,
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
		SubscriptionConsumerID: subscriptionConsumerID,
	}
	accountWithoutParentProvisioningRequest := tenantfetchersvc.TenantSubscriptionRequest{
		AccountTenantID:        tenantExtID,
		Subdomain:              tenantSubdomain,
		SubscriptionConsumerID: subscriptionConsumerID,
	}

	testCases := []struct {
		Name                  string
		TenantProvisionerFn   func() *automock.TenantProvisioner
		TxFn                  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		Request               *http.Request
		ExpectedErrorOutput   string
		ExpectedSuccessOutput string
		ExpectedStatusCode    int
	}{
		{
			Name: "Succeeds",
			TxFn: txGen.ThatSucceeds,
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", txtest.CtxWithDBMatcher(), accountProvisioningRequest).Return(nil).Once()
				return provisioner
			},
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name: "Succeeds when parent tenant is not found in body",
			TxFn: txGen.ThatSucceeds,
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", txtest.CtxWithDBMatcher(), accountWithoutParentProvisioningRequest).Return(nil).Once()
				return provisioner
			},
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingParent)),
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name: "Succeeds when matching child tenant ID is provided",
			TxFn: txGen.ThatSucceeds,
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", txtest.CtxWithDBMatcher(), accountProvisioningRequest).Return(nil).Once()
				return provisioner
			},
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMathcingChild)),
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name:                "Returns error when reading request body fails",
			TxFn:                txGen.ThatDoesntStartTransaction,
			TenantProvisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			Request:             httptest.NewRequest(http.MethodPut, target, errReader(0)),
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name:                "Returns error when request body doesn't contain tenantID",
			TxFn:                txGen.ThatDoesntStartTransaction,
			TenantProvisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingTenant)),
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", tenantProviderTenantIDProperty),
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			Name:                "Returns error when request body doesn't contain tenant subdomain",
			TxFn:                txGen.ThatDoesntStartTransaction,
			TenantProvisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingTenantSubdomain)),
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", tenantProviderSubdomainProperty),
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			Name:                "Returns error when request body doesn't contain SubscriptionConsumerID",
			TxFn:                txGen.ThatDoesntStartTransaction,
			TenantProvisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingSubscriptionConsumerID)),
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", subscriptionConsumerIDProperty),
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			Name:                "Returns error when beginning transaction fails",
			TxFn:                txGen.ThatFailsOnBegin,
			TenantProvisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name: "Returns error when tenant creation fails",
			TxFn: txGen.ThatSucceeds,
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", txtest.CtxWithDBMatcher(), accountProvisioningRequest).Return(testError).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
		},
		{
			Name: "Returns error when transaction commit fails",
			TxFn: txGen.ThatFailsOnCommit,
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", txtest.CtxWithDBMatcher(), accountProvisioningRequest).Return(nil).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, transact := testCase.TxFn()
			provisioner := testCase.TenantProvisionerFn()
			defer mock.AssertExpectationsForObjects(t, transact, provisioner)

			handler := tenantfetchersvc.NewTenantsHTTPHandler(provisioner, &automock.RuntimeService{}, transact, validHandlerConfig)
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
		SubscriptionConsumerID: subscriptionConsumerID,
	})
	assert.NoError(t, err)

	bodyWithMissingParent, err := json.Marshal(regionalTenantCreationRequest{
		SubaccountID:           subaccountTenantExtID,
		Subdomain:              tenantSubdomain,
		SubscriptionConsumerID: subscriptionConsumerID,
	})
	assert.NoError(t, err)

	bodyWithMissingTenantSubdomain, err := json.Marshal(regionalTenantCreationRequest{
		SubaccountID:           subaccountTenantExtID,
		TenantID:               tenantExtID,
		SubscriptionConsumerID: subscriptionConsumerID,
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
		SubscriptionConsumerID: subscriptionConsumerID,
	})
	assert.NoError(t, err)

	validHandlerConfig := tenantfetchersvc.HandlerConfig{
		RegionPathParam:               "region",
		RegionLabelKey:                "regionKey",
		SubscriptionConsumerLabelKey:  "SubscriptionConsumerLabelKey",
		ConsumerSubaccountIDsLabelKey: "ConsumerSubaccountIDsLabelKey",
		TenantProviderConfig: tenantfetchersvc.TenantProviderConfig{
			TenantProvider:                 testProviderName,
			TenantIDProperty:               tenantProviderTenantIDProperty,
			SubaccountTenantIDProperty:     tenantProviderSubaccountTenantIDProperty,
			CustomerIDProperty:             tenantProviderCustomerIDProperty,
			SubdomainProperty:              tenantProviderSubdomainProperty,
			SubscriptionConsumerIDProperty: subscriptionConsumerIDProperty,
		},
	}
	regionalTenant := tenantfetchersvc.TenantSubscriptionRequest{
		SubaccountTenantID:     subaccountTenantExtID,
		AccountTenantID:        tenantExtID,
		Subdomain:              regionalTenantSubdomain,
		Region:                 region,
		SubscriptionConsumerID: subscriptionConsumerID,
	}

	accountTenantReq := tenantfetchersvc.TenantSubscriptionRequest{
		AccountTenantID:        tenantExtID,
		Subdomain:              regionalTenantSubdomain,
		Region:                 region,
		SubscriptionConsumerID: subscriptionConsumerID,
	}

	filters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(validHandlerConfig.SubscriptionConsumerLabelKey, fmt.Sprintf("\"%s\"", subscriptionConsumerID)),
		labelfilter.NewForKeyWithQuery(validHandlerConfig.RegionLabelKey, fmt.Sprintf("\"%s\"", region)),
	}

	// Subscribe flow
	testCases := []struct {
		Name                  string
		provisionerFn         func() *automock.TenantProvisioner
		runtimeServiceFn      func() *automock.RuntimeService
		TxFn                  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		Request               *http.Request
		Region                string
		ExpectedErrorOutput   string
		ExpectedSuccessOutput string
		ExpectedStatusCode    int
	}{
		{
			Name: "Succeeds",
			TxFn: txGen.ThatSucceeds,
			provisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionRegionalTenants", txtest.CtxWithDBMatcher(), regionalTenant).Return(nil).Once()
				return provisioner
			},
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return([]*model.Runtime{}, nil).Once()
				return provisioner
			},
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:                region,
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name: "Succeeds to create account tenant when account ID and subaccount IDs are matching",
			TxFn: txGen.ThatSucceeds,
			provisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionRegionalTenants", txtest.CtxWithDBMatcher(), accountTenantReq).Return(nil).Once()
				return provisioner
			},
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return([]*model.Runtime{}, nil).Once()
				return provisioner
			},
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMatchingAccountAndSubaccountIDs)),
			Region:                region,
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name:                "Returns error when region path parameter is missing",
			TxFn:                txGen.ThatDoesntStartTransaction,
			provisionerFn:       func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn:    func() *automock.RuntimeService { return &automock.RuntimeService{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: "Region path parameter is missing from request",
		},
		{
			Name:                "Returns error when parent tenant is not found in body",
			TxFn:                txGen.ThatDoesntStartTransaction,
			provisionerFn:       func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn:    func() *automock.RuntimeService { return &automock.RuntimeService{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingParent)),
			Region:              region,
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", tenantProviderTenantIDProperty),
		},
		{
			Name:                "Returns error when SubscriptionConsumerID is not found in body",
			TxFn:                txGen.ThatDoesntStartTransaction,
			provisionerFn:       func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn:    func() *automock.RuntimeService { return &automock.RuntimeService{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingSubscriptionConsumerID)),
			Region:              region,
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", subscriptionConsumerIDProperty),
		},
		{
			Name:                "Returns error when request body doesn't contain tenant subdomain",
			TxFn:                txGen.ThatDoesntStartTransaction,
			provisionerFn:       func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn:    func() *automock.RuntimeService { return &automock.RuntimeService{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingTenantSubdomain)),
			Region:              region,
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", tenantProviderSubdomainProperty),
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			Name:                "Returns error when reading request body fails",
			TxFn:                txGen.ThatDoesntStartTransaction,
			provisionerFn:       func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn:    func() *automock.RuntimeService { return &automock.RuntimeService{} },
			Request:             httptest.NewRequest(http.MethodPut, target, errReader(0)),
			Region:              region,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name:                "Returns error when beginning transaction fails",
			TxFn:                txGen.ThatFailsOnBegin,
			provisionerFn:       func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn:    func() *automock.RuntimeService { return &automock.RuntimeService{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name: "Returns error when tenant creation fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			provisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionRegionalTenants", txtest.CtxWithDBMatcher(), regionalTenant).Return(testError).Once()
				return provisioner
			},
			runtimeServiceFn:    func() *automock.RuntimeService { return &automock.RuntimeService{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
		},
		{
			Name: "Succeeds when can't find runtimes",
			TxFn: txGen.ThatSucceeds,
			provisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionRegionalTenants", txtest.CtxWithDBMatcher(), regionalTenant).Return(nil).Once()
				return provisioner
			},
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return(nil, notFoundErr).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: compassURL,
			ExpectedStatusCode:  http.StatusOK,
		},
		{
			Name: "Returns error when could not list runtimes",
			TxFn: txGen.ThatSucceeds,
			provisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionRegionalTenants", txtest.CtxWithDBMatcher(), regionalTenant).Return(nil).Once()
				return provisioner
			},
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return(nil, testError).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name: "Returns error when could not get label for runtime",
			TxFn: txGen.ThatSucceeds,
			provisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionRegionalTenants", txtest.CtxWithDBMatcher(), regionalTenant).Return(nil).Once()
				return provisioner
			},
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", txtest.CtxWithDBMatcher(), testRuntime.ID, validHandlerConfig.ConsumerSubaccountIDsLabelKey).Return(nil, testError).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name: "Returns error when could not parse label value",
			TxFn: txGen.ThatSucceeds,
			provisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionRegionalTenants", txtest.CtxWithDBMatcher(), regionalTenant).Return(nil).Once()
				return provisioner
			},
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", txtest.CtxWithDBMatcher(), testRuntime.ID, validHandlerConfig.ConsumerSubaccountIDsLabelKey).Return(&invalidTestLabel, nil).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name: "Returns error when could not set label for runtime",
			TxFn: txGen.ThatSucceeds,
			provisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionRegionalTenants", txtest.CtxWithDBMatcher(), regionalTenant).Return(nil).Once()
				return provisioner
			},
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", txtest.CtxWithDBMatcher(), testRuntime.ID, validHandlerConfig.ConsumerSubaccountIDsLabelKey).Return(&testLabel, nil).Once()
				provisioner.On("SetLabel", txtest.CtxWithDBMatcher(), &updateLabelInput).Return(testError).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name: "Succeeds and creates label",
			TxFn: txGen.ThatSucceeds,
			provisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionRegionalTenants", txtest.CtxWithDBMatcher(), regionalTenant).Return(nil).Once()
				return provisioner
			},
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", txtest.CtxWithDBMatcher(), testRuntime.ID, validHandlerConfig.ConsumerSubaccountIDsLabelKey).Return(nil, notFoundErr).Once()
				provisioner.On("SetLabel", txtest.CtxWithDBMatcher(), &createLabelInput).Return(nil).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: compassURL,
			ExpectedStatusCode:  http.StatusOK,
		},
		{
			Name: "Succeeds and updates label",
			TxFn: txGen.ThatSucceeds,
			provisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionRegionalTenants", txtest.CtxWithDBMatcher(), regionalTenant).Return(nil).Once()
				return provisioner
			},
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", txtest.CtxWithDBMatcher(), testRuntime.ID, validHandlerConfig.ConsumerSubaccountIDsLabelKey).Return(&testLabel, nil).Once()
				provisioner.On("SetLabel", txtest.CtxWithDBMatcher(), &updateLabelInput).Return(nil).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: compassURL,
			ExpectedStatusCode:  http.StatusOK,
		},
		{
			Name: "Returns error when transaction commit fails",
			TxFn: txGen.ThatFailsOnCommit,
			provisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionRegionalTenants", txtest.CtxWithDBMatcher(), regionalTenant).Return(nil).Once()
				return provisioner
			},
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", txtest.CtxWithDBMatcher(), testRuntime.ID, validHandlerConfig.ConsumerSubaccountIDsLabelKey).Return(&testLabel, nil).Once()
				provisioner.On("SetLabel", txtest.CtxWithDBMatcher(), &updateLabelInput).Return(nil).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, transact := testCase.TxFn()
			provisioner := testCase.provisionerFn()
			runtimeService := testCase.runtimeServiceFn()
			defer mock.AssertExpectationsForObjects(t, transact, provisioner, runtimeService)

			handler := tenantfetchersvc.NewTenantsHTTPHandler(provisioner, runtimeService, transact, validHandlerConfig)
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
	testCases = []struct {
		Name                  string
		provisionerFn         func() *automock.TenantProvisioner
		runtimeServiceFn      func() *automock.RuntimeService
		TxFn                  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		Request               *http.Request
		Region                string
		ExpectedErrorOutput   string
		ExpectedSuccessOutput string
		ExpectedStatusCode    int
	}{
		{
			Name:          "Succeeds",
			TxFn:          txGen.ThatSucceeds,
			provisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return([]*model.Runtime{}, nil).Once()
				return provisioner
			},
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:                region,
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name:                "Returns error when region path parameter is missing",
			TxFn:                txGen.ThatDoesntStartTransaction,
			provisionerFn:       func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn:    func() *automock.RuntimeService { return &automock.RuntimeService{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: "Region path parameter is missing from request",
		},
		{
			Name:                "Returns error when parent tenant is not found in body",
			TxFn:                txGen.ThatDoesntStartTransaction,
			provisionerFn:       func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn:    func() *automock.RuntimeService { return &automock.RuntimeService{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingParent)),
			Region:              region,
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", tenantProviderTenantIDProperty),
		},
		{
			Name:                "Returns error when SubscriptionConsumerID is not found in body",
			TxFn:                txGen.ThatDoesntStartTransaction,
			provisionerFn:       func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn:    func() *automock.RuntimeService { return &automock.RuntimeService{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingSubscriptionConsumerID)),
			Region:              region,
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", subscriptionConsumerIDProperty),
		},
		{
			Name:                "Returns error when request body doesn't contain tenant subdomain",
			TxFn:                txGen.ThatDoesntStartTransaction,
			provisionerFn:       func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn:    func() *automock.RuntimeService { return &automock.RuntimeService{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingTenantSubdomain)),
			Region:              region,
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", tenantProviderSubdomainProperty),
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			Name:                "Returns error when reading request body fails",
			TxFn:                txGen.ThatDoesntStartTransaction,
			provisionerFn:       func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn:    func() *automock.RuntimeService { return &automock.RuntimeService{} },
			Request:             httptest.NewRequest(http.MethodPut, target, errReader(0)),
			Region:              region,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name:                "Returns error when beginning transaction fails",
			TxFn:                txGen.ThatFailsOnBegin,
			provisionerFn:       func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn:    func() *automock.RuntimeService { return &automock.RuntimeService{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name:          "Succeeds when can't find runtimes",
			TxFn:          txGen.ThatSucceeds,
			provisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return(nil, notFoundErr).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: compassURL,
			ExpectedStatusCode:  http.StatusOK,
		},
		{
			Name:          "Returns error when could not list runtimes",
			TxFn:          txGen.ThatSucceeds,
			provisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return(nil, testError).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name:          "Returns error when could not get label for runtime",
			TxFn:          txGen.ThatSucceeds,
			provisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", txtest.CtxWithDBMatcher(), testRuntime.ID, validHandlerConfig.ConsumerSubaccountIDsLabelKey).Return(nil, testError).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name:          "Returns error when could not parse label value",
			TxFn:          txGen.ThatSucceeds,
			provisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", txtest.CtxWithDBMatcher(), testRuntime.ID, validHandlerConfig.ConsumerSubaccountIDsLabelKey).Return(&invalidTestLabel, nil).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name:          "Returns error when could not set label for runtime",
			TxFn:          txGen.ThatSucceeds,
			provisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", txtest.CtxWithDBMatcher(), testRuntime.ID, validHandlerConfig.ConsumerSubaccountIDsLabelKey).Return(&testLabel, nil).Once()
				provisioner.On("SetLabel", txtest.CtxWithDBMatcher(), &removeLabelInput).Return(testError).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name:          "Succeeds and creates label",
			TxFn:          txGen.ThatSucceeds,
			provisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", txtest.CtxWithDBMatcher(), testRuntime.ID, validHandlerConfig.ConsumerSubaccountIDsLabelKey).Return(nil, notFoundErr).Once()
				provisioner.On("SetLabel", txtest.CtxWithDBMatcher(), &emptyLabelInput).Return(nil).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: compassURL,
			ExpectedStatusCode:  http.StatusOK,
		},
		{
			Name:          "Succeeds and updates label",
			TxFn:          txGen.ThatSucceeds,
			provisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", txtest.CtxWithDBMatcher(), testRuntime.ID, validHandlerConfig.ConsumerSubaccountIDsLabelKey).Return(&testLabel, nil).Once()
				provisioner.On("SetLabel", txtest.CtxWithDBMatcher(), &removeLabelInput).Return(nil).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: compassURL,
			ExpectedStatusCode:  http.StatusOK,
		},
		{
			Name:          "Returns error when transaction commit fails",
			TxFn:          txGen.ThatFailsOnCommit,
			provisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			runtimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", txtest.CtxWithDBMatcher(), filters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", txtest.CtxWithDBMatcher(), testRuntime.ID, validHandlerConfig.ConsumerSubaccountIDsLabelKey).Return(&testLabel, nil).Once()
				provisioner.On("SetLabel", txtest.CtxWithDBMatcher(), &removeLabelInput).Return(nil).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: tenantfetchersvc.InternalServerError,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, transact := testCase.TxFn()
			provisioner := testCase.provisionerFn()
			runtimeService := testCase.runtimeServiceFn()
			defer mock.AssertExpectationsForObjects(t, transact, provisioner, runtimeService)

			handler := tenantfetchersvc.NewTenantsHTTPHandler(provisioner, runtimeService, transact, validHandlerConfig)
			req := testCase.Request

			if len(testCase.Region) > 0 {
				vars := map[string]string{
					"region": testCase.Region,
				}
				req = mux.SetURLVars(req, vars)
			}

			w := httptest.NewRecorder()

			//WHEN
			handler.UnSubscribeTenant(w, req)

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
			SubscriptionConsumerID: subscriptionConsumerID,
		})
		assert.NoError(t, err)

		_, transact := txGen.ThatDoesntStartTransaction()
		provisioner := &automock.TenantProvisioner{}
		defer mock.AssertExpectationsForObjects(t, transact, provisioner)

		handler := tenantfetchersvc.NewTenantsHTTPHandler(provisioner, &automock.RuntimeService{}, transact, validHandlerConfig)
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
		provisioner := &automock.TenantProvisioner{}
		defer mock.AssertExpectationsForObjects(t, transact, provisioner)

		handler := tenantfetchersvc.NewTenantsHTTPHandler(provisioner, &automock.RuntimeService{}, transact, validHandlerConfig)
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
			SubscriptionConsumerID: subscriptionConsumerID,
		})
		assert.NoError(t, err)

		_, transact := txGen.ThatDoesntStartTransaction()
		provisioner := &automock.TenantProvisioner{}
		defer mock.AssertExpectationsForObjects(t, transact, provisioner)

		handler := tenantfetchersvc.NewTenantsHTTPHandler(provisioner, &automock.RuntimeService{}, transact, validHandlerConfig)
		req := httptest.NewRequest(http.MethodDelete, target, bytes.NewBuffer(requestBody))
		w := httptest.NewRecorder()

		//WHEN
		handler.DeleteByExternalID(w, req)

		// THEN
		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
