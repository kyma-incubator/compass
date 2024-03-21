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
			FormationID: "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
			ReceiverTenant: types.ReceiverTenant{
				ApplicationURL: "localhost",
			},
			AssignedTenants: []types.AssignedTenant{
				{
					UCLApplicationID:   "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
					UCLApplicationType: "test-app-type",
					LocalTenantID:      "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
					Operation:          types.OperationAssign,
					Parameters: types.AssignedTenantParameters{
						ClientID: "clientID",
					},
					Configuration: types.AssignedTenantConfiguration{
						ConsumedAPIs: []string{},
					},
					ReverseAssignmentState: "",
				},
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
			iasService.On("GetApplication", ctx, mock.Anything, mock.Anything, mock.Anything).Return(types.Application{}, errors.New("error"))
			Expect(tenantMappingsStorage.AssertNotCalled(GinkgoT(), "UpsertTenantMapping")).To(BeTrue())
			err := tms.ProcessTenantMapping(ctx, tenantMapping)
			Expect(err).Error().To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get IAS application"))
		})
	})

	When("receive tenant mapping with S/4 participant", func() {
		BeforeEach(func() {
			tenantMapping.AssignedTenants[0].UCLApplicationType = types.S4ApplicationType
		})
		It("should return error when default S/4 certificate is not provided", func() {
			tenantMappingsStorage.On("ListTenantMappings", ctx, mock.Anything).Return(map[string]types.TenantMapping{}, nil)
			tms := TenantMappingsService{Storage: tenantMappingsStorage, IASService: iasService}
			err := tms.ProcessTenantMapping(ctx, tenantMapping)
			Expect(err).Error().To(MatchError(errors.S4CertificateNotFound))
		})
		It("should create application in IAS", func() {
			tenantMapping.AssignedTenants[0].Configuration.ApiCertificate = "s4TestCert"
			tenantMappingsStorage.On("ListTenantMappings", ctx, mock.Anything).Return(map[string]types.TenantMapping{}, nil)
			tenantMappingsStorage.On("UpsertTenantMapping", ctx, mock.Anything).Return(nil)
			iasService.On("CreateApplication", ctx, mock.Anything, mock.Anything).Return(nil)
			tms := TenantMappingsService{Storage: tenantMappingsStorage, IASService: iasService}
			err := tms.ProcessTenantMapping(ctx, tenantMapping)
			Expect(err).Error().ToNot(HaveOccurred())
		})
	})
})
