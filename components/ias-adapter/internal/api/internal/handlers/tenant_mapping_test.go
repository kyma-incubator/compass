package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/internal/handlers/automock"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Tenant Mapping Handler", func() {
	var (
		tenantMapping *types.TenantMapping
		errExpected   = errors.New("errExpected")
		mockService   *automock.TenantMappingsService
		handler       *TenantMappingsHandler

		expectError = func(w *httptest.ResponseRecorder, expectedCode int, expectedMessage string) {
			responseBody, err := io.ReadAll(w.Body)
			Expect(err).Error().ToNot(HaveOccurred())
			Expect(responseBody).To(ContainSubstring(url.QueryEscape(expectedMessage)))
			Expect(w.Code).To(Equal(expectedCode))
		}
		expectSuccess = func(w *httptest.ResponseRecorder) {
			_, err := io.ReadAll(w.Body)
			Expect(err).Error().ToNot(HaveOccurred())
			Expect(w.Code).To(Equal(http.StatusOK))
		}
	)

	BeforeEach(func() {
		mockService = &automock.TenantMappingsService{}
		tenantMapping = &types.TenantMapping{
			FormationID: "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
			ReceiverTenant: types.ReceiverTenant{
				ApplicationURL: "localhost",
			},
			AssignedTenants: []types.AssignedTenant{
				{
					UCLApplicationID: "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
					UCLApplicationType: "test-app-type",
					LocalTenantID:    "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
					Operation:        types.OperationAssign,
					Parameters: types.AssignedTenantParameters{
						ClientID: "clientID",
					},
					ReverseAssignmentState: "",
				},
			},
		}
		handler = &TenantMappingsHandler{Service: mockService}
	})

	AfterEach(func() {
		mockService.AssertExpectations(test)
	})

	When("Tenant mapping cannot be decoded", func() {
		It("Should fail with 400", func() {
			w, ctx := createTestRequest("unprocessable body")
			handler.Patch(ctx)
			expectError(w, http.StatusBadRequest, "failed to decode tenant mapping body")
		})
	})
	When("Tenant mapping is invalid", func() {
		BeforeEach(func() {
			tenantMapping.AssignedTenants[0].Parameters.ClientID = ""
		})
		When("Operation is assign", func() {
			It("Should fail with 400", func() {
				w, ctx := createTestRequest(tenantMapping)
				handler.Patch(ctx)
				expectError(w, http.StatusBadRequest, "tenant mapping body is invalid")
			})
		})
		When("Operation is unassign", func() {
			BeforeEach(func() {
				tenantMapping.AssignedTenants[0].Operation = types.OperationUnassign
			})
			It("Should fail with 400 if tenantMappings are 2", func() {
				mockService.On("CanSafelyRemoveTenantMapping", mock.Anything, mock.Anything).Return(false, nil)
				w, ctx := createTestRequest(tenantMapping)
				handler.Patch(ctx)
				expectError(w, http.StatusBadRequest, "tenant mapping body is invalid")
			})
			It("Should fail with 500 if tenantMappings check fails", func() {
				mockService.On("CanSafelyRemoveTenantMapping", mock.Anything, mock.Anything).Return(false, errExpected)
				w, ctx := createTestRequest(tenantMapping)
				handler.Patch(ctx)
				expectError(w, http.StatusInternalServerError, errExpected.Error())
			})
			It("Should fail with 500 if tenantMappings remove call fails", func() {
				mockService.On("CanSafelyRemoveTenantMapping", mock.Anything, mock.Anything).Return(true, nil)
				mockService.On("RemoveTenantMapping", mock.Anything, mock.Anything).Return(errExpected)
				w, ctx := createTestRequest(tenantMapping)
				handler.Patch(ctx)
				expectError(w, http.StatusInternalServerError, errExpected.Error())
			})
			It("Should succeed if tenantMappings are less then 2", func() {
				mockService.On("CanSafelyRemoveTenantMapping", mock.Anything, mock.Anything).Return(true, nil)
				mockService.On("RemoveTenantMapping", mock.Anything, mock.Anything).Return(nil)
				w, ctx := createTestRequest(tenantMapping)
				handler.Patch(ctx)
				expectSuccess(w)
			})
		})
	})
	When("Reverse assignment state is neither INITIAL nor READY", func() {
		BeforeEach(func() {
			tenantMapping.AssignedTenants[0].ReverseAssignmentState = "CREATE_ERROR"
		})
		It("Should fail with 422 CONFIG_PENDING", func() {
			w, ctx := createTestRequest(tenantMapping)
			handler.Patch(ctx)
			Expect(w.Code).To(Equal(http.StatusUnprocessableEntity))
			expectError(w, http.StatusUnprocessableEntity, "")
		})
	})
	When("Consumed APIs cannot be updated", func() {
		BeforeEach(func() {
			tenantMapping.AssignedTenants[0].ReverseAssignmentState = types.StateInitial
		})
		It("Should fail with 500", func() {
			mockService.On("ProcessTenantMapping", mock.Anything, mock.Anything).Return(errExpected)
			w, ctx := createTestRequest(tenantMapping)
			handler.Patch(ctx)
			expectError(w, http.StatusInternalServerError, errExpected.Error())
		})
	})
	When("Consumed APIs cannot be updated due to not found IAS application", func() {
		When("Operation is Assign", func() {
			BeforeEach(func() {
				tenantMapping.AssignedTenants[0].ReverseAssignmentState = types.StateInitial
				tenantMapping.AssignedTenants[0].Operation = types.OperationAssign
			})
			It("Should return 404", func() {
				err := errors.Newf("could not process tenant mapping: %w", errors.IASApplicationNotFound)
				mockService.On("ProcessTenantMapping", mock.Anything, mock.Anything).Return(err)
				w, ctx := createTestRequest(tenantMapping)
				handler.Patch(ctx)
				expectError(w, http.StatusNotFound, err.Error())
			})
		})
	})
	When("Consumed APIs are successfully updated", func() {
		BeforeEach(func() {
			tenantMapping.AssignedTenants[0].ReverseAssignmentState = types.StateReady
		})
		It("Should return 200", func() {
			mockService.On("ProcessTenantMapping", mock.Anything, mock.Anything).Return(nil)
			w, ctx := createTestRequest(tenantMapping)
			handler.Patch(ctx)
			expectSuccess(w)
		})
	})
})
