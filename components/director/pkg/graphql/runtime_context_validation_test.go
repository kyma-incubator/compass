package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

func TestRuntimeContextInput_Validate_Key(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "tenant_id",
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - Empty",
			Value:         "",
			ExpectedValid: false,
		},
		{
			Name:          "Invalid - Invalid Name",
			Value:         "value/with-inv@lid-char$",
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidRuntimeContextInput()
			sut.Key = testCase.Value
			// WHEN
			err := sut.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidRuntimeContextInput() graphql.RuntimeContextInput {
	return graphql.RuntimeContextInput{
		Key: "test",
	}
}
