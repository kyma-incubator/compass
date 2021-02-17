package model_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestProductInput_ToProduct(t *testing.T) {
	// given
	id := "foo"
	appID := "bar"
	vendor := "Sample"
	name := "sample"
	tenant := "tenant"
	labels := json.RawMessage("{}")

	testCases := []struct {
		Name     string
		Input    *model.ProductInput
		Expected *model.Product
	}{
		{
			Name: "All properties given",
			Input: &model.ProductInput{
				OrdID:  id,
				Title:  name,
				Vendor: vendor,
				Labels: labels,
			},
			Expected: &model.Product{
				OrdID:         id,
				TenantID:      tenant,
				ApplicationID: appID,
				Title:         name,
				Vendor:        vendor,
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
			result := testCase.Input.ToProduct(tenant, appID)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
