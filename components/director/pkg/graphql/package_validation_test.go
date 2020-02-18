package graphql_test

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPackageCreateInput_Validate(t *testing.T) {

}

func TestPackageUpdateInput_Validate(t *testing.T) {

}

func TestPackageInstanceAuthRequestInput_Validate(t *testing.T) {
	//GIVEN
	var val interface{} = map[string]interface{}{"foo": "bar"}
	testCases := []struct {
		Name          string
		Value         graphql.PackageInstanceAuthRequestInput
		ExpectedValid bool
	}{
		{
			Name:          "Empty",
			Value:         graphql.PackageInstanceAuthRequestInput{},
			ExpectedValid: true,
		},
		{
			Name: "InputParams and Context set",
			Value: graphql.PackageInstanceAuthRequestInput{
				Context:     &val,
				InputParams: &val,
			},
			ExpectedValid: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//WHEN
			err := testCase.Value.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestPackageInstanceAuthSetInput_Validate(t *testing.T) {
	//GIVEN
	authInput := fixValidAuthInput()
	str := "foo"
	testCases := []struct {
		Name          string
		Value         graphql.PackageInstanceAuthSetInput
		ExpectedValid bool
	}{
		{
			Name:          "Auth",
			Value:         graphql.PackageInstanceAuthSetInput{
				Auth: &authInput,
			},
			ExpectedValid: true,
		},
		{
			Name:          "Failed Status",
			Value:         graphql.PackageInstanceAuthSetInput{
				Status:&graphql.PackageInstanceAuthStatusInput{
					Condition: graphql.PackageInstanceAuthSetStatusConditionInputFailed,
				},
			},
			ExpectedValid: true,
		},
		{
			Name:          "Success Status",
			Value:         graphql.PackageInstanceAuthSetInput{
				Status:&graphql.PackageInstanceAuthStatusInput{
					Condition: graphql.PackageInstanceAuthSetStatusConditionInputSucceeded,
				},
			},
			ExpectedValid: true,
		},
		{
			Name:          "Auth and Success Status",
			Value:         graphql.PackageInstanceAuthSetInput{
				Auth: &authInput,
				Status:&graphql.PackageInstanceAuthStatusInput{
					Condition: graphql.PackageInstanceAuthSetStatusConditionInputSucceeded,
					Message:   &str,
					Reason:    &str,
				},
			},
			ExpectedValid: true,
		},
		{
			Name:          "Auth and Failure Status",
			Value:         graphql.PackageInstanceAuthSetInput{
				Auth: &authInput,
				Status:&graphql.PackageInstanceAuthStatusInput{
					Condition: graphql.PackageInstanceAuthSetStatusConditionInputFailed,
				},
			},
			ExpectedValid: false,
		},
		{
			Name:          "Empty objects",
			Value:         graphql.PackageInstanceAuthSetInput{
				Auth:   &graphql.AuthInput{},
				Status: &graphql.PackageInstanceAuthStatusInput{},
			},
			ExpectedValid: false,
		},
		{
			Name:          "Empty",
			Value:         graphql.PackageInstanceAuthSetInput{},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//WHEN
			err := testCase.Value.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
