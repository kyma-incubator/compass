package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

func TestRuntimeInput_Validate_Name(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         inputvalidationtest.ValidName,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - Empty",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "Invalid - Invalid Name",
			Value:         inputvalidationtest.InvalidName,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidRuntimeInput()
			sut.Name = testCase.Value
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

func TestRuntimeInput_Validate_Description(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name: "ExpectedValid",
			Value: str.Ptr("valid	valid"),
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid - Nil",
			Value:         (*string)(nil),
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid - Empty",
			Value:         str.Ptr(inputvalidationtest.EmptyString),
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - Too long",
			Value:         str.Ptr(inputvalidationtest.String2001Long),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidRuntimeInput()
			sut.Description = testCase.Value
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
func TestRuntimeInput_Validate_Labels(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *graphql.Labels
		ExpectedValid bool
	}{
		{
			Name: "ExpectedValid",
			Value: &graphql.Labels{
				"test": "ok",
			},
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid - Nil",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name: "ExpectedValid - Nil map value",
			Value: &graphql.Labels{
				"test": nil,
			},
			ExpectedValid: true,
		},
		{
			Name: "Invalid - Empty map key",
			Value: &graphql.Labels{
				inputvalidationtest.EmptyString: "val",
			},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidRuntimeInput()
			sut.Labels = testCase.Value
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

func fixValidRuntimeInput() graphql.RuntimeInput {
	return graphql.RuntimeInput{
		Name: inputvalidationtest.ValidName,
	}
}
