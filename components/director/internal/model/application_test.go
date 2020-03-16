package model_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

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
	providerName := "provider name"
	timestamp := time.Now()
	testCases := []struct {
		Name     string
		Input    *model.ApplicationRegisterInput
		Expected *model.Application
	}{
		{
			Name: "All properties given",
			Input: &model.ApplicationRegisterInput{
				Name:        "Foo",
				Description: &desc,
				Labels: map[string]interface{}{
					"test": map[string]interface{}{
						"test": "foo",
					},
				},
				HealthCheckURL:      &url,
				IntegrationSystemID: &intSysID,
				ProviderName:        &providerName,
			},
			Expected: &model.Application{
				Name:                "Foo",
				ID:                  id,
				Tenant:              tenant,
				Description:         &desc,
				HealthCheckURL:      &url,
				IntegrationSystemID: &intSysID,
				ProviderName:        &providerName,
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

func TestApplicationUpdateInput_UpdateApplication(t *testing.T) {
	//GIVEN
	filledAppUpdate := model.ApplicationUpdateInput{
		ProviderName:        str.Ptr("provider name"),
		Description:         str.Ptr("description"),
		HealthCheckURL:      str.Ptr("https://kyma-project.io"),
		IntegrationSystemID: str.Ptr("int sys id"),
	}
	app := model.Application{}

	//WHEN
	app.SetFromUpdateInput(filledAppUpdate)

	//THEN
	assert.Equal(t, filledAppUpdate.Description, app.Description)
	assert.Equal(t, filledAppUpdate.HealthCheckURL, app.HealthCheckURL)
	assert.Equal(t, filledAppUpdate.IntegrationSystemID, app.IntegrationSystemID)
	assert.Equal(t, filledAppUpdate.ProviderName, app.ProviderName)
}
