package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

func TestVersionInput_Validate_Value(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "ExpectedValid",
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "String longer than 256 chars",
			Value:         inputvalidationtest.String257Long,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidVersionInput()
			obj.Value = testCase.Value
			//WHEN
			err := obj.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestVersionInput_Validate_Deprecated(t *testing.T) {
	boolean := true

	testCases := []struct {
		Name          string
		Value         *bool
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         &boolean,
			ExpectedValid: true,
		},
		{
			Name:          "Nil value",
			Value:         nil,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			doc := fixValidVersionInput()
			doc.Deprecated = testCase.Value
			//WHEN
			err := doc.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestVersionInput_Validate_DeprecatedSince(t *testing.T) {
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
			Name:          "Nil pointer",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         str.Ptr(inputvalidationtest.EmptyString),
			ExpectedValid: true,
		},
		{
			Name:          "String longer than 256 chars",
			Value:         str.Ptr(inputvalidationtest.String257Long),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidVersionInput()
			obj.DeprecatedSince = testCase.Value
			//WHEN
			err := obj.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestVersionInput_Validate_ForRemoval(t *testing.T) {
	boolean := true

	testCases := []struct {
		Name          string
		Value         *bool
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         &boolean,
			ExpectedValid: true,
		},
		{
			Name:          "Nil value",
			Value:         nil,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			doc := fixValidVersionInput()
			doc.ForRemoval = testCase.Value
			//WHEN
			err := doc.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidVersionInput() graphql.VersionInput {
	boolean := true
	return graphql.VersionInput{
		Value: "value", Deprecated: &boolean, ForRemoval: &boolean}
}
