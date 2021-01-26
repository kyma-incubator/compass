package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestBundleCreateInput_ToBundle(t *testing.T) {
	// given
	id := "foo"
	appID := "bar"
	desc := "Sample"
	name := "sample"
	tenant := "tenant"

	testCases := []struct {
		Name     string
		Input    *model.BundleCreateInput
		Expected *model.Bundle
	}{
		{

			Name: "All properties given",
			Input: &model.BundleCreateInput{
				Name:        name,
				Description: &desc,
			},
			Expected: &model.Bundle{
				ID:            id,
				TenantID:      tenant,
				ApplicationID: appID,
				Name:          name,
				Description:   &desc,
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
			result := testCase.Input.ToBundle(id, appID, tenant)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
