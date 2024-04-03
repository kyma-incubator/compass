package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/internal/handlers/automock"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/service/ucl"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Tenant Mapping Handler", func() {
	var (
		tenantMapping  *types.TenantMapping
		errExpected    = errors.New("errExpected")
		mockService    *automock.TenantMappingsService
		mockUCLService *automock.UCLService
		handler        *TenantMappingsHandler

		expectError = func(w *httptest.ResponseRecorder, expectedCode int, expectedMessage string) {
			responseBody, err := io.ReadAll(w.Body)
			Expect(err).Error().ToNot(HaveOccurred())
			Expect(responseBody).To(ContainSubstring(url.QueryEscape(expectedMessage)))
			Expect(w.Code).To(Equal(expectedCode))
		}
		expectSuccess = func(w *httptest.ResponseRecorder, expectedMessage string) {
			responseBody, err := io.ReadAll(w.Body)
			Expect(err).Error().ToNot(HaveOccurred())
			Expect(responseBody).To(ContainSubstring(url.QueryEscape(expectedMessage)))
			Expect(w.Code).To(Equal(http.StatusOK))
		}
	)

	BeforeEach(func() {
		mockService = &automock.TenantMappingsService{}
		mockUCLService = &automock.UCLService{}
		tenantMapping = &types.TenantMapping{
			Context: types.Context{
				FormationID: "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
				Operation:   types.OperationAssign,
			},
			ReceiverTenant: types.ReceiverTenant{
				ApplicationURL: "localhost",
			},
			AssignedTenant: types.AssignedTenant{
				UCLApplicationID:   "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
				UCLApplicationType: "test-app-type",
				LocalTenantID:      "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
				Parameters: types.AssignedTenantParameters{
					ClientID: "clientID",
				},
				ReverseAssignmentState: "",
			},
		}
		handler = &TenantMappingsHandler{
			Service:    mockService,
			UCLService: mockUCLService,
		}
	})

	AfterEach(func() {
		mockService.AssertExpectations(test)
		mockUCLService.AssertExpectations(test)
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
			tenantMapping.AssignedTenant.Parameters.ClientID = ""
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
				tenantMapping.Operation = types.OperationUnassign
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
				expectSuccess(w, "")
			})
		})
	})

	When("Reverse assignment state is neither INITIAL nor READY", func() {
		BeforeEach(func() {
			tenantMapping.AssignedTenant.ReverseAssignmentState = "CREATE_ERROR"
		})

		It("Should return status 202 and report CONFIG_PENDING status", func() {
			expectedStatusReport := ucl.StatusReport{State: types.StateConfigPending}
			mockUCLService.On("ReportStatus", mock.Anything, mock.Anything, expectedStatusReport).Return(nil)
			w, ctx := createTestRequest(tenantMapping)

			handler.Patch(ctx)
			Expect(w.Code).To(Equal(http.StatusAccepted))
		})
	})

	When("Consumed APIs cannot be updated", func() {
		BeforeEach(func() {
			tenantMapping.AssignedTenant.ReverseAssignmentState = types.StateInitial
		})

		It("Should return status 202 and report CREATE_ERROR", func() {
			mockService.On("ProcessTenantMapping", mock.Anything, mock.Anything).Return(errExpected)
			expectedStatusReport := ucl.StatusReport{
				State: types.StateCreateError,
				Error: fmt.Sprintf("failed to process tenant mapping notification: %s", errExpected.Error()),
			}
			mockUCLService.On("ReportStatus", mock.Anything, mock.Anything, expectedStatusReport).Return(nil)
			w, ctx := createTestRequest(tenantMapping)

			handler.Patch(ctx)
			Expect(w.Code).To(Equal(http.StatusAccepted))
		})
	})
	When("Consumed APIs cannot be updated due to not found IAS application", func() {
		When("Operation is Assign", func() {
			BeforeEach(func() {
				tenantMapping.AssignedTenant.ReverseAssignmentState = types.StateInitial
				tenantMapping.Operation = types.OperationAssign
			})

			It("Should return status 202 and report CREATE_ERROR", func() {
				err := errors.Newf("could not process tenant mapping: %w", errors.IASApplicationNotFound)
				mockService.On("ProcessTenantMapping", mock.Anything, mock.Anything).Return(err)
				expectedStatusReport := ucl.StatusReport{
					State: types.StateCreateError,
					Error: fmt.Sprintf("failed to process tenant mapping notification: %s", err.Error()),
				}
				mockUCLService.On("ReportStatus", mock.Anything, mock.Anything, expectedStatusReport).Return(nil)
				w, ctx := createTestRequest(tenantMapping)

				handler.Patch(ctx)
				Expect(w.Code).To(Equal(http.StatusAccepted))
			})
		})
	})
	When("One of the participants is S/4 and there is no certificate provided", func() {
		BeforeEach(func() {
			tenantMapping.AssignedTenant.ReverseAssignmentState = types.StateInitial
			tenantMapping.AssignedTenant.UCLApplicationType = types.S4ApplicationType
			tenantMapping.AssignedTenant.Parameters.ClientID = ""
		})

		It("Should return 202 and report and CONFIG_PENDING with S/4 configuration", func() {
			mockService.On("ProcessTenantMapping", mock.Anything, mock.Anything).Return(errors.S4CertificateNotFound)
			w, ctx := createTestRequest(tenantMapping)
			expectedStatusReport := ucl.StatusReport{
				State: types.StateConfigPending,
				Configuration: &types.TenantMappingConfiguration{
					Credentials: types.Credentials{
						OutboundCommunicationCredentials: types.CommunicationCredentials{
							OAuth2mTLSAuthentication: types.OAuth2mTLSAuthentication{
								CorrelationIds: []string{S4SAPManagedCommunicationScenario},
							},
						},
					},
				},
			}
			mockUCLService.On("ReportStatus", mock.Anything, mock.Anything, expectedStatusReport).Return(nil)

			handler.Patch(ctx)
			Expect(w.Code).To(Equal(http.StatusAccepted))
		})
	})
	When("Consumed APIs are successfully updated", func() {
		BeforeEach(func() {
			tenantMapping.AssignedTenant.ReverseAssignmentState = types.StateReady
		})

		It("Should return 200", func() {
			mockService.On("ProcessTenantMapping", mock.Anything, mock.Anything).Return(nil)
			expectedStatusReport := ucl.StatusReport{
				State: types.StateCreateReady,
			}
			mockUCLService.On("ReportStatus", mock.Anything, mock.Anything, expectedStatusReport).Return(nil)
			w, ctx := createTestRequest(tenantMapping)

			handler.Patch(ctx)
			Expect(w.Code).To(Equal(http.StatusAccepted))
		})
	})
})
