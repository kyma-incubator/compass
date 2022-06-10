package graphql_test

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestFormationTemplateInput_ValidateName(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "Success",
			Value:         "a normal name for once",
			ExpectedValid: true,
		},
		{
			Name:          "Name longer than 256",
			Value:         strings.Repeat("some-name", 50),
			ExpectedValid: false,
		},
		{
			Name:          "Invalid",
			Value:         "",
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			formationTemplateInput := fixValidFormationTemplateInput()
			formationTemplateInput.Name = testCase.Value
			// WHEN
			err := formationTemplateInput.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestFormationTemplateInput_ValidateMissingMessages(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "Success",
			Value:         "a normal name for once",
			ExpectedValid: true,
		},
		{
			Name:          "Name longer than 512",
			Value:         strings.Repeat("some-name", 100),
			ExpectedValid: false,
		},
		{
			Name:          "Invalid",
			Value:         "",
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run("Warning Message "+testCase.Name, func(t *testing.T) {
			//GIVEN
			formationTemplateInput := fixValidFormationTemplateInput()
			formationTemplateInput.MissingArtifactWarningMessage = testCase.Value
			// WHEN
			err := formationTemplateInput.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
		t.Run("Info Message "+testCase.Name, func(t *testing.T) {
			//GIVEN
			formationTemplateInput := fixValidFormationTemplateInput()
			formationTemplateInput.MissingArtifactInfoMessage = testCase.Value
			// WHEN
			err := formationTemplateInput.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestFormationTemplateInput_ValidateTypes(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         []string
		ExpectedValid bool
	}{
		{
			Name:          "Success",
			Value:         []string{"normal-type", "another-normal-type"},
			ExpectedValid: true,
		},
		{
			Name:          "Empty slice",
			Value:         []string{},
			ExpectedValid: false,
		},
		{
			Name:          "Nil slice",
			Value:         nil,
			ExpectedValid: false,
		},
		{
			Name:          "Empty elements in slice",
			Value:         []string{""},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run("Application Types  "+testCase.Name, func(t *testing.T) {
			//GIVEN
			formationTemplateInput := fixValidFormationTemplateInput()
			formationTemplateInput.ApplicationTypes = testCase.Value
			// WHEN
			err := formationTemplateInput.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
		t.Run("Runtime Types "+testCase.Name, func(t *testing.T) {
			//GIVEN
			formationTemplateInput := fixValidFormationTemplateInput()
			formationTemplateInput.RuntimeTypes = testCase.Value
			// WHEN
			err := formationTemplateInput.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidFormationTemplateInput() graphql.FormationTemplateInput {
	return graphql.FormationTemplateInput{
		Name:                          "formation-template-name",
		ApplicationTypes:              []string{"some-application-type"},
		RuntimeTypes:                  []string{"some-runtime-type"},
		MissingArtifactInfoMessage:    "some missing info message",
		MissingArtifactWarningMessage: "some missing warning message",
	}
}
