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
	// GIVEN
	url := "https://foo.bar"
	desc := "Sample"
	id := "foo"
	intSysID := "bar"
	providerName := "provider name"
	timestamp := time.Now()
	statusCondition := model.ApplicationStatusConditionInitial
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
				StatusCondition:     &statusCondition,
				LocalTenantID:       str.Ptr("localTenantID"),
			},
			Expected: &model.Application{
				Name:                "Foo",
				Description:         &desc,
				HealthCheckURL:      &url,
				IntegrationSystemID: &intSysID,
				ProviderName:        &providerName,
				LocalTenantID:       str.Ptr("localTenantID"),
				Status: &model.ApplicationStatus{
					Timestamp: timestamp,
					Condition: model.ApplicationStatusConditionInitial,
				},
				BaseEntity: &model.BaseEntity{
					ID:    id,
					Ready: true,
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
			// WHEN
			result := testCase.Input.ToApplication(timestamp, id)

			// THEN
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestApplicationUpdateInput_UpdateApplication(t *testing.T) {
	const (
		providerName   = "provider name"
		description    = "description"
		healthCheckURL = "https://kyma-project.io"
		intSysID       = "int sys id"
	)
	t.Run("successfully overrides values with new input", func(t *testing.T) {
		// GIVEN
		timestamp := time.Now()
		statusCondition := model.ApplicationStatusConditionConnected
		filledAppUpdate := model.ApplicationUpdateInput{
			ProviderName:        str.Ptr(providerName),
			Description:         str.Ptr(description),
			HealthCheckURL:      str.Ptr(healthCheckURL),
			IntegrationSystemID: str.Ptr(intSysID),
			StatusCondition:     &statusCondition,
			LocalTenantID:       str.Ptr("localTenantID"),
		}
		app := model.Application{}

		// WHEN
		app.SetFromUpdateInput(filledAppUpdate, timestamp)

		// THEN
		assert.Equal(t, filledAppUpdate.Description, app.Description)
		assert.Equal(t, filledAppUpdate.HealthCheckURL, app.HealthCheckURL)
		assert.Equal(t, filledAppUpdate.IntegrationSystemID, app.IntegrationSystemID)
		assert.Equal(t, filledAppUpdate.ProviderName, app.ProviderName)
		assert.Equal(t, *filledAppUpdate.StatusCondition, app.Status.Condition)
		assert.Equal(t, filledAppUpdate.LocalTenantID, app.LocalTenantID)
	})

	t.Run("does not override values when input is missing", func(t *testing.T) {
		// GIVEN
		timestamp := time.Now()
		statusCondition := model.ApplicationStatusConditionConnected
		filledAppUpdate := model.ApplicationUpdateInput{
			StatusCondition: &statusCondition,
		}

		app := model.Application{
			ProviderName:        str.Ptr(providerName),
			Description:         str.Ptr(description),
			HealthCheckURL:      str.Ptr(healthCheckURL),
			IntegrationSystemID: str.Ptr(intSysID),
		}

		// WHEN
		app.SetFromUpdateInput(filledAppUpdate, timestamp)

		// THEN
		assert.Equal(t, description, *app.Description)
		assert.Equal(t, healthCheckURL, *app.HealthCheckURL)
		assert.Equal(t, intSysID, *app.IntegrationSystemID)
		assert.Equal(t, providerName, *app.ProviderName)
		assert.Equal(t, *filledAppUpdate.StatusCondition, app.Status.Condition)
	})
}
