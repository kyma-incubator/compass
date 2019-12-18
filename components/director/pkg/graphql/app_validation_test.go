package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

func TestApplicationRegisterInput_Validate_Name(t *testing.T) {
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
			Name:          "Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "Invalid Upper Case Letters",
			Value:         "Invalid",
			ExpectedValid: false,
		},
		{
			Name:          "String longer than 37 chars",
			Value:         inputvalidationtest.String37Long,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := fixValidApplicationCreateInput()
			app.Name = testCase.Value
			//WHEN
			err := app.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationRegisterInput_Validate_ProviderDisplayName(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "provider-name",
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "String longer than 128 chars",
			Value:         inputvalidationtest.String257Long,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := fixValidApplicationCreateInput()
			app.ProviderDisplayName = testCase.Value
			//WHEN
			err := app.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationRegisterInput_Validate_Description(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr("this is a valid description"),
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
			Name:          "String longer than 128 chars",
			Value:         str.Ptr(inputvalidationtest.String129Long),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := fixValidApplicationCreateInput()
			app.Description = testCase.Value
			//WHEN
			err := app.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationRegisterInput_Validate_Labels(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *graphql.Labels
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         &graphql.Labels{"key": "value"},
			ExpectedValid: true,
		},
		{
			Name:          "Nil pointer",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Label with array of strings",
			Value:         &graphql.Labels{"scenarios": []string{"ABC", "CBA", "TEST"}},
			ExpectedValid: true,
		},
		{
			Name:          "Invalid key",
			Value:         &graphql.Labels{"": "value"},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := fixValidApplicationCreateInput()
			app.Labels = testCase.Value
			//WHEN
			err := app.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationRegisterInput_Validate_HealthCheckURL(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr(inputvalidationtest.ValidURL),
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid nil value",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "URL longer than 256",
			Value:         str.Ptr(inputvalidationtest.URL257Long),
			ExpectedValid: false,
		},
		{
			Name:          "Invalid",
			Value:         str.Ptr(inputvalidationtest.InvalidURL),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := fixValidApplicationCreateInput()
			app.HealthCheckURL = testCase.Value
			//WHEN
			err := app.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationRegisterInput_Validate_Webhooks(t *testing.T) {
	validObj := fixValidWebhookInput()

	testCases := []struct {
		Name          string
		Value         []*graphql.WebhookInput
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid array",
			Value:         []*graphql.WebhookInput{&validObj},
			ExpectedValid: true,
		},
		{
			Name:          "Empty array",
			Value:         []*graphql.WebhookInput{},
			ExpectedValid: true,
		},
		{
			Name:          "Array with invalid object",
			Value:         []*graphql.WebhookInput{{}},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := fixValidApplicationCreateInput()
			app.Webhooks = testCase.Value
			//WHEN
			err := app.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationRegisterInput_Validate_APIs(t *testing.T) {
	validObj := fixValidAPIDefinitionInput()

	testCases := []struct {
		Name          string
		Value         []*graphql.APIDefinitionInput
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid array",
			Value:         []*graphql.APIDefinitionInput{&validObj},
			ExpectedValid: true,
		},
		{
			Name:          "Empty array",
			Value:         []*graphql.APIDefinitionInput{},
			ExpectedValid: true,
		},
		{
			Name:          "Array with invalid object",
			Value:         []*graphql.APIDefinitionInput{{}},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := fixValidApplicationCreateInput()
			app.APIDefinitions = testCase.Value
			//WHEN
			err := app.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationRegisterInput_Validate_EventAPIs(t *testing.T) {
	validObj := fixValidEventAPIDefinitionInput()

	testCases := []struct {
		Name          string
		Value         []*graphql.EventDefinitionInput
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid array",
			Value:         []*graphql.EventDefinitionInput{&validObj},
			ExpectedValid: true,
		},
		{
			Name:          "Empty array",
			Value:         []*graphql.EventDefinitionInput{},
			ExpectedValid: true,
		},
		{
			Name:          "Array with invalid object",
			Value:         []*graphql.EventDefinitionInput{{}},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := fixValidApplicationCreateInput()
			app.EventDefinitions = testCase.Value
			//WHEN
			err := app.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationRegisterInput_Validate_Documents(t *testing.T) {
	validDoc := fixValidDocument()

	testCases := []struct {
		Name          string
		Value         []*graphql.DocumentInput
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid array",
			Value:         []*graphql.DocumentInput{&validDoc},
			ExpectedValid: true,
		},
		{
			Name:          "Empty array",
			Value:         []*graphql.DocumentInput{},
			ExpectedValid: true,
		},
		{
			Name:          "Array with invalid object",
			Value:         []*graphql.DocumentInput{{}},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := fixValidApplicationCreateInput()
			app.Documents = testCase.Value
			//WHEN
			err := app.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationUpdateInput_Validate_Name(t *testing.T) {
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
			Name:          "Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "Invalid Upper Case Letters",
			Value:         inputvalidationtest.InvalidName,
			ExpectedValid: false,
		},
		{
			Name:          "String longer than 37 chars",
			Value:         inputvalidationtest.String37Long,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := fixValidApplicationUpdateInput()
			app.Name = testCase.Value
			//WHEN
			err := app.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationUpdateInput_Validate_ProviderDisplayName(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "provider-name",
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "String longer than 128 chars",
			Value:         inputvalidationtest.String257Long,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := fixValidApplicationCreateInput()
			app.ProviderDisplayName = testCase.Value
			//WHEN
			err := app.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationUpdateInput_Validate_Description(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr(inputvalidationtest.ValidName),
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
			Name:          "String longer than 128 chars",
			Value:         str.Ptr(inputvalidationtest.String129Long),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := fixValidApplicationUpdateInput()
			app.Description = testCase.Value
			//WHEN
			err := app.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationUpdateInput_Validate_HealthCheckURL(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr(inputvalidationtest.ValidURL),
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid nil value",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "URL longer than 256",
			Value:         str.Ptr(inputvalidationtest.URL257Long),
			ExpectedValid: false,
		},
		{
			Name:          "Invalid",
			Value:         str.Ptr(inputvalidationtest.InvalidURL),
			ExpectedValid: false,
		},
		{
			Name:          "URL without protocol",
			Value:         str.Ptr(inputvalidationtest.InvalidURL),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := fixValidApplicationUpdateInput()
			app.HealthCheckURL = testCase.Value
			//WHEN
			err := app.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidApplicationUpdateInput() graphql.ApplicationUpdateInput {
	return graphql.ApplicationUpdateInput{
		Name:                "application",
		ProviderDisplayName: "provider-name",
	}
}

func fixValidApplicationCreateInput() graphql.ApplicationRegisterInput {
	return graphql.ApplicationRegisterInput{
		Name:                "application",
		ProviderDisplayName: "provider-name",
	}
}
