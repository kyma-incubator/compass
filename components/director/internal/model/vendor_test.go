package model_test

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVendorInput_ToVendor(t *testing.T) {
	// given
	id := "test"
	ordID := "foo"
	appID := "bar"
	partners := json.RawMessage(`["microsoft:vendor:Microsoft:"]`)
	name := "sample"
	tenant := "tenant"
	labels := json.RawMessage("{}")

	testCases := []struct {
		Name     string
		Input    *model.VendorInput
		Expected *model.Vendor
	}{
		{
			Name: "All properties given",
			Input: &model.VendorInput{
				OrdID:    ordID,
				Title:    name,
				Partners: partners,
				Labels:   labels,
			},
			Expected: &model.Vendor{
				ID:            id,
				OrdID:         ordID,
				TenantID:      tenant,
				ApplicationID: appID,
				Title:         name,
				Partners:      partners,
				Labels:        labels,
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {

			// when
			result := testCase.Input.ToVendor(id, tenant, appID)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
