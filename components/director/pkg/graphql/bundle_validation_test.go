package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

func TestBundleCreateInput_Validate_Name(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "name-123.com",
			ExpectedValid: true,
		},
		{
			Name:          "Valid Printable ASCII",
			Value:         "V1 +=_-)(*&^%$#@!?/>.<,|\\\"':;}{][",
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "String longer than 100 chars",
			Value:         inputvalidationtest.String129Long,
			ExpectedValid: false,
		},
		{
			Name:          "String contains invalid ASCII",
			Value:         "ąćńłóęǖǘǚǜ",
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidBundleCreateInput()
			obj.Name = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestBundleCreateInput_Validate_Description(t *testing.T) {
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
			obj := fixValidBundleCreateInput()
			obj.Description = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestBundleCreateInput_Validate_DefaultInstanceAuth(t *testing.T) {
	validObj := fixValidAuthInput()

	testCases := []struct {
		Name          string
		Value         *graphql.AuthInput
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid obj",
			Value:         &validObj,
			ExpectedValid: true,
		},
		{
			Name:          "Nil object",
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
			obj := fixValidBundleCreateInput()
			obj.DefaultInstanceAuth = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestBundleCreateInput_Validate_InstanceAuthRequestInputSchema(t *testing.T) {
	schema := graphql.JSONSchema("Test")
	emptySchema := graphql.JSONSchema("")
	testCases := []struct {
		Name          string
		Value         *graphql.JSONSchema
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         &schema,
			ExpectedValid: true,
		},
		{
			Name:          "Empty schema",
			Value:         &emptySchema,
			ExpectedValid: false,
		},
		{
			Name:          "Nil pointer",
			Value:         nil,
			ExpectedValid: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidBundleCreateInput()
			obj.InstanceAuthRequestInputSchema = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestBundleCreateInput_Validate_APIs(t *testing.T) {
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
			app := fixValidBundleCreateInput()
			app.APIDefinitions = testCase.Value
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

func TestBundleCreateInput_Validate_EventAPIs(t *testing.T) {
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
			app := fixValidBundleCreateInput()
			app.EventDefinitions = testCase.Value
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

func TestBundleCreateInput_Validate_Documents(t *testing.T) {
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
			app := fixValidBundleCreateInput()
			app.Documents = testCase.Value
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

func TestBundleUpdateInput_Validate_Name(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "name-123.com",
			ExpectedValid: true,
		},
		{
			Name:          "Valid Printable ASCII",
			Value:         "V1 +=_-)(*&^%$#@!?/>.<,|\\\"':;}{][",
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "String longer than 100 chars",
			Value:         inputvalidationtest.String129Long,
			ExpectedValid: false,
		},
		{
			Name:          "String contains invalid ASCII",
			Value:         "ąćńłóęǖǘǚǜ",
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidBundleUpdateInput()
			obj.Name = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestBundleUpdateInput_Validate_Description(t *testing.T) {
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
			obj := fixValidBundleUpdateInput()
			obj.Description = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestBundleUpdateInput_Validate_DefaultInstanceAuth(t *testing.T) {
	validObj := fixValidAuthInput()

	testCases := []struct {
		Name          string
		Value         *graphql.AuthInput
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid obj",
			Value:         &validObj,
			ExpectedValid: true,
		},
		{
			Name:          "Nil object",
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
			obj := fixValidBundleUpdateInput()
			obj.DefaultInstanceAuth = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestBundleUpdateInput_Validate_InstanceAuthRequestInputSchema(t *testing.T) {
	schema := graphql.JSONSchema("Test")
	emptySchema := graphql.JSONSchema("")
	testCases := []struct {
		Name          string
		Value         *graphql.JSONSchema
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         &schema,
			ExpectedValid: true,
		},
		{
			Name:          "Empty schema",
			Value:         &emptySchema,
			ExpectedValid: false,
		},
		{
			Name:          "Nil pointer",
			Value:         nil,
			ExpectedValid: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidBundleUpdateInput()
			obj.InstanceAuthRequestInputSchema = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestBundleInstanceAuthRequestInput_Validate(t *testing.T) {
	//GIVEN
	val := graphql.JSON("{\"foo\": \"bar\"}")
	testCases := []struct {
		Name          string
		Value         graphql.BundleInstanceAuthRequestInput
		ExpectedValid bool
	}{
		{
			Name:          "Empty",
			Value:         graphql.BundleInstanceAuthRequestInput{},
			ExpectedValid: true,
		},
		{
			Name: "InputParams and Context set",
			Value: graphql.BundleInstanceAuthRequestInput{
				Context:     &val,
				InputParams: &val,
			},
			ExpectedValid: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			err := testCase.Value.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestBundleInstanceAuthSetInput_Validate(t *testing.T) {
	//GIVEN
	authInput := fixValidAuthInput()
	str := "foo"
	testCases := []struct {
		Name          string
		Value         graphql.BundleInstanceAuthSetInput
		ExpectedValid bool
	}{
		{
			Name: "Auth",
			Value: graphql.BundleInstanceAuthSetInput{
				Auth: &authInput,
			},
			ExpectedValid: true,
		},
		{
			Name: "Failed Status",
			Value: graphql.BundleInstanceAuthSetInput{
				Status: &graphql.BundleInstanceAuthStatusInput{
					Condition: graphql.BundleInstanceAuthSetStatusConditionInputFailed,
					Reason:    str,
					Message:   str,
				},
			},
			ExpectedValid: true,
		},
		{
			Name: "Success Status",
			Value: graphql.BundleInstanceAuthSetInput{
				Status: &graphql.BundleInstanceAuthStatusInput{
					Condition: graphql.BundleInstanceAuthSetStatusConditionInputSucceeded,
				},
			},
			ExpectedValid: false,
		},
		{
			Name: "Auth and Success Status",
			Value: graphql.BundleInstanceAuthSetInput{
				Auth: &authInput,
				Status: &graphql.BundleInstanceAuthStatusInput{
					Condition: graphql.BundleInstanceAuthSetStatusConditionInputSucceeded,
					Message:   str,
					Reason:    str,
				},
			},
			ExpectedValid: true,
		},
		{
			Name: "Auth and Failure Status",
			Value: graphql.BundleInstanceAuthSetInput{
				Auth: &authInput,
				Status: &graphql.BundleInstanceAuthStatusInput{
					Condition: graphql.BundleInstanceAuthSetStatusConditionInputFailed,
				},
			},
			ExpectedValid: false,
		},
		{
			Name: "Empty objects",
			Value: graphql.BundleInstanceAuthSetInput{
				Auth:   &graphql.AuthInput{},
				Status: &graphql.BundleInstanceAuthStatusInput{},
			},
			ExpectedValid: false,
		},
		{
			Name:          "Empty",
			Value:         graphql.BundleInstanceAuthSetInput{},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			err := testCase.Value.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestBundleInstanceAuthStatusInput_Validate(t *testing.T) {
	//GIVEN
	str := "foo"
	testCases := []struct {
		Name          string
		Value         graphql.BundleInstanceAuthStatusInput
		ExpectedValid bool
	}{
		{
			Name: "Success",
			Value: graphql.BundleInstanceAuthStatusInput{
				Condition: graphql.BundleInstanceAuthSetStatusConditionInputSucceeded,
				Message:   str,
				Reason:    str,
			},
			ExpectedValid: true,
		},
		{
			Name: "No reason provided",
			Value: graphql.BundleInstanceAuthStatusInput{
				Condition: graphql.BundleInstanceAuthSetStatusConditionInputSucceeded,
				Message:   str,
			},
			ExpectedValid: false,
		},
		{
			Name: "No message provided",
			Value: graphql.BundleInstanceAuthStatusInput{
				Condition: graphql.BundleInstanceAuthSetStatusConditionInputSucceeded,
				Reason:    str,
			},
			ExpectedValid: false,
		},
		{
			Name: "No condition provided",
			Value: graphql.BundleInstanceAuthStatusInput{
				Message: str,
				Reason:  str,
			},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			err := testCase.Value.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidBundleCreateInput() graphql.BundleCreateInput {
	return graphql.BundleCreateInput{
		Name: inputvalidationtest.ValidName,
	}
}

func fixValidBundleUpdateInput() graphql.BundleUpdateInput {
	return graphql.BundleUpdateInput{
		Name: inputvalidationtest.ValidName,
	}
}
