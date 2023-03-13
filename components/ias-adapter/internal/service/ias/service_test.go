package ias

import (
	"testing"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIASService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IAS Service Test Suite")
}

var testConsumedAPI = types.ApplicationConsumedAPI{
	Name:     "name1",
	APIName:  "apiname1",
	AppID:    "appId1",
	ClientID: "clientId1",
}

var _ = Describe("Adding consumed API", func() {
	When("It doesn't exist in the current consumed APIs", func() {
		It("Should add it", func() {
			consumedAPIs := []types.ApplicationConsumedAPI{testConsumedAPI}
			newConsumedAPI := types.ApplicationConsumedAPI{
				Name:     "name2",
				APIName:  "apiname2",
				AppID:    "appId2",
				ClientID: "clientId2",
			}
			newConsumedAPIs := addConsumedAPI(consumedAPIs, newConsumedAPI)
			Expect(len(newConsumedAPIs)).To(Equal(2))
		})
	})
	When("It already exists in the current consumed APIs", func() {
		It("Shouldn't add it", func() {
			consumedAPIs := []types.ApplicationConsumedAPI{testConsumedAPI}
			newConsumedAPIs := addConsumedAPI(consumedAPIs, testConsumedAPI)
			Expect(len(newConsumedAPIs)).To(Equal(1))
		})
	})
})

var _ = Describe("Removing consumed API", func() {
	When("It doesn't exist in the current consumed APIs", func() {
		It("Shouldn't do anything", func() {
			consumedAPIs := []types.ApplicationConsumedAPI{testConsumedAPI}
			newConsumedAPIs := removeConsumedAPI(consumedAPIs, "non-existent-api-name")
			Expect(len(newConsumedAPIs)).To(Equal(1))
		})
	})
	When("It already exists in the current consumed APIs", func() {
		It("Should remove it", func() {
			consumedAPIs := []types.ApplicationConsumedAPI{testConsumedAPI}
			newConsumedAPIs := removeConsumedAPI(consumedAPIs, "apiname1")
			Expect(len(newConsumedAPIs)).To(Equal(0))
		})
	})
})
