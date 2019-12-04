package model_test

import (
	"fmt"
	"testing"
	"time"

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
				HealthCheckURL:      &url,
				IntegrationSystemID: &intSysID,
			},
			Expected: &model.Application{
				Name:                "Foo",
				ID:                  id,
				Tenant:              tenant,
				Description:         &desc,
				HealthCheckURL:      &url,
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
