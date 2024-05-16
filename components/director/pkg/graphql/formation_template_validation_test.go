package graphql_test

import (
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/require"
)

const (
	runtimeArtifactKindField    = "RuntimeArtifactKind"
	runtimeTypeDisplayNameField = "RuntimeTypeDisplayName"
	runtimeTypesField           = "RuntimeTypes"
)

func TestFormationTemplateRegisterInput_ValidateName(t *testing.T) {
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
			formationTemplateRegisterInput := fixValidFormationTemplateRegisterInput()
			formationTemplateRegisterInput.Name = testCase.Value
			// WHEN
			err := formationTemplateRegisterInput.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestFormationTemplateRegisterInput_ValidateRuntimeDisplayName(t *testing.T) {
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
			formationTemplateRegisterInput := fixValidFormationTemplateRegisterInput()
			formationTemplateRegisterInput.RuntimeTypeDisplayName = &testCase.Value
			// WHEN
			err := formationTemplateRegisterInput.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestFormationTemplateRegisterInput_ValidateRuntimeArtifactKind(t *testing.T) {
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
			formationTemplateRegisterInput := fixValidFormationTemplateRegisterInput()
			formationTemplateRegisterInput.RuntimeArtifactKind = &testCase.Value
			// WHEN
			err := formationTemplateRegisterInput.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestFormationTemplateRegisterInput_ValidateApplicationTypes(t *testing.T) {
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
			formationTemplateRegisterInput := fixValidFormationTemplateRegisterInput()
			formationTemplateRegisterInput.ApplicationTypes = testCase.Value
			// WHEN
			err := formationTemplateRegisterInput.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestFormationTemplateRegisterInput_ValidateRuntimeTypes(t *testing.T) {
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
			formationTemplateRegisterInput := fixValidFormationTemplateRegisterInput()
			formationTemplateRegisterInput.RuntimeTypes = testCase.Value
			// WHEN
			err := formationTemplateRegisterInput.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestFormationTemplateRegisterInput_ValidateDiscoveryConsumers(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         []string
		ExpectedValid bool
	}{
		{
			Name:          "Success",
			Value:         []string{"some-runtime-type", "some-application-type"},
			ExpectedValid: true,
		},
		{
			Name:          "Non-existing type",
			Value:         []string{"non-existing-type", "some-application-type"},
			ExpectedValid: false,
		},
		{
			Name:          "Empty slice",
			Value:         []string{},
			ExpectedValid: true,
		},
		{
			Name:          "Nil slice",
			Value:         nil,
			ExpectedValid: true,
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
			formationTemplateRegisterInput := fixValidFormationTemplateRegisterInput()
			formationTemplateRegisterInput.DiscoveryConsumers = testCase.Value
			// WHEN
			err := formationTemplateRegisterInput.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestFormationTemplateRegisterInput_Validate_Webhooks(t *testing.T) {
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
			sut := fixValidFormationTemplateRegisterInput()
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

func TestFormationTemplateRegisterInput_ValidateRuntimeRelatedFields(t *testing.T) {
	testCases := []struct {
		Name          string
		EmptyFields   []string
		ExpectedValid bool
	}{
		{
			Name:          "Success all fields present",
			EmptyFields:   []string{},
			ExpectedValid: true,
		},
		{
			Name:          "Success all fields missing",
			EmptyFields:   []string{runtimeTypesField, runtimeTypeDisplayNameField, runtimeArtifactKindField},
			ExpectedValid: true,
		},
		{
			Name:          "Missing artifact kind",
			EmptyFields:   []string{runtimeArtifactKindField},
			ExpectedValid: false,
		},
		{
			Name:          "Missing display name",
			EmptyFields:   []string{runtimeTypeDisplayNameField},
			ExpectedValid: false,
		},
		{
			Name:          "Missing runtime types",
			EmptyFields:   []string{runtimeTypesField},
			ExpectedValid: false,
		},
		{
			Name:          "Missing artifact kind and display name",
			EmptyFields:   []string{runtimeArtifactKindField, runtimeTypeDisplayNameField},
			ExpectedValid: false,
		},
		{
			Name:          "Missing artifact kind and runtime types",
			EmptyFields:   []string{runtimeArtifactKindField, runtimeTypesField},
			ExpectedValid: false,
		},
		{
			Name:          "Missing display name and runtime types",
			EmptyFields:   []string{runtimeTypeDisplayNameField, runtimeTypesField},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			formationTemplateRegisterInput := fixValidFormationTemplateRegisterInput()

			for _, field := range testCase.EmptyFields {
				switch field {
				case runtimeTypesField:
					formationTemplateRegisterInput.RuntimeTypes = []string{}
				case runtimeArtifactKindField:
					formationTemplateRegisterInput.RuntimeArtifactKind = nil
				case runtimeTypeDisplayNameField:
					formationTemplateRegisterInput.RuntimeTypeDisplayName = nil
				}
			}

			// WHEN
			err := formationTemplateRegisterInput.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidFormationTemplateRegisterInput() graphql.FormationTemplateRegisterInput {
	kind := graphql.ArtifactTypeSubscription
	return graphql.FormationTemplateRegisterInput{
		Name:                   "formation-template-name",
		ApplicationTypes:       []string{"some-application-type"},
		RuntimeTypes:           []string{"some-runtime-type"},
		RuntimeTypeDisplayName: stringPtr("display-name-for-runtime"),
		RuntimeArtifactKind:    &kind,
	}
}
