package tenantfetchersvc_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	tenantExtID       = "tenant-external-id"
	tenantSubdomain   = "mytenant"
	parentTenantExtID = "parent-tenant-external-id"
	parentTenantIntID = "27e23549-d369-4603-b3b8-e46883ab4a60"

	testProviderName                 = "test-provider"
	autogeneratedProviderName        = "autogenerated"
	tenantProviderTenantIdProperty   = "tenantId"
	tenantProviderCustomerIdProperty = "customerId"
	tenantProviderSubdomainProperty  = "subdomain"

	tenantCreationFailureMsgFmt = "Failed to create tenant with ID %s"
	compassURL                  = "https://github.com/kyma-incubator/compass"
)

var (
	testError          = errors.New("test error")
	validHandlerConfig = tenantfetchersvc.HandlerConfig{
		TenantProviderConfig: tenantfetchersvc.TenantProviderConfig{
			TenantProvider:     testProviderName,
			TenantIdProperty:   tenantProviderTenantIdProperty,
			CustomerIdProperty: tenantProviderCustomerIdProperty,
			SubdomainProperty:  tenantProviderSubdomainProperty,
		},
	}
)

type tenantCreationRequest struct {
	TenantID   string `json:"tenantId"`
	CustomerID string `json:"customerId"`
	Subdomain  string `json:"subdomain"`
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

	accountTenant := model.BusinessTenantMappingInput{
		Name:           tenantExtID,
		ExternalTenant: tenantExtID,
		Parent:         parentTenantExtID,
		Type:           tenantEntity.TypeToStr(tenantEntity.Account),
		Provider:       testProviderName,
		Subdomain:      tenantSubdomain,
	}
	accountTenantWithoutParent := model.BusinessTenantMappingInput{
		Name:           tenantExtID,
		ExternalTenant: tenantExtID,
		Type:           tenantEntity.TypeToStr(tenantEntity.Account),
		Provider:       testProviderName,
		Subdomain:      tenantSubdomain,
	}

	testCases := []struct {
		Name                  string
		TenantProvisionerFn   func() *automock.TenantProvisioner
		TxFn                  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		HandlerCfg            tenantfetchersvc.HandlerConfig
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
				provisioner.On("ProvisionTenant", txtest.CtxWithDBMatcher(), accountTenant).Return(nil).Once()
				return provisioner
			},
			HandlerCfg:            validHandlerConfig,
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name: "Succeeds when parent tenant is not found in body",
			TxFn: txGen.ThatSucceeds,
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenant", txtest.CtxWithDBMatcher(), accountTenantWithoutParent).Return(nil).Once()
				return provisioner
			},
			HandlerCfg:            validHandlerConfig,
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingParent)),
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name:                "Returns error when reading request body fails",
			TxFn:                txGen.ThatDoesntStartTransaction,
			TenantProvisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			HandlerCfg: tenantfetchersvc.HandlerConfig{
				TenantProviderConfig: tenantfetchersvc.TenantProviderConfig{
					TenantProvider:     testProviderName,
					TenantIdProperty:   tenantProviderTenantIdProperty,
					CustomerIdProperty: tenantProviderCustomerIdProperty,
				},
			},
			Request:             httptest.NewRequest(http.MethodPut, target, errReader(0)),
			ExpectedErrorOutput: "Failed to read tenant information from request body",
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name:                "Returns error when request body doesn't contain tenantID",
			TxFn:                txGen.ThatDoesntStartTransaction,
			TenantProvisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			HandlerCfg:          validHandlerConfig,
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingTenant)),
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", tenantProviderTenantIdProperty),
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			Name:                "Returns error when request body doesn't contain tenant subdomain",
			TxFn:                txGen.ThatDoesntStartTransaction,
			TenantProvisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			HandlerCfg:          validHandlerConfig,
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingTenantSubdomain)),
			ExpectedErrorOutput: fmt.Sprintf("mandatory property %q is missing from request body", tenantProviderSubdomainProperty),
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			Name:                "Returns error when beginning transaction fails",
			TxFn:                txGen.ThatFailsOnBegin,
			TenantProvisionerFn: func() *automock.TenantProvisioner { return &automock.TenantProvisioner{} },
			HandlerCfg:          validHandlerConfig,
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedErrorOutput: fmt.Sprintf(tenantCreationFailureMsgFmt, accountTenant.ExternalTenant),
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name: "Returns error when tenant creation fails",
			TxFn: txGen.ThatSucceeds,
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenant", txtest.CtxWithDBMatcher(), accountTenant).Return(testError).Once()
				return provisioner
			},
			HandlerCfg:          validHandlerConfig,
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: fmt.Sprintf(tenantCreationFailureMsgFmt, tenantExtID),
		},
		{
			Name: "Returns error when transaction commit fails",
			TxFn: txGen.ThatFailsOnCommit,
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenant", txtest.CtxWithDBMatcher(), accountTenant).Return(nil).Once()
				return provisioner
			},
			HandlerCfg:          validHandlerConfig,
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedErrorOutput: fmt.Sprintf(tenantCreationFailureMsgFmt, accountTenant.ExternalTenant),
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, transact := testCase.TxFn()
			provisioner := testCase.TenantProvisionerFn()
			defer mock.AssertExpectationsForObjects(t, transact, provisioner)

			handler := tenantfetchersvc.NewTenantsHTTPHandler(provisioner, transact, testCase.HandlerCfg)
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
