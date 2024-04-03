package service

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/service/automock"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tenant Mappings Service Test Suite")
}

var _ = Describe("Tenant mappings service", func() {
	var (
		ctx                   context.Context
		tenantMapping         types.TenantMapping
		tenantMappingsStorage *automock.TenantMappingsStorage
		iasService            *automock.IASService
	)

	BeforeEach(func() {
		ctx = context.Background()
		tenantMappingsStorage = &automock.TenantMappingsStorage{}
		iasService = &automock.IASService{}
		tenantMapping = types.TenantMapping{
			Context: types.Context{
				Operation:   types.OperationAssign,
				FormationID: "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
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
				Configuration: types.AssignedTenantConfiguration{
					ConsumedAPIs: []string{},
				},
				ReverseAssignmentState: "",
			},
		}
	})

	When("receive tenant mapping with empty ConsumedAPIs and the tenant mapping is stored in DB", func() {
		It("should not try to insert it again in the DB", func() {
			tenantMappingsInDB := map[string]types.TenantMapping{
				"2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05": tenantMapping,
				"11111111-10c4-4d6f-b4d4-5e1553e4ff05": tenantMapping,
			}
			tenantMappingsStorage.On("ListTenantMappings", ctx, mock.Anything).Return(tenantMappingsInDB, nil)
			tms := TenantMappingsService{Storage: tenantMappingsStorage, IASService: iasService}
			iasService.On("GetApplicationByClientID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(types.Application{}, errors.New("error"))
			Expect(tenantMappingsStorage.AssertNotCalled(GinkgoT(), "UpsertTenantMapping")).To(BeTrue())
			err := tms.ProcessTenantMapping(ctx, tenantMapping)
			Expect(err).Error().To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get IAS application"))
		})
	})

	When("tenant mapping with S/4 participant is received", func() {
		BeforeEach(func() {
			tenantMapping.AssignedTenant.UCLApplicationType = types.S4ApplicationType
			tenantMapping.AssignedTenant.Parameters.ClientID = ""
			tenantMappingsStorage.On("ListTenantMappings", ctx, mock.Anything).Return(map[string]types.TenantMapping{}, nil)
		})

		It("should return error when default S/4 certificate is not provided", func() {
			tms := TenantMappingsService{Storage: tenantMappingsStorage, IASService: iasService}
			err := tms.ProcessTenantMapping(ctx, tenantMapping)
			Expect(err).Error().To(MatchError(errors.S4CertificateNotFound))
		})

		It("should create application for S/4 in IAS if it doesn't exist", func() {
			iasAppID := "appId"
			tenantMapping.AssignedTenant.Configuration.Credentials.InboundCommunicationCredentials.OAuth2mTLSAuthentication.Certificate = "s4TestCert"

			expectedTenantMapping := tenantMapping
			expectedTenantMapping.AssignedTenant.Parameters.IASApplicationID = iasAppID
			tenantMappingsStorage.On("UpsertTenantMapping", ctx, expectedTenantMapping).Return(nil)

			iasService.On("GetApplicationByName", ctx, mock.Anything, mock.Anything).Return(types.Application{}, errors.IASApplicationNotFound)
			iasService.On("CreateApplication", ctx, mock.Anything, mock.Anything).Return(iasAppID, nil)

			tms := TenantMappingsService{Storage: tenantMappingsStorage, IASService: iasService}
			err := tms.ProcessTenantMapping(ctx, tenantMapping)
			Expect(err).Error().ToNot(HaveOccurred())
		})

		It("should get the application for S/4 in IAS if it exists", func() {
			iasAppID := "appId"
			tenantMapping.AssignedTenant.Configuration.Credentials.InboundCommunicationCredentials.OAuth2mTLSAuthentication.Certificate = "s4TestCert"

			expectedTenantMapping := tenantMapping
			expectedTenantMapping.AssignedTenant.Parameters.IASApplicationID = iasAppID
			tenantMappingsStorage.On("UpsertTenantMapping", ctx, expectedTenantMapping).Return(nil)

			iasService.On("GetApplicationByName", ctx, mock.Anything, mock.Anything).Return(types.Application{ID: iasAppID}, nil)

			tms := TenantMappingsService{Storage: tenantMappingsStorage, IASService: iasService}
			err := tms.ProcessTenantMapping(ctx, tenantMapping)
			Expect(err).Error().ToNot(HaveOccurred())
			Expect(tenantMappingsStorage.AssertNotCalled(GinkgoT(), "CreateApplication")).To(BeTrue())
		})
	})
})
