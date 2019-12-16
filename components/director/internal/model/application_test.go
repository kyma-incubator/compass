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
				ProviderDisplayName: providerName,
			},
			Expected: &model.Application{
				Name:                "Foo",
				ID:                  id,
				Tenant:              tenant,
				Description:         &desc,
				HealthCheckURL:      &url,
				IntegrationSystemID: &intSysID,
				ProviderDisplayName: providerName,
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
	filledAppUpdate := model.ApplicationUpdateInput{
		Name:                "",
		ProviderDisplayName: "provider name",
		Description:         str.Ptr("description"),
		HealthCheckURL:      str.Ptr("https://kyma-project.io"),
		IntegrationSystemID: str.Ptr("int sys id"),
	}

	testCases := []struct {
		Name      string
		AppUpdate model.ApplicationUpdateInput
	}{
		{
			Name:      "All properties filled",
			AppUpdate: filledAppUpdate,
		},
		{
			Name:      "Only needed properties",
			AppUpdate: model.ApplicationUpdateInput{Name: "name"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := model.Application{}

			//WHEN
			app.UpdateApplication(testCase.AppUpdate)

			//THEN
			assert.Equal(t, testCase.AppUpdate.Name, app.Name)
			assert.Equal(t, testCase.AppUpdate.Description, app.Description)
			assert.Equal(t, testCase.AppUpdate.HealthCheckURL, app.HealthCheckURL)
			assert.Equal(t, testCase.AppUpdate.IntegrationSystemID, app.IntegrationSystemID)
			assert.Equal(t, testCase.AppUpdate.ProviderDisplayName, app.ProviderDisplayName)
		})
	}
}
