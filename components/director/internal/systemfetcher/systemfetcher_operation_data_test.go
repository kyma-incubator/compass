package systemfetcher_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemFetcherOperationData_GetData(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name         string
		TenantID     string
		ExpectedData string
		ExpectedErr  error
	}{
		{
			Name:         "Success",
			TenantID:     "tenant-id",
			ExpectedData: "{\"tenantID\":\"tenant-id\"}",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			opData := systemfetcher.NewSystemFetcherOperationData(testCase.TenantID)

			// WHEN
			result, err := opData.GetData()

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
