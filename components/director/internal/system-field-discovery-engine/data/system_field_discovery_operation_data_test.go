package data_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSystemFieldDiscoveryOperationData_GetData(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name         string
		AppID        string
		TenantID     string
		ExpectedData string
		ExpectedErr  error
	}{
		{
			Name:         "Success",
			AppID:        "app-id",
			TenantID:     "tenant-id",
			ExpectedData: "{\"applicationID\":\"app-id\",\"tenantID\":\"tenant-id\"}",
		},
		{
			Name:         "Success - missing tenant id",
			AppID:        "app-id",
			TenantID:     "",
			ExpectedData: "{\"applicationID\":\"app-id\",\"tenantID\":\"\"}",
		},
		{
			Name:         "Success - missing application id",
			TenantID:     "tenant-id",
			ExpectedData: "{\"applicationID\":\"\",\"tenantID\":\"tenant-id\"}",
		},
		{
			Name:         "Success - missing application id and tenant id",
			ExpectedData: "{\"applicationID\":\"\",\"tenantID\":\"\"}",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			data := data.NewSystemFieldDiscoveryOperationData(testCase.AppID, testCase.TenantID)

			// WHEN
			result, err := data.GetData()

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, testCase.ExpectedData, result)
			}
		})
	}
}
