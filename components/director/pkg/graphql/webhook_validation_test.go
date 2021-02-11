package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/stretchr/testify/require"
)

func TestWebhookInput_Validate_Type(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         graphql.WebhookType
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         graphql.WebhookTypeConfigurationChanged,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - Empty",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "Invalid - Not enum",
			Value:         "invalid",
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidWebhookInput()
			sut.Type = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestWebhookInput_Validate_URL(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         inputvalidationtest.ValidURL,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "Invalid - Invalid URL",
			Value:         inputvalidationtest.InvalidURL,
			ExpectedValid: false,
		},
		{
			Name:          "Invalid - Too long",
			Value:         inputvalidationtest.URL257Long,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidWebhookInput()
			sut.URL = &testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestWebhookInput_Validate_Auth(t *testing.T) {
	auth := fixValidAuthInput()
	testCases := []struct {
		Name          string
		Value         *graphql.AuthInput
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         &auth,
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid - nil",
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
			sut := fixValidWebhookInput()
			sut.Auth = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestWebhookInput_Validate_CorrelationIDKey(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         stringPtr(correlation.RequestIDHeaderKey),
			ExpectedValid: true,
		},
		{
			Name:          "Empty",
			Value:         stringPtr(inputvalidationtest.EmptyString),
			ExpectedValid: true,
		},
		{
			Name:          "Nil",
			Value:         nil,
			ExpectedValid: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidWebhookInput()
			sut.CorrelationIDKey = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestWebhookInput_Validate_Mode(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *graphql.WebhookMode
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         webhookModePtr(graphql.WebhookModeSync),
			ExpectedValid: true,
		},
		{
			Name:          "Empty",
			Value:         webhookModePtr(inputvalidationtest.EmptyString),
			ExpectedValid: true,
		},
		{
			Name:          "Nil",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - Not enum",
			Value:         webhookModePtr("invalid"),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidWebhookInput()
			sut.Mode = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestWebhookInput_Validate_RetryInterval(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *int
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         intPtr(120),
			ExpectedValid: true,
		},
		{
			Name:          "Empty",
			Value:         intPtr(0),
			ExpectedValid: true,
		},
		{
			Name:          "Nil",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - negative number",
			Value:         intPtr(-1),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidWebhookInput()
			sut.RetryInterval = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestWebhookInput_Validate_Timeout(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *int
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         intPtr(120),
			ExpectedValid: true,
		},
		{
			Name:          "Empty",
			Value:         intPtr(0),
			ExpectedValid: true,
		},
		{
			Name:          "Nil",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - negative number",
			Value:         intPtr(-1),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidWebhookInput()
			sut.Timeout = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestWebhookInput_Validate_URLTemplate(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         stringPtr("https://my-int-system/api/v1/{{.Application.ID}}/pairing"),
			ExpectedValid: true,
		},
		{
			Name:          "Empty",
			Value:         stringPtr(""),
			ExpectedValid: false, // it should not be valid due to the fact that we also discard the legacy URL from the webhook in the test below
		},
		{
			Name:          "Nil",
			Value:         nil,
			ExpectedValid: false, // it should not be valid due to the fact that we also discard the legacy URL from the webhook in the test below
		},
		{
			Name:          "Invalid - contains unexpected property",
			Value:         stringPtr("https://my-int-system/api/v1/{{.Application.Group}}/pairing"),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidWebhookInput()
			sut.URL = nil
			sut.URLTemplate = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestWebhookInput_Validate_InputTemplate(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name: "ExpectedValid",
			Value: stringPtr(`{
			  "app_id": "{{.Application.ID}}",
			  "app_name": "{{.Application.Name}}"
			}`),
			ExpectedValid: true,
		},
		{
			Name:          "Empty",
			Value:         stringPtr(""),
			ExpectedValid: false,
		},
		{
			Name:          "Nil",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name: "Invalid - contains unexpected property",
			Value: stringPtr(`{
			  "app_id": "{{.Application.ID}}",
			  "app_name": "{{.Application.Group}}"
			}`),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidWebhookInput()
			sut.InputTemplate = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestWebhookInput_Validate_HeaderTemplate(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         stringPtr(`{"Content-Type":["application/json"],"client_user":["{{.Headers.user_id}}"]}`),
			ExpectedValid: true,
		},
		{
			Name:          "Empty",
			Value:         stringPtr(""),
			ExpectedValid: true,
		},
		{
			Name:          "Nil",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - contains unexpected property",
			Value:         stringPtr(`{"Content-Type":["application/json"],"client_user":["{{.Body.Group}}"]}`),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidWebhookInput()
			sut.HeaderTemplate = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestWebhookInput_Validate_OutputTemplate(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name: "ExpectedValid",
			Value: stringPtr(`{
			   "location": "{{.Headers.location}}",
			   "success_status_code": 202,
			   "error": "{{.Body.error}}"
			 }`),
			ExpectedValid: true,
		},
		{
			Name: "Invalid - contains only location",
			Value: stringPtr(`{
			   "location": "{{.Headers.location}}"
			 }`),
			ExpectedValid: false,
		},
		{
			Name: "Invalid - contains only success status code",
			Value: stringPtr(`{
			   "success_status_code": 202
			 }`),
			ExpectedValid: false,
		},
		{
			Name: "Invalid - contains only error",
			Value: stringPtr(`{
			   "error": "{{.Body.error}}"
			 }`),
			ExpectedValid: false,
		},
		{
			Name:          "Empty",
			Value:         stringPtr(""),
			ExpectedValid: false,
		},
		{
			Name:          "Nil",
			Value:         nil,
			ExpectedValid: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidWebhookInput()
			sut.InputTemplate = nil
			sut.OutputTemplate = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestWebhookInput_Validate_StatusTemplate(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name: "ExpectedValid",
			Value: stringPtr(`{
			   "status": "{{.Body.status}}",
			   "success_status_code": 200,
			   "error": "{{.Body.error}}"
			}`),
			ExpectedValid: true,
		},
		{
			Name: "Invalid - missing error",
			Value: stringPtr(`{
			   "status": "{{.Body.status}}",
			   "success_status_code": 200
			 }`),
			ExpectedValid: false,
		},
		{
			Name: "Invalid -  missing status",
			Value: stringPtr(`{
			   "success_status_code": 200,
			   "error": "{{.Body.error}}"
			 }`),
			ExpectedValid: false,
		},
		{
			Name: "Invalid - missing success status code",
			Value: stringPtr(`{
			   "status": "{{.Body.status}}",
			   "error": "{{.Body.error}}"
			 }`),
			ExpectedValid: false,
		},
		{
			Name:          "Empty",
			Value:         stringPtr(""),
			ExpectedValid: false,
		},
		{
			Name:          "Nil",
			Value:         nil,
			ExpectedValid: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidWebhookInput()
			sut.Mode = webhookModePtr(graphql.WebhookModeAsync)
			sut.StatusTemplate = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestWebhookInput_Validate_BothURLAndURLTemplate(t *testing.T) {
	sut := fixValidWebhookInput()
	sut.URL = stringPtr("https://my-int-system/api/v1/123/pairing")
	sut.URLTemplate = stringPtr("https://my-int-system/api/v1/{{.Application.ID}}/pairing")
	//WHEN
	err := sut.Validate()
	//THEN
	require.Error(t, err)
}

func TestWebhookInput_Validate_InputTemplateProvided_MissingOutputTemplate_ShouldReturnError(t *testing.T) {
	sut := fixValidWebhookInput()
	sut.InputTemplate = stringPtr(`{
	  "app_id": "{{.Application.ID}}",
	  "app_name": "{{.Application.Name}}"
	}`)
	sut.OutputTemplate = nil
	//WHEN
	err := sut.Validate()
	//THEN
	require.Error(t, err)
}

func TestWebhookInput_Validate_AsyncWebhook_MissingLocationInOutputTemplate_ShouldReturnError(t *testing.T) {
	sut := fixValidWebhookInput()
	sut.Mode = webhookModePtr(graphql.WebhookModeAsync)
	sut.OutputTemplate = stringPtr(`{
	   "success_status_code": 202,
	   "error": "{{.Body.error}}"
	}`)
	//WHEN
	err := sut.Validate()
	//THEN
	require.Error(t, err)
}

func fixValidWebhookInput() graphql.WebhookInput {
	template := `{}`
	outputTemplate := `{
	   "location": "{{.Headers.location}}",
	   "success_status_code": 202,
	   "error": "{{.Body.error}}"
 }`
	webhookMode := graphql.WebhookModeSync
	webhookURL := inputvalidationtest.ValidURL
	return graphql.WebhookInput{
		Type:           graphql.WebhookTypeConfigurationChanged,
		URL:            &webhookURL,
		Mode:           &webhookMode,
		InputTemplate:  &template,
		HeaderTemplate: &template,
		OutputTemplate: &outputTemplate,
	}
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(n int) *int {
	return &n
}

func webhookModePtr(mode graphql.WebhookMode) *graphql.WebhookMode {
	return &mode
}
