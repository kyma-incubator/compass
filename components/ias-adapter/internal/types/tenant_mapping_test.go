package types_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTypes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Types Test Suite")
}

var _ = Describe("Tenant Mapping Type", func() {
	When("Tenant mapping has $.assignedTenant.configuration", func() {
		It("Should set the Configuration typed field", func() {
			tenantMapping := types.TenantMapping{
				Context: types.Context{
					FormationID: "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
					Operation:   types.OperationAssign,
				},
				ReceiverTenant: types.ReceiverTenant{
					ApplicationURL: "localhost",
				},
				AssignedTenant: types.AssignedTenant{
					AppID:         "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
					LocalTenantID: "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
					Parameters: types.AssignedTenantParameters{
						ClientID: "clientID",
					},
					Config: types.AssignedTenantConfiguration{
						ConsumedAPIs: []string{"qwe"},
					},
				},
			}
			Expect(tenantMapping.AssignedTenant.Configuration).To(Equal(types.AssignedTenantConfiguration{}))
			Expect(tenantMapping.AssignedTenant.SetConfiguration(context.Background())).To(Succeed())
			Expect(tenantMapping.AssignedTenant.Configuration).To(Equal(types.AssignedTenantConfiguration{ConsumedAPIs: []string{"qwe"}}))
		})
	})
})
