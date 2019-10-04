package model_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestApplicationInput_ToApplication(t *testing.T) {
	// given
	url := "https://foo.bar"
	desc := "Sample"
	id := "foo"
	tenant := "sample"
	timestamp := time.Now()
	testCases := []struct {
		Name     string
		Input    *model.ApplicationInput
		Expected *model.Application
	}{
		{
			Name: "All properties given",
			Input: &model.ApplicationInput{
				Name:        "Foo",
				Description: &desc,
				Labels: map[string]interface{}{
					"test": map[string]interface{}{
						"test": "foo",
					},
				},
				HealthCheckURL: &url,
			},
			Expected: &model.Application{
				Name:           "Foo",
				ID:             id,
				Tenant:         tenant,
				Description:    &desc,
				HealthCheckURL: &url,
				Status: &model.ApplicationStatus{
					Timestamp: timestamp,
					Condition: model.ApplicationStatusConditionUnknown,
				},
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
			result := testCase.Input.ToApplication(timestamp, model.ApplicationStatusConditionUnknown, id, tenant)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestApplicationInput_ValidateInput(t *testing.T) {
	//GIVEN
	testError := errors.New("a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character")
	testCases := []struct {
		Name        string
		Input       model.ApplicationInput
		ExpectedErr error
	}{
		{
			Name:        "Correct Application name",
			Input:       model.ApplicationInput{Name: "correct-name.yeah"},
			ExpectedErr: nil,
		},
		{
			Name:        "Returns errors when Application name is empty",
			Input:       model.ApplicationInput{Name: ""},
			ExpectedErr: testError,
		},
		{
			Name:        "Returns errors when Application name contains UpperCase letter",
			Input:       model.ApplicationInput{Name: "Not-correct-name.yeah"},
			ExpectedErr: testError,
		},
		{
			Name:        "Returns errors when Application name contains special not allowed character",
			Input:       model.ApplicationInput{Name: "not-correct-n@me.yeah"},
			ExpectedErr: testError,
		},
		{
			Name:        "Returns error when Application name is too long",
			Input:       model.ApplicationInput{Name: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			ExpectedErr: errors.New("application name is too long, must be maximum 36 characters long"),
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
