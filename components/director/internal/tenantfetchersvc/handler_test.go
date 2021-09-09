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

const (
	tenantExtID     = "tenant-external-id"
	tenantSubdomain = "mytenant"
	tenantRegion    = "myregion"

	subaccountTenantSubdomain = "myregionaltenant"
	subaccountTenantExtID     = "regional-tenant-external-id"

	parentTenantExtID = "parent-tenant-external-id"

	tenantProviderTenantIDProperty           = "tenantId"
	tenantProviderCustomerIDProperty         = "customerId"
	tenantProviderSubdomainProperty          = "subdomain"
	tenantProviderSubaccountTenantIDProperty = "subaccountTenantId"

	tenantCreationFailureMsgFmt = "Failed to create tenant with ID %s"
	compassURL                  = "https://github.com/kyma-incubator/compass"
)

var (
	testError          = errors.New("test error")
	validHandlerConfig = tenantfetchersvc.HandlerConfig{
		TenantProviderConfig: tenantfetchersvc.TenantProviderConfig{
			TenantProvider:     testProviderName,
			TenantIDProperty:   tenantProviderTenantIDProperty,
			CustomerIDProperty: tenantProviderCustomerIDProperty,
			SubdomainProperty:  tenantProviderSubdomainProperty,
		},
	}
)

type tenantCreationRequest struct {
	TenantID   string `json:"tenantId"`
	CustomerID string `json:"customerId"`
	Subdomain  string `json:"subdomain"`
}

