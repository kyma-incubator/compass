package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestTenantBusinessTypeInput_ToModel(t *testing.T) {
	// GIVEN
	id := "test-id"
	testCases := []struct {
		Name     string
		Input    *model.TenantBusinessTypeInput
		Expected *model.TenantBusinessType
	}{
		{
			Name: "All properties given",
			Input: &model.TenantBusinessTypeInput{
				Name: "test-name",
				Code: "test-code",
			},
			Expected: &model.TenantBusinessType{
				ID:   id,
				Name: "test-name",
				Code: "test-code",
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// WHEN
			result := testCase.Input.ToModel(id)

			// THEN
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
