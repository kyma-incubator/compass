package model_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestApplicationCreateInput_ToApplication(t *testing.T) {
	// given
	url := "https://foo.bar"
	desc := "Sample"
	id := "foo"
	tenant := "sample"
	intSysID := "bar"
	timestamp := time.Now()
	testCases := []struct {
		Name     string
		Input    *model.ApplicationCreateInput
		Expected *model.Application
	}{
		{
			Name: "All properties given",
			Input: &model.ApplicationCreateInput{
				Name:        "Foo",
				Description: &desc,
				Labels: map[string]interface{}{
					"test": map[string]interface{}{
						"test": "foo",
					},
				},
				HealthCheckURL: &url,
				IntegrationSystemID: &intSysID,
			},
			Expected: &model.Application{
				Name:           "Foo",
				ID:             id,
				Tenant:         tenant,
				Description:    &desc,
				HealthCheckURL: &url,
				IntegrationSystemID: &intSysID,
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

func TestApplicationCreateInput_ValidateInput(t *testing.T) {
	//GIVEN
	testError := errors.New("a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character")
	testCases := []struct {
		Name        string
		Input       model.ApplicationCreateInput
		ExpectedErr error
	}{
		{
			Name:        "Correct Application name",
			Input:       model.ApplicationCreateInput{Name: "correct-name.yeah"},
			ExpectedErr: nil,
		},
		{
			Name:        "Returns errors when Application name is empty",
			Input:       model.ApplicationCreateInput{Name: ""},
			ExpectedErr: testError,
		},
		{
			Name:        "Returns errors when Application name contains UpperCase letter",
			Input:       model.ApplicationCreateInput{Name: "Not-correct-name.yeah"},
			ExpectedErr: testError,
		},
		{
			Name:        "Returns errors when Application name contains special not allowed character",
			Input:       model.ApplicationCreateInput{Name: "not-correct-n@me.yeah"},
			ExpectedErr: testError,
		},
		{
			Name:        "Returns error when Application name is too long",
			Input:       model.ApplicationCreateInput{Name: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
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

func TestApplicationUpdateInput_ValidateInput(t *testing.T) {
	//GIVEN
	testError := errors.New("a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character")
	testCases := []struct {
		Name        string
		Input       model.ApplicationUpdateInput
		ExpectedErr error
	}{
		{
			Name:        "Correct Application name",
			Input:       model.ApplicationUpdateInput{Name: "correct-name.yeah"},
			ExpectedErr: nil,
		},
		{
			Name:        "Returns errors when Application name is empty",
			Input:       model.ApplicationUpdateInput{Name: ""},
			ExpectedErr: testError,
		},
		{
			Name:        "Returns errors when Application name contains UpperCase letter",
			Input:       model.ApplicationUpdateInput{Name: "Not-correct-name.yeah"},
			ExpectedErr: testError,
		},
		{
			Name:        "Returns errors when Application name contains special not allowed character",
			Input:       model.ApplicationUpdateInput{Name: "not-correct-n@me.yeah"},
			ExpectedErr: testError,
		},
		{
			Name:        "Returns error when Application name is too long",
			Input:       model.ApplicationUpdateInput{Name: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
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
