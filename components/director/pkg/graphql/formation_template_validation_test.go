package graphql_test

import (
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/require"
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

func TestFormationTemplateInput_ValidateRuntimeDisplayName(t *testing.T) {
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
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			formationTemplateInput := fixValidFormationTemplateInput()
			formationTemplateInput.RuntimeTypeDisplayName = testCase.Value
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

func TestFormationTemplateInput_ValidateRuntimeArtifactKind(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         graphql.ArtifactType
		ExpectedValid bool
	}{
		{
			Name:          "Success - Service Instance",
			Value:         graphql.ArtifactTypeServiceInstance,
			ExpectedValid: true,
		},
		{
			Name:          "Success - Subscription",
			Value:         graphql.ArtifactTypeSubscription,
			ExpectedValid: true,
		},
		{
			Name:          "Success - Environment Instance",
			Value:         graphql.ArtifactTypeEnvironmentInstance,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid type",
			Value:         graphql.ArtifactType("Invalid type"),
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
			formationTemplateInput.RuntimeArtifactKind = testCase.Value
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

func TestFormationTemplateInput_ValidateApplicationTypes(t *testing.T) {
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
		t.Run(testCase.Name, func(t *testing.T) {
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
	}
}

func TestFormationTemplateInput_ValidateRuntimeTypes(t *testing.T) {
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
		t.Run(testCase.Name, func(t *testing.T) {
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

func TestFormationTemplateInput_Validate_Webhooks(t *testing.T) {
	webhookInput := fixValidWebhookInput(inputvalidationtest.ValidURL)
	webhookInputWithInvalidOutputTemplate := fixValidWebhookInput(inputvalidationtest.ValidURL)
	webhookInputWithInvalidMode := fixValidWebhookInput(inputvalidationtest.ValidURL)
	webhookInputWithInvalidMode.Type = graphql.WebhookTypeFormationLifecycle
	webhookInputWithInvalidMode.Mode = webhookModePtr(graphql.WebhookModeAsync)
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
			Name:  "Invalid - type is 'FORMATION_LIFECYCLE' and mode is not 'SYNC'",
			Value: []*graphql.WebhookInput{&webhookInputWithInvalidMode},
			Valid: false,
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
			sut := fixValidFormationTemplateInput()
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

func fixValidFormationTemplateInput() graphql.FormationTemplateInput {
	return graphql.FormationTemplateInput{
		Name:                   "formation-template-name",
		ApplicationTypes:       []string{"some-application-type"},
		RuntimeTypes:           []string{"some-runtime-type"},
		RuntimeTypeDisplayName: "display-name-for-runtime",
		RuntimeArtifactKind:    graphql.ArtifactTypeSubscription,
	}
}
