package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

func TestEventAPIDefinitionInput_Validate_Name(t *testing.T) {
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
			obj := fixValidEventAPIDefinitionInput()
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

func TestEventAPIDefinitionInput_Validate_Description(t *testing.T) {
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
			Name:          "nil pointer",
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
			obj := fixValidEventAPIDefinitionInput()
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

func TestEventAPIDefinitionInput_Validate_Spec(t *testing.T) {
	validObj := fixValidEventAPISpecInput()
	emptyObj := graphql.EventSpecInput{}

	testCases := []struct {
		Name          string
		Value         *graphql.EventSpecInput
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
			Name:          "Invalid object",
			Value:         &emptyObj,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidEventAPIDefinitionInput()
			obj.Spec = testCase.Value
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

func TestEventAPIDefinitionInput_Validate_Group(t *testing.T) {
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
			Name:          "nil pointer",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         str.Ptr(inputvalidationtest.EmptyString),
			ExpectedValid: true,
		},
		{
			Name:          "String longer than 100 chars",
			Value:         str.Ptr(inputvalidationtest.String101Long),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidEventAPIDefinitionInput()
			obj.Group = testCase.Value
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

func TestEventAPIDefinitionInput_Validate_Version(t *testing.T) {
	validObj := fixValidVersionInput()
	emptyObj := graphql.VersionInput{}

	testCases := []struct {
		Name          string
		Value         *graphql.VersionInput
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
			Name:          "Invalid object",
			Value:         &emptyObj,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidEventAPIDefinitionInput()
			obj.Version = testCase.Value
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

func TestEventAPISpecInput_Validate_EventSpecType(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         graphql.EventSpecType
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         graphql.EventSpecTypeAsyncAPI,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid object",
			Value:         graphql.EventSpecType("INVALID"),
			ExpectedValid: false,
		},
		{
			Name:          "Invalid default value",
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidEventAPISpecInput()
			obj.Type = testCase.Value
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

func TestEventAPISpecInput_Validate_Format(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         graphql.SpecFormat
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid JSON",
			Value:         graphql.SpecFormatJSON,
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid YAML",
			Value:         graphql.SpecFormatYaml,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid object",
			Value:         graphql.SpecFormat("INVALID"),
			ExpectedValid: false,
		},
		{
			Name:          "Invalid default value",
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidEventAPISpecInput()
			obj.Format = testCase.Value
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

func TestEventAPISpecInput_Validate_FetchRequest(t *testing.T) {
	validObj := fixValidFetchRequestInput()
	emptyObj := graphql.FetchRequestInput{}

	testCases := []struct {
		Name          string
		Value         *graphql.FetchRequestInput
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
			ExpectedValid: false,
		},
		{
			Name:          "Invalid object",
			Value:         &emptyObj,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidEventAPISpecInput()
			obj.FetchRequest = testCase.Value
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

func fixValidEventAPISpecInput() graphql.EventSpecInput {
	req := fixValidFetchRequestInput()
	return graphql.EventSpecInput{
		FetchRequest: &req,
		Format:       graphql.SpecFormatJSON,
		Type:         graphql.EventSpecTypeAsyncAPI,
	}
}

func fixValidEventAPIDefinitionInput() graphql.EventDefinitionInput {
	eventSpec := fixValidEventAPISpecInput()
	return graphql.EventDefinitionInput{Name: "valid-name", Spec: &eventSpec}
}
