package model_test

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestRuntimeInput_ToRuntime(t *testing.T) {
	// given
	desc := "Sample"
	id := "foo"
	tenant := "sample"
	testCases := []struct {
		Name     string
		Input    *model.RuntimeInput
		Expected *model.Runtime
	}{
		{
			Name: "All properties given",
			Input: &model.RuntimeInput{
				Name:        "Foo",
				Description: &desc,
				Labels: map[string]interface{}{
					"test": []string{"val", "val2"},
				},
			},
			Expected: &model.Runtime{
				Name:        "Foo",
				ID:          id,
				Tenant:      tenant,
				Description: &desc,
				Status:      &model.RuntimeStatus{},
				AgentAuth:   &model.Auth{},
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
			// when
			result := testCase.Input.ToRuntime(id, tenant)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestRuntimeInput_ValidateInput(t *testing.T) {
	//GIVEN
	testError := errors.New("a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character")
	testCases := []struct {
		Name        string
		Input       model.RuntimeInput
		ExpectedErr error
	}{
		{
			Name:        "Correct Runtime name",
			Input:       model.RuntimeInput{Name: "correct-name.yeah"},
			ExpectedErr: nil,
		},
		{
			Name:        "Returns errors when Runtime name is empty",
			Input:       model.RuntimeInput{Name: ""},
			ExpectedErr: testError,
		},
		{
			Name:        "Returns errors when Runtime name contains UpperCase letter",
			Input:       model.RuntimeInput{Name: "Not-correct-name.yeah"},
			ExpectedErr: testError,
		},
		{
			Name:        "Returns errors when Runtime name contains special not allowed character",
			Input:       model.RuntimeInput{Name: "not-correct-n@me.yeah"},
			ExpectedErr: testError,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			//WHEN
			err := testCase.Input.Validate()

			//THEN
			if err != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.NoError(t, testCase.ExpectedErr)
			}
		})
	}
}
