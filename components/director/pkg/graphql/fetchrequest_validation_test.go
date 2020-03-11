package graphql_test

import (
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

func TestFetchRequestInput_Validate_URL(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         inputvalidationtest.ValidURL,
			ExpectedValid: true,
		},
		{
			Name:          "URL longer than 256",
			Value:         "https://kyma-project.io/" + strings.Repeat("a", 233),
			ExpectedValid: false,
		},
		{
			Name:          "Invalid",
			Value:         "kyma-project",
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			fr := fixValidFetchRequestInput()
			fr.URL = testCase.Value
			//WHEN
			err := fr.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestFetchRequestInput_Validate_Auth(t *testing.T) {
	validObj := fixValidAuthInput()
	testCases := []struct {
		Name          string
		Value         *graphql.AuthInput
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         &validObj,
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid nil value",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - Nested validation error",
			Value:         &graphql.AuthInput{Credential: &graphql.CredentialDataInput{}},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			fr := fixValidFetchRequestInput()
			fr.Auth = testCase.Value
			//WHEN
			err := fr.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestFetchRequestInput_Validate_Mode(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *graphql.FetchMode
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         (*graphql.FetchMode)(str.Ptr("SINGLE")),
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid nil value",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid object",
			Value:         (*graphql.FetchMode)(str.Ptr("INVALID")),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			fr := fixValidFetchRequestInput()
			fr.Mode = testCase.Value
			//WHEN
			err := fr.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestFetchRequestInput_Validate_Filter(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr("this is a valid string"),
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid nil pointer",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         str.Ptr(inputvalidationtest.EmptyString),
			ExpectedValid: false,
		},
		{
			Name:          "String bigger than 256 chars",
			Value:         str.Ptr(inputvalidationtest.String257Long),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			fr := fixValidFetchRequestInput()
			fr.Filter = testCase.Value
			//WHEN
			err := fr.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidFetchRequestInput() graphql.FetchRequestInput {
	return graphql.FetchRequestInput{
		URL: "https://kyma-project.io",
	}
}
