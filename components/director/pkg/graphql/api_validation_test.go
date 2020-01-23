package graphql_test

import (
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

func TestAPIDefinitionInput_Validate_Name(t *testing.T) {
	testCases := []struct {
		Name        string
		Value       string
		ExpectValid bool
	}{
		{
			Name:        "ExpectedValid",
			Value:       "name-123.com",
			ExpectValid: true,
		},
		{
			Name:        "Valid Printable ASCII",
			Value:       "V1 +=_-)(*&^%$#@!?/>.<,|\\\"':;}{][",
			ExpectValid: true,
		},
		{
			Name:        "Empty string",
			Value:       inputvalidationtest.EmptyString,
			ExpectValid: false,
		},
		{
			Name:        "String longer than 100 chars",
			Value:       inputvalidationtest.String129Long,
			ExpectValid: false,
		},
		{
			Name:        "String contains invalid ASCII",
			Value:       "ąćńłóęǖǘǚǜ",
			ExpectValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidAPIDefinitionInput()
			obj.Name = testCase.Value
			//WHEN
			err := obj.Validate()
			//THEN
			if testCase.ExpectValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAPIDefinitionInput_Validate_Description(t *testing.T) {
	testCases := []struct {
		Name        string
		Value       *string
		ExpectValid bool
	}{
		{
			Name:        "ExpectedValid",
			Value:       str.Ptr("this is a valid description"),
			ExpectValid: true,
		},
		{
			Name:        "Nil pointer",
			Value:       nil,
			ExpectValid: true,
		},
		{
			Name:        "Empty string",
			Value:       str.Ptr(inputvalidationtest.EmptyString),
			ExpectValid: true,
		},
		{
			Name:        "String longer than 128 chars",
			Value:       str.Ptr(inputvalidationtest.String129Long),
			ExpectValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidAPIDefinitionInput()
			obj.Description = testCase.Value
			//WHEN
			err := obj.Validate()
			//THEN
			if testCase.ExpectValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAPIDefinitionInput_Validate_TargetURL(t *testing.T) {
	testCases := []struct {
		Name        string
		Value       string
		ExpectValid bool
	}{
		{
			Name:        "ExpectedValid",
			Value:       inputvalidationtest.ValidURL,
			ExpectValid: true,
		},
		{
			Name:        "URL longer than 256",
			Value:       "kyma-project.io/" + strings.Repeat("a", 241),
			ExpectValid: false,
		},
		{
			Name:        "Invalid, space in URL",
			Value:       "https://kyma test project.io",
			ExpectValid: false,
		},
		{
			Name:        "Invalid, no protocol",
			Value:       "kyma-project.io",
			ExpectValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			app := fixValidAPIDefinitionInput()
			app.TargetURL = testCase.Value
			//WHEN
			err := app.Validate()
			//THEN
			if testCase.ExpectValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAPIDefinitionInput_Validate_Group(t *testing.T) {
	testCases := []struct {
		Name        string
		Value       *string
		ExpectValid bool
	}{
		{
			Name:        "ExpectedValid",
			Value:       str.Ptr("this is a valid description"),
			ExpectValid: true,
		},
		{
			Name:        "Nil pointer",
			Value:       nil,
			ExpectValid: true,
		},
		{
			Name:        "Empty string",
			Value:       str.Ptr(inputvalidationtest.EmptyString),
			ExpectValid: true,
		},
		{
			Name:        "String longer than 36 chars",
			Value:       str.Ptr(inputvalidationtest.String37Long),
			ExpectValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidAPIDefinitionInput()
			obj.Group = testCase.Value
			//WHEN
			err := obj.Validate()
			//THEN
			if testCase.ExpectValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAPIDefinitionInput_Validate_APISpecInput(t *testing.T) {
	validObj := fixValidAPISpecInput()
	emptyObj := graphql.APISpecInput{}

	testCases := []struct {
		Name        string
		Value       *graphql.APISpecInput
		ExpectValid bool
	}{
		{
			Name:        "ExpectedValid obj",
			Value:       &validObj,
			ExpectValid: true,
		},
		{
			Name:        "Nil object",
			Value:       nil,
			ExpectValid: true,
		},
		{
			Name:        "Invalid object",
			Value:       &emptyObj,
			ExpectValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidAPIDefinitionInput()
			obj.Spec = testCase.Value
			//WHEN
			err := obj.Validate()
			//THEN
			if testCase.ExpectValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAPIDefinitionInput_Validate_Version(t *testing.T) {
	validObj := fixValidVersionInput()
	emptyObj := graphql.VersionInput{}

	testCases := []struct {
		Name        string
		Value       *graphql.VersionInput
		ExpectValid bool
	}{
		{
			Name:        "ExpectedValid obj",
			Value:       &validObj,
			ExpectValid: true,
		},
		{
			Name:        "Nil object",
			Value:       nil,
			ExpectValid: true,
		},
		{
			Name:        "Invalid object",
			Value:       &emptyObj,
			ExpectValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidAPIDefinitionInput()
			obj.Version = testCase.Value
			//WHEN
			err := obj.Validate()
			//THEN
			if testCase.ExpectValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAPIDefinitionInput_Validate_DefaultAuth(t *testing.T) {
	validObj := fixValidAuthInput()
	emptyObj := graphql.AuthInput{}

	testCases := []struct {
		Name        string
		Value       *graphql.AuthInput
		ExpectValid bool
	}{
		{
			Name:        "ExpectedValid obj",
			Value:       &validObj,
			ExpectValid: true,
		},
		{
			Name:        "Nil object",
			Value:       nil,
			ExpectValid: true,
		},
		{
			Name:        "Invalid object",
			Value:       &emptyObj,
			ExpectValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidAPIDefinitionInput()
			obj.DefaultAuth = testCase.Value
			//WHEN
			err := obj.Validate()
			//THEN
			if testCase.ExpectValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAPISpecInput_Validate_Type(t *testing.T) {
	testCases := []struct {
		Name        string
		Value       graphql.APISpecType
		ExpectValid bool
	}{
		{
			Name:        "ExpectedValid",
			Value:       graphql.APISpecTypeOpenAPI,
			ExpectValid: true,
		},
		{
			Name:        "Invalid object",
			Value:       graphql.APISpecType("INVALID"),
			ExpectValid: false,
		},
		{
			Name:        "Invalid default value",
			ExpectValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidAPISpecInput()
			obj.Type = testCase.Value
			//WHEN
			err := obj.Validate()
			//THEN
			if testCase.ExpectValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAPISpecInput_Validate_Format(t *testing.T) {
	testCases := []struct {
		Name        string
		Value       graphql.SpecFormat
		ExpectValid bool
	}{
		{
			Name:        "ExpectedValid JSON",
			Value:       graphql.SpecFormatJSON,
			ExpectValid: true,
		},
		{
			Name:        "Invalid object",
			Value:       graphql.SpecFormat("INVALID"),
			ExpectValid: false,
		},
		{
			Name:        "Invalid default value",
			ExpectValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidAPISpecInput()
			obj.Format = testCase.Value
			//WHEN
			err := obj.Validate()
			//THEN
			if testCase.ExpectValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAPISpecInput_Validate_TypeODataWithFormat(t *testing.T) {
	testCases := []struct {
		Name        string
		InputType   graphql.APISpecType
		InputFormat graphql.SpecFormat
		ExpectValid bool
	}{
		{
			Name:        "ExpectedValid ODATA with XML",
			InputType:   graphql.APISpecTypeOdata,
			InputFormat: graphql.SpecFormatXML,
			ExpectValid: true,
		},
		{
			Name:        "ExpectedValid ODATA with JSON",
			InputType:   graphql.APISpecTypeOdata,
			InputFormat: graphql.SpecFormatJSON,
			ExpectValid: true,
		},
		{
			Name:        "Invalid ODATA with YAML",
			InputType:   graphql.APISpecTypeOdata,
			InputFormat: graphql.SpecFormatYaml,
			ExpectValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidAPISpecInput()
			obj.Type = testCase.InputType
			obj.Format = testCase.InputFormat
			//WHEN
			err := obj.Validate()
			//THEN
			if testCase.ExpectValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAPISpecInput_Validate_TypeOpenAPIWithFormat(t *testing.T) {
	testCases := []struct {
		Name        string
		InputType   graphql.APISpecType
		InputFormat graphql.SpecFormat
		ExpectValid bool
	}{
		{
			Name:        "ExpectedValid OpenAPI with JSON",
			InputType:   graphql.APISpecTypeOpenAPI,
			InputFormat: graphql.SpecFormatJSON,
			ExpectValid: true,
		},
		{
			Name:        "ExpectedValid OpenAPI with YAML",
			InputType:   graphql.APISpecTypeOpenAPI,
			InputFormat: graphql.SpecFormatYaml,
			ExpectValid: true,
		},
		{
			Name:        "invalid OpenAPI with XML",
			InputType:   graphql.APISpecTypeOpenAPI,
			InputFormat: graphql.SpecFormatXML,
			ExpectValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidAPISpecInput()
			obj.Type = testCase.InputType
			obj.Format = testCase.InputFormat
			//WHEN
			err := obj.Validate()
			//THEN
			if testCase.ExpectValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAPISpecInput_Validate_FetchRequest(t *testing.T) {
	validObj := fixValidFetchRequestInput()
	emptyObj := graphql.FetchRequestInput{}

	testCases := []struct {
		Name        string
		Value       *graphql.FetchRequestInput
		DataClob    *graphql.CLOB
		ExpectValid bool
	}{
		{
			Name:        "ExpectedValid obj",
			Value:       &validObj,
			ExpectValid: true,
		},
		{
			Name:        "Nil object",
			Value:       nil,
			DataClob:    fixCLOB("data"),
			ExpectValid: true,
		},
		{
			Name:        "Invalid object",
			Value:       &emptyObj,
			ExpectValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidAPISpecInput()
			obj.FetchRequest = testCase.Value
			obj.Data = testCase.DataClob
			//WHEN
			err := obj.Validate()
			//THEN
			if testCase.ExpectValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidAPISpecInput() graphql.APISpecInput {
	return graphql.APISpecInput{
		Type:   graphql.APISpecTypeOpenAPI,
		Format: graphql.SpecFormatJSON,
		Data:   fixCLOB("data"),
	}
}

func fixValidAPIDefinitionInput() graphql.APIDefinitionInput {
	return graphql.APIDefinitionInput{
		Name:      inputvalidationtest.ValidName,
		TargetURL: inputvalidationtest.ValidURL,
	}
}
