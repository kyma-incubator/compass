package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

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
			sut := fixValidWebhookInput(inputvalidationtest.ValidURL)
			sut.Type = testCase.Value
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
			sut := fixValidWebhookInput(inputvalidationtest.ValidURL)
			sut.URL = &testCase.Value
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
			sut := fixValidWebhookInput(inputvalidationtest.ValidURL)
			sut.Auth = testCase.Value
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
			sut := fixValidWebhookInput(inputvalidationtest.ValidURL)
			sut.CorrelationIDKey = testCase.Value
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

func TestWebhookInput_Validate_Mode(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *graphql.WebhookMode
		Type          *graphql.WebhookType
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValidMode",
			Value:         webhookModePtr(graphql.WebhookModeSync),
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValidSyncModeAndConfigurationChangedType",
			Value:         webhookModePtr(graphql.WebhookModeSync),
			Type:          webhookTypePtr(graphql.WebhookTypeConfigurationChanged),
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValidAsyncCallbackModeAndConfigurationChangedType",
			Value:         webhookModePtr(graphql.WebhookModeAsyncCallback),
			Type:          webhookTypePtr(graphql.WebhookTypeConfigurationChanged),
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValidAsyncCallbackModeAndAppTntMappingType",
			Value:         webhookModePtr(graphql.WebhookModeAsyncCallback),
			Type:          webhookTypePtr(graphql.WebhookTypeApplicationTenantMapping),
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValidSyncModeAndFormationLifecycleType",
			Value:         webhookModePtr(graphql.WebhookModeSync),
			Type:          webhookTypePtr(graphql.WebhookTypeFormationLifecycle),
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedInvalidAsyncModeAndConfigurationChangedType",
			Value:         webhookModePtr(graphql.WebhookModeAsync),
			Type:          webhookTypePtr(graphql.WebhookTypeConfigurationChanged),
			ExpectedValid: false,
		},
		{
			Name:          "ExpectedInvalidAsyncModeAndAppTntMappingType",
			Value:         webhookModePtr(graphql.WebhookModeAsync),
			Type:          webhookTypePtr(graphql.WebhookTypeApplicationTenantMapping),
			ExpectedValid: false,
		},
		{
			Name:          "ExpectedInvalidAsyncCallbackModeAndRegisterAppType",
			Value:         webhookModePtr(graphql.WebhookModeAsyncCallback),
			Type:          webhookTypePtr(graphql.WebhookTypeRegisterApplication),
			ExpectedValid: false,
		},
		{
			Name:          "ExpectedInvalidAsyncCallbackModeAndFormationLifecycleType",
			Value:         webhookModePtr(graphql.WebhookModeAsyncCallback),
			Type:          webhookTypePtr(graphql.WebhookTypeFormationLifecycle),
			ExpectedValid: false,
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
			sut := fixValidWebhookInput(inputvalidationtest.ValidURL)
			sut.Mode = testCase.Value
			if testCase.Type != nil {
				sut.Type = *testCase.Type
			}
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
			sut := fixValidWebhookInput(inputvalidationtest.ValidURL)
			sut.RetryInterval = testCase.Value
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
			sut := fixValidWebhookInput(inputvalidationtest.ValidURL)
			sut.Timeout = testCase.Value
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

func TestWebhookInput_Validate_URLTemplate(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name: "ExpectedValid",
			Value: stringPtr(`{
	   "method": "POST",
	   "path": "https://my-int-system/api/v1/{{.Application.ID}}/pairing"
 }`),
			ExpectedValid: true,
		},
		{
			Name: "InvalidURL",
			Value: stringPtr(`{
	   "method": "POST",
	   "path": "abc"
 }`),
			ExpectedValid: false,
		},
		{
			Name: "MissingPath",
			Value: stringPtr(`{
	   "method": "POST"
 }`),
			ExpectedValid: false,
		},
		{
			Name: "MissingMethod",
			Value: stringPtr(`{
	   "path": "https://my-int-system/api/v1/{{.Application.ID}}/pairing"
 }`),
			ExpectedValid: false,
		},
		{
			Name: "MethodNotAllowed",
			Value: stringPtr(`{
		"method": "HEAD",
	   	"path": "https://my-int-system/api/v1/{{.Application.ID}}/pairing"
 }`),
			ExpectedValid: false,
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
			sut := fixValidWebhookInput(inputvalidationtest.ValidURL)
			sut.URL = nil
			sut.URLTemplate = testCase.Value
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
			sut := fixValidWebhookInput(inputvalidationtest.ValidURL)
			sut.InputTemplate = testCase.Value
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

func TestWebhookInput_Validate_HeaderTemplate(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         stringPtr(`{"Content-Type":["application/json"],"client_user":["{{.Headers.User_id}}"]}`),
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
			Name:          "Invalid - contains unexpected property",
			Value:         stringPtr(`{"Content-Type":["application/json"],"client_user":["{{.Body.Group}}"]}`),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidWebhookInput(inputvalidationtest.ValidURL)
			sut.HeaderTemplate = testCase.Value
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

func TestWebhookInput_Validate_OutputTemplate(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name: "ExpectedValid",
			Value: stringPtr(`{
			   "location": "{{.Headers.Location}}",
			   "success_status_code": 202,
			   "error": "{{.Body.error}}"
			 }`),
			ExpectedValid: true,
		},
		{
			Name: "ExpectedValid - missing location when SYNC mode",
			Value: stringPtr(`{
			   "success_status_code": 202,
			   "error": "{{.Body.error}}"
			 }`),
			ExpectedValid: true,
		},
		{
			Name: "Invalid - missing success status code",
			Value: stringPtr(`{
			   "location": "{{.Headers.Location}}",
			   "error": "{{.Body.error}}"
			 }`),
			ExpectedValid: false,
		},
		{
			Name: "Invalid - missing error",
			Value: stringPtr(`{
			   "location": "{{.Headers.Location}}",
			   "success_status_code": 202
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
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidWebhookInput(inputvalidationtest.ValidURL)
			sut.InputTemplate = nil
			sut.OutputTemplate = testCase.Value
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
			   "success_status_identifier": "SUCCESS",
			   "in_progress_status_identifier": "IN_PROGRESS",
			   "failed_status_identifier": "FAILED",
			   "error": "{{.Body.error}}"
			}`),
			ExpectedValid: true,
		},
		{
			Name: "Invalid - missing error",
			Value: stringPtr(`{
			   "status": "{{.Body.status}}",
			   "success_status_code": 200,
			   "success_status_identifier": "SUCCESS",
			   "in_progress_status_identifier": "IN_PROGRESS",
			   "failed_status_identifier": "FAILED"
			 }`),
			ExpectedValid: false,
		},
		{
			Name: "Invalid -  missing status",
			Value: stringPtr(`{
			   "success_status_code": 200,
			   "success_status_identifier": "SUCCESS",
			   "in_progress_status_identifier": "IN_PROGRESS",
			   "failed_status_identifier": "FAILED",
			   "error": "{{.Body.error}}"
			 }`),
			ExpectedValid: false,
		},
		{
			Name: "Invalid - missing success status code",
			Value: stringPtr(`{
			   "status": "{{.Body.status}}",
			   "success_status_identifier": "SUCCESS",
			   "in_progress_status_identifier": "IN_PROGRESS",
			   "failed_status_identifier": "FAILED",
			   "error": "{{.Body.error}}"
			 }`),
			ExpectedValid: false,
		},
		{
			Name: "Invalid - missing success status identifier",
			Value: stringPtr(`{
			   "status": "{{.Body.status}}",
			   "success_status_code": 200,
			   "in_progress_status_identifier": "IN_PROGRESS",
			   "failed_status_identifier": "FAILED",
			   "error": "{{.Body.error}}"
			}`),
			ExpectedValid: false,
		},
		{
			Name: "Invalid - missing in progress status identifier",
			Value: stringPtr(`{
			   "status": "{{.Body.status}}",
			   "success_status_code": 200,
			   "success_status_identifier": "SUCCESS",
			   "failed_status_identifier": "FAILED",
			   "error": "{{.Body.error}}"
			}`),
			ExpectedValid: false,
		},
		{
			Name: "Invalid - missing failed status identifier",
			Value: stringPtr(`{
			   "status": "{{.Body.status}}",
			   "success_status_code": 200,
			   "success_status_identifier": "SUCCESS",
			   "in_progress_status_identifier": "IN_PROGRESS",
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
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidWebhookInput(inputvalidationtest.ValidURL)
			sut.Mode = webhookModePtr(graphql.WebhookModeAsync)
			sut.StatusTemplate = testCase.Value
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

func TestWebhookInput_Validate_BothURLAndURLTemplate(t *testing.T) {
	sut := fixValidWebhookInput(inputvalidationtest.ValidURL)
	sut.URL = stringPtr("https://my-int-system/api/v1/123/pairing")
	sut.URLTemplate = stringPtr("https://my-int-system/api/v1/{{.Application.ID}}/pairing")
	// WHEN
	err := sut.Validate()
	// THEN
	require.Error(t, err)
}

func TestWebhookInput_Validate_AsyncWebhook_MissingLocationInOutputTemplate_ShouldReturnError(t *testing.T) {
	sut := fixValidWebhookInput(inputvalidationtest.ValidURL)
	sut.Mode = webhookModePtr(graphql.WebhookModeAsync)
	sut.OutputTemplate = stringPtr(`{
	   "success_status_code": 202,
	   "error": "{{.Body.error}}"
	}`)
	// WHEN
	err := sut.Validate()
	// THEN
	require.Error(t, err)
}

func TestWebhookInput_Validate_MissingOutputTemplateForCertainTypes(t *testing.T) {
	webhookMode := graphql.WebhookModeSync
	testCases := []struct {
		Name          string
		Input         graphql.WebhookInput
		ExpectedValid bool
	}{
		{
			Name:          "Success when missing for type: OPEN_RESOURCE_DISCOVERY",
			ExpectedValid: true,
			Input: graphql.WebhookInput{
				Type: graphql.WebhookTypeOpenResourceDiscovery,
				Mode: &webhookMode,
				URL:  str.Ptr(inputvalidationtest.ValidURL),
			},
		},
		{
			Name:          "Success when missing for type: CONFIGURATION_CHANGED",
			ExpectedValid: true,
			Input: graphql.WebhookInput{
				Type: graphql.WebhookTypeConfigurationChanged,
				Mode: &webhookMode,
				URL:  str.Ptr(inputvalidationtest.ValidURL),
			},
		},
		{
			Name:          "Fails when missing for type: UNREGISTER_APPLICATION",
			ExpectedValid: false,
			Input: graphql.WebhookInput{
				Type: graphql.WebhookTypeUnregisterApplication,
				Mode: &webhookMode,
				URL:  str.Ptr(inputvalidationtest.ValidURL),
			},
		},
		{
			Name:          "Fails when missing for type: REGISTER_APPLICATION",
			ExpectedValid: false,
			Input: graphql.WebhookInput{
				Type: graphql.WebhookTypeRegisterApplication,
				Mode: &webhookMode,
				URL:  str.Ptr(inputvalidationtest.ValidURL),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			err := testCase.Input.Validate()
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidWebhookInput(url string) graphql.WebhookInput {
	template := `{}`
	outputTemplate := `{
	   "location": "{{.Headers.Location}}",
	   "success_status_code": 202,
	   "error": "{{.Body.error}}"
 }`
	webhookMode := graphql.WebhookModeSync
	webhookInput := graphql.WebhookInput{
		Type:           graphql.WebhookTypeUnregisterApplication,
		Mode:           &webhookMode,
		InputTemplate:  &template,
		HeaderTemplate: &template,
		OutputTemplate: &outputTemplate,
	}

	if url != "" {
		webhookInput.URL = &url
	} else {
		webhookInput.URLTemplate = stringPtr(`{
	   "method": "POST",
	   "path": "https://my-int-system/api/v1/{{.Application.ID}}/pairing"
 }`)
	}

	return webhookInput
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

func webhookTypePtr(whType graphql.WebhookType) *graphql.WebhookType {
	return &whType
}