type regionalTenantCreationRequest struct {
	TenantID  string `json:"subaccountTenantId"`
	ParentID  string `json:"tenantId"`
	Subdomain string `json:"subdomain"`
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
		TenantID:   tenantExtID,
		CustomerID: parentTenantExtID,
		Subdomain:  tenantSubdomain,
	})
	assert.NoError(t, err)

	bodyWithMissingTenant, err := json.Marshal(tenantCreationRequest{
		CustomerID: parentTenantExtID,
		Subdomain:  tenantSubdomain,
	})
	assert.NoError(t, err)

	bodyWithMissingParent, err := json.Marshal(tenantCreationRequest{
		TenantID:  tenantExtID,
		Subdomain: tenantSubdomain,
	})
	assert.NoError(t, err)

	bodyWithMissingTenantSubdomain, err := json.Marshal(tenantCreationRequest{
		TenantID:   tenantExtID,
		CustomerID: parentTenantExtID,
	})
	assert.NoError(t, err)

	accountProvisioningRequest := tenantfetchersvc.TenantProvisioningRequest{
		AccountTenantID:  tenantExtID,
		CustomerTenantID: parentTenantExtID,
		Subdomain:        tenantSubdomain,
	}
	accountWithoutParentProvisioningRequest := tenantfetchersvc.TenantProvisioningRequest{
		AccountTenantID: tenantExtID,
		Subdomain:       tenantSubdomain,
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
			Name:                "Returns error when reading request body fails",
			TxFn:                txGen.ThatDoesntStartTransaction,
			TenantProvisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			Request:             httptest.NewRequest(http.MethodPut, target, errReader(0)),
			ExpectedErrorOutput: "Failed to read tenant information from request body",
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
			Name:                "Returns error when beginning transaction fails",
			TxFn:                txGen.ThatFailsOnBegin,
			TenantProvisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedErrorOutput: fmt.Sprintf(tenantCreationFailureMsgFmt, accountProvisioningRequest.AccountTenantID),
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
			ExpectedErrorOutput: fmt.Sprintf(tenantCreationFailureMsgFmt, tenantExtID),
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
			ExpectedErrorOutput: fmt.Sprintf(tenantCreationFailureMsgFmt, accountProvisioningRequest.AccountTenantID),
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, transact := testCase.TxFn()
			provisioner := testCase.TenantProvisionerFn()
			defer mock.AssertExpectationsForObjects(t, transact, provisioner)

			handler := tenantfetchersvc.NewTenantsHTTPHandler(provisioner, transact, validHandlerConfig)
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

func TestService_CreateRegional(t *testing.T) {
	//GIVEN
	region := "eu-1"

	txGen := txtest.NewTransactionContextGenerator(errors.New("err"))
	target := "http://example.com/foo/:region"
	txtest.CtxWithDBMatcher()

	validRequestBody, err := json.Marshal(regionalTenantCreationRequest{
		TenantID:  subaccountTenantExtID,
		ParentID:  tenantExtID,
		Subdomain: subaccountTenantSubdomain,
	})
	assert.NoError(t, err)

	bodyWithMissingParent, err := json.Marshal(regionalTenantCreationRequest{
		TenantID:  subaccountTenantExtID,
		Subdomain: tenantSubdomain,
	})
	assert.NoError(t, err)

	bodyWithMissingTenantSubdomain, err := json.Marshal(regionalTenantCreationRequest{
		TenantID: subaccountTenantExtID,
		ParentID: tenantExtID,
	})
	assert.NoError(t, err)

	validHandlerConfig := tenantfetchersvc.HandlerConfig{
		RegionPathParam: "region",
		TenantProviderConfig: tenantfetchersvc.TenantProviderConfig{
			TenantProvider:             testProviderName,
			TenantIDProperty:           tenantProviderTenantIDProperty,
			SubaccountTenantIDProperty: tenantProviderSubaccountTenantIDProperty,
			CustomerIDProperty:         tenantProviderCustomerIDProperty,
			SubdomainProperty:          tenantProviderSubdomainProperty,
		},
	}
	regionalTenant := tenantfetchersvc.TenantProvisioningRequest{
		SubaccountTenantID: subaccountTenantExtID,
		AccountTenantID:    tenantExtID,
		Subdomain:          subaccountTenantSubdomain,
		Region:             region,
	}

	testCases := []struct {
		Name                  string
		provisionerFn         func() *automock.TenantProvisioner
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
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:                region,
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name:                "Returns error when region path parameter is missing",
			TxFn:                txGen.ThatDoesntStartTransaction,
			provisionerFn:       func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: "Region path parameter is missing from request",
		},
		{
			Name:                "Returns error when parent tenant is not found in body",
			TxFn:                txGen.ThatDoesntStartTransaction,
			provisionerFn:       func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingParent)),
			Region:              region,
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", tenantProviderTenantIDProperty),
		},
		{
			Name:                "Returns error when reading request body fails",
			TxFn:                txGen.ThatDoesntStartTransaction,
			provisionerFn:       func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			Request:             httptest.NewRequest(http.MethodPut, target, errReader(0)),
			Region:              region,
			ExpectedErrorOutput: "Failed to read tenant information from request body",
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name:                "Returns error when request body doesn't contain tenant subdomain",
			TxFn:                txGen.ThatDoesntStartTransaction,
			provisionerFn:       func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingTenantSubdomain)),
			Region:              region,
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", tenantProviderSubdomainProperty),
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			Name:                "Returns error when beginning transaction fails",
			TxFn:                txGen.ThatFailsOnBegin,
			provisionerFn:       func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: fmt.Sprintf(tenantCreationFailureMsgFmt, regionalTenant.SubaccountTenantID),
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
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: fmt.Sprintf(tenantCreationFailureMsgFmt, subaccountTenantExtID),
		},
		{
			Name: "Returns error when transaction commit fails",
			TxFn: txGen.ThatFailsOnCommit,
			provisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionRegionalTenants", txtest.CtxWithDBMatcher(), regionalTenant).Return(nil).Once()
				return provisioner
			},
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			Region:              region,
			ExpectedErrorOutput: fmt.Sprintf(tenantCreationFailureMsgFmt, regionalTenant.SubaccountTenantID),
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, transact := testCase.TxFn()
			provisioner := testCase.provisionerFn()
			defer mock.AssertExpectationsForObjects(t, transact, provisioner)

			handler := tenantfetchersvc.NewTenantsHTTPHandler(provisioner, transact, validHandlerConfig)
			req := testCase.Request

			if len(testCase.Region) > 0 {
				vars := map[string]string{
					"region": testCase.Region,
				}
				req = mux.SetURLVars(req, vars)
			}

			w := httptest.NewRecorder()

			//WHEN
			handler.CreateRegional(w, req)

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
			TenantID:   tenantExtID,
			CustomerID: parentTenantExtID,
			Subdomain:  tenantSubdomain,
		})
		assert.NoError(t, err)

		_, transact := txGen.ThatDoesntStartTransaction()
		provisioner := &automock.TenantProvisioner{}
		defer mock.AssertExpectationsForObjects(t, transact, provisioner)

		handler := tenantfetchersvc.NewTenantsHTTPHandler(provisioner, transact, validHandlerConfig)
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

		handler := tenantfetchersvc.NewTenantsHTTPHandler(provisioner, transact, validHandlerConfig)
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
			CustomerID: parentTenantExtID,
			Subdomain:  tenantSubdomain,
		})
		assert.NoError(t, err)

		_, transact := txGen.ThatDoesntStartTransaction()
		provisioner := &automock.TenantProvisioner{}
		defer mock.AssertExpectationsForObjects(t, transact, provisioner)

		handler := tenantfetchersvc.NewTenantsHTTPHandler(provisioner, transact, validHandlerConfig)
		req := httptest.NewRequest(http.MethodDelete, target, bytes.NewBuffer(requestBody))
		w := httptest.NewRecorder()

		//WHEN
		handler.DeleteByExternalID(w, req)

		// THEN
		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
