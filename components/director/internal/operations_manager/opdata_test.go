package operationsmanager_test

import (
	"testing"

	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrdOperationData_GetData(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name          string
		AppID         string
		AppTemplateID string
		ExpectedData  string
		ExpectedErr   error
	}{
		{
			Name:          "Success",
			AppID:         "app-id",
			AppTemplateID: "app-template-id",
			ExpectedData:  "{\"applicationID\":\"app-id\",\"applicationTemplateID\":\"app-template-id\"}",
		},
		{
			Name:         "Success - missing application template id",
			AppID:        "app-id",
			ExpectedData: "{\"applicationID\":\"app-id\"}",
		},
		{
			Name:         "Success - missing application id",
			ExpectedData: "{\"applicationID\":\"\"}",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			opData := operationsmanager.NewOrdOperationData(testCase.AppID, testCase.AppTemplateID)

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
