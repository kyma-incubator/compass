package tenantfetcher_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	tenantCreationFailureMsgFmt = "Failed to create tenant with ID %s"

	tenantID          = "17423414-7e86-491b-b087-dh1b787bb3b6"
	tenantExtID       = "tenant-external-id"
	tenantSubdomain   = "mytenant"
	parentTenantExtID = "parent-tenant-external-id"
	parentTenantIntID = "17422414-7e86-481b-b087-dc1b787ba3b6"

	testProviderName                 = "test-provider"
	autogeneratedProviderName        = "autogenerated"
	tenantProviderTenantIdProperty   = "tenantId"
	tenantProviderCustomerIdProperty = "customerId"
	tenantProviderSubdomainProperty  = "subdomain"

	parentTenantErrorMsgFormat = "Failed to ensure parent tenant with ID %s exists"
	compassURL                 = "https://github.com/kyma-incubator/compass"
)

var (
	testError = errors.New("test error")
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

	validHandlerConfig := tenantfetcher.HandlerConfig{
		TenantProvider:                   testProviderName,
		TenantProviderTenantIdProperty:   tenantProviderTenantIdProperty,
		TenantProviderCustomerIdProperty: tenantProviderCustomerIdProperty,
		TenantProviderSubdomainProperty:  tenantProviderSubdomainProperty,
	}
	customerTenant := model.BusinessTenantMapping{
		ID:             parentTenantIntID,
		Name:           parentTenantExtID,
		ExternalTenant: parentTenantExtID,
		Parent:         "",
		Type:           tenantEntity.Customer,
		Provider:       autogeneratedProviderName,
		Status:         tenantEntity.Active,
	}
	accountTenant := model.BusinessTenantMapping{
		ID:             tenantID,
		Name:           tenantExtID,
		ExternalTenant: tenantExtID,
		Parent:         parentTenantIntID,
		Type:           tenantEntity.Account,
		Status:         tenantEntity.Active,
		Provider:       testProviderName,
	}
	accountTenantWithoutParent := model.BusinessTenantMapping{
		ID:             tenantID,
		Name:           tenantExtID,
		ExternalTenant: tenantExtID,
		Type:           tenantEntity.Account,
		Status:         tenantEntity.Active,
		Provider:       testProviderName,
	}

	tenantLabelMatcher := mock.MatchedBy(func(label *model.LabelInput) bool {
		return label.ObjectID == tenantID && label.ObjectType == model.TenantLabelableObject && label.Value == tenantSubdomain && label.Key == "subdomain"
	})

	testCases := []struct {
		Name                  string
		TenantSvcFn           func() *automock.TenantService
		TxFn                  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		UidFn                 func() *automock.UIDService
		HandlerCfg            tenantfetcher.HandlerConfig
		Request               *http.Request
		ExpectedErrorOutput   string
		ExpectedSuccessOutput string
		ExpectedStatusCode    int
	}{
		{
			Name: "Succeeds when parent tenant already exists",
			TxFn: txGen.ThatSucceeds,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", mock.Anything, customerTenant.ExternalTenant).Return(parentTenantIntID, nil).Once()
				tenantSvc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), []model.BusinessTenantMapping{accountTenant}).Return(nil).Once()
				tenantSvc.On("SetLabel", txtest.CtxWithDBMatcher(), tenantLabelMatcher).Return(nil)
				return tenantSvc
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(accountTenant.ID).Once()
				return uidSvc
			},
			HandlerCfg:            validHandlerConfig,
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name: "Succeeds when parent tenant does not exist",
			TxFn: txGen.ThatSucceeds,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", mock.Anything, customerTenant.ExternalTenant).Return("", apperrors.NewNotFoundError(resource.Tenant, customerTenant.ExternalTenant)).Once()
				tenantSvc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), []model.BusinessTenantMapping{customerTenant}).Return(nil).Once()
				tenantSvc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), []model.BusinessTenantMapping{accountTenant}).Return(nil).Once()
				tenantSvc.On("SetLabel", txtest.CtxWithDBMatcher(), tenantLabelMatcher).Return(nil)
				return tenantSvc
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(accountTenant.ID).Once()
				uidSvc.On("Generate").Return(customerTenant.ID).Once()
				return uidSvc
			},
			HandlerCfg:            validHandlerConfig,
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name: "Succeeds when parent tenant is not found in body",
			TxFn: txGen.ThatSucceeds,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), []model.BusinessTenantMapping{accountTenantWithoutParent}).Return(nil).Once()
				tenantSvc.On("SetLabel", txtest.CtxWithDBMatcher(), tenantLabelMatcher).Return(nil)
				return tenantSvc
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(accountTenant.ID).Once()
				return uidSvc
			},
			HandlerCfg:            validHandlerConfig,
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingParent)),
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name: "Succeeds when parent tenant already exists when it tries to create it",
			TxFn: txGen.ThatSucceeds,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", mock.Anything, customerTenant.ExternalTenant).Return("", apperrors.NewNotFoundError(resource.Tenant, customerTenant.ExternalTenant)).Once()
				tenantSvc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), []model.BusinessTenantMapping{customerTenant}).Return(apperrors.NewNotUniqueError(resource.Tenant)).Once()
				tenantSvc.On("GetInternalTenant", mock.Anything, customerTenant.ExternalTenant).Return(customerTenant.ID, nil).Once()
				tenantSvc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), []model.BusinessTenantMapping{accountTenant}).Return(nil).Once()
				tenantSvc.On("SetLabel", txtest.CtxWithDBMatcher(), tenantLabelMatcher).Return(nil)
				return tenantSvc
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(accountTenant.ID).Once()
				uidSvc.On("Generate").Return(customerTenant.ID).Once()
				return uidSvc
			},
			HandlerCfg:            validHandlerConfig,
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedSuccessOutput: compassURL,
			ExpectedStatusCode:    http.StatusOK,
		},
		{
			Name: "Returns error when reading request body fails",
			TxFn: txGen.ThatDoesntStartTransaction,
			TenantSvcFn: func() *automock.TenantService {
				return &automock.TenantService{}
			},
			UidFn: func() *automock.UIDService {
				return &automock.UIDService{}
			},
			HandlerCfg: tenantfetcher.HandlerConfig{
				TenantProvider:                   testProviderName,
				TenantProviderTenantIdProperty:   tenantProviderTenantIdProperty,
				TenantProviderCustomerIdProperty: tenantProviderCustomerIdProperty,
			},
			Request:             httptest.NewRequest(http.MethodPut, target, errReader(0)),
			ExpectedErrorOutput: "Failed to read tenant information from request body",
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name: "Returns error when request body doesn't contain tenantID",
			TxFn: txGen.ThatDoesntStartTransaction,
			TenantSvcFn: func() *automock.TenantService {
				return &automock.TenantService{}
			},
			UidFn: func() *automock.UIDService {
				return &automock.UIDService{}
			},
			HandlerCfg:          validHandlerConfig,
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingTenant)),
			ExpectedErrorOutput: fmt.Sprintf("mandatory tenant ID property %q is missing from request body", tenantProviderTenantIdProperty),
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			Name: "Returns error when request body doesn't contain tenant subdomain",
			TxFn: txGen.ThatDoesntStartTransaction,
			TenantSvcFn: func() *automock.TenantService {
				return &automock.TenantService{}
			},
			UidFn: func() *automock.UIDService {
				return &automock.UIDService{}
			},
			HandlerCfg:          validHandlerConfig,
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyWithMissingTenantSubdomain)),
			ExpectedErrorOutput: fmt.Sprintf("mandatory subdomain property %q is missing from request body", tenantProviderSubdomainProperty),
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			Name: "Returns error when beginning transaction fails",
			TxFn: txGen.ThatFailsOnBegin,
			TenantSvcFn: func() *automock.TenantService {
				return &automock.TenantService{}
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(accountTenant.ID).Once()
				return uidSvc
			},
			HandlerCfg:          validHandlerConfig,
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedErrorOutput: fmt.Sprintf(tenantCreationFailureMsgFmt, accountTenant.ExternalTenant),
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name: "Returns error when creating parent tenant fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), customerTenant.ExternalTenant).Return("", apperrors.NewNotFoundError(resource.Tenant, customerTenant.ExternalTenant)).Once()
				tenantSvc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), []model.BusinessTenantMapping{customerTenant}).Return(testError).Once()
				return tenantSvc
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(accountTenant.ID).Once()
				uidSvc.On("Generate").Return(customerTenant.ID).Once()
				return uidSvc
			},
			HandlerCfg:            validHandlerConfig,
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedErrorOutput:   fmt.Sprintf(parentTenantErrorMsgFormat, parentTenantExtID),
			ExpectedSuccessOutput: "",
			ExpectedStatusCode:    http.StatusInternalServerError,
		},
		{
			Name: "Returns error when getting parent tenant from database fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), customerTenant.ExternalTenant).Return("", testError).Once()
				return tenantSvc
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(accountTenant.ID)
				return uidSvc
			},
			HandlerCfg:          validHandlerConfig,
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedErrorOutput: fmt.Sprintf(parentTenantErrorMsgFormat, parentTenantExtID),
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name: "Returns error when tenant creation fails",
			TxFn: txGen.ThatSucceeds,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", mock.Anything, customerTenant.ExternalTenant).Return("", apperrors.NewNotFoundError(resource.Tenant, customerTenant.ExternalTenant)).Once()
				tenantSvc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), []model.BusinessTenantMapping{customerTenant}).Return(apperrors.NewNotUniqueError(resource.Tenant)).Once()
				tenantSvc.On("GetInternalTenant", mock.Anything, customerTenant.ExternalTenant).Return(customerTenant.ID, nil).Once()
				tenantSvc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), []model.BusinessTenantMapping{accountTenant}).Return(testError).Once()
				return tenantSvc
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(accountTenant.ID).Once()
				uidSvc.On("Generate").Return(customerTenant.ID).Once()
				return uidSvc
			},
			HandlerCfg:          validHandlerConfig,
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: fmt.Sprintf(tenantCreationFailureMsgFmt, tenantExtID),
		},
		{
			Name: "Returns error when subdomain label creation fails",
			TxFn: txGen.ThatSucceeds,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", mock.Anything, customerTenant.ExternalTenant).Return("", apperrors.NewNotFoundError(resource.Tenant, customerTenant.ExternalTenant)).Once()
				tenantSvc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), []model.BusinessTenantMapping{customerTenant}).Return(apperrors.NewNotUniqueError(resource.Tenant)).Once()
				tenantSvc.On("GetInternalTenant", mock.Anything, customerTenant.ExternalTenant).Return(customerTenant.ID, nil).Once()
				tenantSvc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), []model.BusinessTenantMapping{accountTenant}).Return(nil).Once()
				tenantSvc.On("SetLabel", txtest.CtxWithDBMatcher(), tenantLabelMatcher).Return(testError)
				return tenantSvc
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(accountTenant.ID).Once()
				uidSvc.On("Generate").Return(customerTenant.ID).Once()
				return uidSvc
			},
			HandlerCfg:          validHandlerConfig,
			Request:             httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: fmt.Sprintf("Failed to add subdomain label to tenant with ID %s", tenantExtID),
		},
		{
			Name: "Error when committing transaction in database",
			TxFn: txGen.ThatFailsOnCommit,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), customerTenant.ExternalTenant).Return("", apperrors.NewNotFoundError(resource.Tenant, customerTenant.ExternalTenant)).Once()
				tenantSvc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), []model.BusinessTenantMapping{customerTenant}).Return(nil).Once()
				tenantSvc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), []model.BusinessTenantMapping{accountTenant}).Return(nil).Once()
				tenantSvc.On("SetLabel", txtest.CtxWithDBMatcher(), tenantLabelMatcher).Return(nil)
				return tenantSvc
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(accountTenant.ID).Once()
				uidSvc.On("Generate").Return(customerTenant.ID).Once()
				return uidSvc
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
			tenantSvc := testCase.TenantSvcFn()
			uidSvc := testCase.UidFn()
			defer mock.AssertExpectationsForObjects(t, transact, tenantSvc, uidSvc)

			body, err := json.Marshal(customerTenant)
			require.NoError(t, err)

			handler := tenantfetcher.NewTenantsHTTPHandler(tenantSvc, transact, uidSvc, testCase.HandlerCfg)
			req := testCase.Request
			w := httptest.NewRecorder()

			//WHEN
			handler.Create(w, req)

			// THEN
			resp := w.Result()
			body, err = ioutil.ReadAll(resp.Body)
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
	requestBody, err := json.Marshal(tenantCreationRequest{
		TenantID:   tenantExtID,
		CustomerID: parentTenantExtID,
		Subdomain:  tenantSubdomain,
	})
	assert.NoError(t, err)

	t.Run("DeleteByExternalID handler is noop", func(t *testing.T) {
		_, transact := txGen.ThatDoesntStartTransaction()
		tenantSvc := &automock.TenantService{}
		uidSvc := &automock.UIDService{}
		defer transact.AssertExpectations(t)
		defer tenantSvc.AssertExpectations(t)
		defer uidSvc.AssertExpectations(t)

		config := tenantfetcher.HandlerConfig{}

		handler := tenantfetcher.NewTenantsHTTPHandler(tenantSvc, transact, uidSvc, config)
		req := httptest.NewRequest(http.MethodDelete, target, bytes.NewBuffer(requestBody))
		w := httptest.NewRecorder()

		//WHEN
		handler.DeleteByExternalID(w, req)

		// THEN
		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
