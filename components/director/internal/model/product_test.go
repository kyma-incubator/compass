package model_test

import (
	"encoding/json"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestProductInput_ToProduct(t *testing.T) {
	// given
	id := "test"
	ordID := "foo"
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
				OrdID:  ordID,
				Title:  name,
				Vendor: vendor,
				Labels: labels,
			},
			Expected: &model.Product{
				ID:            id,
				OrdID:         ordID,
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
		t.Run(testCase.Name, func(t *testing.T) {
			// when
			result := testCase.Input.ToProduct(id, tenant, appID)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
