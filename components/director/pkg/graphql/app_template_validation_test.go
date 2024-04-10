package graphql_test

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

// ApplicationTemplateInput

func TestApplicationTemplateInput_Validate_Rule_ValidPlaceholders(t *testing.T) {
	testPlaceholderName := "test"

	testCases := []struct {
		Name                             string
		Value                            []*graphql.PlaceholderDefinitionInput
		ApplicationTemplateInputProvider func() graphql.ApplicationTemplateInput
		Error                            error
	}{
		{
			Name: "Valid",
			Value: []*graphql.PlaceholderDefinitionInput{
				{Name: testPlaceholderName, Description: str.Ptr("Test description"), JSONPath: str.Ptr("displayName")},
			},
			ApplicationTemplateInputProvider: applicationTemplateWithPlaceholderProvider(testPlaceholderName),
			Error:                            nil,
		},
		{
			Name:                             "Invalid - no placeholders defined",
			Value:                            []*graphql.PlaceholderDefinitionInput{},
			ApplicationTemplateInputProvider: applicationTemplateWithPlaceholderProvider(testPlaceholderName),
			Error:                            errors.New(`Placeholder [name=test] is used in the application input but it is not defined in the Placeholders array.`),
		},
		{
			Name:  "Invalid - empty placeholder in app input",
			Value: []*graphql.PlaceholderDefinitionInput{},
			ApplicationTemplateInputProvider: func() graphql.ApplicationTemplateInput {
				at := fixValidApplicationTemplateInput()
				at.ApplicationInput.Description = str.Ptr("{{}}")
				return at
			},
			Error: errors.New("Empty placeholder [name=] provided in the Application Input."),
		},
		{
			Name: "Invalid - not unique",
			Value: []*graphql.PlaceholderDefinitionInput{
				{Name: testPlaceholderName, Description: str.Ptr("Test description"), JSONPath: str.Ptr("displayName")},
				{Name: testPlaceholderName, Description: str.Ptr("Different description"), JSONPath: str.Ptr("displayName2")},
			},
			ApplicationTemplateInputProvider: applicationTemplateWithPlaceholderProvider(testPlaceholderName),
			Error:                            errors.New("placeholder [name=test] not unique."),
		},
		{
			Name: "Invalid - not used",
			Value: []*graphql.PlaceholderDefinitionInput{
				{Name: "notused", Description: str.Ptr("Test description"), JSONPath: str.Ptr("displayName")},
			},
			ApplicationTemplateInputProvider: applicationTemplateWithPlaceholderProvider(testPlaceholderName),
			Error:                            errors.New("application input does not use provided placeholder [name=notused].")},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			applicationTemplate := testCase.ApplicationTemplateInputProvider()
			applicationTemplate.Placeholders = testCase.Value
			// WHEN

			err := applicationTemplate.Validate()
			// THEN
			if testCase.Error == nil {
				require.NoError(t, err)
			} else {
				require.Contains(t, err.Error(), testCase.Error.Error())
			}
		})
	}
}

