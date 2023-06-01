package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

func TestApplicationInput_Validate_Name(t *testing.T) {
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
			Name:          "String longer than 100 chars",
			Value:         inputvalidationtest.String101Long,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app1 := fixValidApplicationRegisterInput()
			app1.Name = testCase.Value
			app2 := fixValidApplicationJSONInput()
			app2.Name = testCase.Value
			// WHEN
			err1 := app1.Validate()
			err2 := app2.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err1)
				require.NoError(t, err2)
			} else {
				require.Error(t, err1)
				require.Error(t, err2)
			}
		})
	}
}

func TestApplicationInput_Validate_ProviderName(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr("provider-name"),
			ExpectedValid: true,
		},
		{
			Name:          "Nil",
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
			app1 := fixValidApplicationRegisterInput()
			app1.ProviderName = testCase.Value
			app2 := fixValidApplicationJSONInput()
			app2.ProviderName = testCase.Value
			// WHEN
			err1 := app1.Validate()
			err2 := app2.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err1)
				require.NoError(t, err2)
			} else {
				require.Error(t, err1)
				require.Error(t, err2)
			}
		})
	}
}

func TestApplicationInput_Validate_Description(t *testing.T) {
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
			Name:          "String longer than 2000 chars",
			Value:         str.Ptr(inputvalidationtest.String2001Long),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app1 := fixValidApplicationRegisterInput()
			app1.Description = testCase.Value
			app2 := fixValidApplicationJSONInput()
			app2.Description = testCase.Value
			// WHEN
			err1 := app1.Validate()
			err2 := app2.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err1)
				require.NoError(t, err2)
			} else {
				require.Error(t, err1)
				require.Error(t, err2)
			}
		})
	}
}

func TestApplicationInput_Validate_Labels(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         graphql.Labels
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         graphql.Labels{"key": "value"},
			ExpectedValid: true,
		},
		{
			Name:          "Nil pointer",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Label with array of strings",
			Value:         graphql.Labels{"scenarios": []string{"ABC", "CBA", "TEST"}},
			ExpectedValid: true,
		},
		{
			Name:          "Empty key",
			Value:         graphql.Labels{"": "value"},
			ExpectedValid: false,
		},
		{
			Name:          "Invalid key",
			Value:         graphql.Labels{"not/valid": "value"},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app1 := fixValidApplicationRegisterInput()
			app1.Labels = testCase.Value
			app2 := fixValidApplicationJSONInput()
			app2.Labels = testCase.Value
			// WHEN
			err1 := app1.Validate()
			err2 := app2.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err1)
				require.NoError(t, err2)
			} else {
				require.Error(t, err1)
				require.Error(t, err2)
			}
		})
	}
}

func TestApplicationInput_Validate_HealthCheckURL(t *testing.T) {
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
			app1 := fixValidApplicationRegisterInput()
			app1.HealthCheckURL = testCase.Value
			app2 := fixValidApplicationJSONInput()
			app2.HealthCheckURL = testCase.Value
			// WHEN
			err1 := app1.Validate()
			err2 := app2.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err1)
				require.NoError(t, err2)
			} else {
				require.Error(t, err1)
				require.Error(t, err2)
			}
		})
	}
}

func TestApplicationInput_Validate_Webhooks(t *testing.T) {
	validObj := fixValidWebhookInput(inputvalidationtest.ValidURL)

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
			app1 := fixValidApplicationRegisterInput()
			app1.Webhooks = testCase.Value
			app2 := fixValidApplicationJSONInput()
			app2.Webhooks = testCase.Value
			// WHEN
			err1 := app1.Validate()
			err2 := app2.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err1)
				require.NoError(t, err2)
			} else {
				require.Error(t, err1)
				require.Error(t, err2)
			}
		})
	}
}

func TestApplicationUpdateInput_Validate_ProviderName(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr("provider-name"),
			ExpectedValid: true,
		},
		{
			Name:          "Nil",
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
			app1 := fixValidApplicationRegisterInput()
			app1.ProviderName = testCase.Value
			app2 := fixValidApplicationJSONInput()
			app2.ProviderName = testCase.Value
			// WHEN
			err1 := app1.Validate()
			err2 := app2.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err1)
				require.NoError(t, err2)
			} else {
				require.Error(t, err1)
				require.Error(t, err2)
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
			Name:          "String longer than 2000 chars",
			Value:         str.Ptr(inputvalidationtest.String2001Long),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := fixValidApplicationUpdateInput()
			app.Description = testCase.Value
			// WHEN
			err := app.Validate()
			// THEN
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
			// WHEN
			err := app.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidApplicationUpdateInput() graphql.ApplicationUpdateInput {
	return graphql.ApplicationUpdateInput{}
}

func fixValidApplicationRegisterInput() graphql.ApplicationRegisterInput {
	return graphql.ApplicationRegisterInput{
		Name: "application",
	}
}

func fixValidApplicationJSONInput() graphql.ApplicationRegisterInput {
	return graphql.ApplicationRegisterInput{
		Name: "application",
	}
}
