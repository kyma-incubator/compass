package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

func TestPackageCreateInput_Validate_Name(t *testing.T) {
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
			obj := fixValidPackageCreateInput()
			obj.Name = testCase.Value
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

func TestPackageCreateInput_Validate_Description(t *testing.T) {
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
			obj := fixValidPackageCreateInput()
			obj.Description = testCase.Value
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

func TestPackageCreateInput_Validate_DefaultInstanceAuth(t *testing.T) {
	validObj := fixValidAuthInput()
	emptyObj := graphql.AuthInput{}

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
			Name:          "Invalid object",
			Value:         &emptyObj,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidPackageCreateInput()
			obj.DefaultInstanceAuth = testCase.Value
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

func TestPackageCreateInput_Validate_InstanceAuthRequestInputSchema(t *testing.T) {
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
			obj := fixValidPackageCreateInput()
			obj.InstanceAuthRequestInputSchema = testCase.Value
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

func TestPackageCreateInput_Validate_APIs(t *testing.T) {
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
			app := fixValidPackageCreateInput()
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

func TestPackageCreateInput_Validate_EventAPIs(t *testing.T) {
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
			app := fixValidPackageCreateInput()
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

func TestPackageCreateInput_Validate_Documents(t *testing.T) {
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
			app := fixValidPackageCreateInput()
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

func TestPackageUpdateInput_Validate_Name(t *testing.T) {
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
			obj := fixValidPackageUpdateInput()
			obj.Name = testCase.Value
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

func TestPackageUpdateInput_Validate_Description(t *testing.T) {
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
			obj := fixValidPackageUpdateInput()
			obj.Description = testCase.Value
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

func TestPackageUpdateInput_Validate_DefaultInstanceAuth(t *testing.T) {
	validObj := fixValidAuthInput()
	emptyObj := graphql.AuthInput{}

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
			Name:          "Invalid object",
			Value:         &emptyObj,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidPackageUpdateInput()
			obj.DefaultInstanceAuth = testCase.Value
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

func TestPackageUpdateInput_Validate_InstanceAuthRequestInputSchema(t *testing.T) {
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
			obj := fixValidPackageUpdateInput()
			obj.InstanceAuthRequestInputSchema = testCase.Value
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
			Name: "Auth",
			Value: graphql.PackageInstanceAuthSetInput{
				Auth: &authInput,
			},
			ExpectedValid: true,
		},
		{
			Name: "Failed Status",
			Value: graphql.PackageInstanceAuthSetInput{
				Status: &graphql.PackageInstanceAuthStatusInput{
					Condition: graphql.PackageInstanceAuthSetStatusConditionInputFailed,
					Reason:    &str,
				},
			},
			ExpectedValid: true,
		},
		{
			Name: "Success Status",
			Value: graphql.PackageInstanceAuthSetInput{
				Status: &graphql.PackageInstanceAuthStatusInput{
					Condition: graphql.PackageInstanceAuthSetStatusConditionInputSucceeded,
				},
			},
			ExpectedValid: false,
		},
		{
			Name: "Auth and Success Status",
			Value: graphql.PackageInstanceAuthSetInput{
				Auth: &authInput,
				Status: &graphql.PackageInstanceAuthStatusInput{
					Condition: graphql.PackageInstanceAuthSetStatusConditionInputSucceeded,
					Message:   &str,
					Reason:    &str,
				},
			},
			ExpectedValid: true,
		},
		{
			Name: "Auth and Failure Status",
			Value: graphql.PackageInstanceAuthSetInput{
				Auth: &authInput,
				Status: &graphql.PackageInstanceAuthStatusInput{
					Condition: graphql.PackageInstanceAuthSetStatusConditionInputFailed,
				},
			},
			ExpectedValid: false,
		},
		{
			Name: "Empty objects",
			Value: graphql.PackageInstanceAuthSetInput{
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

func TestPackageInstanceAuthStatusInput_Validate(t *testing.T) {
	//GIVEN
	str := "foo"
	emptyStr := ""
	testCases := []struct {
		Name          string
		Value         graphql.PackageInstanceAuthStatusInput
		ExpectedValid bool
	}{
		{
			Name: "Success condition",
			Value: graphql.PackageInstanceAuthStatusInput{
				Condition: graphql.PackageInstanceAuthSetStatusConditionInputSucceeded,
			},
			ExpectedValid: true,
		},
		{
			Name: "Failed condition without reason",
			Value: graphql.PackageInstanceAuthStatusInput{
				Condition: graphql.PackageInstanceAuthSetStatusConditionInputFailed,
			},
			ExpectedValid: false,
		},
		{
			Name: "Failed condition with reason",
			Value: graphql.PackageInstanceAuthStatusInput{
				Condition: graphql.PackageInstanceAuthSetStatusConditionInputFailed,
				Reason:    &str,
			},
			ExpectedValid: true,
		},
		{
			Name: "Empty Reason",
			Value: graphql.PackageInstanceAuthStatusInput{
				Condition: graphql.PackageInstanceAuthSetStatusConditionInputSucceeded,
				Reason:    &emptyStr,
			},
			ExpectedValid: false,
		},
		{
			Name: "Empty Message",
			Value: graphql.PackageInstanceAuthStatusInput{
				Condition: graphql.PackageInstanceAuthSetStatusConditionInputSucceeded,
				Message:   &emptyStr,
			},
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

func fixValidPackageCreateInput() graphql.PackageCreateInput {
	return graphql.PackageCreateInput{
		Name: inputvalidationtest.ValidName,
	}
}

func fixValidPackageUpdateInput() graphql.PackageUpdateInput {
	return graphql.PackageUpdateInput{
		Name: inputvalidationtest.ValidName,
	}
}
