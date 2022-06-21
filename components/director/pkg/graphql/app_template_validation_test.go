package graphql_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

// ApplicationTemaplteInput

func TestApplicationTemplateInput_Validate_Rule_ValidPlaceholders(t *testing.T) {
	testPlaceholderName := "test"

	testCases := []struct {
		Name  string
		Value []*graphql.PlaceholderDefinitionInput
		Valid bool
	}{
		{
			Name: "Valid",
			Value: []*graphql.PlaceholderDefinitionInput{
				{Name: testPlaceholderName, Description: str.Ptr("Test description")},
			},
			Valid: true,
		},
		{
			Name:  "Valid - no placeholders",
			Value: []*graphql.PlaceholderDefinitionInput{},
			Valid: true,
		},
		{
			Name: "Invalid - not unique",
			Value: []*graphql.PlaceholderDefinitionInput{
				{Name: testPlaceholderName, Description: str.Ptr("Test description")},
				{Name: testPlaceholderName, Description: str.Ptr("Different description")},
			},
			Valid: false,
		},
		{
			Name: "Invalid - not used",
			Value: []*graphql.PlaceholderDefinitionInput{
				{Name: "notused", Description: str.Ptr("Test description")},
			},
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidApplicationTemplateInput()
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
			Name: "Valid",
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
		Name  string
		Value []*graphql.PlaceholderDefinitionInput
		Valid bool
	}{
		{
			Name: "Valid",
			Value: []*graphql.PlaceholderDefinitionInput{
				{Name: testPlaceholderName, Description: str.Ptr("Test description")},
			},
			Valid: true,
		},
		{
			Name:  "Valid - Empty",
			Value: []*graphql.PlaceholderDefinitionInput{},
			Valid: true,
		},
		{
			Name:  "Valid - Nil",
			Value: nil,
			Valid: true,
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
			sut := fixValidApplicationTemplateInput()
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

// ApplicationTemaplteInput

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
				{Name: testPlaceholderName, Description: str.Ptr("Test description")},
			},
			Valid: true,
		},
		{
			Name:  "Valid - no placeholders",
			Value: []*graphql.PlaceholderDefinitionInput{},
			Valid: true,
		},
		{
			Name: "Invalid - not unique",
			Value: []*graphql.PlaceholderDefinitionInput{
				{Name: testPlaceholderName, Description: str.Ptr("Test description")},
				{Name: testPlaceholderName, Description: str.Ptr("Different description")},
			},
			Valid: false,
		},
		{
			Name: "Invalid - not used",
			Value: []*graphql.PlaceholderDefinitionInput{
				{Name: "notused", Description: str.Ptr("Test description")},
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
			Name: "Valid",
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
				{Name: testPlaceholderName, Description: str.Ptr("Test description")},
			},
			Valid: true,
		},
		{
			Name:  "Valid - Empty",
			Value: []*graphql.PlaceholderDefinitionInput{},
			Valid: true,
		},
		{
			Name:  "Valid - Nil",
			Value: nil,
			Valid: true,
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
			Name: "Valid",
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
			Name:  "Valid - no placeholders",
			Value: []*graphql.TemplateValueInput{},
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

func TestApplicationTemplateInput_Validate_Values(t *testing.T) {
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
			Name:  "Valid - Empty",
			Value: []*graphql.TemplateValueInput{},
			Valid: true,
		},
		{
			Name:  "Valid - Nil",
			Value: nil,
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
		ApplicationInput: &graphql.ApplicationRegisterInput{
			Name: "valid",
		},
		AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal,
		Webhooks:    []*graphql.WebhookInput{},
	}
}
func fixValidApplicationTemplateUpdateInput() graphql.ApplicationTemplateUpdateInput {
	return graphql.ApplicationTemplateUpdateInput{
		Name: "valid",
		ApplicationInput: &graphql.ApplicationRegisterInput{
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
