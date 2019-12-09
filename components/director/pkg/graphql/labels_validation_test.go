package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/stretchr/testify/require"
)

func TestLabelInput_Validate_Key(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "valid",
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - Empty",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "Invalid - Too long",
			Value:         inputvalidationtest.String257Long,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidLabelInput()
			sut.Key = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestLabelInput_Validate_Value(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         interface{}
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "valid",
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid - Map with strings",
			Value:         map[string]string{"a": "b"},
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid - Slice of ints",
			Value:         []int{1, 2, 3},
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - Nil",
			Value:         nil,
			ExpectedValid: false,
		},
		{
			Name:          "Invalid - Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidLabelInput()
			sut.Value = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidLabelInput() graphql.LabelInput {
	return graphql.LabelInput{
		Key:   "valid",
		Value: "valid",
	}
}
