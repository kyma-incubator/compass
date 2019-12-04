package graphql_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

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
			//WHEN
			err := sut.Validate()
			//THEN
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
		Name  string
		Value string
		Valid bool
	}{
		{
			Name: "Valid",
			Value: ,
			Valid: true,
		},
		{
			Name: "Invalid",
			Value:,
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidApplicationTemplateInput()
			sut.Name = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidApplicationTemplateInput() graphql.ApplicationTemplateInput {
	return graphql.ApplicationTemplateInput{
		Name: "valid",
		ApplicationInput: &graphql.ApplicationCreateInput{
			Name: "valid",
		},
		AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal,
	}
}
