package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

func TestLabelSelectorInput_Validate_Key(t *testing.T) {
	testCases := []struct {
		Name          string
		Key           string
		ExpectedValid bool
	}{
		{
			Name:          "Valid",
			Key:           "global_subaccount_id",
			ExpectedValid: true,
		},
		{
			Name:          "Invalid",
			Key:           "key",
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			fr := fixValidLabelSelector()
			fr.Key = testCase.Key
			// WHEN
			err := fr.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidLabelSelector() graphql.LabelSelectorInput {
	return graphql.LabelSelectorInput{
		Key:   "global_subaccount_id",
		Value: "value",
	}
}
