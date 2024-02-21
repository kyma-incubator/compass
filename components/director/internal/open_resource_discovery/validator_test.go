package ord_test

import (
	"context"
	"encoding/json"
	"fmt"
	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDocumentValidator_Validate(t *testing.T) {
	testCases := []struct {
		Name                     string
		ClientValidatorFn        func() *automock.ValidatorClient
		InputDocument            string
		InputBaseUrl             string
		ExpectedRuntimeError     error
		ExpectedValidationErrors []*ord.ValidationError
	}{
		{
			Name: "Success without errors",
			ClientValidatorFn: func() *automock.ValidatorClient {
				clientValidator := &automock.ValidatorClient{}
				clientValidator.On("Validate", policyLevelBase, mock.Anything).Return([]ord.ValidationResult{}, nil)
				return clientValidator
			},
			InputDocument: fmt.Sprintf(OrdDocument, baseURL),
			InputBaseUrl:  baseURL,
		},
		{
			Name: "Runtime error from API Metadata Validator",
			ClientValidatorFn: func() *automock.ValidatorClient {
				clientValidator := &automock.ValidatorClient{}
				clientValidator.On("Validate", policyLevelBase, mock.Anything).Return([]ord.ValidationResult{}, errors.New("Test runtime error"))
				return clientValidator
			},
			InputDocument:        fmt.Sprintf(OrdDocument, baseURL),
			InputBaseUrl:         baseURL,
			ExpectedRuntimeError: errors.New("error while validating document with API Metadata validator"),
		},
		{
			Name: "Validation errors only with severity level Error",
			ClientValidatorFn: func() *automock.ValidatorClient {
				clientValidator := &automock.ValidatorClient{}
				clientValidator.On("Validate", policyLevelBase, mock.Anything).Return(validationResultsErrorSeverity, nil)
				return clientValidator
			},
			InputDocument:            fmt.Sprintf(OrdDocument, baseURL),
			InputBaseUrl:             baseURL,
			ExpectedValidationErrors: validationErrorsErrorSeverity,
		},
		{
			Name: "Validation errors with severity level Warning",
			ClientValidatorFn: func() *automock.ValidatorClient {
				clientValidator := &automock.ValidatorClient{}
				clientValidator.On("Validate", policyLevelBase, mock.Anything).Return(validationResultsWarningSeverity, nil)
				return clientValidator
			},
			InputDocument:            fmt.Sprintf(OrdDocument, baseURL),
			InputBaseUrl:             baseURL,
			ExpectedValidationErrors: validationErrorsWarningSeverity,
		},
		{
			Name: "Validation error when there is a duplicate API resource",
			ClientValidatorFn: func() *automock.ValidatorClient {
				clientValidator := &automock.ValidatorClient{}
				clientValidator.On("Validate", policyLevelBase, mock.Anything).Return([]ord.ValidationResult{}, nil)
				return clientValidator
			},
			InputDocument:            fmt.Sprintf(OrdDocumentWithDuplicates, baseURL),
			InputBaseUrl:             baseURL,
			ExpectedValidationErrors: validationErrorDuplicateResources,
		},
		{
			Name: "Validation error when there is a resource with unknown reference",
			ClientValidatorFn: func() *automock.ValidatorClient {
				clientValidator := &automock.ValidatorClient{}
				clientValidator.On("Validate", policyLevelBase, mock.Anything).Return([]ord.ValidationResult{}, nil)
				return clientValidator
			},
			InputDocument:            fmt.Sprintf(OrdDocumentAPIHasUnknownReference, baseURL),
			InputBaseUrl:             baseURL,
			ExpectedValidationErrors: validationErrorUnknownReference,
		},
		{
			Name: "Validation error when baseUrl is missing",
			ClientValidatorFn: func() *automock.ValidatorClient {
				clientValidator := &automock.ValidatorClient{}
				clientValidator.On("Validate", policyLevelBase, mock.Anything).Return([]ord.ValidationResult{}, nil)
				return clientValidator
			},
			InputDocument:            fmt.Sprintf(OrdDocumentWithWrongBaseUrl, ""),
			InputBaseUrl:             "",
			ExpectedValidationErrors: validationErrorMissingBaseUrl,
		},
		{
			Name: "Validation error when there is a mismatch between the given baseUrls",
			ClientValidatorFn: func() *automock.ValidatorClient {
				clientValidator := &automock.ValidatorClient{}
				clientValidator.On("Validate", policyLevelBase, mock.Anything).Return([]ord.ValidationResult{}, nil)
				return clientValidator
			},
			InputDocument:            fmt.Sprintf(OrdDocumentWithWrongBaseUrl, "https://differentbase.com"),
			InputBaseUrl:             baseURL,
			ExpectedValidationErrors: validationErrorMismatchedBaseUrl,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			validator := ord.NewDocumentValidator(test.ClientValidatorFn())

			result := &ord.Document{}
			json.Unmarshal([]byte(test.InputDocument), &result)

			validationErrors, err := validator.Validate(context.TODO(), []*ord.Document{result}, test.InputBaseUrl, map[string]bool{}, []string{test.InputDocument})

			if test.ExpectedRuntimeError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedRuntimeError.Error())
			} else {
				require.NoError(t, err)
			}

			if len(test.ExpectedValidationErrors) > 0 {
				require.Equal(t, len(test.ExpectedValidationErrors), len(validationErrors))

				for i, currentError := range validationErrors {
					require.Equal(t, test.ExpectedValidationErrors[i].OrdID, currentError.OrdID)
					require.Equal(t, test.ExpectedValidationErrors[i].Severity, currentError.Severity)
					require.Equal(t, test.ExpectedValidationErrors[i].Type, currentError.Type)
					require.Equal(t, test.ExpectedValidationErrors[i].Description, currentError.Description)
				}
			} else {
				require.Empty(t, validationErrors)
			}
		})
	}
}
