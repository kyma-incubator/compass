package consumer_test

import (
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapSystemAuthToConsumerType(t *testing.T) {
	// GIVEN
	testCases := []struct {
		name            string
		sysAuthRefInput model.SystemAuthReferenceObjectType
		expected        consumer.ConsumerType
		expectedErr     error
	}{
		{
			name:            "Success - Map to application",
			sysAuthRefInput: model.ApplicationReference,
			expected:        consumer.Application,
		},
		{
			name:            "Success - Map to runtime",
			sysAuthRefInput: model.RuntimeReference,
			expected:        consumer.Runtime,
		},
		{
			name:            "Success - Map to integration system",
			sysAuthRefInput: model.IntegrationSystemReference,
			expected:        consumer.IntegrationSystem,
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
			// WHEN
			consumerType, err := consumer.MapSystemAuthToConsumerType(testCase.sysAuthRefInput)
			// THEN
			if err == nil {
				require.NoError(t, testCase.expectedErr)
				assert.Equal(t, consumerType, testCase.expected)
			} else {
				require.Error(t, testCase.expectedErr)
				assert.EqualError(t, err, apperrors.NewInternalError("unknown reference object type").Error())
			}
		})
	}
}
