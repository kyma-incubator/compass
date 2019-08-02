package model_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

func TestToLabel(t *testing.T) {
	// given
	id := "foo"
	tenant := "sample"

	labelKey := "key"
	labelValue := "value"

	testCases := []struct {
		Name     string
		Input    *model.LabelInput
		Expected *model.Label
	}{
		{
			Name: "All properties given",
			Input: &model.LabelInput{
				Key:        labelKey,
				Value:      labelValue,
				ObjectID:   id,
				ObjectType: model.ApplicationLabelableObject,
			},
			Expected: &model.Label{
				ID:         id,
				Tenant:     tenant,
				Key:        labelKey,
				Value:      labelValue,
				ObjectID:   id,
				ObjectType: model.ApplicationLabelableObject,
			},
		},
		{
			Name:  "Empty",
			Input: &model.LabelInput{},
			Expected: &model.Label{
				ID:     id,
				Tenant: tenant,
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when
			result := testCase.Input.ToLabel(id, tenant)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
