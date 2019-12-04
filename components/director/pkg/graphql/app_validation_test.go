package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

func TestApplicationCreateInput_Validate_Name(t *testing.T) {
	testCases := []struct {
		Name  string
		Value string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: inputvalidationtest.ValidName,
			Valid: true,
		},
		{
			Name:  "Empty string",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
		},
		{
			Name:  "Invalid Upper Case Letters",
			Value: "Invalid",
			Valid: false,
		},
		{
			Name:  "String longer than 37 chars",
			Value: inputvalidationtest.String37Long,
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationCreateInput_Validate_Description(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: str.Ptr("this is a valid description"),
			Valid: true,
		},
		{
			Name:  "Nil pointer",
			Value: nil,
			Valid: true,
		},
		{
			Name:  "Empty string",
			Value: str.Ptr(inputvalidationtest.EmptyString),
			Valid: true,
		},
		{
			Name:  "String longer than 128 chars",
			Value: str.Ptr(inputvalidationtest.String129Long),
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationCreateInput_Validate_Labels(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *graphql.Labels
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: &graphql.Labels{"key": "value"},
			Valid: true,
		},
		{
			Name:  "Nil pointer",
			Value: nil,
			Valid: true,
		},
		{
			Name:  "Label with array of strings",
			Value: &graphql.Labels{"scenarios": []string{"ABC", "CBA", "TEST"}},
			Valid: true,
		},
		{
			Name:  "Invalid key",
			Value: &graphql.Labels{"": "value"},
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationCreateInput_Validate_HealthCheckURL(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: str.Ptr(inputvalidationtest.ValidURL),
			Valid: true,
		},
		{
			Name:  "Valid nil value",
			Value: nil,
			Valid: true,
		},
		{
			Name:  "URL longer than 256",
			Value: str.Ptr(inputvalidationtest.URL257Long),
			Valid: false,
		},
		{
			Name:  "Invalid",
			Value: str.Ptr(inputvalidationtest.InvalidURL),
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationCreateInput_Validate_Webhooks(t *testing.T) {
	validObj := fixValidWebhookInput()

	testCases := []struct {
		Name  string
		Value []*graphql.WebhookInput
		Valid bool
	}{
		{
			Name:  "Valid array",
			Value: []*graphql.WebhookInput{&validObj},
			Valid: true,
		},
		{
			Name:  "Empty array",
			Value: []*graphql.WebhookInput{},
			Valid: true,
		},
		//TODO: uncomment after implementation of webhook validation
		//{
		//	Name: "Array with invalid object",
		//	Value: []*model.WebhookInput{&model.WebhookInput{}},
		//	Valid:false,
		//},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := fixValidApplicationCreateInput()
			app.Webhooks = testCase.Value
			//WHEN
			err := app.Validate()
			//THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationCreateInput_Validate_APIs(t *testing.T) {
	validObj := fixValidAPIDefinitionInput()

	testCases := []struct {
		Name  string
		Value []*graphql.APIDefinitionInput
		Valid bool
	}{
		{
			Name:  "Valid array",
			Value: []*graphql.APIDefinitionInput{&validObj},
			Valid: true,
		},
		{
			Name:  "Empty array",
			Value: []*graphql.APIDefinitionInput{},
			Valid: true,
		},
		//TODO: uncomment after implementation of APIs validation
		//{
		//	Name: "Array with invalid object",
		//	Value: []*model.APIDefinitionInput{&model.APIDefinitionInput{}},
		//	Valid:false,
		//},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := fixValidApplicationCreateInput()
			app.Apis = testCase.Value
			//WHEN
			err := app.Validate()
			//THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationCreateInput_Validate_EventAPIs(t *testing.T) {
	validObj := fixValidEventAPIDefinitionInput()

	testCases := []struct {
		Name  string
		Value []*graphql.EventAPIDefinitionInput
		Valid bool
	}{
		{
			Name:  "Valid array",
			Value: []*graphql.EventAPIDefinitionInput{&validObj},
			Valid: true,
		},
		{
			Name:  "Empty array",
			Value: []*graphql.EventAPIDefinitionInput{},
			Valid: true,
		},
		//TODO: uncomment after implementation of eventAPIs validation
		//{
		//	Name: "Array with invalid object",
		//	Value: []*model.EventAPIDefinitionInput{&model.EventAPIDefinitionInput{}},
		//	Valid:false,
		//},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := fixValidApplicationCreateInput()
			app.EventAPIs = testCase.Value
			//WHEN
			err := app.Validate()
			//THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationCreateInput_Validate_Documents(t *testing.T) {
	validDoc := fixValidDocument()

	testCases := []struct {
		Name  string
		Value []*graphql.DocumentInput
		Valid bool
	}{
		{
			Name:  "Valid array",
			Value: []*graphql.DocumentInput{&validDoc},
			Valid: true,
		},
		{
			Name:  "Empty array",
			Value: []*graphql.DocumentInput{},
			Valid: true,
		},
		{
			Name:  "Array with invalid object",
			Value: []*graphql.DocumentInput{{}},
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationUpdateInput_Validate_Name(t *testing.T) {
	testCases := []struct {
		Name  string
		Value string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: inputvalidationtest.ValidName,
			Valid: true,
		},
		{
			Name:  "Empty string",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
		},
		{
			Name:  "Invalid Upper Case Letters",
			Value: inputvalidationtest.InvalidName,
			Valid: false,
		},
		{
			Name:  "String longer than 37 chars",
			Value: inputvalidationtest.String37Long,
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationUpdateInput_Validate_Description(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: str.Ptr(inputvalidationtest.ValidName),
			Valid: true,
		},
		{
			Name:  "Nil pointer",
			Value: nil,
			Valid: true,
		},
		{
			Name:  "Empty string",
			Value: str.Ptr(inputvalidationtest.EmptyString),
			Valid: true,
		},
		{
			Name:  "String longer than 128 chars",
			Value: str.Ptr(inputvalidationtest.String129Long),
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationUpdateInput_Validate_HealthCheckURL(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: str.Ptr(inputvalidationtest.ValidURL),
			Valid: true,
		},
		{
			Name:  "Valid nil value",
			Value: nil,
			Valid: true,
		},
		{
			Name:  "URL longer than 256",
			Value: str.Ptr(inputvalidationtest.URL257Long),
			Valid: false,
		},
		{
			Name:  "Invalid",
			Value: str.Ptr(inputvalidationtest.InvalidURL),
			Valid: false,
		},
		{
			Name:  "URL without protocol",
			Value: str.Ptr(inputvalidationtest.InvalidURL),
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidApplicationUpdateInput() graphql.ApplicationUpdateInput {
	return graphql.ApplicationUpdateInput{
		Name: "application",
	}
}

func fixValidApplicationCreateInput() graphql.ApplicationCreateInput {
	return graphql.ApplicationCreateInput{
		Name: "application",
	}
}

func fixValidAPIDefinitionInput() graphql.APIDefinitionInput {
	return graphql.APIDefinitionInput{}
}

func fixValidEventAPIDefinitionInput() graphql.EventAPIDefinitionInput {
	return graphql.EventAPIDefinitionInput{}
}
