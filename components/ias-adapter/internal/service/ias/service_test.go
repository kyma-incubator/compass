package ias

import (
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testConsumedAPI = types.ApplicationConsumedAPI{
	Name:    "name1",
	APIName: "apiname1",
	AppID:   "appId1",
}

var _ = Describe("Adding consumed API", func() {
	When("It doesn't exist in the current consumed APIs", func() {
		It("Should add it", func() {
			consumedAPIs := []types.ApplicationConsumedAPI{testConsumedAPI}
			newConsumedAPI := types.ApplicationConsumedAPI{
				Name:    "name2",
				APIName: "apiname2",
				AppID:   "appId2",
			}
			addConsumedAPI(&consumedAPIs, newConsumedAPI)
			Expect(len(consumedAPIs)).To(Equal(2))
		})
	})
	When("It already exists in the current consumed APIs", func() {
		It("Shouldn't add it", func() {
			consumedAPIs := []types.ApplicationConsumedAPI{testConsumedAPI}
			addConsumedAPI(&consumedAPIs, testConsumedAPI)
			Expect(len(consumedAPIs)).To(Equal(1))
		})
	})
})

var _ = Describe("Removing consumed API", func() {
	When("It doesn't exist in the current consumed APIs", func() {
		It("Shouldn't do anything", func() {
			consumedAPIs := []types.ApplicationConsumedAPI{testConsumedAPI}
			removeConsumedAPI(&consumedAPIs, "non-existent-api-name")
			Expect(len(consumedAPIs)).To(Equal(1))
		})
	})
	When("It already exists in the current consumed APIs", func() {
		It("Should remove it", func() {
			consumedAPIs := []types.ApplicationConsumedAPI{testConsumedAPI}
			removeConsumedAPI(&consumedAPIs, "apiname1")
			Expect(len(consumedAPIs)).To(Equal(0))
		})
	})
})

var _ = Describe("Normalize name", func() {
	When("API Name has invalid characters", func() {
		It("Should replace them with -", func() {
			result := normalizeIASDisplayName("api_name+with/invalid-characters%$@!")
			Expect(result).To(Equal("api_name-with-invalid-characters----"))
		})
	})
})
