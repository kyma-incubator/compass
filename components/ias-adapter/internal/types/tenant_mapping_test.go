package types

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTypes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Types Test Suite")
}

var _ = Describe("Tenant Mapping Type", func() {
	When("Tenant mapping has $.assignedTenants[0].configuration", func() {
		It("Should set the Configuration typed field", func() {
			tenantMapping := TenantMapping{
				FormationID: "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
				ReceiverTenant: ReceiverTenant{
					ApplicationURL: "localhost",
				},
				AssignedTenants: []AssignedTenant{
					{
						UCLApplicationID: "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
						LocalTenantID:    "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
						Operation:        OperationAssign,
						Parameters: AssignedTenantParameters{
							ClientID: "clientID",
						},
						Config: AssignedTenantConfiguration{
							ConsumedAPIs: []API{{
								APIName:      "qwe",
								AliasAPIName: "qwe",
							}},
						},
					},
				},
			}
			Expect(tenantMapping.AssignedTenants[0].Configuration).To(Equal(AssignedTenantConfiguration{}))
			Expect(tenantMapping.AssignedTenants[0].SetConfiguration(context.Background())).To(Succeed())
			Expect(tenantMapping.AssignedTenants[0].Configuration).To(Equal(AssignedTenantConfiguration{
				ConsumedAPIs: []API{
					{APIName: "qwe", AliasAPIName: "qwe"},
				},
			}))
		})
	})
})
