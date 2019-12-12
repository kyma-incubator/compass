package tenantmapping_test

import (
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapSystemAuthToConsumerType(t *testing.T) {
	//GIVEN
	testCases := []struct {
		name            string
		sysAuthRefInput model.SystemAuthReferenceObjectType
		expected        tenantmapping.ConsumerType
		expectedErr     error
	}{
		{
			name:            "Success - Map to application",
			sysAuthRefInput: model.ApplicationReference,
			expected:        tenantmapping.APPLICATION,
		},
		{
			name:            "Success - Map to runtime",
			sysAuthRefInput: model.RuntimeReference,
			expected:        tenantmapping.RUNTIME,
		},
		{
			name:            "Success - Map to integration system",
			sysAuthRefInput: model.IntegrationSystemReference,
			expected:        tenantmapping.INTEGRATION_SYSTEM,
		},
		{
			name:            "Error - Not exist reference",
			sysAuthRefInput: "Not Exist",
			expected:        "",
			expectedErr:     errors.New("unknown reference object type"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//WHEN
			consumerType, err := tenantmapping.MapSystemAuthToConsumerType(testCase.sysAuthRefInput)
			//THEN
			if err == nil {
				require.NoError(t, testCase.expectedErr)
				assert.Equal(t, consumerType, testCase.expected)
			} else {
				require.Error(t, testCase.expectedErr)
				assert.EqualError(t, testCase.expectedErr, err.Error())
			}
		})
	}
}