func TestApplicationTemplateInput_Validate_Name(t *testing.T) {
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
			sut := fixValidApplicationTemplateInput()
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

func TestApplicationTemplateInput_Validate_Description(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: str.Ptr("valid	valid"),
			Valid: true,
		},
		{
			Name:  "Valid - Nil",
			Value: (*string)(nil),
			Valid: true,
		},
		{
			Name:  "Valid - Empty",
			Value: str.Ptr(inputvalidationtest.EmptyString),
			Valid: true,
		},
		{
			Name:  "Invalid - Too long",
			Value: str.Ptr(inputvalidationtest.String2001Long),
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidApplicationTemplateInput()
			sut.Description = testCase.Value
			// WHEN
			err := sut.Validate()
			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationTemplateInput_Validate_ApplicationNamespace(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: str.Ptr("valid"),
			Valid: true,
		},
		{
			Name:  "Valid - Nil",
			Value: (*string)(nil),
			Valid: true,
		},
		{
			Name:  "Valid - Empty",
			Value: str.Ptr(inputvalidationtest.EmptyString),
			Valid: true,
		},
		{
			Name:  "Invalid - Too long",
			Value: str.Ptr(inputvalidationtest.String257Long),
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidApplicationTemplateInput()
			sut.ApplicationNamespace = testCase.Value
			// WHEN
			err := sut.Validate()
			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationTemplateInput_Validate_Placeholders(t *testing.T) {
	testPlaceholderName := "test"
	testCases := []struct {
		Name                             string
		Value                            []*graphql.PlaceholderDefinitionInput
		ApplicationTemplateInputProvider func() graphql.ApplicationTemplateInput
		Valid                            bool
	}{
		{
			Name: "Valid",
			Value: []*graphql.PlaceholderDefinitionInput{
				{Name: testPlaceholderName, Description: str.Ptr("Test description"), JSONPath: str.Ptr("displayName")},
			},
			ApplicationTemplateInputProvider: applicationTemplateWithPlaceholderProvider(testPlaceholderName),
			Valid:                            true,
		},
		{
			Name:  "Valid - Empty",
			Value: []*graphql.PlaceholderDefinitionInput{},
			ApplicationTemplateInputProvider: func() graphql.ApplicationTemplateInput {
				at := fixValidApplicationTemplateInput()
				at.ApplicationInput.Description = str.Ptr(testPlaceholderName)
				return at
			},
			Valid: true,
		},
		{
			Name:                             "Invalid - Empty placeholders array but placeholders are used in input",
			Value:                            []*graphql.PlaceholderDefinitionInput{},
			ApplicationTemplateInputProvider: applicationTemplateWithPlaceholderProvider(testPlaceholderName),
			Valid:                            false,
		},
		{
			Name:  "Valid - Nil",
			Value: nil,
			ApplicationTemplateInputProvider: func() graphql.ApplicationTemplateInput {
				at := fixValidApplicationTemplateInput()
				at.ApplicationInput.Description = str.Ptr(testPlaceholderName)
				return at
			},
			Valid: true,
		},
		{
			Name: "Invalid - Nil in slice",
			Value: []*graphql.PlaceholderDefinitionInput{
				nil,
			},
			ApplicationTemplateInputProvider: applicationTemplateWithPlaceholderProvider(testPlaceholderName),
			Valid:                            false,
		},
		{
			Name: "Invalid - Nested validation error",
			Value: []*graphql.PlaceholderDefinitionInput{
				{},
			},
			ApplicationTemplateInputProvider: applicationTemplateWithPlaceholderProvider(testPlaceholderName),
			Valid:                            false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			appTemplate := testCase.ApplicationTemplateInputProvider()
			appTemplate.Placeholders = testCase.Value
			// WHEN
			err := appTemplate.Validate()
			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationTemplateInput_Validate_AccessLevel(t *testing.T) {
	testCases := []struct {
		Name  string
		Value graphql.ApplicationTemplateAccessLevel
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: graphql.ApplicationTemplateAccessLevelGlobal,
			Valid: true,
		},
		{
			Name:  "Invalid - Empty",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
		},
		{
			Name:  "Invalid - Not in enum",
			Value: "invalid",
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidApplicationTemplateInput()
			sut.AccessLevel = testCase.Value
			// WHEN
			err := sut.Validate()
			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationTemplateInput_Validate_Webhooks(t *testing.T) {
	webhookInput := fixValidWebhookInput(inputvalidationtest.ValidURL)
	webhookInputWithInvalidOutputTemplate := fixValidWebhookInput(inputvalidationtest.ValidURL)
	webhookInputWithInvalidOutputTemplate.OutputTemplate = stringPtr(`{ "gone_status_code": 404, "success_status_code": 200}`)
	webhookInputwithInvalidURL := fixValidWebhookInput(inputvalidationtest.ValidURL)
	webhookInputwithInvalidURL.URL = nil
	testCases := []struct {
		Name  string
		Value []*graphql.WebhookInput
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: []*graphql.WebhookInput{&webhookInput},
			Valid: true,
		},
		{
			Name:  "Valid - Empty",
			Value: []*graphql.WebhookInput{},
			Valid: true,
		},
		{
			Name:  "Valid - nil",
			Value: nil,
			Valid: true,
		},
		{
			Name:  "Invalid - some of the webhooks are in invalid state - invalid output template",
			Value: []*graphql.WebhookInput{&webhookInputWithInvalidOutputTemplate},
			Valid: false,
		},
		{
			Name:  "Invalid - some of the webhooks are in invalid state - invalid URL",
			Value: []*graphql.WebhookInput{&webhookInputwithInvalidURL},
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidApplicationTemplateInput()
			sut.Webhooks = testCase.Value
			// WHEN
			err := sut.Validate()
			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationTemplateInput_Validate_Labels(t *testing.T) {
	validSystemRolesOneValue1 := []interface{}{"role1"}
	validSystemRolesOneValue2 := []interface{}{"role2"}
	validSystemRolesTwoValues := []interface{}{"role1", "role2"}
	invalidSystemRoleValueIsNotString := []interface{}{1}

	validSlisFilterOneValue := []interface{}{
		map[string]interface{}{
			"productId": "role1",
			"filter":    []map[string]interface{}{},
		},
	}
	invalidSlisFilterLabelWrongFormat := ""
	invalidSlisFilterValueWrongFormat := []interface{}{1}
	invalidSlisFilterMissingProductId := []interface{}{
		map[string]interface{}{
			"filter": []map[string]interface{}{},
		},
	}

	invalidSlisFilterInvalidFormatOfProductId := []interface{}{
		map[string]interface{}{
			"productId": []interface{}{},
			"filter":    []map[string]interface{}{},
		},
	}

	labelsInputWithSystemRole := fixLabelsInputWithSystemRole(validSystemRolesOneValue1)
	labelsInputWithSystemRoleAndSlisFilter := fixLabelsInputWithSystemRoleAndSlisFilter(validSystemRolesOneValue1, validSlisFilterOneValue)

	labelsInputWithMissingSystemRole := fixLabelsInputWithSystemRoleAndSlisFilter([]interface{}{}, validSlisFilterOneValue)

	testCases := []struct {
		Name  string
		Value graphql.Labels
		Valid bool
	}{
		{
			Name:  "Valid with system roles",
			Value: labelsInputWithSystemRole,
			Valid: true,
		},
		{
			Name:  "Valid with system roles and slis filter",
			Value: labelsInputWithSystemRoleAndSlisFilter,
			Valid: true,
		},
		{
			Name:  "Valid - Empty",
			Value: graphql.Labels{},
			Valid: true,
		},
		{
			Name:  "Valid - nil",
			Value: nil,
			Valid: true,
		},
		{
			Name:  "Not valid - missing system role when slis filter is defined",
			Value: labelsInputWithMissingSystemRole,
			Valid: false,
		},
		{
			Name:  "Not valid - value of cld system role is not a string",
			Value: fixLabelsInputWithSystemRoleAndSlisFilter(invalidSystemRoleValueIsNotString, validSlisFilterOneValue),
			Valid: false,
		},
		{
			Name:  "Not valid - invalid format of slis filter label",
			Value: fixLabelsInputWithSystemRoleAndSlisFilter(validSystemRolesOneValue1, invalidSlisFilterLabelWrongFormat),
			Valid: false,
		},
		{
			Name:  "Not valid - invalid format of slis filter value",
			Value: fixLabelsInputWithSystemRoleAndSlisFilter(validSystemRolesOneValue1, invalidSlisFilterValueWrongFormat),
			Valid: false,
		},
		{
			Name:  "Not valid - missing productId in slis filter",
			Value: fixLabelsInputWithSystemRoleAndSlisFilter(validSystemRolesOneValue1, invalidSlisFilterMissingProductId),
			Valid: false,
		},
		{
			Name:  "Not valid - invalid format of productId value in slis filter",
			Value: fixLabelsInputWithSystemRoleAndSlisFilter(validSystemRolesOneValue1, invalidSlisFilterInvalidFormatOfProductId),
			Valid: false,
		},
		{
			Name:  "Not valid - cld system roles count does not match the product ids count in slis filter",
			Value: fixLabelsInputWithSystemRoleAndSlisFilter(validSystemRolesTwoValues, validSlisFilterOneValue),
			Valid: false,
		},
		{
			Name:  "Not valid - cld system roles don't match with product ids in slis filter",
			Value: fixLabelsInputWithSystemRoleAndSlisFilter(validSystemRolesOneValue2, validSlisFilterOneValue),
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidApplicationTemplateInput()
			sut.Labels = testCase.Value
			// WHEN
			err := sut.Validate()
			fmt.Println(err)
			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// ApplicationTemplateUpdateInput

func TestApplicationTemplateUpdateInput_Validate_Rule_ValidPlaceholders(t *testing.T) {
	testPlaceholderName := "test"

	testCases := []struct {
		Name  string
		Value []*graphql.PlaceholderDefinitionInput
		Valid bool
	}{
		{
			Name: "Valid",
			Value: []*graphql.PlaceholderDefinitionInput{
				{Name: testPlaceholderName, Description: str.Ptr("Test description"), JSONPath: str.Ptr("displayName")},
			},
			Valid: true,
		},
		{
			Name:  "Invalid - no placeholders",
			Value: []*graphql.PlaceholderDefinitionInput{},
			Valid: false,
		},
		{
			Name: "Invalid - not unique",
			Value: []*graphql.PlaceholderDefinitionInput{
				{Name: testPlaceholderName, Description: str.Ptr("Test description"), JSONPath: str.Ptr("displayName")},
				{Name: testPlaceholderName, Description: str.Ptr("Different description"), JSONPath: str.Ptr("displayName2")},
			},
			Valid: false,
		},
		{
			Name: "Invalid - not used",
			Value: []*graphql.PlaceholderDefinitionInput{
				{Name: "notused", Description: str.Ptr("Test description"), JSONPath: str.Ptr("displayName")},
			},
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidApplicationTemplateUpdateInput()
			sut.ApplicationInput.Description = str.Ptr(fmt.Sprintf("{{%s}}", testPlaceholderName))
			sut.Placeholders = testCase.Value
			// WHEN
			err := sut.Validate()
			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationTemplateUpdateInput_Validate_Name(t *testing.T) {
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
			sut := fixValidApplicationTemplateUpdateInput()
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

func TestApplicationTemplateUpdateInput_Validate_Description(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: str.Ptr("valid	valid"),
			Valid: true,
		},
		{
			Name:  "Valid - Nil",
			Value: (*string)(nil),
			Valid: true,
		},
		{
			Name:  "Valid - Empty",
			Value: str.Ptr(inputvalidationtest.EmptyString),
			Valid: true,
		},
		{
			Name:  "Invalid - Too long",
			Value: str.Ptr(inputvalidationtest.String2001Long),
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidApplicationTemplateUpdateInput()
			sut.Description = testCase.Value
			// WHEN
			err := sut.Validate()
			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationTemplateUpdateInput_Validate_Placeholders(t *testing.T) {
	testPlaceholderName := "test"
	testCases := []struct {
		Name  string
		Value []*graphql.PlaceholderDefinitionInput
		Valid bool
	}{
		{
			Name: "Valid",
			Value: []*graphql.PlaceholderDefinitionInput{
				{Name: testPlaceholderName, Description: str.Ptr("Test description"), JSONPath: str.Ptr("displayName")},
			},
			Valid: true,
		},
		{
			Name:  "Invalid - Empty",
			Value: []*graphql.PlaceholderDefinitionInput{},
			Valid: false,
		},
		{
			Name:  "Invalid - Nil",
			Value: nil,
			Valid: false,
		},
		{
			Name: "Invalid - Nil in slice",
			Value: []*graphql.PlaceholderDefinitionInput{
				nil,
			},
			Valid: false,
		},
		{
			Name: "Invalid - Nested validation error",
			Value: []*graphql.PlaceholderDefinitionInput{
				{},
			},
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidApplicationTemplateUpdateInput()
			sut.ApplicationInput.Description = str.Ptr(fmt.Sprintf("{{%s}}", testPlaceholderName))
			sut.Placeholders = testCase.Value
			// WHEN
			err := sut.Validate()
			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationTemplateUpdateInput_Validate_AccessLevel(t *testing.T) {
	testCases := []struct {
		Name  string
		Value graphql.ApplicationTemplateAccessLevel
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: graphql.ApplicationTemplateAccessLevelGlobal,
			Valid: true,
		},
		{
			Name:  "Invalid - Empty",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
		},
		{
			Name:  "Invalid - Not in enum",
			Value: "invalid",
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidApplicationTemplateUpdateInput()
			sut.AccessLevel = testCase.Value
			// WHEN
			err := sut.Validate()
			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationTemplateUpdateInput_Validate_Webhooks(t *testing.T) {
	webhookInput := fixValidWebhookInput(inputvalidationtest.ValidURL)
	webhookInputWithInvalidOutputTemplate := fixValidWebhookInput(inputvalidationtest.ValidURL)
	webhookInputWithInvalidOutputTemplate.OutputTemplate = stringPtr(`{ "gone_status_code": 404, "success_status_code": 200}`)
	webhookInputWithInvalidURL := fixValidWebhookInput(inputvalidationtest.ValidURL)
	webhookInputWithInvalidURL.URL = nil
	testCases := []struct {
		Name  string
		Value []*graphql.WebhookInput
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: []*graphql.WebhookInput{&webhookInput},
			Valid: true,
		},
		{
			Name:  "Valid - Empty",
			Value: []*graphql.WebhookInput{},
			Valid: true,
		},
		{
			Name:  "Valid - nil",
			Value: nil,
			Valid: true,
		},
		{
			Name:  "Invalid - some of the webhooks are in invalid state - invalid output template",
			Value: []*graphql.WebhookInput{&webhookInputWithInvalidOutputTemplate},
			Valid: false,
		},
		{
			Name:  "Invalid - some of the webhooks are in invalid state - invalid URL",
			Value: []*graphql.WebhookInput{&webhookInputWithInvalidURL},
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidApplicationTemplateUpdateInput()
			sut.ApplicationInput.Webhooks = testCase.Value
			// WHEN
			err := sut.Validate()
			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// PlaceholderDefinitionInput

func TestPlaceholderDefinitionInput_Validate_Name(t *testing.T) {
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
			Name:  "Invalid - Empty",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
		},
		{
			Name:  "Invalid - Invalid Name",
			Value: inputvalidationtest.InvalidName,
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidPlaceholderDefintionInput()
			sut.Name = testCase.Value
			// WHEN
			err := sut.Validate()
			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestPlaceholderDefinitionInput_Validate_Description(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: str.Ptr("valid	valid"),
			Valid: true,
		},
		{
			Name:  "Valid - Nil",
			Value: (*string)(nil),
			Valid: true,
		},
		{
			Name:  "Valid - Empty",
			Value: str.Ptr(inputvalidationtest.EmptyString),
			Valid: true,
		},
		{
			Name:  "Invalid - Too long",
			Value: str.Ptr(inputvalidationtest.String2001Long),
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidPlaceholderDefintionInput()
			sut.Description = testCase.Value
			// WHEN
			err := sut.Validate()
			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationFromTemplateInput_Validate_Rule_EitherPlaceholdersOrPlaceholdersPayloadExists(t *testing.T) {
	testPlaceholderName := "test"
	testPlacehoderPayload := "{\"a\":\"b\"}"

	testCases := []struct {
		Name                string
		Value               []*graphql.TemplateValueInput
		PlaceholdersPayload *string
		Valid               bool
	}{
		{
			Name: "Valid - only Value",
			Value: []*graphql.TemplateValueInput{
				{Placeholder: testPlaceholderName, Value: "abc"},
			},
			Valid: true,
		},
		{
			Name:                "Valid - only PlaceholdersPayload",
			PlaceholdersPayload: &testPlacehoderPayload,
			Valid:               true,
		},
		{
			Name: "Invalid - both Value and PlaceholdersPayload",
			Value: []*graphql.TemplateValueInput{
				{Placeholder: testPlaceholderName, Value: "abc"},
			},
			PlaceholdersPayload: &testPlacehoderPayload,
			Valid:               false,
		},
		{
			Name:  "Invalid - neither Value nor PlaceholdersPayload",
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidApplicationFromTemplateInput()
			sut.Values = testCase.Value
			sut.PlaceholdersPayload = testCase.PlaceholdersPayload
			// WHEN
			err := sut.Validate()
			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// ApplicationFromTemplateInput

func TestApplicationFromTemplateInput_Validate_Rule_UniquePlaceholders(t *testing.T) {
	testPlaceholderName := "test"

	testCases := []struct {
		Name  string
		Value []*graphql.TemplateValueInput
		Valid bool
	}{
		{
			Name: "Valid",
			Value: []*graphql.TemplateValueInput{
				{Placeholder: testPlaceholderName, Value: ""},
			},
			Valid: true,
		},
		{
			Name: "Invalid - not unique",
			Value: []*graphql.TemplateValueInput{
				{Placeholder: testPlaceholderName, Value: "one"},
				{Placeholder: testPlaceholderName, Value: "two"},
			},
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidApplicationFromTemplateInput()
			sut.Values = testCase.Value
			// WHEN
			err := sut.Validate()
			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestApplicationFromTemplateInput_Validate_TemplateName(t *testing.T) {
	testPlacehoderPayload := "{\"a\":\"b\"}"
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
			sut := fixValidApplicationFromTemplateInput()
			sut.TemplateName = testCase.Value
			sut.PlaceholdersPayload = &testPlacehoderPayload
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

func TestApplicationTemplateInput_Validate_Value(t *testing.T) {
	testPlaceholderName := "test"
	testCases := []struct {
		Name  string
		Value []*graphql.TemplateValueInput
		Valid bool
	}{
		{
			Name: "Valid",
			Value: []*graphql.TemplateValueInput{
				{Placeholder: testPlaceholderName, Value: "valid"},
			},
			Valid: true,
		},
		{
			Name: "Invalid - Nil in slice",
			Value: []*graphql.TemplateValueInput{
				nil,
			},
			Valid: false,
		},
		{
			Name: "Invalid - Nested validation error",
			Value: []*graphql.TemplateValueInput{
				{},
			},
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidApplicationFromTemplateInput()
			sut.Values = testCase.Value
			// WHEN
			err := sut.Validate()
			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// TemplateValueInput

func TestTemplateValueInput_Validate_Name(t *testing.T) {
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
			Name:  "Invalid - Empty",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
		},
		{
			Name:  "Invalid - Invalid Name",
			Value: inputvalidationtest.InvalidName,
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidTemplateValueInput()
			sut.Placeholder = testCase.Value
			// WHEN
			err := sut.Validate()
			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestTemplateValueInput_Validate_Description(t *testing.T) {
	testCases := []struct {
		Name  string
		Value string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: "valid",
			Valid: true,
		},
		{
			Name:  "Valid - Empty",
			Value: inputvalidationtest.EmptyString,
			Valid: true,
		},
		{
			Name:  "Invalid - Too long",
			Value: inputvalidationtest.String129Long,
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidTemplateValueInput()
			sut.Value = testCase.Value
			// WHEN
			err := sut.Validate()
			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// fixtures

func fixValidApplicationTemplateInput() graphql.ApplicationTemplateInput {
	return graphql.ApplicationTemplateInput{
		Name: "valid",
		ApplicationInput: &graphql.ApplicationJSONInput{
			Name: "valid",
			Webhooks: []*graphql.WebhookInput{
				{
					URL:           stringPtr("http://localhost.com"),
					Type:          graphql.WebhookTypeConfigurationChanged,
					InputTemplate: stringPtr(`{"context":{ {{ if .CustomerTenantContext.AccountID }}"btp": {"uclFormationId":"{{.FormationID}}","globalAccountId":"{{.CustomerTenantContext.AccountID}}","crmId":"{{.CustomerTenantContext.CustomerID}}"} {{ else }}"atom": {"uclFormationId":"{{.FormationID}}","path":"{{.CustomerTenantContext.Path}}","crmId":"{{.CustomerTenantContext.CustomerID}}"} {{ end }} },"items": [ {"uclAssignmentId":"{{ .Assignment.ID }}","operation":"{{.Operation}}","deploymentRegion":"{{if .Application.Labels.region }}{{.Application.Labels.region}}{{ else }}{{.ApplicationTemplate.Labels.region}}{{end }}","applicationNamespace":"{{ if .Application.ApplicationNamespace }}{{.Application.ApplicationNamespace}}{{else }}{{.ApplicationTemplate.ApplicationNamespace}}{{ end }}","applicationTenantId":"{{.Application.LocalTenantID}}","uclSystemTenantId":"{{.Application.ID}}",{{ if .ApplicationTemplate.Labels.parameters }}"parameters": {{.ApplicationTemplate.Labels.parameters}},{{ end }}"configuration": {{.ReverseAssignment.Value}} } ] }`),
				},
			},
		},
		AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal,
		Webhooks:    []*graphql.WebhookInput{},
	}
}
func fixValidApplicationTemplateUpdateInput() graphql.ApplicationTemplateUpdateInput {
	return graphql.ApplicationTemplateUpdateInput{
		Name: "valid",
		ApplicationInput: &graphql.ApplicationJSONInput{
			Name: "valid",
		},
		AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixValidPlaceholderDefintionInput() graphql.PlaceholderDefinitionInput {
	return graphql.PlaceholderDefinitionInput{
		Name: "valid",
	}
}

func fixValidApplicationFromTemplateInput() graphql.ApplicationFromTemplateInput {
	return graphql.ApplicationFromTemplateInput{
		TemplateName: "valid",
	}
}

func fixValidTemplateValueInput() graphql.TemplateValueInput {
	return graphql.TemplateValueInput{
		Placeholder: "test",
		Value:       "",
	}
}

func applicationTemplateWithPlaceholderProvider(testPlaceholderName string) func() graphql.ApplicationTemplateInput {
	return func() graphql.ApplicationTemplateInput {
		at := fixValidApplicationTemplateInput()
		at.ApplicationInput.Description = str.Ptr(fmt.Sprintf("{{%s}}", testPlaceholderName))
		return at
	}
}
