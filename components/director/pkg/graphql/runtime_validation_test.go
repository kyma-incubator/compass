package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

func TestRuntimeRegisterInput_Validate_Name(t *testing.T) {
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
			Name:          "Expected valid with digit",
			Value:         inputvalidationtest.ValidRuntimeNameWithDigit,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - Empty",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "Invalid - too long",
			Value:         inputvalidationtest.String257Long,
			ExpectedValid: false,
		},
		{
			Name:          "Invalid - invalid characters",
			Value:         inputvalidationtest.InValidRuntimeNameInvalidCharacters,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidRuntimeRegisterInput()
			sut.Name = testCase.Value
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

func TestRuntimeRegisterInput_Validate_Description(t *testing.T) {
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
			sut := fixValidRuntimeRegisterInput()
			sut.Description = testCase.Value
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

func TestRuntimeRegisterInput_Validate_Labels(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         graphql.Labels
		ExpectedValid bool
	}{
		{
			Name: "ExpectedValid",
			Value: graphql.Labels{
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
			Value: graphql.Labels{
				"test": nil,
			},
			ExpectedValid: true,
		},
		{
			Name: "Invalid - Empty map key",
			Value: graphql.Labels{
				inputvalidationtest.EmptyString: "val",
			},
			ExpectedValid: false,
		},
		{
			Name: "Invalid - Unsupported characters in key",
			Value: graphql.Labels{
				"not/valid": "val",
			},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidRuntimeRegisterInput()
			sut.Labels = testCase.Value
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

func TestRuntimeRegisterInput_Validate_Webhooks(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         graphql.WebhookType
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         graphql.WebhookTypeConfigurationChanged,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - Empty",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "Invalid - Not enum",
			Value:         "invalid",
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidRuntimeRegisterInput()
			webhook := fixValidWebhookInput(inputvalidationtest.ValidURL)
			sut.Webhooks = []*graphql.WebhookInput{&webhook}
			sut.Webhooks[0].Type = testCase.Value
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

func fixValidRuntimeRegisterInput() graphql.RuntimeRegisterInput {
	return graphql.RuntimeRegisterInput{
		Name: inputvalidationtest.ValidName,
	}
}

func TestRuntimeUpdateInput_Validate_Name(t *testing.T) {
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
			Name:          "Expected valid with digit",
			Value:         inputvalidationtest.ValidRuntimeNameWithDigit,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - Empty",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "Invalid - too long",
			Value:         inputvalidationtest.String257Long,
			ExpectedValid: false,
		},
		{
			Name:          "Invalid - invalid characters",
			Value:         inputvalidationtest.InValidRuntimeNameInvalidCharacters,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidRuntimeUpdateInput()
			sut.Name = testCase.Value
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

func TestRuntimeUpdateInput_Validate_Description(t *testing.T) {
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
			sut := fixValidRuntimeUpdateInput()
			sut.Description = testCase.Value
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

func TestRuntimeUpdateInput_Validate_Labels(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         graphql.Labels
		ExpectedValid bool
	}{
		{
			Name: "ExpectedValid",
			Value: graphql.Labels{
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
			Value: graphql.Labels{
				"test": nil,
			},
			ExpectedValid: true,
		},
		{
			Name: "Invalid - Empty map key",
			Value: graphql.Labels{
				inputvalidationtest.EmptyString: "val",
			},
			ExpectedValid: false,
		},
		{
			Name: "Invalid - Unsupported characters in key",
			Value: graphql.Labels{
				"not/valid": "val",
			},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidRuntimeUpdateInput()
			sut.Labels = testCase.Value
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

func fixValidRuntimeUpdateInput() graphql.RuntimeUpdateInput {
	return graphql.RuntimeUpdateInput{
		Name: inputvalidationtest.ValidName,
	}
}
