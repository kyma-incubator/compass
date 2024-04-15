package processor

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/service/processor/automock"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/service/ucl"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var test *testing.T

func TestProcessor(t *testing.T) {
	test = t
	RegisterFailHandler(Fail)
	RunSpecs(t, "Processor Test Suite")
}

var _ = Describe("Processor", func() {
	var (
		ctx            = context.WithValue(context.Background(), locationHeader, "valid.url")
		tenantMapping  *types.TenantMapping
		errExpected    = errors.New("errExpected")
		mockTMService  *automock.TenantMappingsService
		mockUCLService *automock.UCLService
		asyncProcessor *AsyncProcessor
	)

	BeforeEach(func() {
		mockTMService = &automock.TenantMappingsService{}
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
				AppID:         "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
				AppNamespace:  "sap.test.namespace",
				LocalTenantID: "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
				Parameters: types.AssignedTenantParameters{
					ClientID: "clientID",
				},
				ReverseAssignmentState: "",
			},
		}
		asyncProcessor = &AsyncProcessor{
			TenantMappingsService: mockTMService,
			UCLService:            mockUCLService,
		}
	})

	AfterEach(func() {
		mockTMService.AssertExpectations(test)
		mockUCLService.AssertExpectations(test)
	})

	When("Reverse assignment state is neither INITIAL nor READY", func() {
		BeforeEach(func() {
			tenantMapping.AssignedTenant.ReverseAssignmentState = "CREATE_ERROR"
		})

		It("Should report CONFIG_PENDING status", func() {
			expectedStatusReport := ucl.StatusReport{State: types.StateConfigPending}
			mockUCLService.On("ReportStatus", mock.Anything, mock.Anything, expectedStatusReport).Return(nil)

			asyncProcessor.ProcessTMRequest(ctx, *tenantMapping)
		})
	})

	When("Consumed APIs cannot be updated", func() {
		BeforeEach(func() {
			tenantMapping.AssignedTenant.ReverseAssignmentState = types.StateInitial
		})

		It("Should report CREATE_ERROR status", func() {
			mockTMService.On("ProcessTenantMapping", mock.Anything, mock.Anything).Return(errExpected)
			expectedStatusReport := ucl.StatusReport{
				State: types.StateCreateError,
				Error: fmt.Sprintf("failed to process tenant mapping notification: %s", errExpected.Error()),
			}
			mockUCLService.On("ReportStatus", mock.Anything, mock.Anything, expectedStatusReport).Return(nil)

			asyncProcessor.ProcessTMRequest(ctx, *tenantMapping)
		})
	})

	When("Consumed APIs cannot be updated due to not found IAS application", func() {
		When("Operation is Assign", func() {
			BeforeEach(func() {
				tenantMapping.AssignedTenant.ReverseAssignmentState = types.StateInitial
				tenantMapping.Operation = types.OperationAssign
			})

			It("Should report CREATE_ERROR status", func() {
				err := errors.Newf("could not process tenant mapping: %w", errors.IASApplicationNotFound)
				mockTMService.On("ProcessTenantMapping", mock.Anything, mock.Anything).Return(err)
				expectedStatusReport := ucl.StatusReport{
					State: types.StateCreateError,
					Error: fmt.Sprintf("failed to process tenant mapping notification: %s", err.Error()),
				}
				mockUCLService.On("ReportStatus", mock.Anything, mock.Anything, expectedStatusReport).Return(nil)

				asyncProcessor.ProcessTMRequest(ctx, *tenantMapping)
			})
		})
	})

	When("One of the participants is S/4 and there is no certificate provided", func() {
		BeforeEach(func() {
			tenantMapping.AssignedTenant.ReverseAssignmentState = types.StateInitial
			tenantMapping.AssignedTenant.AppNamespace = types.S4ApplicationNamespace
			tenantMapping.AssignedTenant.Parameters.ClientID = ""
		})

		It("Should report CONFIG_PENDING status with S/4 configuration", func() {
			mockTMService.On("ProcessTenantMapping", mock.Anything, mock.Anything).Return(errors.S4CertificateNotFound)
			expectedStatusReport := ucl.StatusReport{
				State: types.StateConfigPending,
				Configuration: &types.TenantMappingConfiguration{
					Credentials: types.Credentials{
						OutboundCommunicationCredentials: types.CommunicationCredentials{
							OAuth2mTLSAuthentication: types.OAuth2mTLSAuthentication{
								CorrelationIds: []string{types.S4SAPManagedCommunicationScenario},
							},
						},
					},
				},
			}
			mockUCLService.On("ReportStatus", mock.Anything, mock.Anything, expectedStatusReport).Return(nil)

			asyncProcessor.ProcessTMRequest(ctx, *tenantMapping)
		})
	})

	When("Consumed APIs are successfully updated", func() {
		BeforeEach(func() {
			tenantMapping.AssignedTenant.ReverseAssignmentState = types.StateReady
		})

		It("Should report CREATE_READY status", func() {
			mockTMService.On("ProcessTenantMapping", mock.Anything, mock.Anything).Return(nil)
			expectedStatusReport := ucl.StatusReport{
				State: types.StateCreateReady,
			}
			mockUCLService.On("ReportStatus", mock.Anything, mock.Anything, expectedStatusReport).Return(nil)

			asyncProcessor.ProcessTMRequest(ctx, *tenantMapping)
		})
	})
})
