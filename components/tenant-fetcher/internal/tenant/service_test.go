package tenant_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"

	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/model"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/tenant"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/tenant/automock"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	//GIVEN
	txGen := txtest.NewTransactionContextGenerator(testError)
	target := "http://example.com/foo"
	tenantModel := model.TenantModel{
		ID:             testID,
		TenantId:       testID,
		Status:         tenantEntity.Active,
		CustomerId:     customerID,
		Subdomain:      subdomain,
		TenantProvider: testProviderName,
	}
	tenantModelWithoutCustomerID := model.TenantModel{
		ID:             testID,
		TenantId:       testID,
		Status:         tenantEntity.Active,
		Subdomain:      subdomain,
		TenantProvider: testProviderName,
	}
	tenantModelWithoutTenantID := model.TenantModel{
		ID:             testID,
		Status:         tenantEntity.Active,
		CustomerId:     customerID,
		Subdomain:      subdomain,
		TenantProvider: testProviderName,
	}
	validRequestBody, err := json.Marshal(tenantModel)
	assert.NoError(t, err)
	validRequestBodyWithoutCustomerID, err := json.Marshal(tenantModelWithoutCustomerID)
	assert.NoError(t, err)
	validRequestBodyWithoutTenantID, err := json.Marshal(tenantModelWithoutTenantID)
	assert.NoError(t, err)
	invalidRequestBody, err := json.Marshal(model.TenantModel{})
	assert.NoError(t, err)

	testCases := []struct {
		Name                   string
		TenantRepoFn           func() *automock.TenantRepository
		TxFn                   func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		UidFn                  func() *automock.UIDService
		ConfigFn               func() tenant.Config
		Request                *http.Request
		ExpectedErrorOutput    error
		ExpectedSuccessOutput  string
		ExpectedStatusCode     int
		ExpectedTenantResponse model.TenantModel
	}{
		{
			Name: "Success with all properties provided",
			TxFn: txGen.ThatSucceeds,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.On("Create", txtest.CtxWithDBMatcher(), tenantModel).Return(nil).Once()
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(testID)
				return uidSvc
			},
			ConfigFn: func() tenant.Config {
				return tenant.Config{
					TenantProvider:                   testProviderName,
					TenantProviderTenantIdProperty:   tenantProviderTenantIdProperty,
					TenantProviderCustomerIdProperty: tenantProviderCustomerIdProperty,
					TenantProviderSubdomainProperty:  tenantProviderSubdomainProperty,
				}
			},
			Request:                httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedErrorOutput:    nil,
			ExpectedSuccessOutput:  "https://github.com/kyma-incubator/compass",
			ExpectedStatusCode:     200,
			ExpectedTenantResponse: tenantModel,
		}, {
			Name: "Success with missing customer id but with tenant id provided",
			TxFn: txGen.ThatSucceeds,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.On("Create", txtest.CtxWithDBMatcher(), tenantModelWithoutCustomerID).Return(nil).Once()
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(testID)
				return uidSvc
			},
			ConfigFn: func() tenant.Config {
				return tenant.Config{
					TenantProvider:                   testProviderName,
					TenantProviderTenantIdProperty:   tenantProviderTenantIdProperty,
					TenantProviderCustomerIdProperty: tenantProviderCustomerIdProperty,
					TenantProviderSubdomainProperty:  tenantProviderSubdomainProperty,
				}
			},
			Request:                httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBodyWithoutCustomerID)),
			ExpectedErrorOutput:    nil,
			ExpectedSuccessOutput:  "https://github.com/kyma-incubator/compass",
			ExpectedStatusCode:     200,
			ExpectedTenantResponse: tenantModelWithoutCustomerID,
		},
		{
			Name: "Success with missing tenant id but with customer id provided",
			TxFn: txGen.ThatSucceeds,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.On("Create", txtest.CtxWithDBMatcher(), tenantModelWithoutTenantID).Return(nil).Once()
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(testID)
				return uidSvc
			},
			ConfigFn: func() tenant.Config {
				return tenant.Config{
					TenantProvider:                   testProviderName,
					TenantProviderTenantIdProperty:   tenantProviderTenantIdProperty,
					TenantProviderCustomerIdProperty: tenantProviderCustomerIdProperty,
					TenantProviderSubdomainProperty:  tenantProviderSubdomainProperty,
				}
			},
			Request:                httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBodyWithoutTenantID)),
			ExpectedErrorOutput:    nil,
			ExpectedSuccessOutput:  "https://github.com/kyma-incubator/compass",
			ExpectedStatusCode:     200,
			ExpectedTenantResponse: tenantModelWithoutTenantID,
		},
		{
			Name: "Error when extracting request body",
			TxFn: txGen.ThatSucceeds,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.AssertNotCalled(t, "Create")
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.AssertNotCalled(t, "Generate")
				return uidSvc
			},
			ConfigFn: func() tenant.Config {
				return tenant.Config{
					TenantProvider:                   testProviderName,
					TenantProviderTenantIdProperty:   tenantProviderTenantIdProperty,
					TenantProviderCustomerIdProperty: tenantProviderCustomerIdProperty,
					TenantProviderSubdomainProperty:  tenantProviderSubdomainProperty,
				}
			},
			Request:                httptest.NewRequest(http.MethodPut, target, errReader(0)),
			ExpectedErrorOutput:    testError,
			ExpectedSuccessOutput:  "",
			ExpectedStatusCode:     500,
			ExpectedTenantResponse: tenantModel,
		},
		{
			Name: "Error when request body is invalid",
			TxFn: txGen.ThatSucceeds,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.AssertNotCalled(t, "Create")
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.AssertNotCalled(t, "Generate")
				return uidSvc
			},
			ConfigFn: func() tenant.Config {
				return tenant.Config{
					TenantProvider:                   testProviderName,
					TenantProviderTenantIdProperty:   tenantProviderTenantIdProperty,
					TenantProviderCustomerIdProperty: tenantProviderCustomerIdProperty,
					TenantProviderSubdomainProperty:  tenantProviderSubdomainProperty,
				}
			},
			Request:                httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(invalidRequestBody)),
			ExpectedErrorOutput:    fmt.Errorf("Both %q and %q not found in body or both are not valid strings", tenantProviderTenantIdProperty, tenantProviderCustomerIdProperty),
			ExpectedSuccessOutput:  "",
			ExpectedStatusCode:     500,
			ExpectedTenantResponse: tenantModel,
		},
		{
			Name: "Error when beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.AssertNotCalled(t, "Create")
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(testID)
				return uidSvc
			},
			ConfigFn: func() tenant.Config {
				return tenant.Config{
					TenantProvider:                   testProviderName,
					TenantProviderTenantIdProperty:   tenantProviderTenantIdProperty,
					TenantProviderCustomerIdProperty: tenantProviderCustomerIdProperty,
					TenantProviderSubdomainProperty:  tenantProviderSubdomainProperty,
				}
			},
			Request:                httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedErrorOutput:    testError,
			ExpectedSuccessOutput:  "",
			ExpectedStatusCode:     500,
			ExpectedTenantResponse: tenantModel,
		},
		{
			Name: "Error when creating tenant in database",
			TxFn: txGen.ThatSucceeds,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.On("Create", txtest.CtxWithDBMatcher(), tenantModel).Return(testError)
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(testID)
				return uidSvc
			},
			ConfigFn: func() tenant.Config {
				return tenant.Config{
					TenantProvider:                   testProviderName,
					TenantProviderTenantIdProperty:   tenantProviderTenantIdProperty,
					TenantProviderCustomerIdProperty: tenantProviderCustomerIdProperty,
					TenantProviderSubdomainProperty:  tenantProviderSubdomainProperty,
				}
			},
			Request:                httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedErrorOutput:    testError,
			ExpectedSuccessOutput:  "",
			ExpectedStatusCode:     500,
			ExpectedTenantResponse: tenantModel,
		},
		{
			Name: "Object Not Unique error when creating tenant in database should not fail",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.On("Create", txtest.CtxWithDBMatcher(), tenantModel).Return(apperrors.NewNotUniqueError(resource.Tenant))
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(testID)
				return uidSvc
			},
			ConfigFn: func() tenant.Config {
				return tenant.Config{
					TenantProvider:                   testProviderName,
					TenantProviderTenantIdProperty:   tenantProviderTenantIdProperty,
					TenantProviderCustomerIdProperty: tenantProviderCustomerIdProperty,
					TenantProviderSubdomainProperty:  tenantProviderSubdomainProperty,
				}
			},
			Request:                httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedSuccessOutput:  "https://github.com/kyma-incubator/compass",
			ExpectedStatusCode:     200,
			ExpectedTenantResponse: tenantModel,
		},
		{
			Name: "Error when committing transaction in database",
			TxFn: txGen.ThatFailsOnCommit,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.On("Create", txtest.CtxWithDBMatcher(), tenantModel).Return(nil)
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(testID)
				return uidSvc
			},
			ConfigFn: func() tenant.Config {
				return tenant.Config{
					TenantProvider:                   testProviderName,
					TenantProviderTenantIdProperty:   tenantProviderTenantIdProperty,
					TenantProviderCustomerIdProperty: tenantProviderCustomerIdProperty,
					TenantProviderSubdomainProperty:  tenantProviderSubdomainProperty,
				}
			},
			Request:                httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(validRequestBody)),
			ExpectedErrorOutput:    testError,
			ExpectedSuccessOutput:  "",
			ExpectedStatusCode:     500,
			ExpectedTenantResponse: tenantModel,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, transact := testCase.TxFn()
			tenantRepo := testCase.TenantRepoFn()
			uidSvc := testCase.UidFn()
			config := testCase.ConfigFn()

			body, err := json.Marshal(testCase.ExpectedTenantResponse)
			require.NoError(t, err)
			handler := tenant.NewService(tenantRepo, transact, uidSvc, config)
			req := testCase.Request
			w := httptest.NewRecorder()

			//WHEN
			handler.Create(w, req)

			// THEN
			resp := w.Result()
			body, err = ioutil.ReadAll(resp.Body)
			assert.NoError(t, err)

			if testCase.ExpectedErrorOutput != nil {
				assert.Contains(t, string(body), testCase.ExpectedErrorOutput.Error())
			} else {
				assert.NoError(t, err)
			}

			if testCase.ExpectedSuccessOutput != "" {
				assert.Equal(t, testCase.ExpectedSuccessOutput, string(body))
			}

			assert.Equal(t, testCase.ExpectedStatusCode, resp.StatusCode)

			tenantRepo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	//GIVEN
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)
	target := "http://example.com/foo"
	tenantModel := model.TenantModel{
		ID:             testID,
		TenantId:       testID,
		Status:         tenantEntity.Active,
		CustomerId:     customerID,
		Subdomain:      subdomain,
		TenantProvider: testProviderName,
	}
	requestBody, err := json.Marshal(tenantModel)
	assert.NoError(t, err)
	invalidRequestBody, err := json.Marshal(model.TenantModel{})
	assert.NoError(t, err)

	testCases := []struct {
		Name                  string
		TenantRepoFn          func() *automock.TenantRepository
		TxFn                  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		UidFn                 func() *automock.UIDService
		ConfigFn              func() tenant.Config
		Request               *http.Request
		ExpectedErrorOutput   error
		ExpectedSuccessOutput string
		ExpectedStatusCode    int
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.On("DeleteByExternalID", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				return uidSvc
			},
			ConfigFn: func() tenant.Config {
				return tenant.Config{
					TenantProvider:                   testProviderName,
					TenantProviderTenantIdProperty:   tenantProviderTenantIdProperty,
					TenantProviderCustomerIdProperty: tenantProviderCustomerIdProperty,
					TenantProviderSubdomainProperty:  tenantProviderSubdomainProperty,
				}
			},
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(requestBody)),
			ExpectedErrorOutput:   nil,
			ExpectedSuccessOutput: "",
			ExpectedStatusCode:    200,
		},
		{
			Name: "Error when extracting request body",
			TxFn: txGen.ThatSucceeds,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.AssertNotCalled(t, "DeleteByExternalID")
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				return uidSvc
			},
			ConfigFn: func() tenant.Config {
				return tenant.Config{
					TenantProvider:                   testProviderName,
					TenantProviderTenantIdProperty:   tenantProviderTenantIdProperty,
					TenantProviderCustomerIdProperty: tenantProviderCustomerIdProperty,
					TenantProviderSubdomainProperty:  tenantProviderSubdomainProperty,
				}
			},
			Request:               httptest.NewRequest(http.MethodPut, target, errReader(0)),
			ExpectedErrorOutput:   testErr,
			ExpectedSuccessOutput: "",
			ExpectedStatusCode:    500,
		},
		{
			Name: "Error when validating request body",
			TxFn: txGen.ThatSucceeds,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.AssertNotCalled(t, "DeleteByExternalID")
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				return uidSvc
			},
			ConfigFn: func() tenant.Config {
				return tenant.Config{
					TenantProvider:                   testProviderName,
					TenantProviderTenantIdProperty:   tenantProviderTenantIdProperty,
					TenantProviderCustomerIdProperty: tenantProviderCustomerIdProperty,
					TenantProviderSubdomainProperty:  tenantProviderSubdomainProperty,
				}
			},
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(invalidRequestBody)),
			ExpectedErrorOutput:   fmt.Errorf("Property %q not found in body or it is not of String type", tenantProviderTenantIdProperty),
			ExpectedSuccessOutput: "",
			ExpectedStatusCode:    500,
		},
		{
			Name: "Error when beginning transaction in database",
			TxFn: txGen.ThatFailsOnBegin,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.AssertNotCalled(t, "DeleteByExternalID")
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				return uidSvc
			},
			ConfigFn: func() tenant.Config {
				return tenant.Config{
					TenantProvider:                   testProviderName,
					TenantProviderTenantIdProperty:   tenantProviderTenantIdProperty,
					TenantProviderCustomerIdProperty: tenantProviderCustomerIdProperty,
					TenantProviderSubdomainProperty:  tenantProviderSubdomainProperty,
				}
			},
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(requestBody)),
			ExpectedErrorOutput:   testError,
			ExpectedSuccessOutput: "",
			ExpectedStatusCode:    500,
		},
		{
			Name: "Error when deleting tenant from database",
			TxFn: txGen.ThatSucceeds,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.On("DeleteByExternalID", txtest.CtxWithDBMatcher(), testID).Return(testErr)
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				return uidSvc
			},
			ConfigFn: func() tenant.Config {
				return tenant.Config{
					TenantProvider:                   testProviderName,
					TenantProviderTenantIdProperty:   tenantProviderTenantIdProperty,
					TenantProviderCustomerIdProperty: tenantProviderCustomerIdProperty,
					TenantProviderSubdomainProperty:  tenantProviderSubdomainProperty,
				}
			},
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(requestBody)),
			ExpectedErrorOutput:   testErr,
			ExpectedSuccessOutput: "",
			ExpectedStatusCode:    500,
		},
		{
			Name: "Object Not Found error when deleting tenant from database should not fail",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.On("DeleteByExternalID", txtest.CtxWithDBMatcher(), testID).Return(apperrors.NewNotFoundError(resource.Tenant, testID))
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				return uidSvc
			},
			ConfigFn: func() tenant.Config {
				return tenant.Config{
					TenantProvider:                   testProviderName,
					TenantProviderTenantIdProperty:   tenantProviderTenantIdProperty,
					TenantProviderCustomerIdProperty: tenantProviderCustomerIdProperty,
					TenantProviderSubdomainProperty:  tenantProviderSubdomainProperty,
				}
			},
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(requestBody)),
			ExpectedSuccessOutput: "",
			ExpectedStatusCode:    200,
		},
		{
			Name: "Error when committing transaction in database",
			TxFn: txGen.ThatFailsOnCommit,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.On("DeleteByExternalID", txtest.CtxWithDBMatcher(), testID).Return(nil)
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				return uidSvc
			},
			ConfigFn: func() tenant.Config {
				return tenant.Config{
					TenantProvider:                   testProviderName,
					TenantProviderTenantIdProperty:   tenantProviderTenantIdProperty,
					TenantProviderCustomerIdProperty: tenantProviderCustomerIdProperty,
					TenantProviderSubdomainProperty:  tenantProviderSubdomainProperty,
				}
			},
			Request:               httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(requestBody)),
			ExpectedErrorOutput:   testErr,
			ExpectedSuccessOutput: "",
			ExpectedStatusCode:    500,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, transact := testCase.TxFn()
			tenantRepo := testCase.TenantRepoFn()
			uidSvc := testCase.UidFn()
			config := testCase.ConfigFn()

			handler := tenant.NewService(tenantRepo, transact, uidSvc, config)
			req := testCase.Request
			w := httptest.NewRecorder()

			//WHEN
			handler.DeleteByExternalID(w, req)

			// THEN
			resp := w.Result()
			body, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err)

			if testCase.ExpectedErrorOutput != nil {
				assert.Contains(t, string(body), testCase.ExpectedErrorOutput.Error())
			} else {
				assert.NoError(t, err)
			}

			if testCase.ExpectedSuccessOutput != "" {
				assert.Equal(t, testCase.ExpectedSuccessOutput, string(body))
			}

			assert.Equal(t, testCase.ExpectedStatusCode, resp.StatusCode)

			tenantRepo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}

}
